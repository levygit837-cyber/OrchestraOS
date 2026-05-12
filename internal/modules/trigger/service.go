// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, boundary rules
// Ignoring these files will cause architecture test failures.

package trigger

import (
	"context"
	"database/sql"
	"encoding/json"
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

// RunReader abstracts run reads to avoid cyclic imports.
type RunReader interface {
	GetByID(id string) (*domain.Run, error)
}

// AgentSessionReader abstracts agent-session reads to avoid cyclic imports.
type AgentSessionReader interface {
	GetByID(id string) (*domain.AgentSession, error)
}

// WorkUnitReader abstracts work-unit reads to avoid cyclic imports.
type WorkUnitReader interface {
	GetByID(id string) (*domain.WorkUnit, error)
}

// TriggerService manages trigger lifecycle and anomaly detection.
type TriggerService struct {
	db                    *sql.DB
	newRunReader          func(dbcore.DBTX) RunReader
	newAgentSessionReader func(dbcore.DBTX) AgentSessionReader
	newWorkUnitReader     func(dbcore.DBTX) WorkUnitReader
	newEventStore         func(executor dbcore.DBTX) (*eventstore.Store, error)
	clock                 func() time.Time
	thresholds            domain.ThresholdConfig
}

// CreateTriggerInput holds all fields needed to create a trigger.
type CreateTriggerInput struct {
	ID               string
	EventID          string
	RunID            string
	TaskID           string
	AgentSessionID   string
	TriggerType      domain.TriggerType
	Status           domain.TriggerStatus
	AnomalyType      *domain.AnomalyType
	ThresholdValue   json.RawMessage
	CurrentValue     json.RawMessage
	ResolutionAction *domain.ResolutionAction
}

// NewTriggerService creates a new TriggerService with default thresholds and clock.
func NewTriggerService(
	db *sql.DB,
	newRunReader func(dbcore.DBTX) RunReader,
	newAgentSessionReader func(dbcore.DBTX) AgentSessionReader,
	newWorkUnitReader func(dbcore.DBTX) WorkUnitReader,
) *TriggerService {
	return &TriggerService{
		db:                    db,
		newRunReader:          newRunReader,
		newAgentSessionReader: newAgentSessionReader,
		newWorkUnitReader:     newWorkUnitReader,
		newEventStore:         eventstore.NewStoreWithExecutor,
		clock:                 func() time.Time { return time.Now().UTC() },
		thresholds:            DefaultThresholds(),
	}
}

// SetClock overrides the clock function for testing.
func (s *TriggerService) SetClock(fn func() time.Time) {
	s.clock = fn
}

// SetThresholds overrides the threshold config for testing.
func (s *TriggerService) SetThresholds(cfg domain.ThresholdConfig) {
	s.thresholds = cfg
}

// Create persists a new trigger and emits a domain event.
func (s *TriggerService) Create(ctx context.Context, input CreateTriggerInput) (*transition.OperationResult[*domain.Trigger], error) {
	if input.ID == "" {
		input.ID = uuid.New().String()
	}
	if input.Status == "" {
		input.Status = domain.TriggerStatusActive
	}
	if err := validateCreateTriggerInput(input); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "trigger_service.begin_create")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	trigger := &domain.Trigger{
		ID:               input.ID,
		RunID:            nilIfEmpty(input.RunID),
		TaskID:           nilIfEmpty(input.TaskID),
		AgentSessionID:   nilIfEmpty(input.AgentSessionID),
		TriggerType:      input.TriggerType,
		Status:           input.Status,
		AnomalyType:      input.AnomalyType,
		ThresholdValue:   input.ThresholdValue,
		CurrentValue:     input.CurrentValue,
		ResolutionAction: input.ResolutionAction,
		CreatedAt:        s.clock(),
	}

	if err := NewRepository(tx).Create(trigger); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "trigger_service.create_projection", err)
	}

	payload, err := serialization.MarshalPayload("trigger_service.create_payload", map[string]interface{}{
		"trigger_id": trigger.ID,
		"run_id":     input.RunID,
		"task_id":    input.TaskID,
		"status":     trigger.Status,
		"type":       trigger.TriggerType,
	})
	if err != nil {
		return nil, err
	}

	appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        EventTypeForStatus(trigger.Status),
		Version:     transition.EventVersionV1,
		TaskID:      input.TaskID,
		RunID:       input.RunID,
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}

	if err := dbcore.CommitTx(tx, "trigger_service.commit_create"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.Trigger]{Value: trigger, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

// EvaluateRun evaluates a run for anomalies and persists detected triggers.
func (s *TriggerService) EvaluateRun(ctx context.Context, runID string) ([]*domain.Trigger, error) {
	if err := validation.RequiredUUID(runID, "run_id", "trigger_service.evaluate_run"); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "trigger_service.begin_evaluate_run")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	run, err := s.requireRunByID(tx, runID)
	if err != nil {
		return nil, err
	}

	events, err := s.listRunEvents(tx, runID)
	if err != nil {
		return nil, err
	}

	now := s.clock()
	var detected []*domain.Trigger

	// Stall detection
	lastEventAt := lastEventTime(events)
	if stall := (StallDetector{}).Detect(lastEventAt, run.StartedAt, now, s.thresholds.StallSeconds); stall != nil {
		stall.RunID = &runID
		stall.TaskID = &run.TaskID
		detected = append(detected, stall)
	}

	// Loop detection
	if len(events) > 0 {
		types := eventTypes(events)
		if loop := (LoopDetector{}).Detect(types, s.thresholds.LoopRepetitions, now); loop != nil {
			loop.RunID = &runID
			loop.TaskID = &run.TaskID
			detected = append(detected, loop)
		}
	}

	// Time threshold
	if timeTrigger := (TimeThresholdDetector{}).Detect(run.StartedAt, now, s.thresholds.TimeMaxSeconds); timeTrigger != nil {
		timeTrigger.RunID = &runID
		timeTrigger.TaskID = &run.TaskID
		detected = append(detected, timeTrigger)
	}

	// Steps threshold (count non-trigger events as proxy for steps)
	stepCount := countNonTriggerEvents(events)
	if stepTrigger := (StepsThresholdDetector{}).Detect(stepCount, s.thresholds.StepsMax, now); stepTrigger != nil {
		stepTrigger.RunID = &runID
		stepTrigger.TaskID = &run.TaskID
		detected = append(detected, stepTrigger)
	}

	var result []*domain.Trigger
	for _, t := range detected {
		created, err := s.persistDetectedTrigger(ctx, tx, t)
		if err != nil {
			return nil, err
		}
		result = append(result, created)
	}

	if err := dbcore.CommitTx(tx, "trigger_service.commit_evaluate_run"); err != nil {
		return nil, err
	}
	return result, nil
}

