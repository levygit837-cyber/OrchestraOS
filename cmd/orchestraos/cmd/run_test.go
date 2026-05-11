package cmd

import (
	"testing"

	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
)

func TestCheckpointTriggerForRuntimeEventIncludesDirectToolExecution(t *testing.T) {
	tests := map[string]agentsessionmod.CheckpointTrigger{
		"agent.tool_requested": agentsessionmod.CheckpointTriggerToolRequest,
		"agent.tool_executed":  agentsessionmod.CheckpointTriggerToolExecuted,
		"tool.completed":       agentsessionmod.CheckpointTriggerToolExecuted,
		"agent.completed":      agentsessionmod.CheckpointTriggerBeforeCompletion,
	}

	for eventType, want := range tests {
		got, ok := checkpointTriggerForRuntimeEvent(eventType)
		if !ok {
			t.Fatalf("expected %s to trigger checkpoint", eventType)
		}
		if got != want {
			t.Fatalf("expected %s to trigger %s, got %s", eventType, want, got)
		}
	}

	if _, ok := checkpointTriggerForRuntimeEvent("tool.failed"); ok {
		t.Fatal("tool.failed should be handled as failure state, not automatic checkpoint")
	}
}
