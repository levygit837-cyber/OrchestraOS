package agent_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
)

func TestNewGeminiRuntime(t *testing.T) {
	rt := agent.NewGeminiRuntime()
	if rt == nil {
		t.Fatal("expected non-nil runtime")
	}
	if rt.Started {
		t.Error("expected runtime to not be started")
	}
	status := rt.Status()
	if status.State != "" {
		t.Errorf("expected empty initial state, got %q", status.State)
	}
}

func TestGeminiRuntime_Start_RequiresAPIKey(t *testing.T) {
	// Ensure no API key is set.
	_ = os.Unsetenv("GEMINI_API_KEY")
	_ = os.Unsetenv("GOOGLE_API_KEY")

	rt := agent.NewGeminiRuntime()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := rt.Start(ctx, agent.RuntimeConfig{
		RunID:      "run-test-001",
		WorkUnitID: "wu-test-001",
		TaskID:     "task-test-001",
		AgentID:    "agent-test-001",
		Prompt:     "test prompt",
	})
	if err == nil {
		t.Fatal("expected error when starting without API key")
	}
}

func TestSanitizeFunctionName(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"filesystem.read", "filesystem.read"},
		{"tests.run_local", "tests.run_local"},
		{"shell.diagnostic_timeout", "shell.diagnostic_timeout"},
		{"", "tool"},
		{"123tool", "_123tool"},
		{"tool/with/slash", "tool_with_slash"},
		{"tool:with:colon", "tool_with_colon"},
		{"a_very_long_tool_name_that_exceeds_the_maximum_allowed_length_of_64_characters", "a_very_long_tool_name_that_exceeds_the_maximum_allowed_length_of"},
	}

	for _, tc := range cases {
		got := agent.SanitizeFunctionName(tc.input)
		if got != tc.expected {
			t.Errorf("SanitizeFunctionName(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestGeminiRuntime_buildTools(t *testing.T) {
	rt := agent.NewGeminiRuntime()
	rt.Config = agent.RuntimeConfig{
		Toolset: []string{
			"filesystem.read",
			"tests.run_local",
		},
	}

	tools := rt.BuildTools()
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool group, got %d", len(tools))
	}

	decls := tools[0].FunctionDeclarations
	if len(decls) != 2 {
		t.Fatalf("expected 2 function declarations, got %d", len(decls))
	}

	if decls[0].Name != "filesystem.read" {
		t.Errorf("expected first declaration name %q, got %q", "filesystem.read", decls[0].Name)
	}
	if decls[1].Name != "tests.run_local" {
		t.Errorf("expected second declaration name %q, got %q", "tests.run_local", decls[1].Name)
	}

	if decls[0].Parameters == nil {
		t.Error("expected parameters schema for first declaration")
	}
}

func TestGeminiRuntime_buildTools_Empty(t *testing.T) {
	rt := agent.NewGeminiRuntime()
	rt.Config = agent.RuntimeConfig{Toolset: []string{}}

	tools := rt.BuildTools()
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}
}

func TestGeminiRuntime_ReceiveEvent_BeforeStart(t *testing.T) {
	rt := agent.NewGeminiRuntime()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := rt.ReceiveEvent(ctx)
	if err == nil {
		t.Error("expected error receiving event before start")
	}
}

func TestGeminiRuntime_SendEvent_WithoutStart(t *testing.T) {
	rt := agent.NewGeminiRuntime()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	event := &domain.EventEnvelope{
		Type: "tool.approved",
	}
	// Should not block or panic even if runtime not started.
	err := rt.SendEvent(ctx, event)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
