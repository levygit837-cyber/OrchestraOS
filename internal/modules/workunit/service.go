// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package workunit

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// TaskReader abstracts task reads to avoid cyclic imports.
type TaskReader interface {
	GetByID(id string) (*domain.Task, error)
}

// TaskGraphManager abstracts task-graph operations to avoid cyclic imports.
type TaskGraphManager interface {
	GetActiveByTask(taskID string) (*domain.TaskGraph, error)
	GetByID(id string) (*domain.TaskGraph, error)
	GetNextVersion(taskID string) (int, error)
	Create(graph *domain.TaskGraph) error
}

type WorkUnitService struct {
	db                  *sql.DB
	newTaskReader       func(*sql.Tx) TaskReader
	newTaskGraphManager func(*sql.Tx) TaskGraphManager
}

type CreateWorkUnitInput struct {
	ID                   string
	EventID              string
	TaskID               string
	TaskGraphID          string
	Title                string
	Objective            string
	AssignedAgentProfile string
	OwnedPaths           []string
	ReadPaths            []string
	AcceptanceCriteria   []string
	ValidationPlan       []string
	DependsOn            []string
}

func NewWorkUnitService(database *sql.DB, newTaskReader func(*sql.Tx) TaskReader, newTaskGraphManager func(*sql.Tx) TaskGraphManager) *WorkUnitService {
	return &WorkUnitService{db: database, newTaskReader: newTaskReader, newTaskGraphManager: newTaskGraphManager}
}

func (s *WorkUnitService) Create(ctx context.Context, input CreateWorkUnitInput) (*transition.OperationResult[*WorkUnit], error) {
	result, err := s.createMany(ctx, []CreateWorkUnitInput{input})
	if err != nil {
		return nil, err
	}
	return result[0], nil
}

func (s *WorkUnitService) CreateMany(ctx context.Context, inputs []CreateWorkUnitInput) ([]*transition.OperationResult[*WorkUnit], error) {
	return s.createMany(ctx, inputs)
}

func (s *WorkUnitService) Assign(ctx context.Context, workUnitID, agentProfile string, input transition.TransitionInput) (*transition.OperationResult[*WorkUnit], error) {
	op := "work_unit_service.assign"
	if err := validation.RequiredUUID(workUnitID, "work_unit_id", op); err != nil {
		return nil, err
	}
	if err := validation.RequiredText(agentProfile, "assigned_agent_profile", op); err != nil {
		return nil, err
	}
	tx, err := dbcore.BeginTx(ctx, s.db, "work_unit_service.begin_assign")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	wu, err := RequireByID(ctx, tx, workUnitID)
	if err != nil {
		return nil, err
	}
	event, duplicate, err := transition.AppendTransition(ctx, tx, input.EventID, "work_unit.assigned", wu.TaskID, "", wu.ID, input.AgentID, map[string]interface{}{
		"from_agent_profile": wu.AssignedAgentProfile,
		"to_agent_profile":   agentProfile,
		"justification":      input.Justification,
	})
	if err != nil {
		return nil, err
	}
	if !duplicate {
		res, err := tx.ExecContext(ctx, QueryUpdateAssignment, wu.ID, agentProfile, time.Now().UTC())
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.update_assignment", err)
		}
		if err := dbcore.EnsureRowsAffected(res, "work unit", "work_unit_service.update_assignment"); err != nil {
			return nil, err
		}
		wu.AssignedAgentProfile = agentProfile
	}
	if err := dbcore.CommitTx(tx, "work_unit_service.commit_assign"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*WorkUnit]{Value: wu, Event: event, Duplicate: duplicate}, nil
}

func (s *WorkUnitService) Block(ctx context.Context, workUnitID string, input transition.TransitionInput) (*transition.OperationResult[*WorkUnit], error) {
	return s.transition(ctx, workUnitID, StatusBlocked, input)
}

func (s *WorkUnitService) Schedule(ctx context.Context, workUnitID string, input transition.TransitionInput) (*transition.OperationResult[*WorkUnit], error) {
	return s.transition(ctx, workUnitID, StatusScheduled, input)
}

func (s *WorkUnitService) Start(ctx context.Context, workUnitID string, input transition.TransitionInput) (*transition.OperationResult[*WorkUnit], error) {
	return s.transition(ctx, workUnitID, StatusRunning, input)
}

func (s *WorkUnitService) Validate(ctx context.Context, workUnitID string, input transition.TransitionInput) (*transition.OperationResult[*WorkUnit], error) {
	return s.transition(ctx, workUnitID, StatusValidating, input)
}

