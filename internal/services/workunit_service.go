package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
)

type WorkUnitService struct {
	db *sql.DB
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

func NewWorkUnitService(database *sql.DB) *WorkUnitService {
	return &WorkUnitService{db: database}
}

func (s *WorkUnitService) Create(ctx context.Context, input CreateWorkUnitInput) (*OperationResult[*domain.WorkUnit], error) {
	result, err := s.createMany(ctx, []CreateWorkUnitInput{input})
	if err != nil {
		return nil, err
	}
	return result[0], nil
}

func (s *WorkUnitService) CreateMany(ctx context.Context, inputs []CreateWorkUnitInput) ([]*OperationResult[*domain.WorkUnit], error) {
	return s.createMany(ctx, inputs)
}

func (s *WorkUnitService) Assign(ctx context.Context, workUnitID, agentProfile string, input TransitionInput) (*OperationResult[*domain.WorkUnit], error) {
	op := "work_unit_service.assign"
	if err := validateRequiredUUID(workUnitID, "work_unit_id", op); err != nil {
		return nil, err
	}
	if err := validateRequiredText(agentProfile, "assigned_agent_profile", op); err != nil {
		return nil, err
	}
	tx, err := beginTx(ctx, s.db, "work_unit_service.begin_assign")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	wu, err := getWorkUnit(ctx, tx, workUnitID)
	if err != nil {
		return nil, err
	}
	event, duplicate, err := appendTransition(ctx, tx, input.EventID, "work_unit.assigned", wu.TaskID, "", wu.ID, input.AgentID, map[string]interface{}{
		"from_agent_profile": wu.AssignedAgentProfile,
		"to_agent_profile":   agentProfile,
		"justification":      input.Justification,
	})
	if err != nil {
		return nil, err
	}
	if !duplicate {
		res, err := tx.ExecContext(ctx, `UPDATE work_units SET assigned_agent_profile = $2, updated_at = $3 WHERE id = $1`, wu.ID, agentProfile, time.Now().UTC())
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.update_assignment", err)
		}
		if err := ensureRowsAffected(res, "work unit", "work_unit_service.update_assignment"); err != nil {
			return nil, err
		}
		wu.AssignedAgentProfile = agentProfile
	}
	if err := commitTx(tx, "work_unit_service.commit_assign"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.WorkUnit]{Value: wu, Event: event, Duplicate: duplicate}, nil
}

func (s *WorkUnitService) Block(ctx context.Context, workUnitID string, input TransitionInput) (*OperationResult[*domain.WorkUnit], error) {
	return s.transition(ctx, workUnitID, domain.WorkUnitStatusBlocked, input)
}

func (s *WorkUnitService) Schedule(ctx context.Context, workUnitID string, input TransitionInput) (*OperationResult[*domain.WorkUnit], error) {
	return s.transition(ctx, workUnitID, domain.WorkUnitStatusScheduled, input)
}

func (s *WorkUnitService) Start(ctx context.Context, workUnitID string, input TransitionInput) (*OperationResult[*domain.WorkUnit], error) {
	return s.transition(ctx, workUnitID, domain.WorkUnitStatusRunning, input)
}

func (s *WorkUnitService) Validate(ctx context.Context, workUnitID string, input TransitionInput) (*OperationResult[*domain.WorkUnit], error) {
	return s.transition(ctx, workUnitID, domain.WorkUnitStatusValidating, input)
}

func (s *WorkUnitService) Complete(ctx context.Context, workUnitID string, input TransitionInput) (*OperationResult[*domain.WorkUnit], error) {
	return s.transition(ctx, workUnitID, domain.WorkUnitStatusCompleted, input)
}

func (s *WorkUnitService) Fail(ctx context.Context, workUnitID string, input TransitionInput) (*OperationResult[*domain.WorkUnit], error) {
	return s.transition(ctx, workUnitID, domain.WorkUnitStatusFailed, input)
}

func (s *WorkUnitService) Cancel(ctx context.Context, workUnitID string, input TransitionInput) (*OperationResult[*domain.WorkUnit], error) {
	return s.transition(ctx, workUnitID, domain.WorkUnitStatusCancelled, input)
}

