package agent

import (
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// This file now primarily uses domain.Agent for persistence.
// The local RuntimeType constants are kept for runtime interface compatibility.

type RuntimeType string

const (
	RuntimeTypeCodexCLI RuntimeType = "codex_cli"
	RuntimeTypeFake     RuntimeType = "fake"
	RuntimeTypeExternal RuntimeType = "external"
	RuntimeTypeGemini   RuntimeType = "gemini"
)

// AgentStatus represents the status of an agent in the database
type AgentStatus string

const (
	AgentStatusActive   AgentStatus = "active"
	AgentStatusInactive AgentStatus = "inactive"
)

// ToDomainRuntimeType converts local RuntimeType to domain.AgentRuntimeType
func ToDomainRuntimeType(rt RuntimeType) domain.AgentRuntimeType {
	return domain.AgentRuntimeType(rt)
}

// FromDomainRuntimeType converts domain.AgentRuntimeType to local RuntimeType
func FromDomainRuntimeType(rt domain.AgentRuntimeType) RuntimeType {
	return RuntimeType(rt)
}
