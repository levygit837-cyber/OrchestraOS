package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/db"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// AgentSessionRepository handles agent session persistence
type AgentSessionRepository struct {
	db *sql.DB
}

// NewAgentSessionRepository creates a new agent session repository
func NewAgentSessionRepository(db *sql.DB) *AgentSessionRepository {
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

func (r *AgentSessionRepository) scanAgentSession(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.AgentSession, error) {
	var session domain.AgentSession
	var sandboxID, connectionID sql.NullString
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

	return &session, nil
}
