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

---

## File Map

- `doc.go` → package documentation and context briefing
- `models.go` → domain type aliases and constants
- `events.go` → event-type mappings for status transitions
- `queries.go` → SQL constants (SELECT / INSERT / UPDATE)
- `repository.go` → database reads and writes
- `service.go` → domain logic, state transitions, orchestration calls

---

## Allowed Dependencies

- `internal/core/*`
- `internal/domain`
- `internal/modules/event` (indirectly via orchestration)
- {{ALLOWED_MODULE}}

Forbidden:
- Direct imports of other modules' service logic
- Cross-module mutations outside `core/orchestration`

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors — keep changes minimal and localized.
5. Every mutation MUST emit an event via `core/orchestration` helpers.
6. SQL belongs only in `queries.go`.