func (s *WorkUnitService) createMany(ctx context.Context, inputs []CreateWorkUnitInput) ([]*OperationResult[*domain.WorkUnit], error) {
	op := "work_unit_service.validate_create_many"
	if len(inputs) == 0 {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "at least one work unit is required")
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

	tx, err := beginTx(ctx, s.db, "work_unit_service.begin_create_many")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	task, err := getTask(ctx, tx, taskID)
	if err != nil {
		return nil, err
	}

	graphRepo := repository.NewTaskGraphRepository(tx)
	var graph *domain.TaskGraph
	if taskGraphID == "" {
		graph, err = ensureActiveManualTaskGraph(ctx, tx, task)
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

	existing, err := repository.NewWorkUnitRepository(tx).ListByTaskGraph(taskGraphID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.list_existing", err)
	}
	if err := validateWorkUnitDependencies(inputs, existing); err != nil {
		return nil, err
	}

	repo := repository.NewWorkUnitRepository(tx)
	results := make([]*OperationResult[*domain.WorkUnit], 0, len(inputs))
	for _, input := range inputs {
		wu := &domain.WorkUnit{
			ID:                   input.ID,
			TaskID:               input.TaskID,
			TaskGraphID:          input.TaskGraphID,
			Title:                input.Title,
			Objective:            input.Objective,
			AssignedAgentProfile: input.AssignedAgentProfile,
			Status:               domain.WorkUnitStatusCreated,
			OwnedPaths:           input.OwnedPaths,
			ReadPaths:            input.ReadPaths,
			AcceptanceCriteria:   input.AcceptanceCriteria,
			ValidationPlan:       input.ValidationPlan,
			DependsOn:            input.DependsOn,
		}
		if err := repo.Create(wu); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.create_projection", err)
		}
		payload, err := marshalPayload("work_unit_service.create_payload", map[string]interface{}{
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
		eventID := input.EventID
		if len(inputs) > 1 {
			eventID = ""
		}
		appendResult, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
			ID:          eventID,
			Type:        "work_unit.created",
			Version:     eventVersionV1,
			TaskID:      wu.TaskID,
			WorkUnitID:  wu.ID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			Payload:     payload,
		})
		if err != nil {
			return nil, err
		}
		results = append(results, &OperationResult[*domain.WorkUnit]{Value: wu, Event: &appendResult.Event, Duplicate: appendResult.Duplicate})
	}

	if err := commitTx(tx, "work_unit_service.commit_create_many"); err != nil {
		return nil, err
	}
	return results, nil
}

func ensureActiveManualTaskGraph(ctx context.Context, tx *sql.Tx, task *domain.Task) (*domain.TaskGraph, error) {
	if err := acquireAdvisoryTxLock(ctx, tx, "task_graph:"+task.ID, "work_unit_service.task_graph_lock"); err != nil {
		return nil, err
	}
	repo := repository.NewTaskGraphRepository(tx)
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
	version, err := repo.NextVersion(task.ID)
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

func (s *WorkUnitService) transition(ctx context.Context, workUnitID string, target domain.WorkUnitStatus, input TransitionInput) (*OperationResult[*domain.WorkUnit], error) {
	op := "work_unit_service.transition"
	if err := validateRequiredUUID(workUnitID, "work_unit_id", op); err != nil {
		return nil, err
	}
	if err := requireFinalAudit(string(target), input, op); err != nil {
		return nil, err
	}

	tx, err := beginTx(ctx, s.db, "work_unit_service.begin_transition")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	wu, err := getWorkUnit(ctx, tx, workUnitID)
	if err != nil {
		return nil, err
	}
	if target == domain.WorkUnitStatusScheduled || target == domain.WorkUnitStatusRunning {
		if err := acquireAdvisoryTxLock(ctx, tx, "work_unit_paths:"+wu.TaskID, "work_unit_service.path_lock"); err != nil {
			return nil, err
		}
		if err := validateDependenciesCompleted(ctx, tx, wu); err != nil {
			return nil, err
		}
		if err := validateOwnedPathAvailability(ctx, tx, wu); err != nil {
			return nil, err
		}
	}
	if target == domain.WorkUnitStatusCompleted && len(wu.AcceptanceCriteria) == 0 && input.Justification == "" {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "work unit completion requires acceptance criteria or explicit justification")
	}
	if err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(wu.Status), string(target), transitionContext(input)); err != nil {
		return nil, err
	}

	event, duplicate, err := appendTransition(ctx, tx, input.EventID, eventTypeForWorkUnitStatus(target), wu.TaskID, "", wu.ID, input.AgentID, transitionPayload(wu.Status, target, input))
	if err != nil {
		return nil, err
	}
	if !duplicate {
		res, err := tx.ExecContext(ctx, db.QueryWorkUnitUpdateStatus, wu.ID, target, time.Now().UTC())
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.update_projection", err)
		}
		if err := ensureRowsAffected(res, "work unit", "work_unit_service.update_projection"); err != nil {
			return nil, err
		}
		wu.Status = target
	}
	if err := commitTx(tx, "work_unit_service.commit_transition"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.WorkUnit]{Value: wu, Event: event, Duplicate: duplicate}, nil
}

