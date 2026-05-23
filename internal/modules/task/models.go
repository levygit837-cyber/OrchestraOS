package task

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// Aliases to shared domain types per ADR-0019.
// These allow existing code in the task module to continue using task.Task,
// task.Status, etc., while other modules import domain directly.

type Status = domain.TaskStatus
type Priority = domain.TaskPriority
type RiskLevel = domain.TaskRiskLevel
type Task = domain.Task

const (
	StatusCreated          = domain.TaskStatusCreated
	StatusTriaged          = domain.TaskStatusTriaged
	StatusPlanned          = domain.TaskStatusPlanned
	StatusScheduled        = domain.TaskStatusScheduled
	StatusSandboxPreparing = domain.TaskStatusSandboxPreparing
	StatusRunning          = domain.TaskStatusRunning
	StatusWaitingApproval  = domain.TaskStatusWaitingApproval
	StatusPaused           = domain.TaskStatusPaused
	StatusValidating       = domain.TaskStatusValidating
	StatusCompleted        = domain.TaskStatusCompleted
	StatusFailed           = domain.TaskStatusFailed
	StatusCancelled        = domain.TaskStatusCancelled

	PriorityP0 = domain.TaskPriorityP0
	PriorityP1 = domain.TaskPriorityP1
	PriorityP2 = domain.TaskPriorityP2
	PriorityP3 = domain.TaskPriorityP3

	RiskLevelLow      = domain.TaskRiskLevelLow
	RiskLevelMedium   = domain.TaskRiskLevelMedium
	RiskLevelHigh     = domain.TaskRiskLevelHigh
	RiskLevelCritical = domain.TaskRiskLevelCritical
)
