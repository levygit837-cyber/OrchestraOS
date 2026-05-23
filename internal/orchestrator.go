package internal

import (
	"context"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/executor"
	"github.com/levygit837-cyber/OrchestraOS/internal/planner"
	"github.com/levygit837-cyber/OrchestraOS/internal/store"
)

// Orchestrator coordinates the pipeline: plan → execute.
type Orchestrator struct {
	store   store.Store
	planner planner.Planner
	exec    *executor.Executor
}

func NewOrchestrator(s store.Store, p planner.Planner, e *executor.Executor) *Orchestrator {
	return &Orchestrator{store: s, planner: p, exec: e}
}

// RunTask plans and executes a task end-to-end.
func (o *Orchestrator) RunTask(ctx context.Context, taskID string) (*executor.Result, error) {
	task, err := o.store.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if err := o.store.UpdateTaskStatus(ctx, taskID, domain.TaskStatusRunning); err != nil {
		return nil, err
	}

	plan, err := o.planner.Plan(ctx, task)
	if err != nil {
		return nil, err
	}

	graph := &domain.TaskGraph{
		ID:              plan.GraphID,
		TaskID:          taskID,
		Version:         1,
		Status:          domain.TaskGraphStatusActive,
		PlannerStrategy: "heuristic",
		Rationale:       plan.Rationale,
		CreatedBy:       "orchestrator",
		NodeCount:       len(plan.WorkUnits),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := o.store.CreateTaskGraph(ctx, graph); err != nil {
		return nil, err
	}

	for i := range plan.WorkUnits {
		if err := o.store.CreateWorkUnit(ctx, &plan.WorkUnits[i]); err != nil {
			return nil, err
		}
	}

	result, err := o.exec.Execute(ctx, task, plan.WorkUnits)
	if err != nil {
		_ = o.store.UpdateTaskStatus(ctx, taskID, domain.TaskStatusFailed)
		return nil, err
	}

	if err := o.store.UpdateTaskStatus(ctx, taskID, result.Status); err != nil {
		return nil, err
	}

	return result, nil
}
