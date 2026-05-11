package event

// Aliases temporários para domain. Serão convertidos em definições próprias na Onda 10.
import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Envelope = domain.EventEnvelope
type Priority = domain.EventPriority

const (
	PriorityInterrupt    = domain.EventPriorityInterrupt
	PriorityCheckpoint   = domain.EventPriorityCheckpoint
	PriorityNotification = domain.EventPriorityNotification
	PriorityBackground   = domain.EventPriorityBackground
)
