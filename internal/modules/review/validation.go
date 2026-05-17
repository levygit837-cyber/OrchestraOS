package review

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
)

func validateGateType(gate ValidationGate, op string) error {
	switch gate {
	case GateHard, GateSoft, GatePolicy:
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid gate_type")
	}
}

func validateVerdict(verdict Decision, op string) error {
	switch verdict {
	case StatusApproved, StatusChangesRequested, StatusNeedsDiscussion:
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid verdict")
	}
}

func validateCreateReviewInput(input CreateReviewInput) error {
	op := "review_service.validate_create"
	if err := validation.OptionalUUID(input.ID, "review_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.EventID, "event_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.RunID, "run_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.WorkUnitID, "work_unit_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.TaskID, "task_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.AgentSessionID, "agent_session_id", op); err != nil {
		return err
	}
	if err := validateGateType(input.GateType, op); err != nil {
		return err
	}
	return nil
}
