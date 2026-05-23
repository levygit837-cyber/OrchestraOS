package decomposer_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/decomposer"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

type fakeStrategy struct {
	result *domain.DecompositionResult
	err    error
	calls  int
}

func (f *fakeStrategy) Name() string { return "fake" }

func (f *fakeStrategy) Decompose(_ context.Context, req *domain.DecompositionRequest) (*domain.DecompositionResult, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func TestPipeline_Success(t *testing.T) {
	t.Parallel()

	n1 := uuid.New().String()
	n2 := uuid.New().String()
	graphID := uuid.New().String()

	strategy := &fakeStrategy{
		result: &domain.DecompositionResult{
			TaskID: "t1",
			Graph:  domain.DAGGraph{ID: graphID, TaskID: "t1"},
			WorkUnits: []domain.WUSpec{
				{NodeID: n1, Title: "Auth", Objective: "Setup auth", Context: domain.WUContext{Domain: "auth", Description: "Auth setup"}, AcceptanceCriteria: []string{"auth works"}, DependsOn: []string{}, SuggestedAgent: "code_worker"},
				{NodeID: n2, Title: "API", Objective: "Build API", Context: domain.WUContext{Domain: "api", Description: "API layer"}, AcceptanceCriteria: []string{"API responds"}, DependsOn: []string{n1}, SuggestedAgent: "code_worker"},
			},
			Rationale: "Separated by domain context.",
			Strategy:  "fake",
			CreatedAt: time.Now(),
		},
	}

	cfg := decomposer.DefaultPipelineConfig()
	cfg.Retry.MaxRetries = 0
	p := decomposer.NewPipeline(strategy, cfg)

	task := &domain.Task{
		ID:                 "t1",
		Title:              "Add login feature",
		Description:        "Implement authentication and API endpoints",
		AcceptanceCriteria: []string{"auth works", "API responds"},
	}

	result, err := p.Run(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.WorkUnits) != 2 {
		t.Errorf("expected 2 work units, got %d", len(result.WorkUnits))
	}
	if result.Graph.Status != "validated" {
		t.Errorf("expected validated, got %s", result.Graph.Status)
	}
	if result.Strategy != "fake" {
		t.Errorf("expected fake strategy, got %s", result.Strategy)
	}
}

func TestPipeline_ToPlan(t *testing.T) {
	t.Parallel()

	n1 := uuid.New().String()
	graphID := uuid.New().String()

	strategy := &fakeStrategy{
		result: &domain.DecompositionResult{
			TaskID: "t1",
			Graph:  domain.DAGGraph{ID: graphID, TaskID: "t1"},
			WorkUnits: []domain.WUSpec{
				{NodeID: n1, Title: "Work", Objective: "Do work", Context: domain.WUContext{Domain: "core", Description: "Core work"}, AcceptanceCriteria: []string{"done"}, DependsOn: []string{}},
			},
			Rationale: "Single unit.",
			Strategy:  "fake",
		},
	}

	cfg := decomposer.DefaultPipelineConfig()
	cfg.Retry.MaxRetries = 0
	p := decomposer.NewPipeline(strategy, cfg)

	task := &domain.Task{ID: "t1", Title: "Simple task"}
	result, err := p.Run(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	plan := result.ToPlan()
	if plan.GraphID != graphID {
		t.Errorf("expected graph ID %s, got %s", graphID, plan.GraphID)
	}
	if len(plan.WorkUnits) != 1 {
		t.Errorf("expected 1 work unit, got %d", len(plan.WorkUnits))
	}
}
