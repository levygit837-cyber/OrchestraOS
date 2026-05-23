package trigger_test

import (
	"testing"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/modules/trigger"
)

func TestStallDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	startedAt := now.Add(-time.Minute * 10)

	t.Run("detects stall when no events and exceeded threshold", func(t *testing.T) {
		d := trigger.StallDetector{}
		tr := d.Detect(nil, startedAt, now, 300)
		if tr == nil {
			t.Fatal("expected stall detection")
		}
		if tr.AnomalyType == nil || *tr.AnomalyType != string(trigger.AnomalyStall) {
			t.Fatalf("expected anomaly stall, got %v", tr.AnomalyType)
		}
	})

	t.Run("no stall when within threshold", func(t *testing.T) {
		d := trigger.StallDetector{}
		lastEvent := now.Add(-time.Minute * 2)
		tr := d.Detect(&lastEvent, startedAt, now, 300)
		if tr != nil {
			t.Fatal("expected no stall")
		}
	})
}

func TestLoopDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects loop with repeated events", func(t *testing.T) {
		d := trigger.LoopDetector{}
		eventTypes := []string{"run.started", "run.started", "run.started", "run.started", "run.started"}
		tr := d.Detect(eventTypes, 3, now)
		if tr == nil {
			t.Fatal("expected loop detection")
		}
		if tr.AnomalyType == nil || *tr.AnomalyType != string(trigger.AnomalyLoop) {
			t.Fatalf("expected anomaly loop, got %v", tr.AnomalyType)
		}
	})

	t.Run("no loop below threshold", func(t *testing.T) {
		d := trigger.LoopDetector{}
		eventTypes := []string{"run.started", "tool.requested", "run.started"}
		tr := d.Detect(eventTypes, 5, now)
		if tr != nil {
			t.Fatal("expected no loop")
		}
	})

	t.Run("ignores trigger events to prevent feedback loop", func(t *testing.T) {
		d := trigger.LoopDetector{}
		eventTypes := []string{"trigger.created", "trigger.created", "trigger.created", "trigger.created"}
		tr := d.Detect(eventTypes, 3, now)
		if tr != nil {
			t.Fatal("expected no loop for trigger events")
		}
	})
}

func TestDriftDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects drift outside owned and read paths", func(t *testing.T) {
		d := trigger.DriftDetector{}
		tr := d.Detect(
			[]string{"internal/modules/task"},
			[]string{"docs/adr"},
			[]string{"internal/modules/task", "internal/modules/run", "unexpected/path"},
			now,
		)
		if tr == nil {
			t.Fatal("expected drift detection")
		}
		if tr.AnomalyType == nil || *tr.AnomalyType != string(trigger.AnomalyDrift) {
			t.Fatalf("expected anomaly drift, got %v", tr.AnomalyType)
		}
	})

	t.Run("no drift within scope", func(t *testing.T) {
		d := trigger.DriftDetector{}
		tr := d.Detect(
			[]string{"internal/modules/task"},
			[]string{"docs/adr"},
			[]string{"internal/modules/task", "docs/adr/0001.md"},
			now,
		)
		if tr != nil {
			t.Fatal("expected no drift")
		}
	})

	t.Run("subpath is allowed", func(t *testing.T) {
		d := trigger.DriftDetector{}
		tr := d.Detect(
			[]string{"internal/modules/task"},
			[]string{},
			[]string{"internal/modules/task/service.go"},
			now,
		)
		if tr != nil {
			t.Fatal("expected no drift for subpath")
		}
	})
}

func TestPathViolationDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects violation outside owned paths", func(t *testing.T) {
		d := trigger.PathViolationDetector{}
		tr := d.Detect(
			[]string{"internal/modules/task"},
			[]string{"internal/modules/run/service.go"},
			now,
		)
		if tr == nil {
			t.Fatal("expected path violation")
		}
		if tr.AnomalyType == nil || *tr.AnomalyType != string(trigger.AnomalyPathViolation) {
			t.Fatalf("expected anomaly path_violation, got %v", tr.AnomalyType)
		}
	})

	t.Run("no violation within owned paths", func(t *testing.T) {
		d := trigger.PathViolationDetector{}
		tr := d.Detect(
			[]string{"internal/modules/task"},
			[]string{"internal/modules/task/service.go"},
			now,
		)
		if tr != nil {
			t.Fatal("expected no violation")
		}
	})

	t.Run("subpath is allowed for owned", func(t *testing.T) {
		d := trigger.PathViolationDetector{}
		tr := d.Detect(
			[]string{"internal/modules/task"},
			[]string{"internal/modules/task/repository.go"},
			now,
		)
		if tr != nil {
			t.Fatal("expected no violation for subpath")
		}
	})
}

func TestTokenThresholdDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects when tokens exceed threshold", func(t *testing.T) {
		d := trigger.TokenThresholdDetector{}
		tr := d.Detect(1000, 500, now)
		if tr == nil {
			t.Fatal("expected token threshold trigger")
		}
	})

	t.Run("no trigger when within threshold", func(t *testing.T) {
		d := trigger.TokenThresholdDetector{}
		tr := d.Detect(300, 500, now)
		if tr != nil {
			t.Fatal("expected no trigger")
		}
	})

	t.Run("no trigger at exact threshold", func(t *testing.T) {
		d := trigger.TokenThresholdDetector{}
		tr := d.Detect(500, 500, now)
		if tr != nil {
			t.Fatal("expected no trigger at exact threshold")
		}
	})
}

func TestStepsThresholdDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects when steps exceed threshold", func(t *testing.T) {
		d := trigger.StepsThresholdDetector{}
		tr := d.Detect(15, 10, now)
		if tr == nil {
			t.Fatal("expected steps threshold trigger")
		}
	})

	t.Run("no trigger when within threshold", func(t *testing.T) {
		d := trigger.StepsThresholdDetector{}
		tr := d.Detect(5, 10, now)
		if tr != nil {
			t.Fatal("expected no trigger")
		}
	})
}

func TestTimeThresholdDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects when elapsed time exceeds threshold", func(t *testing.T) {
		d := trigger.TimeThresholdDetector{}
		startedAt := now.Add(-time.Minute * 10)
		tr := d.Detect(startedAt, now, 300)
		if tr == nil {
			t.Fatal("expected time threshold trigger")
		}
	})

	t.Run("no trigger when within threshold", func(t *testing.T) {
		d := trigger.TimeThresholdDetector{}
		startedAt := now.Add(-time.Minute * 2)
		tr := d.Detect(startedAt, now, 300)
		if tr != nil {
			t.Fatal("expected no trigger")
		}
	})

	t.Run("no trigger when startedAt is zero", func(t *testing.T) {
		d := trigger.TimeThresholdDetector{}
		tr := d.Detect(time.Time{}, now, 300)
		if tr != nil {
			t.Fatal("expected no trigger when startedAt is zero")
		}
	})
}
