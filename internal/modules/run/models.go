package run

import "time"

type Status string

const (
	StatusCreated         Status = "created"
	StatusRunning         Status = "running"
	StatusWaitingApproval Status = "waiting_approval"
	StatusPaused          Status = "paused"
	StatusValidating      Status = "validating"
	StatusCompleted       Status = "completed"
	StatusFailed          Status = "failed"
	StatusCancelled       Status = "cancelled"
)

type Result string

const (
	ResultSucceeded Result = "succeeded"
	ResultFailed    Result = "failed"
	ResultCancelled Result = "cancelled"
)

type Run struct {
	ID            string     `json:"id"`
	TaskID        string     `json:"task_id"`
	WorkUnitID    string     `json:"work_unit_id"`
	Status        Status     `json:"status"`
	Attempt       int        `json:"attempt"`
	StartedAt     time.Time  `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	Result        *Result    `json:"result,omitempty"`
	FailureReason *string    `json:"failure_reason,omitempty"`
}
