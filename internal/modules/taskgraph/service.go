// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package taskgraph

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

const localHeuristicPlanner = "local_heuristic_v1"

// TaskReader abstracts task reads to avoid importing the task module.
type TaskReader interface {
	GetByID(id string) (*domain.Task, error)
}

// WorkUnitCreator abstracts work-unit writes to avoid importing the workunit module.
// TODO[ADR-0022]: migrar para *workunit.WorkUnit
type WorkUnitCreator interface {
	Create(wu *domain.WorkUnit) error
}

// WorkUnitLister abstracts work-unit reads to avoid importing the workunit module.
// TODO[ADR-0022]: migrar para []workunit.WorkUnit
type WorkUnitLister interface {
	ListByTaskGraph(graphID string) ([]domain.WorkUnit, error)
}

type TaskGraphService struct {
	db                 *sql.DB
	newTaskReader      func(dbcore.DBTX) TaskReader
	newWorkUnitCreator func(dbcore.DBTX) WorkUnitCreator
	newWorkUnitLister  func(dbcore.DBTX) WorkUnitLister
}

type DecomposeTaskGraphInput struct {
	TaskID          string
	EventID         string
	ReplaceActive   bool
	CreatedBy       string
	PlannerStrategy string
}

type TaskGraphDecomposeResult struct {
	Graph     *domain.TaskGraph
	WorkUnits []domain.WorkUnit
	Event     *domain.EventEnvelope
	Duplicate bool
}

func NewTaskGraphService(
	database *sql.DB,
	newTaskReader func(dbcore.DBTX) TaskReader,
	newWorkUnitCreator func(dbcore.DBTX) WorkUnitCreator,
	newWorkUnitLister func(dbcore.DBTX) WorkUnitLister,
) *TaskGraphService {
	return &TaskGraphService{
		db:                 database,
		newTaskReader:      newTaskReader,
		newWorkUnitCreator: newWorkUnitCreator,
		newWorkUnitLister:  newWorkUnitLister,
	}
}

