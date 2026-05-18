package domain

import (
	"encoding/json"
	"time"
)

type EventPriority string

const (
	EventPriorityInterrupt    EventPriority = "interrupt"
	EventPriorityCheckpoint   EventPriority = "checkpoint"
	EventPriorityNotification EventPriority = "notification"
	EventPriorityBackground   EventPriority = "background"
)

type EventEnvelope struct {
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	Version      string          `json:"version"`
	TaskID       string          `json:"task_id"`
	RunID        string          `json:"run_id,omitempty"`
	WorkUnitID   string          `json:"work_unit_id,omitempty"`
	AgentID      string          `json:"agent_id,omitempty"`
	TraceID      string          `json:"trace_id,omitempty"`
	SpanID       string          `json:"span_id,omitempty"`
	ParentSpanID string          `json:"parent_span_id,omitempty"`
	Sequence     int64           `json:"sequence"`
	Priority     EventPriority   `json:"priority"`
	RequiresAck  bool            `json:"requires_ack"`
	CreatedAt    time.Time       `json:"created_at"`
	Payload      json.RawMessage `json:"payload"`
}
