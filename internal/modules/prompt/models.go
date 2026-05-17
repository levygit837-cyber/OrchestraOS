package prompt

import (
	"encoding/json"
	"time"
)

// PromptSnapshot stores a composed prompt for a specific run.
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

// PromptFragment stores a reusable prompt fragment.
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

// PromptFragmentRef is a lightweight reference to a fragment.
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

// ToolsetSnapshot stores a composed toolset for a specific run.
type ToolsetSnapshot struct {
	ID             string        `json:"id"`
	RunID          string        `json:"run_id"`
	AgentSessionID string        `json:"agent_session_id"`
	Tools          []ToolsetTool `json:"tools"`
	CreatedReason  string        `json:"created_reason"`
	CreatedAt      time.Time     `json:"created_at"`
}

// ToolsetTool represents a single tool in a toolset.
type ToolsetTool struct {
	Name   string `json:"name"`
	Scope  string `json:"scope"`
	Risk   string `json:"risk"`
	Reason string `json:"reason,omitempty"`
}
