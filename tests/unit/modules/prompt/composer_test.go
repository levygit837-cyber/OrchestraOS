package prompt

import (
	"fmt"
	"strings"
	"testing"
)

func TestComposerBuildsPromptWithMandatoryFragments(t *testing.T) {
	toolset, err := SelectToolset("code_worker")
	if err != nil {
		t.Fatalf("select toolset: %v", err)
	}
	composed, err := Compose(TaskContext{
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
		t.Fatalf("expected work unit decomposition scope in task prompt")
	}
	if composed.SystemPromptHash == "" || composed.TaskPromptHash == "" || composed.CombinedPromptHash == "" {
		t.Fatalf("expected prompt hashes")
	}
	if composed.CompositionHash == "" || composed.CategorySignature == "" {
		t.Fatalf("expected composition and category hashes")
	}
	if len(composed.FragmentRefs) == 0 || len(composed.AssemblyOrder) != len(composed.FragmentRefs) {
		t.Fatalf("expected fragment refs and assembly order")
	}
}

func TestComposerHashIsStable(t *testing.T) {
	ctx := TaskContext{
		TaskID:            "task-1",
		TaskTitle:         "Stable",
		RunID:             "run-1",
		WorkUnitID:        "wu-1",
		WorkUnitObjective: "Stable prompt",
		AgentProfile:      "fake",
	}
	first, err := Compose(ctx)
	if err != nil {
		t.Fatalf("first compose: %v", err)
	}
	second, err := Compose(ctx)
	if err != nil {
		t.Fatalf("second compose: %v", err)
	}
	if first.CombinedPromptHash != second.CombinedPromptHash {
		t.Fatalf("expected stable combined hash, got %s then %s", first.CombinedPromptHash, second.CombinedPromptHash)
	}
	if first.CompositionHash != second.CompositionHash {
		t.Fatalf("expected stable composition hash, got %s then %s", first.CompositionHash, second.CompositionHash)
	}
}

func TestComposerTaskPromptChangesWithSystemProfile(t *testing.T) {
	codeToolset, err := SelectToolset("code_worker")
	if err != nil {
		t.Fatalf("select code worker toolset: %v", err)
	}
	reviewerToolset, err := SelectToolset("reviewer")
	if err != nil {
		t.Fatalf("select reviewer toolset: %v", err)
	}
	base := TaskContext{
		TaskID:            "task-1",
		TaskTitle:         "Profile-aware prompt",
		TaskGraphID:       "tg-1",
		WorkUnitID:        "wu-1",
		WorkUnitObjective: "Handle profile-aware WorkUnit.",
		OwnedPaths:        []string{"internal/prompting"},
	}
	codeCtx := base
	codeCtx.AgentProfile = "code_worker"
	codeCtx.Toolset = codeToolset
	reviewerCtx := base
	reviewerCtx.AgentProfile = "reviewer"
	reviewerCtx.Toolset = reviewerToolset

	codePrompt, err := Compose(codeCtx)
	if err != nil {
		t.Fatalf("compose code worker: %v", err)
	}
	reviewerPrompt, err := Compose(reviewerCtx)
	if err != nil {
		t.Fatalf("compose reviewer: %v", err)
	}
	if codePrompt.TaskPrompt == reviewerPrompt.TaskPrompt {
		t.Fatal("expected task prompt to change when persona/mode changes")
	}
	if !strings.Contains(codePrompt.TaskPrompt, "Implement the smallest sufficient change") {
		t.Fatalf("expected code worker task focus, got %s", codePrompt.TaskPrompt)
	}
	if !strings.Contains(reviewerPrompt.TaskPrompt, "findings-first") {
		t.Fatalf("expected reviewer task focus, got %s", reviewerPrompt.TaskPrompt)
	}
}

func TestValidateSelectedFragmentsRejectsExclusiveGroupCollision(t *testing.T) {
	fragments := minimalValidFragments()
	fragments[0].ExclusiveGroup = "same-group"
	fragments[1].ExclusiveGroup = "same-group"
	if err := validateSelectedFragments(fragments); err == nil {
		t.Fatal("expected duplicate exclusive group to be rejected")
	}
}

func TestValidateSelectedFragmentsRejectsCategoryCollision(t *testing.T) {
	fragments := minimalValidFragments()
	fragments = append(fragments, Fragment{
		ID:             "fragment.duplicate_persona",
		Version:        "1.0.0",
		Category:       CategoryPersona,
		Kind:           FragmentKindPersona,
		ExclusiveGroup: "persona.secondary",
	})
	if err := validateSelectedFragments(fragments); err == nil {
		t.Fatal("expected duplicate category to be rejected")
	}
}

func TestValidateSelectedFragmentsRequiresMandatoryCategories(t *testing.T) {
	fragments := minimalValidFragments()
	fragments = fragments[:len(fragments)-1]
	if err := validateSelectedFragments(fragments); err == nil {
		t.Fatal("expected missing mandatory category to be rejected")
	}
}

func TestValidateSelectedFragmentsRejectsExplicitConflict(t *testing.T) {
	fragments := minimalValidFragments()
	fragments[0].ConflictsWith = []string{fragments[1].ID}
	if err := validateSelectedFragments(fragments); err == nil {
		t.Fatal("expected explicit conflict to be rejected")
	}
}

func TestValidateSelectedFragmentsRejectsAllowDenyConflict(t *testing.T) {
	fragments := minimalValidFragments()
	fragments[0].Allows = []string{"network_access"}
	fragments[1].Denies = []string{"network_access"}
	if err := validateSelectedFragments(fragments); err == nil {
		t.Fatal("expected allow/deny conflict to be rejected")
	}
}

func TestValidateSelectedFragmentsRejectsMissingRequiredFragment(t *testing.T) {
	fragments := minimalValidFragments()
	fragments[0].Requires = []string{"fragment.parent"}
	if err := validateSelectedFragments(fragments); err == nil {
		t.Fatal("expected missing required fragment to be rejected")
	}
}

func TestValidateSelectedFragmentsRejectsAutonomyAboveMVP(t *testing.T) {
	fragments := minimalValidFragments()
	fragments[0].AutonomyLevel = MaxAutonomyLevel + 1
	if err := validateSelectedFragments(fragments); err == nil {
		t.Fatal("expected autonomy above MVP to be rejected")
	}
}

func minimalValidFragments() []Fragment {
	fragments := make([]Fragment, 0, len(RequiredCategories))
	for i, category := range RequiredCategories {
		kind := FragmentKindGlobalPolicy
		if category == CategoryTaskTemplate {
			kind = FragmentKindTaskTemplate
		}
		fragments = append(fragments, Fragment{
			ID:             fmt.Sprintf("fragment.%02d", i),
			Version:        "1.0.0",
			Category:       category,
			Kind:           kind,
			ExclusiveGroup: string(category),
		})
	}
	return fragments
}
