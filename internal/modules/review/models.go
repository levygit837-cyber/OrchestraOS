package review

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.ReviewStatus

const (
	StatusPending          = domain.ReviewStatusPending
	StatusInProgress       = domain.ReviewStatusInProgress
	StatusApproved         = domain.ReviewStatusApproved
	StatusChangesRequested = domain.ReviewStatusChangesRequested
	StatusNeedsDiscussion  = domain.ReviewStatusNeedsDiscussion
)

type Gate = domain.ValidationGate

const (
	GateHard   = domain.ValidationGateHard
	GateSoft   = domain.ValidationGateSoft
	GatePolicy = domain.ValidationGatePolicy
)
