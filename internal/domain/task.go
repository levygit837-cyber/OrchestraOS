package domain

import "time"

// ============================================================================
// Task Domain
// ============================================================================

type TaskStatus string

const (
	TaskStatusCreated          TaskStatus = "created"
	TaskStatusTriaged          TaskStatus = "triaged"
	TaskStatusPlanned          TaskStatus = "planned"
	TaskStatusScheduled        TaskStatus = "scheduled"
	TaskStatusSandboxPreparing TaskStatus = "sandbox_preparing"
	TaskStatusRunning          TaskStatus = "running"
	TaskStatusWaitingApproval  TaskStatus = "waiting_approval"
	TaskStatusPaused           TaskStatus = "paused"
	TaskStatusValidating       TaskStatus = "validating"
	TaskStatusCompleted        TaskStatus = "completed"
	TaskStatusFailed           TaskStatus = "failed"
	TaskStatusCancelled        TaskStatus = "cancelled"
)

type TaskPriority string

const (
	TaskPriorityP0 TaskPriority = "P0"
	TaskPriorityP1 TaskPriority = "P1"
	TaskPriorityP2 TaskPriority = "P2"
	TaskPriorityP3 TaskPriority = "P3"
)

type TaskRiskLevel string

const (
	TaskRiskLevelLow      TaskRiskLevel = "low"
	TaskRiskLevelMedium   TaskRiskLevel = "medium"
	TaskRiskLevelHigh     TaskRiskLevel = "high"
	TaskRiskLevelCritical TaskRiskLevel = "critical"
)

type Task struct {
	ID                   string        `json:"id"`
	Title                string        `json:"title"`
	Description          string        `json:"description"`
	Status               TaskStatus    `json:"status"`
	Priority             TaskPriority  `json:"priority"`
	RiskLevel            TaskRiskLevel `json:"risk_level"`
	CreatedFromMessageID string        `json:"created_from_message_id"`
	AcceptanceCriteria   []string      `json:"acceptance_criteria"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}
