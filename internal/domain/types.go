package domain

import (
	"encoding/json"
	"time"
)

type TaskStatus string

const (
	TaskStatusCreated          TaskStatus = "created"
	TaskStatusTriaged          TaskStatus = "triaged"
	TaskStatusPlanned          TaskStatus = "planned"
	TaskStatusScheduled        TaskStatus = "scheduled"
	TaskStatusSandboxPreparing TaskStatus = "sandbox_preparing"
	TaskStatusRunning          TaskStatus = "running"
	TaskStatusWaitingApproval  TaskStatus = "waiting_approval"
	TaskStatusPaused           TaskStatus = "paused"
	TaskStatusValidating       TaskStatus = "validating"
	TaskStatusCompleted        TaskStatus = "completed"
	TaskStatusFailed           TaskStatus = "failed"
	TaskStatusCancelled        TaskStatus = "cancelled"
)

type Priority string

const (
	PriorityP0 Priority = "P0"
	PriorityP1 Priority = "P1"
	PriorityP2 Priority = "P2"
	PriorityP3 Priority = "P3"
)

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

type Task struct {
	ID                   string     `json:"id"`
	Title                string     `json:"title"`
	Description          string     `json:"description"`
	Status               TaskStatus `json:"status"`
	Priority             Priority   `json:"priority"`
	RiskLevel            RiskLevel  `json:"risk_level"`
	CreatedFromMessageID string     `json:"created_from_message_id"`
	AcceptanceCriteria   []string   `json:"acceptance_criteria"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type WorkUnitStatus string

const (
	WorkUnitStatusCreated         WorkUnitStatus = "created"
	WorkUnitStatusPlanned         WorkUnitStatus = "planned"
	WorkUnitStatusScheduled       WorkUnitStatus = "scheduled"
	WorkUnitStatusBlocked         WorkUnitStatus = "blocked"
	WorkUnitStatusRunning         WorkUnitStatus = "running"
	WorkUnitStatusWaitingApproval WorkUnitStatus = "waiting_approval"
	WorkUnitStatusPaused          WorkUnitStatus = "paused"
	WorkUnitStatusValidating      WorkUnitStatus = "validating"
	WorkUnitStatusCompleted       WorkUnitStatus = "completed"
	WorkUnitStatusFailed          WorkUnitStatus = "failed"
	WorkUnitStatusCancelled       WorkUnitStatus = "cancelled"
)

type WorkUnit struct {
	ID                   string         `json:"id"`
	TaskGraphID          string         `json:"task_graph_id"`
	Title                string         `json:"title"`
	Objective            string         `json:"objective"`
	AssignedAgentProfile string         `json:"assigned_agent_profile"`
	Status               WorkUnitStatus `json:"status"`
	OwnedPaths           []string       `json:"owned_paths"`
	ReadPaths            []string       `json:"read_paths"`
	AcceptanceCriteria   []string       `json:"acceptance_criteria"`
	ValidationPlan       []string       `json:"validation_plan"`
	DependsOn            []string       `json:"depends_on"`
}

type RunStatus string

const (
	RunStatusCreated         RunStatus = "created"
	RunStatusRunning         RunStatus = "running"
	RunStatusWaitingApproval RunStatus = "waiting_approval"
	RunStatusPaused          RunStatus = "paused"
	RunStatusValidating      RunStatus = "validating"
	RunStatusCompleted       RunStatus = "completed"
	RunStatusFailed          RunStatus = "failed"
	RunStatusCancelled       RunStatus = "cancelled"
)

type RunResult string

const (
	RunResultSucceeded RunResult = "succeeded"
	RunResultFailed    RunResult = "failed"
	RunResultCancelled RunResult = "cancelled"
)

type Run struct {
	ID            string     `json:"id"`
	TaskID        string     `json:"task_id"`
	WorkUnitID    string     `json:"work_unit_id"`
	Status        RunStatus  `json:"status"`
	Attempt       int        `json:"attempt"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	Result        *RunResult `json:"result"`
	FailureReason *string    `json:"failure_reason"`
}

type AgentRuntimeType string

const (
	AgentRuntimeTypeCodexCLI AgentRuntimeType = "codex_cli"
	AgentRuntimeTypeFake     AgentRuntimeType = "fake"
	AgentRuntimeTypeExternal AgentRuntimeType = "external"
)

type Agent struct {
	ID                     string           `json:"id"`
	Name                   string           `json:"name"`
	Profile                string           `json:"profile"`
	Capabilities           []string         `json:"capabilities"`
	AllowedTools           []string         `json:"allowed_tools"`
	DefaultPromptFragments []string         `json:"default_prompt_fragments"`
	RuntimeType            AgentRuntimeType `json:"runtime_type"`
}

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
	SandboxID        string             `json:"sandbox_id"`
	ConnectionID     string             `json:"connection_id"`
	Status           AgentSessionStatus `json:"status"`
	LastHeartbeatAt  *time.Time         `json:"last_heartbeat_at"`
	LastCheckpointAt *time.Time         `json:"last_checkpoint_at"`
	LastSeenEventID  string             `json:"last_seen_event_id,omitempty"`
	RecoverableState json.RawMessage    `json:"recoverable_state,omitempty"`
}

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
