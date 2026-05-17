// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package prompt

import (
	"fmt"
	"sort"
	"strings"
)

type Composer struct {
	catalog *Catalog
}

func NewComposer(catalog *Catalog) *Composer {
	return &Composer{catalog: catalog}
}

func Compose(ctx TaskContext) (*ComposedPrompt, error) {
	catalog, err := LoadCatalog()
	if err != nil {
		return nil, err
	}
	return NewComposer(catalog).Compose(ctx)
}

func (c *Composer) Compose(ctx TaskContext) (*ComposedPrompt, error) {
	if c == nil || c.catalog == nil {
		return nil, fmt.Errorf("prompt composer requires a catalog")
	}
	selected, err := c.selectFragments(ctx.AgentProfile)
	if err != nil {
		return nil, err
	}
	if err := ValidateSelectedFragments(selected); err != nil {
		return nil, err
	}

	refs := make([]FragmentRef, 0, len(selected))
	order := make([]string, 0, len(selected))
	for i, fragment := range selected {
		refs = append(refs, FragmentRef{
			ID:           fragment.ID,
			Version:      fragment.Version,
			Category:     fragment.Category,
			Kind:         fragment.Kind,
			Order:        i + 1,
			BodyHash:     fragment.BodyHash,
			MetadataHash: fragment.MetadataHash,
			Title:        fragment.Title,
		})
		order = append(order, fragment.ID+"@"+fragment.Version)
	}

	var taskTemplateFragment *Fragment
	var systemFragments []Fragment
	for i := range selected {
		fragment := selected[i]
		if fragment.Kind == FragmentKindTaskTemplate {
			copyFragment := fragment
			taskTemplateFragment = &copyFragment
			continue
		}
		systemFragments = append(systemFragments, fragment)
	}
	if taskTemplateFragment == nil {
		return nil, fmt.Errorf("task template fragment is required")
	}
	systemProfile := buildSystemProfile(systemFragments, refs, ctx.Toolset, ctx.AgentProfile)

	var system strings.Builder
	system.WriteString("# System Prompt\n\n")
	for _, fragment := range systemFragments {
		system.WriteString("## ")
		system.WriteString(fragment.Title)
		system.WriteString("\n\n")
		system.WriteString(strings.TrimSpace(fragment.Body))
		system.WriteString("\n\n")
	}

	variables := variablesForContext(ctx, systemProfile)
	taskPrompt, err := renderTemplate(taskTemplateFragment.Body, variables)
	if err != nil {
		return nil, err
	}
	combined := strings.TrimSpace(system.String()) + "\n\n--- TASK PROMPT ---\n\n" + strings.TrimSpace(taskPrompt)

	systemPrompt := strings.TrimSpace(system.String())
	taskPrompt = strings.TrimSpace(taskPrompt)
	systemPromptHash := HashText(systemPrompt)
	taskPromptHash := HashText(taskPrompt)
	combinedPromptHash := HashText(combined)
	compositionHash := HashText(strings.Join([]string{
		systemProfile.CategorySignature,
		systemPromptHash,
		taskPromptHash,
		combinedPromptHash,
	}, "\n"))

	return &ComposedPrompt{
		SystemPrompt:       systemPrompt,
		TaskPrompt:         taskPrompt,
		CombinedPrompt:     combined,
		SystemPromptHash:   systemPromptHash,
		TaskPromptHash:     taskPromptHash,
		CombinedPromptHash: combinedPromptHash,
		CompositionHash:    compositionHash,
		CategorySignature:  systemProfile.CategorySignature,
		SystemProfile:      systemProfile,
		Fragments:          selected,
		FragmentRefs:       refs,
		AssemblyOrder:      order,
		VariablesApplied:   variables,
	}, nil
}

func (c *Composer) selectFragments(profile string) ([]Fragment, error) {
	normalized := normalizeProfile(profile)
	fragments := c.catalog.Fragments()
	selected := make([]Fragment, 0, len(fragments))
	for _, fragment := range fragments {
		if appliesToProfile(fragment, normalized) {
			selected = append(selected, fragment)
		}
	}
	sort.SliceStable(selected, func(i, j int) bool {
		if selected[i].Priority == selected[j].Priority {
			return selected[i].ID < selected[j].ID
		}
		return selected[i].Priority < selected[j].Priority
	})
	return selected, nil
}

func appliesToProfile(fragment Fragment, profile string) bool {
	profiles := fragment.AppliesWhen.AgentProfiles
	if len(profiles) == 0 {
		return true
	}
	for _, candidate := range profiles {
		if normalizeProfile(candidate) == profile {
			return true
		}
	}
	return false
}

func ValidateSelectedFragments(fragments []Fragment) error {
	selectedByID := map[string]Fragment{}
	exclusiveGroups := map[string]string{}
	categories := map[FragmentCategory]string{}
	allowed := map[string]string{}
	denied := map[string]string{}

	for _, fragment := range fragments {
		if fragment.Category == "" {
			return fmt.Errorf("fragment %s is missing category", fragment.ID)
		}
		if existing, ok := categories[fragment.Category]; ok {
			return fmt.Errorf("fragments %s and %s share category %s", existing, fragment.ID, fragment.Category)
		}
		categories[fragment.Category] = fragment.ID
		if fragment.AutonomyLevel > MaxAutonomyLevel {
			return fmt.Errorf("fragment %s requests autonomy level %d above maximum %d", fragment.ID, fragment.AutonomyLevel, MaxAutonomyLevel)
		}
		if _, ok := selectedByID[fragment.ID]; ok {
			return fmt.Errorf("fragment %s selected more than once", fragment.ID)
		}
		selectedByID[fragment.ID] = fragment
		if fragment.ExclusiveGroup != "" {
			if existing, ok := exclusiveGroups[fragment.ExclusiveGroup]; ok {
				return fmt.Errorf("fragments %s and %s share exclusive group %s", existing, fragment.ID, fragment.ExclusiveGroup)
			}
			exclusiveGroups[fragment.ExclusiveGroup] = fragment.ID
		}
		for _, operation := range fragment.Allows {
			allowed[operation] = fragment.ID
		}
		for _, operation := range fragment.Denies {
			denied[operation] = fragment.ID
		}
	}
	for _, requiredCategory := range RequiredCategories {
		if _, ok := categories[requiredCategory]; !ok {
			return fmt.Errorf("required prompt category %s is missing", requiredCategory)
		}
	}

	for _, fragment := range fragments {
		for _, requiredID := range fragment.Requires {
			if _, ok := selectedByID[requiredID]; !ok {
				return fmt.Errorf("fragment %s requires missing fragment %s", fragment.ID, requiredID)
			}
		}
		for _, conflictID := range fragment.ConflictsWith {
			if _, ok := selectedByID[conflictID]; ok {
				return fmt.Errorf("fragment %s conflicts with selected fragment %s", fragment.ID, conflictID)
			}
		}
	}

	for operation, allowFragmentID := range allowed {
		if denyFragmentID, ok := denied[operation]; ok {
			return fmt.Errorf("operation %s is both allowed by %s and denied by %s", operation, allowFragmentID, denyFragmentID)
		}
	}
	return nil
}
