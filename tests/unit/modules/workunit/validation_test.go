package workunit_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
)

func TestValidateWorkUnitDependenciesRejectsCycles(t *testing.T) {
	a := uuid.New().String()
	b := uuid.New().String()

	err := workunit.ValidateWorkUnitDependencies([]workunit.CreateWorkUnitInput{
		{ID: a, TaskID: uuid.New().String(), Title: "A", Objective: "A", AssignedAgentProfile: "default", DependsOn: []string{b}},
		{ID: b, TaskID: uuid.New().String(), Title: "B", Objective: "B", AssignedAgentProfile: "default", DependsOn: []string{a}},
	}, nil)
	if err == nil {
		t.Fatal("expected cyclic dependency to be rejected")
	}
}
