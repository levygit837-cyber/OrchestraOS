package review

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// Aliases to shared domain types per ADR-0030.

type Review = domain.Review
type Status = domain.ReviewStatus
type ValidationGate = domain.ReviewValidationGate
type CriteriaChecked = domain.CriteriaChecked

const (
	StatusPending          = domain.ReviewStatusPending
	StatusInProgress       = domain.ReviewStatusInProgress
	StatusApproved         = domain.ReviewStatusApproved
	StatusChangesRequested = domain.ReviewStatusChangesRequested
	StatusNeedsDiscussion  = domain.ReviewStatusNeedsDiscussion

	GateHard   = domain.ReviewValidationGateHard
	GateSoft   = domain.ReviewValidationGateSoft
	GatePolicy = domain.ReviewValidationGatePolicy
)

// Local types (not shared across modules).

type Decision = Status
