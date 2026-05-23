package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
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
	runtime domain.Runtime
}

func New(s store.Store, rt domain.Runtime) *Executor {
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
		if err := e.executeWorkUnit(ctx, task, &wu, result); err != nil {
			return nil, err
		}
		if result.Status == domain.TaskStatusFailed {
			return result, nil
		}
	}

	return result, nil
}

func (e *Executor) executeWorkUnit(ctx context.Context, task *domain.Task, wu *domain.WorkUnit, result *Result) error {
	if err := e.store.UpdateWorkUnitStatus(ctx, wu.ID, domain.WorkUnitStatusRunning); err != nil {
		return err
	}

	run := newRun(task.ID, wu.ID)
	if err := e.store.CreateRun(ctx, run); err != nil {
		return err
	}
	result.RunIDs = append(result.RunIDs, run.ID)

	rtResult, err := e.runtime.Execute(ctx, wu, task)
	if err != nil {
		return e.failRun(ctx, run, wu.ID, err.Error(), result, fmt.Sprintf("wu %s: %v", wu.ID, err))
	}

	return e.finalizeRun(ctx, run, wu.ID, rtResult, result)
}

func newRun(taskID, wuID string) *domain.Run {
	now := time.Now()
	return &domain.Run{
		ID:         uuid.New().String(),
		TaskID:     taskID,
		WorkUnitID: wuID,
		Status:     domain.RunStatusRunning,
		Attempt:    1,
		StartedAt:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (e *Executor) failRun(ctx context.Context, run *domain.Run, wuID, reason string, result *Result, errMsg string) error {
	run.Status = domain.RunStatusFailed
	run.FailureReason = &reason
	failed := domain.RunResultFailed
	run.Result = &failed
	now := time.Now()
	run.FinishedAt = &now
	run.UpdatedAt = now
	if err := e.store.UpdateRun(ctx, run); err != nil {
		return err
	}
	if err := e.store.UpdateWorkUnitStatus(ctx, wuID, domain.WorkUnitStatusFailed); err != nil {
		return err
	}
	result.Status = domain.TaskStatusFailed
	result.Errors = append(result.Errors, errMsg)
	return nil
}

func (e *Executor) finalizeRun(ctx context.Context, run *domain.Run, wuID string, rt *domain.RuntimeResult, result *Result) error {
	run.Result = &rt.Status
	now := time.Now()
	run.FinishedAt = &now
	run.UpdatedAt = now

	if rt.Status == domain.RunResultSucceeded {
		run.Status = domain.RunStatusCompleted
		if err := e.store.UpdateRun(ctx, run); err != nil {
			return err
		}
		return e.store.UpdateWorkUnitStatus(ctx, wuID, domain.WorkUnitStatusCompleted)
	}

	return e.failRun(ctx, run, wuID, rt.FailureReason, result, fmt.Sprintf("wu %s: %s", wuID, rt.FailureReason))
}

// topologicalSort sorts work units respecting DependsOn edges.
func topologicalSort(wus []domain.WorkUnit) ([]domain.WorkUnit, error) {
	byID := map[string]*domain.WorkUnit{}
	for i := range wus {
		byID[wus[i].ID] = &wus[i]
	}

	state := &topoState{byID: byID, visiting: map[string]bool{}, visited: map[string]bool{}}
	for _, wu := range wus {
		if err := state.visit(wu.ID); err != nil {
			return nil, err
		}
	}
	return state.sorted, nil
}

type topoState struct {
	byID     map[string]*domain.WorkUnit
	visiting map[string]bool
	visited  map[string]bool
	sorted   []domain.WorkUnit
}

func (s *topoState) visit(id string) error {
	if s.visited[id] {
		return nil
	}
	if s.visiting[id] {
		return apperrors.New(apperrors.KindValidation, "executor.sort", "cycle detected in work unit dependencies")
	}
	s.visiting[id] = true
	wu := s.byID[id]
	for _, dep := range wu.DependsOn {
		if _, ok := s.byID[dep]; !ok {
			return apperrors.New(apperrors.KindValidation, "executor.sort", "unknown dependency: "+dep)
		}
		if err := s.visit(dep); err != nil {
			return err
		}
	}
	s.visiting[id] = false
	s.visited[id] = true
	s.sorted = append(s.sorted, *wu)
	return nil
}
