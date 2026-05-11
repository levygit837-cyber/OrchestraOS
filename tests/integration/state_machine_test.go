package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/eventstore"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/orchestration"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
)

func TestCommanderRunStateMachine(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	taskID := createTestTask(t, db)
	workUnitID := createTestWorkUnit(t, db, taskID)
	runID := createTestRun(t, db, taskID, workUnitID)

	commander := orchestration.NewCommander(db)
	runRepo := repository.NewRunRepository(db)

	t.Run("invalid completed transition does not update projection", func(t *testing.T) {
		err := commander.TransitionRun(context.Background(), runID, domain.RunStatusCompleted, orchestration.TransitionOptions{
			EvidenceRefs: []string{"validation.completed:test"},
		})
		if err == nil {
			t.Fatal("expected invalid transition error, got nil")
		}

		run, err := runRepo.GetByID(runID)
		if err != nil {
			t.Fatalf("failed to get run: %v", err)
		}
		if run.Status != domain.RunStatusCreated {
			t.Fatalf("expected run to remain created, got %s", run.Status)
		}
	})

	t.Run("valid lifecycle persists events and projection", func(t *testing.T) {
		if err := commander.TransitionRun(context.Background(), runID, domain.RunStatusRunning, orchestration.TransitionOptions{}); err != nil {
			t.Fatalf("failed to start run: %v", err)
		}

		validationID := uuid.New().String()
		store, err := eventstore.NewStore(db)
		if err != nil {
			t.Fatalf("failed to create event store: %v", err)
		}
		if _, err := store.AppendRaw("validation.completed", "v1", domain.ValidationCompletedPayload{
			ValidationID: validationID,
			Status:       "passed",
		}, taskID, runID); err != nil {
			t.Fatalf("failed to append validation event: %v", err)
		}

		if err := commander.TransitionRun(context.Background(), runID, domain.RunStatusValidating, orchestration.TransitionOptions{}); err != nil {
			t.Fatalf("failed to validate run: %v", err)
		}
		result := domain.RunResultSucceeded
		if err := commander.TransitionRun(context.Background(), runID, domain.RunStatusCompleted, orchestration.TransitionOptions{
			Result:            &result,
			ValidationEventID: validationID,
		}); err != nil {
			t.Fatalf("failed to complete run: %v", err)
		}

		run, err := runRepo.GetByID(runID)
		if err != nil {
			t.Fatalf("failed to get run: %v", err)
		}
		if run.Status != domain.RunStatusCompleted {
			t.Fatalf("expected run completed, got %s", run.Status)
		}

		state, err := store.ReplayRunState(runID)
		if err != nil {
			t.Fatalf("failed to replay run state: %v", err)
		}
		if state.RunStatuses[runID] != domain.RunStatusCompleted {
			t.Fatalf("expected replay status completed, got %s", state.RunStatuses[runID])
		}
	})
}
