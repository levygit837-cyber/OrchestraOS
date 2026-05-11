package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
)

type AgentSessionService struct {
	db *sql.DB
}

type CreateAgentSessionInput struct {
	ID               string
	EventID          string
	AgentID          string
	RunID            string
	SandboxID        string
	ConnectionID     string
	LastSeenEventID  string
	RecoverableState json.RawMessage
}

type HeartbeatInput struct {
	EventID string
	Payload map[string]interface{}
}

type CheckpointInput struct {
	EventID        string
	CheckpointID   string
	CurrentGoal    string
	MinimalSummary string
	Ledger         map[string]interface{}
	EvidenceRefs   []string
	OccurredAt     time.Time
	Source         string
	Extra          map[string]interface{}
}

func NewAgentSessionService(database *sql.DB) *AgentSessionService {
	return &AgentSessionService{db: database}
}

func (s *AgentSessionService) Create(ctx context.Context, input CreateAgentSessionInput) (*OperationResult[*domain.AgentSession], error) {
	if input.ID == "" {
		input.ID = uuid.New().String()
	}
	if err := validateCreateAgentSessionInput(input); err != nil {
		return nil, err
	}

	tx, err := beginTx(ctx, s.db, "agent_session_service.begin_create")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	run, err := getRun(ctx, tx, input.RunID)
	if err != nil {
		return nil, err
	}
	if run.Status == domain.RunStatusCompleted || run.Status == domain.RunStatusFailed || run.Status == domain.RunStatusCancelled {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, "agent_session_service.create", "cannot create session for terminal run")
	}
	session := &domain.AgentSession{
		ID:               input.ID,
		AgentID:          input.AgentID,
		RunID:            run.ID,
		SandboxID:        input.SandboxID,
		ConnectionID:     input.ConnectionID,
		Status:           domain.AgentSessionStatusStarting,
		LastSeenEventID:  input.LastSeenEventID,
		RecoverableState: input.RecoverableState,
	}
	if err := repository.NewAgentSessionRepository(tx).Create(session); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.create_projection", err)
	}
	payload, err := marshalPayload("agent_session_service.create_payload", map[string]interface{}{
		"agent_session_id": session.ID,
		"agent_id":         session.AgentID,
		"run_id":           session.RunID,
		"sandbox_id":       session.SandboxID,
		"connection_id":    session.ConnectionID,
		"status":           session.Status,
	})
	if err != nil {
		return nil, err
	}
	appendResult, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "agent.session_starting",
		Version:     eventVersionV1,
		TaskID:      run.TaskID,
		RunID:       run.ID,
		WorkUnitID:  run.WorkUnitID,
		AgentID:     session.AgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}
	if err := commitTx(tx, "agent_session_service.commit_create"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.AgentSession]{Value: session, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *AgentSessionService) Connect(ctx context.Context, sessionID, connectionID, sandboxID string, input TransitionInput) (*OperationResult[*domain.AgentSession], error) {
	if input.Extra == nil {
		input.Extra = map[string]interface{}{}
	}
	if connectionID != "" {
		input.Extra["connection_id"] = connectionID
	}
	if sandboxID != "" {
		input.Extra["sandbox_id"] = sandboxID
	}
	return s.transition(ctx, sessionID, domain.AgentSessionStatusRunning, input, func(ctx context.Context, tx *sql.Tx, session *domain.AgentSession) error {
		if connectionID == "" && session.ConnectionID == "" {
			return apperrors.New(apperrors.CodeInvalidInput, "agent_session_service.connect", "connection_id is required")
		}
		if connectionID != "" {
			session.ConnectionID = connectionID
		}
		if sandboxID != "" {
			session.SandboxID = sandboxID
		}
		res, err := tx.ExecContext(ctx, `UPDATE agent_sessions SET connection_id = $2, sandbox_id = $3, updated_at = $4 WHERE id = $1`, session.ID, session.ConnectionID, session.SandboxID, time.Now().UTC())
		if err != nil {
			return apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_connection", err)
		}
		return ensureRowsAffected(res, "agent session", "agent_session_service.update_connection")
	})
}

