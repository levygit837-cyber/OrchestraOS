package trigger

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

// DefaultThresholds returns conservative default thresholds.
func DefaultThresholds() ThresholdConfig {
	return ThresholdConfig{
		StallSeconds:    300,
		LoopRepetitions: 5,
		TokenMax:        100000,
		StepsMax:        100,
		TimeMaxSeconds:  3600,
	}
}

// ValidateThresholds validates that threshold values are positive and reasonable.
func ValidateThresholds(cfg ThresholdConfig) error {
	if cfg.StallSeconds < 1 {
		return apperrors.New(apperrors.CodeInvalidInput, "trigger.validate_thresholds", "stall_seconds must be >= 1")
	}
	if cfg.LoopRepetitions < 2 {
		return apperrors.New(apperrors.CodeInvalidInput, "trigger.validate_thresholds", "loop_repetitions must be >= 2")
	}
	if cfg.TokenMax < 1 {
		return apperrors.New(apperrors.CodeInvalidInput, "trigger.validate_thresholds", "token_max must be >= 1")
	}
	if cfg.StepsMax < 1 {
		return apperrors.New(apperrors.CodeInvalidInput, "trigger.validate_thresholds", "steps_max must be >= 1")
	}
	if cfg.TimeMaxSeconds < 1 {
		return apperrors.New(apperrors.CodeInvalidInput, "trigger.validate_thresholds", "time_max_seconds must be >= 1")
	}
	return nil
}
