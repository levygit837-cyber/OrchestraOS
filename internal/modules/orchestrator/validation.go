package orchestrator

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// ValidateRunTaskOptions validates the options for RunTask.
func ValidateRunTaskOptions(options RunTaskOptions) error {
	const op = "orchestrator.validate_run_task_options"

	// Validate runtime type
	if options.RuntimeType == "" {
		options.RuntimeType = RuntimeTypeFake
	}
	switch options.RuntimeType {
	case RuntimeTypeFake, RuntimeTypeGemini, RuntimeTypeCodexCLI:
		// valid
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid runtime_type: must be fake, gemini, or codex_cli")
	}

	// Validate planner strategy
	if options.PlannerStrategy == "" {
		options.PlannerStrategy = PlannerStrategyLocalHeuristic
	}
	switch options.PlannerStrategy {
	case PlannerStrategyLocalHeuristic, PlannerStrategyLLMGemini:
		// valid
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid planner_strategy: must be local_heuristic_v1 or llm_gemini_v1")
	}

	// Validate max steps
	if options.MaxSteps <= 0 {
		options.MaxSteps = DefaultMaxSteps
	}
	if options.MaxSteps < 1 || options.MaxSteps > 1000 {
		return apperrors.New(apperrors.CodeInvalidInput, op, "max_steps must be between 1 and 1000")
	}

	// Validate timeout
	if options.TimeoutSeconds <= 0 {
		options.TimeoutSeconds = DefaultTimeoutSeconds
	}
	if options.TimeoutSeconds < 10 || options.TimeoutSeconds > 3600 {
		return apperrors.New(apperrors.CodeInvalidInput, op, "timeout_seconds must be between 10 and 3600")
	}

	return nil
}

// ValidateRuntimeType validates a runtime type string.
func ValidateRuntimeType(runtimeType string) error {
	const op = "orchestrator.validate_runtime_type"
	switch runtimeType {
	case RuntimeTypeFake, RuntimeTypeGemini, RuntimeTypeCodexCLI:
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid runtime_type")
	}
}

// ValidatePlannerStrategy validates a planner strategy string.
func ValidatePlannerStrategy(strategy string) error {
	const op = "orchestrator.validate_planner_strategy"
	switch strategy {
	case PlannerStrategyLocalHeuristic, PlannerStrategyLLMGemini:
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid planner_strategy")
	}
}

// ValidateAgentProfile validates an agent profile string.
func ValidateAgentProfile(profile string) error {
	op := "orchestrator.validate_agent_profile"
	if err := validation.RequiredText(profile, "profile", op); err != nil {
		return err
	}
	switch profile {
	case "code_worker", "docs_writer", "reviewer", "debugger", "default":
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, "invalid agent profile")
	}
}

// ValidateTaskID validates a task ID.
func ValidateTaskID(taskID string) error {
	return validation.RequiredUUID(taskID, "task_id", "orchestrator.validate_task_id")
}

// ConvertRuntimeType converts a string to domain.AgentRuntimeType.
func ConvertRuntimeType(runtimeType string) domain.AgentRuntimeType {
	switch runtimeType {
	case RuntimeTypeFake:
		return domain.AgentRuntimeTypeFake
	case RuntimeTypeGemini:
		return domain.AgentRuntimeTypeGemini
	case RuntimeTypeCodexCLI:
		return domain.AgentRuntimeTypeCodexCLI
	default:
		return domain.AgentRuntimeTypeFake
	}
}
