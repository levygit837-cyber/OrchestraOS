package trigger

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

func EventTypeForStatus(status domain.TriggerStatus) string {
	switch status {
	case domain.TriggerStatusActive:
		return "trigger.created"
	case domain.TriggerStatusTriggered:
		return "trigger.triggered"
	case domain.TriggerStatusResolved:
		return "trigger.resolved"
	case domain.TriggerStatusDismissed:
		return "trigger.dismissed"
	default:
		return "trigger.unknown"
	}
}
