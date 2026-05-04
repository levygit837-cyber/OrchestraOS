package agent

import (
	"context"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
)

// Runtime defines the interface for agent runtimes
type Runtime interface {
	// Start starts the agent runtime with the given configuration
	Start(ctx context.Context, config RuntimeConfig) error

	// Stop stops the agent runtime
	Stop(ctx context.Context) error

	// SendEvent sends an event to the runtime
	SendEvent(ctx context.Context, event *domain.EventEnvelope) error

	// ReceiveEvent receives events from the runtime
	ReceiveEvent(ctx context.Context) (*domain.EventEnvelope, error)

	// Status returns the current runtime status
	Status() RuntimeStatus
}

// RuntimeConfig holds configuration for starting a runtime
type RuntimeConfig struct {
	RunID      string
	WorkUnitID string
	TaskID     string
	AgentID    string
	Prompt     string
	Toolset    []string
	OwnedPaths []string
	ReadPaths  []string
	MaxSteps   int
	Timeout    int // in seconds
}

// RuntimeStatus represents the current status of a runtime
type RuntimeStatus struct {
	State         string // starting, running, paused, stopping, stopped, failed
	CurrentStep   int
	LastHeartbeat int64 // unix timestamp
}

// RuntimeType represents the type of agent runtime
type RuntimeType string

const (
	RuntimeTypeCodexCLI RuntimeType = "codex_cli"
	RuntimeTypeFake     RuntimeType = "fake"
	RuntimeTypeExternal RuntimeType = "external"
)
