package integration

import (
	"context"
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
)

// TestRunServiceStateMachine validates that RunService enforces state-machine
// rules and persists events correctly through the run lifecycle.
func TestRunServiceStateMachine(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)
	runID := createTestRun(t, db, taskID, workUnitID)

	runService := bootstrap.RunService(db)
	runRepo := runmod.NewRepository(db)

	t.Run("invalid completed transition does not update projection", func(t *testing.T) {
		_, err := runService.Complete(context.Background(), runID, transition.TransitionInput{
			Runtime:       "fake",
			AgentID:       "test-agent",
			EvidenceRefs:  []string{"validation.completed:test"},
			Justification: "attempting invalid transition from created to completed",
		})
		if err == nil {
			t.Fatal("expected invalid transition error, got nil")
		}

		run, err := runRepo.GetByID(runID)
		if err != nil {
			t.Fatalf("failed to get run: %v", err)
		}
		if run.Status != runmod.StatusCreated {
			t.Fatalf("expected run to remain created, got %s", run.Status)
		}
	})

	t.Run("valid lifecycle persists events and projection", func(t *testing.T) {
		if _, err := runService.Start(context.Background(), runID, transition.TransitionInput{
			Runtime: "fake",
			AgentID: "test-agent",
		}); err != nil {
			t.Fatalf("failed to start run: %v", err)
		}

		store, err := eventstore.NewStore(db)
		if err != nil {
			t.Fatalf("failed to create event store: %v", err)
		}

		if _, err := runService.Validate(context.Background(), runID, transition.TransitionInput{
			Runtime:       "fake",
			AgentID:       "test-agent",
			Justification: "validation passed",
		}); err != nil {
			t.Fatalf("failed to validate run: %v", err)
		}

		if _, err := runService.Complete(context.Background(), runID, transition.TransitionInput{
			Runtime:       "fake",
			AgentID:       "test-agent",
			EvidenceRefs:  []string{"validation.completed:test"},
			Justification: "run completed successfully",
		}); err != nil {
			t.Fatalf("failed to complete run: %v", err)
		}

		run, err := runRepo.GetByID(runID)
		if err != nil {
			t.Fatalf("failed to get run: %v", err)
		}
		if run.Status != runmod.StatusCompleted {
			t.Fatalf("expected run completed, got %s", run.Status)
		}

		state, err := store.ReplayRunState(runID)
		if err != nil {
			t.Fatalf("failed to replay run state: %v", err)
		}
		if string(state.RunStatuses[runID]) != string(runmod.StatusCompleted) {
			t.Fatalf("expected replay status completed, got %s", state.RunStatuses[runID])
		}
	})
}
