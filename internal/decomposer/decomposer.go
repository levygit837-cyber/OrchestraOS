package decomposer

import (
	"context"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Strategy decomposes a task into work unit specs and a rationale.
// Implementations include agent-based (LLM) and heuristic strategies.
type Strategy interface {
	Decompose(ctx context.Context, req *domain.DecompositionRequest) (*domain.DecompositionResult, error)
	Name() string
}

// Decomposer is the main interface for task decomposition.
type Decomposer interface {
	Decompose(ctx context.Context, req *domain.DecompositionRequest) (*domain.DecompositionResult, error)
}
