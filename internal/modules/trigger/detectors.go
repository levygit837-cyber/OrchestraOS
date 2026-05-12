package trigger

import (
	"encoding/json"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// StallDetector detects when no events have been emitted for a threshold duration.
type StallDetector struct{}

// Detect returns a trigger if the time since the last event (or run start) exceeds thresholdSeconds.
func (d StallDetector) Detect(lastEventAt *time.Time, startedAt time.Time, now time.Time, thresholdSeconds int) *domain.Trigger {
	reference := startedAt
	if lastEventAt != nil && lastEventAt.After(reference) {
		reference = *lastEventAt
	}
	if now.Sub(reference) >= time.Duration(thresholdSeconds)*time.Second {
		anomaly := domain.AnomalyTypeStall
		return &domain.Trigger{
			TriggerType: domain.TriggerTypeAnomaly,
			Status:      domain.TriggerStatusTriggered,
			AnomalyType: &anomaly,
			ThresholdValue: mustMarshal(map[string]interface{}{
				"stall_seconds": thresholdSeconds,
			}),
			CurrentValue: mustMarshal(map[string]interface{}{
				"seconds_since_last_event": int(now.Sub(reference).Seconds()),
			}),
			TriggeredAt: &now,
		}
	}
	return nil
}

// LoopDetector detects when the same event type repeats N times in sequence.
type LoopDetector struct{}

// Detect returns a trigger if the last threshold event types are identical.
// It skips trigger events to avoid infinite loops.
func (d LoopDetector) Detect(eventTypes []string, threshold int, now time.Time) *domain.Trigger {
	if threshold < 2 || len(eventTypes) < threshold {
		return nil
	}
	// Filter out trigger events to avoid feedback loops
	filtered := make([]string, 0, len(eventTypes))
	for _, et := range eventTypes {
		if len(et) > 8 && et[:8] == "trigger." {
			continue
		}
		filtered = append(filtered, et)
	}
	if len(filtered) < threshold {
		return nil
	}
	lastType := filtered[len(filtered)-1]
	for i := len(filtered) - 2; i >= len(filtered)-threshold; i-- {
		if filtered[i] != lastType {
			return nil
		}
	}
	anomaly := domain.AnomalyTypeLoop
	return &domain.Trigger{
		TriggerType: domain.TriggerTypeAnomaly,
		Status:      domain.TriggerStatusTriggered,
		AnomalyType: &anomaly,
		ThresholdValue: mustMarshal(map[string]interface{}{
			"loop_repetitions": threshold,
		}),
		CurrentValue: mustMarshal(map[string]interface{}{
			"repeated_event_type": lastType,
			"actual_repetitions":  threshold,
		}),
		TriggeredAt: &now,
	}
}

// DriftDetector detects when accessed paths fall outside owned_paths and read_paths.
type DriftDetector struct{}

// Detect returns a trigger for each accessed path outside the allowed scope.
func (d DriftDetector) Detect(ownedPaths, readPaths, accessedPaths []string, now time.Time) *domain.Trigger {
	allowed := make(map[string]bool)
	for _, p := range ownedPaths {
		allowed[normalizePath(p)] = true
	}
	for _, p := range readPaths {
		allowed[normalizePath(p)] = true
	}
	var driftPaths []string
	for _, p := range accessedPaths {
		norm := normalizePath(p)
		if !allowed[norm] && !isSubPathOfAny(norm, ownedPaths) && !isSubPathOfAny(norm, readPaths) {
			driftPaths = append(driftPaths, p)
		}
	}
	if len(driftPaths) == 0 {
		return nil
	}
	anomaly := domain.AnomalyTypeDrift
	return &domain.Trigger{
		TriggerType: domain.TriggerTypeAnomaly,
		Status:      domain.TriggerStatusTriggered,
		AnomalyType: &anomaly,
		CurrentValue: mustMarshal(map[string]interface{}{
			"drift_paths": driftPaths,
		}),
		TriggeredAt: &now,
	}
}

// PathViolationDetector detects when modified paths are not in owned_paths.
type PathViolationDetector struct{}

// Detect returns a trigger for each modified path outside owned_paths.
func (d PathViolationDetector) Detect(ownedPaths, modifiedPaths []string, now time.Time) *domain.Trigger {
	owned := make(map[string]bool)
	for _, p := range ownedPaths {
		owned[normalizePath(p)] = true
	}
	var violations []string
	for _, p := range modifiedPaths {
		norm := normalizePath(p)
		if !owned[norm] && !isSubPathOfAny(norm, ownedPaths) {
			violations = append(violations, p)
		}
	}
	if len(violations) == 0 {
		return nil
	}
	anomaly := domain.AnomalyTypePathViolation
	return &domain.Trigger{
		TriggerType: domain.TriggerTypeAnomaly,
		Status:      domain.TriggerStatusTriggered,
		AnomalyType: &anomaly,
		CurrentValue: mustMarshal(map[string]interface{}{
			"violation_paths": violations,
		}),
		TriggeredAt: &now,
	}
}

// TokenThresholdDetector detects when consumed tokens exceed a threshold.
type TokenThresholdDetector struct{}

// Detect returns a trigger if currentTokens > maxTokens.
func (d TokenThresholdDetector) Detect(currentTokens, maxTokens int, now time.Time) *domain.Trigger {
	if currentTokens <= maxTokens {
		return nil
	}
	anomaly := domain.AnomalyTypeTokenExceeded
	return &domain.Trigger{
		TriggerType: domain.TriggerTypeThreshold,
		Status:      domain.TriggerStatusTriggered,
		AnomalyType: &anomaly,
		ThresholdValue: mustMarshal(map[string]interface{}{
			"token_max": maxTokens,
		}),
		CurrentValue: mustMarshal(map[string]interface{}{
			"current_tokens": currentTokens,
		}),
		TriggeredAt: &now,
	}
}

// StepsThresholdDetector detects when executed steps exceed a threshold.
type StepsThresholdDetector struct{}

// Detect returns a trigger if currentSteps > maxSteps.
func (d StepsThresholdDetector) Detect(currentSteps, maxSteps int, now time.Time) *domain.Trigger {
	if currentSteps <= maxSteps {
		return nil
	}
	anomaly := domain.AnomalyTypeStepsExceeded
	return &domain.Trigger{
		TriggerType: domain.TriggerTypeThreshold,
		Status:      domain.TriggerStatusTriggered,
		AnomalyType: &anomaly,
		ThresholdValue: mustMarshal(map[string]interface{}{
			"steps_max": maxSteps,
		}),
		CurrentValue: mustMarshal(map[string]interface{}{
			"current_steps": currentSteps,
		}),
		TriggeredAt: &now,
	}
}

// TimeThresholdDetector detects when total execution time exceeds a threshold.
type TimeThresholdDetector struct{}

// Detect returns a trigger if now-startedAt exceeds maxSeconds.
func (d TimeThresholdDetector) Detect(startedAt time.Time, now time.Time, maxSeconds int) *domain.Trigger {
	if startedAt.IsZero() || now.Sub(startedAt) < time.Duration(maxSeconds)*time.Second {
		return nil
	}
	elapsed := int(now.Sub(startedAt).Seconds())
	anomaly := domain.AnomalyTypeTimeExceeded
	return &domain.Trigger{
		TriggerType: domain.TriggerTypeThreshold,
		Status:      domain.TriggerStatusTriggered,
		AnomalyType: &anomaly,
		ThresholdValue: mustMarshal(map[string]interface{}{
			"time_max_seconds": maxSeconds,
		}),
		CurrentValue: mustMarshal(map[string]interface{}{
			"elapsed_seconds": elapsed,
		}),
		TriggeredAt: &now,
	}
}

// normalizePath normalizes a path for comparison.
func normalizePath(p string) string {
	// simplistic normalization; handles basic absolute vs relative
	if len(p) > 0 && p[0] != '/' {
		return "/" + p
	}
	return p
}

// isSubPathOfAny checks if child is a subpath of any path in parents.
func isSubPathOfAny(child string, parents []string) bool {
	for _, parent := range parents {
		normParent := normalizePath(parent)
		if child == normParent {
			return true
		}
		if len(child) > len(normParent) && child[:len(normParent)+1] == normParent+"/" {
			return true
		}
	}
	return false
}

func mustMarshal(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return json.RawMessage(b)
}
