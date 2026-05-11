// Package event implements the event-append service.
//
// # Responsibility
// Provides idempotent event appending with schema validation and deduplication.
// Consumers append envelopes; the service validates, checks for duplicates,
// and persists through the core/eventstore.
//
// # Key Types
//   - Service: event-append service
//   - AppendResult: result containing the persisted (or duplicate) envelope
//   - Envelope: event envelope with type, version, payload and metadata
//
// # Dependencies
//   - core/db: DBTX interface
//   - core/eventstore: schema validation and persistence
//   - core/statemachine: aggregate type constants
//
// # Related Packages
//   - core/eventstore/: lower-level storage with JSON-Schema validation
//   - All modules/: every domain service appends events via this package
//
// CRITICAL RULES (violating these fails architecture tests):
//   - Event append is idempotent: same ID + same content = no-op, returning existing envelope.
//   - Event ID collision with different content returns CodeConflict.
//   - Every envelope passes JSON-Schema validation before storage.
//   - NEVER import internal/modules/* (event is a leaf dependency).
//   - NEVER write inline SQL — all persistence goes through core/eventstore.
//
// For full contracts, invariants, and boundary rules:
//   READ: README.md  → purpose, dependencies, file map
//   READ: CONTRACTS.md → invariants, execution rules, boundary rules
//
// Quick code reference: see ModuleContract in contract.go
package event
