package agent

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/core/validation"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

var validProfiles = map[string]bool{
	"code_worker": true,
	"docs_writer": true,
	"reviewer":    true,
	"debugger":    true,
	"default":     true,
}

var validRuntimeTypes = map[domain.AgentRuntimeType]bool{
	domain.AgentRuntimeTypeFake:     true,
	domain.AgentRuntimeTypeGemini:   true,
	domain.AgentRuntimeTypeCodexCLI: true,
	domain.AgentRuntimeTypeExternal: true,
}

// ValidateProfile checks if the profile is valid
func ValidateProfile(profile string, op string) error {
	if err := validation.RequiredText(profile, "profile", op); err != nil {
		return err
	}
	if !validProfiles[profile] {
		return apperrors.New(apperrors.CodeValidation, op, "invalid profile: must be one of code_worker, docs_writer, reviewer, debugger, default")
	}
	return nil
}

// ValidateRuntimeType checks if the runtime type is valid
func ValidateRuntimeType(runtimeType domain.AgentRuntimeType, op string) error {
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
