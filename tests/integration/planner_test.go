package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
)

func TestPlannerValidator_ValidPlan(t *testing.T) {
	plan := &taskgraphmod.GraphPlan{
		GraphID: uuid.New().String(),
		WorkUnits: []domain.WorkUnit{
			{
				ID:                   uuid.New().String(),
				Title:                "Work Unit 1",
				Objective:            "Objective 1",
				AssignedAgentProfile: "code_worker",
				AcceptanceCriteria:   []string{"Criterion 1"},
				ValidationPlan:       []string{"Validate 1"},
				DependsOn:            []string{},
			},
			{
				ID:                   uuid.New().String(),
				Title:                "Work Unit 2",
				Objective:            "Objective 2",
				AssignedAgentProfile: "docs_writer",
				AcceptanceCriteria:   []string{"Criterion 2"},
				ValidationPlan:       []string{"Validate 2"},
				DependsOn:            []string{},
			},
		},
	}
	if err := bootstrap.ValidateGraphPlan(plan); err != nil {
		t.Fatalf("expected valid plan, got error: %v", err)
	}
}

func TestPlannerValidator_CycleDetected(t *testing.T) {
	wu1ID := uuid.New().String()
	wu2ID := uuid.New().String()
	wu3ID := uuid.New().String()

	plan := &taskgraphmod.GraphPlan{
		GraphID: uuid.New().String(),
		WorkUnits: []domain.WorkUnit{
			{
				ID:                   wu1ID,
				Title:                "Work Unit 1",
				Objective:            "Objective 1",
				AssignedAgentProfile: "code_worker",
				AcceptanceCriteria:   []string{"Criterion 1"},
				ValidationPlan:       []string{"Validate 1"},
				DependsOn:            []string{wu3ID},
			},
			{
				ID:                   wu2ID,
				Title:                "Work Unit 2",
				Objective:            "Objective 2",
				AssignedAgentProfile: "code_worker",
				AcceptanceCriteria:   []string{"Criterion 2"},
				ValidationPlan:       []string{"Validate 2"},
				DependsOn:            []string{wu1ID},
			},
			{
				ID:                   wu3ID,
				Title:                "Work Unit 3",
				Objective:            "Objective 3",
				AssignedAgentProfile: "code_worker",
				AcceptanceCriteria:   []string{"Criterion 3"},
				ValidationPlan:       []string{"Validate 3"},
				DependsOn:            []string{wu2ID},
			},
		},
	}

	err := bootstrap.ValidateGraphPlan(plan)
	if err == nil {
		t.Fatal("expected cycle detection error, got nil")
	}
	if !contains(err.Error(), "cycle") {
		t.Fatalf("expected error to mention cycle, got: %v", err)
	}
}

func TestPlannerValidator_InvalidProfile(t *testing.T) {
	plan := &taskgraphmod.GraphPlan{
		GraphID: uuid.New().String(),
		WorkUnits: []domain.WorkUnit{
			{
				ID:                   uuid.New().String(),
				Title:                "Work Unit 1",
				Objective:            "Objective 1",
				AssignedAgentProfile: "hacker",
				AcceptanceCriteria:   []string{"Criterion 1"},
				ValidationPlan:       []string{"Validate 1"},
				DependsOn:            []string{},
			},
		},
	}

	err := bootstrap.ValidateGraphPlan(plan)
	if err == nil {
		t.Fatal("expected invalid profile error, got nil")
	}
	if !contains(err.Error(), "profile") {
		t.Fatalf("expected error to mention profile, got: %v", err)
	}
}

func TestPlannerValidator_CountBounds(t *testing.T) {
	// Empty plan
	emptyPlan := &taskgraphmod.GraphPlan{
		GraphID:   uuid.New().String(),
		WorkUnits: []domain.WorkUnit{},
	}
	if err := bootstrap.ValidateGraphPlan(emptyPlan); err == nil {
		t.Fatal("expected error for empty plan, got nil")
	}

	// Too many work units
	var tooMany []domain.WorkUnit
	for i := 0; i < 11; i++ {
		tooMany = append(tooMany, domain.WorkUnit{
			ID:                   uuid.New().String(),
			Title:                "Work Unit",
			Objective:            "Objective",
			AssignedAgentProfile: "code_worker",
			AcceptanceCriteria:   []string{"Criterion"},
			ValidationPlan:       []string{"Validate"},
			DependsOn:            []string{},
		})
	}
	bigPlan := &taskgraphmod.GraphPlan{
		GraphID:   uuid.New().String(),
		WorkUnits: tooMany,
	}
	if err := bootstrap.ValidateGraphPlan(bigPlan); err == nil {
		t.Fatal("expected error for too many work units, got nil")
	}
}

func TestPlannerValidator_MissingDependency(t *testing.T) {
	plan := &taskgraphmod.GraphPlan{
		GraphID: uuid.New().String(),
		WorkUnits: []domain.WorkUnit{
			{
				ID:                   uuid.New().String(),
				Title:                "Work Unit 1",
				Objective:            "Objective 1",
				AssignedAgentProfile: "code_worker",
				AcceptanceCriteria:   []string{"Criterion 1"},
				ValidationPlan:       []string{"Validate 1"},
				DependsOn:            []string{"non-existent-id"},
			},
		},
	}

	err := bootstrap.ValidateGraphPlan(plan)
	if err == nil {
		t.Fatal("expected missing dependency error, got nil")
	}
	if !contains(err.Error(), "unknown") {
		t.Fatalf("expected error to mention unknown dependency, got: %v", err)
	}
}

func TestPlannerValidator_DependencyOrder(t *testing.T) {
	wu1ID := uuid.New().String()
	wu2ID := uuid.New().String()

	plan := &taskgraphmod.GraphPlan{
		GraphID: uuid.New().String(),
		WorkUnits: []domain.WorkUnit{
			{
				ID:                   wu1ID,
				Title:                "Work Unit 1",
				Objective:            "Objective 1",
				AssignedAgentProfile: "code_worker",
				AcceptanceCriteria:   []string{"Criterion 1"},
				ValidationPlan:       []string{"Validate 1"},
				DependsOn:            []string{},
			},
			{
				ID:                   wu2ID,
				Title:                "Work Unit 2",
				Objective:            "Objective 2",
				AssignedAgentProfile: "reviewer",
				AcceptanceCriteria:   []string{"Criterion 2"},
				ValidationPlan:       []string{"Validate 2"},
				DependsOn:            []string{wu1ID},
			},
		},
	}

	if err := bootstrap.ValidateGraphPlan(plan); err != nil {
		t.Fatalf("expected valid plan with dependency, got error: %v", err)
	}
}

func TestPlannerPrompt_Render(t *testing.T) {
	task := &domain.Task{
		Title:              "Test Task",
		Description:        "A task for testing",
		Priority:           domain.PriorityP1,
		RiskLevel:          domain.RiskLevelLow,
		AcceptanceCriteria: []string{"Criterion A", "Criterion B"},
	}

	prompt, err := bootstrap.PlannerPrompt(task)
	if err != nil {
		t.Fatalf("failed to render planner prompt: %v", err)
	}

	if !contains(prompt, "Test Task") {
		t.Error("prompt should contain task title")
	}
	if !contains(prompt, "A task for testing") {
		t.Error("prompt should contain task description")
	}
	if !contains(prompt, "Criterion A") {
		t.Error("prompt should contain acceptance criteria")
	}
	if !contains(prompt, "code_worker") {
		t.Error("prompt should mention valid agent profiles")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
