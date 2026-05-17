// Package eventstore provides schema-validated event persistence.
//
// # Responsibility
// Stores and retrieves event envelopes with JSON-Schema validation,
// operational payload checks, and idempotent deduplication.
// The Store is the single entry-point for event persistence; modules
// should prefer the higher-level event.Service.
//
// # Key Types
//   - Store: validates and persists events
//   - Repository: SQL CRUD for events
//   - Validator: JSON-Schema validator loaded from embedded schemas
//   - NewStoreWithExecutor: factory accepting DBTX for transactional use
//
// # Dependencies
//   - core/apperrors: error wrapping
//   - core/db: DBTX interface
//   - core/statemachine: aggregate constants
//   - domain: EventEnvelope
//
// # Related Packages
//   - modules/event/: service layer that wraps Store with business rules
//   - core/transition/: appends transition events via Store
package eventstore
