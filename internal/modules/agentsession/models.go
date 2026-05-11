package agentsession

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.AgentSessionStatus

const (
	StatusStarting        = domain.AgentSessionStatusStarting
	StatusRunning         = domain.AgentSessionStatusRunning
	StatusWaitingApproval = domain.AgentSessionStatusWaitingApproval
	StatusPaused          = domain.AgentSessionStatusPaused
	StatusStopping        = domain.AgentSessionStatusStopping
	StatusStopped         = domain.AgentSessionStatusStopped
	StatusDisconnected    = domain.AgentSessionStatusDisconnected
	StatusFailed          = domain.AgentSessionStatusFailed
)
