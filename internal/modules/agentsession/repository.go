// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package agentsession

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/db"
)

// Repository handles agent session persistence
type Repository struct {
	db db.DBTX
}

// NewRepository creates a new agent session repository
func NewRepository(db db.DBTX) *Repository {
	return &Repository{db: db}
}

// Create inserts a new agent session
func (r *Repository) Create(session *AgentSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	now := time.Now()

	_, err := r.db.Exec(
		QueryInsert,
		session.ID,
		session.AgentID,
		session.RunID,
		session.TaskID,
		session.WorkUnitID,
		session.SandboxID,
		session.ConnectionID,
		session.Status,
		session.LastHeartbeatAt,
		session.LastCheckpointAt,
		nullString(session.LastSeenEventID),
		nullableRawJSON(session.RecoverableState),
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to create agent session: %w", err)
	}

	return nil
}

// GetByID retrieves an agent session by ID
func (r *Repository) GetByID(id string) (*AgentSession, error) {
	row := r.db.QueryRow(QueryGetByID, id)
	return r.scanAgentSession(row)
}

// GetByRunID retrieves the most recent agent session for a run
func (r *Repository) GetByRunID(runID string) (*AgentSession, error) {
	row := r.db.QueryRow(QueryGetByRunID, runID)
	return r.scanAgentSession(row)
}

// UpdateStatus updates agent session status and timestamps
func (r *Repository) UpdateStatus(id string, status Status) error {
	var heartbeatAt, checkpointAt *time.Time
	now := time.Now()

	if status == StatusRunning {
		heartbeatAt = &now
	}

	_, err := r.db.Exec(
		QueryUpdateStatus,
		id,
		status,
		heartbeatAt,
		checkpointAt,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update agent session status: %w", err)
	}

	return nil
}

// UpdateHeartbeat updates the last heartbeat timestamp
func (r *Repository) UpdateHeartbeat(id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		QueryUpdateHeartbeat,
		id,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}
	return nil
}

// UpdateHeartbeat updates the last heartbeat timestamp and optional last seen event.
func (r *Repository) UpdateHeartbeatWithEvent(id, lastSeenEventID string) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		QueryUpdateHeartbeatWithEvent,
		id,
		now,
		nullString(lastSeenEventID),
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}
	return nil
}

// UpdateCheckpoint updates the last checkpoint timestamp
func (r *Repository) UpdateCheckpoint(id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		QueryUpdateCheckpoint,
		id,
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update checkpoint: %w", err)
	}
	return nil
}

// UpdateCheckpointWithEvent updates the last checkpoint timestamp and optional last seen event.
func (r *Repository) UpdateCheckpointWithEvent(id, lastSeenEventID string) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		QueryUpdateCheckpointWithEvent,
		id,
		now,
		nullString(lastSeenEventID),
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update checkpoint: %w", err)
	}
	return nil
}

// UpdateRecoverableState stores resume context for a disconnected or timed-out session.
func (r *Repository) UpdateRecoverableState(id string, state json.RawMessage) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		QueryUpdateRecoverableState,
		id,
		nullableRawJSON(state),
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update recoverable state: %w", err)
	}
	return nil
}

func (r *Repository) scanAgentSession(scanner interface {
	Scan(dest ...interface{}) error
}) (*AgentSession, error) {
	var session AgentSession
	var sandboxID, connectionID, lastSeenEventID sql.NullString
	var recoverableState []byte
	var createdAt, updatedAt time.Time

	var taskID, workUnitID sql.NullString
	err := scanner.Scan(
		&session.ID,
		&session.AgentID,
		&session.RunID,
		&taskID,
		&workUnitID,
		&sandboxID,
		&connectionID,
		&session.Status,
		&session.LastHeartbeatAt,
		&session.LastCheckpointAt,
		&lastSeenEventID,
		&recoverableState,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan agent session: %w", err)
	}

	if taskID.Valid {
		session.TaskID = taskID.String
	}
	if workUnitID.Valid {
		session.WorkUnitID = workUnitID.String
	}
	if sandboxID.Valid {
		session.SandboxID = sandboxID.String
	}
	if connectionID.Valid {
		session.ConnectionID = connectionID.String
	}
	if lastSeenEventID.Valid {
		session.LastSeenEventID = lastSeenEventID.String
	}
	if len(recoverableState) > 0 {
		session.RecoverableState = json.RawMessage(recoverableState)
	}

	return &session, nil
}

func nullableRawJSON(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	return raw
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
