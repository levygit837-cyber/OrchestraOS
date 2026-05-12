package trigger

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.TriggerStatus
type Type = domain.TriggerType
type Anomaly = domain.AnomalyType
type Resolution = domain.ResolutionAction

const (
	StatusActive    = domain.TriggerStatusActive
	StatusTriggered = domain.TriggerStatusTriggered
	StatusResolved  = domain.TriggerStatusResolved
	StatusDismissed = domain.TriggerStatusDismissed
)

const (
	TypeThreshold        = domain.TriggerTypeThreshold
	TypeAnomaly          = domain.TriggerTypeAnomaly
	TypeHeartbeatTimeout = domain.TriggerTypeHeartbeatTimeout
	TypePolicy           = domain.TriggerTypePolicy
)

const (
	AnomalyStall         = domain.AnomalyTypeStall
	AnomalyLoop          = domain.AnomalyTypeLoop
	AnomalyDrift         = domain.AnomalyTypeDrift
	AnomalyPathViolation = domain.AnomalyTypePathViolation
	AnomalyTokenExceeded = domain.AnomalyTypeTokenExceeded
	AnomalyStepsExceeded = domain.AnomalyTypeStepsExceeded
	AnomalyTimeExceeded  = domain.AnomalyTypeTimeExceeded
)

const (
	ResolutionPause    = domain.ResolutionActionPause
	ResolutionCancel   = domain.ResolutionActionCancel
	ResolutionNotify   = domain.ResolutionActionNotify
	ResolutionEscalate = domain.ResolutionActionEscalate
)
