package agent

import "testing"

func TestValidateProfile(t *testing.T) {
	op := "test"

	// Valid profiles (snake_case)
	validProfiles := []string{"code_worker", "docs_writer", "reviewer", "debugger", "default", "new_profile"}
	for _, profile := range validProfiles {
		if err := ValidateProfile(profile, op); err != nil {
			t.Fatalf("expected valid profile %s to pass validation, got error: %v", profile, err)
		}
	}

	// Invalid profiles
	invalidProfiles := []string{"Invalid-Profile", "123profile", "profile with spaces", "UPPER_CASE"}
	for _, profile := range invalidProfiles {
		if err := ValidateProfile(profile, op); err == nil {
			t.Fatalf("expected invalid profile %s to be rejected", profile)
		}
	}

	// Empty profile
	if err := ValidateProfile("", op); err == nil {
		t.Fatal("expected empty profile to be rejected")
	}
}

func TestValidateRuntimeType(t *testing.T) {
	op := "test"

	// Valid runtime types
	validTypes := []RuntimeType{
		RuntimeTypeFake,
		RuntimeTypeGemini,
		RuntimeTypeCodexCLI,
		RuntimeTypeExternal,
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

// Note: Integration tests for AgentService.Create, GetByID, and FindOrCreate
// would require a test database setup. These are typically placed in tests/integration/
// following the project's testing strategy. The validation logic is tested above.
