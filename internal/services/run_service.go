package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
	"github.com/levygit837-cyber/OrchestraOS/internal/statemachine"
)

type RunService struct {
	db *sql.DB
}

type CreateRunInput struct {
	ID         string
	EventID    string
	TaskID     string
	WorkUnitID string
	Attempt    int
}

func NewRunService(database *sql.DB) *RunService {
	return &RunService{db: database}
}

func (s *RunService) Create(ctx context.Context, input CreateRunInput) (*OperationResult[*domain.Run], error) {
	if input.ID == "" {
		input.ID = uuid.New().String()
	}
	if input.Attempt == 0 {
		input.Attempt = 1
	}
	if err := validateCreateRunInput(input); err != nil {
		return nil, err
	}

	tx, err := beginTx(ctx, s.db, "run_service.begin_create")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	task, err := getTask(ctx, tx, input.TaskID)
	if err != nil {
		return nil, err
	}
	wu, err := getWorkUnit(ctx, tx, input.WorkUnitID)
	if err != nil {
		return nil, err
	}
	if wu.TaskGraphID != task.ID {
		return nil, apperrors.New(apperrors.CodeInvalidInput, "run_service.validate_refs", "work_unit_id does not belong to task_id")
	}

	run := &domain.Run{
		ID:         input.ID,
		TaskID:     task.ID,
		WorkUnitID: wu.ID,
		Status:     domain.RunStatusCreated,
		Attempt:    input.Attempt,
	}
	if err := repository.NewRunRepository(tx).Create(run); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "run_service.create_projection", err)
	}

	payload, err := marshalPayload("run_service.create_payload", map[string]interface{}{
		"run_id":       run.ID,
		"task_id":      run.TaskID,
		"work_unit_id": run.WorkUnitID,
		"status":       run.Status,
		"attempt":      run.Attempt,
	})
	if err != nil {
		return nil, err
	}
	appendResult, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "run.created",
		Version:     eventVersionV1,
		TaskID:      run.TaskID,
		RunID:       run.ID,
		WorkUnitID:  run.WorkUnitID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}
	if err := commitTx(tx, "run_service.commit_create"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.Run]{Value: run, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *RunService) Start(ctx context.Context, runID string, input TransitionInput) (*OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusRunning, input, true)
}

func (s *RunService) Resume(ctx context.Context, runID string, input TransitionInput) (*OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusRunning, input, false)
}

func (s *RunService) Validate(ctx context.Context, runID string, input TransitionInput) (*OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusValidating, input, true)
}

func (s *RunService) Complete(ctx context.Context, runID string, input TransitionInput) (*OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusCompleted, input, true)
}

func (s *RunService) Fail(ctx context.Context, runID string, input TransitionInput) (*OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusFailed, input, true)
}

func (s *RunService) Cancel(ctx context.Context, runID string, input TransitionInput) (*OperationResult[*domain.Run], error) {
	return s.transition(ctx, runID, domain.RunStatusCancelled, input, true)
}

func (s *RunService) Timeout(ctx context.Context, runID string, input TransitionInput) (*OperationResult[*domain.Run], error) {
	if input.FailureReason == "" {
		input.FailureReason = "run timed out"
	}
	if input.Justification == "" {
		input.Justification = "timeout reached before run completed"
	}
	if input.Extra == nil {
		input.Extra = map[string]interface{}{}
	}
	input.Extra["timeout"] = true
	return s.transition(ctx, runID, domain.RunStatusFailed, input, true)
}

