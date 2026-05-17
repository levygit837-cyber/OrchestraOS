// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package agentsession

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// AgentReader abstracts agent reads to avoid cyclic imports.
type AgentReader interface {
	GetByID(ctx context.Context, id string) (*domain.Agent, error)
}

type AgentSessionService struct {
	db             *sql.DB
	newAgentReader func(*sql.Tx) AgentReader
}

type CreateAgentSessionInput struct {
	ID               string
	EventID          string
	AgentID          string
	RunID            string
	TaskID           string
	WorkUnitID       string
	SandboxID        string
	ConnectionID     string
	LastSeenEventID  string
	RecoverableState json.RawMessage
}

func NewAgentSessionService(database *sql.DB, newAgentReader func(*sql.Tx) AgentReader) *AgentSessionService {
	return &AgentSessionService{db: database, newAgentReader: newAgentReader}
}

func (s *AgentSessionService) Create(ctx context.Context, input CreateAgentSessionInput) (*transition.OperationResult[*domain.AgentSession], error) {
	if input.ID == "" {
		input.ID = uuid.New().String()
	}
	if err := validateCreateAgentSessionInput(input); err != nil {
		return nil, err
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "agent_session_service.begin_create")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	// Validate AgentID exists
	_, err = s.requireAgentByID(ctx, tx, input.AgentID)
	if err != nil {
		return nil, err
	}

	session := &domain.AgentSession{
		ID:               input.ID,
		AgentID:          input.AgentID,
		RunID:            input.RunID,
		TaskID:           input.TaskID,
		WorkUnitID:       input.WorkUnitID,
		SandboxID:        input.SandboxID,
		ConnectionID:     input.ConnectionID,
		Status:           domain.AgentSessionStatusStarting,
		LastSeenEventID:  input.LastSeenEventID,
		RecoverableState: input.RecoverableState,
	}
	if err := NewRepository(tx).Create(session); err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.create_projection", err)
	}
	payload, err := serialization.MarshalPayload("agent_session_service.create_payload", map[string]interface{}{
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
	appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "agent.session_starting",
		Version:     transition.EventVersionV1,
		TaskID:      session.TaskID,
		RunID:       session.RunID,
		WorkUnitID:  session.WorkUnitID,
		AgentID:     session.AgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		return nil, err
	}
	if err := dbcore.CommitTx(tx, "agent_session_service.commit_create"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.AgentSession]{Value: session, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}

func (s *AgentSessionService) Connect(ctx context.Context, sessionID, connectionID, sandboxID string, input transition.TransitionInput) (*transition.OperationResult[*domain.AgentSession], error) {
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
		return dbcore.EnsureRowsAffected(res, "agent session", "agent_session_service.update_connection")
	})
}

func (s *AgentSessionService) Disconnect(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*domain.AgentSession], error) {
	return s.transition(ctx, sessionID, domain.AgentSessionStatusDisconnected, input, nil)
}

func (s *AgentSessionService) Resume(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*domain.AgentSession], error) {
	return s.transition(ctx, sessionID, domain.AgentSessionStatusRunning, input, nil)
}

func (s *AgentSessionService) Stop(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*domain.AgentSession], error) {
	op := "agent_session_service.stop"
	if err := validation.RequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if input.Justification == "" {
		input.Justification = "session stop requested"
	}
	tx, err := dbcore.BeginTx(ctx, s.db, "agent_session_service.begin_stop")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	session, err := RequireByID(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	var lastEvent *domain.EventEnvelope
	var duplicate bool
	if session.Status != domain.AgentSessionStatusStopping {
		stoppingInput := input
		stoppingInput.EventID = ""
		_, dup, err := transitionAgentSessionInTx(ctx, tx, session, session.TaskID, session.WorkUnitID, domain.AgentSessionStatusStopping, stoppingInput)
		if err != nil {
			return nil, err
		}
		duplicate = dup
		session.Status = domain.AgentSessionStatusStopping
	}
	event, dup, err := transitionAgentSessionInTx(ctx, tx, session, session.TaskID, session.WorkUnitID, domain.AgentSessionStatusStopped, input)
	if err != nil {
		return nil, err
	}
	lastEvent = event
	duplicate = duplicate || dup
	session.Status = domain.AgentSessionStatusStopped

	if err := dbcore.CommitTx(tx, "agent_session_service.commit_stop"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.AgentSession]{Value: session, Event: lastEvent, Duplicate: duplicate}, nil
}

func (s *AgentSessionService) Timeout(ctx context.Context, sessionID string, recoverableState json.RawMessage, input transition.TransitionInput) (*transition.OperationResult[*domain.AgentSession], error) {
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
			if err := NewRepository(tx).UpdateRecoverableState(session.ID, recoverableState); err != nil {
				return apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_recoverable_state", err)
			}
			session.RecoverableState = recoverableState
		}
		return nil
	})
}

