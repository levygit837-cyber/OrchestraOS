// Package orchestration provides shared transition helpers used by all
// domain modules.
//
// # Responsibility
// Provides common types, payload builders, and atomic update helpers for
// status transitions across aggregates (Task, WorkUnit, Run, AgentSession).
// Domain services in internal/modules/* use these primitives to implement
// their own transition logic.
//
// # Key Types
//   - TransitionInput: payload builder input for domain-service transitions
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
