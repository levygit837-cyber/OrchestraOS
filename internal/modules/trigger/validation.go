package trigger

import (
	"encoding/json"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
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
		input.Status = domain.TriggerStatusActive
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

func isValidTriggerType(t domain.TriggerType) bool {
	switch t {
	case domain.TriggerTypeThreshold, domain.TriggerTypeAnomaly, domain.TriggerTypeHeartbeatTimeout, domain.TriggerTypePolicy:
		return true
	}
	return false
}

func isValidTriggerStatus(s domain.TriggerStatus) bool {
	switch s {
	case domain.TriggerStatusActive, domain.TriggerStatusTriggered, domain.TriggerStatusResolved, domain.TriggerStatusDismissed:
		return true
	}
	return false
}

func isValidAnomalyType(a domain.AnomalyType) bool {
	switch a {
	case domain.AnomalyTypeStall, domain.AnomalyTypeLoop, domain.AnomalyTypeDrift,
		domain.AnomalyTypePathViolation, domain.AnomalyTypeTokenExceeded,
		domain.AnomalyTypeStepsExceeded, domain.AnomalyTypeTimeExceeded:
		return true
	}
	return false
}

func isValidResolutionAction(r domain.ResolutionAction) bool {
	switch r {
	case domain.ResolutionActionPause, domain.ResolutionActionCancel,
		domain.ResolutionActionNotify, domain.ResolutionActionEscalate:
		return true
	}
	return false
}
