// Package orchestration provides cross-domain state-transition commands.
//
// # Responsibility
// Orchestrates atomic status transitions across aggregates (Task, WorkUnit,
// Run, AgentSession) while enforcing state-machine rules and appending
// transition events. Also provides shared transition helpers used by all
// domain modules.
//
// # Key Types
//   - Commander: executes cross-domain transitions
//   - TransitionOptions: common options for any transition
//   - TransitionInput: payload builder input
//   - OperationResult[T]: generic result wrapper (value + event + duplicate flag)
//   - UpdateRunProjection: exported helper for updating run projections
//   - AppendServiceEvent, TransitionPayload, TransitionContext: shared helpers
//
// # Dependencies
//   - core/apperrors: error typing
//   - core/db: EnsureRowsAffected
//   - core/eventstore: event append
//   - core/serialization: payload marshalling
//   - core/statemachine: transition rules
//   - domain: all aggregate types and statuses
//
// # Related Packages
//   - All modules/: import this package for transitions and helpers
package orchestration
