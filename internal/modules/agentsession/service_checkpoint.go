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

func (s *AgentSessionService) Checkpoint(ctx context.Context, sessionID string, input domain.CheckpointInput) (*transition.OperationResult[*AgentSession], error) {
	op := "agent_session_service.checkpoint"
	if err := validation.RequiredUUID(sessionID, "agent_session_id", op); err != nil {
		return nil, err
	}
	if err := validation.OptionalUUID(input.EventID, "event_id", op); err != nil {
		return nil, err
	}
	if err := validation.RequiredText(input.CheckpointID, "checkpoint_id", op); err != nil {
		return nil, err
	}
	if err := validation.RequiredText(input.CurrentGoal, "current_goal", op); err != nil {
		return nil, err
	}
	if err := validation.RequiredText(input.MinimalSummary, "minimal_summary", op); err != nil {
		return nil, err
	}
	if len(input.Ledger) == 0 {
		return nil, apperrors.New(apperrors.CodeInvalidInput, op, "ledger is required")
	}

	tx, err := dbcore.BeginTx(ctx, s.db, "agent_session_service.begin_checkpoint")
	if err != nil {
		return nil, err
	}
	defer dbcore.RollbackTx(tx)

	session, err := RequireByID(ctx, tx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != StatusRunning && session.Status != StatusWaitingApproval && session.Status != StatusPaused {
		return nil, apperrors.New(apperrors.CodeInvalidTransition, op, "checkpoint requires an active session")
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
	payloadBytes, err := serialization.MarshalPayload("agent_session_service.checkpoint_payload", payload)
	if err != nil {
		return nil, err
	}
	appendResult, err := transition.AppendServiceEvent(ctx, tx, &domain.EventEnvelope{
		ID:          input.EventID,
		Type:        "agent.checkpoint_reached",
		Version:     transition.EventVersionV1,
		TaskID:      session.TaskID,
		RunID:       session.RunID,
		WorkUnitID:  session.WorkUnitID,
		AgentID:     session.AgentID,
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload:     payloadBytes,
	})
	if err != nil {
		return nil, err
	}
	if !appendResult.Duplicate {
		repo := NewRepository(tx)
		if err := repo.UpdateCheckpointWithEvent(session.ID, appendResult.Event.ID, appendResult.Event.CreatedAt, appendResult.Event.CreatedAt); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_checkpoint", err)
		}
		recoverableState, err := serialization.MarshalPayload("agent_session_service.recoverable_checkpoint_state", map[string]interface{}{
			"agent_session_id":         session.ID,
			"agent_id":                 session.AgentID,
			"run_id":                   session.RunID,
			"work_unit_id":             session.WorkUnitID,
			"last_checkpoint_event_id": appendResult.Event.ID,
			"checkpoint":               payload,
			"recoverable_at":           appendResult.Event.CreatedAt.Format(time.RFC3339Nano),
		})
		if err != nil {
			return nil, err
		}
		if err := repo.UpdateRecoverableState(session.ID, recoverableState, appendResult.Event.CreatedAt); err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "agent_session_service.update_recoverable_checkpoint_state", err)
		}
		session.LastSeenEventID = appendResult.Event.ID
		now := appendResult.Event.CreatedAt
		session.LastCheckpointAt = &now
		session.RecoverableState = recoverableState
	}
	if err := dbcore.CommitTx(tx, "agent_session_service.commit_checkpoint"); err != nil {
		return nil, err
	}
	return &transition.OperationResult[*AgentSession]{Value: session, Event: &appendResult.Event, Duplicate: appendResult.Duplicate}, nil
}
