package domain

import "time"

// ============================================================================
// Review Domain
// ============================================================================

type ReviewStatus string

const (
	ReviewStatusPending          ReviewStatus = "pending"
	ReviewStatusInProgress       ReviewStatus = "in_progress"
	ReviewStatusApproved         ReviewStatus = "approved"
	ReviewStatusChangesRequested ReviewStatus = "changes_requested"
	ReviewStatusNeedsDiscussion  ReviewStatus = "needs_discussion"
)

type ReviewValidationGate string

const (
	ReviewValidationGateHard   ReviewValidationGate = "hard"
	ReviewValidationGateSoft   ReviewValidationGate = "soft"
	ReviewValidationGatePolicy ReviewValidationGate = "policy"
)

type Review struct {
	ID              string               `json:"id"`
	RunID           *string              `json:"run_id,omitempty"`
	WorkUnitID      *string              `json:"work_unit_id,omitempty"`
	TaskID          *string              `json:"task_id,omitempty"`
	AgentSessionID  *string              `json:"agent_session_id,omitempty"`
	ReviewerAgentID *string              `json:"reviewer_agent_id,omitempty"`
	GateType        ReviewValidationGate `json:"gate_type"`
	Status          ReviewStatus         `json:"status"`
	VerdictReason   string               `json:"verdict_reason,omitempty"`
	EvidenceRefs    []string             `json:"evidence_refs,omitempty"`
	CriteriaChecked []CriteriaChecked    `json:"criteria_checked,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	CompletedAt     *time.Time           `json:"completed_at,omitempty"`
}

type CriteriaChecked struct {
	Criterion string `json:"criterion"`
	Passed    bool   `json:"passed"`
	Reason    string `json:"reason,omitempty"`
}
