package review

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func validateGateType(gate domain.ValidationGate, op string) error {
	switch gate {
	case domain.ValidationGateHard, domain.ValidationGateSoft, domain.ValidationGatePolicy:
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid gate_type")
	}
}

func validateVerdict(verdict domain.ReviewDecision, op string) error {
	switch verdict {
	case domain.ReviewStatusApproved, domain.ReviewStatusChangesRequested, domain.ReviewStatusNeedsDiscussion:
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
