package prompt

// EventTypeForAction returns the event type string for a given prompt action.
// The prompt module does not have a traditional lifecycle state machine,
// but emits events for snapshot creation and usage tracking.
func EventTypeForAction(action string) string {
	switch action {
	case "snapshot_created":
		return "prompt.snapshot_created"
	case "toolset_snapshot_created":
		return "toolset.snapshot_created"
	default:
		return "prompt." + action
	}
}
