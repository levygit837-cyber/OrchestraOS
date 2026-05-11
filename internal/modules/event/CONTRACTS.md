# Contracts: event

## Invariants

- Event append is idempotent: identical ID + identical content returns the existing envelope.
- Event ID collision with different content returns `CodeConflict`.
- Every envelope passes JSON-Schema validation before storage.
- Operational payload validation runs after schema validation.
- The service never mutates the envelope ID after receiving it.
- `AppendResult.Duplicate` is `true` iff the event already existed with identical content.

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

Events do not have a lifecycle state machine. They are immutable records.

---

## Execution Rules

- Always validate the envelope before storage.
- Never bypass schema validation.
- Never overwrite an existing event.
- Deduplication must compare full intent (type, payload, metadata) — not just ID.
- Return the existing envelope on duplicate without error.

---

## Boundary Rules

Allowed:
- Read and append to the `events` table via `core/eventstore.Store`.
- Validate envelopes against JSON Schema.

Forbidden:
- Direct mutation of events after storage.
- Imports of `internal/modules/*`.
- Business logic beyond validation and storage.
- Inline SQL — all persistence goes through `core/eventstore`.

---

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw database errors leaked outside the module.
- `CodeConflict` for idempotency violations (same ID, different content).
- `CodeValidation` for schema or operational payload failures.

---

## Persistence Rules

- All persistence goes through `core/eventstore.Store`.
- No SQL inside this module.
- No business logic inside repositories (none owned by this module).

---

## LLM Execution Rules

LLM executors MUST:

1. Read `README.md` first.
2. Read `CONTRACTS.md` before editing.
3. Modify only files related to the task.
4. Preserve all invariants.
5. Avoid speculative refactors.
6. Avoid introducing new abstractions unless required.
7. Keep implementations deterministic.
8. Preserve module boundaries.

---

## Forbidden Patterns

- Adding imports to `internal/modules/*`.
- Business logic beyond validation and storage.
- Mutable event envelopes.
- Inline SQL strings.
