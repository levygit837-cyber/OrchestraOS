package daggen_test

import (
	"testing"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/daggen"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func TestBuildGraph_Success(t *testing.T) {
	t.Parallel()
	result := &domain.DecompositionResult{
		TaskID: "t1",
		Graph:  domain.DAGGraph{ID: "g1", TaskID: "t1"},
		WorkUnits: []domain.WUSpec{
			{NodeID: "n1", Title: "Auth Setup", Objective: "Set up auth", Context: domain.WUContext{Domain: "auth", Description: "Auth work"}, AcceptanceCriteria: []string{"auth works"}, DependsOn: []string{}},
			{NodeID: "n2", Title: "API Layer", Objective: "Build API", Context: domain.WUContext{Domain: "api", Description: "API work"}, AcceptanceCriteria: []string{"API responds"}, DependsOn: []string{"n1"}},
		},
	}

	g, err := daggen.BuildGraph(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Status != "validated" {
		t.Errorf("expected validated, got %s", g.Status)
	}
	if len(g.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(g.Edges))
	}
}

func TestBuildGraph_NilResult(t *testing.T) {
	t.Parallel()
	_, err := daggen.BuildGraph(nil)
	if err == nil {
		t.Fatal("expected error for nil result")
	}
}

func TestBuildGraph_ZeroWorkUnits(t *testing.T) {
	t.Parallel()
	result := &domain.DecompositionResult{TaskID: "t1", Graph: domain.DAGGraph{ID: "g1"}}
	_, err := daggen.BuildGraph(result)
	if err == nil {
		t.Fatal("expected error for zero work units")
	}
}

func TestBuildGraph_InvalidGraphRejected(t *testing.T) {
	t.Parallel()
	result := &domain.DecompositionResult{
		TaskID: "t1",
		Graph:  domain.DAGGraph{ID: "g1", TaskID: "t1"},
		WorkUnits: []domain.WUSpec{
			{NodeID: "n1", Title: "A", Objective: "do A", Context: domain.WUContext{Domain: "x"}, DependsOn: []string{"n2"}},
			{NodeID: "n2", Title: "B", Objective: "do B", Context: domain.WUContext{Domain: "y"}, DependsOn: []string{"n1"}},
		},
	}

	g, err := daggen.BuildGraph(result)
	if err == nil {
		t.Fatal("expected error for cyclic graph")
	}
	if g != nil && g.Status != "rejected" {
		t.Errorf("expected rejected status, got %s", g.Status)
	}
}

func TestBuildWorkUnits_Success(t *testing.T) {
	t.Parallel()
	task := &domain.Task{ID: "t1", Title: "Test Task"}
	graph := &domain.DAGGraph{
		ID:     "g1",
		TaskID: "t1",
		Status: "validated",
		Nodes: []domain.DAGNode{
			{ID: "n1", GraphID: "g1", Label: "Auth", Context: domain.WUContext{Domain: "auth"}, DependsOn: []string{}, CreatedAt: time.Now()},
		},
	}
	specs := []domain.WUSpec{
		{NodeID: "n1", Title: "Auth Setup", Objective: "Set up auth", Context: domain.WUContext{Domain: "auth"}, AcceptanceCriteria: []string{"auth works"}, DependsOn: []string{}, SuggestedAgent: "code_worker"},
	}

	wus, err := daggen.BuildWorkUnits(task, graph, specs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(wus) != 1 {
		t.Fatalf("expected 1 work unit, got %d", len(wus))
	}
	if wus[0].TaskID != "t1" {
		t.Errorf("expected task ID t1, got %s", wus[0].TaskID)
	}
	if wus[0].AssignedAgentProfile != "code_worker" {
		t.Errorf("expected agent code_worker, got %s", wus[0].AssignedAgentProfile)
	}
}

func TestBuildWorkUnits_NonValidatedGraph(t *testing.T) {
	t.Parallel()
	task := &domain.Task{ID: "t1"}
	graph := &domain.DAGGraph{ID: "g1", Status: "pending_validation"}
	_, err := daggen.BuildWorkUnits(task, graph, nil)
	if err == nil {
		t.Fatal("expected error for non-validated graph")
	}
}

func TestBuildWorkUnits_UnknownNode(t *testing.T) {
	t.Parallel()
	task := &domain.Task{ID: "t1"}
	graph := &domain.DAGGraph{
		ID:     "g1",
		Status: "validated",
		Nodes:  []domain.DAGNode{{ID: "n1", Label: "A", Context: domain.WUContext{Domain: "x"}}},
	}
	specs := []domain.WUSpec{
		{NodeID: "n99", Title: "Unknown", Objective: "?", Context: domain.WUContext{Domain: "x"}},
	}
	_, err := daggen.BuildWorkUnits(task, graph, specs)
	if err == nil {
		t.Fatal("expected error for unknown node reference")
	}
}
