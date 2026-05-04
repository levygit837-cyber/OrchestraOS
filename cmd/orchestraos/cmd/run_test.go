package cmd

import (
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/services"
)

func TestCheckpointTriggerForRuntimeEventIncludesDirectToolExecution(t *testing.T) {
	tests := map[string]services.CheckpointTrigger{
		"agent.tool_requested": services.CheckpointTriggerToolRequest,
		"agent.tool_executed":  services.CheckpointTriggerToolExecuted,
		"tool.completed":       services.CheckpointTriggerToolExecuted,
		"agent.completed":      services.CheckpointTriggerBeforeCompletion,
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
