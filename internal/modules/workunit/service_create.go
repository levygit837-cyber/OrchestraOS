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
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
)

func (s *WorkUnitService) createMany(ctx context.Context, inputs []CreateWorkUnitInput) ([]*transition.OperationResult[*WorkUnit], error) {
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
	var graph *taskgraphmod.TaskGraph
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
		eventID := input.EventID
		if len(inputs) > 1 {
			eventID = ""
		}
		appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
			ID:          eventID,
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

func (s *WorkUnitService) ensureActiveManualTaskGraph(ctx context.Context, tx *sql.Tx, task *task.Task) (*taskgraphmod.TaskGraph, error) {
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
	version, err := repo.NextVersion(task.ID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "work_unit_service.next_task_graph_version", err)
	}
	now := time.Now().UTC()
	graph := &taskgraphmod.TaskGraph{
		ID:              uuid.New().String(),
		TaskID:          task.ID,
		Version:         version,
		Status:          taskgraphmod.StatusActive,
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

func isManualTaskGraph(graph *taskgraphmod.TaskGraph) bool {
	if graph == nil {
		return false
	}
	return graph.PlannerStrategy == "manual" || graph.PlannerStrategy == "legacy_manual"
}
