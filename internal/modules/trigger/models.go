package trigger

import (
	"encoding/json"
	"time"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusTriggered Status = "triggered"
	StatusResolved  Status = "resolved"
	StatusDismissed Status = "dismissed"
)

type Type string

const (
	TypeThreshold        Type = "threshold"
	TypeAnomaly          Type = "anomaly"
	TypeHeartbeatTimeout Type = "heartbeat_timeout"
	TypePolicy           Type = "policy"
)

type AnomalyType string

const (
	AnomalyStall         AnomalyType = "stall"
	AnomalyLoop          AnomalyType = "loop"
	AnomalyDrift         AnomalyType = "drift"
	AnomalyPathViolation AnomalyType = "path_violation"
	AnomalyTokenExceeded AnomalyType = "token_exceeded"
	AnomalyStepsExceeded AnomalyType = "steps_exceeded"
	AnomalyTimeExceeded  AnomalyType = "time_exceeded"
)

type ResolutionAction string

const (
	ResolutionPause    ResolutionAction = "pause"
	ResolutionCancel   ResolutionAction = "cancel"
	ResolutionNotify   ResolutionAction = "notify"
	ResolutionEscalate ResolutionAction = "escalate"
)

type ThresholdConfig struct {
	StallSeconds    int `json:"stall_seconds"`
	LoopRepetitions int `json:"loop_repetitions"`
	TokenMax        int `json:"token_max"`
	StepsMax        int `json:"steps_max"`
	TimeMaxSeconds  int `json:"time_max_seconds"`
}

type Trigger struct {
	ID               string            `json:"id"`
	RunID            *string           `json:"run_id,omitempty"`
	TaskID           *string           `json:"task_id,omitempty"`
	AgentSessionID   *string           `json:"agent_session_id,omitempty"`
	TriggerType      Type              `json:"trigger_type"`
	Status           Status            `json:"status"`
	AnomalyType      *AnomalyType      `json:"anomaly_type,omitempty"`
	ThresholdValue   json.RawMessage   `json:"threshold_value,omitempty"`
	CurrentValue     json.RawMessage   `json:"current_value,omitempty"`
	TriggeredAt      *time.Time        `json:"triggered_at,omitempty"`
	ResolvedAt       *time.Time        `json:"resolved_at,omitempty"`
	ResolutionAction *ResolutionAction `json:"resolution_action,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
}
