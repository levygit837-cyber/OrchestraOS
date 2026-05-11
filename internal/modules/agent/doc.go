// Package agent defines agent runtime interfaces and implementations.
//
// # Responsibility
// Abstracts the execution surface for AI agents. Provides a common Runtime
// interface and concrete implementations:
//   - fake: deterministic stub for testing
//   - gemini: Google Gemini inference integration
//
// # Key Types
//   - Runtime: interface for agent execution (Plan, Execute, etc.)
//   - GeminiPlanner: LLM-based task decomposition using Gemini
//   - FakeRuntime: test double with canned responses
//
// # Dependencies
//   - domain: Task, WorkUnit
//
// # Related Packages
//   - taskgraph/: uses GeminiPlanner for decomposition
//   - agentsession/: binds a runtime to a session
//
// CRITICAL RULES (violating these fails architecture tests):
//   - Every runtime must implement the Runtime interface completely.
//   - FakeRuntime responses must be deterministic for the same input.
//   - GeminiPlanner returns either a fully valid GraphPlan or an error — no partial plans.
//   - NEVER import internal/modules/* or internal/core/orchestration.
//   - NEVER mutate tasks, work_units, runs, or agent_sessions tables directly.
//
// For full contracts, invariants, and boundary rules:
//   READ: README.md  → purpose, dependencies, file map
//   READ: CONTRACTS.md → invariants, execution rules, boundary rules
//
// Quick code reference: see ModuleContract in contract.go
package agent
