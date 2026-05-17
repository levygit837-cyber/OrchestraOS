package taskgraph

import (
	"context"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
)

// GraphPlan is the result of any Planner implementation.
type GraphPlan struct {
	GraphID   string
	WorkUnits []PlanWorkUnit
	Nodes     []domain.TaskGraphNodeInfo
	Edges     []domain.TaskGraphEdgeInfo
	Rationale string
}

// Planner decomposes a Task into an acyclic graph of WorkUnits.
type Planner interface {
	Plan(ctx context.Context, task *task.Task) (*GraphPlan, error)
}
