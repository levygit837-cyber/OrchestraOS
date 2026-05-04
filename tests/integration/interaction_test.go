package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
	_ "github.com/lib/pq"
)

// TestTaskWorkUnitRunInteraction tests the full lifecycle of Task -> WorkUnit -> Run
func TestTaskWorkUnitRunInteraction(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	taskRepo := repository.NewTaskRepository(db)
	wuRepo := repository.NewWorkUnitRepository(db)
	runRepo := repository.NewRunRepository(db)
	eventStore, _ := eventstore.NewStore(db)

	t.Run("create task generates event", func(t *testing.T) {
		task := &domain.Task{
			ID:          uuid.New().String(),
			Title:       "Test Task",
			Description: "Test Description",
			Status:      domain.TaskStatusCreated,
			Priority:    domain.PriorityP1,
			RiskLevel:   domain.RiskLevelLow,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := taskRepo.Create(task); err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		// Create event
		payload := map[string]string{"task_id": task.ID, "title": task.Title}
		payloadBytes, _ := json.Marshal(payload)
		event := &domain.EventEnvelope{
			Type:        "task.created",
			Version:     "v1",
			TaskID:      task.ID,
			Priority:    domain.EventPriorityNotification,
			RequiresAck: false,
			CreatedAt:   time.Now(),
			Payload:     payloadBytes,
		}

		if err := eventStore.Append(event); err != nil {
			t.Fatalf("Failed to append event: %v", err)
		}

		// Verify event was stored
		events, err := eventStore.ListByTask(task.ID)
		if err != nil {
			t.Fatalf("Failed to list events: %v", err)
		}
		if len(events) == 0 {
			t.Error("Expected events for task, got none")
		}
	})

	t.Run("task with multiple work units", func(t *testing.T) {
		// Create task
		task := &domain.Task{
			ID:          uuid.New().String(),
			Title:       "Task with WorkUnits",
			Description: "Testing work units",
			Status:      domain.TaskStatusCreated,
			Priority:    domain.PriorityP2,
			RiskLevel:   domain.RiskLevelMedium,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := taskRepo.Create(task); err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		// Create multiple work units
		workUnits := []domain.WorkUnit{
			{
				TaskGraphID:          task.ID,
				Title:                "Work Unit 1",
				Objective:            "First objective",
				AssignedAgentProfile: "default",
				Status:               domain.WorkUnitStatusCreated,
			},
			{
				TaskGraphID:          task.ID,
				Title:                "Work Unit 2",
				Objective:            "Second objective",
				AssignedAgentProfile: "default",
				Status:               domain.WorkUnitStatusCreated,
			},
		}

		for i := range workUnits {
			if err := wuRepo.Create(&workUnits[i]); err != nil {
				t.Fatalf("Failed to create work unit %d: %v", i+1, err)
			}

			// Create work_unit.created event
			payload := map[string]string{
				"work_unit_id": workUnits[i].ID,
				"task_id":      task.ID,
				"title":        workUnits[i].Title,
			}
			payloadBytes, _ := json.Marshal(payload)
			event := &domain.EventEnvelope{
				Type:        "work_unit.created",
				Version:     "v1",
				TaskID:      task.ID,
				WorkUnitID:  workUnits[i].ID,
				Priority:    domain.EventPriorityNotification,
				RequiresAck: false,
				CreatedAt:   time.Now(),
				Payload:     payloadBytes,
			}
			if err := eventStore.Append(event); err != nil {
				t.Fatalf("Failed to append work_unit.created event: %v", err)
			}
		}

		// List work units and verify
		storedWUs, err := wuRepo.ListByTask(task.ID)
		if err != nil {
			t.Fatalf("Failed to list work units: %v", err)
		}
		if len(storedWUs) != 2 {
			t.Errorf("Expected 2 work units, got %d", len(storedWUs))
		}
	})

	t.Run("work unit with runs", func(t *testing.T) {
		// Create task and work unit
		task := &domain.Task{
			ID:        uuid.New().String(),
			Title:     "Task for Run",
			Status:    domain.TaskStatusCreated,
			Priority:  domain.PriorityP1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := taskRepo.Create(task); err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		wu := &domain.WorkUnit{
			TaskGraphID:          task.ID,
			Title:                "Work Unit with Run",
			AssignedAgentProfile: "default",
			Status:               domain.WorkUnitStatusCreated,
		}
		if err := wuRepo.Create(wu); err != nil {
			t.Fatalf("Failed to create work unit: %v", err)
		}

		// Create run
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

		// Create run.started event
		payload := map[string]string{
			"run_id":       run.ID,
			"work_unit_id": wu.ID,
			"status":       "running",
		}
		payloadBytes, _ := json.Marshal(payload)
		event := &domain.EventEnvelope{
			Type:        "run.started",
			Version:     "v1",
			TaskID:      task.ID,
			RunID:       run.ID,
			WorkUnitID:  wu.ID,
			Priority:    domain.EventPriorityCheckpoint,
			RequiresAck: false,
			CreatedAt:   time.Now(),
			Payload:     payloadBytes,
		}
		if err := eventStore.Append(event); err != nil {
			t.Fatalf("Failed to append event: %v", err)
		}

		// Complete run
		result := domain.RunResultSucceeded
		if err := runRepo.UpdateStatus(run.ID, domain.RunStatusCompleted, &result, nil); err != nil {
			t.Fatalf("Failed to complete run: %v", err)
		}

		// Verify run exists
		storedRun, err := runRepo.GetByID(run.ID)
		if err != nil {
			t.Fatalf("Failed to get run: %v", err)
		}
		if storedRun == nil {
			t.Error("Run was not stored")
		}
		if storedRun.Status != domain.RunStatusCompleted {
			t.Errorf("Expected status completed, got %s", storedRun.Status)
		}
		if storedRun.StartedAt.IsZero() {
			t.Error("Expected started_at to be preserved after completion")
		}
		if storedRun.FinishedAt == nil {
			t.Error("Expected finished_at to be set after completion")
		}
		if storedRun.Result == nil || *storedRun.Result != domain.RunResultSucceeded {
			t.Error("Expected result to be succeeded")
		}
	})
}

// TestAgentSessionWithRun tests AgentSession lifecycle with Run
func TestAgentSessionWithRun(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	taskRepo := repository.NewTaskRepository(db)
	wuRepo := repository.NewWorkUnitRepository(db)
	runRepo := repository.NewRunRepository(db)
	sessionRepo := repository.NewAgentSessionRepository(db)

	t.Run("agent session lifecycle", func(t *testing.T) {
		// Create task
		task := &domain.Task{
			ID:        uuid.New().String(),
			Title:     "Task with Agent Session",
			Status:    domain.TaskStatusCreated,
			Priority:  domain.PriorityP1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := taskRepo.Create(task); err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		// Create work unit
		wu := &domain.WorkUnit{
			TaskGraphID:          task.ID,
			Title:                "Work Unit for Agent",
			AssignedAgentProfile: "default",
			Status:               domain.WorkUnitStatusCreated,
		}
		if err := wuRepo.Create(wu); err != nil {
			t.Fatalf("Failed to create work unit: %v", err)
		}

		// Create run
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

		// Create agent session
		session := &domain.AgentSession{
			ID:      uuid.New().String(),
			AgentID: "agent-test-001",
			RunID:   run.ID,
			Status:  domain.AgentSessionStatusStarting,
		}
		if err := sessionRepo.Create(session); err != nil {
			t.Fatalf("Failed to create agent session: %v", err)
		}

		// Transition to running
		if err := sessionRepo.UpdateStatus(session.ID, domain.AgentSessionStatusRunning); err != nil {
			t.Fatalf("Failed to update session to running: %v", err)
		}

		// Update heartbeat
		if err := sessionRepo.UpdateHeartbeat(session.ID); err != nil {
			t.Fatalf("Failed to update heartbeat: %v", err)
		}

		// Update checkpoint
		if err := sessionRepo.UpdateCheckpoint(session.ID); err != nil {
			t.Fatalf("Failed to update checkpoint: %v", err)
		}

		// Transition to stopped
		if err := sessionRepo.UpdateStatus(session.ID, domain.AgentSessionStatusStopped); err != nil {
			t.Fatalf("Failed to update session to stopped: %v", err)
		}

		// Verify final state
		storedSession, err := sessionRepo.GetByID(session.ID)
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}
		if storedSession == nil {
			t.Fatal("Session was not stored")
		}
		if storedSession.Status != domain.AgentSessionStatusStopped {
			t.Errorf("Expected status stopped, got %s", storedSession.Status)
		}
		if storedSession.LastHeartbeatAt == nil {
			t.Error("Expected heartbeat to be set")
		}
		if storedSession.LastCheckpointAt == nil {
			t.Error("Expected checkpoint to be set")
		}
	})
}
