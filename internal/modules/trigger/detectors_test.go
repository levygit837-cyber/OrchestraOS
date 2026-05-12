package trigger

import (
	"testing"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func TestStallDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	startedAt := now.Add(-time.Minute * 10)

	t.Run("detects stall when no events and exceeded threshold", func(t *testing.T) {
		d := StallDetector{}
		trigger := d.Detect(nil, startedAt, now, 300)
		if trigger == nil {
			t.Fatal("expected stall detection")
		}
		if trigger.AnomalyType == nil || *trigger.AnomalyType != domain.AnomalyTypeStall {
			t.Fatalf("expected anomaly stall, got %v", trigger.AnomalyType)
		}
	})

	t.Run("no stall when within threshold", func(t *testing.T) {
		d := StallDetector{}
		lastEvent := now.Add(-time.Minute * 2)
		trigger := d.Detect(&lastEvent, startedAt, now, 300)
		if trigger != nil {
			t.Fatal("expected no stall")
		}
	})

	t.Run("detects stall from last event", func(t *testing.T) {
		d := StallDetector{}
		lastEvent := now.Add(-time.Minute * 10)
		trigger := d.Detect(&lastEvent, startedAt, now, 300)
		if trigger == nil {
			t.Fatal("expected stall detection from last event")
		}
	})
}

func TestLoopDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects loop with repeated events", func(t *testing.T) {
		d := LoopDetector{}
		events := []string{"heartbeat", "heartbeat", "heartbeat", "heartbeat", "heartbeat"}
		trigger := d.Detect(events, 5, now)
		if trigger == nil {
			t.Fatal("expected loop detection")
		}
		if trigger.AnomalyType == nil || *trigger.AnomalyType != domain.AnomalyTypeLoop {
			t.Fatalf("expected anomaly loop, got %v", trigger.AnomalyType)
		}
	})

	t.Run("no loop below threshold", func(t *testing.T) {
		d := LoopDetector{}
		events := []string{"heartbeat", "heartbeat", "heartbeat"}
		trigger := d.Detect(events, 5, now)
		if trigger != nil {
			t.Fatal("expected no loop")
		}
	})

	t.Run("no loop with varied events", func(t *testing.T) {
		d := LoopDetector{}
		events := []string{"heartbeat", "checkpoint", "heartbeat", "checkpoint", "heartbeat"}
		trigger := d.Detect(events, 3, now)
		if trigger != nil {
			t.Fatal("expected no loop with varied events")
		}
	})

	t.Run("ignores trigger events to prevent feedback loop", func(t *testing.T) {
		d := LoopDetector{}
		events := []string{"heartbeat", "trigger.created", "heartbeat", "trigger.created", "heartbeat"}
		trigger := d.Detect(events, 4, now)
		if trigger != nil {
			t.Fatal("expected no loop when trigger events break sequence")
		}
	})
}

func TestDriftDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects drift outside owned and read paths", func(t *testing.T) {
		d := DriftDetector{}
		owned := []string{"/app/src"}
		read := []string{"/app/docs"}
		accessed := []string{"/app/src", "/app/config"}
		trigger := d.Detect(owned, read, accessed, now)
		if trigger == nil {
			t.Fatal("expected drift detection")
		}
		if trigger.AnomalyType == nil || *trigger.AnomalyType != domain.AnomalyTypeDrift {
			t.Fatalf("expected anomaly drift, got %v", trigger.AnomalyType)
		}
	})

	t.Run("no drift within scope", func(t *testing.T) {
		d := DriftDetector{}
		owned := []string{"/app/src"}
		read := []string{"/app/docs"}
		accessed := []string{"/app/src", "/app/docs"}
		trigger := d.Detect(owned, read, accessed, now)
		if trigger != nil {
			t.Fatal("expected no drift")
		}
	})

	t.Run("subpath is allowed", func(t *testing.T) {
		d := DriftDetector{}
		owned := []string{"/app/src"}
		read := []string{}
		accessed := []string{"/app/src/components"}
		trigger := d.Detect(owned, read, accessed, now)
		if trigger != nil {
			t.Fatal("expected no drift for subpath")
		}
	})
}

func TestPathViolationDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects violation outside owned paths", func(t *testing.T) {
		d := PathViolationDetector{}
		owned := []string{"/app/src"}
		modified := []string{"/app/config"}
		trigger := d.Detect(owned, modified, now)
		if trigger == nil {
			t.Fatal("expected path violation")
		}
		if trigger.AnomalyType == nil || *trigger.AnomalyType != domain.AnomalyTypePathViolation {
			t.Fatalf("expected anomaly path_violation, got %v", trigger.AnomalyType)
		}
	})

	t.Run("no violation within owned paths", func(t *testing.T) {
		d := PathViolationDetector{}
		owned := []string{"/app/src"}
		modified := []string{"/app/src/file.go"}
		trigger := d.Detect(owned, modified, now)
		if trigger != nil {
			t.Fatal("expected no violation")
		}
	})

	t.Run("subpath is allowed for owned", func(t *testing.T) {
		d := PathViolationDetector{}
		owned := []string{"/app/src"}
		modified := []string{"/app/src/components/button.go"}
		trigger := d.Detect(owned, modified, now)
		if trigger != nil {
			t.Fatal("expected no violation for subpath")
		}
	})
}

func TestTokenThresholdDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects when tokens exceed threshold", func(t *testing.T) {
		d := TokenThresholdDetector{}
		trigger := d.Detect(150, 100, now)
		if trigger == nil {
			t.Fatal("expected token threshold trigger")
		}
		if trigger.AnomalyType == nil || *trigger.AnomalyType != domain.AnomalyTypeTokenExceeded {
			t.Fatalf("expected token_exceeded, got %v", trigger.AnomalyType)
		}
	})

	t.Run("no trigger when within threshold", func(t *testing.T) {
		d := TokenThresholdDetector{}
		trigger := d.Detect(50, 100, now)
		if trigger != nil {
			t.Fatal("expected no trigger")
		}
	})

	t.Run("no trigger at exact threshold", func(t *testing.T) {
		d := TokenThresholdDetector{}
		trigger := d.Detect(100, 100, now)
		if trigger != nil {
			t.Fatal("expected no trigger at exact threshold")
		}
	})
}

func TestStepsThresholdDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects when steps exceed threshold", func(t *testing.T) {
		d := StepsThresholdDetector{}
		trigger := d.Detect(101, 100, now)
		if trigger == nil {
			t.Fatal("expected steps threshold trigger")
		}
		if trigger.AnomalyType == nil || *trigger.AnomalyType != domain.AnomalyTypeStepsExceeded {
			t.Fatalf("expected steps_exceeded, got %v", trigger.AnomalyType)
		}
	})

	t.Run("no trigger when within threshold", func(t *testing.T) {
		d := StepsThresholdDetector{}
		trigger := d.Detect(50, 100, now)
		if trigger != nil {
			t.Fatal("expected no trigger")
		}
	})
}

func TestTimeThresholdDetector(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("detects when elapsed time exceeds threshold", func(t *testing.T) {
		d := TimeThresholdDetector{}
		startedAt := now.Add(-time.Hour * 2)
		trigger := d.Detect(startedAt, now, 3600)
		if trigger == nil {
			t.Fatal("expected time threshold trigger")
		}
		if trigger.AnomalyType == nil || *trigger.AnomalyType != domain.AnomalyTypeTimeExceeded {
			t.Fatalf("expected time_exceeded, got %v", trigger.AnomalyType)
		}
	})

	t.Run("no trigger when within threshold", func(t *testing.T) {
		d := TimeThresholdDetector{}
		startedAt := now.Add(-time.Minute * 30)
		trigger := d.Detect(startedAt, now, 3600)
		if trigger != nil {
			t.Fatal("expected no trigger")
		}
	})

	t.Run("no trigger when startedAt is zero", func(t *testing.T) {
		d := TimeThresholdDetector{}
		trigger := d.Detect(time.Time{}, now, 3600)
		if trigger != nil {
			t.Fatal("expected no trigger for zero startedAt")
		}
	})
}