func (s *TaskGraphService) Decompose(ctx context.Context, input DecomposeTaskGraphInput) (*TaskGraphDecomposeResult, error) {
	op := "task_graph_service.decompose"
	if err := validation.RequiredUUID(input.TaskID, "task_id", op); err != nil {
		return nil, err
	}
	if err := validation.OptionalUUID(input.EventID, "event_id", op); err != nil {
		return nil, err
	}
	if input.CreatedBy == "" {
		input.CreatedBy = "task_graph_service"
	}
	if input.PlannerStrategy == "" {
		input.PlannerStrategy = localHeuristicPlanner
	}
	if input.EventID != "" {
		if result, duplicate, err := s.duplicateResult(ctx, input); err != nil {
			return nil, err
		} else if duplicate {
			return result, nil
		}
	}

	task, err := s.newTaskReader(s.db).GetByID(input.TaskID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.get_task", err)
	}
	if task == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "task_graph_service.get_task", "task not found")
	}

	plan, strategyUsed, rationale := s.buildPlan(ctx, task, input.PlannerStrategy)

	tx, err := dbcore.BeginTx(ctx, s.db, "task_graph_service.begin_decompose")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	if err := dbcore.AcquireAdvisoryTxLock(ctx, tx, "task_graph:"+task.ID, "task_graph_service.task_graph_lock"); err != nil {
		return nil, err
	}
	if input.EventID != "" {
		if result, duplicate, err := s.duplicateResultInTx(ctx, tx, input); err != nil {
			return nil, err
		} else if duplicate {
			if err := dbcore.CommitTx(tx, "task_graph_service.commit_duplicate"); err != nil {
				return nil, err
			}
			return result, nil
		}
	}

	task, err = s.newTaskReader(tx).GetByID(input.TaskID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "taskgraph.get_task", err)
	}
	if task == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "taskgraph.get_task", "task not found")
	}
	graphRepo := NewRepository(tx)
	active, err := graphRepo.GetActiveByTask(task.ID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.get_active_graph", err)
	}
	if active != nil && !input.ReplaceActive {
		return nil, apperrors.New(apperrors.CodeConflict, "task_graph_service.active_graph_exists", "active task graph already exists")
	}
	version, err := graphRepo.NextVersion(task.ID)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.next_version", err)
	}
	now := time.Now().UTC()
	if active != nil {
		if err := graphRepo.SupersedeActiveByTask(task.ID, now); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.supersede_active", err)
		}
	}

	graph := &domain.TaskGraph{
		ID:              plan.GraphID,
		TaskID:          task.ID,
		Version:         version,
		Status:          domain.TaskGraphStatusActive,
		PlannerStrategy: strategyUsed,
		Rationale:       rationale,
		CreatedBy:       input.CreatedBy,
		NodeCount:       len(plan.WorkUnits),
		EdgeCount:       len(plan.Edges),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := graphRepo.Create(graph); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.persistence", err)
	}

	payload, err := serialization.MarshalPayload("task_graph_service.graph_payload", map[string]interface{}{
		"task_id":          graph.TaskID,
		"graph_id":         graph.ID,
		"graph_version":    graph.Version,
		"planner_strategy": graph.PlannerStrategy,
		"rationale":        graph.Rationale,
		"created_by":       graph.CreatedBy,
		"nodes":            plan.Nodes,
		"edges":            plan.Edges,
	})
	if err != nil {
		return nil, err
	}
	graphAppend, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "task.graph_created",
		Version:     transition.EventVersionV1,
		TaskID:      task.ID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}
	if graphAppend.Duplicate {
		return nil, apperrors.New(apperrors.CodeConflict, "task_graph_service.idempotency", "event_id became duplicate during graph creation")
	}

	workUnitCreator := s.newWorkUnitCreator(tx)
	for i := range plan.WorkUnits {
		wu := plan.WorkUnits[i]
		if err := workUnitCreator.Create(&wu); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.persistence", err)
		}
		wuPayload, err := serialization.MarshalPayload("task_graph_service.work_unit_payload", map[string]interface{}{
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
		if _, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
			Type:        "work_unit.created",
			Version:     transition.EventVersionV1,
			TaskID:      wu.TaskID,
			WorkUnitID:  wu.ID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			Payload:     wuPayload,
		}); err != nil {
			return nil, err
		}
		plan.WorkUnits[i] = wu
	}

	if err := dbcore.CommitTx(tx, "task_graph_service.commit_decompose"); err != nil {
		return nil, err
	}
	return &TaskGraphDecomposeResult{
		Graph:     graph,
		WorkUnits: plan.WorkUnits,
		Event:     &graphAppend.Event,
	}, nil
}

// buildPlan selects and executes the appropriate planner, with automatic fallback to heuristic on failure.
func (s *TaskGraphService) buildPlan(ctx context.Context, task *domain.Task, strategy string) (*GraphPlan, string, string) {
	if strategy == localHeuristicPlanner {
		plan, err := BuildLocalHeuristicGraphPlan(task)
		if err != nil {
			// Heuristic should rarely fail, but if it does we still return a minimal plan
			return s.buildFallbackPlan(task, fmt.Sprintf("heuristic failed: %v", err))
		}
		return plan, localHeuristicPlanner, plan.Rationale
	}

	if strategy == llmGeminiPlanner {
		planner, err := NewGeminiPlanner()
		if err != nil {
			plan, _ := BuildLocalHeuristicGraphPlan(task)
			rationale := fmt.Sprintf("LLM planner initialization failed (%v), fallback to %s", err, localHeuristicPlanner)
			return plan, localHeuristicPlanner, rationale
		}

		plan, err := planner.Plan(ctx, task)
		if err != nil {
			plan, _ := BuildLocalHeuristicGraphPlan(task)
			rationale := fmt.Sprintf("LLM planner failed (%v), fallback to %s", err, localHeuristicPlanner)
			return plan, localHeuristicPlanner, rationale
		}

		if err := ValidateGraphPlan(plan); err != nil {
			plan, _ := BuildLocalHeuristicGraphPlan(task)
			rationale := fmt.Sprintf("LLM plan validation failed (%v), fallback to %s", err, localHeuristicPlanner)
			return plan, localHeuristicPlanner, rationale
		}

		return plan, llmGeminiPlanner, plan.Rationale
	}

	// Unknown strategy: fallback to heuristic
	plan, err := BuildLocalHeuristicGraphPlan(task)
	if err != nil {
		return s.buildFallbackPlan(task, fmt.Sprintf("unknown strategy %q and heuristic failed: %v", strategy, err))
	}
	rationale := fmt.Sprintf("Unknown strategy %q, fallback to %s", strategy, localHeuristicPlanner)
	return plan, localHeuristicPlanner, rationale
}