// EvaluateSession evaluates an agent session for heartbeat timeout and stall.
func (s *TriggerService) EvaluateSession(ctx context.Context, sessionID string) ([]*domain.Trigger, error) {
	if err := validation.RequiredUUID(sessionID, "session_id", "trigger_service.evaluate_session"); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "trigger_service.begin_evaluate_session")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	session, err := s.requireSessionByID(tx, sessionID)
	if err != nil {
		return nil, err
	}

	var detected []*domain.Trigger
	now := s.clock()

	// Heartbeat timeout
	if session.LastHeartbeatAt == nil {
		anomaly := domain.AnomalyTypeStall
		detected = append(detected, &domain.Trigger{
			TriggerType: domain.TriggerTypeHeartbeatTimeout,
			Status:      domain.TriggerStatusTriggered,
			AnomalyType: &anomaly,
			AgentSessionID: &sessionID,
			RunID:       &session.RunID,
			TaskID:      &session.TaskID,
			ThresholdValue: mustMarshal(map[string]interface{}{
				"stall_seconds": s.thresholds.StallSeconds,
			}),
			CurrentValue: mustMarshal(map[string]interface{}{
				"last_heartbeat_at": nil,
			}),
			TriggeredAt: &now,
		})
	} else if now.Sub(*session.LastHeartbeatAt) >= time.Duration(s.thresholds.StallSeconds)*time.Second {
		anomaly := domain.AnomalyTypeStall
		detected = append(detected, &domain.Trigger{
			TriggerType: domain.TriggerTypeHeartbeatTimeout,
			Status:      domain.TriggerStatusTriggered,
			AnomalyType: &anomaly,
			AgentSessionID: &sessionID,
			RunID:       &session.RunID,
			TaskID:      &session.TaskID,
			ThresholdValue: mustMarshal(map[string]interface{}{
				"stall_seconds": s.thresholds.StallSeconds,
			}),
			CurrentValue: mustMarshal(map[string]interface{}{
				"seconds_since_heartbeat": int(now.Sub(*session.LastHeartbeatAt).Seconds()),
			}),
			TriggeredAt: &now,
		})
	}

	var result []*domain.Trigger
	for _, t := range detected {
		created, err := s.persistDetectedTrigger(ctx, tx, t)
		if err != nil {
			return nil, err
		}
		result = append(result, created)
	}

	if err := dbcore.CommitTx(tx, "trigger_service.commit_evaluate_session"); err != nil {
		return nil, err
	}
	return result, nil
}

