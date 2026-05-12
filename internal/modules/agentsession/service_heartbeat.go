// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package agentsession

import (
	"context"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func (s *AgentSessionService) Heartbeat(ctx context.Context, sessionID string, input HeartbeatInput) (*transition.OperationResult[*domain.AgentSession], error) {
	op := "agent_session_service.heartbeat"
	if err := validation.RequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if err := validation.OptionalUUID(input.EventID, "event_id", op); err != nil {
		return nil, err
	}
	tx, err := dbcore.BeginTx(ctx, s.db, "agent_session_service.begin_heartbeat")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	session, err := RequireByID(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != domain.AgentSessionStatusRunning && session.Status != domain.AgentSessionStatusWaitingApproval && session.Status != domain.AgentSessionStatusPaused {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "heartbeat requires an active session")
	}
	payload := input.Payload
	if payload == nil {
		payload = map[string]interface{}{}
	}
	payload["agent_session_id"] = session.ID
	payload["agent_id"] = session.AgentID
	payload["heartbeat_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	payloadBytes, err := serialization.MarshalPayload("agent_session_service.heartbeat_payload", payload)
	if err != nil {
		return nil, err
	}
	appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "agent.heartbeat",
		Version:     transition.EventVersionV1,
		TaskID:      session.TaskID,
		RunID:       session.RunID,
		WorkUnitID:  session.WorkUnitID,
		AgentID:     session.AgentID,
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		Payload:     payloadBytes,
	})
	if err != nil {
		return nil, err
	}
	if !appendResult.Duplicate {
		if err := NewRepository(tx).UpdateHeartbeatWithEvent(session.ID, appendResult.Event.ID); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_heartbeat", err)
		}
		session.LastSeenEventID = appendResult.Event.ID
		now := appendResult.Event.CreatedAt
		session.LastHeartbeatAt = &now
	}
	if err := dbcore.CommitTx(tx, "agent_session_service.commit_heartbeat"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*domain.AgentSession]{Value: session, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}
