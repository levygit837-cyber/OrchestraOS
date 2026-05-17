package agent

// Agent represents a configured agent runtime in the system.
type Agent struct {
	ID                     string      `json:"id"`
	Name                   string      `json:"name"`
	Profile                string      `json:"profile"`
	Capabilities           []string    `json:"capabilities"`
	AllowedTools           []string    `json:"allowed_tools"`
	DefaultPromptFragments []string    `json:"default_prompt_fragments"`
	RuntimeType            RuntimeType `json:"runtime_type"`
}

// RuntimeType defines the available agent runtime implementations.
type RuntimeType string

const (
	RuntimeTypeCodexCLI RuntimeType = "codex_cli"
	RuntimeTypeFake     RuntimeType = "fake"
	RuntimeTypeExternal RuntimeType = "external"
	RuntimeTypeGemini   RuntimeType = "gemini"
)

// AgentStatus represents the status of an agent in the database.
type AgentStatus string

const (
	AgentStatusActive   AgentStatus = "active"
	AgentStatusInactive AgentStatus = "inactive"
)