func validateCreateWorkUnitInput(input CreateWorkUnitInput) error {
	op := "work_unit_service.validate_create"
	if err := validateRequiredUUID(input.ID, "work_unit_id", op); err != nil {
		return err
	}
	if err := validateOptionalUUID(input.EventID, "event_id", op); err != nil {
		return err
	}
	if err := validateRequiredUUID(input.TaskID, "task_id", op); err != nil {
		return err
	}
	if err := validateOptionalUUID(input.TaskGraphID, "task_graph_id", op); err != nil {
		return err
	}
	if err := validateRequiredText(input.Title, "title", op); err != nil {
		return err
	}
	if err := validateRequiredText(input.Objective, "objective", op); err != nil {
		return err
	}
	if err := validateRequiredText(input.AssignedAgentProfile, "assigned_agent_profile", op); err != nil {
		return err
	}
	if err := validateStringList(input.OwnedPaths, "owned_paths", op, false); err != nil {
		return err
	}
	if err := validateStringList(input.ReadPaths, "read_paths", op, false); err != nil {
		return err
	}
	if err := validateStringList(input.AcceptanceCriteria, "acceptance_criteria", op, false); err != nil {
		return err
	}
	if err := validateStringList(input.ValidationPlan, "validation_plan", op, false); err != nil {
		return err
	}
	for _, dep := range input.DependsOn {
		if err := validateRequiredUUID(dep, "depends_on", op); err != nil {
			return err
		}
	}
	return nil
}

func validateWorkUnitDependencies(inputs []CreateWorkUnitInput, existing []domain.WorkUnit) error {
	op := "work_unit_service.validate_dependencies"
	known := map[string]bool{}
	for _, wu := range existing {
		known[wu.ID] = true
	}
	graph := map[string][]string{}
	for _, input := range inputs {
		known[input.ID] = true
		graph[input.ID] = append([]string{}, input.DependsOn...)
	}
	for _, input := range inputs {
		for _, dep := range input.DependsOn {
			if !known[dep] {
				return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("dependency %s does not exist in task graph", dep))
			}
		}
	}
	visiting := map[string]bool{}
	visited := map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return apperrors.New(apperrors.CodeInvalidInput, op, "work unit dependencies must be acyclic")
		}
		visiting[id] = true
		for _, dep := range graph[id] {
			if _, ok := graph[dep]; ok {
				if err := visit(dep); err != nil {
					return err
				}
			}
		}
		visiting[id] = false
		visited[id] = true
		return nil
	}
	for id := range graph {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

func validateDependenciesCompleted(ctx context.Context, tx *sql.Tx, wu *domain.WorkUnit) error {
	_ = ctx
	if len(wu.DependsOn) == 0 {
		return nil
	}
	repo := repository.NewWorkUnitRepository(tx)
	for _, depID := range wu.DependsOn {
		dep, err := repo.GetByID(depID)
		if err != nil {
			return apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.get_dependency", err)
		}
		if dep == nil {
			return apperrors.New(apperrors.CodeInvalidInput, "work_unit_service.get_dependency", "dependency not found")
		}
		if dep.TaskID != wu.TaskID || dep.TaskGraphID != wu.TaskGraphID {
			return apperrors.New(apperrors.CodeInvalidInput, "work_unit_service.get_dependency", "dependency belongs to a different task graph")
		}
		if dep.Status != domain.WorkUnitStatusCompleted {
			return apperrors.New(apperrors.CodeInvalidTransition, "work_unit_service.dependencies", "dependencies must be completed before scheduling or starting")
		}
	}
	return nil
}

func validateOwnedPathAvailability(ctx context.Context, tx *sql.Tx, wu *domain.WorkUnit) error {
	_ = ctx
	if len(wu.OwnedPaths) == 0 {
		return nil
	}
	all, err := repository.NewWorkUnitRepository(tx).ListByTask(wu.TaskID)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.path_conflict", err)
	}
	for _, other := range all {
		if other.ID == wu.ID {
			continue
		}
		if other.Status != domain.WorkUnitStatusScheduled && other.Status != domain.WorkUnitStatusRunning && other.Status != domain.WorkUnitStatusValidating {
			continue
		}
		if pathsOverlap(wu.OwnedPaths, other.OwnedPaths) {
			return apperrors.New(apperrors.CodeConflict, "work_unit_service.path_conflict", "owned_paths conflict with another active work unit")
		}
	}
	return nil
}

func pathsOverlap(left, right []string) bool {
	for _, l := range left {
		for _, r := range right {
			if l == r {
				return true
			}
		}
	}
	return false
}
