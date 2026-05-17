package trigger

import (
	"encoding/json"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
)

func validateCreateTriggerInput(input CreateTriggerInput) error {
	op := "trigger_service.validate_create"
	if err := validation.RequiredUUID(input.ID, "id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.RunID, "run_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.TaskID, "task_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.AgentSessionID, "agent_session_id", op); err != nil {
		return err
	}
	if !isValidTriggerType(input.TriggerType) {
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid trigger_type")
	}
	if input.Status == "" {
		input.Status = StatusActive
	}
	if !isValidTriggerStatus(input.Status) {
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid status")
	}
	if input.AnomalyType != nil && !isValidAnomalyType(*input.AnomalyType) {
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid anomaly_type")
	}
	if input.ResolutionAction != nil && !isValidResolutionAction(*input.ResolutionAction) {
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid resolution_action")
	}
	if input.ThresholdValue != nil {
		if !json.Valid(input.ThresholdValue) {
			return apperrors.New(apperrors.CodeInvalidInput, op, "threshold_value must be valid JSON")
		}
	}
	if input.CurrentValue != nil {
		if !json.Valid(input.CurrentValue) {
			return apperrors.New(apperrors.CodeInvalidInput, op, "current_value must be valid JSON")
		}
	}
	return nil
}

func isValidTriggerType(t Type) bool {
	switch t {
	case TypeThreshold, TypeAnomaly, TypeHeartbeatTimeout, TypePolicy:
		return true
	}
	return false
}

func isValidTriggerStatus(s Status) bool {
	switch s {
	case StatusActive, StatusTriggered, StatusResolved, StatusDismissed:
		return true
	}
	return false
}

func isValidAnomalyType(a AnomalyType) bool {
	switch a {
	case AnomalyStall, AnomalyLoop, AnomalyDrift,
		AnomalyPathViolation, AnomalyTokenExceeded,
		AnomalyStepsExceeded, AnomalyTimeExceeded:
		return true
	}
	return false
}

func isValidResolutionAction(r ResolutionAction) bool {
	switch r {
	case ResolutionPause, ResolutionCancel,
		ResolutionNotify, ResolutionEscalate:
		return true
	}
	return false
}
