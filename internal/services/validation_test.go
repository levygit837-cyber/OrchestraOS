package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func TestServiceValidationRejectsInvalidTaskInput(t *testing.T) {
	err := validateCreateTaskInput(CreateTaskInput{
		Title:     "   ",
		Priority:  domain.Priority("P9"),
		RiskLevel: domain.RiskLevelLow,
	})
	if err == nil {
		t.Fatal("expected invalid input for blank title")
	}

	err = validateCreateTaskInput(CreateTaskInput{
		Title:     "Valid",
		Priority:  domain.Priority("P9"),
		RiskLevel: domain.RiskLevelLow,
	})
	if err == nil {
		t.Fatal("expected invalid priority to be rejected")
	}
}

func TestWorkUnitDependencyValidationRejectsCycles(t *testing.T) {
	a := uuid.New().String()
	b := uuid.New().String()

	err := validateWorkUnitDependencies([]CreateWorkUnitInput{
		{ID: a, TaskID: uuid.New().String(), Title: "A", Objective: "A", AssignedAgentProfile: "default", DependsOn: []string{b}},
		{ID: b, TaskID: uuid.New().String(), Title: "B", Objective: "B", AssignedAgentProfile: "default", DependsOn: []string{a}},
	}, nil)
	if err == nil {
		t.Fatal("expected cyclic dependency to be rejected")
	}
}

func TestFinalTransitionsRequireAuditData(t *testing.T) {
	if err := requireFinalAudit("completed", TransitionInput{}, "test"); err == nil {
		t.Fatal("expected final state without audit data to be rejected")
	}
	if err := requireFinalAudit("completed", TransitionInput{EvidenceRefs: []string{"validation:test"}}, "test"); err != nil {
		t.Fatalf("expected final state with evidence to be accepted: %v", err)
	}
}
