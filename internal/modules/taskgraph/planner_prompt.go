package taskgraph

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

const plannerPromptTemplate = `You are an expert software project planner. Decompose the following task into a directed acyclic graph (DAG) of work units.

Each work unit must be self-contained, have a clear objective, and include specific acceptance criteria.

## Task Information

**Title:** {{.Task.Title}}
**Description:** {{.Task.Description}}
**Priority:** {{.Task.Priority}}
**Risk Level:** {{.Task.RiskLevel}}

{{- if .Task.AcceptanceCriteria }}
**Original Acceptance Criteria:**
{{- range $i, $c := .Task.AcceptanceCriteria }}
{{add $i 1}}. {{$c}}
{{- end }}
{{- end }}

## Planning Rules

1. Create between 1 and {{.MaxWorkUnits}} work units.
2. Each work unit must have:
   - A descriptive title
   - A clear objective explaining what must be achieved
   - An assigned_agent_profile from the allowed list
   - At least one acceptance criterion
   - At least one validation step
   - Optional dependencies on other work units (use 0-based indices)
   - Optional owned_paths and read_paths if relevant
3. The dependency graph MUST be acyclic (no circular dependencies).
4. Choose agent profiles wisely:
   - code_worker: implementation, coding, refactoring
   - docs_writer: documentation, README, ADRs
   - reviewer: code review, validation, quality checks
   - debugger: debugging, troubleshooting, fixing bugs
   - default: general purpose work
5. The decomposition should reflect semantic understanding of the task, not just word-count splitting.
6. Provide a rationale explaining your decomposition strategy.

## Output Format

Respond ONLY with a valid JSON object matching this exact structure (no markdown, no explanation outside JSON):

{
  "rationale": "string explaining the decomposition strategy",
  "work_units": [
    {
      "title": "string",
      "objective": "string",
      "assigned_agent_profile": "code_worker|docs_writer|reviewer|debugger|default",
      "acceptance_criteria": ["string"],
      "validation_plan": ["string"],
      "depends_on": [0],
      "owned_paths": ["string"],
      "read_paths": ["string"]
    }
  ]
}`

var plannerPromptFuncs = template.FuncMap{
	"add": func(a, b int) int { return a + b },
}

var parsedPlannerPromptTemplate = template.Must(
	template.New("planner").Funcs(plannerPromptFuncs).Parse(plannerPromptTemplate),
)

// PlannerPromptInput holds the data for rendering the planner prompt.
type PlannerPromptInput struct {
	Task         *domain.Task
	MaxWorkUnits int
}

// BuildPlannerPrompt renders the planner prompt for a given task.
func BuildPlannerPrompt(input PlannerPromptInput) (string, error) {
	if input.Task == nil {
		return "", apperrors.New(apperrors.CodeInvalidInput, "planner_prompt.build", "task is required")
	}
	if input.MaxWorkUnits <= 0 {
		input.MaxWorkUnits = maxGraphWorkUnits
	}

	var buf bytes.Buffer
	if err := parsedPlannerPromptTemplate.Execute(&buf, input); err != nil {
		return "", apperrors.Wrap(apperrors.CodeValidation, "planner_prompt.render", err)
	}
	return strings.TrimSpace(buf.String()), nil
}

// PlannerPrompt returns a simple string prompt for the given task (convenience function).
func PlannerPrompt(task *domain.Task) (string, error) {
	return BuildPlannerPrompt(PlannerPromptInput{Task: task, MaxWorkUnits: maxGraphWorkUnits})
}

// ValidatePlannerProfile checks if the given agent profile is valid for planner output.
func ValidatePlannerProfile(profile string) error {
	validProfiles := map[string]bool{
		"code_worker": true,
		"docs_writer": true,
		"reviewer":    true,
		"debugger":    true,
		"default":     true,
	}
	if !validProfiles[profile] {
		return apperrors.New(
			apperrors.CodeValidation,
			"planner.validate_profile",
			fmt.Sprintf("invalid agent profile %q", profile),
		)
	}
	return nil
}
