package store

import (
	"context"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Store is the unified persistence interface for all entities.
type Store interface {
	// Task CRUD
	CreateTask(ctx context.Context, task *domain.Task) error
	GetTask(ctx context.Context, id string) (*domain.Task, error)
	UpdateTaskStatus(ctx context.Context, id string, status domain.TaskStatus) error
	ListTasks(ctx context.Context) ([]domain.Task, error)

	// TaskGraph CRUD
	CreateTaskGraph(ctx context.Context, graph *domain.TaskGraph) error
	GetActiveTaskGraph(ctx context.Context, taskID string) (*domain.TaskGraph, error)

	// WorkUnit CRUD
	CreateWorkUnit(ctx context.Context, wu *domain.WorkUnit) error
	ListWorkUnitsByGraph(ctx context.Context, graphID string) ([]domain.WorkUnit, error)
	UpdateWorkUnitStatus(ctx context.Context, id string, status domain.WorkUnitStatus) error

	// Run CRUD
	CreateRun(ctx context.Context, run *domain.Run) error
	GetRun(ctx context.Context, id string) (*domain.Run, error)
	UpdateRun(ctx context.Context, run *domain.Run) error

	// Event append
	AppendEvent(ctx context.Context, event *domain.EventEnvelope) error
}
