package prompt

import (
	"encoding/json"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// ============================================================================
// Aliases to shared domain types per ADR-0019
// ============================================================================

type PromptSnapshot = domain.PromptSnapshot
type PromptFragment = domain.PromptFragment
type PromptFragmentRef = domain.PromptFragmentRef
type ToolsetSnapshot = domain.ToolsetSnapshot
type ToolsetTool = domain.ToolsetTool

// ============================================================================
// Constants
// ============================================================================

const (
	MaxAutonomyLevel = 2
)

type FragmentKind string

const (
	FragmentKindGlobalPolicy    FragmentKind = "global_policy"
	FragmentKindRepoContext     FragmentKind = "repo_context"
	FragmentKindAutonomyPolicy  FragmentKind = "autonomy_policy"
	FragmentKindPersona         FragmentKind = "persona"
	FragmentKindOperatingMode   FragmentKind = "operating_mode"
	FragmentKindTechnicalDomain FragmentKind = "technical_domain"
	FragmentKindToolPolicy      FragmentKind = "tool_policy"
	FragmentKindCommunication   FragmentKind = "communication"
	FragmentKindOutputContract  FragmentKind = "output_contract"
	FragmentKindValidation      FragmentKind = "validation"
	FragmentKindLedger          FragmentKind = "ledger"
	FragmentKindTaskTemplate    FragmentKind = "task_template"
)

type FragmentCategory string

const (
	CategoryPolicyGlobal      FragmentCategory = "policy.global"
	CategoryPolicyAutonomy    FragmentCategory = "policy.autonomy"
	CategoryRepositoryContext FragmentCategory = "repository.context"
	CategoryPersona           FragmentCategory = "persona"
	CategoryOperatingMode     FragmentCategory = "operating_mode"
	CategoryTechnicalDomain   FragmentCategory = "technical_domain"
	CategoryToolPolicy        FragmentCategory = "tool_policy"
	CategoryCommunication     FragmentCategory = "communication"
	CategoryValidation        FragmentCategory = "validation"
	CategoryLedger            FragmentCategory = "ledger"
	CategoryOutputContract    FragmentCategory = "output_contract"
	CategoryTaskTemplate      FragmentCategory = "task_template"
)

var RequiredCategories = []FragmentCategory{
	CategoryPolicyGlobal,
	CategoryPolicyAutonomy,
	CategoryRepositoryContext,
	CategoryPersona,
	CategoryOperatingMode,
	CategoryTechnicalDomain,
	CategoryToolPolicy,
	CategoryCommunication,
	CategoryValidation,
	CategoryLedger,
	CategoryOutputContract,
	CategoryTaskTemplate,
}

// ============================================================================
// Composition Types (local to prompt module)
// ============================================================================

type AppliesWhen struct {
	AgentProfiles []string `json:"agent_profiles,omitempty"`
}

// Fragment is the runtime representation of a prompt fragment used by the composer.
// It mirrors PromptFragment but uses typed fields.
type Fragment struct {
	ID               string           `json:"id"`
	Version          string           `json:"version"`
	Category         FragmentCategory `json:"category"`
	Kind             FragmentKind     `json:"kind"`
	Title            string           `json:"title"`
	Priority         int              `json:"priority"`
	ExclusiveGroup   string           `json:"exclusive_group,omitempty"`
	BodyPath         string           `json:"body_path"`
	BodyHash         string           `json:"body_hash"`
	MetadataHash     string           `json:"metadata_hash"`
	Body             string           `json:"body"`
	AppliesWhen      AppliesWhen      `json:"applies_when,omitempty"`
	Requires         []string         `json:"requires,omitempty"`
	ConflictsWith    []string         `json:"conflicts_with,omitempty"`
	Allows           []string         `json:"allows,omitempty"`
	Denies           []string         `json:"denies,omitempty"`
	ApprovalRequired []string         `json:"approval_required,omitempty"`
	AutonomyLevel    int              `json:"autonomy_level,omitempty"`
	CreatedAt        time.Time        `json:"created_at,omitempty"`
	Metadata         json.RawMessage  `json:"metadata,omitempty"`
}

// FragmentRef is the runtime reference used by the composer.
type FragmentRef struct {
	ID           string           `json:"id"`
	Version      string           `json:"version"`
	Category     FragmentCategory `json:"category"`
	Kind         FragmentKind     `json:"kind"`
	Order        int              `json:"order"`
	BodyHash     string           `json:"body_hash"`
	MetadataHash string           `json:"metadata_hash"`
	Title        string           `json:"title"`
}

type ToolRisk string

const (
	ToolRiskSafe             ToolRisk = "safe"
	ToolRiskGuarded          ToolRisk = "guarded"
	ToolRiskApprovalRequired ToolRisk = "approval_required"
	ToolRiskDestructive      ToolRisk = "destructive"
	ToolRiskForbidden        ToolRisk = "forbidden"
)

type Tool struct {
	Name   string   `json:"name"`
	Scope  string   `json:"scope"`
	Risk   ToolRisk `json:"risk"`
	Reason string   `json:"reason,omitempty"`
}

type TaskContext struct {
	TaskID             string
	TaskTitle          string
	TaskDescription    string
	RunID              string
	WorkUnitID         string
	TaskGraphID        string
	WorkUnitTitle      string
	WorkUnitObjective  string
	AgentProfile       string
	OwnedPaths         []string
	ReadPaths          []string
	DependsOn          []string
	AcceptanceCriteria []string
	ValidationPlan     []string
	Toolset            ToolsetSelection
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
	Categories            map[string]FragmentRef `json:"categories"`
	CategorySignature     string                 `json:"category_signature"`
	TaskExecutionFocus    string                 `json:"task_execution_focus"`
	CanonicalAgentProfile string                 `json:"canonical_agent_profile"`
}

type ToolsetSelection struct {
	Profile       string `json:"profile"`
	Tools         []Tool `json:"tools"`
	CreatedReason string `json:"created_reason"`
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
	Fragments          []Fragment             `json:"fragments"`
	FragmentRefs       []FragmentRef          `json:"fragment_refs"`
	AssemblyOrder      []string               `json:"assembly_order"`
	VariablesApplied   map[string]interface{} `json:"variables_applied"`
	Toolset            ToolsetSelection       `json:"toolset"`
}
