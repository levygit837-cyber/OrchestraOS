package runtime

import (
	"context"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Result holds the outcome of a work unit execution.
type Result struct {
	Status        domain.RunResult
	Output        string
	FailureReason string
}

// Runtime executes a single work unit and returns the result.
type Runtime interface {
	Execute(ctx context.Context, wu *domain.WorkUnit, task *domain.Task) (*Result, error)
}
