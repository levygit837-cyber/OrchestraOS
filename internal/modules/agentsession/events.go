package agentsession

func EventTypeForStatus(status Status) string {
	return "agent.session_" + string(status)
}
