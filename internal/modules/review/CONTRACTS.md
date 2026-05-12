# Contracts: review

## Invariants

- Verdicts (`approved`, `changes_requested`, `needs_discussion`) are immutable after submission.
- A review can only transition: `pending` → `in_progress` → verdict.
- Reviews must be created with a valid `gate_type` (`hard`, `soft`, `policy`).
- Every status change emits exactly one domain event.
- `criteria_checked` JSONB must be valid before persistence.

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

Valid transitions:

| From | To |
|---|---|
| pending | in_progress |
| pending | approved, changes_requested, needs_discussion |
| in_progress | approved, changes_requested, needs_discussion |

Invalid transitions:
- Any verdict → any status.
- `in_progress` → `pending`.
- Any status → invalid status.

---

## Execution Rules

- Always validate input UUIDs and gate_type before persistence.
- Verdict submission requires the review to be `pending` or `in_progress`.
- Verdict submission updates `completed_at` timestamp.
- Idempotency: duplicate event append returns the existing envelope without error.

---

## Boundary Rules

Allowed:
- Read and mutate the `reviews` table via `repository.go`.
- Append events via `core/orchestration` helpers.
- Read other aggregates via their repositories (not services) for validation context.

Forbidden:
- Direct mutation of `tasks`, `work_units`, `runs`, or `agent_sessions` tables.
- Calling service methods from other modules.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module orchestration belongs ONLY to:
- `internal/core/orchestration`

---

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw database errors leaked outside the module.
- `CodeValidation` for invalid gate types or verdicts.
- `CodeInvalidTransition` for illegal status changes.
- `CodeNotFound` for missing reviews.

---

## Persistence Rules

- All writes must go through `repository.go`.
- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD + row-scanning.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.

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

- Shared helpers inside the module (move to `core/` if reusable).
- Hidden side effects.
- Cross-module mutations via service imports.
- Business logic inside repositories.
- Inline SQL strings.
