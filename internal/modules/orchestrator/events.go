package orchestrator

// EventTypeForStatus returns the event type string for orchestrator events.
// The orchestrator does not have a traditional state machine like domain modules,
// but emits lifecycle events for tasks and work units.
func EventTypeForStatus(status string) string {
	return "orchestrator." + status
}
