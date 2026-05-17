// Package task implements the task lifecycle domain.
//
// # Responsibility
// Manages the full lifecycle of a Task from creation through completion,
// including triage, planning, scheduling, execution, validation and cancellation.
// When a task is cancelled, it cascades the cancellation to dependent WorkUnits
// and Runs using local repository calls (not service imports) to preserve
// transactional atomicity.
//
// # Key Types
//   - TaskService: domain service for task operations
//   - CreateTaskInput: input for creating a new task
//   - RequireByID: exported repository factory used as TaskReader by other modules
//
// # Dependencies
//   - core/db: transaction helpers
//   - core/transition: TransitionInput, OperationResult, state transitions
//   - core/statemachine: status transition rules
//   - core/validation: input validation
//   - domain: EventEnvelope (event types only)
//
// # Related Packages
//   - workunit/: tasks create work units via decomposition
//   - run/: tasks own runs
//   - taskgraph/: decomposes tasks into directed graphs of work units
//
// CRITICAL RULES (violating these fails architecture tests):
//   - Task status transitions are atomic and emit exactly one domain event.
//   - Terminal statuses (completed, failed, cancelled) are immutable.
//   - Cancellation cascades to dependent WorkUnits and Runs in the SAME transaction.
//   - NEVER call run.Service or workunit.Service methods — use repositories only.
//   - NEVER write SQL outside queries.go.
//
// For full contracts, invariants, state machine, and boundary rules:
//
//	READ: README.md  → purpose, dependencies, file map
//	READ: CONTRACTS.md → invariants, state machine, boundary rules
//
// Quick code reference: see ModuleContract in contract.go
package task
