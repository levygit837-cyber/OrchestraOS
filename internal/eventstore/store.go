package eventstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
)

// Store handles event storage and retrieval with validation
type Store struct {
	repo      *repository.EventRepository
	validator *Validator
}

// NewStore creates a new event store
func NewStore(db *sql.DB) (*Store, error) {
	validator, err := NewValidator()
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "eventstore.new_validator", err)
	}

	repo := repository.NewEventRepository(db)

	return &Store{
		repo:      repo,
		validator: validator,
	}, nil
}

// Append validates and stores a new event
func (s *Store) Append(envelope *domain.EventEnvelope) error {
	if err := s.completeEnvelopeBeforeValidation(envelope); err != nil {
		return err
	}

	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeValidation, "eventstore.marshal_envelope", err)
	}

	if err := s.validator.Validate(envelopeBytes); err != nil {
		return apperrors.Wrap(apperrors.CodeValidation, "eventstore.validate_envelope", err)
	}

	if err := s.repo.Create(envelope); err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "eventstore.store_event", err)
	}

	return nil
}

// completeEnvelopeBeforeValidation owns all generated envelope fields.
// The repository only persists already complete envelopes, so this must run
// before the JSON Schema validator.
func (s *Store) completeEnvelopeBeforeValidation(envelope *domain.EventEnvelope) error {
	if envelope == nil {
		return apperrors.New(apperrors.CodeInvalidInput, "eventstore.complete_envelope", "event envelope is required")
	}
	if envelope.ID == "" {
		envelope.ID = uuid.New().String()
	}
	if envelope.CreatedAt.IsZero() {
		envelope.CreatedAt = time.Now().UTC()
	}
	if envelope.Sequence == 0 {
		seq, err := s.repo.GetNextSequence()
		if err != nil {
			return apperrors.Wrap(apperrors.CodePersistence, "eventstore.next_sequence", err)
		}
		envelope.Sequence = seq
	}
	if envelope.Priority == "" {
		envelope.Priority = domain.EventPriorityBackground
	}
	if len(envelope.Payload) == 0 {
		envelope.Payload = json.RawMessage(`{}`)
	}
	return nil
}

// AppendRaw validates and stores a raw event
func (s *Store) AppendRaw(eventType, version string, payload interface{}, taskID, runID string) (*domain.EventEnvelope, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, "eventstore.marshal_payload", err)
	}

	envelope := &domain.EventEnvelope{
		Type:    eventType,
		Version: version,
		TaskID:  taskID,
		RunID:   runID,
		Payload: payloadBytes,
	}

	if err := s.Append(envelope); err != nil {
		return nil, err
	}

	return envelope, nil
}

// Get retrieves an event by ID
func (s *Store) Get(id string) (*domain.EventEnvelope, error) {
	return s.repo.GetByID(id)
}

// List retrieves all events
func (s *Store) List() ([]domain.EventEnvelope, error) {
	return s.repo.List()
}

// ListByTask retrieves all events for a task
func (s *Store) ListByTask(taskID string) ([]domain.EventEnvelope, error) {
	return s.repo.ListByTask(taskID)
}

// ListByRun retrieves all events for a run
func (s *Store) ListByRun(runID string) ([]domain.EventEnvelope, error) {
	return s.repo.ListByRun(runID)
}

// ListByWorkUnit retrieves all events for a work unit
func (s *Store) ListByWorkUnit(workUnitID string) ([]domain.EventEnvelope, error) {
	return s.repo.ListByWorkUnit(workUnitID)
}

// Replay reconstructs state by replaying events for a task
func (s *Store) Replay(taskID string) ([]domain.EventEnvelope, error) {
	events, err := s.repo.ListByTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to replay events: %w", err)
	}
	return events, nil
}
