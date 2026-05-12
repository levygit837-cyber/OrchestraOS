package transition

import (
	"context"
	"database/sql"

	eventmod "github.com/levygit837-cyber/OrchestraOS/internal/core/event"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/serialization"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

const EventVersionV1 = "v1"

// AppendServiceEvent appends a generic service event to the event store.
func AppendServiceEvent(ctx context.Context, tx *sql.Tx, envelope *domain.EventEnvelope) (*eventmod.AppendResult, error) {
	service := eventmod.NewService(tx)
	return service.Append(ctx, envelope)
}

// AppendTransition appends a state-machine transition event to the event store.
func AppendTransition(ctx context.Context, tx *sql.Tx, eventID, eventType, taskID, runID, workUnitID, agentID string, payload map[string]interface{}) (*domain.EventEnvelope, bool, error) {
	payloadBytes, err := serialization.MarshalPayload("transition.transition_payload", payload)
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