func (s *RunService) Retry(ctx context.Context, runID string, input TransitionInput) (*OperationResult[*domain.Run], error) {
	op := "run_service.retry"
	if err := validateRequiredUUID(runID, "run_id", op); err != nil {
		return nil, err
	}
	if err := validateRequiredUUID(input.EventID, "event_id", op); err != nil {
		return nil, err
	}
	if input.Justification == "" && input.FailureReason == "" {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "retry requires justification or failure reason")
	}
	policy, err := retryPolicyFromInput(input, op)
	if err != nil {
		return nil, err
	}
	operationCtx, operationCancel := context.WithTimeout(ctx, policy.OperationTimeout)
	defer operationCancel()
	attemptCtx, attemptCancel := context.WithTimeout(operationCtx, policy.AttemptTimeout)
	defer attemptCancel()

	tx, err := beginTx(attemptCtx, s.db, "run_service.begin_retry")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	if err := acquireAdvisoryTxLock(attemptCtx, tx, "run_retry:"+input.EventID, "run_service.retry_lock"); err != nil {
		return nil, err
	}
	previous, err := getRun(attemptCtx, tx, runID)
	if err != nil {
		return nil, err
	}
	if existing, duplicate, err := existingRetryResult(attemptCtx, tx, input.EventID, previous.ID); err != nil {
		return nil, err
	} else if duplicate {
		if err := commitTx(tx, "run_service.commit_retry_duplicate"); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if previous.Status != domain.RunStatusFailed && previous.Status != domain.RunStatusCancelled {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "only failed or cancelled runs can be retried")
	}
	if previous.Attempt >= policy.MaxAttempts {
		return nil, apperrors.New(apperrors.CodePolicy, op, "retry limit reached")
	}

	nextAttempt := previous.Attempt + 1
	backoffDelay := policy.backoffDelayForAttempt(nextAttempt)
	if err := waitForRetryBackoff(attemptCtx, backoffDelay); err != nil {
		return nil, err
	}

	next := &domain.Run{
		ID:         uuid.NewSHA1(uuid.NameSpaceURL, []byte("orchestraos:run_retry:"+input.EventID)).String(),
		TaskID:     previous.TaskID,
		WorkUnitID: previous.WorkUnitID,
		Status:     domain.RunStatusCreated,
		Attempt:    nextAttempt,
	}
	if err := repository.NewRunRepository(tx).Create(next); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "run_service.retry_create_projection", err)
	}
	payload, err := marshalPayload("run_service.retry_payload", map[string]interface{}{
		"run_id":          next.ID,
		"retry_of":        previous.ID,
		"task_id":         next.TaskID,
		"work_unit_id":    next.WorkUnitID,
		"attempt":         next.Attempt,
		"idempotency_key": input.EventID,
		"retry_policy":    policy.payload(backoffDelay),
		"justification":   input.Justification,
		"failure_reason":  input.FailureReason,
	})
	if err != nil {
		return nil, err
	}
	appendResult, err := appendServiceEvent(attemptCtx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "run.created",
		Version:     eventVersionV1,
		TaskID:      next.TaskID,
		RunID:       next.ID,
		WorkUnitID:  next.WorkUnitID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}
	if err := commitTx(tx, "run_service.commit_retry"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.Run]{Value: next, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func existingRetryResult(ctx context.Context, tx *sql.Tx, eventID, previousRunID string) (*OperationResult[*domain.Run], bool, error) {
	_ = ctx
	store, err := newEventStore(tx)
	if err != nil {
		return nil, false, err
	}
	existing, err := store.Get(eventID)
	if err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodePersistence, "run_service.retry_get_existing_event", err)
	}
	if existing == nil {
		return nil, false, nil
	}
	var payload struct {
		RunID   string `json:"run_id"`
		RetryOf string `json:"retry_of"`
	}
	if err := json.Unmarshal(existing.Payload, &payload); err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodeValidation, "run_service.retry_existing_payload", err)
	}
	if existing.Type != "run.created" || payload.RetryOf != previousRunID || payload.RunID == "" {
		return nil, false, apperrors.New(apperrors.CodeConflict, "run_service.retry_idempotency", "event_id already exists for a different retry operation")
	}
	run, err := getRun(ctx, tx, payload.RunID)
	if err != nil {
		return nil, false, err
	}
	return &OperationResult[*domain.Run]{Value: run, Event: existing, Duplicate: true}, true, nil
}

func (s *RunService) transition(ctx context.Context, runID string, target domain.RunStatus, input TransitionInput, updateWorkUnit bool) (*OperationResult[*domain.Run], error) {
	op := "run_service.transition"
	if err := validateRequiredUUID(runID, "run_id", op); err != nil {
		return nil, err
	}
	if err := validateRuntime(input.Runtime, op); err != nil {
		return nil, err
	}
	if err := requireFinalAudit(string(target), input, op); err != nil {
		return nil, err
	}

	tx, err := beginTx(ctx, s.db, "run_service.begin_transition")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	run, err := getRun(ctx, tx, runID)
	if err != nil {
		return nil, err
	}
	task, err := getTask(ctx, tx, run.TaskID)
	if err != nil {
		return nil, err
	}
	if target == domain.RunStatusRunning {
		if err := validateRunStartPolicy(task, input); err != nil {
			return nil, err
		}
	}
	if err := statemachine.CanTransition(statemachine.AggregateRun, string(run.Status), string(target), transitionContext(input)); err != nil {
		return nil, err
	}

	if updateWorkUnit && run.WorkUnitID != "" {
		if err := transitionRelatedWorkUnit(ctx, tx, run, target, input); err != nil {
			return nil, err
		}
	}

	payload := transitionPayload(run.Status, target, input)
	payload["attempt"] = run.Attempt
	payload["run_id"] = run.ID
	if target == domain.RunStatusCompleted {
		payload["result"] = domain.RunResultSucceeded
	}
	event, duplicate, err := appendTransition(ctx, tx, input.EventID, eventTypeForRunStatus(target), run.TaskID, run.ID, run.WorkUnitID, input.AgentID, payload)
	if err != nil {
		return nil, err
	}
	if !duplicate {
		result := runResultForStatus(target)
		var failureReason *string
		if input.FailureReason != "" {
			failureReason = &input.FailureReason
		}
		if err := updateRunProjection(ctx, tx, run.ID, target, result, failureReason); err != nil {
			return nil, err
		}
		run.Status = target
		if target == domain.RunStatusRunning && run.StartedAt.IsZero() {
			run.StartedAt = time.Now().UTC()
		}
	}
	if err := commitTx(tx, "run_service.commit_transition"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.Run]{Value: run, Event: event, Duplicate: duplicate}, nil
}

