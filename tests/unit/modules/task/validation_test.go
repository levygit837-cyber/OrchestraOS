package task_test

import (
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
)

func TestValidateCreateTaskInputRejectsBlankTitle(t *testing.T) {
	if err := task.ValidateCreateTaskInput(task.CreateTaskInput{
		Title:     "   ",
		Priority:  task.Priority("P9"),
		RiskLevel: task.RiskLevelLow,
	}); err == nil {
		t.Fatal("expected invalid input for blank title")
	}
}

func TestValidateCreateTaskInputRejectsInvalidPriority(t *testing.T) {
	if err := task.ValidateCreateTaskInput(task.CreateTaskInput{
		Title:     "Valid",
		Priority:  task.Priority("P9"),
		RiskLevel: task.RiskLevelLow,
	}); err == nil {
		t.Fatal("expected invalid priority to be rejected")
	}
}
