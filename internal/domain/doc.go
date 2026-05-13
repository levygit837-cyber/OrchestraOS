// Package domain defines shared domain types and enums.
//
// # Responsibility
// Holds the canonical structs and constants used by multiple modules.
// This package has ZERO dependencies on other internal packages to avoid
// import cycles.
//
// # Key Types
//   - Task, TaskStatus, Priority, RiskLevel
//   - WorkUnit, WorkUnitStatus
//   - Run, RunStatus, RunResult
//   - AgentSession, AgentSessionStatus
//   - TaskGraph, TaskGraphStatus
//   - EventEnvelope, EventPayload
//   - PromptSnapshot, ToolsetSnapshot
//   - CheckpointTrigger, HeartbeatInput, CheckpointInput, AutoCheckpointInput,
//     CheckpointSuggestion: shared checkpoint/heartbeat types used by agentsession
//       and orchestration packages
//
// # Dependencies
//   - None (this is the innermost layer)
//
// # Related Packages
//   - All other internal packages: import domain for types
package domain