func (s *AgentSessionService) Heartbeat(ctx context.Context, sessionID string, input HeartbeatInput) (*OperationResult[*domain.AgentSession], error) {
	op := "agent_session_service.heartbeat"
	if err := validateRequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if err := validateOptionalUUID(input.EventID, "event_id", op); err != nil {
		return nil, err
	}
	tx, err := beginTx(ctx, s.db, "agent_session_service.begin_heartbeat")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	session, err := getAgentSession(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != domain.AgentSessionStatusRunning && session.Status != domain.AgentSessionStatusWaitingApproval && session.Status != domain.AgentSessionStatusPaused {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "heartbeat requires an active session")
	}
	run, err := getRun(ctx, tx, session.RunID)
	if err != nil {
		return nil, err
	}
	payload := input.Payload
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["agent_session_id"] = session.ID
	payload["agent_id"] = session.AgentID
	payload["heartbeat_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	payloadBytes, err := marshalPayload("agent_session_service.heartbeat_payload", payload)
	if err != nil {
		return nil, err
	}
	appendResult, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "agent.heartbeat",
		Version:     eventVersionV1,
		TaskID:      run.TaskID,
		RunID:       run.ID,
		WorkUnitID:  run.WorkUnitID,
		AgentID:     session.AgentID,
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		Payload:     payloadBytes,
	})
	if err != nil {
		return nil, err
	}
	if !appendResult.Duplicate {
		if err := repository.NewAgentSessionRepository(tx).UpdateHeartbeatWithEvent(session.ID, appendResult.Event.ID); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_heartbeat", err)
		}
		session.LastSeenEventID = appendResult.Event.ID
		now := appendResult.Event.CreatedAt
		session.LastHeartbeatAt = &now
	}
	if err := commitTx(tx, "agent_session_service.commit_heartbeat"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.AgentSession]{Value: session, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *AgentSessionService) Checkpoint(ctx context.Context, sessionID string, input CheckpointInput) (*OperationResult[*domain.AgentSession], error) {
	op := "agent_session_service.checkpoint"
	if err := validateRequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if err := validateOptionalUUID(input.EventID, "event_id", op); err != nil {
		return nil, err
	}
	if err := validateRequiredText(input.CheckpointID, "checkpoint_id", op); err != nil {
		return nil, err
	}
	if err := validateRequiredText(input.CurrentGoal, "current_goal", op); err != nil {
		return nil, err
	}
	if err := validateRequiredText(input.MinimalSummary, "minimal_summary", op); err != nil {
		return nil, err
	}
	if len(input.Ledger) == 0 {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "ledger is required")
	}

	tx, err := beginTx(ctx, s.db, "agent_session_service.begin_checkpoint")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	session, err := getAgentSession(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != domain.AgentSessionStatusRunning && session.Status != domain.AgentSessionStatusWaitingApproval && session.Status != domain.AgentSessionStatusPaused {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "checkpoint requires an active session")
	}
	run, err := getRun(ctx, tx, session.RunID)
	if err != nil {
		return nil, err
	}
	occurredAt := input.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	} else {
		occurredAt = occurredAt.UTC()
	}
	payload := map[string]interface{}{
		"agent_session_id": session.ID,
		"checkpoint_id":    input.CheckpointID,
		"current_goal":     input.CurrentGoal,
		"minimal_summary":  input.MinimalSummary,
		"ledger":           input.Ledger,
		"occurred_at":      occurredAt.Format(time.RFC3339Nano),
	}
	if len(input.EvidenceRefs) > 0 {
		payload["evidence_refs"] = input.EvidenceRefs
	}
	if input.Source != "" {
		payload["source"] = input.Source
	}
	for key, value := range input.Extra {
		payload[key] = value
	}
	payloadBytes, err := marshalPayload("agent_session_service.checkpoint_payload", payload)
	if err != nil {
		return nil, err
	}
	appendResult, err := appendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "agent.checkpoint_reached",
		Version:     eventVersionV1,
		TaskID:      run.TaskID,
		RunID:       run.ID,
		WorkUnitID:  run.WorkUnitID,
		AgentID:     session.AgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payloadBytes,
	})
	if err != nil {
		return nil, err
	}
	if !appendResult.Duplicate {
		repo := repository.NewAgentSessionRepository(tx)
		if err := repo.UpdateCheckpointWithEvent(session.ID, appendResult.Event.ID); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_checkpoint", err)
		}
		recoverableState, err := marshalPayload("agent_session_service.recoverable_checkpoint_state", map[string]interface{}{
			"agent_session_id":         session.ID,
			"agent_id":                 session.AgentID,
			"run_id":                   run.ID,
			"work_unit_id":             run.WorkUnitID,
			"last_checkpoint_event_id": appendResult.Event.ID,
			"checkpoint":               payload,
			"recoverable_at":           appendResult.Event.CreatedAt.Format(time.RFC3339Nano),
		})
		if err != nil {
			return nil, err
		}
		if err := repo.UpdateRecoverableState(session.ID, recoverableState); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_recoverable_checkpoint_state", err)
		}
		session.LastSeenEventID = appendResult.Event.ID
		now := appendResult.Event.CreatedAt
		session.LastCheckpointAt = &now
		session.RecoverableState = recoverableState
	}
	if err := commitTx(tx, "agent_session_service.commit_checkpoint"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.AgentSession]{Value: session, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *AgentSessionService) Disconnect(ctx context.Context, sessionID string, input TransitionInput) (*OperationResult[*domain.AgentSession], error) {
	return s.transition(ctx, sessionID, domain.AgentSessionStatusDisconnected, input, nil)
}

