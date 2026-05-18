package prompt

import _ "embed"

// GLOBAL RULES (apply to ALL modules — do NOT remove):
//   1. NEVER import internal/modules/* for services, repositories, or business logic.
//      ALLOWED: import types (structs, enums) from another module ONLY for DI
//      interface return types. See ADR-0026 for full policy.
//   2. NEVER import internal/domain for entity structs or entity enums.
//      ALLOWED: EventEnvelope, EventPriority, checkpoint types, generic payloads.
//   3. NEVER write SQL outside queries.go.
//   4. NEVER call panic() — return apperrors.Error.
//   5. NEVER put business logic in repository.go.
//   6. ALWAYS emit a domain event on mutation.
//   7. ALWAYS validate inputs with core/validation on boundaries.
//
// MODULE-TYPE RULES (prompt is a support module — no state machine):
//   - Snapshots are immutable after creation.
//   - Composition must resolve all Requires and ConflictsWith relationships.
//
// MODULE-SPECIFIC RULES (prompt only):
//   - Prompt snapshots are deduplicated by composition_hash (UPSERT semantics).
//   - MaxAutonomyLevel = 2 is the highest allowed autonomy level for any prompt.
//   - All RequiredCategories must be present in a composed prompt.
//   - Toolset snapshots are immutable once created.
//   - NEVER call service methods from other modules.
//
// ALLOWED core/* imports:
//   - core/apperrors, core/db, core/validation, core/event
//   - core/serialization, core/transition
// FORBIDDEN core/* imports:
//   - core/statemachine (prompt does not manage lifecycle states)
//
// For full contracts, read CONTRACTS.md in this directory.

//go:embed README.md
//nolint:unused // embed placeholder for architecture test
var _readme string

//go:embed CONTRACTS.md
//nolint:unused // embed placeholder for architecture test
var _contracts string

var ModuleContract = struct {
	Name    string
	Purpose string
}{
	Name:    "prompt",
	Purpose: "Prompt composition, snapshot deduplication, and toolset management for agent runs",
}
