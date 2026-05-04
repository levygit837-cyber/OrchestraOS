package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
	"github.com/levygit837-cyber/OrchestraOS/internal/statemachine"
)

type TaskService struct {
	db *sql.DB
}

type CreateTaskInput struct {
	ID                   string
	EventID              string
	Title                string
	Description          string
	Priority             domain.Priority
	RiskLevel            domain.RiskLevel
	CreatedFromMessageID string
	AcceptanceCriteria   []string
}

func NewTaskService(database *sql.DB) *TaskService {
	return &TaskService{db: database}
}

func (s *TaskService) Create(ctx context.Context, input CreateTaskInput) (*OperationResult[*domain.Task], error) {
	if input.Priority == "" {
		input.Priority = domain.PriorityP2
	}
	if input.RiskLevel == "" {
		input.RiskLevel = domain.RiskLevelLow
	}
	if err := validateCreateTaskInput(input); err != nil {
		return nil, err
	}

	tx, err := beginTx(ctx, s.db, "task_service.begin_create")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	now := time.Now().UTC()
	task := &domain.Task{
		ID:                   input.ID,
		Title:                input.Title,
		Description:          input.Description,
		Status:               domain.TaskStatusCreated,
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

	repo := repository.NewTaskRepository(tx)
	if err := repo.Create(task); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "task_service.create_projection", err)
	}

	payload, err := marshalPayload("task_service.create_payload", map[string]interface{}{
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
	appendResult, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "task.created",
		Version:     eventVersionV1,
		TaskID:      task.ID,
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}

	if err := commitTx(tx, "task_service.commit_create"); err != nil {
		return nil, err
	}

	return &OperationResult[*domain.Task]{Value: task, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *TaskService) Triage(ctx context.Context, taskID string, input TransitionInput) (*OperationResult[*domain.Task], error) {
	return s.transition(ctx, taskID, domain.TaskStatusTriaged, input)
}

func (s *TaskService) Plan(ctx context.Context, taskID string, input TransitionInput) (*OperationResult[*domain.Task], error) {
	return s.transition(ctx, taskID, domain.TaskStatusPlanned, input)
}

func (s *TaskService) Schedule(ctx context.Context, taskID string, input TransitionInput) (*OperationResult[*domain.Task], error) {
	return s.transition(ctx, taskID, domain.TaskStatusScheduled, input)
}

func (s *TaskService) Pause(ctx context.Context, taskID string, input TransitionInput) (*OperationResult[*domain.Task], error) {
	return s.transition(ctx, taskID, domain.TaskStatusPaused, input)
}

func (s *TaskService) Resume(ctx context.Context, taskID string, input TransitionInput) (*OperationResult[*domain.Task], error) {
	return s.transition(ctx, taskID, domain.TaskStatusRunning, input)
}

func (s *TaskService) Start(ctx context.Context, taskID string, input TransitionInput) (*OperationResult[*domain.Task], error) {
	return s.transition(ctx, taskID, domain.TaskStatusRunning, input)
}

func (s *TaskService) Validate(ctx context.Context, taskID string, input TransitionInput) (*OperationResult[*domain.Task], error) {
	return s.transition(ctx, taskID, domain.TaskStatusValidating, input)
}

func (s *TaskService) Complete(ctx context.Context, taskID string, input TransitionInput) (*OperationResult[*domain.Task], error) {
	return s.transition(ctx, taskID, domain.TaskStatusCompleted, input)
}

func (s *TaskService) Fail(ctx context.Context, taskID string, input TransitionInput) (*OperationResult[*domain.Task], error) {
	return s.transition(ctx, taskID, domain.TaskStatusFailed, input)
}

func (s *TaskService) Cancel(ctx context.Context, taskID string, input TransitionInput) (*OperationResult[*domain.Task], error) {
	return s.transition(ctx, taskID, domain.TaskStatusCancelled, input)
}

func (s *TaskService) transition(ctx context.Context, taskID string, target domain.TaskStatus, input TransitionInput) (*OperationResult[*domain.Task], error) {
	op := "task_service.transition"
	if err := validateRequiredUUID(taskID, "task_id", op); err != nil {
		return nil, err
	}
	if err := requireFinalAudit(string(target), input, op); err != nil {
		return nil, err
	}

	tx, err := beginTx(ctx, s.db, "task_service.begin_transition")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	task, err := getTask(ctx, tx, taskID)
	if err != nil {
		return nil, err
	}
	if err := statemachine.CanTransition(statemachine.AggregateTask, string(task.Status), string(target), transitionContext(input)); err != nil {
		return nil, err
	}
	if target == domain.TaskStatusCancelled {
		if err := cancelTaskDependents(ctx, tx, task.ID, input); err != nil {
			return nil, err
		}
	}

	event, duplicate, err := appendTransition(ctx, tx, input.EventID, eventTypeForTaskStatus(target), task.ID, "", "", input.AgentID, transitionPayload(task.Status, target, input))
	if err != nil {
		return nil, err
	}
	if !duplicate {
		now := time.Now().UTC()
		acceptanceCriteria, err := json.Marshal(task.AcceptanceCriteria)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodeValidation, "task_service.marshal_acceptance_criteria", err)
		}
		res, err := tx.ExecContext(ctx, db.QueryTaskUpdate, task.ID, task.Title, task.Description, target, task.Priority, task.RiskLevel, acceptanceCriteria, now)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "task_service.update_projection", err)
		}
		if err := ensureRowsAffected(res, "task", "task_service.update_projection"); err != nil {
			return nil, err
		}
		task.Status = target
		task.UpdatedAt = now
	}

	if err := commitTx(tx, "task_service.commit_transition"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.Task]{Value: task, Event: event, Duplicate: duplicate}, nil
}

