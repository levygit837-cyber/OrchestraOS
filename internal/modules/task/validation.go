package task

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
)

// ValidateStatus checks if the status is valid.
func ValidateStatus(s Status, op string) error {
	switch s {
	case StatusCreated, StatusTriaged, StatusPlanned, StatusScheduled,
		StatusSandboxPreparing, StatusRunning, StatusWaitingApproval,
		StatusPaused, StatusValidating, StatusCompleted, StatusFailed, StatusCancelled:
		return nil
	default:
		return apperrors.New(apperrors.CodeValidation, op, "invalid task status")
	}
}

// ValidatePriority checks if the priority is valid.
func ValidatePriority(p Priority, op string) error {
	switch p {
	case PriorityP0, PriorityP1, PriorityP2, PriorityP3:
		return nil
	default:
		return apperrors.New(apperrors.CodeValidation, op, "invalid priority")
	}
}

// ValidateRiskLevel checks if the risk level is valid.
func ValidateRiskLevel(r RiskLevel, op string) error {
	switch r {
	case RiskLevelLow, RiskLevelMedium, RiskLevelHigh, RiskLevelCritical:
		return nil
	default:
		return apperrors.New(apperrors.CodeValidation, op, "invalid risk level")
	}
}

// ValidateCreateTaskInput validates the input for creating a task.
func ValidateCreateTaskInput(input CreateTaskInput) error {
	op := "task.validate_create"
	if err := validation.OptionalUUID(input.ID, "task_id", op); err != nil {
		return err
	}
	if err := validation.OptionalUUID(input.EventID, "event_id", op); err != nil {
		return err
	}
	if err := validation.RequiredText(input.Title, "title", op); err != nil {
		return err
	}
	if err := ValidatePriority(input.Priority, op); err != nil {
		return err
	}
	if err := ValidateRiskLevel(input.RiskLevel, op); err != nil {
		return err
	}
	if err := validation.StringList(input.AcceptanceCriteria, "acceptance_criteria", op, false); err != nil {
		return err
	}
	return nil
}
