package orchestration

import (
	"context"
	"database/sql"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	eventmod "github.com/levygit837-cyber/OrchestraOS/internal/core/event"
)

const EventVersionV1 = "v1"

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

func TransitionPayload(from, to interface{}, input TransitionInput) map[string]interface{} {
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

func TransitionContext(input TransitionInput) statemachine.TransitionContext {
	return statemachine.TransitionContext{
		EvidenceRefs:      input.EvidenceRefs,
		ValidationEventID: input.ValidationEventID,
		Justification:     input.Justification,
	}
}

func RequireFinalAudit(target string, input TransitionInput, op string) error {
	if !IsFinalStatus(target) {
		return nil
	}
	if len(input.EvidenceRefs) > 0 || input.ValidationEventID != "" || input.Justification != "" || input.FailureReason != "" {
		return nil
	}
	return apperrors.New(apperrors.CodeInvalidInput, op, "final state requires evidence, validation event, failure reason, or justification")
}

func IsFinalStatus(status string) bool {
	switch status {
	case "completed", "failed", "cancelled", "stopped":
		return true
	default:
		return false
	}
}

func AppendServiceEvent(ctx context.Context, tx *sql.Tx, envelope *domain.EventEnvelope) (*eventmod.AppendResult, error) {
	service := eventmod.NewService(tx)
	return service.Append(ctx, envelope)
}

func AppendTransition(ctx context.Context, tx *sql.Tx, eventID, eventType, taskID, runID, workUnitID, agentID string, payload map[string]interface{}) (*domain.EventEnvelope, bool, error) {
	payloadBytes, err := serialization.MarshalPayload("orchestration.transition_payload", payload)
	if err != nil {
		return nil, false, err
	}
	result, err := eventmod.NewService(tx).Append(ctx, &domain.EventEnvelope{
		ID:          eventID,
		Type:        eventType,
		Version:     EventVersionV1,
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
