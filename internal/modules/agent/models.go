package agent

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// Aliases to shared domain types per ADR-0019.

type Agent = domain.Agent
type RuntimeType = domain.AgentRuntimeType

const (
	RuntimeTypeCodexCLI = domain.AgentRuntimeTypeCodexCLI
	RuntimeTypeFake     = domain.AgentRuntimeTypeFake
	RuntimeTypeExternal = domain.AgentRuntimeTypeExternal
	RuntimeTypeGemini   = domain.AgentRuntimeTypeGemini
)

type AgentStatus = domain.AgentStatus

const (
	AgentStatusActive   = domain.AgentStatusActive
	AgentStatusInactive = domain.AgentStatusInactive
)
