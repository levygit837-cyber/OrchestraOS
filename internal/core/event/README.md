# Package: event

## Purpose

This package is responsible for:
- Providing idempotent event appending with schema validation and deduplication.
- Serving as the single entry-point for domain event persistence consumed by all modules.
- Enforcing event envelope integrity (type, version, payload, metadata).

This package DOES NOT:
- Manage task, run, or work-unit lifecycle.
- Execute business logic beyond validation and storage.
- Compose prompts or manage agent sessions.

---

## Contract Summary

This package is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- Event append is idempotent: same ID + same content = no-op, returning the existing envelope.
- Event ID conflict with different content returns `CodeConflict`.
- Every envelope is validated against a JSON Schema before storage.
- Operational payload validation runs after schema validation.

State Flow:
```
Envelope → validation → schema check → idempotency check → persist → AppendResult
```

---

## File Map

- `doc.go` → package documentation and context briefing
- `models.go` → type aliases (`Envelope`, `Priority`)
- `service.go` → event append service with validation and deduplication

---

## Allowed Dependencies

- `internal/core/apperrors`
- `internal/core/db` (DBTX interface)
- `internal/core/eventstore` (Store, Validator)
- `internal/core/statemachine` (aggregate constants)
- `internal/domain`

Forbidden:
- Imports of any module under `internal/modules/*`.
- Business logic beyond validation and storage.
- Inline SQL — all persistence goes through `core/eventstore`.

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. This is a `core/` package — it must remain a leaf in the module dependency graph.
