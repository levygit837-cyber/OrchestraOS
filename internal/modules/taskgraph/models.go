package taskgraph

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.TaskGraphStatus

const (
	StatusActive     = domain.TaskGraphStatusActive
	StatusSuperseded = domain.TaskGraphStatusSuperseded
)