// buildFallbackPlan creates a minimal valid plan when everything else fails.
func (s *TaskGraphService) buildFallbackPlan(task *domain.Task, reason string) (*GraphPlan, string, string) {
	graphID := uuid.New().String()
	wu := domain.WorkUnit{
		ID:                   uuid.New().String(),
		TaskID:               task.ID,
		TaskGraphID:          graphID,
		Title:                task.Title,
		Objective:            task.Description,
		AssignedAgentProfile: "default",
		Status:               domain.WorkUnitStatusCreated,
		OwnedPaths:           []string{},
		ReadPaths:            []string{},
		AcceptanceCriteria:   task.AcceptanceCriteria,
		ValidationPlan:       []string{"Validar critérios de aceite da task e registrar evidência."},
		DependsOn:            []string{},
	}
	return &GraphPlan{
		GraphID:   graphID,
		WorkUnits: []domain.WorkUnit{wu},
		Nodes: []domain.TaskGraphNodeInfo{{
			ID:                 wu.ID,
			Title:              wu.Title,
			Objective:          wu.Objective,
			AgentProfile:       wu.AssignedAgentProfile,
			OwnedPaths:         wu.OwnedPaths,
			ReadPaths:          wu.ReadPaths,
			AcceptanceCriteria: wu.AcceptanceCriteria,
			ValidationPlan:     wu.ValidationPlan,
		}},
		Edges:     []domain.TaskGraphEdgeInfo{},
		Rationale: fmt.Sprintf("Emergency fallback plan: %s", reason),
	}, localHeuristicPlanner, fmt.Sprintf("Emergency fallback plan: %s", reason)
}

func (s *TaskGraphService) ListByTask(ctx context.Context, taskID string) ([]domain.TaskGraph, error) {
	_ = ctx
	if err := validation.RequiredUUID(taskID, "task_id", "task_graph_service.list_by_task"); err != nil {
		return nil, err
	}
	return NewRepository(s.db).ListByTask(taskID)
}

func (s *TaskGraphService) duplicateResult(ctx context.Context, input DecomposeTaskGraphInput) (*TaskGraphDecomposeResult, bool, error) {
	return s.duplicateResultWithExecutor(ctx, s.db, input)
}

func (s *TaskGraphService) duplicateResultInTx(ctx context.Context, tx *sql.Tx, input DecomposeTaskGraphInput) (*TaskGraphDecomposeResult, bool, error) {
	return s.duplicateResultWithExecutor(ctx, tx, input)
}

func (s *TaskGraphService) duplicateResultWithExecutor(ctx context.Context, executor dbcore.DBTX, input DecomposeTaskGraphInput) (*TaskGraphDecomposeResult, bool, error) {
	_ = ctx
	store, err := eventstore.NewStoreWithExecutor(executor)
	if err != nil {
		return nil, false, err
	}
	existing, err := store.Get(input.EventID)
	if err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.get_existing_event", err)
	}
	if existing == nil {
		return nil, false, nil
	}
	if existing.Type != "task.graph_created" || existing.TaskID != input.TaskID {
		return nil, false, apperrors.New(apperrors.CodeConflict, "task_graph_service.idempotency", "event_id already exists for a different operation")
	}
	var payload domain.TaskGraphCreatedPayload
	if err := json.Unmarshal(existing.Payload, &payload); err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodeValidation, "task_graph_service.existing_payload", err)
	}
	graph, err := NewRepository(executor).GetByID(payload.GraphID)
	if err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.get_existing_graph", err)
	}
	if graph == nil {
		return nil, false, apperrors.New(apperrors.CodePersistence, "task_graph_service.get_existing_graph", "existing graph event has no graph projection")
	}
	workUnits, err := s.newWorkUnitLister(executor).ListByTaskGraph(graph.ID)
	if err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodePersistence, "task_graph_service.list_existing_work_units", err)
	}
	return &TaskGraphDecomposeResult{
		Graph:     graph,
		WorkUnits: workUnits,
		Event:     existing,
		Duplicate: true,
	}, true, nil
}
