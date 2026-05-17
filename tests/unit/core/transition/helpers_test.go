package transition_test

import (
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
)

func TestRequireFinalAuditRejectsMissingData(t *testing.T) {
	if err := transition.RequireFinalAudit("completed", transition.TransitionInput{}, "test"); err == nil {
		t.Fatal("expected final state without audit data to be rejected")
	}
	if err := transition.RequireFinalAudit("completed", transition.TransitionInput{EvidenceRefs: []string{"validation:test"}}, "test"); err != nil {
		t.Fatalf("expected final state with evidence to be accepted: %v", err)
	}
}
