package integration

import (
	"context"
	"testing"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/bootstrap"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/transition"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
	triggermod "github.com/levygit837-cyber/OrchestraOS/internal/modules/trigger"
	runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
	workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
	taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
)

func TestTriggerServiceCreateAndList(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	triggerService := bootstrap.TriggerService(db)

	t.Run("create trigger persists and emits event", func(t *testing.T) {
		result, err := triggerService.Create(ctx, triggermod.CreateTriggerInput{
			TriggerType: domain.TriggerTypeThreshold,
			Status:      domain.TriggerStatusActive,
		})
		if err != nil {
			t.Fatalf("create trigger: %v", err)
		}
		if result.Value.ID == "" {
			t.Fatal("expected trigger ID to be generated")
		}
		if result.Value.Status != domain.TriggerStatusActive {
			t.Fatalf("expected status active, got %s", result.Value.Status)
		}
		if result.Event == nil || result.Event.Type != "trigger.created" {
			t.Fatalf("expected trigger.created event, got %+v", result.Event)
		}
	})

	t.Run("list active returns non-resolved triggers", func(t *testing.T) {
		active, err := triggerService.ListActive(ctx)
		if err != nil {
			t.Fatalf("list active: %v", err)
		}
		if len(active) == 0 {
			t.Fatal("expected at least one active trigger")
		}
		for _, tr := range active {
			if tr.Status == domain.TriggerStatusResolved || tr.Status == domain.TriggerStatusDismissed {
				t.Fatalf("expected no resolved/dismissed triggers in ListActive, got %s", tr.Status)
			}
		}
	})
}

func TestTriggerServiceResolveAndDismiss(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	triggerService := bootstrap.TriggerService(db)

	result, err := triggerService.Create(ctx, triggermod.CreateTriggerInput{
		TriggerType: domain.TriggerTypeAnomaly,
		Status:      domain.TriggerStatusActive,
	})
	if err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	triggerID := result.Value.ID

	t.Run("resolve trigger", func(t *testing.T) {
		resolved, err := triggerService.Resolve(ctx, triggerID, domain.ResolutionActionNotify, "test resolution")
		if err != nil {
			t.Fatalf("resolve trigger: %v", err)
		}
		if resolved.Value.Status != domain.TriggerStatusResolved {
			t.Fatalf("expected resolved, got %s", resolved.Value.Status)
		}
		if resolved.Event == nil || resolved.Event.Type != "trigger.resolved" {
			t.Fatalf("expected trigger.resolved event, got %+v", resolved.Event)
		}
	})

	t.Run("cannot resolve already resolved trigger", func(t *testing.T) {
		_, err := triggerService.Resolve(ctx, triggerID, domain.ResolutionActionNotify, "double resolution")
		if err == nil {
			t.Fatal("expected error resolving already resolved trigger")
		}
	})

	t.Run("dismiss trigger", func(t *testing.T) {
		created, err := triggerService.Create(ctx, triggermod.CreateTriggerInput{
			TriggerType: domain.TriggerTypePolicy,
			Status:      domain.TriggerStatusActive,
		})
		if err != nil {
			t.Fatalf("create trigger for dismiss: %v", err)
		}
		dismissed, err := triggerService.Dismiss(ctx, created.Value.ID, "test dismiss")
		if err != nil {
			t.Fatalf("dismiss trigger: %v", err)
		}
		if dismissed.Value.Status != domain.TriggerStatusDismissed {
			t.Fatalf("expected dismissed, got %s", dismissed.Value.Status)
		}
	})
}

