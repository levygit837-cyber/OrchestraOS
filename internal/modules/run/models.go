package run

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.RunStatus
type Result = domain.RunResult

const (
	StatusCreated         = domain.RunStatusCreated
	StatusRunning         = domain.RunStatusRunning
	StatusWaitingApproval = domain.RunStatusWaitingApproval
	StatusPaused          = domain.RunStatusPaused
	StatusValidating      = domain.RunStatusValidating
	StatusCompleted       = domain.RunStatusCompleted
	StatusFailed          = domain.RunStatusFailed
	StatusCancelled       = domain.RunStatusCancelled
)

const (
	ResultSucceeded = domain.RunResultSucceeded
	ResultFailed    = domain.RunResultFailed
	ResultCancelled = domain.RunResultCancelled
)
