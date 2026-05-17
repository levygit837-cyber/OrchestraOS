// Package taskgraph implements task decomposition into directed acyclic graphs (DAGs).
//
// # Responsibility
// Decomposes a Task into a TaskGraph containing WorkUnits, nodes and edges.
// Supports two planner strategies:
//   - local_heuristic_v1: rule-based decomposition without external calls
//   - gemini: LLM-based decomposition using Google Gemini
//
// Also manages graph versioning, activation, and supersession.
//
// # Key Types
//   - TaskGraphService: domain service for graph lifecycle
//   - DecomposeTaskGraphInput: input for decomposition
//   - TaskGraphDecomposeResult: output with graph, work units and event
//   - GraphPlan: intermediate plan produced by a Planner
//   - GeminiPlanner: concrete LLM planner implementation
//   - PlannerPrompt, ValidateGraphPlan: exported helpers for consumers
//
// # Dependencies
//   - core/db: transaction helpers
//   - core/eventstore: event storage
//   - core/coordination: TransitionInput, OperationResult
//   - core/validation: input validation
//   - domain: TaskGraph, TaskGraphStatus
//
// # Related Packages
//   - task/: graphs are created from tasks
//   - workunit/: graphs contain work units
//   - agent/: GeminiPlanner uses agent runtime interfaces
//
// CRITICAL RULES (violating these fails architecture tests):
//   - A Task can have at most ONE active TaskGraph at any time.
//   - Graph plans MUST be validated before persistence (cycle detection, node count).
//   - NEVER call task.Service or workunit.Service directly — use DI interfaces only.
//   - NEVER put business logic in repository.go — pure CRUD only.
//   - NEVER write SQL outside queries.go.
//
// For full contracts, invariants, and boundary rules:
//
//	READ: README.md  → purpose, dependencies, file map
//	READ: CONTRACTS.md → invariants, state machine, boundary rules
//
// Quick code reference: see ModuleContract in contract.go
package taskgraph