func TestTriggerServiceEvaluateRun(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := bootstrap.TaskService(db)
	workUnitService := bootstrap.WorkUnitService(db)
	runService := bootstrap.RunService(db)
	triggerService := bootstrap.TriggerService(db)

	// Override clock and thresholds for determinism
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	triggerService.SetClock(func() time.Time { return now })
	triggerService.SetThresholds(domain.ThresholdConfig{
		StallSeconds:    60,
		LoopRepetitions: 3,
		TokenMax:        100,
		StepsMax:        5,
		TimeMaxSeconds:  300,
	})

	taskResult, err := taskService.Create(ctx, taskmod.CreateTaskInput{
		Title:     "Trigger evaluate run",
		Priority:  taskmod.PriorityP2,
		RiskLevel: taskmod.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	wuResult, err := workUnitService.Create(ctx, workunitmod.CreateWorkUnitInput{
		TaskID:               taskResult.Value.ID,
		Title:                "Evaluate run work unit",
		Objective:            "Trigger detection",
		AssignedAgentProfile: "fake",
	})
	if err != nil {
		t.Fatalf("create work unit: %v", err)
	}
	runResult, err := runService.Create(ctx, runmod.CreateRunInput{
		TaskID:     taskResult.Value.ID,
		WorkUnitID: wuResult.Value.ID,
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}

	// Start the run so it has a started_at
	if _, err := runService.Start(ctx, runResult.Value.ID, transition.TransitionInput{Runtime: "fake"}); err != nil {
		t.Fatalf("start run: %v", err)
	}

	// Move clock forward to trigger time threshold
	now = now.Add(time.Hour)
	triggerService.SetClock(func() time.Time { return now })

	detected, err := triggerService.EvaluateRun(ctx, runResult.Value.ID)
	if err != nil {
		t.Fatalf("evaluate run: %v", err)
	}

	// Should detect time exceeded at minimum
	foundTime := false
	for _, tr := range detected {
		if tr.AnomalyType != nil && *tr.AnomalyType == domain.AnomalyTypeTimeExceeded {
			foundTime = true
		}
	}
	if !foundTime {
		t.Fatalf("expected time exceeded trigger, got %+v", detected)
	}

	// Verify triggers are persisted
	byRun, err := triggerService.ListByRun(ctx, runResult.Value.ID)
	if err != nil {
		t.Fatalf("list by run: %v", err)
	}
	if len(byRun) == 0 {
		t.Fatal("expected triggers persisted for run")
	}
}

func TestTriggerServiceEvaluateSession(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := bootstrap.TaskService(db)
	workUnitService := bootstrap.WorkUnitService(db)
	runService := bootstrap.RunService(db)
	sessionService := bootstrap.AgentSessionService(db)
	triggerService := bootstrap.TriggerService(db)

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	triggerService.SetClock(func() time.Time { return now })
	triggerService.SetThresholds(domain.ThresholdConfig{
		StallSeconds: 60,
	})

	taskResult, err := taskService.Create(ctx, taskmod.CreateTaskInput{
		Title:     "Trigger evaluate session",
		Priority:  taskmod.PriorityP2,
		RiskLevel: taskmod.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	wuResult, err := workUnitService.Create(ctx, workunitmod.CreateWorkUnitInput{
		TaskID:               taskResult.Value.ID,
		Title:                "Session work unit",
		Objective:            "Trigger detection",
		AssignedAgentProfile: "fake",
	})
	if err != nil {
		t.Fatalf("create work unit: %v", err)
	}
	runResult, err := runService.Create(ctx, runmod.CreateRunInput{
		TaskID:     taskResult.Value.ID,
		WorkUnitID: wuResult.Value.ID,
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}

	sessionResult, err := sessionService.Create(ctx, agentsessionmod.CreateAgentSessionInput{
		AgentID:    "agent-trigger-test",
		RunID:      runResult.Value.ID,
		TaskID:     runResult.Value.TaskID,
		WorkUnitID: runResult.Value.WorkUnitID,
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	detected, err := triggerService.EvaluateSession(ctx, sessionResult.Value.ID)
	if err != nil {
		t.Fatalf("evaluate session: %v", err)
	}

	// Should detect heartbeat timeout because no heartbeat was sent
	foundTimeout := false
	for _, tr := range detected {
		if tr.TriggerType == domain.TriggerTypeHeartbeatTimeout {
			foundTimeout = true
		}
	}
	if !foundTimeout {
		t.Fatalf("expected heartbeat timeout trigger, got %+v", detected)
	}
}

func TestTriggerServiceEvaluateWorkUnit(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	ctx := context.Background()

	taskService := bootstrap.TaskService(db)
	workUnitService := bootstrap.WorkUnitService(db)
	triggerService := bootstrap.TriggerService(db)

	taskResult, err := taskService.Create(ctx, taskmod.CreateTaskInput{
		Title:     "Trigger evaluate work unit",
		Priority:  taskmod.PriorityP2,
		RiskLevel: taskmod.RiskLevelLow,
	})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	wuResult, err := workUnitService.Create(ctx, workunitmod.CreateWorkUnitInput{
		TaskID:               taskResult.Value.ID,
		Title:                "Owned paths work unit",
		Objective:            "Path violation detection",
		AssignedAgentProfile: "fake",
		OwnedPaths:           []string{"/app/src"},
		ReadPaths:            []string{"/app/docs"},
	})
	if err != nil {
		t.Fatalf("create work unit: %v", err)
	}

	detected, err := triggerService.EvaluateWorkUnit(ctx, wuResult.Value.ID)
	if err != nil {
		t.Fatalf("evaluate work unit: %v", err)
	}

	// No events exist, so no drift/violation should be detected
	if len(detected) != 0 {
		t.Fatalf("expected no triggers without events, got %+v", detected)
	}
}
