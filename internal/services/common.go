package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
	"github.com/levygit837-cyber/OrchestraOS/internal/statemachine"
)

const eventVersionV1 = "v1"

type TransitionInput struct {
	EventID           string
	AgentID           string
	Runtime           string
	EvidenceRefs      []string
	ValidationEventID string
	Justification     string
	FailureReason     string
	Extra             map[string]interface{}
}

type OperationResult[T any] struct {
	Value     T
	Event     *domain.EventEnvelope
	Duplicate bool
}

type retryPolicy struct {
	MaxAttempts       int
	AttemptTimeout    time.Duration
	OperationTimeout  time.Duration
	InitialBackoff    time.Duration
	BackoffMultiplier int
}

func defaultRetryPolicy() retryPolicy {
	return retryPolicy{
		MaxAttempts:       3,
		AttemptTimeout:    5 * time.Second,
		OperationTimeout:  30 * time.Second,
		InitialBackoff:    100 * time.Millisecond,
		BackoffMultiplier: 2,
	}
}

func retryPolicyFromInput(input TransitionInput, op string) (retryPolicy, error) {
	policy := defaultRetryPolicy()
	if input.Extra == nil {
		return policy, nil
	}
	if value, ok, err := intExtra(input.Extra, "max_attempts"); err != nil {
		return policy, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	} else if ok {
		policy.MaxAttempts = value
	}
	if value, ok, err := intExtra(input.Extra, "attempt_timeout_seconds"); err != nil {
		return policy, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	} else if ok {
		policy.AttemptTimeout = time.Duration(value) * time.Second
	}
	if value, ok, err := intExtra(input.Extra, "operation_timeout_seconds"); err != nil {
		return policy, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	} else if ok {
		policy.OperationTimeout = time.Duration(value) * time.Second
	}
	if value, ok, err := intExtra(input.Extra, "initial_backoff_millis"); err != nil {
		return policy, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	} else if ok {
		policy.InitialBackoff = time.Duration(value) * time.Millisecond
	}
	if value, ok, err := intExtra(input.Extra, "backoff_multiplier"); err != nil {
		return policy, apperrors.Wrap(apperrors.CodeInvalidInput, op, err)
	} else if ok {
		policy.BackoffMultiplier = value
	}
	if policy.MaxAttempts < 1 {
		return policy, apperrors.New(apperrors.CodeInvalidInput, op, "max_attempts must be greater than zero")
	}
	if policy.AttemptTimeout <= 0 {
		return policy, apperrors.New(apperrors.CodeInvalidInput, op, "attempt_timeout_seconds must be greater than zero")
	}
	if policy.OperationTimeout <= 0 {
		return policy, apperrors.New(apperrors.CodeInvalidInput, op, "operation_timeout_seconds must be greater than zero")
	}
	if policy.InitialBackoff < 0 {
		return policy, apperrors.New(apperrors.CodeInvalidInput, op, "initial_backoff_millis must not be negative")
	}
	if policy.BackoffMultiplier < 1 {
		return policy, apperrors.New(apperrors.CodeInvalidInput, op, "backoff_multiplier must be greater than zero")
	}
	return policy, nil
}

func intExtra(extra map[string]interface{}, key string) (int, bool, error) {
	raw, ok := extra[key]
	if !ok {
		return 0, false, nil
	}
	switch value := raw.(type) {
	case int:
		return value, true, nil
	case int8:
		return int(value), true, nil
	case int16:
		return int(value), true, nil
	case int32:
		return int(value), true, nil
	case int64:
		return int(value), true, nil
	case uint:
		return int(value), true, nil
	case uint8:
		return int(value), true, nil
	case uint16:
		return int(value), true, nil
	case uint32:
		return int(value), true, nil
	case uint64:
		return int(value), true, nil
	case float64:
		converted := int(value)
		if value != float64(converted) {
			return 0, true, fmt.Errorf("%s must be an integer", key)
		}
		return converted, true, nil
	case json.Number:
		converted, err := strconv.Atoi(value.String())
		if err != nil {
			return 0, true, fmt.Errorf("%s must be an integer: %w", key, err)
		}
		return converted, true, nil
	default:
		return 0, true, fmt.Errorf("%s must be an integer", key)
	}
}

func (p retryPolicy) backoffDelayForAttempt(attempt int) time.Duration {
	if attempt <= 1 || p.InitialBackoff == 0 {
		return 0
	}
	delay := p.InitialBackoff
	for i := 2; i < attempt; i++ {
		delay *= time.Duration(p.BackoffMultiplier)
	}
	return delay
}

