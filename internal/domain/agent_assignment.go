package domain

import "time"

// ============================================================================
// Agent Assignment Domain
// ============================================================================

type AssignmentStatus string

const (
	AssignmentStatusActive   AssignmentStatus = "active"
	AssignmentStatusReplaced AssignmentStatus = "replaced"
	AssignmentStatusRemoved  AssignmentStatus = "removed"
)

type AgentAssignment struct {
	ID           string           `json:"id"`
	WorkUnitID   string           `json:"work_unit_id"`
	AgentProfile string           `json:"agent_profile"`
	Status       AssignmentStatus `json:"status"`
	Reason       string           `json:"reason"`
	AssignedAt   time.Time        `json:"assigned_at"`
	RemovedAt    *time.Time       `json:"removed_at,omitempty"`
}
