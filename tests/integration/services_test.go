package integration

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/repository"
	"github.com/levygit837-cyber/OrchestraOS/internal/services"
)

func TestDomainServicesFullLifecycle(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := services.NewTaskService(db)
	workUnitService := services.NewWorkUnitService(db)
	runService := services.NewRunService(db)
	sessionService := services.NewAgentSessionService(db)
	eventService := services.NewEventService(db)

	taskResult, err := taskService.Create(ctx, services.CreateTaskInput{
		Title:              "Service lifecycle",
		Description:        "Validate service orchestration",
		Priority:           domain.PriorityP1,
		RiskLevel:          domain.RiskLevelLow,
		AcceptanceCriteria: []string{"run completes with validation evidence"},
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	taskID := taskResult.Value.ID

	wuResult, err := workUnitService.Create(ctx, services.CreateWorkUnitInput{
		TaskID:               taskID,
		Title:                "Implement service lifecycle",
		Objective:            "Exercise all service boundaries",
		AssignedAgentProfile: "fake",
		OwnedPaths:           []string{"internal/services/lifecycle"},
		AcceptanceCriteria:   []string{"checkpoint persisted"},
		ValidationPlan:       []string{"go test ./..."},
	})
	if err != nil {
		t.Fatalf("create work unit: %v", err)
	}

	runResult, err := runService.Create(ctx, services.CreateRunInput{
		TaskID:     taskID,
		WorkUnitID: wuResult.Value.ID,
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	if _, err := runService.Start(ctx, runResult.Value.ID, services.TransitionInput{Runtime: "fake"}); err != nil {
		t.Fatalf("start run: %v", err)
	}

	sessionResult, err := sessionService.Create(ctx, services.CreateAgentSessionInput{
		AgentID: "agent-service-test",
		RunID:   runResult.Value.ID,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := sessionService.Connect(ctx, sessionResult.Value.ID, "conn-service-test", "sandbox-service-test", services.TransitionInput{Runtime: "fake"}); err != nil {
		t.Fatalf("connect session: %v", err)
	}
	if _, err := sessionService.Heartbeat(ctx, sessionResult.Value.ID, services.HeartbeatInput{
		Payload: map[string]interface{}{"source": "test"},
	}); err != nil {
		t.Fatalf("heartbeat: %v", err)
	}
	if _, err := sessionService.Checkpoint(ctx, sessionResult.Value.ID, services.CheckpointInput{
		CheckpointID:   "checkpoint-service-test",
		CurrentGoal:    "validate lifecycle",
		MinimalSummary: "checkpoint persisted",
		Ledger: map[string]interface{}{
			"pending_todos": []interface{}{},
		},
		EvidenceRefs: []string{"checkpoint:service-test"},
	}); err != nil {
		t.Fatalf("checkpoint: %v", err)
	}

	if _, err := runService.Validate(ctx, runResult.Value.ID, services.TransitionInput{Runtime: "fake"}); err != nil {
		t.Fatalf("validate run: %v", err)
	}
	if _, err := runService.Complete(ctx, runResult.Value.ID, services.TransitionInput{
		Runtime:       "fake",
		EvidenceRefs:  []string{"validation:service-test"},
		Justification: "service lifecycle validated",
	}); err != nil {
		t.Fatalf("complete run: %v", err)
	}

	run, err := repository.NewRunRepository(db).GetByID(runResult.Value.ID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if run.Status != domain.RunStatusCompleted {
		t.Fatalf("expected run completed, got %s", run.Status)
	}
	wu, err := repository.NewWorkUnitRepository(db).GetByID(wuResult.Value.ID)
	if err != nil {
		t.Fatalf("get work unit: %v", err)
	}
	if wu.Status != domain.WorkUnitStatusCompleted {
		t.Fatalf("expected work unit completed, got %s", wu.Status)
	}
	session, err := repository.NewAgentSessionRepository(db).GetByID(sessionResult.Value.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if session.LastHeartbeatAt == nil || session.LastCheckpointAt == nil || session.LastSeenEventID == "" {
		t.Fatalf("expected heartbeat, checkpoint, and last seen event to be persisted, got %+v", session)
	}
	state, err := eventService.ReplayRun(ctx, runResult.Value.ID)
	if err != nil {
		t.Fatalf("replay run: %v", err)
	}
	if state.RunStatuses[runResult.Value.ID] != domain.RunStatusCompleted {
		t.Fatalf("expected replayed run completed, got %s", state.RunStatuses[runResult.Value.ID])
	}
}

func TestDomainServicesRejectUnsafeTransitionsAndCascadeCancel(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := services.NewTaskService(db)
	workUnitService := services.NewWorkUnitService(db)
	runService := services.NewRunService(db)

	taskResult, err := taskService.Create(ctx, services.CreateTaskInput{
		Title:     "Cascade cancel",
		Priority:  domain.PriorityP2,
		RiskLevel: domain.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	wuResult, err := workUnitService.Create(ctx, services.CreateWorkUnitInput{
		TaskID:               taskResult.Value.ID,
		Title:                "Cancelable work unit",
		Objective:            "Be cancelled",
		AssignedAgentProfile: "fake",
	})
	if err != nil {
		t.Fatalf("create work unit: %v", err)
	}
	runResult, err := runService.Create(ctx, services.CreateRunInput{
		TaskID:     taskResult.Value.ID,
		WorkUnitID: wuResult.Value.ID,
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	if _, err := runService.Complete(ctx, runResult.Value.ID, services.TransitionInput{EvidenceRefs: []string{"validation:test"}}); err == nil {
		t.Fatal("expected run completion from created state to be rejected")
	}

	if _, err := taskService.Cancel(ctx, taskResult.Value.ID, services.TransitionInput{Justification: "test cascade cancel"}); err != nil {
		t.Fatalf("cancel task: %v", err)
	}
	run, err := repository.NewRunRepository(db).GetByID(runResult.Value.ID)
	if err != nil {
		t.Fatalf("get run: %v", err)
	}
	if run.Status != domain.RunStatusCancelled {
		t.Fatalf("expected related run cancelled, got %s", run.Status)
	}
	wu, err := repository.NewWorkUnitRepository(db).GetByID(wuResult.Value.ID)
	if err != nil {
		t.Fatalf("get work unit: %v", err)
	}
	if wu.Status != domain.WorkUnitStatusCancelled {
		t.Fatalf("expected related work unit cancelled, got %s", wu.Status)
	}
}

func TestEventServiceIdempotencyAndConflict(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskResult, err := services.NewTaskService(db).Create(ctx, services.CreateTaskInput{
		Title:     "Event service idempotency",
		Priority:  domain.PriorityP2,
		RiskLevel: domain.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	eventID := uuid.New().String()
	payload := json.RawMessage(`{"note":"same"}`)
	eventService := services.NewEventService(db)
	first, err := eventService.Append(ctx, &domain.EventEnvelope{
		ID:          eventID,
		Type:        "task.triaged",
		Version:     "v1",
		TaskID:      taskResult.Value.ID,
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		t.Fatalf("append first event: %v", err)
	}
	second, err := eventService.Append(ctx, &domain.EventEnvelope{
		ID:          eventID,
		Type:        "task.triaged",
		Version:     "v1",
		TaskID:      taskResult.Value.ID,
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		Payload:     payload,
	})
	if err != nil {
		t.Fatalf("append duplicate event: %v", err)
	}
	if !second.Duplicate || second.Event.ID != first.Event.ID {
		t.Fatalf("expected duplicate append to return existing event, got %+v", second)
	}
	if _, err := eventService.Append(ctx, &domain.EventEnvelope{
		ID:          eventID,
		Type:        "task.failed",
		Version:     "v1",
		TaskID:      taskResult.Value.ID,
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		Payload:     json.RawMessage(`{"note":"different"}`),
	}); err == nil {
		t.Fatal("expected conflicting idempotency key to be rejected")
	}
}

func TestEventServiceConcurrentIdempotencyConflict(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskResult, err := services.NewTaskService(db).Create(ctx, services.CreateTaskInput{
		Title:     "Concurrent event idempotency",
		Priority:  domain.PriorityP2,
		RiskLevel: domain.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	eventID := uuid.New().String()
	eventService := services.NewEventService(db)
	start := make(chan struct{})
	errs := make(chan error, 2)
	for _, payload := range []json.RawMessage{json.RawMessage(`{"note":"first"}`), json.RawMessage(`{"note":"second"}`)} {
		payload := payload
		go func() {
			<-start
			_, err := eventService.Append(ctx, &domain.EventEnvelope{
				ID:          eventID,
				Type:        "task.triaged",
				Version:     "v1",
				TaskID:      taskResult.Value.ID,
				Priority:    domain.EventPriorityNotification,
				RequiresAck: false,
				Payload:     payload,
			})
			errs <- err
		}()
	}
	close(start)

	successes := 0
	conflicts := 0
	for i := 0; i < 2; i++ {
		err := <-errs
		if err == nil {
			successes++
			continue
		}
		var appErr *apperrors.Error
		if errors.As(err, &appErr) && appErr.Code == apperrors.CodeConflict {
			conflicts++
			continue
		}
		t.Fatalf("expected conflict or success, got %v", err)
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("expected one successful append and one conflict, got successes=%d conflicts=%d", successes, conflicts)
	}

	events, err := eventService.ListByTask(ctx, taskResult.Value.ID)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	count := 0
	for _, event := range events {
		if event.ID == eventID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one persisted event with id %s, got %d", eventID, count)
	}
}

func TestDomainServicesParallelRunsAndPathConflicts(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := services.NewTaskService(db)
	workUnitService := services.NewWorkUnitService(db)
	runService := services.NewRunService(db)

	taskResult, err := taskService.Create(ctx, services.CreateTaskInput{
		Title:     "Parallel services",
		Priority:  domain.PriorityP2,
		RiskLevel: domain.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	runIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		wuResult, err := workUnitService.Create(ctx, services.CreateWorkUnitInput{
			TaskID:               taskResult.Value.ID,
			Title:                "Parallel work unit",
			Objective:            "Run independently",
			AssignedAgentProfile: "fake",
			OwnedPaths:           []string{string(rune('a' + i))},
		})
		if err != nil {
			t.Fatalf("create work unit %d: %v", i, err)
		}
		runResult, err := runService.Create(ctx, services.CreateRunInput{
			TaskID:     taskResult.Value.ID,
			WorkUnitID: wuResult.Value.ID,
		})
		if err != nil {
			t.Fatalf("create run %d: %v", i, err)
		}
		runIDs[i] = runResult.Value.ID
	}

	var wg sync.WaitGroup
	errs := make(chan error, len(runIDs))
	for _, runID := range runIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			_, err := runService.Start(ctx, id, services.TransitionInput{Runtime: "fake"})
			errs <- err
		}(runID)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("parallel start failed: %v", err)
		}
	}

	conflictA, err := workUnitService.Create(ctx, services.CreateWorkUnitInput{
		TaskID:               taskResult.Value.ID,
		Title:                "Conflict A",
		Objective:            "Own same path",
		AssignedAgentProfile: "fake",
		OwnedPaths:           []string{"shared/path"},
	})
	if err != nil {
		t.Fatalf("create conflict A: %v", err)
	}
	conflictB, err := workUnitService.Create(ctx, services.CreateWorkUnitInput{
		TaskID:               taskResult.Value.ID,
		Title:                "Conflict B",
		Objective:            "Own same path",
		AssignedAgentProfile: "fake",
		OwnedPaths:           []string{"shared/path"},
	})
	if err != nil {
		t.Fatalf("create conflict B: %v", err)
	}
	runA, err := runService.Create(ctx, services.CreateRunInput{TaskID: taskResult.Value.ID, WorkUnitID: conflictA.Value.ID})
	if err != nil {
		t.Fatalf("create run A: %v", err)
	}
	runB, err := runService.Create(ctx, services.CreateRunInput{TaskID: taskResult.Value.ID, WorkUnitID: conflictB.Value.ID})
	if err != nil {
		t.Fatalf("create run B: %v", err)
	}
	if _, err := runService.Start(ctx, runA.Value.ID, services.TransitionInput{Runtime: "fake"}); err != nil {
		t.Fatalf("start run A: %v", err)
	}
	if _, err := runService.Start(ctx, runB.Value.ID, services.TransitionInput{Runtime: "fake"}); err == nil {
		t.Fatal("expected owned path conflict")
	} else if appErr, ok := err.(*apperrors.Error); !ok || appErr.Code != apperrors.CodeConflict {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestDomainServicesConcurrentOwnedPathConflict(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := services.NewTaskService(db)
	workUnitService := services.NewWorkUnitService(db)
	runService := services.NewRunService(db)

	taskResult, err := taskService.Create(ctx, services.CreateTaskInput{
		Title:     "Concurrent path conflict",
		Priority:  domain.PriorityP2,
		RiskLevel: domain.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	runIDs := make([]string, 0, 2)
	for _, title := range []string{"Conflict A", "Conflict B"} {
		wuResult, err := workUnitService.Create(ctx, services.CreateWorkUnitInput{
			TaskID:               taskResult.Value.ID,
			Title:                title,
			Objective:            "Own the same path concurrently",
			AssignedAgentProfile: "fake",
			OwnedPaths:           []string{"shared/concurrent/path"},
		})
		if err != nil {
			t.Fatalf("create work unit %s: %v", title, err)
		}
		runResult, err := runService.Create(ctx, services.CreateRunInput{
			TaskID:     taskResult.Value.ID,
			WorkUnitID: wuResult.Value.ID,
		})
		if err != nil {
			t.Fatalf("create run %s: %v", title, err)
		}
		runIDs = append(runIDs, runResult.Value.ID)
	}

	start := make(chan struct{})
	errs := make(chan error, len(runIDs))
	for _, runID := range runIDs {
		runID := runID
		go func() {
			<-start
			_, err := runService.Start(ctx, runID, services.TransitionInput{Runtime: "fake"})
			errs <- err
		}()
	}
	close(start)

	successes := 0
	conflicts := 0
	for i := 0; i < len(runIDs); i++ {
		err := <-errs
		if err == nil {
			successes++
			continue
		}
		var appErr *apperrors.Error
		if errors.As(err, &appErr) && appErr.Code == apperrors.CodeConflict {
			conflicts++
			continue
		}
		t.Fatalf("expected conflict or success, got %v", err)
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("expected one start and one path conflict, got successes=%d conflicts=%d", successes, conflicts)
	}
}

func TestAgentSessionStartingEventReplays(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskResult, err := services.NewTaskService(db).Create(ctx, services.CreateTaskInput{
		Title:     "Replay starting session",
		Priority:  domain.PriorityP2,
		RiskLevel: domain.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	wuResult, err := services.NewWorkUnitService(db).Create(ctx, services.CreateWorkUnitInput{
		TaskID:               taskResult.Value.ID,
		Title:                "Replay session work unit",
		Objective:            "Create a session only",
		AssignedAgentProfile: "fake",
	})
	if err != nil {
		t.Fatalf("create work unit: %v", err)
	}
	runResult, err := services.NewRunService(db).Create(ctx, services.CreateRunInput{
		TaskID:     taskResult.Value.ID,
		WorkUnitID: wuResult.Value.ID,
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	agentID := "agent-replay-starting"
	if _, err := services.NewAgentSessionService(db).Create(ctx, services.CreateAgentSessionInput{
		AgentID: agentID,
		RunID:   runResult.Value.ID,
	}); err != nil {
		t.Fatalf("create agent session: %v", err)
	}

	state, err := services.NewEventService(db).ReplayRun(ctx, runResult.Value.ID)
	if err != nil {
		t.Fatalf("replay run: %v", err)
	}
	if state.AgentSessionStatus[agentID] != domain.AgentSessionStatusStarting {
		t.Fatalf("expected replayed agent session starting, got %s", state.AgentSessionStatus[agentID])
	}
}

func TestAgentSessionAutomaticCheckpointRecoveryAndOrdering(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskResult, err := services.NewTaskService(db).Create(ctx, services.CreateTaskInput{
		Title:     "Automatic checkpoints",
		Priority:  domain.PriorityP2,
		RiskLevel: domain.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	wuResult, err := services.NewWorkUnitService(db).Create(ctx, services.CreateWorkUnitInput{
		TaskID:               taskResult.Value.ID,
		Title:                "Checkpointed work unit",
		Objective:            "Persist recoverable state",
		AssignedAgentProfile: "fake",
	})
	if err != nil {
		t.Fatalf("create work unit: %v", err)
	}
	runService := services.NewRunService(db)
	runResult, err := runService.Create(ctx, services.CreateRunInput{
		TaskID:     taskResult.Value.ID,
		WorkUnitID: wuResult.Value.ID,
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	if _, err := runService.Start(ctx, runResult.Value.ID, services.TransitionInput{Runtime: "fake"}); err != nil {
		t.Fatalf("start run: %v", err)
	}

	sessionService := services.NewAgentSessionService(db)
	sessionResult, err := sessionService.Create(ctx, services.CreateAgentSessionInput{
		AgentID: "agent-auto-checkpoint",
		RunID:   runResult.Value.ID,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := sessionService.Connect(ctx, sessionResult.Value.ID, "conn-auto-checkpoint", "sandbox-auto-checkpoint", services.TransitionInput{Runtime: "fake"}); err != nil {
		t.Fatalf("connect session: %v", err)
	}

	if result, suggestion, err := sessionService.AutomaticCheckpoint(ctx, sessionResult.Value.ID, services.AutoCheckpointInput{
		Trigger:        services.CheckpointTriggerGoalCompleted,
		CurrentGoal:    "analysis",
		MinimalSummary: "analysis completed",
		PendingTodos:   []string{"implementation"},
		EvidenceRefs:   []string{"checkpoint:analysis"},
		Runtime:        "fake",
	}); err != nil {
		t.Fatalf("automatic checkpoint 1: %v", err)
	} else if result == nil || suggestion == nil || !suggestion.ShouldCheckpoint {
		t.Fatalf("expected first automatic checkpoint to persist, result=%+v suggestion=%+v", result, suggestion)
	}

	second, suggestion, err := sessionService.AutomaticCheckpoint(ctx, sessionResult.Value.ID, services.AutoCheckpointInput{
		Trigger:        services.CheckpointTriggerDiffProduced,
		CurrentGoal:    "implementation",
		MinimalSummary: "diff produced and ready for validation",
		CompletedGoals: []string{"analysis"},
		PendingTodos:   []string{"validation"},
		FilesModified:  []string{"internal/services/checkpoint_policy.go"},
		EvidenceRefs:   []string{"checkpoint:diff"},
		Runtime:        "fake",
	})
	if err != nil {
		t.Fatalf("automatic checkpoint 2: %v", err)
	}
	if second == nil || suggestion == nil || !suggestion.ShouldCheckpoint {
		t.Fatalf("expected second automatic checkpoint to persist, result=%+v suggestion=%+v", second, suggestion)
	}

	if result, suggestion, err := sessionService.AutomaticCheckpoint(ctx, sessionResult.Value.ID, services.AutoCheckpointInput{
		Trigger: services.CheckpointTriggerHeartbeat,
	}); err != nil {
		t.Fatalf("heartbeat checkpoint suggestion: %v", err)
	} else if result != nil || suggestion == nil || suggestion.ShouldCheckpoint {
		t.Fatalf("expected heartbeat trigger to be suggestion-only, result=%+v suggestion=%+v", result, suggestion)
	}

	checkpoints, err := sessionService.ListCheckpoints(ctx, sessionResult.Value.ID)
	if err != nil {
		t.Fatalf("list checkpoints: %v", err)
	}
	if len(checkpoints) != 2 {
		t.Fatalf("expected two persisted checkpoints, got %d", len(checkpoints))
	}
	if checkpoints[0].Event.Sequence >= checkpoints[1].Event.Sequence {
		t.Fatalf("expected checkpoints ordered by event sequence, got %d then %d", checkpoints[0].Event.Sequence, checkpoints[1].Event.Sequence)
	}
	if checkpoints[1].CurrentGoal != "implementation" {
		t.Fatalf("expected latest checkpoint to cover implementation state, got %+v", checkpoints[1])
	}

	recovered, err := sessionService.RecoverableCheckpoint(ctx, sessionResult.Value.ID)
	if err != nil {
		t.Fatalf("recover checkpoint: %v", err)
	}
	if recovered.LastCheckpoint == nil || recovered.LastCheckpoint.Event.ID != second.Event.ID {
		t.Fatalf("expected recoverable state to point at latest checkpoint %s, got %+v", second.Event.ID, recovered.LastCheckpoint)
	}
	if len(recovered.RecoverableState) == 0 {
		t.Fatal("expected session recoverable_state to be persisted")
	}
	var state map[string]interface{}
	if err := json.Unmarshal(recovered.RecoverableState, &state); err != nil {
		t.Fatalf("decode recoverable_state: %v", err)
	}
	if state["last_checkpoint_event_id"] != second.Event.ID {
		t.Fatalf("expected recoverable_state last checkpoint %s, got %+v", second.Event.ID, state)
	}
}

func TestRunRetryRequiresPolicyAndIsIdempotent(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskResult, err := services.NewTaskService(db).Create(ctx, services.CreateTaskInput{
		Title:     "Retry policy",
		Priority:  domain.PriorityP2,
		RiskLevel: domain.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	wuResult, err := services.NewWorkUnitService(db).Create(ctx, services.CreateWorkUnitInput{
		TaskID:               taskResult.Value.ID,
		Title:                "Retryable work unit",
		Objective:            "Retry failed run",
		AssignedAgentProfile: "fake",
	})
	if err != nil {
		t.Fatalf("create work unit: %v", err)
	}
	runService := services.NewRunService(db)
	runResult, err := runService.Create(ctx, services.CreateRunInput{
		TaskID:     taskResult.Value.ID,
		WorkUnitID: wuResult.Value.ID,
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	if _, err := runService.Fail(ctx, runResult.Value.ID, services.TransitionInput{
		FailureReason: "transient runtime failure",
		Justification: "prepare retry test",
	}); err != nil {
		t.Fatalf("fail run: %v", err)
	}
	if _, err := runService.Retry(ctx, runResult.Value.ID, services.TransitionInput{
		Justification: "missing idempotency key",
	}); err == nil {
		t.Fatal("expected retry without event_id to be rejected")
	}

	retryEventID := uuid.New().String()
	retryInput := services.TransitionInput{
		EventID:       retryEventID,
		FailureReason: "transient runtime failure",
		Justification: "retry transient runtime failure",
		Extra: map[string]interface{}{
			"max_attempts":              3,
			"attempt_timeout_seconds":   1,
			"operation_timeout_seconds": 5,
			"initial_backoff_millis":    1,
			"backoff_multiplier":        2,
		},
	}
	retryResult, err := runService.Retry(ctx, runResult.Value.ID, retryInput)
	if err != nil {
		t.Fatalf("retry run: %v", err)
	}
	if retryResult.Value.Attempt != 2 || retryResult.Value.Status != domain.RunStatusCreated {
		t.Fatalf("expected retry attempt 2 in created status, got %+v", retryResult.Value)
	}
	duplicate, err := runService.Retry(ctx, runResult.Value.ID, retryInput)
	if err != nil {
		t.Fatalf("retry duplicate: %v", err)
	}
	if !duplicate.Duplicate || duplicate.Value.ID != retryResult.Value.ID {
		t.Fatalf("expected idempotent duplicate retry, got %+v", duplicate)
	}
	if duplicate.Event == nil || duplicate.Event.ID != retryEventID {
		t.Fatalf("expected duplicate retry event %s, got %+v", retryEventID, duplicate.Event)
	}
}
