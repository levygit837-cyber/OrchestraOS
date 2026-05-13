package orchestrator

import (
	"context"
	"database/sql"
	"fmt"
)

// RunTaskOptions configures how a task is executed.
type RunTaskOptions struct {
	RuntimeType     string // fake | gemini | codex_cli
	PlannerStrategy string // local_heuristic_v1 | llm_gemini_v1
	MaxSteps        int    // default: 10
	TimeoutSeconds  int    // default: 300
}

// RunTaskResult contains the outcome of an orchestrated task run.
type RunTaskResult struct {
	TaskID    string
	RunIDs    []string
	Status    string // completed | failed | partial
	ReviewIDs []string
}

// Service is the central orchestrator that automates task execution.
// TODO(ORCH-F05-R02-A01): Replace stub with full implementation.
type Service struct {
	db *sql.DB
}

// NewService creates a new OrchestratorService.
// TODO(ORCH-F05-R02-A01): Accept full Dependencies struct when implemented.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// RunTask executes a task end-to-end by coordinating all domain services.
// This is a STUB. The real implementation will:
//   - Decompose the task into a DAG of work units
//   - Execute each work unit sequentially via Runtime + Relay
//   - Manage agent lifecycle, prompts, reviews and triggers
//
// TODO(ORCH-F05-R02-A01): Implement full RunTask logic.
func (s *Service) RunTask(ctx context.Context, taskID string, options RunTaskOptions) (*RunTaskResult, error) {
	return nil, fmt.Errorf("OrchestratorService.RunTask is not yet implemented (pending ORCH-F05-R02-A01)")
}