func cancelTaskDependents(ctx context.Context, tx *sql.Tx, taskID string, input TransitionInput) error {
	runRepo := repository.NewRunRepository(tx)
	runs, err := runRepo.ListByTask(taskID)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "task_service.list_runs_for_cancel", err)
	}
	for _, run := range runs {
		if isFinalStatus(string(run.Status)) {
			continue
		}
		if err := statemachine.CanTransition(statemachine.AggregateRun, string(run.Status), string(domain.RunStatusCancelled), transitionContext(input)); err != nil {
			return err
		}
		if _, _, err := appendTransition(ctx, tx, "", "run.cancelled", taskID, run.ID, run.WorkUnitID, input.AgentID, transitionPayload(run.Status, domain.RunStatusCancelled, input)); err != nil {
			return err
		}
		if err := updateRunProjection(ctx, tx, run.ID, domain.RunStatusCancelled, runResultForStatus(domain.RunStatusCancelled), nil); err != nil {
			return err
		}
	}

	wuRepo := repository.NewWorkUnitRepository(tx)
	workUnits, err := wuRepo.ListByTask(taskID)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "task_service.list_work_units_for_cancel", err)
	}
	for _, wu := range workUnits {
		if isFinalStatus(string(wu.Status)) {
			continue
		}
		if err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(wu.Status), string(domain.WorkUnitStatusCancelled), transitionContext(input)); err != nil {
			return err
		}
		if _, _, err := appendTransition(ctx, tx, "", "work_unit.cancelled", taskID, "", wu.ID, input.AgentID, transitionPayload(wu.Status, domain.WorkUnitStatusCancelled, input)); err != nil {
			return err
		}
		res, err := tx.ExecContext(ctx, db.QueryWorkUnitUpdateStatus, wu.ID, domain.WorkUnitStatusCancelled, time.Now().UTC())
		if err != nil {
			return apperrors.Wrap(apperrors.CodePersistence, "task_service.cancel_work_unit_projection", err)
		}
		if err := ensureRowsAffected(res, "work unit", "task_service.cancel_work_unit_projection"); err != nil {
			return err
		}
	}
	return nil
}

func validateCreateTaskInput(input CreateTaskInput) error {
	op := "task_service.validate_create"
	if err := validateOptionalUUID(input.ID, "task_id", op); err != nil {
		return err
	}
	if err := validateOptionalUUID(input.EventID, "event_id", op); err != nil {
		return err
	}
	if err := validateRequiredText(input.Title, "title", op); err != nil {
		return err
	}
	if err := validatePriority(input.Priority, op); err != nil {
		return err
	}
	if err := validateRiskLevel(input.RiskLevel, op); err != nil {
		return err
	}
	if err := validateStringList(input.AcceptanceCriteria, "acceptance_criteria", op, false); err != nil {
		return err
	}
	return nil
}
