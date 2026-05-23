package domain

import (
	"encoding/json"
	"time"
)

// ============================================================================
// Agentsession Domain
// ============================================================================

type AgentSessionStatus string

const (
	AgentSessionStatusStarting        AgentSessionStatus = "starting"
	AgentSessionStatusRunning         AgentSessionStatus = "running"
	AgentSessionStatusWaitingApproval AgentSessionStatus = "waiting_approval"
	AgentSessionStatusPaused          AgentSessionStatus = "paused"
	AgentSessionStatusStopping        AgentSessionStatus = "stopping"
	AgentSessionStatusStopped         AgentSessionStatus = "stopped"
	AgentSessionStatusDisconnected    AgentSessionStatus = "disconnected"
	AgentSessionStatusFailed          AgentSessionStatus = "failed"
)

type AgentSession struct {
	ID               string             `json:"id"`
	AgentID          string             `json:"agent_id"`
	RunID            string             `json:"run_id"`
	TaskID           string             `json:"task_id"`
	WorkUnitID       string             `json:"work_unit_id"`
	SandboxID        string             `json:"sandbox_id"`
	ConnectionID     string             `json:"connection_id"`
	Status           AgentSessionStatus `json:"status"`
	LastHeartbeatAt  *time.Time         `json:"last_heartbeat_at,omitempty"`
	LastCheckpointAt *time.Time         `json:"last_checkpoint_at,omitempty"`
	LastSeenEventID  string             `json:"last_seen_event_id,omitempty"`
	RecoverableState json.RawMessage    `json:"recoverable_state,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}
