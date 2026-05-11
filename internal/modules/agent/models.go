package agent

type RuntimeType string

const (
	RuntimeTypeCodexCLI RuntimeType = "codex_cli"
	RuntimeTypeFake     RuntimeType = "fake"
	RuntimeTypeExternal RuntimeType = "external"
	RuntimeTypeGemini   RuntimeType = "gemini"
)

type Agent struct {
	ID                     string           `json:"id"`
	Name                   string           `json:"name"`
	Profile                string           `json:"profile"`
	Capabilities           []string         `json:"capabilities"`
	AllowedTools           []string         `json:"allowed_tools"`
	DefaultPromptFragments []string         `json:"default_prompt_fragments"`
	RuntimeType            RuntimeType      `json:"runtime_type"`
}
