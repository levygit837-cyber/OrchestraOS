package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
	"github.com/levygit837-cyber/OrchestraOS/internal/statemachine"
)

type EventService struct {
	db       *sql.DB
	executor repository.DBTX
}

type AppendResult struct {
	Event     domain.EventEnvelope
	Duplicate bool
}

func NewEventService(database *sql.DB) *EventService {
	return &EventService{db: database, executor: database}
}

func NewEventServiceWithExecutor(executor repository.DBTX) *EventService {
	return &EventService{executor: executor}
}

func (s *EventService) Append(ctx context.Context, envelope *domain.EventEnvelope) (*AppendResult, error) {
	if envelope == nil {
		return nil, apperrors.New(apperrors.CodeInvalidInput, "event_service.append", "event envelope is required")
	}
	if err := s.validateEnvelopeInput(envelope); err != nil {
		return nil, err
	}
	store, err := newEventStore(s.executor)
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

	if err := s.validateReferences(ctx, envelope); err != nil {
		return nil, err
	}

	persisted, duplicate, err := store.AppendResult(envelope)
	if err != nil {
		return nil, err
	}
	return &AppendResult{Event: *persisted, Duplicate: duplicate}, nil
}

func (s *EventService) AppendInTx(ctx context.Context, tx *sql.Tx, envelope *domain.EventEnvelope) (*AppendResult, error) {
	return NewEventServiceWithExecutor(tx).Append(ctx, envelope)
}

func (s *EventService) List(ctx context.Context) ([]domain.EventEnvelope, error) {
	_ = ctx
	store, err := newEventStore(s.executor)
	if err != nil {
		return nil, err
	}
	return store.List()
}

func (s *EventService) ListByTask(ctx context.Context, taskID string) ([]domain.EventEnvelope, error) {
	_ = ctx
	if err := validateRequiredUUID(taskID, "task_id", "event_service.list_by_task"); err != nil {
		return nil, err
	}
	store, err := newEventStore(s.executor)
	if err != nil {
		return nil, err
	}
	return store.ListByTask(taskID)
}

func (s *EventService) ListByRun(ctx context.Context, runID string) ([]domain.EventEnvelope, error) {
	_ = ctx
	if err := validateRequiredUUID(runID, "run_id", "event_service.list_by_run"); err != nil {
		return nil, err
	}
	store, err := newEventStore(s.executor)
	if err != nil {
		return nil, err
	}
	return store.ListByRun(runID)
}

func (s *EventService) ListByWorkUnit(ctx context.Context, workUnitID string) ([]domain.EventEnvelope, error) {
	_ = ctx
	if err := validateRequiredUUID(workUnitID, "work_unit_id", "event_service.list_by_work_unit"); err != nil {
		return nil, err
	}
	store, err := newEventStore(s.executor)
	if err != nil {
		return nil, err
	}
	return store.ListByWorkUnit(workUnitID)
}

func (s *EventService) ReplayTask(ctx context.Context, taskID string) (*statemachine.ReplayState, error) {
	_ = ctx
	if err := validateRequiredUUID(taskID, "task_id", "event_service.replay_task"); err != nil {
		return nil, err
	}
	store, err := newEventStore(s.executor)
	if err != nil {
		return nil, err
	}
	return store.ReplayState(taskID)
}

func (s *EventService) ReplayRun(ctx context.Context, runID string) (*statemachine.ReplayState, error) {
	_ = ctx
	if err := validateRequiredUUID(runID, "run_id", "event_service.replay_run"); err != nil {
		return nil, err
	}
	store, err := newEventStore(s.executor)
	if err != nil {
		return nil, err
	}
	return store.ReplayRunState(runID)
}

func (s *EventService) LastCheckpointByRun(ctx context.Context, runID string) (*domain.EventEnvelope, error) {
	_ = ctx
	if err := validateRequiredUUID(runID, "run_id", "event_service.last_checkpoint_by_run"); err != nil {
		return nil, err
	}
	store, err := newEventStore(s.executor)
	if err != nil {
		return nil, err
	}
	return store.LastCheckpointByRun(runID)
}

