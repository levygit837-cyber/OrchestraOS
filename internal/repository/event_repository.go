package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// EventRepository handles event persistence
type EventRepository struct {
	db *sql.DB
}

// NewEventRepository creates a new event repository
func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Create inserts a new event
func (r *EventRepository) Create(event *domain.EventEnvelope) error {
	if event == nil {
		return apperrors.New(apperrors.CodeInvalidInput, "event_repository.create", "event envelope is required")
	}
	if event.ID == "" || event.Sequence == 0 || event.CreatedAt.IsZero() {
		return apperrors.New(apperrors.CodeInvalidInput, "event_repository.create", "event envelope must be completed before persistence")
	}

	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = r.db.Exec(
		db.QueryEventInsert,
		event.ID,
		event.Type,
		event.Version,
		nullString(event.TaskID),
		nullString(event.RunID),
		nullString(event.WorkUnitID),
		nullString(event.AgentID),
		nullString(event.TraceID),
		nullString(event.SpanID),
		nullString(event.ParentSpanID),
		event.Sequence,
		event.Priority,
		event.RequiresAck,
		event.CreatedAt,
		payloadJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// GetNextSequence returns the next event sequence number
func (r *EventRepository) GetNextSequence() (int64, error) {
	var seq int64
	err := r.db.QueryRow(db.QueryEventNextSequence).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence: %w", err)
	}
	return seq, nil
}

// GetByID retrieves an event by ID
func (r *EventRepository) GetByID(id string) (*domain.EventEnvelope, error) {
	row := r.db.QueryRow(db.QueryEventGetByID, id)
	return r.scanEvent(row)
}

// List retrieves all events ordered by sequence
func (r *EventRepository) List() ([]domain.EventEnvelope, error) {
	rows, err := r.db.Query(db.QueryEventList)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []domain.EventEnvelope
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}

	return events, rows.Err()
}

// ListByTask retrieves all events for a task
func (r *EventRepository) ListByTask(taskID string) ([]domain.EventEnvelope, error) {
	rows, err := r.db.Query(db.QueryEventListByTask, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []domain.EventEnvelope
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}

	return events, rows.Err()
}

// ListByRun retrieves all events for a run
func (r *EventRepository) ListByRun(runID string) ([]domain.EventEnvelope, error) {
	rows, err := r.db.Query(db.QueryEventListByRun, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []domain.EventEnvelope
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}

	return events, rows.Err()
}

// ListByWorkUnit retrieves all events for a work unit
func (r *EventRepository) ListByWorkUnit(workUnitID string) ([]domain.EventEnvelope, error) {
	rows, err := r.db.Query(db.QueryEventListByWorkUnit, workUnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()

	var events []domain.EventEnvelope
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}

	return events, rows.Err()
}

func (r *EventRepository) scanEvent(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.EventEnvelope, error) {
	var event domain.EventEnvelope
	var taskID, runID, workUnitID, agentID, traceID, spanID, parentSpanID sql.NullString
	var payloadJSON []byte

	err := scanner.Scan(
		&event.ID,
		&event.Type,
		&event.Version,
		&taskID,
		&runID,
		&workUnitID,
		&agentID,
		&traceID,
		&spanID,
		&parentSpanID,
		&event.Sequence,
		&event.Priority,
		&event.RequiresAck,
		&event.CreatedAt,
		&payloadJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan event: %w", err)
	}

	if taskID.Valid {
		event.TaskID = taskID.String
	}
	if runID.Valid {
		event.RunID = runID.String
	}
	if workUnitID.Valid {
		event.WorkUnitID = workUnitID.String
	}
	if agentID.Valid {
		event.AgentID = agentID.String
	}
	if traceID.Valid {
		event.TraceID = traceID.String
	}
	if spanID.Valid {
		event.SpanID = spanID.String
	}
	if parentSpanID.Valid {
		event.ParentSpanID = parentSpanID.String
	}

	if len(payloadJSON) > 0 {
		if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}
	}

	return &event, nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
