// Package trigger implements configurable anomaly detection and threshold triggers.
//
// # Responsibility
// Manages Triggers — anomaly detectors, threshold checks, and heartbeat timeouts.
// Handles creation, evaluation of runs/sessions/workunits, resolution, and dismissal.
// Emits domain events for all state changes.
//
// # Key Types
//   - TriggerService: domain service for trigger operations
//   - CreateTriggerInput: input for creating a trigger
//   - ThresholdConfig: default and custom threshold configurations
//   - Detectors: StallDetector, LoopDetector, DriftDetector, PathViolationDetector,
//     TokenThresholdDetector, StepsThresholdDetector, TimeThresholdDetector
//
// # Dependencies
//   - core/db: transaction helpers
//   - core/eventstore: event storage
//   - core/transition: event emission helpers
//   - core/validation: input validation
//   - core/apperrors: error handling
//   - domain: Trigger, TriggerType, TriggerStatus, AnomalyType, ResolutionAction
//
// # Related Packages
//   - run/: triggers evaluate runs
//   - agentsession/: triggers evaluate sessions
//   - workunit/: triggers evaluate work units
//
// CRITICAL RULES (violating these fails architecture tests):
//   - Detectors are deterministic: same input always produces same output.
//   - Detectors have no side effects; they only analyze and return triggers.
//   - TriggerService persists triggers and emits events.
//   - NEVER write SQL outside queries.go.
//   - NEVER call run.Service or agentsession.Service methods directly.
//
// For full contracts, invariants, and boundary rules:
//   READ: README.md  → purpose, dependencies, file map
//   READ: CONTRACTS.md → invariants, boundary rules
//
// Quick code reference: see ModuleContract in contract.go
package trigger
