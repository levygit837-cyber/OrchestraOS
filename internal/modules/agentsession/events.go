package agentsession

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

func EventTypeForStatus(status domain.AgentSessionStatus) string {
	return "agent.session_" + string(status)
}
