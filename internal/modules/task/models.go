package task

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.TaskStatus
type Priority = domain.Priority
type RiskLevel = domain.RiskLevel

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
)

const (
	PriorityP0 = domain.PriorityP0
	PriorityP1 = domain.PriorityP1
	PriorityP2 = domain.PriorityP2
	PriorityP3 = domain.PriorityP3
)

const (
	RiskLevelLow      = domain.RiskLevelLow
	RiskLevelMedium   = domain.RiskLevelMedium
	RiskLevelHigh     = domain.RiskLevelHigh
	RiskLevelCritical = domain.RiskLevelCritical
)
