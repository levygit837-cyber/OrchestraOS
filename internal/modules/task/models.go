package task

import (
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
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

// ToDomain converts a local Task to the legacy domain.Task for external consumers.
func ToDomain(t *Task) *domain.Task {
	if t == nil {
		return nil
	}
	return &domain.Task{
		ID:                   t.ID,
		Title:                t.Title,
		Description:          t.Description,
		Status:               domain.TaskStatus(t.Status),
		Priority:             domain.Priority(t.Priority),
		RiskLevel:            domain.RiskLevel(t.RiskLevel),
		CreatedFromMessageID: t.CreatedFromMessageID,
		AcceptanceCriteria:   t.AcceptanceCriteria,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
}

// FromDomain converts a legacy domain.Task to the local Task.
func FromDomain(t *domain.Task) *Task {
	if t == nil {
		return nil
	}
	return &Task{
		ID:                   t.ID,
		Title:                t.Title,
		Description:          t.Description,
		Status:               Status(t.Status),
		Priority:             Priority(t.Priority),
		RiskLevel:            RiskLevel(t.RiskLevel),
		CreatedFromMessageID: t.CreatedFromMessageID,
		AcceptanceCriteria:   t.AcceptanceCriteria,
		CreatedAt:            t.CreatedAt,
		UpdatedAt:            t.UpdatedAt,
	}
}
