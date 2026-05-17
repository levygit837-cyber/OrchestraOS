package taskgraph

// EventTypeForStatus returns the event type string for a given status.
func EventTypeForStatus(status Status) string {
	return "taskgraph." + string(status)
}