// EvaluateWorkUnit evaluates a work unit for drift and path violations.
func (s *TriggerService) EvaluateWorkUnit(ctx context.Context, workUnitID string) ([]*domain.Trigger, error) {
	if err := validation.RequiredUUID(workUnitID, "work_unit_id", "trigger_service.evaluate_workunit"); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "trigger_service.begin_evaluate_workunit")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	wu, err := s.requireWorkUnitByID(tx, workUnitID)
	if err != nil {
		return nil, err
	}

	events, err := s.listWorkUnitEvents(tx, workUnitID)
	if err != nil {
		return nil, err
	}

	accessedPaths, modifiedPaths := extractPathsFromEvents(events)

	now := s.clock()
	var detected []*domain.Trigger

	if drift := (DriftDetector{}).Detect(wu.OwnedPaths, wu.ReadPaths, accessedPaths, now); drift != nil {
		drift.TaskID = &wu.TaskID
		detected = append(detected, drift)
	}

	if violation := (PathViolationDetector{}).Detect(wu.OwnedPaths, modifiedPaths, now); violation != nil {
		violation.TaskID = &wu.TaskID
		detected = append(detected, violation)
	}

	var result []*domain.Trigger
	for _, t := range detected {
		created, err := s.persistDetectedTrigger(ctx, tx, t)
		if err != nil {
			return nil, err
		}
		result = append(result, created)
	}

	if err := dbcore.CommitTx(tx, "trigger_service.commit_evaluate_workunit"); err != nil {
		return nil, err
	}
	return result, nil
}

// Resolve marks a trigger as resolved.
func (s *TriggerService) Resolve(ctx context.Context, triggerID string, action domain.ResolutionAction, reason string) (*transition.OperationResult[*domain.Trigger], error) {
	return s.transition(ctx, triggerID, domain.TriggerStatusResolved, action, reason)
}

// Dismiss marks a trigger as dismissed.
func (s *TriggerService) Dismiss(ctx context.Context, triggerID string, reason string) (*transition.OperationResult[*domain.Trigger], error) {
	return s.transition(ctx, triggerID, domain.TriggerStatusDismissed, "", reason)
}

// ListActive returns all active or triggered triggers.
func (s *TriggerService) ListActive(ctx context.Context) ([]*domain.Trigger, error) {
	_ = ctx
	return NewRepository(s.db).ListActive()
}

// ListByRun returns all triggers for a run.
func (s *TriggerService) ListByRun(ctx context.Context, runID string) ([]*domain.Trigger, error) {
	if err := validation.RequiredUUID(runID, "run_id", "trigger_service.list_by_run"); err != nil {
		return nil, err
	}
	_ = ctx
	return NewRepository(s.db).ListByRun(runID)
}

func (s *TriggerService) transition(ctx context.Context, triggerID string, target domain.TriggerStatus, action domain.ResolutionAction, reason string) (*transition.OperationResult[*domain.Trigger], error) {
	op := "trigger_service.transition"
	if err := validation.RequiredUUID(triggerID, "trigger_id", op); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "trigger_service.begin_transition")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	trigger, err := RequireByID(ctx, tx, triggerID)
	if err != nil {
		return nil, err
	}

	if trigger.Status == domain.TriggerStatusResolved || trigger.Status == domain.TriggerStatusDismissed {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "cannot transition from terminal status")
	}
	if target == trigger.Status {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "target status equals current status")
	}

	now := s.clock()
	var resolvedAt *time.Time
	if target == domain.TriggerStatusResolved || target == domain.TriggerStatusDismissed {
		resolvedAt = &now
	}

	var resolutionPtr *domain.ResolutionAction
	if action != "" {
		resolutionPtr = &action
	}

	if err := NewRepository(tx).UpdateStatus(trigger.ID, target, trigger.TriggeredAt, resolvedAt, resolutionPtr); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "trigger_service.update_status", err)
	}

	trigger.Status = target
	trigger.ResolvedAt = resolvedAt
	trigger.ResolutionAction = resolutionPtr

	payload, err := serialization.MarshalPayload("trigger_service.transition_payload", map[string]interface{}{
		"trigger_id": trigger.ID,
		"from_status": string(trigger.Status),
		"to_status":   string(target),
		"reason":      reason,
		"action":      string(action),
	})
	if err != nil {
		return nil, err
	}

	appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:         uuid.New().String(),
		Type:       EventTypeForStatus(target),
		Version:    transition.EventVersionV1,
		TaskID:     ptrValue(trigger.TaskID),
		RunID:      ptrValue(trigger.RunID),
		Priority:   domain.EventPriorityNotification,
		Payload:    payload,
	})
	if err != nil {
		return nil, err
	}

	if err := dbcore.CommitTx(tx, "trigger_service.commit_transition"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.Trigger]{Value: trigger, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *TriggerService) persistDetectedTrigger(ctx context.Context, tx *sql.Tx, trigger *domain.Trigger) (*domain.Trigger, error) {
	trigger.ID = uuid.New().String()
	trigger.CreatedAt = s.clock()
	if trigger.Status == "" {
		trigger.Status = domain.TriggerStatusTriggered
	}

	if err := NewRepository(tx).Create(trigger); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "trigger_service.persist_detected", err)
	}

	payload, err := serialization.MarshalPayload("trigger_service.detected_payload", map[string]interface{}{
		"trigger_id": trigger.ID,
		"anomaly_type": ptrValueString(trigger.AnomalyType),
		"run_id": ptrValue(trigger.RunID),
	})
	if err != nil {
		return nil, err
	}

	_, err = transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:         uuid.New().String(),
		Type:       EventTypeForStatus(trigger.Status),
		Version:    transition.EventVersionV1,
		TaskID:     ptrValue(trigger.TaskID),
		RunID:      ptrValue(trigger.RunID),
		Priority:   domain.EventPriorityNotification,
		Payload:    payload,
	})
	if err != nil {
		return nil, err
	}

	return trigger, nil
}

