package runtime

import (
	"fmt"
	"strings"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// BuildPrompt constructs a structured prompt from a WorkUnit and Task.
func BuildPrompt(wu *domain.WorkUnit, task *domain.Task) domain.Prompt {
	system := fmt.Sprintf(
		"You are an AI agent executing a work unit for task: %s\n"+
			"Task description: %s\n"+
			"Your objective: %s\n"+
			"You must satisfy the acceptance criteria listed below.",
		task.Title, task.Description, wu.Objective,
	)

	var user strings.Builder
	user.WriteString("Acceptance Criteria:\n")
	for i, ac := range wu.AcceptanceCriteria {
		fmt.Fprintf(&user, "%d. %s\n", i+1, ac)
	}

	if len(wu.OwnedPaths) > 0 {
		user.WriteString("\nFiles you may modify:\n")
		for _, p := range wu.OwnedPaths {
			fmt.Fprintf(&user, "- %s\n", p)
		}
	}

	if len(wu.ReadPaths) > 0 {
		user.WriteString("\nFiles you may read:\n")
		for _, p := range wu.ReadPaths {
			fmt.Fprintf(&user, "- %s\n", p)
		}
	}

	return domain.Prompt{
		SystemMessage: system,
		UserMessage:   user.String(),
		WorkUnitID:    wu.ID,
		TaskID:        task.ID,
	}
}
