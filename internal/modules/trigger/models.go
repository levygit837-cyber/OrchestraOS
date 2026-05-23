package trigger

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// Aliases to shared domain types per ADR-0030.

type Trigger = domain.Trigger
type Status = domain.TriggerStatus
type Type = domain.TriggerType

const (
	StatusActive    = domain.TriggerStatusActive
	StatusTriggered = domain.TriggerStatusTriggered
	StatusResolved  = domain.TriggerStatusResolved
	StatusDismissed = domain.TriggerStatusDismissed

	TypeThreshold        = domain.TriggerTypeThreshold
	TypeAnomaly          = domain.TriggerTypeAnomaly
	TypeHeartbeatTimeout = domain.TriggerTypeHeartbeatTimeout
	TypePolicy           = domain.TriggerTypePolicy
)

// Helper functions for type conversions.

func stringPtr(s string) *string {
	return &s
}

func anomalyTypePtr(a AnomalyType) *string {
	return stringPtr(string(a))
}

func anomalyTypePtrDeref(a *AnomalyType) *string {
	if a == nil {
		return nil
	}
	return stringPtr(string(*a))
}

func resolutionActionPtr(r ResolutionAction) *string {
	return stringPtr(string(r))
}

func resolutionActionPtrDeref(r *ResolutionAction) *string {
	if r == nil {
		return nil
	}
	return stringPtr(string(*r))
}

// Local types (not shared across modules).

type AnomalyType string

const (
	AnomalyStall         AnomalyType = "stall"
	AnomalyLoop          AnomalyType = "loop"
	AnomalyDrift         AnomalyType = "drift"
	AnomalyPathViolation AnomalyType = "path_violation"
	AnomalyTokenExceeded AnomalyType = "token_exceeded"
	AnomalyStepsExceeded AnomalyType = "steps_exceeded"
	AnomalyTimeExceeded  AnomalyType = "time_exceeded"
)

type ResolutionAction string

const (
	ResolutionPause    ResolutionAction = "pause"
	ResolutionCancel   ResolutionAction = "cancel"
	ResolutionNotify   ResolutionAction = "notify"
	ResolutionEscalate ResolutionAction = "escalate"
)

type ThresholdConfig struct {
	StallSeconds    int `json:"stall_seconds"`
	LoopRepetitions int `json:"loop_repetitions"`
	TokenMax        int `json:"token_max"`
	StepsMax        int `json:"steps_max"`
	TimeMaxSeconds  int `json:"time_max_seconds"`
}
