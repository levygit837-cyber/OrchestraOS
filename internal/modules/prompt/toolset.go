package prompt

import "fmt"

func SelectToolset(profile string) (ToolsetSelection, error) {
	normalized := normalizeProfile(profile)
	switch normalized {
	case "default", "codex", "general_engineering", "gemini":
		normalized = "code_worker"
	}

	base := []Tool{
		{Name: "event.emit", Scope: "current_run", Risk: ToolRiskSafe, Reason: "Emit structured run and agent events."},
		{Name: "ledger.update", Scope: "current_work_unit", Risk: ToolRiskSafe, Reason: "Update the work unit operational ledger at checkpoints."},
		{Name: "toolset.request_change", Scope: "current_agent_session", Risk: ToolRiskSafe, Reason: "Request a missing tool without executing it."},
	}

	var tools []Tool
	switch normalized {
	case "fake":
		tools = append([]Tool{
			{Name: "runtime.fake.emit", Scope: "current_run", Risk: ToolRiskSafe, Reason: "Emit deterministic fake runtime events."},
		}, base...)
	case "docs_writer":
		tools = append([]Tool{
			{Name: "filesystem.read", Scope: "approved_read_paths", Risk: ToolRiskSafe, Reason: "Read authorized documentation and context."},
			{Name: "filesystem.write_docs", Scope: "owned_documentation_paths", Risk: ToolRiskGuarded, Reason: "Edit documentation paths owned by the work unit."},
			{Name: "git.diff", Scope: "owned_worktree", Risk: ToolRiskSafe, Reason: "Inspect local documentation diff."},
		}, base...)
	case "code_worker":
		tools = append([]Tool{
			{Name: "filesystem.read", Scope: "approved_read_paths", Risk: ToolRiskSafe, Reason: "Read authorized code and docs."},
			{Name: "filesystem.write_owned", Scope: "owned_paths", Risk: ToolRiskGuarded, Reason: "Edit files owned by the work unit."},
			{Name: "tests.run_local", Scope: "owned_worktree_no_network", Risk: ToolRiskGuarded, Reason: "Run local validation without network."},
			{Name: "lint.run_local", Scope: "owned_worktree_no_network", Risk: ToolRiskGuarded, Reason: "Run local lint or format checks without installing dependencies."},
			{Name: "git.diff", Scope: "owned_worktree", Risk: ToolRiskSafe, Reason: "Inspect local diff."},
		}, base...)
	case "reviewer":
		tools = append([]Tool{
			{Name: "filesystem.read", Scope: "approved_read_paths", Risk: ToolRiskSafe, Reason: "Read authorized code and evidence."},
			{Name: "git.diff", Scope: "owned_worktree", Risk: ToolRiskSafe, Reason: "Inspect proposed changes."},
			{Name: "tests.run_local", Scope: "owned_worktree_no_network", Risk: ToolRiskGuarded, Reason: "Run local validation when approved by the work unit."},
			{Name: "review.comment_structured", Scope: "current_work_unit", Risk: ToolRiskSafe, Reason: "Emit structured review findings."},
		}, base...)
	case "debugger":
		tools = append([]Tool{
			{Name: "filesystem.read", Scope: "approved_read_paths", Risk: ToolRiskSafe, Reason: "Read authorized code and logs."},
			{Name: "tests.run_local", Scope: "owned_worktree_no_network", Risk: ToolRiskGuarded, Reason: "Run local reproduction or regression checks."},
			{Name: "logs.read", Scope: "current_run_artifacts", Risk: ToolRiskSafe, Reason: "Inspect run logs and artifacts."},
			{Name: "shell.diagnostic_timeout", Scope: "owned_worktree_no_network", Risk: ToolRiskGuarded, Reason: "Run bounded diagnostic commands."},
		}, base...)
	default:
		return ToolsetSelection{}, fmt.Errorf("unknown agent profile %q", profile)
	}

	return ToolsetSelection{
		Profile:       normalized,
		Tools:         tools,
		CreatedReason: fmt.Sprintf("minimum toolset for %s profile", normalized),
	}, nil
}

// IsValidAgentProfile returns true if the profile is recognized by the system.
func IsValidAgentProfile(profile string) bool {
	normalized := normalizeProfile(profile)
	switch normalized {
	case "default", "codex", "general_engineering", "gemini":
		normalized = "code_worker"
	}
	switch normalized {
	case "fake", "docs_writer", "code_worker", "reviewer", "debugger":
		return true
	default:
		return false
	}
}

func ToolNames(tools []Tool) []string {
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}
	return names
}
