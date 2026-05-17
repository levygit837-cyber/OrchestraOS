// Package statemachine defines status-transition rules for aggregates.
//
// # Responsibility
// Encodes which status transitions are legal for each aggregate type
// (Task, WorkUnit, Run, AgentSession) and provides a CanTransition guard
// that returns an error for illegal transitions.
//
// # Key Types
//   - CanTransition: validates a proposed transition
//   - TransitionContext: carries evidence, justification and validation event
//   - AggregateTask, AggregateWorkUnit, AggregateRun, AggregateAgentSession:
//     constants identifying the aggregate kinds
//
// # Dependencies
//   - core/apperrors: error typing
//
// # Related Packages
//   - core/coordination/: calls CanTransition before executing transitions
//   - All modules/: may reference aggregate constants
package statemachine
