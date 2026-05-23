// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
)

type Service struct {
	executor db.DBTX
}

type AppendResult struct {
	Event     Envelope
	Duplicate bool
}

func NewService(executor db.DBTX) *Service {
	return &Service{executor: executor}
}

func (s *Service) Append(ctx context.Context, envelope *Envelope) (*AppendResult, error) {
	if envelope == nil {
		return nil, apperrors.New(apperrors.CodeInvalidInput, "event_service.append", "event envelope is required")
	}
	if err := s.validateEnvelopeInput(envelope); err != nil {
		return nil, err
	}
	store, err := eventstore.NewStoreWithExecutor(s.executor)
	if err != nil {
		return nil, err
	}

	if envelope.ID != "" {
		existing, err := store.Get(envelope.ID)
		if err != nil {
			return nil, apperrors.Wrap(apperrors.CodePersistence, "event_service.get_existing", err)
		}
		if existing != nil {
			if !sameEventIntent(existing, envelope) {
				return nil, apperrors.New(apperrors.CodeConflict, "event_service.idempotency", "event_id already exists with different event content")
			}
			return &AppendResult{Event: *existing, Duplicate: true}, nil
		}
	}

	persisted, duplicate, err := store.AppendResult(envelope)
	if err != nil {
		return nil, err
	}
	return &AppendResult{Event: *persisted, Duplicate: duplicate}, nil
}

func (s *Service) List(ctx context.Context) ([]Envelope, error) {
	// ctx reserved for future cancellation; intentionally ignored
	_ = ctx //nolint:ctx-ignored // ctx reserved for future cancellation; intentionally ignored
	store, err := eventstore.NewStoreWithExecutor(s.executor)
	if err != nil {
		return nil, err
	}
	return store.List()
}

func (s *Service) ListByTask(ctx context.Context, taskID string) ([]Envelope, error) {
	// ctx reserved for future cancellation; intentionally ignored
	_ = ctx //nolint:ctx-ignored // ctx reserved for future cancellation; intentionally ignored
	if err := validation.RequiredUUID(taskID, "task_id", "event_service.list_by_task"); err != nil {
		return nil, err
	}
	store, err := eventstore.NewStoreWithExecutor(s.executor)
	if err != nil {
		return nil, err
	}
	return store.ListByTask(taskID)
}

func (s *Service) ListByRun(ctx context.Context, runID string) ([]Envelope, error) {
	// ctx reserved for future cancellation; intentionally ignored
	_ = ctx //nolint:ctx-ignored // ctx reserved for future cancellation; intentionally ignored
	if err := validation.RequiredUUID(runID, "run_id", "event_service.list_by_run"); err != nil {
		return nil, err
	}
	store, err := eventstore.NewStoreWithExecutor(s.executor)
	if err != nil {
		return nil, err
	}
	return store.ListByRun(runID)
}

func (s *Service) ListByWorkUnit(ctx context.Context, workUnitID string) ([]Envelope, error) {
	// ctx reserved for future cancellation; intentionally ignored
	_ = ctx //nolint:ctx-ignored // ctx reserved for future cancellation; intentionally ignored
	if err := validation.RequiredUUID(workUnitID, "work_unit_id", "event_service.list_by_work_unit"); err != nil {
		return nil, err
	}
	store, err := eventstore.NewStoreWithExecutor(s.executor)
	if err != nil {
		return nil, err
	}
	return store.ListByWorkUnit(workUnitID)
}

func (s *Service) ReplayTask(ctx context.Context, taskID string) (*statemachine.ReplayState, error) {
	// ctx reserved for future cancellation; intentionally ignored
	_ = ctx //nolint:ctx-ignored // ctx reserved for future cancellation; intentionally ignored
	if err := validation.RequiredUUID(taskID, "task_id", "event_service.replay_task"); err != nil {
		return nil, err
	}
	store, err := eventstore.NewStoreWithExecutor(s.executor)
	if err != nil {
		return nil, err
	}
	return store.ReplayState(taskID)
}

