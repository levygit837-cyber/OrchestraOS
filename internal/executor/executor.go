package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/runtime"
	"github.com/levygit837-cyber/OrchestraOS/internal/store"
)

// Result holds the outcome of executing a full DAG.
type Result struct {
	TaskID string
	Status domain.TaskStatus
	RunIDs []string
	Errors []string
}

// Executor executes a DAG of work units in topological order.
type Executor struct {
	store   store.Store
	runtime runtime.Runtime
}

func New(s store.Store, rt runtime.Runtime) *Executor {
	return &Executor{store: s, runtime: rt}
}

// Execute runs all work units in topological order for a given task graph.
func (e *Executor) Execute(ctx context.Context, task *domain.Task, workUnits []domain.WorkUnit) (*Result, error) {
	sorted, err := topologicalSort(workUnits)
	if err != nil {
		return nil, err
	}

	result := &Result{TaskID: task.ID, Status: domain.TaskStatusCompleted}

	for _, wu := range sorted {
		if err := e.store.UpdateWorkUnitStatus(ctx, wu.ID, domain.WorkUnitStatusRunning); err != nil {
			return nil, err
		}

		run := &domain.Run{
			ID:         uuid.New().String(),
			TaskID:     task.ID,
			WorkUnitID: wu.ID,
			Status:     domain.RunStatusRunning,
			Attempt:    1,
			StartedAt:  time.Now(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := e.store.CreateRun(ctx, run); err != nil {
			return nil, err
		}
		result.RunIDs = append(result.RunIDs, run.ID)

		rtResult, err := e.runtime.Execute(ctx, &wu, task)
		if err != nil {
			run.Status = domain.RunStatusFailed
			reason := err.Error()
			run.FailureReason = &reason
			runResult := domain.RunResultFailed
			run.Result = &runResult
			now := time.Now()
			run.FinishedAt = &now
			run.UpdatedAt = now
			if updateErr := e.store.UpdateRun(ctx, run); updateErr != nil {
				return nil, updateErr
			}
			if updateErr := e.store.UpdateWorkUnitStatus(ctx, wu.ID, domain.WorkUnitStatusFailed); updateErr != nil {
				return nil, updateErr
			}
			result.Status = domain.TaskStatusFailed
			result.Errors = append(result.Errors, fmt.Sprintf("wu %s: %v", wu.ID, err))
			return result, nil
		}

		run.Result = &rtResult.Status
		now := time.Now()
		run.FinishedAt = &now
		run.UpdatedAt = now

		if rtResult.Status == domain.RunResultSucceeded {
			run.Status = domain.RunStatusCompleted
			if err := e.store.UpdateRun(ctx, run); err != nil {
				return nil, err
			}
			if err := e.store.UpdateWorkUnitStatus(ctx, wu.ID, domain.WorkUnitStatusCompleted); err != nil {
				return nil, err
			}
		} else {
			run.Status = domain.RunStatusFailed
			run.FailureReason = &rtResult.FailureReason
			if err := e.store.UpdateRun(ctx, run); err != nil {
				return nil, err
			}
			if err := e.store.UpdateWorkUnitStatus(ctx, wu.ID, domain.WorkUnitStatusFailed); err != nil {
				return nil, err
			}
			result.Status = domain.TaskStatusFailed
			result.Errors = append(result.Errors, fmt.Sprintf("wu %s: %s", wu.ID, rtResult.FailureReason))
			return result, nil
		}
	}

	return result, nil
}

// topologicalSort sorts work units respecting DependsOn edges.
func topologicalSort(wus []domain.WorkUnit) ([]domain.WorkUnit, error) {
	byID := map[string]*domain.WorkUnit{}
	for i := range wus {
		byID[wus[i].ID] = &wus[i]
	}

	visited := map[string]bool{}
	visiting := map[string]bool{}
	var sorted []domain.WorkUnit

	var visit func(string) error
	visit = func(id string) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return apperrors.New(apperrors.KindValidation, "executor.sort", "cycle detected in work unit dependencies")
		}
		visiting[id] = true
		wu := byID[id]
		for _, dep := range wu.DependsOn {
			if _, ok := byID[dep]; !ok {
				return apperrors.New(apperrors.KindValidation, "executor.sort", "unknown dependency: "+dep)
			}
			if err := visit(dep); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true
		sorted = append(sorted, *wu)
		return nil
	}

	for _, wu := range wus {
		if err := visit(wu.ID); err != nil {
			return nil, err
		}
	}

	return sorted, nil
}
