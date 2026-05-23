package domain

import (
	"context"
	"time"
)

// Runtime defines the execution contract for an LLM provider.
type Runtime interface {
	Execute(ctx context.Context, wu *WorkUnit, task *Task) (*RuntimeResult, error)
}

// StreamRuntime extends Runtime with streaming support.
type StreamRuntime interface {
	Runtime
	ExecuteStream(ctx context.Context, wu *WorkUnit, task *Task) (<-chan StreamChunk, <-chan error)
}

// RuntimeResult holds the outcome of a single Execute call.
type RuntimeResult struct {
	Status        RunResult
	Output        string
	FailureReason string
	Provider      string
	Model         string
	TokensUsed    int
	Latency       time.Duration
}

// StreamChunk represents a single streaming chunk emitted during ExecuteStream.
type StreamChunk struct {
	Delta         string
	ThinkingDelta string
	TokensUsed    int
	IsThinking    bool
	IsFinal       bool
	Provider      string
	Model         string
}

// Prompt holds the structured prompt sent to an LLM provider.
type Prompt struct {
	SystemMessage string
	UserMessage   string
	WorkUnitID    string
	TaskID        string
}