func (s *Service) ReplayRun(ctx context.Context, runID string) (*statemachine.ReplayState, error) {
	// ctx reserved for future cancellation; intentionally ignored
	_ = ctx //nolint:ctx-ignored // ctx reserved for future cancellation; intentionally ignored
	if err := validation.RequiredUUID(runID, "run_id", "event_service.replay_run"); err != nil {
		return nil, err
	}
	store, err := eventstore.NewStoreWithExecutor(s.executor)
	if err != nil {
		return nil, err
	}
	return store.ReplayRunState(runID)
}

func (s *Service) GetLastCheckpointByRun(ctx context.Context, runID string) (*Envelope, error) {
	// ctx reserved for future cancellation; intentionally ignored
	_ = ctx //nolint:ctx-ignored // ctx reserved for future cancellation; intentionally ignored
	if err := validation.RequiredUUID(runID, "run_id", "event_service.last_checkpoint_by_run"); err != nil {
		return nil, err
	}
	store, err := eventstore.NewStoreWithExecutor(s.executor)
	if err != nil {
		return nil, err
	}
	return store.GetLastCheckpointByRun(runID)
}

func (s *Service) validateEnvelopeInput(envelope *Envelope) error {
	op := "event_service.validate_envelope"
	if err := validation.RequiredText(envelope.Type, "type", op); err != nil {
		return err
	}
	if err := validation.RequiredText(envelope.Version, "version", op); err != nil {
		return err
	}
	if err := validation.RequiredUUID(envelope.TaskID, "task_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(envelope.ID, "event_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(envelope.RunID, "run_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(envelope.WorkUnitID, "work_unit_id", op); err != nil {
		return err
	}
	if requiresRunID(envelope.Type) && strings.TrimSpace(envelope.RunID) == "" {
		return apperrors.New(apperrors.CodeInvalidInput, op, "run_id is required for run, agent, and tool events")
	}
	if envelope.Priority != "" {
		switch envelope.Priority {
		case PriorityInterrupt, PriorityCheckpoint, PriorityNotification, PriorityBackground:
		default:
			return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("invalid event priority %q", envelope.Priority))
		}
	}
	if len(envelope.Payload) > 0 && !json.Valid(envelope.Payload) {
		return apperrors.New(apperrors.CodeValidation, op, "payload must be valid JSON")
	}
	return nil
}

func requiresRunID(eventType string) bool {
	return strings.HasPrefix(eventType, "run.") || strings.HasPrefix(eventType, "agent.") || strings.HasPrefix(eventType, "tool.")
}

func sameEventIntent(existing *Envelope, candidate *Envelope) bool {
	if existing == nil || candidate == nil {
		return false
	}
	existingPriority := existing.Priority
	if existingPriority == "" {
		existingPriority = PriorityBackground
	}
	candidatePriority := candidate.Priority
	if candidatePriority == "" {
		candidatePriority = PriorityBackground
	}
	if existing.Type != candidate.Type ||
		existing.Version != candidate.Version ||
		existing.TaskID != candidate.TaskID ||
		existing.RunID != candidate.RunID ||
		existing.WorkUnitID != candidate.WorkUnitID ||
		existing.AgentID != candidate.AgentID ||
		existing.TraceID != candidate.TraceID ||
		existing.SpanID != candidate.SpanID ||
		existing.ParentSpanID != candidate.ParentSpanID ||
		existingPriority != candidatePriority ||
		existing.RequiresAck != candidate.RequiresAck {
		return false
	}
	if len(candidate.Payload) == 0 {
		return len(existing.Payload) == 0 || bytes.Equal(existing.Payload, []byte(`{}`))
	}
	var existingPayload any
	var candidatePayload any
	if err := json.Unmarshal(existing.Payload, &existingPayload); err != nil {
		return false
	}
	if err := json.Unmarshal(candidate.Payload, &candidatePayload); err != nil {
		return false
	}
	return reflect.DeepEqual(existingPayload, candidatePayload)
}