func transitionRelatedWorkUnit(ctx context.Context, tx *sql.Tx, run *domain.Run, target domain.RunStatus, input TransitionInput) error {
	wu, err := getWorkUnit(ctx, tx, run.WorkUnitID)
	if err != nil {
		return err
	}
	var wuTarget domain.WorkUnitStatus
	switch target {
	case domain.RunStatusRunning:
		wuTarget = domain.WorkUnitStatusRunning
	case domain.RunStatusValidating:
		wuTarget = domain.WorkUnitStatusValidating
	case domain.RunStatusCompleted:
		wuTarget = domain.WorkUnitStatusCompleted
	case domain.RunStatusFailed:
		wuTarget = domain.WorkUnitStatusFailed
	case domain.RunStatusCancelled:
		wuTarget = domain.WorkUnitStatusCancelled
	default:
		return nil
	}
	if wu.Status == wuTarget {
		return nil
	}
	if wuTarget == domain.WorkUnitStatusRunning {
		if err := acquireAdvisoryTxLock(ctx, tx, "work_unit_paths:"+wu.TaskGraphID, "run_service.work_unit_path_lock"); err != nil {
			return err
		}
		if err := validateDependenciesCompleted(ctx, tx, wu); err != nil {
			return err
		}
		if err := validateOwnedPathAvailability(ctx, tx, wu); err != nil {
			return err
		}
	}
	if wuTarget == domain.WorkUnitStatusCompleted && len(wu.AcceptanceCriteria) == 0 && input.Justification == "" {
		return apperrors.New(apperrors.CodeInvalidInput, "run_service.work_unit_completion", "related work unit completion requires acceptance criteria or explicit justification")
	}
	if err := statemachine.CanTransition(statemachine.AggregateWorkUnit, string(wu.Status), string(wuTarget), transitionContext(input)); err != nil {
		if wuTarget == domain.WorkUnitStatusFailed && wu.Status == domain.WorkUnitStatusCreated {
			return nil
		}
		return err
	}
	if _, _, err := appendTransition(ctx, tx, "", eventTypeForWorkUnitStatus(wuTarget), run.TaskID, run.ID, wu.ID, input.AgentID, transitionPayload(wu.Status, wuTarget, input)); err != nil {
		return err
	}
	res, err := tx.ExecContext(ctx, `UPDATE work_units SET status = $2, updated_at = $3 WHERE id = $1`, wu.ID, wuTarget, time.Now().UTC())
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "run_service.update_work_unit_projection", err)
	}
	return ensureRowsAffected(res, "work unit", "run_service.update_work_unit_projection")
}

func validateCreateRunInput(input CreateRunInput) error {
	op := "run_service.validate_create"
	if err := validateRequiredUUID(input.ID, "run_id", op); err != nil {
		return err
	}
	if err := validateOptionalUUID(input.EventID, "event_id", op); err != nil {
		return err
	}
	if err := validateRequiredUUID(input.TaskID, "task_id", op); err != nil {
		return err
	}
	if err := validateRequiredUUID(input.WorkUnitID, "work_unit_id", op); err != nil {
		return err
	}
	if input.Attempt < 1 {
		return apperrors.New(apperrors.CodeInvalidInput, op, "attempt must be greater than zero")
	}
	return nil
}

func validateRuntime(runtime, op string) error {
	if runtime == "" {
		return nil
	}
	switch domain.AgentRuntimeType(runtime) {
	case domain.AgentRuntimeTypeFake, domain.AgentRuntimeTypeCodexCLI, domain.AgentRuntimeTypeExternal:
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("invalid runtime %q", runtime))
	}
}

func validateRunStartPolicy(task *domain.Task, input TransitionInput) error {
	if task.RiskLevel == domain.RiskLevelHigh || task.RiskLevel == domain.RiskLevelCritical {
		if input.Justification == "" {
			return apperrors.New(apperrors.CodePolicy, "run_service.policy", "starting high or critical risk tasks requires explicit justification")
		}
	}
	return nil
}
