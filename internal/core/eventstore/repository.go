package eventstore

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type Repository struct {
	db db.DBTX
}

func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(event *domain.EventEnvelope) (bool, error) {
	if event == nil {
		return false, apperrors.New(apperrors.CodeInvalidInput, "event_repository.create", "event envelope is required")
	}
	if event.ID == "" || event.Sequence == 0 || event.CreatedAt.IsZero() {
		return false, apperrors.New(apperrors.CodeInvalidInput, "event_repository.create", "event envelope must be completed before persistence")
	}

	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	result, err := r.db.Exec(
		QueryInsert,
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
		return false, fmt.Errorf("failed to create event: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to inspect event insert result: %w", err)
	}

	return rows > 0, nil
}

func (r *Repository) GetNextSequence() (int64, error) {
	var seq int64
	err := r.db.QueryRow(QueryNextSequence).Scan(&seq)
	if err != nil {
		return 0, fmt.Errorf("failed to get next sequence: %w", err)
	}
	return seq, nil
}

func (r *Repository) GetByID(id string) (*domain.EventEnvelope, error) {
	row := r.db.QueryRow(QueryGetByID, id)
	return r.scanEvent(row)
}

func (r *Repository) List() ([]domain.EventEnvelope, error) {
	rows, err := r.db.Query(QueryList)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer func() { _ = rows.Close() }()

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

func (r *Repository) ListByTask(taskID string) ([]domain.EventEnvelope, error) {
	rows, err := r.db.Query(QueryListByTask, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer func() { _ = rows.Close() }()

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

func (r *Repository) ListByRun(runID string) ([]domain.EventEnvelope, error) {
	rows, err := r.db.Query(QueryListByRun, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer func() { _ = rows.Close() }()

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

func (r *Repository) ListByWorkUnit(workUnitID string) ([]domain.EventEnvelope, error) {
	rows, err := r.db.Query(QueryListByWorkUnit, workUnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}
	defer func() { _ = rows.Close() }()

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

func (r *Repository) LastCheckpointByRun(runID string) (*domain.EventEnvelope, error) {
	row := r.db.QueryRow(QueryLastCheckpointByRun, runID)
	return r.scanEvent(row)
}

func (r *Repository) scanEvent(scanner interface {
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
