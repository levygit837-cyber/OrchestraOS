package workunit

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// Aliases to shared domain types per ADR-0030.

type Status = domain.WorkUnitStatus
type WorkUnit = domain.WorkUnit

const (
	StatusCreated         = domain.WorkUnitStatusCreated
	StatusPlanned         = domain.WorkUnitStatusPlanned
	StatusScheduled       = domain.WorkUnitStatusScheduled
	StatusBlocked         = domain.WorkUnitStatusBlocked
	StatusRunning         = domain.WorkUnitStatusRunning
	StatusWaitingApproval = domain.WorkUnitStatusWaitingApproval
	StatusPaused          = domain.WorkUnitStatusPaused
	StatusValidating      = domain.WorkUnitStatusValidating
	StatusCompleted       = domain.WorkUnitStatusCompleted
	StatusFailed          = domain.WorkUnitStatusFailed
	StatusCancelled       = domain.WorkUnitStatusCancelled
)
