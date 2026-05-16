package task

import (
	"time"
)

type Status string

const (
	StatusCreated          Status = "created"
	StatusTriaged          Status = "triaged"
	StatusPlanned          Status = "planned"
	StatusScheduled        Status = "scheduled"
	StatusSandboxPreparing Status = "sandbox_preparing"
	StatusRunning          Status = "running"
	StatusWaitingApproval  Status = "waiting_approval"
	StatusPaused           Status = "paused"
	StatusValidating       Status = "validating"
	StatusCompleted        Status = "completed"
	StatusFailed           Status = "failed"
	StatusCancelled        Status = "cancelled"
)

type Priority string

const (
	PriorityP0 Priority = "P0"
	PriorityP1 Priority = "P1"
	PriorityP2 Priority = "P2"
	PriorityP3 Priority = "P3"
)

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

type Task struct {
	ID                   string    `json:"id"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	Status               Status    `json:"status"`
	Priority             Priority  `json:"priority"`
	RiskLevel            RiskLevel `json:"risk_level"`
	CreatedFromMessageID string    `json:"created_from_message_id"`
	AcceptanceCriteria   []string  `json:"acceptance_criteria"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
