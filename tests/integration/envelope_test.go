package integration

import (
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
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

	return database
}

func TestEventEnvelopeValidation(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

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
			t.Error("Envelope was not stored")
		}
		if stored.ID == "" || stored.Sequence == 0 || stored.CreatedAt.IsZero() {
			t.Errorf("Expected store to fill id, sequence, and created_at, got %+v", stored)
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
	defer db.Close()

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
			Sequence:    int64(i),
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
	for i, event := range events {
		if event.Sequence != int64(i+1) {
			t.Errorf("Expected sequence %d, got %d", i+1, event.Sequence)
		}
	}
}

func TestEventQueries(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

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

	repo := repository.NewTaskRepository(db)
	task := &domain.Task{
		ID:          id,
		Title:       "Integration Test Task",
		Description: "Created by integration test",
		Status:      domain.TaskStatusCreated,
		Priority:    domain.PriorityP2,
		RiskLevel:   domain.RiskLevelLow,
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

	repo := repository.NewWorkUnitRepository(db)
	wu := &domain.WorkUnit{
		ID:                   uuid.New().String(),
		TaskGraphID:          taskID,
		Title:                "Integration Test Work Unit",
		Objective:            "Validate event persistence",
		AssignedAgentProfile: "default",
		Status:               domain.WorkUnitStatusCreated,
	}
	if err := repo.Create(wu); err != nil {
		t.Fatalf("Failed to create test work unit: %v", err)
	}
	return wu.ID
}

func createTestRun(t *testing.T, db *sql.DB, taskID, workUnitID string) string {
	t.Helper()

	repo := repository.NewRunRepository(db)
	run := &domain.Run{
		ID:         uuid.New().String(),
		TaskID:     taskID,
		WorkUnitID: workUnitID,
		Status:     domain.RunStatusCreated,
		Attempt:    1,
	}
	if err := repo.Create(run); err != nil {
		t.Fatalf("Failed to create test run: %v", err)
	}
	return run.ID
}
