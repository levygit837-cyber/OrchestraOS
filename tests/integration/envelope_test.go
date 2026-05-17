package integration

import (
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/migrations"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
	taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
	_ "github.com/lib/pq"
)

func getTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=orchestraos password=orchestraos dbname=orchestraos sslmode=disable"
	}

	database, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := database.Ping(); err != nil {
		t.Skipf("Database not available: %v (skipping integration test)", err)
	}
	if err := migrations.Run(database); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestEventEnvelopeValidation(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()

	store, err := eventstore.NewStore(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	t.Run("valid envelope should be stored", func(t *testing.T) {
		taskID := createTestTask(t, db)

		envelope := &domain.EventEnvelope{
			Type:        "task.created",
			Version:     "v1",
			TaskID:      taskID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			Payload:     json.RawMessage(`{"key": "value"}`),
		}

		err := store.Append(envelope)
		if err != nil {
			t.Errorf("Failed to append valid envelope: %v", err)
		}

		// Verify it was stored
		stored, err := store.Get(envelope.ID)
		if err != nil {
			t.Errorf("Failed to get stored envelope: %v", err)
		}
		if stored == nil {
			t.Fatal("Envelope was not stored")
		}
		if stored.ID == "" || stored.Sequence == 0 || stored.CreatedAt.IsZero() {
			t.Errorf("Expected store to fill id, sequence, and created_at, got %+v", stored)
		}
		if envelope.ID == "" || envelope.Sequence == 0 || envelope.CreatedAt.IsZero() {
			t.Errorf("Expected Append to complete envelope before validation, got %+v", envelope)
		}
	})

	t.Run("invalid envelope should be rejected", func(t *testing.T) {
		taskID := createTestTask(t, db)

		// Missing required fields like Type and Version
		envelope := &domain.EventEnvelope{
			TaskID:  taskID,
			Payload: json.RawMessage(`{}`),
		}

		err := store.Append(envelope)
		if err == nil {
			t.Error("Expected validation error for invalid envelope, got nil")
		}
	})

	t.Run("runtime event without run should be rejected", func(t *testing.T) {
		taskID := createTestTask(t, db)

		envelope := &domain.EventEnvelope{
			Type:        "agent.started",
			Version:     "v1",
			TaskID:      taskID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			Payload:     json.RawMessage(`{}`),
		}

		if err := store.Append(envelope); err == nil {
			t.Error("Expected validation error for runtime event without run_id, got nil")
		}
	})
}

func TestEventReplay(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()

	store, err := eventstore.NewStore(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	taskID := createTestTask(t, db)

	// Create multiple events for the task
	for i := 1; i <= 3; i++ {
		envelope := &domain.EventEnvelope{
			ID:          uuid.New().String(),
			Type:        "test.event",
			Version:     "v1",
			TaskID:      taskID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			CreatedAt:   time.Now(),
			Payload:     json.RawMessage(`{"index": ` + string(rune('0'+i)) + `}`),
		}

		if err := store.Append(envelope); err != nil {
			t.Fatalf("Failed to append event %d: %v", i, err)
		}
	}

	// Replay events
	events, err := store.Replay(taskID)
	if err != nil {
		t.Fatalf("Failed to replay events: %v", err)
	}

	if len(events) < 3 {
		t.Errorf("Expected at least 3 events, got %d", len(events))
	}

	// Verify sequence order
	for i := 1; i < len(events); i++ {
		if events[i].Sequence <= events[i-1].Sequence {
			t.Errorf("Expected increasing sequence order, got %d then %d", events[i-1].Sequence, events[i].Sequence)
		}
	}
}

func TestEventIdempotencyAndCheckpointLookup(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()

	store, err := eventstore.NewStore(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)
	runID := createTestRun(t, db, taskID, workUnitID)

	event := &domain.EventEnvelope{
		ID:          uuid.New().String(),
		Type:        "agent.checkpoint_reached",
		Version:     "v1",
		TaskID:      taskID,
		RunID:       runID,
		WorkUnitID:  workUnitID,
		AgentID:     "agent-test",
		Priority:    domain.EventPriorityCheckpoint,
		RequiresAck: false,
		Payload: json.RawMessage(`{
			"checkpoint_id": "checkpoint-1",
			"current_goal": "validate idempotency",
			"ledger": {"pending_todos": []},
			"minimal_summary": "checkpoint ready"
		}`),
	}

	if err := store.Append(event); err != nil {
		t.Fatalf("Failed to append checkpoint event: %v", err)
	}
	if err := store.Append(event); err != nil {
		t.Fatalf("Expected duplicate event append to be idempotent, got %v", err)
	}

	events, err := store.ListByRun(runID)
	if err != nil {
		t.Fatalf("Failed to list events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected one persisted event after duplicate append, got %d", len(events))
	}

	checkpoint, err := store.LastCheckpointByRun(runID)
	if err != nil {
		t.Fatalf("Failed to get latest checkpoint: %v", err)
	}
	if checkpoint == nil || checkpoint.ID != event.ID {
		t.Fatalf("Expected latest checkpoint %s, got %+v", event.ID, checkpoint)
	}
}

func TestEventQueries(t *testing.T) {
	db := getTestDB(t)
	defer func() { _ = db.Close() }()

	store, err := eventstore.NewStore(db)
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}

	taskID := uuid.New().String()
	taskID = createTestTaskWithID(t, db, taskID)
	workUnitID := createTestWorkUnit(t, db, taskID)
	runID := createTestRun(t, db, taskID, workUnitID)

	// Create events with different associations
	events := []struct {
		taskID     string
		runID      string
		workUnitID string
		eventType  string
	}{
		{taskID, runID, workUnitID, "event.with.all.ids"},
		{taskID, runID, "", "event.without.workunit"},
		{taskID, "", "", "event.only.task"},
	}

	for _, e := range events {
		envelope := &domain.EventEnvelope{
			ID:          uuid.New().String(),
			Type:        e.eventType,
			Version:     "v1",
			TaskID:      e.taskID,
			RunID:       e.runID,
			WorkUnitID:  e.workUnitID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			CreatedAt:   time.Now(),
			Payload:     json.RawMessage(`{}`),
		}

		if err := store.Append(envelope); err != nil {
			t.Fatalf("Failed to append event: %v", err)
		}
	}

	t.Run("query by task ID", func(t *testing.T) {
		events, err := store.ListByTask(taskID)
		if err != nil {
			t.Errorf("Failed to list events by task: %v", err)
		}
		if len(events) < 3 {
			t.Errorf("Expected at least 3 events for task, got %d", len(events))
		}
	})

	t.Run("query by run ID", func(t *testing.T) {
		events, err := store.ListByRun(runID)
		if err != nil {
			t.Errorf("Failed to list events by run: %v", err)
		}
		if len(events) < 2 {
			t.Errorf("Expected at least 2 events for run, got %d", len(events))
		}
	})

	t.Run("query by work unit ID", func(t *testing.T) {
		events, err := store.ListByWorkUnit(workUnitID)
		if err != nil {
			t.Errorf("Failed to list events by work unit: %v", err)
		}
		if len(events) < 1 {
			t.Errorf("Expected at least 1 event for work unit, got %d", len(events))
		}
	})
}

func createTestTask(t *testing.T, db *sql.DB) string {
	t.Helper()
	return createTestTaskWithID(t, db, uuid.New().String())
}

func createTestTaskWithID(t *testing.T, db *sql.DB, id string) string {
	t.Helper()

	repo := taskmod.NewRepository(db)
	task := &taskmod.Task{
		ID:          id,
		Title:       "Integration Test Task",
		Description: "Created by integration test",
		Status:      taskmod.StatusCreated,
		Priority:    taskmod.PriorityP2,
		RiskLevel:   taskmod.RiskLevelLow,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repo.Create(task); err != nil {
		t.Fatalf("Failed to create test task: %v", err)
	}
	return task.ID
}

func createTestWorkUnit(t *testing.T, db *sql.DB, taskID string) string {
	t.Helper()

	taskGraphID := createTestTaskGraph(t, db, taskID)
	repo := workunitmod.NewRepository(db)
	wu := &workunitmod.WorkUnit{
		ID:                   uuid.New().String(),
		TaskID:               taskID,
		TaskGraphID:          taskGraphID,
		Title:                "Integration Test Work Unit",
		Objective:            "Validate event persistence",
		AssignedAgentProfile: "default",
		Status:               workunitmod.StatusCreated,
	}
	if err := repo.Create(wu); err != nil {
		t.Fatalf("Failed to create test work unit: %v", err)
	}
	return wu.ID
}

func createTestTaskGraph(t *testing.T, db *sql.DB, taskID string) string {
	t.Helper()

	repo := taskgraphmod.NewRepository(db)
	existing, err := repo.GetActiveByTask(taskID)
	if err != nil {
		t.Fatalf("Failed to get active task graph: %v", err)
	}
	if existing != nil {
		return existing.ID
	}
	graph := &domain.TaskGraph{
		ID:              uuid.New().String(),
		TaskID:          taskID,
		Version:         1,
		Status:          domain.TaskGraphStatusActive,
		PlannerStrategy: "integration_test",
		Rationale:       "Integration test graph",
		CreatedBy:       "integration_test",
		NodeCount:       0,
		EdgeCount:       0,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := repo.Create(graph); err != nil {
		t.Fatalf("Failed to create test task graph: %v", err)
	}
	return graph.ID
}

func createTestRun(t *testing.T, db *sql.DB, taskID, workUnitID string) string {
	t.Helper()

	repo := runmod.NewRepository(db)
	run := &runmod.Run{
		ID:         uuid.New().String(),
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		Status:     runmod.StatusCreated,
		Attempt:    1,
	}
	if err := repo.Create(run); err != nil {
		t.Fatalf("Failed to create test run: %v", err)
	}
	return run.ID
}
