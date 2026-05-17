package agent_test

import (
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
)

func TestValidateProfile(t *testing.T) {
	op := "test"

	// Valid profiles (snake_case)
	validProfiles := []string{"code_worker", "docs_writer", "reviewer", "debugger", "default", "new_profile"}
	for _, profile := range validProfiles {
		if err := agent.ValidateProfile(profile, op); err != nil {
			t.Fatalf("expected valid profile %s to pass validation, got error: %v", profile, err)
		}
	}

	// Invalid profiles
	invalidProfiles := []string{"Invalid-Profile", "123profile", "profile with spaces", "UPPER_CASE"}
	for _, profile := range invalidProfiles {
		if err := agent.ValidateProfile(profile, op); err == nil {
			t.Fatalf("expected invalid profile %s to be rejected", profile)
		}
	}

	// Empty profile
	if err := agent.ValidateProfile("", op); err == nil {
		t.Fatal("expected empty profile to be rejected")
	}
}

func TestValidateRuntimeType(t *testing.T) {
	op := "test"

	// Valid runtime types
	validTypes := []agent.RuntimeType{
		agent.RuntimeTypeFake,
		agent.RuntimeTypeGemini,
		agent.RuntimeTypeCodexCLI,
		agent.RuntimeTypeExternal,
	}
	for _, rt := range validTypes {
		if err := agent.ValidateRuntimeType(rt, op); err != nil {
			t.Fatalf("expected valid runtime type %s to pass validation, got error: %v", rt, err)
		}
	}

	// Invalid runtime type
	if err := agent.ValidateRuntimeType("invalid_type", op); err == nil {
		t.Fatal("expected invalid runtime type to be rejected")
	}

	// Empty runtime type
	if err := agent.ValidateRuntimeType("", op); err == nil {
		t.Fatal("expected empty runtime type to be rejected")
	}
}

func TestValidateName(t *testing.T) {
	op := "test"

	// Valid name
	if err := agent.ValidateName("test agent", op); err != nil {
		t.Fatalf("expected valid name to pass validation, got error: %v", err)
	}

	// Empty name
	if err := agent.ValidateName("", op); err == nil {
		t.Fatal("expected empty name to be rejected")
	}

	// Blank name
	if err := agent.ValidateName("   ", op); err == nil {
		t.Fatal("expected blank name to be rejected")
	}
}

// Note: Integration tests for AgentService.Create, GetByID, and FindOrCreate
// would require a test database setup. These are typically placed in tests/integration/
// following the project's testing strategy. The validation logic is tested above.
