package store

import (
	"context"
	"sync"

	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Memory is an in-memory Store implementation for testing.
type Memory struct {
	mu         sync.RWMutex
	tasks      map[string]*domain.Task
	graphs     map[string]*domain.TaskGraph
	workUnits  map[string]*domain.WorkUnit
	runs       map[string]*domain.Run
	events     []domain.EventEnvelope
	taskOrder  []string
	graphOrder []string
}

func NewMemory() *Memory {
	return &Memory{
		tasks:     make(map[string]*domain.Task),
		graphs:    make(map[string]*domain.TaskGraph),
		workUnits: make(map[string]*domain.WorkUnit),
		runs:      make(map[string]*domain.Run),
	}
}

func (m *Memory) CreateTask(_ context.Context, task *domain.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task
	m.taskOrder = append(m.taskOrder, task.ID)
	return nil
}

func (m *Memory) GetTask(_ context.Context, id string) (*domain.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tasks[id]
	if !ok {
		return nil, apperrors.New(apperrors.KindNotFound, "store.get_task", "task not found: "+id)
	}
	return t, nil
}

func (m *Memory) UpdateTaskStatus(_ context.Context, id string, status domain.TaskStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tasks[id]
	if !ok {
		return apperrors.New(apperrors.KindNotFound, "store.update_task_status", "task not found: "+id)
	}
	t.Status = status
	return nil
}

func (m *Memory) ListTasks(_ context.Context) ([]domain.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]domain.Task, 0, len(m.tasks))
	for _, id := range m.taskOrder {
		if t, ok := m.tasks[id]; ok {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (m *Memory) CreateTaskGraph(_ context.Context, graph *domain.TaskGraph) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.graphs[graph.ID] = graph
	m.graphOrder = append(m.graphOrder, graph.ID)
	return nil
}

func (m *Memory) GetActiveTaskGraph(_ context.Context, taskID string) (*domain.TaskGraph, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := len(m.graphOrder) - 1; i >= 0; i-- {
		g := m.graphs[m.graphOrder[i]]
		if g.TaskID == taskID && g.Status == domain.TaskGraphStatusActive {
			return g, nil
		}
	}
	return nil, nil
}

func (m *Memory) CreateWorkUnit(_ context.Context, wu *domain.WorkUnit) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workUnits[wu.ID] = wu
	return nil
}

func (m *Memory) ListWorkUnitsByGraph(_ context.Context, graphID string) ([]domain.WorkUnit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []domain.WorkUnit
	for _, wu := range m.workUnits {
		if wu.TaskGraphID == graphID {
			result = append(result, *wu)
		}
	}
	return result, nil
}

func (m *Memory) UpdateWorkUnitStatus(_ context.Context, id string, status domain.WorkUnitStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	wu, ok := m.workUnits[id]
	if !ok {
		return apperrors.New(apperrors.KindNotFound, "store.update_wu_status", "work unit not found: "+id)
	}
	wu.Status = status
	return nil
}

func (m *Memory) CreateRun(_ context.Context, run *domain.Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[run.ID] = run
	return nil
}

func (m *Memory) GetRun(_ context.Context, id string) (*domain.Run, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.runs[id]
	if !ok {
		return nil, apperrors.New(apperrors.KindNotFound, "store.get_run", "run not found: "+id)
	}
	return r, nil
}

func (m *Memory) UpdateRun(_ context.Context, run *domain.Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[run.ID] = run
	return nil
}

func (m *Memory) AppendEvent(_ context.Context, event *domain.EventEnvelope) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, *event)
	return nil
}