func (s *AgentSessionService) Fail(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*domain.AgentSession], error) {
	return s.transition(ctx, sessionID, domain.AgentSessionStatusFailed, input, nil)
}

func (s *AgentSessionService) transition(ctx context.Context, sessionID string, target domain.AgentSessionStatus, input transition.TransitionInput, after func(context.Context, *sql.Tx, *domain.AgentSession) error) (*transition.OperationResult[*domain.AgentSession], error) {
	op := "agent_session_service.transition"
	if err := validation.RequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if err := transition.RequireFinalAudit(string(target), input, op); err != nil {
		return nil, err
	}
	tx, err := dbcore.BeginTx(ctx, s.db, "agent_session_service.begin_transition")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	session, err := RequireByID(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	event, duplicate, err := transitionAgentSessionInTx(ctx, tx, session, session.TaskID, session.WorkUnitID, target, input)
	if err != nil {
		return nil, err
	}
	if after != nil && !duplicate {
		if err := after(ctx, tx, session); err != nil {
			return nil, err
		}
	}
	session.Status = target
	if err := dbcore.CommitTx(tx, "agent_session_service.commit_transition"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.AgentSession]{Value: session, Event: event, Duplicate: duplicate}, nil
}

func transitionAgentSessionInTx(ctx context.Context, tx *sql.Tx, session *domain.AgentSession, taskID, workUnitID string, target domain.AgentSessionStatus, input transition.TransitionInput) (*domain.EventEnvelope, bool, error) {
	if err := statemachine.CanTransition(statemachine.AggregateAgentSession, string(session.Status), string(target), transition.TransitionContext(input)); err != nil {
		return nil, false, err
	}
	event, duplicate, err := transition.AppendTransition(ctx, tx, input.EventID, EventTypeForStatus(target), taskID, session.RunID, workUnitID, session.AgentID, transition.TransitionPayload(session.Status, target, input))
	if err != nil {
		return nil, false, err
	}
	if !duplicate {
		now := time.Now().UTC()
		var heartbeatAt, checkpointAt *time.Time
		if target == domain.AgentSessionStatusRunning {
			heartbeatAt = &now
		}
		res, err := tx.ExecContext(ctx, QueryUpdateStatus, session.ID, target, heartbeatAt, checkpointAt, now)
		if err != nil {
			return nil, false, apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_projection", err)
		}
		if err := dbcore.EnsureRowsAffected(res, "agent session", "agent_session_service.update_projection"); err != nil {
			return nil, false, err
		}
	}
	return event, duplicate, nil
}

func validateCreateAgentSessionInput(input CreateAgentSessionInput) error {
	op := "agent_session_service.validate_create"
	if err := validation.RequiredUUID(input.ID, "agent_session_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.EventID, "event_id", op); err != nil {
		return err
	}
	if err := validation.RequiredText(input.AgentID, "agent_id", op); err != nil {
		return err
	}
	if err := validation.RequiredUUID(input.RunID, "run_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.LastSeenEventID, "last_seen_event_id", op); err != nil {
		return err
	}
	if len(input.RecoverableState) > 0 && !json.Valid(input.RecoverableState) {
		return apperrors.New(apperrors.CodeValidation, op, "recoverable_state must be valid JSON")
	}
	return nil
}

func (s *AgentSessionService) requireAgentByID(ctx context.Context, tx *sql.Tx, id string) (*domain.Agent, error) {
	op := "agent_session_service.require_agent"
	agent, err := s.newAgentReader(tx).GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodePersistence, op, err)
	}
	if agent == nil {
		return nil, apperrors.New(apperrors.CodeNotFound, op, "agent not found")
	}
	return agent, nil
}
