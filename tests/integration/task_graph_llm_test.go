package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/services"
)

func TestTaskGraphService_Decompose_Heuristic_Default(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := services.NewTaskService(db)
	graphService := services.NewTaskGraphService(db)

	taskResult, err := taskService.Create(ctx, services.CreateTaskInput{
		Title:              "Heuristic decomposition test",
		Description:        "Test default heuristic decomposition",
		Priority:           domain.PriorityP1,
		RiskLevel:          domain.RiskLevelLow,
		AcceptanceCriteria: []string{"Criterion 1", "Criterion 2", "Criterion 3"},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	result, err := graphService.Decompose(ctx, services.DecomposeTaskGraphInput{
		TaskID:          taskResult.Value.ID,
		PlannerStrategy: "local_heuristic_v1",
	})
	if err != nil {
		t.Fatalf("decompose task: %v", err)
	}

	if result.Graph.PlannerStrategy != "local_heuristic_v1" {
		t.Errorf("expected planner strategy local_heuristic_v1, got %s", result.Graph.PlannerStrategy)
	}
	if len(result.WorkUnits) < 1 {
		t.Errorf("expected at least 1 work unit, got %d", len(result.WorkUnits))
	}
	for _, wu := range result.WorkUnits {
		if wu.AssignedAgentProfile != "default" {
			t.Errorf("expected default profile for heuristic, got %s", wu.AssignedAgentProfile)
		}
	}
}

func TestTaskGraphService_Decompose_LLM_FallbackToHeuristic(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := services.NewTaskService(db)
	graphService := services.NewTaskGraphService(db)

	taskResult, err := taskService.Create(ctx, services.CreateTaskInput{
		Title:              "LLM fallback test",
		Description:        "Test LLM fallback when API key is missing",
		Priority:           domain.PriorityP1,
		RiskLevel:          domain.RiskLevelLow,
		AcceptanceCriteria: []string{"Criterion 1", "Criterion 2"},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	result, err := graphService.Decompose(ctx, services.DecomposeTaskGraphInput{
		TaskID:          taskResult.Value.ID,
		PlannerStrategy: "llm_gemini_v1",
	})
	if err != nil {
		t.Fatalf("decompose task: %v", err)
	}

	// Should fallback to heuristic because API key is not set in test environment
	if result.Graph.PlannerStrategy != "local_heuristic_v1" {
		t.Errorf("expected fallback to local_heuristic_v1, got %s", result.Graph.PlannerStrategy)
	}
	if !contains(result.Graph.Rationale, "fallback") && !contains(result.Graph.Rationale, "failed") {
		t.Errorf("expected rationale to mention fallback, got: %s", result.Graph.Rationale)
	}
}

func TestTaskGraphService_Decompose_UnknownStrategy(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := services.NewTaskService(db)
	graphService := services.NewTaskGraphService(db)

	taskResult, err := taskService.Create(ctx, services.CreateTaskInput{
		Title:              "Unknown strategy test",
		Description:        "Test fallback on unknown strategy",
		Priority:           domain.PriorityP1,
		RiskLevel:          domain.RiskLevelLow,
		AcceptanceCriteria: []string{"Criterion 1", "Criterion 2"},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	result, err := graphService.Decompose(ctx, services.DecomposeTaskGraphInput{
		TaskID:          taskResult.Value.ID,
		PlannerStrategy: "unknown_strategy_v999",
	})
	if err != nil {
		t.Fatalf("decompose task: %v", err)
	}

	if result.Graph.PlannerStrategy != "local_heuristic_v1" {
		t.Errorf("expected fallback to local_heuristic_v1, got %s", result.Graph.PlannerStrategy)
	}
}

func TestTaskGraphService_Decompose_ReplaceActive(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := services.NewTaskService(db)
	graphService := services.NewTaskGraphService(db)

	taskResult, err := taskService.Create(ctx, services.CreateTaskInput{
		Title:              "Replace active graph test",
		Description:        "Test replacing active graph",
		Priority:           domain.PriorityP1,
		RiskLevel:          domain.RiskLevelLow,
		AcceptanceCriteria: []string{"A", "B", "C"},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	// First decomposition
	_, err = graphService.Decompose(ctx, services.DecomposeTaskGraphInput{
		TaskID:          taskResult.Value.ID,
		PlannerStrategy: "local_heuristic_v1",
	})
	if err != nil {
		t.Fatalf("first decompose: %v", err)
	}

	// Second decomposition without ReplaceActive should fail
	_, err = graphService.Decompose(ctx, services.DecomposeTaskGraphInput{
		TaskID:          taskResult.Value.ID,
		PlannerStrategy: "local_heuristic_v1",
	})
	if err == nil {
		t.Fatal("expected error for duplicate active graph")
	}

	// Third decomposition with ReplaceActive should succeed
	result, err := graphService.Decompose(ctx, services.DecomposeTaskGraphInput{
		TaskID:          taskResult.Value.ID,
		PlannerStrategy: "local_heuristic_v1",
		ReplaceActive:   true,
	})
	if err != nil {
		t.Fatalf("replace active decompose: %v", err)
	}
	if result.Graph.Version != 2 {
		t.Errorf("expected version 2, got %d", result.Graph.Version)
	}
}

func TestTaskGraphService_Decompose_Idempotency(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := services.NewTaskService(db)
	graphService := services.NewTaskGraphService(db)

	taskResult, err := taskService.Create(ctx, services.CreateTaskInput{
		Title:              "Idempotency test",
		Description:        "Test idempotent decomposition",
		Priority:           domain.PriorityP1,
		RiskLevel:          domain.RiskLevelLow,
		AcceptanceCriteria: []string{"X", "Y"},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	eventID := uuid.New().String()

	// First call
	result1, err := graphService.Decompose(ctx, services.DecomposeTaskGraphInput{
		TaskID:          taskResult.Value.ID,
		EventID:         eventID,
		PlannerStrategy: "local_heuristic_v1",
	})
	if err != nil {
		t.Fatalf("first decompose: %v", err)
	}

	// Second call with same event ID should return duplicate
	result2, err := graphService.Decompose(ctx, services.DecomposeTaskGraphInput{
		TaskID:          taskResult.Value.ID,
		EventID:         eventID,
		PlannerStrategy: "local_heuristic_v1",
	})
	if err != nil {
		t.Fatalf("second decompose: %v", err)
	}
	if !result2.Duplicate {
		t.Error("expected duplicate result on second call")
	}
	if result1.Graph.ID != result2.Graph.ID {
		t.Error("expected same graph ID for duplicate")
	}
}
