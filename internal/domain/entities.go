package domain

import (
	"encoding/json"
	"time"
)

// ============================================================================
// Task Domain
// ============================================================================

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

type TaskPriority string

const (
	TaskPriorityP0 TaskPriority = "P0"
	TaskPriorityP1 TaskPriority = "P1"
	TaskPriorityP2 TaskPriority = "P2"
	TaskPriorityP3 TaskPriority = "P3"
)

type TaskRiskLevel string

const (
	TaskRiskLevelLow      TaskRiskLevel = "low"
	TaskRiskLevelMedium   TaskRiskLevel = "medium"
	TaskRiskLevelHigh     TaskRiskLevel = "high"
	TaskRiskLevelCritical TaskRiskLevel = "critical"
)

type Task struct {
	ID                   string        `json:"id"`
	Title                string        `json:"title"`
	Description          string        `json:"description"`
	Status               TaskStatus    `json:"status"`
	Priority             TaskPriority  `json:"priority"`
	RiskLevel            TaskRiskLevel `json:"risk_level"`
	CreatedFromMessageID string        `json:"created_from_message_id"`
	AcceptanceCriteria   []string      `json:"acceptance_criteria"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

// ============================================================================
// Run Domain
// ============================================================================

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
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	Result        *RunResult `json:"result,omitempty"`
	FailureReason *string    `json:"failure_reason,omitempty"`
}

// ============================================================================
// WorkUnit Domain
// ============================================================================

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
	ID                   string   `json:"id"`
	TaskID               string   `json:"task_id"`
	TaskGraphID          string   `json:"task_graph_id"`
	Title                string   `json:"title"`
	Objective            string   `json:"objective"`
	AssignedAgentProfile string   `json:"assigned_agent_profile"`
	Status               WorkUnitStatus `json:"status"`
	OwnedPaths           []string `json:"owned_paths"`
	ReadPaths            []string `json:"read_paths"`
	AcceptanceCriteria   []string `json:"acceptance_criteria"`
	ValidationPlan       []string `json:"validation_plan"`
	DependsOn            []string `json:"depends_on"`
}

// ============================================================================
// Agent Domain
// ============================================================================

type AgentRuntimeType string

const (
	AgentRuntimeTypeCodexCLI AgentRuntimeType = "codex_cli"
	AgentRuntimeTypeFake     AgentRuntimeType = "fake"
	AgentRuntimeTypeExternal AgentRuntimeType = "external"
	AgentRuntimeTypeGemini   AgentRuntimeType = "gemini"
)

type AgentStatus string

const (
	AgentStatusActive   AgentStatus = "active"
	AgentStatusInactive AgentStatus = "inactive"
)

type Agent struct {
	ID                     string           `json:"id"`
	Name                   string           `json:"name"`
	Profile                string           `json:"profile"`
	Capabilities           []string         `json:"capabilities"`
	AllowedTools           []string         `json:"allowed_tools"`
	DefaultPromptFragments []string         `json:"default_prompt_fragments"`
	RuntimeType            AgentRuntimeType `json:"runtime_type"`
	Status                 AgentStatus      `json:"status"`
}

// ============================================================================
// AgentSession Domain
// ============================================================================

type AgentSessionStatus string

const (
	AgentSessionStatusStarting     AgentSessionStatus = "starting"
	AgentSessionStatusRunning      AgentSessionStatus = "running"
	AgentSessionStatusWaitingApproval AgentSessionStatus = "waiting_approval"
	AgentSessionStatusPaused       AgentSessionStatus = "paused"
	AgentSessionStatusStopping     AgentSessionStatus = "stopping"
	AgentSessionStatusStopped      AgentSessionStatus = "stopped"
	AgentSessionStatusDisconnected AgentSessionStatus = "disconnected"
	AgentSessionStatusFailed       AgentSessionStatus = "failed"
)

type AgentSession struct {
	ID               string          `json:"id"`
	AgentID          string          `json:"agent_id"`
	RunID            string          `json:"run_id"`
	TaskID           string          `json:"task_id"`
	WorkUnitID       string          `json:"work_unit_id"`
	SandboxID        string          `json:"sandbox_id"`
	ConnectionID     string          `json:"connection_id"`
	Status           AgentSessionStatus `json:"status"`
	LastHeartbeatAt  *time.Time      `json:"last_heartbeat_at,omitempty"`
	LastCheckpointAt *time.Time      `json:"last_checkpoint_at,omitempty"`
	LastSeenEventID  string          `json:"last_seen_event_id,omitempty"`
	RecoverableState json.RawMessage `json:"recoverable_state,omitempty"`
}

// ============================================================================
// TaskGraph Domain
// ============================================================================

type TaskGraphStatus string

const (
	TaskGraphStatusActive     TaskGraphStatus = "active"
	TaskGraphStatusSuperseded TaskGraphStatus = "superseded"
)

type TaskGraph struct {
	ID              string        `json:"id"`
	TaskID          string        `json:"task_id"`
	Version         int           `json:"version"`
	Status          TaskGraphStatus `json:"status"`
	PlannerStrategy string        `json:"planner_strategy"`
	Rationale       string        `json:"rationale"`
	CreatedBy       string        `json:"created_by"`
	NodeCount       int           `json:"node_count"`
	EdgeCount       int           `json:"edge_count"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// ============================================================================
// Prompt Domain
// ============================================================================

type PromptSnapshot struct {
	ID                 string              `json:"id"`
	RunID              string              `json:"run_id"`
	WorkUnitID         string              `json:"work_unit_id"`
	AgentSessionID     string              `json:"agent_session_id"`
	SystemPrompt       string              `json:"system_prompt"`
	TaskPrompt         string              `json:"task_prompt"`
	CombinedPrompt     string              `json:"combined_prompt"`
	SystemPromptHash   string              `json:"system_prompt_hash"`
	TaskPromptHash     string              `json:"task_prompt_hash"`
	CombinedPromptHash string              `json:"combined_prompt_hash"`
	CompositionHash    string              `json:"composition_hash"`
	CategorySignature  string              `json:"category_signature"`
	FragmentRefs       []PromptFragmentRef `json:"fragment_refs"`
	AssemblyOrder      []string            `json:"assembly_order"`
	VariablesApplied   json.RawMessage     `json:"variables_applied"`
	CountUsed          int                 `json:"count_used"`
	FirstUsedAt        time.Time           `json:"first_used_at"`
	LastUsedAt         time.Time           `json:"last_used_at"`
	CreatedAt          time.Time           `json:"created_at"`
}

type PromptFragment struct {
	ID               string          `json:"id"`
	Version          string          `json:"version"`
	Category         string          `json:"category"`
	Kind             string          `json:"kind"`
	Title            string          `json:"title"`
	Priority         int             `json:"priority"`
	ExclusiveGroup   string          `json:"exclusive_group"`
	BodyHash         string          `json:"body_hash"`
	MetadataHash     string          `json:"metadata_hash"`
	Body             string          `json:"body"`
	AppliesWhen      json.RawMessage `json:"applies_when,omitempty"`
	Requires         []string        `json:"requires,omitempty"`
	ConflictsWith    []string        `json:"conflicts_with,omitempty"`
	Allows           []string        `json:"allows,omitempty"`
	Denies           []string        `json:"denies,omitempty"`
	ApprovalRequired []string        `json:"approval_required,omitempty"`
	AutonomyLevel    int             `json:"autonomy_level,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type PromptFragmentRef struct {
	ID           string `json:"id"`
	Version      string `json:"version"`
	Category     string `json:"category"`
	Kind         string `json:"kind"`
	Order        int    `json:"order"`
	BodyHash     string `json:"body_hash"`
	MetadataHash string `json:"metadata_hash"`
	Title        string `json:"title"`
}

type ToolsetSnapshot struct {
	ID             string        `json:"id"`
	RunID          string        `json:"run_id"`
	AgentSessionID string        `json:"agent_session_id"`
	Tools          []ToolsetTool `json:"tools"`
	CreatedReason  string        `json:"created_reason"`
	CreatedAt      time.Time     `json:"created_at"`
}

type ToolsetTool struct {
	Name   string `json:"name"`
	Scope  string `json:"scope"`
	Risk   string `json:"risk"`
	Reason string `json:"reason,omitempty"`
}

type ComposedPrompt struct {
	SystemPrompt       string                 `json:"system_prompt"`
	TaskPrompt         string                 `json:"task_prompt"`
	CombinedPrompt     string                 `json:"combined_prompt"`
	SystemPromptHash   string                 `json:"system_prompt_hash"`
	TaskPromptHash     string                 `json:"task_prompt_hash"`
	CombinedPromptHash string                 `json:"combined_prompt_hash"`
	CompositionHash    string                 `json:"composition_hash"`
	CategorySignature  string                 `json:"category_signature"`
	SystemProfile      SystemProfile          `json:"system_profile"`
	Fragments          []PromptFragment       `json:"fragments"`
	FragmentRefs       []PromptFragmentRef    `json:"fragment_refs"`
	AssemblyOrder      []string               `json:"assembly_order"`
	VariablesApplied   map[string]interface{} `json:"variables_applied"`
	Toolset            ToolsetSelection       `json:"toolset"`
}

type SystemProfile struct {
	Persona               string                 `json:"persona"`
	OperatingMode         string                 `json:"operating_mode"`
	TechnicalDomain       string                 `json:"technical_domain"`
	OutputContract        string                 `json:"output_contract"`
	ToolNames             []string               `json:"tool_names"`
	Allows                []string               `json:"allows"`
	Denies                []string               `json:"denies"`
	ApprovalRequired      []string               `json:"approval_required"`
	Categories            map[string]PromptFragmentRef `json:"categories"`
	CategorySignature     string                 `json:"category_signature"`
	TaskExecutionFocus    string                 `json:"task_execution_focus"`
	CanonicalAgentProfile string                 `json:"canonical_agent_profile"`
}

type ToolsetSelection struct {
	Profile       string       `json:"profile"`
	Tools         []ToolsetTool `json:"tools"`
	CreatedReason string       `json:"created_reason"`
}

// ============================================================================
// Review Domain
// ============================================================================

type ReviewStatus string

const (
	ReviewStatusPending          ReviewStatus = "pending"
	ReviewStatusInProgress       ReviewStatus = "in_progress"
	ReviewStatusApproved         ReviewStatus = "approved"
	ReviewStatusChangesRequested ReviewStatus = "changes_requested"
	ReviewStatusNeedsDiscussion  ReviewStatus = "needs_discussion"
)

type ReviewValidationGate string

const (
	ReviewValidationGateHard   ReviewValidationGate = "hard"
	ReviewValidationGateSoft   ReviewValidationGate = "soft"
	ReviewValidationGatePolicy ReviewValidationGate = "policy"
)

type Review struct {
	ID              string            `json:"id"`
	RunID           *string           `json:"run_id,omitempty"`
	WorkUnitID      *string           `json:"work_unit_id,omitempty"`
	TaskID          *string           `json:"task_id,omitempty"`
	AgentSessionID  *string           `json:"agent_session_id,omitempty"`
	ReviewerAgentID *string           `json:"reviewer_agent_id,omitempty"`
	GateType        ReviewValidationGate `json:"gate_type"`
	Status          ReviewStatus      `json:"status"`
	VerdictReason   string            `json:"verdict_reason,omitempty"`
	EvidenceRefs    []string          `json:"evidence_refs,omitempty"`
	CriteriaChecked []CriteriaChecked `json:"criteria_checked,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
}

type CriteriaChecked struct {
	Criterion string `json:"criterion"`
	Passed    bool   `json:"passed"`
	Reason    string `json:"reason,omitempty"`
}

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
	ID               string            `json:"id"`
	RunID            *string           `json:"run_id,omitempty"`
	TaskID           *string           `json:"task_id,omitempty"`
	AgentSessionID   *string           `json:"agent_session_id,omitempty"`
	TriggerType      TriggerType       `json:"trigger_type"`
	Status           TriggerStatus     `json:"status"`
	AnomalyType      *string           `json:"anomaly_type,omitempty"`
	ThresholdValue   json.RawMessage   `json:"threshold_value,omitempty"`
	CurrentValue     json.RawMessage   `json:"current_value,omitempty"`
	TriggeredAt      *time.Time        `json:"triggered_at,omitempty"`
	ResolvedAt       *time.Time        `json:"resolved_at,omitempty"`
	ResolutionAction *string           `json:"resolution_action,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
}
