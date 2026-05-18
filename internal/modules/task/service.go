// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
)

type TaskService struct {
	db       *sql.DB
	onCancel func(ctx context.Context, tx *sql.Tx, taskID string, input transition.TransitionInput) error
}

type CreateTaskInput struct {
	ID                   string
	EventID              string
	Title                string
	Description          string
	Priority             Priority
	RiskLevel            RiskLevel
	CreatedFromMessageID string
	AcceptanceCriteria   []string
}

func NewTaskService(database *sql.DB, onCancel func(ctx context.Context, tx *sql.Tx, taskID string, input transition.TransitionInput) error) *TaskService {
	return &TaskService{db: database, onCancel: onCancel}
}

func (s *TaskService) Create(ctx context.Context, input CreateTaskInput) (*transition.OperationResult[*Task], error) {
	if input.Priority == "" {
		input.Priority = PriorityP2
	}
	if input.RiskLevel == "" {
		input.RiskLevel = RiskLevelLow
	}
	if err := ValidateCreateTaskInput(input); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "task_service.begin_create")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	now := time.Now().UTC()
	task := &Task{
		ID:                   input.ID,
		Title:                input.Title,
		Description:          input.Description,
		Status:               StatusCreated,
		Priority:             input.Priority,
		RiskLevel:            input.RiskLevel,
		CreatedFromMessageID: input.CreatedFromMessageID,
		AcceptanceCriteria:   input.AcceptanceCriteria,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	repo := NewRepository(tx)
	if err := repo.Create(task); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task_service.create_projection", err)
	}

	payload, err := serialization.MarshalPayload("task_service.create_payload", map[string]interface{}{
		"task_id":             task.ID,
		"title":               task.Title,
		"description":         task.Description,
		"status":              task.Status,
		"priority":            task.Priority,
		"risk_level":          task.RiskLevel,
		"acceptance_criteria": task.AcceptanceCriteria,
	})
	if err != nil {
		return nil, err
	}
	appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "task.created",
		Version:     transition.EventVersionV1,
		TaskID:      task.ID,
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}

	if err := dbcore.CommitTx(tx, "task_service.commit_create"); err != nil {
		return nil, err
	}

	return &transition.OperationResult[*Task]{Value: task, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *TaskService) Triage(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	return s.transition(ctx, taskID, StatusTriaged, input)
}

func (s *TaskService) Plan(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	return s.transition(ctx, taskID, StatusPlanned, input)
}

func (s *TaskService) Schedule(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	return s.transition(ctx, taskID, StatusScheduled, input)
}

func (s *TaskService) Pause(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	return s.transition(ctx, taskID, StatusPaused, input)
}

func (s *TaskService) Resume(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	return s.transition(ctx, taskID, StatusRunning, input)
}

func (s *TaskService) Start(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	return s.transition(ctx, taskID, StatusRunning, input)
}

func (s *TaskService) Validate(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	return s.transition(ctx, taskID, StatusValidating, input)
}

func (s *TaskService) Complete(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	return s.transition(ctx, taskID, StatusCompleted, input)
}

func (s *TaskService) Fail(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	return s.transition(ctx, taskID, StatusFailed, input)
}

func (s *TaskService) Cancel(ctx context.Context, taskID string, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	return s.transition(ctx, taskID, StatusCancelled, input)
}

func (s *TaskService) transition(ctx context.Context, taskID string, target Status, input transition.TransitionInput) (*transition.OperationResult[*Task], error) {
	op := "task_service.transition"
	if err := validation.RequiredUUID(taskID, "task_id", op); err != nil {
		return nil, err
	}
	if err := transition.RequireFinalAudit(string(target), input, op); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "task_service.begin_transition")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	task, err := RequireByID(ctx, tx, taskID)
	if err != nil {
		return nil, err
	}
	if err := statemachine.CanTransition(statemachine.AggregateTask, string(task.Status), string(target), transition.TransitionContext(input)); err != nil {
		return nil, err
	}
	if target == StatusCancelled {
		if s.onCancel != nil {
			if err := s.onCancel(ctx, tx, task.ID, input); err != nil {
				return nil, err
			}
		}
	}

	event, duplicate, err := transition.AppendTransition(ctx, tx, input.EventID, EventTypeForStatus(target), task.ID, "", "", input.AgentID, transition.TransitionPayload(task.Status, target, input))
	if err != nil {
		return nil, err
	}
	if !duplicate {
		now := time.Now().UTC()
		acceptanceCriteria, err := json.Marshal(task.AcceptanceCriteria)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodeValidation, "task_service.marshal_acceptance_criteria", err)
		}
		res, err := tx.ExecContext(ctx, QueryUpdate, task.ID, task.Title, task.Description, target, task.Priority, task.RiskLevel, acceptanceCriteria, now)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "task_service.update_projection", err)
		}
		if err := dbcore.EnsureRowsAffected(res, "task", "task_service.update_projection"); err != nil {
			return nil, err
		}
		task.Status = target
		task.UpdatedAt = now
	}

	if err := dbcore.CommitTx(tx, "task_service.commit_transition"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*Task]{Value: task, Event: event, Duplicate: duplicate}, nil
}
