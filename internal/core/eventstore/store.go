package eventstore

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	dbcore "github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/statemachine"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Store handles event storage and retrieval with validation
type Store struct {
	repo      *Repository
	validator *Validator
}

// NewStore creates a new event store
func NewStore(db *sql.DB) (*Store, error) {
	return NewStoreWithExecutor(db)
}

// NewStoreWithExecutor creates a store bound to a DB or transaction executor.
func NewStoreWithExecutor(executor dbcore.DBTX) (*Store, error) {
	validator, err := NewValidator()
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeInternal, "eventstore.new_validator", err)
	}

	repo := NewRepository(executor)

	return &Store{
		repo:      repo,
		validator: validator,
	}, nil
}

// Append validates and stores a new event
func (s *Store) Append(envelope *domain.EventEnvelope) error {
	_, _, err := s.AppendResult(envelope)
	return err
}

// AppendResult validates and stores a new event, returning the persisted event
// and whether the operation was an idempotent duplicate.
func (s *Store) AppendResult(envelope *domain.EventEnvelope) (*domain.EventEnvelope, bool, error) {
	if err := s.completeEnvelopeBeforeValidation(envelope); err != nil {
		return nil, false, err
	}

	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodeValidation, "eventstore.marshal_envelope", err)
	}

	if err := s.validator.Validate(envelopeBytes); err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodeValidation, "eventstore.validate_envelope", err)
	}
	if err := ValidateOperationalPayload(envelope); err != nil {
		return nil, false, err
	}

	inserted, err := s.repo.Create(envelope)
	if err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodePersistence, "eventstore.store_event", err)
	}
	if inserted {
		return envelope, false, nil
	}

	existing, err := s.repo.GetByID(envelope.ID)
	if err != nil {
		return nil, false, apperrors.Wrap(apperrors.CodePersistence, "eventstore.get_duplicate", err)
	}
	if existing == nil {
		return nil, false, apperrors.New(apperrors.CodeConflict, "eventstore.idempotency", "event_id conflict did not return an existing event")
	}
	if !sameEventIntent(existing, envelope) {
		return nil, false, apperrors.New(apperrors.CodeConflict, "eventstore.idempotency", "event_id already exists with different event content")
	}
	*envelope = *existing
	return existing, true, nil
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
	seq, err := s.repo.GetNextSequence()
	if err != nil {
		return apperrors.Wrap(apperrors.CodePersistence, "eventstore.next_sequence", err)
	}
	envelope.Sequence = seq
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

// GetLastCheckpointByRun retrieves the latest checkpoint event for a run.
func (s *Store) GetLastCheckpointByRun(runID string) (*domain.EventEnvelope, error) {
	return s.repo.GetLastCheckpointByRun(runID)
}

// Replay reconstructs state by replaying events for a task
func (s *Store) Replay(taskID string) ([]domain.EventEnvelope, error) {
	events, err := s.repo.ListByTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to replay events: %w", err)
	}
	return events, nil
}

// ReplayState reconstructs aggregate read-model state from task events.
func (s *Store) ReplayState(taskID string) (*statemachine.ReplayState, error) {
	events, err := s.Replay(taskID)
	if err != nil {
		return nil, err
	}
	state, err := statemachine.ProjectStrict(events)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// ReplayRunState reconstructs run-scoped state from run events.
func (s *Store) ReplayRunState(runID string) (*statemachine.ReplayState, error) {
	events, err := s.repo.ListByRun(runID)
	if err != nil {
		return nil, fmt.Errorf("failed to replay run events: %w", err)
	}
	state, err := statemachine.ProjectStrict(events)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func ValidateOperationalPayload(envelope *domain.EventEnvelope) error {
	switch envelope.Type {
	case "task.graph_created":
		var payload domain.TaskGraphCreatedPayload
		if err := decodePayload(envelope.Payload, &payload); err != nil {
			return err
		}
		if payload.TaskID == "" || payload.GraphID == "" || payload.GraphVersion < 1 || payload.PlannerStrategy == "" || len(payload.Nodes) == 0 || payload.Edges == nil {
			return payloadError(envelope.Type, "task_id, graph_id, graph_version, planner_strategy, nodes, and edges are required")
		}
		if payload.TaskID != envelope.TaskID {
			return payloadError(envelope.Type, "payload task_id must match envelope task_id")
		}
	case "agent.ledger_updated":
		var payload domain.AgentLedgerUpdatedPayload
		if err := decodePayload(envelope.Payload, &payload); err != nil {
			return err
		}
		if len(payload.Ledger) == 0 {
			return payloadError(envelope.Type, "ledger is required")
		}
	case "agent.checkpoint_reached":
		var payload domain.AgentCheckpointReachedPayload
		if err := decodePayload(envelope.Payload, &payload); err != nil {
			return err
		}
		if payload.CheckpointID == "" || payload.CurrentGoal == "" || payload.MinimalSummary == "" || len(payload.Ledger) == 0 {
			return payloadError(envelope.Type, "checkpoint_id, current_goal, ledger, and minimal_summary are required")
		}
	case "artifact.created":
		var payload domain.ArtifactCreatedPayload
		if err := decodePayload(envelope.Payload, &payload); err != nil {
			return err
		}
		if payload.ArtifactID == "" || payload.Kind == "" || payload.URI == "" {
			return payloadError(envelope.Type, "artifact_id, kind, and uri are required")
		}
	case "validation.completed":
		var payload domain.ValidationCompletedPayload
		if err := decodePayload(envelope.Payload, &payload); err != nil {
			return err
		}
		if payload.ValidationID == "" || payload.Status == "" {
			return payloadError(envelope.Type, "validation_id and status are required")
		}
	case "prompt.snapshot_created":
		var payload domain.PromptSnapshotCreatedPayload
		if err := decodePayload(envelope.Payload, &payload); err != nil {
			return err
		}
		if payload.PromptSnapshotID == "" || payload.Hash == "" {
			return payloadError(envelope.Type, "prompt_snapshot_id and hash are required")
		}
	case "toolset.snapshot_created":
		var payload domain.ToolsetSnapshotCreatedPayload
		if err := decodePayload(envelope.Payload, &payload); err != nil {
			return err
		}
		if payload.ToolsetSnapshotID == "" || payload.AgentSessionID == "" {
			return payloadError(envelope.Type, "toolset_snapshot_id and agent_session_id are required")
		}
	}
	return nil
}

func decodePayload(payload json.RawMessage, target interface{}) error {
	if err := json.Unmarshal(payload, target); err != nil {
		return apperrors.Wrap(apperrors.CodeValidation, "eventstore.validate_payload", err)
	}
	return nil
}

func payloadError(eventType, message string) error {
	return apperrors.New(apperrors.CodeValidation, "eventstore.validate_payload", fmt.Sprintf("%s: %s", eventType, message))
}

func sameEventIntent(existing *domain.EventEnvelope, candidate *domain.EventEnvelope) bool {
	if existing == nil || candidate == nil {
		return false
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
		existing.Priority != candidate.Priority ||
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
