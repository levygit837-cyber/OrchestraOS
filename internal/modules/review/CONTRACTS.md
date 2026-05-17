# Contracts: review

## Invariants

- Verdicts (`approved`, `changes_requested`, `needs_discussion`) are immutable after submission.
- A review can only transition: `pending` → `in_progress` → verdict.
- Reviews must be created with a valid `gate_type` (`hard`, `soft`, `policy`).
- Every status change emits exactly one domain event.
- `criteria_checked` JSONB must be valid before persistence.
- Duplicate active reviews are prevented for the same `(work_unit_id, gate_type)`, `(run_id, gate_type)`, and `(task_id, gate_type)`.
- All read operations (`GetByID`, `ListByTask`, `ListPending`) accept and propagate `context.Context`.

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
- `Create` must check for existing active reviews across work_unit_id, run_id, and task_id before inserting.

---

## Boundary Rules

Allowed:
- Read and mutate the `reviews` table via `repository.go`.
- Append events via `core/transition` helpers.
- Read other aggregates via their repositories (not services) for validation context.

Forbidden:
- Direct mutation of `tasks`, `work_units`, `runs`, or `agent_sessions` tables.
- Calling service methods from other modules.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module orchestration belongs ONLY to:
- `internal/modules/orchestrator`

---

## Error Rules

| Code | When to Use |
|------|-------------|
| `CodeValidation` | Invalid gate types or verdicts |
| `CodeInvalidInput` | Semantically invalid input |
| `CodeNotFound` | Missing reviews |
| `CodeInvalidTransition` | State machine violation |
| `CodeConflict` | Duplicate active reviews |
| `CodePersistence` | Database errors |

---

## Persistence Rules

- All writes must go through `repository.go`.
- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD + row-scanning.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.
- All query methods accept `context.Context` and use `QueryContext` / `QueryRowContext`.

---

## File Decomposition

No service decomposition at this time. `service.go` is the single file for review lifecycle logic.

---

## Related ADRs

- ADR-0022: Vertical Slice Architecture
- ADR-0025: Module Standardization
