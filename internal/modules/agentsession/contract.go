package agentsession

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
// MODULE-TYPE RULES (apply to ALL domain modules):
//   1. Status transitions are atomic and emit exactly one domain event.
//   2. Terminal statuses are immutable.
//   3. ALWAYS call core/statemachine.CanTransition before mutating state.
//   4. NEVER call another module's Service methods — use DI interfaces.
//
// MODULE-SPECIFIC RULES (agentsession only):
//   - Checkpoint persists recoverable_state, ledger, and evidence_refs atomically.
//   - Timeout must pause the associated Run in the SAME transaction.
//   - NEVER call run.Service methods — use run.NewRepository(tx) for Run pause.
//   - Heartbeat updates last_heartbeat_at without changing status.
//
// ALLOWED core/* imports:
//   - core/apperrors, core/db, core/validation, core/event
//   - core/statemachine, core/transition, core/serialization
// FORBIDDEN core/* imports:
//   - core/coordination (reserved for orchestrator module only)
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
	Name:    "agentsession",
	Purpose: "Agent session lifecycle, heartbeat, checkpoint, and timeout detection",
}