func (s *EventService) validateEnvelopeInput(envelope *domain.EventEnvelope) error {
	op := "event_service.validate_envelope"
	if err := validateRequiredText(envelope.Type, "type", op); err != nil {
		return err
	}
	if err := validateRequiredText(envelope.Version, "version", op); err != nil {
		return err
	}
	if err := validateRequiredUUID(envelope.TaskID, "task_id", op); err != nil {
		return err
	}
	if err := validateOptionalUUID(envelope.ID, "event_id", op); err != nil {
		return err
	}
	if err := validateOptionalUUID(envelope.RunID, "run_id", op); err != nil {
		return err
	}
	if err := validateOptionalUUID(envelope.WorkUnitID, "work_unit_id", op); err != nil {
		return err
	}
	if requiresRunID(envelope.Type) && strings.TrimSpace(envelope.RunID) == "" {
		return apperrors.New(apperrors.CodeInvalidInput, op, "run_id is required for run, agent, and tool events")
	}
	if envelope.Priority != "" {
		switch envelope.Priority {
		case domain.EventPriorityInterrupt, domain.EventPriorityCheckpoint, domain.EventPriorityNotification, domain.EventPriorityBackground:
		default:
			return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("invalid event priority %q", envelope.Priority))
		}
	}
	if len(envelope.Payload) > 0 && !json.Valid(envelope.Payload) {
		return apperrors.New(apperrors.CodeValidation, op, "payload must be valid JSON")
	}
	return nil
}

func (s *EventService) validateReferences(ctx context.Context, envelope *domain.EventEnvelope) error {
	_ = ctx
	if envelope.TaskID != "" {
		taskRepo := repository.NewTaskRepository(s.executor)
		task, err := taskRepo.GetByID(envelope.TaskID)
		if err != nil {
			return apperrors.Wrap(apperrors.CodePersistence, "event_service.validate_task_ref", err)
		}
		if task == nil {
			return apperrors.New(apperrors.CodeNotFound, "event_service.validate_task_ref", "task not found")
		}
	}
	if envelope.WorkUnitID != "" {
		wuRepo := repository.NewWorkUnitRepository(s.executor)
		wu, err := wuRepo.GetByID(envelope.WorkUnitID)
		if err != nil {
			return apperrors.Wrap(apperrors.CodePersistence, "event_service.validate_work_unit_ref", err)
		}
		if wu == nil {
			return apperrors.New(apperrors.CodeNotFound, "event_service.validate_work_unit_ref", "work unit not found")
		}
		if envelope.TaskID != "" && wu.TaskID != envelope.TaskID {
			return apperrors.New(apperrors.CodeInvalidInput, "event_service.validate_work_unit_ref", "work_unit_id does not belong to task_id")
		}
	}
	if envelope.RunID != "" {
		runRepo := repository.NewRunRepository(s.executor)
		run, err := runRepo.GetByID(envelope.RunID)
		if err != nil {
			return apperrors.Wrap(apperrors.CodePersistence, "event_service.validate_run_ref", err)
		}
		if run == nil {
			return apperrors.New(apperrors.CodeNotFound, "event_service.validate_run_ref", "run not found")
		}
		if envelope.TaskID != "" && run.TaskID != envelope.TaskID {
			return apperrors.New(apperrors.CodeInvalidInput, "event_service.validate_run_ref", "run_id does not belong to task_id")
		}
		if envelope.WorkUnitID != "" && run.WorkUnitID != "" && run.WorkUnitID != envelope.WorkUnitID {
			return apperrors.New(apperrors.CodeInvalidInput, "event_service.validate_run_ref", "run_id does not belong to work_unit_id")
		}
	}
	return nil
}

func requiresRunID(eventType string) bool {
	return strings.HasPrefix(eventType, "run.") || strings.HasPrefix(eventType, "agent.") || strings.HasPrefix(eventType, "tool.")
}

func sameEventIntent(existing *domain.EventEnvelope, candidate *domain.EventEnvelope) bool {
	if existing == nil || candidate == nil {
		return false
	}
	existingPriority := existing.Priority
	if existingPriority == "" {
		existingPriority = domain.EventPriorityBackground
	}
	candidatePriority := candidate.Priority
	if candidatePriority == "" {
		candidatePriority = domain.EventPriorityBackground
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
