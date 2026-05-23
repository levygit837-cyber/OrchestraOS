package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/store"
)

// Emitter provides a simple interface to emit events to the store.
type Emitter struct {
	store store.Store
}

func NewEmitter(s store.Store) *Emitter {
	return &Emitter{store: s}
}

func (e *Emitter) Emit(ctx context.Context, eventType, taskID string, payload []byte) error {
	env := &domain.EventEnvelope{
		ID:        uuid.New().String(),
		Type:      eventType,
		Version:   "1.0",
		TaskID:    taskID,
		Priority:  domain.EventPriorityNotification,
		CreatedAt: time.Now(),
		Payload:   payload,
	}
	return e.store.AppendEvent(ctx, env)
}
