package agent

import (
	"regexp"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
)

var validRuntimeTypes = map[RuntimeType]bool{
	RuntimeTypeFake:     true,
	RuntimeTypeGemini:   true,
	RuntimeTypeCodexCLI: true,
	RuntimeTypeExternal: true,
}

var profilePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ValidateProfile checks if the profile is valid (snake_case, non-empty).
// The database CHECK constraint enforces the allowed set of values.
func ValidateProfile(profile string, op string) error {
	if err := validation.RequiredText(profile, "profile", op); err != nil {
		return err
	}
	if !profilePattern.MatchString(profile) {
		return apperrors.New(apperrors.CodeValidation, op, "invalid profile: must be snake_case starting with a letter")
	}
	return nil
}

// ValidateRuntimeType checks if the runtime type is valid
func ValidateRuntimeType(runtimeType RuntimeType, op string) error {
	if runtimeType == "" {
		return apperrors.New(apperrors.CodeValidation, op, "runtime_type is required")
	}
	if !validRuntimeTypes[runtimeType] {
		return apperrors.New(apperrors.CodeValidation, op, "invalid runtime_type: must be one of fake, gemini, codex_cli, external")
	}
	return nil
}

// ValidateName checks if the name is valid
func ValidateName(name string, op string) error {
	return validation.RequiredText(name, "name", op)
}
