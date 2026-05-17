# Module: review

## Purpose

This module is responsible for:
- Managing reviews and validation gates for work units, runs, and tasks.
- Enforcing structured verdicts (approved, changes_requested, needs_discussion).
- Ensuring verdict immutability after submission.
- Emitting domain events for review lifecycle transitions.
- Preventing duplicate active reviews for the same work unit/run/task + gate combination.

This module DOES NOT:
- Manage task or run lifecycle directly.
- Execute agent code.
- Modify work units or tasks (only reads for validation).

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- Verdicts are immutable once submitted.
- Reviews must be created with a valid gate_type.
- Every status change emits exactly one domain event.
- Reviews can only transition from pending → in_progress → verdict.
- Duplicate active reviews are blocked for work_unit, run, and task scopes.
- All list/read operations accept and propagate context.Context.

---

## File Map

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules
- `models.go` → aliases for domain types and constants
- `events.go` → event type constants
- `queries.go` → SQL constants for reviews table
- `repository.go` → review CRUD, no business logic
- `service.go` → review lifecycle service (Create, Start, SubmitVerdict)
- `validation.go` → input validation for reviews

### Optional Files
- None at this time.

---

## Allowed Dependencies

- `internal/core/apperrors`, `core/db`, `core/validation`, `core/event`
- `internal/core/statemachine`, `core/transition`, `core/serialization`
- `internal/domain`: ONLY `EventEnvelope` and generic types (never entity structs)

- DI interface types: `run.Run`, `workunit.WorkUnit`, `task.Task`, `agentsession.AgentSession`
  — see ADR-0026: types may be imported ONLY for DI interface return types.

Forbidden:
- `internal/modules/*` services, repositories, or business logic imports
- Direct imports of service logic from other modules.

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. State transitions MUST use `core/statemachine.CanTransition`.
6. Every mutation MUST emit an event.
7. SQL belongs only in `queries.go`.
