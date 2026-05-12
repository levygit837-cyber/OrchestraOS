package transition

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// TransitionInput carries metadata for state-machine transitions.
type TransitionInput struct {
	EventID           string
	AgentID           string
	Runtime           string
	EvidenceRefs      []string
	ValidationEventID string
	Justification     string
	FailureReason     string
	Extra             map[string]interface{}
}

// OperationResult pairs a domain value with the event that produced it.
type OperationResult[T any] struct {
	Value     T
	Event     *domain.EventEnvelope
	Duplicate bool
}
