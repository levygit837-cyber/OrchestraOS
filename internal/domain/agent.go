package domain

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
