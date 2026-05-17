package prompt_test

import (
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/modules/prompt"
)

func TestSelectToolsetProfiles(t *testing.T) {
	for _, profile := range []string{"fake", "docs_writer", "code_worker", "reviewer", "debugger"} {
		t.Run(profile, func(t *testing.T) {
			selection, err := prompt.SelectToolset(profile)
			if err != nil {
				t.Fatalf("select toolset: %v", err)
			}
			if selection.Profile == "" || len(selection.Tools) == 0 {
				t.Fatalf("expected populated toolset, got %+v", selection)
			}
			if !containsTool(selection.Tools, "toolset.request_change") {
				t.Fatalf("expected every profile to request missing tools")
			}
		})
	}
}

func TestSelectToolsetAliases(t *testing.T) {
	for _, alias := range []string{"", "default", "codex", "general_engineering"} {
		selection, err := prompt.SelectToolset(alias)
		if err != nil {
			t.Fatalf("select alias %q: %v", alias, err)
		}
		if selection.Profile != "code_worker" {
			t.Fatalf("expected alias %q to resolve to code_worker, got %s", alias, selection.Profile)
		}
	}
}

func containsTool(tools []prompt.Tool, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
}
