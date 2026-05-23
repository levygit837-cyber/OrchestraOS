package agentsession

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// Aliases to shared domain types per ADR-0019.

type Status = domain.AgentSessionStatus
type AgentSession = domain.AgentSession

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
