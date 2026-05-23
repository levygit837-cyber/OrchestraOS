package daggen_test

import (
	"testing"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/daggen"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func TestValidate_ValidGraph(t *testing.T) {
	t.Parallel()
	g := &domain.DAGGraph{
		ID:     "g1",
		TaskID: "t1",
		Nodes: []domain.DAGNode{
			{ID: "n1", GraphID: "g1", Label: "Auth", Context: domain.WUContext{Domain: "auth"}, DependsOn: []string{}, CreatedAt: time.Now()},
			{ID: "n2", GraphID: "g1", Label: "API", Context: domain.WUContext{Domain: "api"}, DependsOn: []string{"n1"}, CreatedAt: time.Now()},
		},
		Edges: []domain.DAGEdge{{From: "n1", To: "n2"}},
	}
	if err := daggen.Validate(g); err != nil {
		t.Fatalf("expected valid graph, got: %v", err)
	}
}

func TestValidate_EmptyGraph(t *testing.T) {
	t.Parallel()
	g := &domain.DAGGraph{ID: "g1", TaskID: "t1"}
	if err := daggen.Validate(g); err == nil {
		t.Fatal("expected error for empty graph")
	}
}

func TestValidate_DuplicateNodeIDs(t *testing.T) {
	t.Parallel()
	g := &domain.DAGGraph{
		ID:     "g1",
		TaskID: "t1",
		Nodes: []domain.DAGNode{
			{ID: "n1", Label: "A", Context: domain.WUContext{Domain: "x"}},
			{ID: "n1", Label: "B", Context: domain.WUContext{Domain: "y"}},
		},
	}
	if err := daggen.Validate(g); err == nil {
		t.Fatal("expected error for duplicate node IDs")
	}
}

func TestValidate_CycleDetected(t *testing.T) {
	t.Parallel()
	g := &domain.DAGGraph{
		ID:     "g1",
		TaskID: "t1",
		Nodes: []domain.DAGNode{
			{ID: "n1", Label: "A", Context: domain.WUContext{Domain: "x"}, DependsOn: []string{"n2"}},
			{ID: "n2", Label: "B", Context: domain.WUContext{Domain: "y"}, DependsOn: []string{"n1"}},
		},
		Edges: []domain.DAGEdge{{From: "n2", To: "n1"}, {From: "n1", To: "n2"}},
	}
	if err := daggen.Validate(g); err == nil {
		t.Fatal("expected error for cyclic graph")
	}
}

func TestValidate_UnknownEdgeRef(t *testing.T) {
	t.Parallel()
	g := &domain.DAGGraph{
		ID:     "g1",
		TaskID: "t1",
		Nodes: []domain.DAGNode{
			{ID: "n1", Label: "A", Context: domain.WUContext{Domain: "x"}},
		},
		Edges: []domain.DAGEdge{{From: "n1", To: "n99"}},
	}
	if err := daggen.Validate(g); err == nil {
		t.Fatal("expected error for unknown edge reference")
	}
}

func TestValidate_SelfRefEdge(t *testing.T) {
	t.Parallel()
	g := &domain.DAGGraph{
		ID:     "g1",
		TaskID: "t1",
		Nodes: []domain.DAGNode{
			{ID: "n1", Label: "A", Context: domain.WUContext{Domain: "x"}},
		},
		Edges: []domain.DAGEdge{{From: "n1", To: "n1"}},
	}
	if err := daggen.Validate(g); err == nil {
		t.Fatal("expected error for self-referencing edge")
	}
}

func TestValidate_MissingLabel(t *testing.T) {
	t.Parallel()
	g := &domain.DAGGraph{
		ID:     "g1",
		TaskID: "t1",
		Nodes: []domain.DAGNode{
			{ID: "n1", Label: "", Context: domain.WUContext{Domain: "x"}},
		},
	}
	if err := daggen.Validate(g); err == nil {
		t.Fatal("expected error for missing label")
	}
}

func TestValidate_MissingDomain(t *testing.T) {
	t.Parallel()
	g := &domain.DAGGraph{
		ID:     "g1",
		TaskID: "t1",
		Nodes: []domain.DAGNode{
			{ID: "n1", Label: "A", Context: domain.WUContext{Domain: ""}},
		},
	}
	if err := daggen.Validate(g); err == nil {
		t.Fatal("expected error for missing domain")
	}
}
