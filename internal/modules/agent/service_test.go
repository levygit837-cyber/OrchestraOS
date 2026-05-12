package agent

import (
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func TestValidateProfile(t *testing.T) {
	op := "test"

	// Valid profiles
	validProfiles := []string{"code_worker", "docs_writer", "reviewer", "debugger", "default"}
	for _, profile := range validProfiles {
		if err := ValidateProfile(profile, op); err != nil {
			t.Fatalf("expected valid profile %s to pass validation, got error: %v", profile, err)
		}
	}

	// Invalid profile
	if err := ValidateProfile("invalid_profile", op); err == nil {
		t.Fatal("expected invalid profile to be rejected")
	}

	// Empty profile
	if err := ValidateProfile("", op); err == nil {
		t.Fatal("expected empty profile to be rejected")
	}
}

func TestValidateRuntimeType(t *testing.T) {
	op := "test"

	// Valid runtime types
	validTypes := []domain.AgentRuntimeType{
		domain.AgentRuntimeTypeFake,
		domain.AgentRuntimeTypeGemini,
		domain.AgentRuntimeTypeCodexCLI,
		domain.AgentRuntimeTypeExternal,
	}
	for _, rt := range validTypes {
		if err := ValidateRuntimeType(rt, op); err != nil {
			t.Fatalf("expected valid runtime type %s to pass validation, got error: %v", rt, err)
		}
	}

	// Invalid runtime type
	if err := ValidateRuntimeType("invalid_type", op); err == nil {
		t.Fatal("expected invalid runtime type to be rejected")
	}

	// Empty runtime type
	if err := ValidateRuntimeType("", op); err == nil {
		t.Fatal("expected empty runtime type to be rejected")
	}
}

func TestValidateName(t *testing.T) {
	op := "test"

	// Valid name
	if err := ValidateName("test agent", op); err != nil {
		t.Fatalf("expected valid name to pass validation, got error: %v", err)
	}

	// Empty name
	if err := ValidateName("", op); err == nil {
		t.Fatal("expected empty name to be rejected")
	}

	// Blank name
	if err := ValidateName("   ", op); err == nil {
		t.Fatal("expected blank name to be rejected")
	}
}

func TestToDomainRuntimeType(t *testing.T) {
	tests := []struct {
		local    RuntimeType
		expected domain.AgentRuntimeType
	}{
		{RuntimeTypeFake, domain.AgentRuntimeTypeFake},
		{RuntimeTypeGemini, domain.AgentRuntimeTypeGemini},
		{RuntimeTypeCodexCLI, domain.AgentRuntimeTypeCodexCLI},
		{RuntimeTypeExternal, domain.AgentRuntimeTypeExternal},
	}

	for _, tt := range tests {
		result := ToDomainRuntimeType(tt.local)
		if result != tt.expected {
			t.Fatalf("expected %s, got %s", tt.expected, result)
		}
	}
}

func TestFromDomainRuntimeType(t *testing.T) {
	tests := []struct {
		domain   domain.AgentRuntimeType
		expected RuntimeType
	}{
		{domain.AgentRuntimeTypeFake, RuntimeTypeFake},
		{domain.AgentRuntimeTypeGemini, RuntimeTypeGemini},
		{domain.AgentRuntimeTypeCodexCLI, RuntimeTypeCodexCLI},
		{domain.AgentRuntimeTypeExternal, RuntimeTypeExternal},
	}

	for _, tt := range tests {
		result := FromDomainRuntimeType(tt.domain)
		if result != tt.expected {
			t.Fatalf("expected %s, got %s", tt.expected, result)
		}
	}
}

// Note: Integration tests for AgentService.Create, GetByID, and FindOrCreate
// would require a test database setup. These are typically placed in tests/integration/
// following the project's testing strategy. The validation logic is tested above.
