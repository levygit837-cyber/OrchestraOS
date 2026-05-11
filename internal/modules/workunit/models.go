package workunit

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.WorkUnitStatus

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
