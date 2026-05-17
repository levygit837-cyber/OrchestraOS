// Package workunit implements the work-unit domain.
//
// # Responsibility
// Manages WorkUnits — the smallest assignable unit of work within a TaskGraph.
// Handles creation, status transitions, dependency validation, owned-path
// availability checks, and manual task-graph activation.
//
// # Key Types
//   - WorkUnitService: domain service for work-unit operations
//   - CreateWorkUnitInput: input for creating a work unit
//   - TaskReader: DI interface to avoid cyclic imports from task/
//   - TaskGraphManager: DI interface to avoid cyclic imports from taskgraph/
//   - ValidateWorkUnitDependencies, ValidateDependenciesCompleted,
//     ValidateOwnedPathAvailability: exported validation helpers used by run/
//
// # Dependencies
//   - core/db: transaction helpers
//   - core/coordination: TransitionInput, OperationResult
//   - core/statemachine: status transition rules
//   - core/validation: input validation
//   - domain: WorkUnit, WorkUnitStatus
//
// # Related Packages
//   - task/: tasks contain work units
//   - run/: runs execute work units
//   - taskgraph/: graphs organize work units into DAGs
//
// CRITICAL RULES (violating these fails architecture tests):
//   - WorkUnit status transitions are atomic and emit exactly one domain event.
//   - Terminal statuses (completed, failed, cancelled) are immutable.
//   - ValidateWorkUnitDependencies ensures the dependency graph is ACYCLIC.
//   - ValidateOwnedPathAvailability prevents path collisions between active work units.
//   - NEVER call task.Service or run.Service methods.
//   - NEVER write SQL outside queries.go.
//
// For full contracts, invariants, state machine, and boundary rules:
//
//	READ: README.md  → purpose, dependencies, file map
//	READ: CONTRACTS.md → invariants, state machine, boundary rules
//
// Quick code reference: see ModuleContract in contract.go
package workunit
