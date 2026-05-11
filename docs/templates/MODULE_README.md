# Module: {{MODULE_NAME}}

## Purpose

This module is responsible for:
- {{RESPONSIBILITY}}
- {{RESPONSIBILITY}}
- {{RESPONSIBILITY}}

This module DOES NOT:
- {{NON_RESPONSIBILITY}}
- {{NON_RESPONSIBILITY}}
- {{NON_RESPONSIBILITY}}

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- {{INVARIANT}}
- {{INVARIANT}}
- {{INVARIANT}}

State Flow:
{{STATE_MACHINE}}

---

## File Map

- `doc.go` → package documentation and context briefing
- `models.go` → domain type aliases and constants
- `events.go` → event-type mappings for status transitions
- `fetch.go` → exported read helpers used by other modules (DI)
- `queries.go` → SQL constants (SELECT / INSERT / UPDATE)
- `repository.go` → database reads and writes
- `service.go` → domain logic, state transitions, orchestration calls
- `validation_test.go` → invariant and rule tests

---

## Allowed Dependencies

- `internal/core/apperrors`
- `internal/core/db`
- `internal/core/eventstore`
- `internal/core/orchestration`
- `internal/core/serialization`
- `internal/core/statemachine`
- `internal/core/validation`
- `internal/domain`
- `internal/core/event`
- {{ALLOWED_MODULE}}

Forbidden:
- Direct imports of other modules' service logic
- Cross-module mutations outside `core/orchestration`
- `internal/services` (removed)
- `internal/repository` (removed)

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors — keep changes minimal and localized.
5. State transitions MUST use `core/statemachine.CanTransition`.
6. Every mutation MUST emit an event via `core/orchestration` helpers.
7. SQL belongs only in `queries.go` — never inline in services.
