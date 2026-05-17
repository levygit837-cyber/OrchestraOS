package taskgraph_test

import (
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
)

func TestLocalHeuristicDecomposesTwoCriteria(t *testing.T) {
	task := taskgraph.TaskForGraphTest([]string{
		"Criar schema do task graph",
		"Criar repository do task graph",
	})

	plan, err := taskgraph.BuildLocalHeuristicGraphPlan(task)
	if err != nil {
		t.Fatalf("expected graph plan: %v", err)
	}
	if len(plan.WorkUnits) != 2 {
		t.Fatalf("expected 2 work units, got %d", len(plan.WorkUnits))
	}
	for _, wu := range plan.WorkUnits {
		if wu.TaskID != task.ID || wu.TaskGraphID != plan.GraphID {
			t.Fatalf("expected work unit to reference task and graph, got %+v", wu)
		}
		if len(wu.AcceptanceCriteria) != 1 {
			t.Fatalf("expected one criterion per work unit, got %+v", wu.AcceptanceCriteria)
		}
	}
}

func TestLocalHeuristicLimitsWorkUnits(t *testing.T) {
	task := taskgraph.TaskForGraphTest([]string{
		"Criterio um pronto",
		"Criterio dois pronto",
		"Criterio tres pronto",
		"Criterio quatro pronto",
		"Criterio cinco pronto",
		"Criterio seis pronto",
	})

	plan, err := taskgraph.BuildLocalHeuristicGraphPlan(task)
	if err != nil {
		t.Fatalf("expected graph plan: %v", err)
	}
	if len(plan.WorkUnits) > taskgraph.MaxGraphWorkUnits {
		t.Fatalf("expected at most %d work units, got %d", taskgraph.MaxGraphWorkUnits, len(plan.WorkUnits))
	}
}

func TestLocalHeuristicRejectsInsufficientInput(t *testing.T) {
	if _, err := taskgraph.BuildLocalHeuristicGraphPlan(taskgraph.TaskForGraphTest(nil)); err == nil {
		t.Fatal("expected empty acceptance criteria to be rejected")
	}
	if _, err := taskgraph.BuildLocalHeuristicGraphPlan(taskgraph.TaskForGraphTest([]string{"Somente um criterio"})); err == nil {
		t.Fatal("expected single acceptance criterion to be rejected")
	}
}

func TestLocalHeuristicRejectsUnbalancedWorkUnits(t *testing.T) {
	task := taskgraph.TaskForGraphTest([]string{
		"Curto",
		"Este criterio tem muitas palavras para deixar a work unit muito maior que a outra parte do plano",
	})

	if _, err := taskgraph.BuildLocalHeuristicGraphPlan(task); err == nil {
		t.Fatal("expected unbalanced criteria to be rejected")
	}
}

func TestLocalHeuristicCreatesExplicitDependencies(t *testing.T) {
	task := taskgraph.TaskForGraphTest([]string{
		"Criar schema",
		"[after: 1] Criar repositorio",
	})

	plan, err := taskgraph.BuildLocalHeuristicGraphPlan(task)
	if err != nil {
		t.Fatalf("expected graph plan: %v", err)
	}
	if len(plan.Edges) != 1 {
		t.Fatalf("expected one edge, got %d", len(plan.Edges))
	}
	if plan.WorkUnits[1].DependsOn[0] != plan.WorkUnits[0].ID {
		t.Fatalf("expected second work unit to depend on first, got %+v", plan.WorkUnits[1].DependsOn)
	}
}

func TestLocalHeuristicRejectsUnknownDependency(t *testing.T) {
	task := taskgraph.TaskForGraphTest([]string{
		"Criar schema",
		"[after: 3] Criar repositorio",
	})

	if _, err := taskgraph.BuildLocalHeuristicGraphPlan(task); err == nil {
		t.Fatal("expected unknown dependency to be rejected")
	}
}

func TestLocalHeuristicRejectsDependencyCycle(t *testing.T) {
	task := taskgraph.TaskForGraphTest([]string{
		"[after: 2] Criar schema",
		"[after: 1] Criar repositorio",
	})

	if _, err := taskgraph.BuildLocalHeuristicGraphPlan(task); err == nil {
		t.Fatal("expected cycle to be rejected")
	}
}