func (s *AgentSessionService) Resume(ctx context.Context, sessionID string, input TransitionInput) (*OperationResult[*domain.AgentSession], error) {
	return s.transition(ctx, sessionID, domain.AgentSessionStatusRunning, input, nil)
}

func (s *AgentSessionService) Stop(ctx context.Context, sessionID string, input TransitionInput) (*OperationResult[*domain.AgentSession], error) {
	op := "agent_session_service.stop"
	if err := validateRequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if input.Justification == "" {
		input.Justification = "session stop requested"
	}
	tx, err := beginTx(ctx, s.db, "agent_session_service.begin_stop")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	session, err := getAgentSession(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	run, err := getRun(ctx, tx, session.RunID)
	if err != nil {
		return nil, err
	}
	var lastEvent *domain.EventEnvelope
	var duplicate bool
	if session.Status != domain.AgentSessionStatusStopping {
		stoppingInput := input
		stoppingInput.EventID = ""
		event, dup, err := transitionAgentSessionInTx(ctx, tx, session, run, domain.AgentSessionStatusStopping, stoppingInput)
		if err != nil {
			return nil, err
		}
		lastEvent = event
		duplicate = dup
		session.Status = domain.AgentSessionStatusStopping
	}
	event, dup, err := transitionAgentSessionInTx(ctx, tx, session, run, domain.AgentSessionStatusStopped, input)
	if err != nil {
		return nil, err
	}
	lastEvent = event
	duplicate = duplicate || dup
	session.Status = domain.AgentSessionStatusStopped

	if err := commitTx(tx, "agent_session_service.commit_stop"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.AgentSession]{Value: session, Event: lastEvent, Duplicate: duplicate}, nil
}

func (s *AgentSessionService) Timeout(ctx context.Context, sessionID string, recoverableState json.RawMessage, input TransitionInput) (*OperationResult[*domain.AgentSession], error) {
	if input.Justification == "" {
		input.Justification = "session heartbeat timeout reached"
	}
	if input.Extra == nil {
		input.Extra = map[string]interface{}{}
	}
	input.Extra["timeout"] = true
	input.Extra["recoverable"] = true
	return s.transition(ctx, sessionID, domain.AgentSessionStatusDisconnected, input, func(ctx context.Context, tx *sql.Tx, session *domain.AgentSession) error {
		if len(recoverableState) > 0 {
			if err := repository.NewAgentSessionRepository(tx).UpdateRecoverableState(session.ID, recoverableState); err != nil {
				return apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_recoverable_state", err)
			}
			session.RecoverableState = recoverableState
		}
		run, err := getRun(ctx, tx, session.RunID)
		if err != nil {
			return err
		}
		if run.Status == domain.RunStatusRunning || run.Status == domain.RunStatusWaitingApproval {
			if err := statemachine.CanTransition(statemachine.AggregateRun, string(run.Status), string(domain.RunStatusPaused), transitionContext(input)); err != nil {
				return err
			}
			if _, _, err := appendTransition(ctx, tx, "", "run.paused", run.TaskID, run.ID, run.WorkUnitID, session.AgentID, transitionPayload(run.Status, domain.RunStatusPaused, input)); err != nil {
				return err
			}
			if err := updateRunProjection(ctx, tx, run.ID, domain.RunStatusPaused, nil, nil); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *AgentSessionService) Fail(ctx context.Context, sessionID string, input TransitionInput) (*OperationResult[*domain.AgentSession], error) {
	return s.transition(ctx, sessionID, domain.AgentSessionStatusFailed, input, nil)
}

func (s *AgentSessionService) transition(ctx context.Context, sessionID string, target domain.AgentSessionStatus, input TransitionInput, after func(context.Context, *sql.Tx, *domain.AgentSession) error) (*OperationResult[*domain.AgentSession], error) {
	op := "agent_session_service.transition"
	if err := validateRequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if err := requireFinalAudit(string(target), input, op); err != nil {
		return nil, err
	}
	tx, err := beginTx(ctx, s.db, "agent_session_service.begin_transition")
	if err != nil {
		return nil, err
	}
	defer rollbackTx(tx)

	session, err := getAgentSession(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	run, err := getRun(ctx, tx, session.RunID)
	if err != nil {
		return nil, err
	}
	event, duplicate, err := transitionAgentSessionInTx(ctx, tx, session, run, target, input)
	if err != nil {
		return nil, err
	}
	if after != nil && !duplicate {
		if err := after(ctx, tx, session); err != nil {
			return nil, err
		}
	}
	session.Status = target
	if err := commitTx(tx, "agent_session_service.commit_transition"); err != nil {
		return nil, err
	}
	return &OperationResult[*domain.AgentSession]{Value: session, Event: event, Duplicate: duplicate}, nil
}

func transitionAgentSessionInTx(ctx context.Context, tx *sql.Tx, session *domain.AgentSession, run *domain.Run, target domain.AgentSessionStatus, input TransitionInput) (*domain.EventEnvelope, bool, error) {
	if err := statemachine.CanTransition(statemachine.AggregateAgentSession, string(session.Status), string(target), transitionContext(input)); err != nil {
		return nil, false, err
	}
	event, duplicate, err := appendTransition(ctx, tx, input.EventID, eventTypeForAgentSessionStatus(target), run.TaskID, run.ID, run.WorkUnitID, session.AgentID, transitionPayload(session.Status, target, input))
	if err != nil {
		return nil, false, err
	}
	if !duplicate {
		now := time.Now().UTC()
		var heartbeatAt, checkpointAt *time.Time
		if target == domain.AgentSessionStatusRunning {
			heartbeatAt = &now
		}
		res, err := tx.ExecContext(ctx, db.QueryAgentSessionUpdateStatus, session.ID, target, heartbeatAt, checkpointAt, now)
		if err != nil {
			return nil, false, apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_projection", err)
		}
		if err := ensureRowsAffected(res, "agent session", "agent_session_service.update_projection"); err != nil {
			return nil, false, err
		}
	}
	return event, duplicate, nil
}

func validateCreateAgentSessionInput(input CreateAgentSessionInput) error {
	op := "agent_session_service.validate_create"
	if err := validateRequiredUUID(input.ID, "agent_session_id", op); err != nil {
		return err
	}
	if err := validateOptionalUUID(input.EventID, "event_id", op); err != nil {
		return err
	}
	if err := validateRequiredText(input.AgentID, "agent_id", op); err != nil {
		return err
	}
	if err := validateRequiredUUID(input.RunID, "run_id", op); err != nil {
		return err
	}
	if err := validateOptionalUUID(input.LastSeenEventID, "last_seen_event_id", op); err != nil {
		return err
	}
	if len(input.RecoverableState) > 0 && !json.Valid(input.RecoverableState) {
		return apperrors.New(apperrors.CodeValidation, op, "recoverable_state must be valid JSON")
	}
	return nil
}
