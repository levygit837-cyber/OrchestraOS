package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
	_ "github.com/lib/pq"
)

// TestFakeRuntimeEvents tests that the fake runtime emits the expected event types
func TestFakeRuntimeEvents(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	eventStore, err := eventstore.NewStore(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	t.Run("fake runtime emits expected events", func(t *testing.T) {
		taskID := createTestTask(t, db)
		workUnitID := createTestWorkUnit(t, db, taskID)
		runID := createTestRun(t, db, taskID, workUnitID)
		agentID := "agent-test-001"

		fakeRuntime := agent.NewFakeRuntime()
		config := agent.RuntimeConfig{
			RunID:             runID,
			WorkUnitID:        workUnitID,
			TaskID:            taskID,
			AgentID:           agentID,
			Prompt:            "Test work unit",
			PromptSnapshotID:  uuid.New().String(),
			ToolsetSnapshotID: uuid.New().String(),
			PromptHash:        "sha256:test",
			Toolset:           []string{"runtime.fake.emit"},
			MaxSteps:          10,
			Timeout:           300,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Start runtime
		if err := fakeRuntime.Start(ctx, config); err != nil {
			t.Fatalf("Failed to start fake runtime: %v", err)
		}

		// Collect events
		var events []*domain.EventEnvelope
		eventCount := 0
		timeout := time.After(8 * time.Second)

	collectLoop:
		for {
			select {
			case <-timeout:
				break collectLoop
			default:
				event, err := fakeRuntime.ReceiveEvent(ctx)
				if err != nil {
					// Context cancelled or runtime stopped
					break collectLoop
				}
				events = append(events, event)
				eventCount++

				if err := eventStore.Append(event); err != nil {
					t.Fatalf("Failed to append runtime event %q: %v", event.Type, err)
				}

				// Stop after we get the completed event
				if event.Type == "agent.completed" {
					break collectLoop
				}
			}
		}

		// Verify we got expected event types
		expectedTypes := map[string]bool{
			"agent.connected":          false,
			"agent.started":            false,
			"agent.heartbeat":          false,
			"agent.checkpoint_reached": false,
			"agent.tool_requested":     false,
			"agent.completed":          false,
		}

		for _, event := range events {
			if _, exists := expectedTypes[event.Type]; exists {
				expectedTypes[event.Type] = true
			}
			if event.Type == "agent.started" {
				var payload map[string]interface{}
				if err := json.Unmarshal(event.Payload, &payload); err != nil {
					t.Fatalf("Failed to decode agent.started payload: %v", err)
				}
				if _, exists := payload["prompt"]; exists {
					t.Fatalf("agent.started should reference prompt snapshot instead of embedding prompt body")
				}
				if payload["prompt_snapshot_id"] != config.PromptSnapshotID || payload["toolset_snapshot_id"] != config.ToolsetSnapshotID {
					t.Fatalf("agent.started should reference prompt/toolset snapshots, got %+v", payload)
				}
			}

			// Verify event structure
			if event.TaskID != taskID {
				t.Errorf("Expected task ID %s, got %s", taskID, event.TaskID)
			}
			if event.RunID != runID {
				t.Errorf("Expected run ID %s, got %s", runID, event.RunID)
			}
			if event.AgentID != agentID {
				t.Errorf("Expected agent ID %s, got %s", agentID, event.AgentID)
			}

			// Validate event envelope
			if event.Type == "" {
				t.Error("Event type should not be empty")
			}
			if event.Version == "" {
				t.Error("Event version should not be empty")
			}
		}

		// Check that all expected types were received
		for eventType, received := range expectedTypes {
			if !received {
				t.Errorf("Expected event type %s was not received", eventType)
			}
		}

		t.Logf("Received %d events from fake runtime", len(events))
	})
}

// TestFakeRuntimeWithAgentSession tests the full integration of FakeRuntime with AgentSession
func TestFakeRuntimeWithAgentSession(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	taskRepo := taskmod.NewRepository(db)
	wuRepo := workunitmod.NewRepository(db)
	runRepo := runmod.NewRepository(db)
	sessionRepo := agentsessionmod.NewRepository(db)
	eventStore, _ := eventstore.NewStore(db)

	t.Run("full integration flow", func(t *testing.T) {
		// 1. Create task
		task := &taskmod.Task{
			ID:        uuid.New().String(),
			Title:     "Integration Test Task",
			Status:    taskmod.StatusCreated,
			Priority:  taskmod.PriorityP1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := taskRepo.Create(task); err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		// 2. Create work unit
		taskGraphID := createTestTaskGraph(t, db, task.ID)
		wu := &domain.WorkUnit{
			TaskID:               task.ID,
			TaskGraphID:          taskGraphID,
			Title:                "Integration Work Unit",
			AssignedAgentProfile: "default",
			Status:               domain.WorkUnitStatusCreated,
		}
		if err := wuRepo.Create(wu); err != nil {
			t.Fatalf("Failed to create work unit: %v", err)
		}

		// 3. Create run
		run := &domain.Run{
			ID:         uuid.New().String(),
			TaskID:     task.ID,
			WorkUnitID: wu.ID,
			Status:     domain.RunStatusCreated,
			Attempt:    1,
		}
		if err := runRepo.Create(run); err != nil {
			t.Fatalf("Failed to create run: %v", err)
		}

		// Update run to running
		if err := runRepo.UpdateStatus(run.ID, domain.RunStatusRunning, nil, nil); err != nil {
			t.Fatalf("Failed to update run status: %v", err)
		}

		// 4. Create agent session
		agentID := "agent-integration-001"
		session := &domain.AgentSession{
			ID:      uuid.New().String(),
			AgentID: agentID,
			RunID:   run.ID,
			Status:  domain.AgentSessionStatusStarting,
		}
		if err := sessionRepo.Create(session); err != nil {
			t.Fatalf("Failed to create agent session: %v", err)
		}

		// Update session to running
		if err := sessionRepo.UpdateStatus(session.ID, domain.AgentSessionStatusRunning); err != nil {
			t.Fatalf("Failed to update session status: %v", err)
		}

		// 5. Start fake runtime
		fakeRuntime := agent.NewFakeRuntime()
		config := agent.RuntimeConfig{
			RunID:      run.ID,
			WorkUnitID: wu.ID,
			TaskID:     task.ID,
			AgentID:    agentID,
			Prompt:     wu.Title,
			MaxSteps:   10,
			Timeout:    300,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := fakeRuntime.Start(ctx, config); err != nil {
			t.Fatalf("Failed to start fake runtime: %v", err)
		}

		// 6. Collect and store events
		var eventTypes []string
		timeout := time.After(8 * time.Second)
		sessionService := bootstrap.AgentSessionService(db)

	collectLoop:
		for {
			select {
			case <-timeout:
				break collectLoop
			default:
				event, err := fakeRuntime.ReceiveEvent(ctx)
				if err != nil {
					break collectLoop
				}

				eventTypes = append(eventTypes, event.Type)
				switch event.Type {
				case "agent.heartbeat":
					payload := map[string]interface{}{}
					if err := json.Unmarshal(event.Payload, &payload); err != nil {
						t.Fatalf("Failed to decode heartbeat payload: %v", err)
					}
					if _, err := sessionService.Heartbeat(ctx, session.ID, domain.HeartbeatInput{
						EventID: event.ID,
						Payload: payload,
					}); err != nil {
						t.Fatalf("Failed to persist heartbeat via service: %v", err)
					}
				case "agent.checkpoint_reached":
					if _, err := sessionService.CheckpointFromEvent(ctx, session.ID, event); err != nil {
						t.Fatalf("Failed to persist checkpoint via service: %v", err)
					}
				default:
					if err := eventStore.Append(event); err != nil {
						t.Fatalf("Failed to append runtime event %q: %v", event.Type, err)
					}
				}

				// End after completion
				if event.Type == "agent.completed" {
					break collectLoop
				}
			}
		}

		// 7. Verify final state
		storedSession, err := sessionRepo.GetByID(session.ID)
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}

		if storedSession.LastHeartbeatAt == nil {
			t.Error("Expected heartbeat to be recorded")
		}
		if storedSession.LastCheckpointAt == nil {
			t.Error("Expected checkpoint to be recorded")
		}

		// 8. Verify events were persisted
		events, err := eventStore.ListByRun(run.ID)
		if err != nil {
			t.Fatalf("Failed to list events: %v", err)
		}

		if len(events) == 0 {
			t.Error("Expected events to be persisted for run")
		}

		t.Logf("Integration flow completed with %d events", len(events))
		t.Logf("Event types received: %v", eventTypes)
	})
}

// TestEventPayloads verifies that event payloads are correctly serialized and deserialized
func TestEventPayloads(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	eventStore, err := eventstore.NewStore(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	t.Run("checkpoint event with complex payload", func(t *testing.T) {
		taskID := createTestTask(t, db)
		workUnitID := createTestWorkUnit(t, db, taskID)
		runID := createTestRun(t, db, taskID, workUnitID)

		payload := map[string]interface{}{
			"checkpoint_id":   uuid.New().String(),
			"current_goal":    "Implement feature X",
			"minimal_summary": "Implementation checkpoint captured",
			"ledger": map[string]interface{}{
				"pending_todos": []string{},
				"blockers":      []string{},
			},
			"files_modified": []string{"main.go", "utils.go"},
			"evidence_refs":  []string{"artifact:test-report"},
			"metrics": map[string]int{
				"lines_added":   50,
				"lines_removed": 10,
			},
		}

		payloadBytes, _ := json.Marshal(payload)

		event := &domain.EventEnvelope{
			ID:          uuid.New().String(),
			Type:        "agent.checkpoint_reached",
			Version:     "v1",
			TaskID:      taskID,
			RunID:       runID,
			WorkUnitID:  workUnitID,
			Sequence:    1,
			Priority:    domain.EventPriorityCheckpoint,
			RequiresAck: false,
			CreatedAt:   time.Now(),
			Payload:     payloadBytes,
		}

		if err := eventStore.Append(event); err != nil {
			t.Fatalf("Failed to append event: %v", err)
		}

		// Retrieve and verify payload
		stored, err := eventStore.Get(event.ID)
		if err != nil {
			t.Fatalf("Failed to get event: %v", err)
		}

		var storedPayload map[string]interface{}
		if err := json.Unmarshal(stored.Payload, &storedPayload); err != nil {
			t.Fatalf("Failed to unmarshal payload: %v", err)
		}

		if storedPayload["current_goal"].(string) != "Implement feature X" {
			t.Errorf("Expected current goal, got %v", storedPayload["current_goal"])
		}

		files := storedPayload["files_modified"].([]interface{})
		if len(files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(files))
		}
	})
}
