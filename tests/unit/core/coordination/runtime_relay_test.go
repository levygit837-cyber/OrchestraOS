package coordination_test

import (
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/coordination"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func TestCheckpointTriggerForRuntimeEventIncludesDirectToolExecution(t *testing.T) {
	tests := map[string]domain.CheckpointTrigger{
		"agent.tool_requested": domain.CheckpointTriggerToolRequest,
		"agent.tool_executed":  domain.CheckpointTriggerToolExecuted,
		"tool.completed":       domain.CheckpointTriggerToolExecuted,
		"agent.completed":      domain.CheckpointTriggerBeforeCompletion,
	}

	for eventType, want := range tests {
		got, ok := coordination.CheckpointTriggerForRuntimeEvent(eventType)
		if !ok {
			t.Fatalf("expected %s to trigger checkpoint", eventType)
		}
		if got != want {
			t.Fatalf("expected %s to trigger %s, got %s", eventType, want, got)
		}
	}

	if _, ok := coordination.CheckpointTriggerForRuntimeEvent("tool.failed"); ok {
		t.Fatal("tool.failed should be handled as failure state, not automatic checkpoint")
	}
}
