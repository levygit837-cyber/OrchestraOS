// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package prompt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

func buildSystemProfile(systemFragments []Fragment, refs []FragmentRef, toolset ToolsetSelection, agentProfile string) SystemProfile {
	refsByCategory := make(map[string]FragmentRef, len(refs))
	for _, ref := range refs {
		refsByCategory[string(ref.Category)] = ref
	}

	profile := normalizeProfile(valueOrDefault(toolset.Profile, agentProfile))
	allows := make([]string, 0)
	denies := make([]string, 0)
	approvalRequired := make([]string, 0)
	for _, fragment := range systemFragments {
		allows = append(allows, fragment.Allows...)
		denies = append(denies, fragment.Denies...)
		approvalRequired = append(approvalRequired, fragment.ApprovalRequired...)
	}

	return SystemProfile{
		Persona:               titleForCategory(refsByCategory, CategoryPersona),
		OperatingMode:         titleForCategory(refsByCategory, CategoryOperatingMode),
		TechnicalDomain:       titleForCategory(refsByCategory, CategoryTechnicalDomain),
		OutputContract:        titleForCategory(refsByCategory, CategoryOutputContract),
		ToolNames:             ToolNames(toolset.Tools),
		Allows:                uniqueSorted(allows),
		Denies:                uniqueSorted(denies),
		ApprovalRequired:      uniqueSorted(approvalRequired),
		Categories:            refsByCategory,
		CategorySignature:     categorySignature(refs),
		TaskExecutionFocus:    taskExecutionFocus(profile),
		CanonicalAgentProfile: profile,
	}
}

func titleForCategory(refs map[string]FragmentRef, category FragmentCategory) string {
	if ref, ok := refs[string(category)]; ok {
		return ref.Title
	}
	return ""
}

func categorySignature(refs []FragmentRef) string {
	parts := make([]string, 0, len(refs))
	for _, ref := range refs {
		parts = append(parts, fmt.Sprintf("%s=%s@%s:%s:%s", ref.Category, ref.ID, ref.Version, ref.BodyHash, ref.MetadataHash))
	}
	sort.Strings(parts)
	return HashText(strings.Join(parts, "\n"))
}

func uniqueSorted(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func taskExecutionFocus(profile string) string {
	switch normalizeProfile(profile) {
	case "docs_writer":
		return "Produce documentation changes for this WorkUnit only, preserving repository truth and documenting behavior changes without expanding into implementation ownership."
	case "reviewer":
		return "Review the WorkUnit with a findings-first lens. Prioritize bugs, regressions, contract violations, missing tests, and scope conflicts; do not implement fixes unless the WorkUnit explicitly owns that edit."
	case "debugger":
		return "Diagnose the failure or anomaly assigned to this WorkUnit. Keep reproduction, evidence, and proposed fixes tied to the owned scope and dependencies."
	case "fake":
		return "Emit deterministic runtime progress for the WorkUnit, using the provided snapshot and toolset references without inventing extra execution scope."
	default:
		return "Implement the smallest sufficient change for this WorkUnit, edit only owned paths, and validate with the requested local checks."
	}
}

func renderTemplate(body string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("task_prompt").Option("missingkey=error").Parse(body)
	if err != nil {
		return "", fmt.Errorf("parse task prompt template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("render task prompt template: %w", err)
	}
	return buf.String(), nil
}

func variablesForContext(ctx TaskContext, profile SystemProfile) map[string]interface{} {
	return map[string]interface{}{
		"TaskID":                  ctx.TaskID,
		"TaskTitle":               valueOrDefault(ctx.TaskTitle, "(untitled task)"),
		"TaskDescription":         valueOrDefault(ctx.TaskDescription, "(no description provided)"),
		"WorkUnitID":              ctx.WorkUnitID,
		"TaskGraphID":             valueOrDefault(ctx.TaskGraphID, "(not linked to task graph)"),
		"WorkUnitTitle":           valueOrDefault(ctx.WorkUnitTitle, "(untitled work unit)"),
		"WorkUnitObjective":       valueOrDefault(ctx.WorkUnitObjective, ctx.WorkUnitTitle),
		"AgentProfile":            valueOrDefault(profile.CanonicalAgentProfile, valueOrDefault(ctx.AgentProfile, "code_worker")),
		"OwnedPaths":              formatList(ctx.OwnedPaths),
		"ReadPaths":               formatList(ctx.ReadPaths),
		"DependsOnInline":         formatInlineList(ctx.DependsOn),
		"AcceptanceCriteria":      formatList(ctx.AcceptanceCriteria),
		"ValidationPlan":          formatList(ctx.ValidationPlan),
		"Toolset":                 formatTools(ctx.Toolset.Tools),
		"SystemProfile":           profile,
		"SystemPersona":           valueOrDefault(profile.Persona, "(no persona selected)"),
		"SystemOperatingMode":     valueOrDefault(profile.OperatingMode, "(no operating mode selected)"),
		"SystemTechnicalDomain":   valueOrDefault(profile.TechnicalDomain, "(no technical domain selected)"),
		"SystemOutputContract":    valueOrDefault(profile.OutputContract, "(no output contract selected)"),
		"SystemAllowedOperations": formatInlineList(profile.Allows),
		"SystemDeniedOperations":  formatInlineList(profile.Denies),
		"SystemApprovalRequired":  formatInlineList(profile.ApprovalRequired),
		"SystemToolNames":         formatInlineList(profile.ToolNames),
		"SystemCategorySignature": profile.CategorySignature,
		"TaskExecutionFocus":      profile.TaskExecutionFocus,
	}
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func formatList(values []string) string {
	if len(values) == 0 {
		return "- (none specified)"
	}
	var out strings.Builder
	for _, value := range values {
		out.WriteString("- ")
		out.WriteString(value)
		out.WriteString("\n")
	}
	return strings.TrimRight(out.String(), "\n")
}

func formatInlineList(values []string) string {
	if len(values) == 0 {
		return "(none)"
	}
	return strings.Join(values, ", ")
}

func formatTools(tools []Tool) string {
	if len(tools) == 0 {
		return "- (no tools granted)"
	}
	var out strings.Builder
	for _, tool := range tools {
		out.WriteString("- ")
		out.WriteString(tool.Name)
		out.WriteString(" [")
		out.WriteString(string(tool.Risk))
		out.WriteString("] scope=")
		out.WriteString(tool.Scope)
		if tool.Reason != "" {
			out.WriteString(" reason=")
			out.WriteString(tool.Reason)
		}
		out.WriteString("\n")
	}
	return strings.TrimRight(out.String(), "\n")
}

func MarshalVariables(variables map[string]interface{}) (json.RawMessage, error) {
	if variables == nil {
		variables = map[string]interface{}{}
	}
	raw, err := json.Marshal(variables)
	if err != nil {
		return nil, fmt.Errorf("marshal prompt variables: %w", err)
	}
	return raw, nil
}