func (s *WorkUnitService) Complete(ctx context.Context, workUnitID string, input transition.TransitionInput) (*transition.OperationResult[*WorkUnit], error) {
	return s.transition(ctx, workUnitID, StatusCompleted, input)
}

func (s *WorkUnitService) Fail(ctx context.Context, workUnitID string, input transition.TransitionInput) (*transition.OperationResult[*WorkUnit], error) {
	return s.transition(ctx, workUnitID, StatusFailed, input)
}

func (s *WorkUnitService) Cancel(ctx context.Context, workUnitID string, input transition.TransitionInput) (*transition.OperationResult[*WorkUnit], error) {
	return s.transition(ctx, workUnitID, StatusCancelled, input)
}

func (s *WorkUnitService) transition(ctx context.Context, workUnitID string, target Status, input transition.TransitionInput) (*transition.OperationResult[*WorkUnit], error) {
	op := "work_unit_service.transition"
	if err := validation.RequiredUUID(workUnitID, "work_unit_id", op); err != nil {
		return nil, err
	}
	if err := transition.RequireFinalAudit(string(target), input, op); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "work_unit_service.begin_transition")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	wu, err := RequireByID(ctx, tx, workUnitID)
	if err != nil {
		return nil, err
	}
	if target == StatusScheduled || target == StatusRunning {
		if err := dbcore.AcquireAdvisoryTxLock(ctx, tx, "work_unit_paths:"+wu.TaskID, "work_unit_service.path_lock"); err != nil {
			return nil, err
		}
		if err := ValidateDependenciesCompleted(ctx, tx, wu); err != nil {
			return nil, err
		}
		if err := ValidateOwnedPathAvailability(ctx, tx, wu); err != nil {
			return nil, err
		}
	}
	if target == StatusCompleted && len(wu.AcceptanceCriteria) == 0 && input.Justification == "" {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "work unit completion requires acceptance criteria or explicit justification")
	}
	if err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(wu.Status), string(target), transition.TransitionContext(input)); err != nil {
		return nil, err
	}

	event, duplicate, err := transition.AppendTransition(ctx, tx, input.EventID, EventTypeForStatus(target), wu.TaskID, "", wu.ID, input.AgentID, transition.TransitionPayload(wu.Status, target, input))
	if err != nil {
		return nil, err
	}
	if !duplicate {
		res, err := NewRepository(tx).UpdateStatus(wu.ID, target, time.Now().UTC())
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.update_projection", err)
		}
		if err := dbcore.EnsureRowsAffected(res, "work unit", "work_unit_service.update_projection"); err != nil {
			return nil, err
		}
		wu.Status = target
	}
	if err := dbcore.CommitTx(tx, "work_unit_service.commit_transition"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*WorkUnit]{Value: wu, Event: event, Duplicate: duplicate}, nil
}

func (s *WorkUnitService) requireTaskByID(tx *sql.Tx, id string) (*domain.Task, error) {
	task, err := s.newTaskReader(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task.get", err)
	}
	if task == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "task.get", "task not found")
	}
	return task, nil
}

