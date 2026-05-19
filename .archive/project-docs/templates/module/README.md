# Module: {{MODULE_NAME}}

## Purpose

[TODO: describe what this module is responsible for]

This module is responsible for:
- [TODO: list responsibilities]

This module DOES NOT:
- [TODO: list non-responsibilities]

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- [TODO: list 2-3 critical invariants]

State Flow:
```
[TODO: add state machine diagram if applicable]
```

---

## File Map

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules (global, type, specific)
- `models.go` → domain types (structs, enums, constants)
- `events.go` → event-type mapping for status transitions
- `queries.go` → SQL constants
- `repository.go` → pure CRUD, no business logic
- `service.go` → main domain logic, transactions, event emission
- `validation.go` → input validation

### Decomposed Files (service.go > 300 lines)
- [TODO: list if applicable, e.g. `service_create.go`]

### Optional Files
- `fetch.go` → RequireByID exported for DI by other modules

---

## Allowed Dependencies

- `internal/core/*`: apperrors, db, validation, event, statemachine, transition, serialization, eventstore
- `internal/domain`: ONLY EventEnvelope and generic types (never entity structs)

Forbidden:
- `internal/modules/*` (direct imports)
- `internal/core/*` packages not listed above
- Inline SQL outside `queries.go`
- Direct mutation of other module's tables

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors — keep changes minimal and localized.
5. State transitions MUST use `core/statemachine.CanTransition`.
6. Every mutation MUST emit an event via `core/transition` helpers.
7. SQL belongs only in `queries.go`.
