package validation

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

// RequiredUUID validates that a non-empty string is a valid UUID.
func RequiredUUID(value, field, op string) error {
	if strings.TrimSpace(value) == "" {
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("%s is required", field))
	}
	if _, err := uuid.Parse(value); err != nil {
		return apperrors.Wrap(apperrors.CodeInvalidInput, op, fmt.Errorf("%s must be a UUID: %w", field, err))
	}
	return nil
}

// OptionalUUID validates that, if non-empty, the value is a valid UUID.
func OptionalUUID(value, field, op string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return RequiredUUID(value, field, op)
}

// RequiredText validates that a string is non-empty after trimming.
func RequiredText(value, field, op string) error {
	if strings.TrimSpace(value) == "" {
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("%s is required", field))
	}
	return nil
}

// StringList validates a slice of strings. If required=true, it must be non-empty.
// No element may be empty or whitespace-only.
func StringList(values []string, field, op string, required bool) error {
	if required && len(values) == 0 {
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("%s is required", field))
	}
	for i, value := range values {
		if strings.TrimSpace(value) == "" {
			return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("%s[%d] must not be empty", field, i))
		}
	}
	return nil
}

// Priority validates a priority string against the allowed values.
func Priority(priority string, op string) error {
	switch priority {
	case "P0", "P1", "P2", "P3":
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("invalid priority %q", priority))
	}
}

// RiskLevel validates a risk level string against the allowed values.
func RiskLevel(risk string, op string) error {
	switch risk {
	case "low", "medium", "high", "critical":
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("invalid risk level %q", risk))
	}
}

// Runtime validates an agent runtime string against the allowed values.
func Runtime(runtime, op string) error {
	if runtime == "" {
		return nil
	}
	switch runtime {
	case "codex_cli", "fake", "external", "gemini":
		return nil
	default:
		return apperrors.New(apperrors.CodeInvalidInput, op, fmt.Sprintf("invalid runtime %q", runtime))
	}
}