func (s *WorkUnitService) createMany(ctx context.Context, inputs []CreateWorkUnitInput) ([]*transition.OperationResult[*WorkUnit], error) {
	op := "work_unit_service.validate_create_many"
	if len(inputs) == 0 {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "at least one work unit is required")
	}
	if len(inputs) > 1 && inputs[0].EventID != "" {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "batch creation does not support external event IDs")
	}
	taskID := inputs[0].TaskID
	taskGraphID := inputs[0].TaskGraphID
	for i := range inputs {
		if inputs[i].ID == "" {
			inputs[i].ID = uuid.New().String()
		}
		if inputs[i].AssignedAgentProfile == "" {
			inputs[i].AssignedAgentProfile = "default"
		}
		if err := validateCreateWorkUnitInput(inputs[i]); err != nil {
			return nil, err
		}
		if inputs[i].TaskID != taskID {
			return nil, apperrors.New(apperrors.CodeInvalidInput, op, "all work units in one create batch must belong to the same task")
		}
		if taskGraphID == "" {
			taskGraphID = inputs[i].TaskGraphID
		}
		if inputs[i].TaskGraphID != "" && inputs[i].TaskGraphID != taskGraphID {
			return nil, apperrors.New(apperrors.CodeInvalidInput, op, "all work units in one create batch must belong to the same task graph")
		}
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "work_unit_service.begin_create_many")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	task, err := s.requireTaskByID(tx, taskID)
	if err != nil {
		return nil, err
	}

	graphRepo := s.newTaskGraphManager(tx)
	var graph *domain.TaskGraph
	if taskGraphID == "" {
		graph, err = s.ensureActiveManualTaskGraph(ctx, tx, task)
		if err != nil {
			return nil, err
		}
		taskGraphID = graph.ID
	} else {
		graph, err = graphRepo.GetByID(taskGraphID)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.get_task_graph", err)
		}
		if graph == nil {
			return nil, apperrors.New(apperrors.CodeNotFound, "work_unit_service.get_task_graph", "task graph not found")
		}
		if graph.TaskID != taskID {
			return nil, apperrors.New(apperrors.CodeInvalidInput, "work_unit_service.task_mismatch", "task_graph_id does not belong to task_id")
		}
		if !isManualTaskGraph(graph) {
			return nil, apperrors.New(apperrors.CodeConflict, "work_unit_service.graph_immutable", "work units for planned task graphs must be created by task decomposition")
		}
	}
	for i := range inputs {
		inputs[i].TaskGraphID = taskGraphID
	}

	existing, err := NewRepository(tx).ListByTaskGraph(taskGraphID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.list_existing", err)
	}
	if err := ValidateWorkUnitDependencies(inputs, existing); err != nil {
		return nil, err
	}

	repo := NewRepository(tx)
	results := make([]*transition.OperationResult[*WorkUnit], 0, len(inputs))
	for _, input := range inputs {
		wu := &WorkUnit{
			ID:                   input.ID,
			TaskID:               input.TaskID,
			TaskGraphID:          input.TaskGraphID,
			Title:                input.Title,
			Objective:            input.Objective,
			AssignedAgentProfile: input.AssignedAgentProfile,
			Status:               StatusCreated,
			OwnedPaths:           input.OwnedPaths,
			ReadPaths:            input.ReadPaths,
			AcceptanceCriteria:   input.AcceptanceCriteria,
			ValidationPlan:       input.ValidationPlan,
			DependsOn:            input.DependsOn,
		}
		if err := repo.Create(wu); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.create_projection", err)
		}
		payload, err := serialization.MarshalPayload("work_unit_service.create_payload", map[string]interface{}{
			"work_unit_id":           wu.ID,
			"task_id":                wu.TaskID,
			"task_graph_id":          wu.TaskGraphID,
			"title":                  wu.Title,
			"objective":              wu.Objective,
			"assigned_agent_profile": wu.AssignedAgentProfile,
			"status":                 wu.Status,
			"owned_paths":            wu.OwnedPaths,
			"read_paths":             wu.ReadPaths,
			"acceptance_criteria":    wu.AcceptanceCriteria,
			"validation_plan":        wu.ValidationPlan,
			"depends_on":             wu.DependsOn,
		})
		if err != nil {
			return nil, err
		}
		appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
			ID:          input.EventID,
			Type:        "work_unit.created",
			Version:     transition.EventVersionV1,
			TaskID:      wu.TaskID,
			WorkUnitID:  wu.ID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			Payload:     payload,
		})
		if err != nil {
			return nil, err
		}
		results = append(results, &transition.OperationResult[*WorkUnit]{Value: wu, Event: &appendResult.Event, Duplicate: appendResult.Duplicate})
	}

	if err := dbcore.CommitTx(tx, "work_unit_service.commit_create_many"); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *WorkUnitService) ensureActiveManualTaskGraph(ctx context.Context, tx *sql.Tx, task *domain.Task) (*domain.TaskGraph, error) {
	if err := dbcore.AcquireAdvisoryTxLock(ctx, tx, "task_graph:"+task.ID, "work_unit_service.task_graph_lock"); err != nil {
		return nil, err
	}
	repo := s.newTaskGraphManager(tx)
	active, err := repo.GetActiveByTask(task.ID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.get_active_task_graph", err)
	}
	if active != nil {
		if !isManualTaskGraph(active) {
			return nil, apperrors.New(apperrors.CodeConflict, "work_unit_service.active_graph_exists", "active task graph was created by task decomposition")
		}
		return active, nil
	}
	version, err := repo.GetNextVersion(task.ID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.next_task_graph_version", err)
	}
	now := time.Now().UTC()
	graph := &domain.TaskGraph{
		ID:              uuid.New().String(),
		TaskID:          task.ID,
		Version:         version,
		Status:          domain.TaskGraphStatusActive,
		PlannerStrategy: "manual",
		Rationale:       "Manual graph for work units created outside task decomposition.",
		CreatedBy:       "workunit_service",
		NodeCount:       0,
		EdgeCount:       0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := repo.Create(graph); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.create_manual_task_graph", err)
	}
	return graph, nil
}

func isManualTaskGraph(graph *domain.TaskGraph) bool {
	if graph == nil {
		return false
	}
	return graph.PlannerStrategy == "manual" || graph.PlannerStrategy == "legacy_manual"
}
