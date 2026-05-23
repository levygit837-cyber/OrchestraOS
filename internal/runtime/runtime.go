package runtime

import (
	"context"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Config holds provider configuration for LLM runtimes.
type Config struct {
	Provider    string
	Model       string
	APIKey      string
	BaseURL     string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration
}

// Prompt represents the structured input sent to an LLM provider.
type Prompt struct {
	SystemMessage string
	UserMessage   string
	WorkUnitID    string
	TaskID        string
}

// Result holds the outcome of a work unit execution.
type Result struct {
	Status        domain.RunResult
	Output        string
	FailureReason string
	Provider      string
	Model         string
	TokensUsed    int
	Latency       time.Duration
}

// Runtime executes a single work unit and returns the result.
type Runtime interface {
	Execute(ctx context.Context, wu *domain.WorkUnit, task *domain.Task) (*Result, error)
}
