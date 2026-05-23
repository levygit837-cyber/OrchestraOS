package domain

import "time"

// ============================================================================
// Run Domain
// ============================================================================

type RunStatus string

const (
	RunStatusCreated         RunStatus = "created"
	RunStatusRunning         RunStatus = "running"
	RunStatusWaitingApproval RunStatus = "waiting_approval"
	RunStatusPaused          RunStatus = "paused"
	RunStatusValidating      RunStatus = "validating"
	RunStatusCompleted       RunStatus = "completed"
	RunStatusFailed          RunStatus = "failed"
	RunStatusCancelled       RunStatus = "cancelled"
)

type RunResult string

const (
	RunResultSucceeded RunResult = "succeeded"
	RunResultFailed    RunResult = "failed"
	RunResultCancelled RunResult = "cancelled"
)

type Run struct {
	ID            string     `json:"id"`
	TaskID        string     `json:"task_id"`
	WorkUnitID    string     `json:"work_unit_id"`
	Status        RunStatus  `json:"status"`
	Attempt       int        `json:"attempt"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	Result        *RunResult `json:"result,omitempty"`
	FailureReason *string    `json:"failure_reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
