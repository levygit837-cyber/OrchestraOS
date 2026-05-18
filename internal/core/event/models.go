package event

// Aliases para os tipos canônicos em internal/domain/. internal/domain/ é a fonte de verdade.
import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Envelope = domain.EventEnvelope
type Priority = domain.EventPriority

const (
	PriorityInterrupt    = domain.EventPriorityInterrupt
	PriorityCheckpoint   = domain.EventPriorityCheckpoint
	PriorityNotification = domain.EventPriorityNotification
	PriorityBackground   = domain.EventPriorityBackground
)
