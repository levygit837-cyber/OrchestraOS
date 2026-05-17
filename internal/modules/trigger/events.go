package trigger

func EventTypeForStatus(status Status) string {
	switch status {
	case StatusActive:
		return "trigger.created"
	case StatusTriggered:
		return "trigger.triggered"
	case StatusResolved:
		return "trigger.resolved"
	case StatusDismissed:
		return "trigger.dismissed"
	default:
		return "trigger.unknown"
	}
}
