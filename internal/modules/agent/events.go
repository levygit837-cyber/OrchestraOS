package agent

// EventTypeForStatus returns the event type for a given agent status
func EventTypeForStatus(status string) string {
	return "agent." + status
}

// EventCreated is the event type for agent creation
const EventCreated = "agent.created"
