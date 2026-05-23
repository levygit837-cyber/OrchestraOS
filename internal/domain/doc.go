// Package domain defines shared infrastructure types for events and checkpoints.
//
// # Responsibility
// Holds types genuinely shared by multiple modules: event envelope, event priority,
// checkpoint/heartbeat types, and generic event payloads.
// This package has ZERO dependencies on other internal packages to avoid
// import cycles.
//
// # Key Types
//   - EventEnvelope, EventPriority
//   - CheckpointTrigger, HeartbeatInput, CheckpointInput, AutoCheckpointInput,
//     CheckpointSuggestion: shared checkpoint/heartbeat types used by agentsession
//     and transition packages
//   - Generic event payloads: TaskGraphCreatedPayload, AgentLedgerUpdatedPayload,
//     AgentCheckpointReachedPayload, ArtifactCreatedPayload, ValidationCompletedPayload,
//     PromptSnapshotCreatedPayload, ToolsetSnapshotCreatedPayload
//
// # What does NOT belong here
// Entity structs (Task, WorkUnit, Run, Agent, AgentSession, etc.) and their
// status enums live in their respective vertical modules under internal/modules/.
// See ADR-0015 Section 4 (Pilar 1).
//
// # Dependencies
//   - None (this is the innermost layer)
//
// # Related Packages
//   - All other internal packages: import domain for EventEnvelope and checkpoint types
package domain