func (p retryPolicy) payload(delay time.Duration) map[string]interface{} {
	return map[string]interface{}{
		"max_attempts":              p.MaxAttempts,
		"attempt_timeout_seconds":   int(p.AttemptTimeout / time.Second),
		"operation_timeout_seconds": int(p.OperationTimeout / time.Second),
		"initial_backoff_millis":    int(p.InitialBackoff / time.Millisecond),
		"backoff_multiplier":        p.BackoffMultiplier,
		"applied_backoff_millis":    int(delay / time.Millisecond),
	}
}

func waitForRetryBackoff(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return apperrors.Wrap(apperrors.CodeTimeout, "services.retry_backoff", ctx.Err())
	case <-timer.C:
		return nil
	}
}

func beginTx(ctx context.Context, database *sql.DB, op string) (*sql.Tx, error) {
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	return tx, nil
}

func commitTx(tx *sql.Tx, op string) error {
	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	return nil
}

func rollbackTx(tx *sql.Tx) {
	_ = tx.Rollback()
}

func acquireAdvisoryTxLock(ctx context.Context, tx *sql.Tx, key, op string) error {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(key))
	lockID := int64(hasher.Sum64())
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, lockID); err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	return nil
}

func appendServiceEvent(ctx context.Context, tx *sql.Tx, envelope *domain.EventEnvelope) (*AppendResult, error) {
	service := NewEventServiceWithExecutor(tx)
	return service.Append(ctx, envelope)
}

func marshalPayload(op string, payload map[string]interface{}) (json.RawMessage, error) {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, op, err)
	}
	return bytes, nil
}

func validateRequiredUUID(value, field, op string) error {
	if strings.TrimSpace(value) == "" {
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("%s is required", field))
	}
	if _, err := uuid.Parse(value); err != nil {
		return apperrors.Wrap(apperrors.CodeInvalidInput, op, fmt.Errorf("%s must be a UUID: %w", field, err))
	}
	return nil
}

func validateOptionalUUID(value, field, op string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return validateRequiredUUID(value, field, op)
}

func validateRequiredText(value, field, op string) error {
	if strings.TrimSpace(value) == "" {
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("%s is required", field))
	}
	return nil
}

func validateStringList(values []string, field, op string, required bool) error {
	if required && len(values) == 0 {
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("%s is required", field))
	}
	for i, value := range values {
		if strings.TrimSpace(value) == "" {
			return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("%s[%d] must not be empty", field, i))
		}
	}
	return nil
}

func validatePriority(priority domain.Priority, op string) error {
	switch priority {
	case domain.PriorityP0, domain.PriorityP1, domain.PriorityP2, domain.PriorityP3:
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("invalid priority %q", priority))
	}
}

func validateRiskLevel(risk domain.RiskLevel, op string) error {
	switch risk {
	case domain.RiskLevelLow, domain.RiskLevelMedium, domain.RiskLevelHigh, domain.RiskLevelCritical:
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("invalid risk level %q", risk))
	}
}

func transitionContext(input TransitionInput) statemachine.TransitionContext {
	return statemachine.TransitionContext{
		EvidenceRefs:      input.EvidenceRefs,
		ValidationEventID: input.ValidationEventID,
		Justification:     input.Justification,
	}
}

func transitionPayload(from, to interface{}, input TransitionInput) map[string]interface{} {
	payload := map[string]interface{}{
		"from_status": from,
		"to_status":   to,
	}
	if input.Runtime != "" {
		payload["runtime"] = input.Runtime
	}
	if len(input.EvidenceRefs) > 0 {
		payload["evidence_refs"] = input.EvidenceRefs
	}
	if input.ValidationEventID != "" {
		payload["validation_event_id"] = input.ValidationEventID
	}
	if input.Justification != "" {
		payload["justification"] = input.Justification
	}
	if input.FailureReason != "" {
		payload["failure_reason"] = input.FailureReason
	}
	for key, value := range input.Extra {
		payload[key] = value
	}
	return payload
}

func requireFinalAudit(target string, input TransitionInput, op string) error {
	if !isFinalStatus(target) {
		return nil
	}
	if len(input.EvidenceRefs) > 0 || input.ValidationEventID != "" || input.Justification != "" || input.FailureReason != "" {
		return nil
	}
	return apperrors.New(apperrors.CodeInvalidInput, op, "final state requires evidence, validation event, failure reason, or justification")
}