func (s *TriggerService) requireRunByID(tx *sql.Tx, id string) (*domain.Run, error) {
	run, err := s.newRunReader(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "run.get", err)
	}
	if run == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "run.get", "run not found")
	}
	return run, nil
}

func (s *TriggerService) requireSessionByID(tx *sql.Tx, id string) (*domain.AgentSession, error) {
	session, err := s.newAgentSessionReader(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "agentsession.get", err)
	}
	if session == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "agentsession.get", "agent session not found")
	}
	return session, nil
}

func (s *TriggerService) requireWorkUnitByID(tx *sql.Tx, id string) (*domain.WorkUnit, error) {
	wu, err := s.newWorkUnitReader(tx).GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "workunit.get", err)
	}
	if wu == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, "workunit.get", "work unit not found")
	}
	return wu, nil
}

func (s *TriggerService) listRunEvents(tx *sql.Tx, runID string) ([]domain.EventEnvelope, error) {
	store, err := s.newEventStore(tx)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "trigger_service.new_eventstore", err)
	}
	return store.ListByRun(runID)
}

func (s *TriggerService) listWorkUnitEvents(tx *sql.Tx, workUnitID string) ([]domain.EventEnvelope, error) {
	store, err := s.newEventStore(tx)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "trigger_service.new_eventstore", err)
	}
	return store.ListByWorkUnit(workUnitID)
}

func lastEventTime(events []domain.EventEnvelope) *time.Time {
	if len(events) == 0 {
		return nil
	}
	latest := events[0].CreatedAt
	for _, e := range events[1:] {
		if e.CreatedAt.After(latest) {
			latest = e.CreatedAt
		}
	}
	return &latest
}

func eventTypes(events []domain.EventEnvelope) []string {
	types := make([]string, len(events))
	for i, e := range events {
		types[i] = e.Type
	}
	return types
}

func countNonTriggerEvents(events []domain.EventEnvelope) int {
	count := 0
	for _, e := range events {
		if len(e.Type) < 8 || e.Type[:8] != "trigger." {
			count++
		}
	}
	return count
}

func extractPathsFromEvents(events []domain.EventEnvelope) (accessed []string, modified []string) {
	seenAccessed := make(map[string]bool)
	seenModified := make(map[string]bool)
	for _, e := range events {
		paths := extractPathsFromPayload(e.Payload)
		for _, p := range paths {
			if !seenAccessed[p] {
				seenAccessed[p] = true
				accessed = append(accessed, p)
			}
		}
		// Heuristic: file write events count as modified
		if e.Type == "tool.file_write" || e.Type == "artifact.created" {
			for _, p := range paths {
				if !seenModified[p] {
					seenModified[p] = true
					modified = append(modified, p)
				}
			}
		}
	}
	return accessed, modified
}

func extractPathsFromPayload(payload json.RawMessage) []string {
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil
	}
	var paths []string
	paths = append(paths, collectPaths(data)...)
	return paths
}

func collectPaths(data interface{}) []string {
	var result []string
	switch v := data.(type) {
	case string:
		if len(v) > 0 && (v[0] == '/' || v[0] == '.') {
			result = append(result, v)
		}
	case map[string]interface{}:
		for _, val := range v {
			result = append(result, collectPaths(val)...)
		}
	case []interface{}:
		for _, val := range v {
			result = append(result, collectPaths(val)...)
		}
	}
	return result
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func ptrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrValueString(a *domain.AnomalyType) string {
	if a == nil {
		return ""
	}
	return string(*a)
}
