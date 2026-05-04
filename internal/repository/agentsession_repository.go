package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// AgentSessionRepository handles agent session persistence
type AgentSessionRepository struct {
	db DBTX
}

// NewAgentSessionRepository creates a new agent session repository
func NewAgentSessionRepository(db DBTX) *AgentSessionRepository {
	return &AgentSessionRepository{db: db}
}

// Create inserts a new agent session
func (r *AgentSessionRepository) Create(session *domain.AgentSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	now := time.Now()

	_, err := r.db.Exec(
		db.QueryAgentSessionInsert,
		session.ID,
		session.AgentID,
		session.RunID,
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
func (r *AgentSessionRepository) GetByID(id string) (*domain.AgentSession, error) {
	row := r.db.QueryRow(db.QueryAgentSessionGetByID, id)
	return r.scanAgentSession(row)
}

// GetByRunID retrieves the most recent agent session for a run
func (r *AgentSessionRepository) GetByRunID(runID string) (*domain.AgentSession, error) {
	row := r.db.QueryRow(db.QueryAgentSessionGetByRunID, runID)
	return r.scanAgentSession(row)
}

// UpdateStatus updates agent session status and timestamps
func (r *AgentSessionRepository) UpdateStatus(id string, status domain.AgentSessionStatus) error {
	var heartbeatAt, checkpointAt *time.Time
	now := time.Now()

	if status == domain.AgentSessionStatusRunning {
		heartbeatAt = &now
	}

	_, err := r.db.Exec(
		db.QueryAgentSessionUpdateStatus,
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
func (r *AgentSessionRepository) UpdateHeartbeat(id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE agent_sessions SET last_heartbeat_at = $2, updated_at = $3 WHERE id = $1`,
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
func (r *AgentSessionRepository) UpdateHeartbeatWithEvent(id, lastSeenEventID string) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE agent_sessions SET last_heartbeat_at = $2, last_seen_event_id = COALESCE($3, last_seen_event_id), updated_at = $4 WHERE id = $1`,
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
func (r *AgentSessionRepository) UpdateCheckpoint(id string) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE agent_sessions SET last_checkpoint_at = $2, updated_at = $3 WHERE id = $1`,
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
func (r *AgentSessionRepository) UpdateCheckpointWithEvent(id, lastSeenEventID string) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE agent_sessions SET last_checkpoint_at = $2, last_seen_event_id = COALESCE($3, last_seen_event_id), updated_at = $4 WHERE id = $1`,
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
func (r *AgentSessionRepository) UpdateRecoverableState(id string, state json.RawMessage) error {
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`UPDATE agent_sessions SET recoverable_state = $2, updated_at = $3 WHERE id = $1`,
		id,
		nullableRawJSON(state),
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to update recoverable state: %w", err)
	}
	return nil
}

func (r *AgentSessionRepository) scanAgentSession(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.AgentSession, error) {
	var session domain.AgentSession
	var sandboxID, connectionID, lastSeenEventID sql.NullString
	var recoverableState []byte
	var createdAt, updatedAt time.Time

	err := scanner.Scan(
		&session.ID,
		&session.AgentID,
		&session.RunID,
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
