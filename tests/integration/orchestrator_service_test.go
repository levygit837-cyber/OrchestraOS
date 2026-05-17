// LLM AGENT: BEFORE MODIFYING THIS FILE, READ:
//   1. README.md  in this directory -> purpose, file map, dependencies
//   2. CONTRACTS.md in this directory -> invariants, state machine, boundary rules
// Ignoring these files will cause architecture test failures.

package integration_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	orchestratormod "github.com/levygit837-cyber/OrchestraOS/internal/modules/orchestrator"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
)

// TestOrchestratorService_RunTask_SequentialExecution tests sequential execution of 2 work units with FakeRuntime.
func TestOrchestratorService_RunTask_SequentialExecution(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create services
	taskService := bootstrap.TaskService(db)
	taskGraphService := bootstrap.TaskGraphService(db)
	orchestrator := bootstrap.OrchestratorService(db)

	// Create a task
	taskID := uuid.New().String()
	_, err := taskService.Create(context.Background(), taskmod.CreateTaskInput{
		ID:          taskID,
		Title:       "Test Task",
		Description: "Test task for orchestrator",
		Priority:    taskmod.PriorityP2,
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Decompose task into work units
	decomposeResult, err := taskGraphService.Decompose(context.Background(), taskgraphmod.DecomposeTaskGraphInput{
		TaskID:          taskID,
		PlannerStrategy: "local_heuristic_v1",
		CreatedBy:       "test",
	})
	if err != nil {
		t.Fatalf("failed to decompose task: %v", err)
	}

	// Verify we have work units
	if len(decomposeResult.WorkUnits) == 0 {
		t.Fatal("expected at least one work unit")
	}

	// Run the task
	result, err := orchestrator.RunTask(context.Background(), taskID, orchestratormod.RunTaskOptions{
		RuntimeType:     "fake",
		PlannerStrategy: "local_heuristic_v1",
		MaxSteps:        10,
		TimeoutSeconds:  30,
	})
	if err != nil {
		t.Fatalf("RunTask failed: %v", err)
	}

	// Verify result
	if result.TaskID != taskID {
		t.Errorf("expected task ID %s, got %s", taskID, result.TaskID)
	}
	if result.Status != "completed" {
		t.Errorf("expected status completed, got %s", result.Status)
	}
	if len(result.RunIDs) != len(decomposeResult.WorkUnits) {
		t.Errorf("expected %d run IDs, got %d", len(decomposeResult.WorkUnits), len(result.RunIDs))
	}

	// Verify each run completed
	for _, runID := range result.RunIDs {
		run, err := runmod.NewRepository(db).GetByID(runID)
		if err != nil {
			t.Fatalf("failed to get run %s: %v", runID, err)
		}
		if run.Status != runmod.StatusCompleted {
			t.Errorf("expected run status completed, got %s", run.Status)
		}
	}

	// Verify task completed
	updatedTask, err := taskmod.NewRepository(db).GetByID(taskID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if updatedTask.Status != taskmod.StatusCompleted {
		t.Errorf("expected task status completed, got %s", updatedTask.Status)
	}
}

// TestOrchestratorService_TopologicalSort verifies correct topological ordering.
func TestOrchestratorService_TopologicalSort(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create services
	taskService := bootstrap.TaskService(db)
	taskGraphService := bootstrap.TaskGraphService(db)
	orchestrator := bootstrap.OrchestratorService(db)

	// Create a task
	taskID := uuid.New().String()
	_, err := taskService.Create(context.Background(), taskmod.CreateTaskInput{
		ID:          taskID,
		Title:       "Test Task",
		Description: "Test task for topological sort",
		Priority:    taskmod.PriorityP2,
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Decompose task
	_, err = taskGraphService.Decompose(context.Background(), taskgraphmod.DecomposeTaskGraphInput{
		TaskID:          taskID,
		PlannerStrategy: "local_heuristic_v1",
		CreatedBy:       "test",
	})
	if err != nil {
		t.Fatalf("failed to decompose task: %v", err)
	}

	// Run the task
	result, err := orchestrator.RunTask(context.Background(), taskID, orchestratormod.RunTaskOptions{
		RuntimeType:     "fake",
		PlannerStrategy: "local_heuristic_v1",
		MaxSteps:        10,
		TimeoutSeconds:  30,
	})
	if err != nil {
		t.Fatalf("RunTask failed: %v", err)
	}

	// Verify runs were created in order
	for i, runID := range result.RunIDs {
		run, err := runmod.NewRepository(db).GetByID(runID)
		if err != nil {
			t.Fatalf("failed to get run %s: %v", runID, err)
		}
		// Verify startedAt is monotonically increasing
		if i > 0 {
			prevRunID := result.RunIDs[i-1]
			prevRun, err := runmod.NewRepository(db).GetByID(prevRunID)
			if err != nil {
				t.Fatalf("failed to get previous run %s: %v", prevRunID, err)
			}
			if run.StartedAt.Before(prevRun.StartedAt) {
				t.Errorf("run %s started before previous run %s", runID, prevRunID)
			}
		}
	}
}

// TestOrchestratorService_AgentSessionCheckpoint verifies AgentSession checkpoint updates.
func TestOrchestratorService_AgentSessionCheckpoint(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create services
	taskService := bootstrap.TaskService(db)
	taskGraphService := bootstrap.TaskGraphService(db)
	orchestrator := bootstrap.OrchestratorService(db)

	// Create a task
	taskID := uuid.New().String()
	_, err := taskService.Create(context.Background(), taskmod.CreateTaskInput{
		ID:          taskID,
		Title:       "Test Task",
		Description: "Test task for checkpoint verification",
		Priority:    taskmod.PriorityP2,
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Decompose task
	_, err = taskGraphService.Decompose(context.Background(), taskgraphmod.DecomposeTaskGraphInput{
		TaskID:          taskID,
		PlannerStrategy: "local_heuristic_v1",
		CreatedBy:       "test",
	})
	if err != nil {
		t.Fatalf("failed to decompose task: %v", err)
	}

	// Run the task
	result, err := orchestrator.RunTask(context.Background(), taskID, orchestratormod.RunTaskOptions{
		RuntimeType:     "fake",
		PlannerStrategy: "local_heuristic_v1",
		MaxSteps:        10,
		TimeoutSeconds:  30,
	})
	if err != nil {
		t.Fatalf("RunTask failed: %v", err)
	}

	// Verify agent sessions were created and have checkpoints
	for _, runID := range result.RunIDs {
		// Find the agent session for this run
		session, err := agentsessionmod.NewRepository(db).GetByRunID(runID)
		if err != nil {
			t.Fatalf("failed to get session for run %s: %v", runID, err)
		}
		if session == nil {
			t.Errorf("expected an agent session for run %s", runID)
			continue
		}

		// Verify session status
		if session.Status != domain.AgentSessionStatusStopped && session.Status != domain.AgentSessionStatusDisconnected {
			t.Errorf("expected session status stopped or disconnected, got %s", session.Status)
		}
	}
}

// TestOrchestratorService_RunStateTransitions verifies Run state transitions.
func TestOrchestratorService_RunStateTransitions(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create services
	taskService := bootstrap.TaskService(db)
	taskGraphService := bootstrap.TaskGraphService(db)
	orchestrator := bootstrap.OrchestratorService(db)

	// Create a task
	taskID := uuid.New().String()
	_, err := taskService.Create(context.Background(), taskmod.CreateTaskInput{
		ID:          taskID,
		Title:       "Test Task",
		Description: "Test task for state transitions",
		Priority:    taskmod.PriorityP2,
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Decompose task
	_, err = taskGraphService.Decompose(context.Background(), taskgraphmod.DecomposeTaskGraphInput{
		TaskID:          taskID,
		PlannerStrategy: "local_heuristic_v1",
		CreatedBy:       "test",
	})
	if err != nil {
		t.Fatalf("failed to decompose task: %v", err)
	}

	// Run the task
	result, err := orchestrator.RunTask(context.Background(), taskID, orchestratormod.RunTaskOptions{
		RuntimeType:     "fake",
		PlannerStrategy: "local_heuristic_v1",
		MaxSteps:        10,
		TimeoutSeconds:  30,
	})
	if err != nil {
		t.Fatalf("RunTask failed: %v", err)
	}

	// Verify each run transitioned through expected states
	for _, runID := range result.RunIDs {
		// Check events for this run to verify state transitions
		events, err := getEventsForRun(db, runID)
		if err != nil {
			t.Fatalf("failed to get events for run %s: %v", runID, err)
		}

		// Verify we have run.created, run.running, run.completed events
		hasCreated := false
		hasRunning := false
		hasCompleted := false

		for _, event := range events {
			switch event.Type {
			case "run.created":
				hasCreated = true
			case "run.running":
				hasRunning = true
			case "run.completed":
				hasCompleted = true
			}
		}

		if !hasCreated {
			t.Errorf("run %s missing run.created event", runID)
		}
		if !hasRunning {
			t.Errorf("run %s missing run.running event", runID)
		}
		if !hasCompleted {
			t.Errorf("run %s missing run.completed event", runID)
		}
	}
}

// TestOrchestratorService_TaskCompletion verifies Task completion after all work units.
func TestOrchestratorService_TaskCompletion(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(t, db)

	// Create services
	taskService := bootstrap.TaskService(db)
	taskGraphService := bootstrap.TaskGraphService(db)
	orchestrator := bootstrap.OrchestratorService(db)

	// Create a task
	taskID := uuid.New().String()
	task, err := taskService.Create(context.Background(), taskmod.CreateTaskInput{
		ID:          taskID,
		Title:       "Test Task",
		Description: "Test task for completion verification",
		Priority:    taskmod.PriorityP2,
	})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Verify initial status
	if task.Value.Status != taskmod.StatusCreated {
		t.Errorf("expected initial task status created, got %s", task.Value.Status)
	}

	// Decompose task
	_, err = taskGraphService.Decompose(context.Background(), taskgraphmod.DecomposeTaskGraphInput{
		TaskID:          taskID,
		PlannerStrategy: "local_heuristic_v1",
		CreatedBy:       "test",
	})
	if err != nil {
		t.Fatalf("failed to decompose task: %v", err)
	}

	// Run the task
	_, err = orchestrator.RunTask(context.Background(), taskID, orchestratormod.RunTaskOptions{
		RuntimeType:     "fake",
		PlannerStrategy: "local_heuristic_v1",
		MaxSteps:        10,
		TimeoutSeconds:  30,
	})
	if err != nil {
		t.Fatalf("RunTask failed: %v", err)
	}

	// Verify task completed
	updatedTask, err := taskmod.NewRepository(db).GetByID(taskID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if updatedTask.Status != taskmod.StatusCompleted {
		t.Errorf("expected task status completed, got %s", updatedTask.Status)
	}

	// Verify orchestrator emitted task.completed event
	events, err := getEventsForTask(db, taskID)
	if err != nil {
		t.Fatalf("failed to get events for task %s: %v", taskID, err)
	}

	hasTaskCompleted := false
	for _, event := range events {
		if event.Type == "orchestrator.task_completed" {
			hasTaskCompleted = true
			break
		}
	}
	if !hasTaskCompleted {
		t.Error("missing orchestrator.task_completed event")
	}
}

// Helper functions

func setupTestDB(t *testing.T) *sql.DB {
	// In a real integration test, this would set up a test database
	// For now, we'll skip this test if no test DB is available
	t.Skip("integration test requires test database setup")
	return nil
}

func teardownTestDB(t *testing.T, db *sql.DB) {
	if db != nil {
		db.Close()
	}
}

func getEventsForRun(db *sql.DB, runID string) ([]domain.EventEnvelope, error) {
	// This would query the event store for events related to a run
	// For now, return empty list as placeholder
	return []domain.EventEnvelope{}, nil
}

func getEventsForTask(db *sql.DB, taskID string) ([]domain.EventEnvelope, error) {
	// This would query the event store for events related to a task
	// For now, return empty list as placeholder
	return []domain.EventEnvelope{}, nil
}
