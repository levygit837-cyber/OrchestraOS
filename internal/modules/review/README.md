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

- `doc.go` → package documentation and context briefing
- `models.go` → aliases for domain types and constants
- `queries.go` → SQL constants for reviews table
- `repository.go` → review CRUD (context-aware queries)
- `service.go` → review lifecycle service (Create, Start, SubmitVerdict)
- `events.go` → event type constants
- `validation.go` → input validation for reviews
- `contract.go` → module contract for LLM agents

---

## Allowed Dependencies

- `internal/core/*` (db, apperrors, validation, transition, serialization, statemachine)
- `internal/domain`

Forbidden:
- Direct imports of service logic from other modules.
- Cross-module mutations outside `core/orchestration`.

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. SQL belongs only in `queries.go`.
