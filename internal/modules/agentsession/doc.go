// Package agentsession implements the agent-session domain.
//
// # Responsibility
// Manages AgentSessions — the binding between an agent runtime and a Run.
// Handles creation, heartbeat, manual/automatic checkpoint, timeout detection,
// and status transitions (including pausing the associated run on timeout).
//
// # Key Types
//   - AgentSessionService: domain service for session operations
//   - CreateAgentSessionInput: input for creating a session
//
// Shared types (defined in internal/domain for cross-package use):
//   - domain.HeartbeatInput: heartbeat payload
//   - domain.CheckpointInput: checkpoint payload with ledger and evidence
//   - domain.AutoCheckpointInput: parameters for automatic checkpoints
//   - domain.CheckpointTrigger: trigger reason enumeration
//   - domain.CheckpointSuggestion: result of checkpoint evaluation
//
// # Dependencies
//   - core/db: transaction helpers
//   - core/transition: TransitionInput, OperationResult
//   - core/statemachine: status transition rules
//   - core/validation: input validation
//   - domain: AgentSession, AgentSessionStatus, HeartbeatInput, CheckpointInput, etc.
//
// # Related Packages
//   - run/: sessions are bound to runs
//   - taskgraph/: planner strategies may influence session behaviour
//
// CRITICAL RULES (violating these fails architecture tests):
//   - Session status transitions are atomic and emit exactly one domain event.
//   - Terminal statuses (stopped, failed) are immutable.
//   - Checkpoint persists recoverable_state, ledger, and evidence_refs atomically.
//   - Timeout must pause the associated Run in the SAME transaction.
//   - NEVER call run.Service methods — use run.NewRepository(tx) for Run pause.
//   - NEVER write SQL outside queries.go.
//
// For full contracts, invariants, state machine, and boundary rules:
//
//	READ: README.md  → purpose, dependencies, file map
//	READ: CONTRACTS.md → invariants, state machine, boundary rules
//
// Quick code reference: see ModuleContract in contract.go
package agentsession
