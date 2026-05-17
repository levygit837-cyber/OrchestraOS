// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package workunit

import (
	"context"
	"database/sql"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
)

// TaskReader abstracts task reads to avoid cyclic imports.
type TaskReader interface {
	GetByID(id string) (*task.Task, error)
}

// TaskGraphManager abstracts task-graph operations to avoid cyclic imports.
type TaskGraphManager interface {
	GetActiveByTask(taskID string) (*taskgraphmod.TaskGraph, error)
	GetByID(id string) (*taskgraphmod.TaskGraph, error)
	NextVersion(taskID string) (int, error)
	Create(graph *taskgraphmod.TaskGraph) error
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
		res, err := tx.ExecContext(ctx, `UPDATE work_units SET assigned_agent_profile = $2, updated_at = $3 WHERE id = $1`, wu.ID, agentProfile, time.Now().UTC())
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
		res, err := tx.ExecContext(ctx, QueryUpdateStatus, wu.ID, target, time.Now().UTC())
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

func (s *WorkUnitService) requireTaskByID(tx *sql.Tx, id string) (*task.Task, error) {
	task, err := s.newTaskReader(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task.get", err)
	}
	if task == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "task.get", "task not found")
	}
	return task, nil
}
