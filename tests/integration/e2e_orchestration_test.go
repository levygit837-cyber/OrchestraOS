package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/orchestration"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/services"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	promptmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/prompt"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
	_ "github.com/lib/pq"
)

// TestE2EFakeRuntimeTaskToComplete validates the full flow:
// Task → Graph → Run → Session → FakeRuntime → Complete
func TestE2EFakeRuntimeTaskToComplete(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()
	taskService := bootstrap.TaskService(db)
	taskGraphService := bootstrap.TaskGraphService(db)
	runService := bootstrap.RunService(db)
	sessionService := bootstrap.AgentSessionService(db)
	promptService := bootstrap.PromptService(db)
	relay := bootstrap.RuntimeEventRelay(db)

	// 1. Create task
	taskResult, err := taskService.Create(ctx, taskmod.CreateTaskInput{
		Title:       "E2E Integration Test Task",
		Description: "Validate full orchestration flow",
		Priority:    domain.PriorityP1,
		RiskLevel:   domain.RiskLevelLow,
		AcceptanceCriteria: []string{
			"Work unit can be created",
			"Runtime can execute",
			"Events are persisted",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	task := taskResult.Value

	// 2. Decompose task into graph
	graphResult, err := taskGraphService.Decompose(ctx, taskgraphmod.DecomposeTaskGraphInput{
		TaskID:          task.ID,
		ReplaceActive:   false,
		CreatedBy:       "e2e_test",
		PlannerStrategy: "local_heuristic_v1",
	})
	if err != nil {
		t.Fatalf("Failed to decompose task: %v", err)
	}
	if len(graphResult.WorkUnits) == 0 {
		t.Fatal("Expected at least one work unit from decomposition")
	}
	wu := graphResult.WorkUnits[0]

	// 3. Create and start run
	runtimeType := "fake"
	runResult, err := runService.Create(ctx, runmod.CreateRunInput{
		TaskID:     task.ID,
		WorkUnitID: wu.ID,
		Attempt:    1,
	})
	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}
	run := runResult.Value
	if _, err := runService.Start(ctx, run.ID, orchestration.TransitionInput{Runtime: runtimeType}); err != nil {
		t.Fatalf("Failed to start run: %v", err)
	}

	// 4. Create and connect agent session
	agentID := fmt.Sprintf("agent-e2e-%s", uuid.New().String()[:8])
	sessionResult, err := sessionService.Create(ctx, agentsessionmod.CreateAgentSessionInput{
		AgentID: agentID,
		RunID:   run.ID,
	})
	if err != nil {
		t.Fatalf("Failed to create agent session: %v", err)
	}
	session := sessionResult.Value
	connectionID := fmt.Sprintf("conn-e2e-%s", uuid.New().String()[:8])
	if _, err := sessionService.Connect(ctx, session.ID, connectionID, "", orchestration.TransitionInput{Runtime: runtimeType}); err != nil {
		t.Fatalf("Failed to connect agent session: %v", err)
	}

	// 5. Prepare prompt
	preparedPrompt, err := promptService.PrepareRunPrompt(ctx, promptmod.PrepareRunPromptInput{
		RunID:          run.ID,
		AgentSessionID: session.ID,
	})
	if err != nil {
		t.Fatalf("Failed to prepare prompt: %v", err)
	}

	// 6. Start FakeRuntime
	fakeRuntime := agent.NewFakeRuntime()
	config := agent.RuntimeConfig{
		RunID:             run.ID,
		WorkUnitID:        wu.ID,
		TaskID:            task.ID,
		AgentID:           agentID,
		Prompt:            preparedPrompt.CombinedPrompt,
		SystemPrompt:      preparedPrompt.SystemPrompt,
		TaskPrompt:        preparedPrompt.TaskPrompt,
		PromptSnapshotID:  preparedPrompt.PromptSnapshot.ID,
		ToolsetSnapshotID: preparedPrompt.ToolsetSnapshot.ID,
		PromptHash:        preparedPrompt.PromptHash,
		Toolset:           preparedPrompt.Toolset,
		MaxSteps:          10,
		Timeout:           300,
	}

	rtCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := fakeRuntime.Start(rtCtx, config); err != nil {
		t.Fatalf("Failed to start fake runtime: %v", err)
	}

	// 7. Run relay
	relayConfig := services.RelayConfig{
		SessionID:   session.ID,
		RunID:       run.ID,
		RuntimeType: runtimeType,
		AgentID:     agentID,
	}

	finalStatus, err := relay.Run(rtCtx, fakeRuntime, relayConfig)
	if err != nil {
		t.Fatalf("Relay failed: %v", err)
	}
	if finalStatus != domain.RunStatusCompleted {
		t.Fatalf("Expected run status %s, got %s", domain.RunStatusCompleted, finalStatus)
	}

	// 8. Assertions
	// Run status
	runRepo := runmod.NewRepository(db)
	finalRun, err := runRepo.GetByID(run.ID)
	if err != nil {
		t.Fatalf("Failed to get final run: %v", err)
	}
	if finalRun.Status != domain.RunStatusCompleted {
		t.Errorf("Expected run status %s, got %s", domain.RunStatusCompleted, finalRun.Status)
	}

	// Work unit status
	wuRepo := workunitmod.NewRepository(db)
	finalWU, err := wuRepo.GetByID(wu.ID)
	if err != nil {
		t.Fatalf("Failed to get final work unit: %v", err)
	}
	if finalWU.Status != domain.WorkUnitStatusCompleted {
		t.Errorf("Expected work unit status %s, got %s", domain.WorkUnitStatusCompleted, finalWU.Status)
	}

	// Agent session checkpoint
	sessionRepo := agentsessionmod.NewRepository(db)
	finalSession, err := sessionRepo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("Failed to get final session: %v", err)
	}
	if finalSession.LastCheckpointAt == nil {
		t.Error("Expected last_checkpoint_at to be set")
	}
	if finalSession.LastHeartbeatAt == nil {
		t.Error("Expected last_heartbeat_at to be set")
	}

	// Events persisted
	eventStore, err := eventstore.NewStore(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	events, err := eventStore.ListByRun(run.ID)
	if err != nil {
		t.Fatalf("Failed to list events: %v", err)
	}
	if len(events) == 0 {
		t.Error("Expected events to be persisted")
	}

	requiredTypes := map[string]bool{
		"run.started":              false,
		"agent.session_starting":   false,
		"agent.checkpoint_reached": false,
		"agent.completed":          false,
		"run.completed":            false,
	}
	// Heartbeat is best-effort: if the runtime completes faster than the
	// 5s heartbeat interval, no heartbeat event is emitted.
	hasHeartbeat := false
	for _, e := range events {
		if _, exists := requiredTypes[e.Type]; exists {
			requiredTypes[e.Type] = true
		}
		if e.Type == "agent.heartbeat" {
			hasHeartbeat = true
		}
	}
	for eventType, found := range requiredTypes {
		if !found {
			t.Errorf("Expected event type %s was not found", eventType)
		}
	}
	if !hasHeartbeat {
		t.Logf("No heartbeat event found (runtime completed faster than heartbeat interval)")
	}

	// Replay validation
	replayState, err := eventStore.ReplayRunState(run.ID)
	if err != nil {
		t.Fatalf("Failed to replay run: %v", err)
	}
	if replayState == nil {
		t.Fatal("Expected replay state to be non-nil")
	}
	if runStatus, ok := replayState.RunStatuses[run.ID]; !ok || runStatus != domain.RunStatusCompleted {
		t.Errorf("Expected replay run status %s, got %v", domain.RunStatusCompleted, runStatus)
	}

	t.Logf("E2E flow completed: task=%s run=%s session=%s events=%d", task.ID, run.ID, session.ID, len(events))
}

// TestE2EGeminiRuntimeTaskToComplete validates the full flow with the real
// Gemini API. It is skipped when GEMINI_API_KEY is not available.
func TestE2EGeminiRuntimeTaskToComplete(t *testing.T) {
	if os.Getenv("GEMINI_API_KEY") == "" && os.Getenv("GOOGLE_API_KEY") == "" {
		t.Skip("Skipping Gemini E2E test: GEMINI_API_KEY or GOOGLE_API_KEY not set")
	}

	db := getTestDB(t)
	defer db.Close()

	ctx := context.Background()
	taskService := bootstrap.TaskService(db)
	taskGraphService := bootstrap.TaskGraphService(db)
	runService := bootstrap.RunService(db)
	sessionService := bootstrap.AgentSessionService(db)
	promptService := bootstrap.PromptService(db)
	relay := bootstrap.RuntimeEventRelay(db)

	// 1. Create task
	taskResult, err := taskService.Create(ctx, taskmod.CreateTaskInput{
		Title:       "E2E Gemini Integration Test",
		Description: "Validate Gemini runtime through full orchestration flow",
		Priority:    domain.PriorityP1,
		RiskLevel:   domain.RiskLevelLow,
		AcceptanceCriteria: []string{
			"Agent can process a simple request",
			"Runtime emits completion event",
		},
	})
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	task := taskResult.Value

	// 2. Decompose task into graph
	graphResult, err := taskGraphService.Decompose(ctx, taskgraphmod.DecomposeTaskGraphInput{
		TaskID:          task.ID,
		ReplaceActive:   false,
		CreatedBy:       "e2e_gemini_test",
		PlannerStrategy: "local_heuristic_v1",
	})
	if err != nil {
		t.Fatalf("Failed to decompose task: %v", err)
	}
	if len(graphResult.WorkUnits) == 0 {
		t.Fatal("Expected at least one work unit from decomposition")
	}
	wu := graphResult.WorkUnits[0]

	// 3. Create and start run
	runtimeType := "gemini"
	runResult, err := runService.Create(ctx, runmod.CreateRunInput{
		TaskID:     task.ID,
		WorkUnitID: wu.ID,
		Attempt:    1,
	})
	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}
	run := runResult.Value
	if _, err := runService.Start(ctx, run.ID, orchestration.TransitionInput{Runtime: runtimeType}); err != nil {
		t.Fatalf("Failed to start run: %v", err)
	}

	// 4. Create and connect agent session
	agentID := fmt.Sprintf("agent-gemini-e2e-%s", uuid.New().String()[:8])
	sessionResult, err := sessionService.Create(ctx, agentsessionmod.CreateAgentSessionInput{
		AgentID: agentID,
		RunID:   run.ID,
	})
	if err != nil {
		t.Fatalf("Failed to create agent session: %v", err)
	}
	session := sessionResult.Value
	connectionID := fmt.Sprintf("conn-gemini-e2e-%s", uuid.New().String()[:8])
	if _, err := sessionService.Connect(ctx, session.ID, connectionID, "", orchestration.TransitionInput{Runtime: runtimeType}); err != nil {
		t.Fatalf("Failed to connect agent session: %v", err)
	}

	// 5. Prepare prompt
	preparedPrompt, err := promptService.PrepareRunPrompt(ctx, promptmod.PrepareRunPromptInput{
		RunID:          run.ID,
		AgentSessionID: session.ID,
	})
	if err != nil {
		t.Fatalf("Failed to prepare prompt: %v", err)
	}

	// 6. Start GeminiRuntime with an empty toolset to avoid blocking
	// on tool requests that the relay is not yet configured to auto-approve.
	geminiRuntime := agent.NewGeminiRuntime()
	config := agent.RuntimeConfig{
		RunID:             run.ID,
		WorkUnitID:        wu.ID,
		TaskID:            task.ID,
		AgentID:           agentID,
		Prompt:            preparedPrompt.CombinedPrompt + "\n\nYour task: briefly confirm you received these instructions and that you are ready. Keep your response very short.",
		SystemPrompt:      preparedPrompt.SystemPrompt,
		TaskPrompt:        preparedPrompt.TaskPrompt,
		PromptSnapshotID:  preparedPrompt.PromptSnapshot.ID,
		ToolsetSnapshotID: preparedPrompt.ToolsetSnapshot.ID,
		PromptHash:        preparedPrompt.PromptHash,
		Toolset:           []string{}, // Empty toolset: no function calls
		MaxSteps:          3,
		Timeout:           300,
	}

	rtCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	if err := geminiRuntime.Start(rtCtx, config); err != nil {
		t.Fatalf("Failed to start Gemini runtime: %v", err)
	}

	// 7. Run relay
	relayConfig := services.RelayConfig{
		SessionID:   session.ID,
		RunID:       run.ID,
		RuntimeType: runtimeType,
		AgentID:     agentID,
		OnEvent: func(event *domain.EventEnvelope) {
			t.Logf("[gemini] %s", event.Type)
		},
	}

	finalStatus, err := relay.Run(rtCtx, geminiRuntime, relayConfig)
	if err != nil {
		t.Fatalf("Relay failed: %v", err)
	}
	if finalStatus != domain.RunStatusCompleted {
		t.Fatalf("Expected run status %s, got %s", domain.RunStatusCompleted, finalStatus)
	}

	// 8. Assertions
	runRepo := runmod.NewRepository(db)
	finalRun, err := runRepo.GetByID(run.ID)
	if err != nil {
		t.Fatalf("Failed to get final run: %v", err)
	}
	if finalRun.Status != domain.RunStatusCompleted {
		t.Errorf("Expected run status %s, got %s", domain.RunStatusCompleted, finalRun.Status)
	}

	wuRepo := workunitmod.NewRepository(db)
	finalWU, err := wuRepo.GetByID(wu.ID)
	if err != nil {
		t.Fatalf("Failed to get final work unit: %v", err)
	}
	if finalWU.Status != domain.WorkUnitStatusCompleted {
		t.Errorf("Expected work unit status %s, got %s", domain.WorkUnitStatusCompleted, finalWU.Status)
	}

	sessionRepo := agentsessionmod.NewRepository(db)
	finalSession, err := sessionRepo.GetByID(session.ID)
	if err != nil {
		t.Fatalf("Failed to get final session: %v", err)
	}
	if finalSession.LastCheckpointAt == nil {
		t.Error("Expected last_checkpoint_at to be set")
	}
	if finalSession.LastHeartbeatAt == nil {
		t.Error("Expected last_heartbeat_at to be set")
	}

	eventStore, err := eventstore.NewStore(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	events, err := eventStore.ListByRun(run.ID)
	if err != nil {
		t.Fatalf("Failed to list events: %v", err)
	}
	if len(events) == 0 {
		t.Error("Expected events to be persisted")
	}

	requiredTypes := map[string]bool{
		"run.started":              false,
		"agent.session_starting":   false,
		"agent.heartbeat":          false,
		"agent.checkpoint_reached": false,
		"agent.completed":          false,
		"run.completed":            false,
	}
	for _, e := range events {
		if _, exists := requiredTypes[e.Type]; exists {
			requiredTypes[e.Type] = true
		}
	}
	for eventType, found := range requiredTypes {
		if !found {
			t.Errorf("Expected event type %s was not found", eventType)
		}
	}

	replayState, err := eventStore.ReplayRunState(run.ID)
	if err != nil {
		t.Fatalf("Failed to replay run: %v", err)
	}
	if replayState == nil {
		t.Fatal("Expected replay state to be non-nil")
	}
	if runStatus, ok := replayState.RunStatuses[run.ID]; !ok || runStatus != domain.RunStatusCompleted {
		t.Errorf("Expected replay run status %s, got %v", domain.RunStatusCompleted, runStatus)
	}

	t.Logf("Gemini E2E flow completed: task=%s run=%s session=%s events=%d", task.ID, run.ID, session.ID, len(events))
}
