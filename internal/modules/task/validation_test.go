package task

import (
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

func TestValidateCreateTaskInputRejectsBlankTitle(t *testing.T) {
	if err := ValidateCreateTaskInput(CreateTaskInput{
		Title:     "   ",
		Priority:  domain.Priority("P9"),
		RiskLevel: domain.RiskLevelLow,
	}); err == nil {
		t.Fatal("expected invalid input for blank title")
	}
}

func TestValidateCreateTaskInputRejectsInvalidPriority(t *testing.T) {
	if err := ValidateCreateTaskInput(CreateTaskInput{
		Title:     "Valid",
		Priority:  domain.Priority("P9"),
		RiskLevel: domain.RiskLevelLow,
	}); err == nil {
		t.Fatal("expected invalid priority to be rejected")
	}
}
