package prompt

import (
	"fmt"
	"strings"
)

// ValidateFragment checks that a prompt fragment has all required fields.
func ValidateFragment(fragment *PromptFragment) error {
	if fragment == nil {
		return fmt.Errorf("fragment is nil")
	}
	if fragment.ID == "" {
		return fmt.Errorf("fragment ID is required")
	}
	if fragment.Version == "" {
		return fmt.Errorf("fragment version is required")
	}
	if fragment.Category == "" {
		return fmt.Errorf("fragment category is required")
	}
	if fragment.Kind == "" {
		return fmt.Errorf("fragment kind is required")
	}
	if fragment.Title == "" {
		return fmt.Errorf("fragment title is required")
	}
	if fragment.Body == "" {
		return fmt.Errorf("fragment body is required")
	}
	if fragment.AutonomyLevel > MaxAutonomyLevel {
		return fmt.Errorf("fragment autonomy level %d exceeds maximum %d", fragment.AutonomyLevel, MaxAutonomyLevel)
	}
	return nil
}

// ValidateToolsetInput checks that a toolset snapshot has required fields.
func ValidateToolsetInput(snapshot *ToolsetSnapshot) error {
	if snapshot == nil {
		return fmt.Errorf("toolset snapshot is nil")
	}
	if strings.TrimSpace(snapshot.RunID) == "" {
		return fmt.Errorf("toolset run_id is required")
	}
	if strings.TrimSpace(snapshot.AgentSessionID) == "" {
		return fmt.Errorf("toolset agent_session_id is required")
	}
	return nil
}
