package domain

import (
	"encoding/json"
	"time"
)

// ============================================================================
// Trigger Domain
// ============================================================================

type TriggerStatus string

const (
	TriggerStatusActive    TriggerStatus = "active"
	TriggerStatusTriggered TriggerStatus = "triggered"
	TriggerStatusResolved  TriggerStatus = "resolved"
	TriggerStatusDismissed TriggerStatus = "dismissed"
)

type TriggerType string

const (
	TriggerTypeThreshold        TriggerType = "threshold"
	TriggerTypeAnomaly          TriggerType = "anomaly"
	TriggerTypeHeartbeatTimeout TriggerType = "heartbeat_timeout"
	TriggerTypePolicy           TriggerType = "policy"
)

type Trigger struct {
	ID               string          `json:"id"`
	RunID            *string         `json:"run_id,omitempty"`
	TaskID           *string         `json:"task_id,omitempty"`
	AgentSessionID   *string         `json:"agent_session_id,omitempty"`
	TriggerType      TriggerType     `json:"trigger_type"`
	Status           TriggerStatus   `json:"status"`
	AnomalyType      *string         `json:"anomaly_type,omitempty"`
	ThresholdValue   json.RawMessage `json:"threshold_value,omitempty"`
	CurrentValue     json.RawMessage `json:"current_value,omitempty"`
	TriggeredAt      *time.Time      `json:"triggered_at,omitempty"`
	ResolvedAt       *time.Time      `json:"resolved_at,omitempty"`
	ResolutionAction *string         `json:"resolution_action,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}
