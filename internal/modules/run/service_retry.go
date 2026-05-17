// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package run

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func (s *RunService) Retry(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*Run], error) {
	op := "run_service.retry"
	if err := validation.RequiredUUID(runID, "run_id", op); err != nil {
		return nil, err
	}
	if err := validation.RequiredUUID(input.EventID, "event_id", op); err != nil {
		return nil, err
	}
	if input.Justification == "" && input.FailureReason == "" {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "retry requires justification or failure reason")
	}
	policy, err := RetryPolicyFromInput(input.Extra, op)
	if err != nil {
		return nil, err
	}
	operationCtx, operationCancel := context.WithTimeout(ctx, policy.OperationTimeout)
	defer operationCancel()
	attemptCtx, attemptCancel := context.WithTimeout(operationCtx, policy.AttemptTimeout)
	defer attemptCancel()

	tx, err := dbcore.BeginTx(attemptCtx, s.db, "run_service.begin_retry")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	if err := dbcore.AcquireAdvisoryTxLock(attemptCtx, tx, "run_retry:"+input.EventID, "run_service.retry_lock"); err != nil {
		return nil, err
	}
	previous, err := RequireByID(attemptCtx, tx, runID)
	if err != nil {
		return nil, err
	}
	if existing, duplicate, err := existingRetryResult(attemptCtx, tx, input.EventID, previous.ID); err != nil {
		return nil, err
	} else if duplicate {
		if err := dbcore.CommitTx(tx, "run_service.commit_retry_duplicate"); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if previous.Status != StatusFailed && previous.Status != StatusCancelled {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "only failed or cancelled runs can be retried")
	}
	if previous.Attempt >= policy.MaxAttempts {
		return nil, apperrors.New(apperrors.CodePolicy, op, "retry limit reached")
	}

	nextAttempt := previous.Attempt + 1
	backoffDelay := policy.BackoffDelayForAttempt(nextAttempt)
	if err := WaitForRetryBackoff(attemptCtx, backoffDelay); err != nil {
		return nil, err
	}

	next := &Run{
		ID:         uuid.NewSHA1(uuid.NameSpaceURL, []byte("orchestraos:run_retry:"+input.EventID)).String(),
		TaskID:     previous.TaskID,
		WorkUnitID: previous.WorkUnitID,
		Status:     StatusCreated,
		Attempt:    nextAttempt,
	}
	if err := NewRepository(tx).Create(next); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "run_service.retry_create_projection", err)
	}
	payload, err := serialization.MarshalPayload("run_service.retry_payload", map[string]interface{}{
		"run_id":          next.ID,
		"retry_of":        previous.ID,
		"task_id":         next.TaskID,
		"work_unit_id":    next.WorkUnitID,
		"attempt":         next.Attempt,
		"idempotency_key": input.EventID,
		"retry_policy":    policy.Payload(backoffDelay),
		"justification":   input.Justification,
		"failure_reason":  input.FailureReason,
	})
	if err != nil {
		return nil, err
	}
	appendResult, err := transition.AppendServiceEvent(attemptCtx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "run.created",
		Version:     transition.EventVersionV1,
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
	if err := dbcore.CommitTx(tx, "run_service.commit_retry"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*Run]{Value: next, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func existingRetryResult(ctx context.Context, tx *sql.Tx, eventID, previousRunID string) (*transition.OperationResult[*Run], bool, error) {
	_ = ctx
	store, err := eventstore.NewStoreWithExecutor(tx)
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
	run, err := RequireByID(ctx, tx, payload.RunID)
	if err != nil {
		return nil, false, err
	}
	return &transition.OperationResult[*Run]{Value: run, Event: existing, Duplicate: true}, true, nil
}
