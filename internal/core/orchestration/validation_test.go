package orchestration

import "testing"

func TestRequireFinalAuditRejectsMissingData(t *testing.T) {
	if err := RequireFinalAudit("completed", TransitionInput{}, "test"); err == nil {
		t.Fatal("expected final state without audit data to be rejected")
	}
	if err := RequireFinalAudit("completed", TransitionInput{EvidenceRefs: []string{"validation:test"}}, "test"); err != nil {
		t.Fatalf("expected final state with evidence to be accepted: %v", err)
	}
}
