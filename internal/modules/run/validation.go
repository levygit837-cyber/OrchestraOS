package run

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
)

// ValidateStatus checks if the status is valid.
func ValidateStatus(s Status, op string) error {
	switch s {
	case StatusCreated, StatusRunning, StatusWaitingApproval, StatusPaused,
		StatusValidating, StatusCompleted, StatusFailed, StatusCancelled:
		return nil
	default:
		return apperrors.New(apperrors.CodeValidation, op, "invalid run status")
	}
}

// ValidateResult checks if the result is valid.
func ValidateResult(r Result, op string) error {
	switch r {
	case ResultSucceeded, ResultFailed, ResultCancelled:
		return nil
	default:
		return apperrors.New(apperrors.CodeValidation, op, "invalid run result")
	}
}

// ValidateCreateRunInput validates the input for creating a run.
func ValidateCreateRunInput(input CreateRunInput) error {
	op := "run.validate_create"
	if err := validation.RequiredUUID(input.TaskID, "task_id", op); err != nil {
		return err
	}
	if err := validation.RequiredUUID(input.WorkUnitID, "work_unit_id", op); err != nil {
		return err
	}
	if input.Attempt < 1 {
		return apperrors.New(apperrors.CodeInvalidInput, op, "attempt must be greater than zero")
	}
	return nil
}
