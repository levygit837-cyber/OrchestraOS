package agentsession

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusStarting        Status = "starting"
	StatusRunning         Status = "running"
	StatusWaitingApproval Status = "waiting_approval"
	StatusPaused          Status = "paused"
	StatusStopping        Status = "stopping"
	StatusStopped         Status = "stopped"
	StatusDisconnected    Status = "disconnected"
	StatusFailed          Status = "failed"
)

type AgentSession struct {
	ID               string          `json:"id"`
	AgentID          string          `json:"agent_id"`
	RunID            string          `json:"run_id"`
	TaskID           string          `json:"task_id"`
	WorkUnitID       string          `json:"work_unit_id"`
	SandboxID        string          `json:"sandbox_id"`
	ConnectionID     string          `json:"connection_id"`
	Status           Status          `json:"status"`
	LastHeartbeatAt  *time.Time      `json:"last_heartbeat_at,omitempty"`
	LastCheckpointAt *time.Time      `json:"last_checkpoint_at,omitempty"`
	LastSeenEventID  string          `json:"last_seen_event_id,omitempty"`
	RecoverableState json.RawMessage `json:"recoverable_state,omitempty"`
}
