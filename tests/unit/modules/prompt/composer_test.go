package prompt_test

import (
	"strings"
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/modules/prompt"
)

func TestComposerBuildsPromptWithMandatoryFragments(t *testing.T) {
	toolset, err := prompt.SelectToolset("code_worker")
	if err != nil {
		t.Fatalf("select toolset: %v", err)
	}
	composed, err := prompt.Compose(prompt.TaskContext{
		TaskID:             "task-1",
		TaskTitle:          "Implement prompts",
		TaskDescription:    "Create prompt composition.",
		RunID:              "run-1",
		WorkUnitID:         "wu-1",
		TaskGraphID:        "tg-1",
		WorkUnitTitle:      "Prompt composer",
		WorkUnitObjective:  "Build deterministic prompt composition.",
		AgentProfile:       "code_worker",
		OwnedPaths:         []string{"internal/prompting"},
		ReadPaths:          []string{"docs/adr/0007-prompt-composition-system.md"},
		DependsOn:          []string{"wu-0"},
		AcceptanceCriteria: []string{"snapshot has hashes"},
		ValidationPlan:     []string{"go test ./..."},
		Toolset:            toolset,
	})
	if err != nil {
		t.Fatalf("compose prompt: %v", err)
	}

	if !strings.Contains(composed.SystemPrompt, "Global Prompt Policy") {
		t.Fatalf("expected global policy in system prompt")
	}
	if !strings.Contains(composed.SystemPrompt, "Code Worker") {
		t.Fatalf("expected code worker persona in system prompt")
	}
	if strings.Contains(composed.SystemPrompt, "Reviewer\n") {
		t.Fatalf("did not expect reviewer persona in code worker prompt")
	}
	if !strings.Contains(composed.TaskPrompt, "Build deterministic prompt composition.") {
		t.Fatalf("expected work unit objective in task prompt")
	}
	if !strings.Contains(composed.TaskPrompt, "Initial Todo Ledger") {
		t.Fatalf("expected initial todo ledger in task prompt")
	}
	if !strings.Contains(composed.TaskPrompt, "TaskPromptDecompose") || !strings.Contains(composed.TaskPrompt, "tg-1") || !strings.Contains(composed.TaskPrompt, "wu-0") {
		t.Fatalf("expected decomposition context variables in task prompt")
	}
	if composed.CompositionHash == "" {
		t.Fatalf("expected non-empty prompt hash")
	}
}

func TestComposerHashIsStable(t *testing.T) {
	toolset, _ := prompt.SelectToolset("code_worker")
	ctx := prompt.TaskContext{
		TaskID: "task-1", TaskTitle: "Hash test", TaskDescription: "Hash stability.",
		RunID: "run-1", WorkUnitID: "wu-1", TaskGraphID: "tg-1",
		WorkUnitTitle: "Hash work unit", WorkUnitObjective: "Build hash.",
		AgentProfile: "code_worker", OwnedPaths: []string{}, ReadPaths: []string{},
		DependsOn: []string{}, AcceptanceCriteria: []string{"hash stable"},
		ValidationPlan: []string{"go test ./..."}, Toolset: toolset,
	}

	first, err := prompt.Compose(ctx)
	if err != nil {
		t.Fatalf("first compose: %v", err)
	}
	second, err := prompt.Compose(ctx)
	if err != nil {
		t.Fatalf("second compose: %v", err)
	}
	if first.CompositionHash != second.CompositionHash {
		t.Fatalf("expected stable hash, got %q vs %q", first.CompositionHash, second.CompositionHash)
	}
}

func TestComposerTaskPromptChangesWithSystemProfile(t *testing.T) {
	toolset, _ := prompt.SelectToolset("code_worker")
	ctx := prompt.TaskContext{
		TaskID: "task-1", TaskTitle: "Profile test", TaskDescription: "Profile changes prompt.",
		RunID: "run-1", WorkUnitID: "wu-1", TaskGraphID: "tg-1",
		WorkUnitTitle: "Profile work unit", WorkUnitObjective: "Build profile.",
		AgentProfile: "code_worker", OwnedPaths: []string{}, ReadPaths: []string{},
		DependsOn: []string{}, AcceptanceCriteria: []string{"profile changes prompt"},
		ValidationPlan: []string{"go test ./..."}, Toolset: toolset,
	}

	codeWorker, _ := prompt.Compose(ctx)
	ctx.AgentProfile = "reviewer"
	reviewer, _ := prompt.Compose(ctx)

	if strings.Contains(codeWorker.SystemPrompt, "Reviewer") {
		t.Fatalf("did not expect reviewer in code worker prompt")
	}
	if !strings.Contains(reviewer.SystemPrompt, "Reviewer") {
		t.Fatalf("expected reviewer persona in reviewer prompt")
	}
}

func TestValidateSelectedFragmentsRejectsExclusiveGroupCollision(t *testing.T) {
	fragments := []prompt.Fragment{
		{ID: "a", Category: "goal", ExclusiveGroup: "group1"},
		{ID: "b", Category: "goal", ExclusiveGroup: "group1"},
	}
	if err := prompt.ValidateSelectedFragments(fragments); err == nil {
		t.Fatal("expected exclusive group collision to be rejected")
	}
}

func TestValidateSelectedFragmentsRejectsCategoryCollision(t *testing.T) {
	fragments := []prompt.Fragment{
		{ID: "a", Category: "goal"},
		{ID: "b", Category: "goal"},
	}
	if err := prompt.ValidateSelectedFragments(fragments); err == nil {
		t.Fatal("expected category collision to be rejected")
	}
}

func TestValidateSelectedFragmentsRequiresMandatoryCategories(t *testing.T) {
	if err := prompt.ValidateSelectedFragments([]prompt.Fragment{}); err == nil {
		t.Fatal("expected empty fragments to be rejected")
	}
}

func TestValidateSelectedFragmentsRejectsExplicitConflict(t *testing.T) {
	fragments := []prompt.Fragment{
		{ID: "a", Category: "goal", ConflictsWith: []string{"b"}},
		{ID: "b", Category: "context"},
	}
	if err := prompt.ValidateSelectedFragments(fragments); err == nil {
		t.Fatal("expected explicit conflict to be rejected")
	}
}

func TestValidateSelectedFragmentsRejectsAllowDenyConflict(t *testing.T) {
	fragments := []prompt.Fragment{
		{ID: "a", Category: "goal", Allows: []string{"docs/"}},
		{ID: "b", Category: "context", Denies: []string{"docs/"}},
	}
	if err := prompt.ValidateSelectedFragments(fragments); err == nil {
		t.Fatal("expected allow/deny conflict to be rejected")
	}
}

func TestValidateSelectedFragmentsRejectsMissingRequiredFragment(t *testing.T) {
	fragments := []prompt.Fragment{
		{ID: "a", Category: "goal", Requires: []string{"missing"}},
	}
	if err := prompt.ValidateSelectedFragments(fragments); err == nil {
		t.Fatal("expected missing required fragment to be rejected")
	}
}

func TestValidateSelectedFragmentsRejectsAutonomyAboveMVP(t *testing.T) {
	fragments := []prompt.Fragment{
		{ID: "a", Category: "goal", AutonomyLevel: 3},
	}
	if err := prompt.ValidateSelectedFragments(fragments); err == nil {
		t.Fatal("expected autonomy above MVP to be rejected")
	}
}
