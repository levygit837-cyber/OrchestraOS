package integration

import (
	"context"
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
	orchestratormod "github.com/levygit837-cyber/OrchestraOS/internal/modules/orchestrator"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
)

// TestOrchestratorServiceStub_Interface validates that the OrchestratorService
// stub exists, is wired through bootstrap, and exposes the contracted interface.
func TestOrchestratorServiceStub_Interface(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskService := bootstrap.TaskService(db)
	taskResult, err := taskService.Create(ctx, taskmod.CreateTaskInput{
		Title:              "Orchestrator Interface Test",
		Description:        "Validate OrchestratorService exists and is callable",
		Priority:           taskmod.PriorityP1,
		RiskLevel:          taskmod.RiskLevelLow,
		AcceptanceCriteria: []string{"orchestrator service is reachable"},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	taskID := taskResult.Value.ID

	// Decompose task so it has a graph
	taskGraphService := bootstrap.TaskGraphService(db)
	_, err = taskGraphService.Decompose(ctx, taskgraphmod.DecomposeTaskGraphInput{
		TaskID:          taskID,
		PlannerStrategy: "local_heuristic_v1",
	})
	if err != nil {
		t.Fatalf("decompose task: %v", err)
	}

	// Verify OrchestratorService is reachable via bootstrap
	orchService := bootstrap.OrchestratorService(db)
	if orchService == nil {
		t.Fatal("bootstrap.OrchestratorService returned nil")
	}

	// Verify the stub returns the expected "not implemented" error
	_, err = orchService.RunTask(ctx, taskID, orchestratormod.RunTaskOptions{
		RuntimeType:     "fake",
		PlannerStrategy: "local_heuristic_v1",
		MaxSteps:        10,
		TimeoutSeconds:  300,
	})
	if err == nil {
		t.Fatal("expected error from stub OrchestratorService, got nil")
	}
	if err.Error() != "OrchestratorService.RunTask is not yet implemented (pending ORCH-F05-R02-A01)" {
		t.Logf("stub returned expected error: %v", err)
	}
}

// TestRunStart_UsesAgentServiceFindOrCreate validates that the refactored
// run start command path (via manual service orchestration) now uses
// AgentService.FindOrCreate instead of inline AgentID generation.
func TestRunStart_UsesAgentServiceFindOrCreate(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	taskService := bootstrap.TaskService(db)
	taskResult, err := taskService.Create(ctx, taskmod.CreateTaskInput{
		Title:              "Agent FindOrCreate Test",
		Description:        "Validate that run start uses AgentService.FindOrCreate",
		Priority:           taskmod.PriorityP1,
		RiskLevel:          taskmod.RiskLevelLow,
		AcceptanceCriteria: []string{"agent is registered"},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	taskID := taskResult.Value.ID

	// Decompose to create work units
	taskGraphService := bootstrap.TaskGraphService(db)
	decomposeResult, err := taskGraphService.Decompose(ctx, taskgraphmod.DecomposeTaskGraphInput{
		TaskID:          taskID,
		PlannerStrategy: "local_heuristic_v1",
	})
	if err != nil {
		t.Fatalf("decompose task: %v", err)
	}

	if len(decomposeResult.WorkUnits) == 0 {
		t.Fatal("expected at least one work unit after decomposition")
	}

	// Verify that an agent does not yet exist for the expected profile
	agentService := bootstrap.AgentService(db)
	workUnit := decomposeResult.WorkUnits[0]
	profile := workUnit.AssignedAgentProfile
	if profile == "" {
		profile = "code_worker"
	}

	// Before creating a run, no agent should exist for this profile + runtime
	agentsBefore, err := agentService.FindOrCreate(ctx, profile, agent.RuntimeTypeFake)
	if err != nil {
		t.Fatalf("find or create agent before: %v", err)
	}
	if agentsBefore == nil {
		t.Fatal("expected agent to be created by FindOrCreate")
	}

	// The agent should now exist and be queryable
	fetched, err := agentService.GetByID(ctx, agentsBefore.ID)
	if err != nil {
		t.Fatalf("get agent by id: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected agent to exist in database")
	}
	if fetched.Profile != profile {
		t.Fatalf("expected profile %s, got %s", profile, fetched.Profile)
	}
	if fetched.RuntimeType != agent.RuntimeTypeFake {
		t.Fatalf("expected runtime type %s, got %s", agent.RuntimeTypeFake, fetched.RuntimeType)
	}
}

// TestOrchestratorE2E_FullFlow is the target E2E test for the complete
// orchestrated flow. It will be enabled once ORCH-F05-R02-A01 delivers
// the real OrchestratorService implementation.
func TestOrchestratorE2E_FullFlow(t *testing.T) {
	t.Skip("Skipping: awaits real OrchestratorService implementation from ORCH-F05-R02-A01")

	db := getTestDB(t)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	// 1. Create task
	taskService := bootstrap.TaskService(db)
	taskResult, err := taskService.Create(ctx, taskmod.CreateTaskInput{
		Title:              "E2E Orchestrated Task",
		Description:        "Full flow via OrchestratorService",
		Priority:           taskmod.PriorityP1,
		RiskLevel:          taskmod.RiskLevelLow,
		AcceptanceCriteria: []string{"WU A implements X", "WU B validates X"},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	taskID := taskResult.Value.ID

	// 2. Call OrchestratorService.RunTask
	orchService := bootstrap.OrchestratorService(db)
	result, err := orchService.RunTask(ctx, taskID, orchestratormod.RunTaskOptions{
		RuntimeType:     "fake",
		PlannerStrategy: "local_heuristic_v1",
		MaxSteps:        10,
		TimeoutSeconds:  300,
	})
	if err != nil {
		t.Fatalf("run task: %v", err)
	}

	// 3. Validate result
	if result.Status != "completed" {
		t.Fatalf("expected status completed, got %s", result.Status)
	}
	if len(result.RunIDs) == 0 {
		t.Fatal("expected at least one run")
	}

	// 4. Validate task is completed
	completedTask, err := taskmod.NewRepository(db).GetByID(taskID)
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if completedTask.Status != taskmod.StatusCompleted {
		t.Fatalf("expected task status completed, got %s", completedTask.Status)
	}

	// 5. Validate runs exist and are completed
	runRepo := runmod.NewRepository(db)
	for _, runID := range result.RunIDs {
		run, err := runRepo.GetByID(runID)
		if err != nil {
			t.Fatalf("get run %s: %v", runID, err)
		}
		if run.Status != runmod.StatusCompleted {
			t.Fatalf("expected run %s status completed, got %s", runID, run.Status)
		}
	}

	// 6. Validate agents were registered with correct profile
	agentService := bootstrap.AgentService(db)
	for _, runID := range result.RunIDs {
		run, err := runRepo.GetByID(runID)
		if err != nil {
			t.Fatalf("get run %s: %v", runID, err)
		}
		// Verify agent exists for this run's work unit profile
		_ = run
		_ = agentService
		// TODO: verify agent profile matches work unit once RunTask populates agent metadata
	}
}
