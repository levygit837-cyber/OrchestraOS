// Package run implements the run execution domain.
//
// # Responsibility
// Manages Runs — executions of WorkUnits by agent sessions.
// Handles creation, retry policies, status transitions, result projection,
// and cascading updates to related WorkUnits.
//
// # Key Types
//   - RunService: domain service for run operations
//   - CreateRunInput: input for creating a run
//   - TaskReader: DI interface to avoid cyclic imports from task/
//   - WorkUnitReader: DI interface to avoid cyclic imports from workunit/
//   - ResultForStatus: exported helper that maps a terminal status to a RunResult
//
// # Dependencies
//   - core/db: transaction helpers
//   - core/eventstore: event storage
//   - core/orchestration: TransitionInput, OperationResult, state transitions
//   - core/statemachine: status transition rules
//   - core/validation: input validation
//   - domain: Run, RunStatus, RunResult
//
// # Related Packages
//   - task/: runs belong to tasks
//   - workunit/: runs execute work units
//   - agentsession/: sessions execute runs
//
// CRITICAL RULES (violating these fails architecture tests):
//   - Run status transitions are atomic and emit exactly one domain event.
//   - Terminal statuses (completed, failed, cancelled) are immutable.
//   - Retry policy must respect max_attempts and exponential backoff.
//   - NEVER call task.Service or workunit.Service methods.
//   - NEVER write SQL outside queries.go.
//
// For full contracts, invariants, state machine, and boundary rules:
//   READ: README.md  → purpose, dependencies, file map
//   READ: CONTRACTS.md → invariants, state machine, boundary rules
//
// Quick code reference: see ModuleContract in contract.go
package run
