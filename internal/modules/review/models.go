package review

import "time"

type Status string

const (
	StatusPending          Status = "pending"
	StatusInProgress       Status = "in_progress"
	StatusApproved         Status = "approved"
	StatusChangesRequested Status = "changes_requested"
	StatusNeedsDiscussion  Status = "needs_discussion"
)

type ValidationGate string

const (
	GateHard   ValidationGate = "hard"
	GateSoft   ValidationGate = "soft"
	GatePolicy ValidationGate = "policy"
)

type Decision = Status

type CriteriaChecked struct {
	Criterion string `json:"criterion"`
	Passed    bool   `json:"passed"`
	Reason    string `json:"reason,omitempty"`
}

type Review struct {
	ID              string            `json:"id"`
	RunID           *string           `json:"run_id,omitempty"`
	WorkUnitID      *string           `json:"work_unit_id,omitempty"`
	TaskID          *string           `json:"task_id,omitempty"`
	AgentSessionID  *string           `json:"agent_session_id,omitempty"`
	ReviewerAgentID *string           `json:"reviewer_agent_id,omitempty"`
	GateType        ValidationGate    `json:"gate_type"`
	Status          Status            `json:"status"`
	VerdictReason   string            `json:"verdict_reason,omitempty"`
	EvidenceRefs    []string          `json:"evidence_refs,omitempty"`
	CriteriaChecked []CriteriaChecked `json:"criteria_checked,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
}
