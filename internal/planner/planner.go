package planner

import (
	"context"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Plan is the result of task decomposition into a DAG of WorkUnits.
type Plan struct {
	GraphID   string
	WorkUnits []domain.WorkUnit
	Rationale string
}

// Planner decomposes a Task into an acyclic graph of WorkUnits.
type Planner interface {
	Plan(ctx context.Context, task *domain.Task) (*Plan, error)
}
