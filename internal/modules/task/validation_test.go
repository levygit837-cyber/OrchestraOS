package task

import "testing"

func TestValidateCreateTaskInputRejectsBlankTitle(t *testing.T) {
	if err := ValidateCreateTaskInput(CreateTaskInput{
		Title:     "   ",
		Priority:  Priority("P9"),
		RiskLevel: RiskLevelLow,
	}); err == nil {
		t.Fatal("expected invalid input for blank title")
	}
}

func TestValidateCreateTaskInputRejectsInvalidPriority(t *testing.T) {
	if err := ValidateCreateTaskInput(CreateTaskInput{
		Title:     "Valid",
		Priority:  Priority("P9"),
		RiskLevel: RiskLevelLow,
	}); err == nil {
		t.Fatal("expected invalid priority to be rejected")
	}
}