func isFinalStatus(status string) bool {
	switch status {
	case "completed", "failed", "cancelled", "stopped":
		return true
	default:
		return false
	}
}

func ensureRowsAffected(result sql.Result, entity, op string) error {
	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	if rows == 0 {
		return apperrors.New(apperrors.CodeConflict, op, fmt.Sprintf("%s projection was not updated", entity))
	}
	return nil
}

func getTask(ctx context.Context, tx *sql.Tx, id string) (*domain.Task, error) {
	_ = ctx
	repo := repository.NewTaskRepository(tx)
	task, err := repo.GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "services.get_task", err)
	}
	if task == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "services.get_task", "task not found")
	}
	return task, nil
}

func getWorkUnit(ctx context.Context, tx *sql.Tx, id string) (*domain.WorkUnit, error) {
	_ = ctx
	repo := repository.NewWorkUnitRepository(tx)
	wu, err := repo.GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "services.get_work_unit", err)
	}
	if wu == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "services.get_work_unit", "work unit not found")
	}
	return wu, nil
}

func getRun(ctx context.Context, tx *sql.Tx, id string) (*domain.Run, error) {
	_ = ctx
	repo := repository.NewRunRepository(tx)
	run, err := repo.GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "services.get_run", err)
	}
	if run == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "services.get_run", "run not found")
	}
	return run, nil
}

func getAgentSession(ctx context.Context, tx *sql.Tx, id string) (*domain.AgentSession, error) {
	_ = ctx
	repo := repository.NewAgentSessionRepository(tx)
	session, err := repo.GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "services.get_agent_session", err)
	}
	if session == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "services.get_agent_session", "agent session not found")
	}
	return session, nil
}

func appendTransition(ctx context.Context, tx *sql.Tx, eventID, eventType, taskID, runID, workUnitID, agentID string, payload map[string]interface{}) (*domain.EventEnvelope, bool, error) {
	payloadBytes, err := marshalPayload("services.transition_payload", payload)
	if err != nil {
		return nil, false, err
	}
	result, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          eventID,
		Type:        eventType,
		Version:     eventVersionV1,
		TaskID:      taskID,
		RunID:       runID,
		WorkUnitID:  workUnitID,
		AgentID:     agentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payloadBytes,
	})
	if err != nil {
		return nil, false, err
	}
	return &result.Event, result.Duplicate, nil
}

func updateRunProjection(ctx context.Context, tx *sql.Tx, runID string, status domain.RunStatus, result *domain.RunResult, failureReason *string) error {
	now := time.Now().UTC()
	var startedAt, finishedAt *time.Time
	if status == domain.RunStatusRunning {
		startedAt = &now
	}
	if status == domain.RunStatusCompleted || status == domain.RunStatusFailed || status == domain.RunStatusCancelled {
		finishedAt = &now
	}

	var resultStr *string
	if result != nil {
		value := string(*result)
		resultStr = &value
	}

	res, err := tx.ExecContext(ctx, db.QueryRunUpdateStatus, runID, status, startedAt, finishedAt, resultStr, failureReason, now)
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "services.update_run_projection", err)
	}
	return ensureRowsAffected(res, "run", "services.update_run_projection")
}

func runResultForStatus(status domain.RunStatus) *domain.RunResult {
	switch status {
	case domain.RunStatusCompleted:
		result := domain.RunResultSucceeded
		return &result
	case domain.RunStatusFailed:
		result := domain.RunResultFailed
		return &result
	case domain.RunStatusCancelled:
		result := domain.RunResultCancelled
		return &result
	default:
		return nil
	}
}

func eventTypeForTaskStatus(status domain.TaskStatus) string {
	if status == domain.TaskStatusRunning {
		return "task.started"
	}
	return "task." + string(status)
}

func eventTypeForWorkUnitStatus(status domain.WorkUnitStatus) string {
	if status == domain.WorkUnitStatusRunning {
		return "work_unit.started"
	}
	return "work_unit." + string(status)
}

func eventTypeForRunStatus(status domain.RunStatus) string {
	if status == domain.RunStatusRunning {
		return "run.started"
	}
	return "run." + string(status)
}

func eventTypeForAgentSessionStatus(status domain.AgentSessionStatus) string {
	return "agent.session_" + string(status)
}

func newEventStore(executor repository.DBTX) (*eventstore.Store, error) {
	store, err := eventstore.NewStoreWithExecutor(executor)
	if err != nil {
		return nil, err
	}
	return store, nil
}
