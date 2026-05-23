package run

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// Aliases to shared domain types per ADR-0019.

type Status = domain.RunStatus
type Result = domain.RunResult
type Run = domain.Run

const (
	StatusCreated         = domain.RunStatusCreated
	StatusRunning         = domain.RunStatusRunning
	StatusWaitingApproval = domain.RunStatusWaitingApproval
	StatusPaused          = domain.RunStatusPaused
	StatusValidating      = domain.RunStatusValidating
	StatusCompleted       = domain.RunStatusCompleted
	StatusFailed          = domain.RunStatusFailed
	StatusCancelled       = domain.RunStatusCancelled

	ResultSucceeded = domain.RunResultSucceeded
	ResultFailed    = domain.RunResultFailed
	ResultCancelled = domain.RunResultCancelled
)
