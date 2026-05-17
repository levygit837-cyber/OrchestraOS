package taskgraph

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
)

// ValidateStatus checks if the status is valid.
func ValidateStatus(s Status, op string) error {
	switch s {
	case StatusActive, StatusSuperseded:
		return nil
	default:
		return apperrors.New(apperrors.CodeValidation, op, "invalid task graph status")
	}
}

// ValidatePlannerStrategy checks if the strategy is valid.
func ValidatePlannerStrategy(strategy string, op string) error {
	switch strategy {
	case "local_heuristic_v1", "llm_gemini_v1":
		return nil
	default:
		return apperrors.New(apperrors.CodeValidation, op, "invalid planner strategy")
	}
}

// ValidateDecomposeInput validates the input for decomposition.
func ValidateDecomposeInput(input DecomposeTaskGraphInput, op string) error {
	if err := validation.RequiredUUID(input.TaskID, "task_id", op); err != nil {
		return err
	}
	if err := ValidatePlannerStrategy(input.PlannerStrategy, op); err != nil {
		return err
	}
	return nil
}
