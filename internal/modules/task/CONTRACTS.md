# Contracts: task

## Invariants

- Task status transitions are atomic and emit exactly one domain event.
- Terminal statuses (`completed`, `failed`, `cancelled`) cannot transition to any other status.
- `completed` transitions require evidence, validation event, or justification.
- Cancellation cascades to all dependent WorkUnits and Runs within the same transaction.
- Default `priority=P2` and `risk_level=low` when omitted at creation.
- `RequireByID` must return `apperrors.CodeNotFound` when a task does not exist.

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

Valid transitions:

| From | To |
|---|---|
| created | triaged, failed, cancelled |
| triaged | planned, failed, cancelled |
| planned | scheduled, failed, cancelled |
| scheduled | sandbox_preparing, running, paused, cancelled |
| sandbox_preparing | running, failed, cancelled |
| running | waiting_approval, paused, validating, failed, cancelled |
| waiting_approval | running, paused, failed, cancelled |
| paused | running, failed, cancelled |
| validating | completed, running, failed, cancelled |

Invalid transitions:
- Any transition from `completed`, `failed`, or `cancelled`.
- Any transition not listed in the valid table.
- `completed` without `EvidenceRefs`, `ValidationEventID`, or `Justification`.

Rules enforced by `core/statemachine.CanTransition`:
1. Terminal statuses are immutable.
2. `completed` requires completion evidence.
3. No transition is allowed unless explicitly listed above.

---

## Execution Rules

- Always validate state before mutation (`CanTransition`).
- Never bypass the state machine.
- Never update a task without emitting a domain event.
- State transitions must be atomic (single transaction).
- Cancellation must update the task projection AND cascade to WorkUnits and Runs in the same tx.
- Idempotency: duplicate event append returns the existing envelope without error.

---

## Boundary Rules

Allowed:
- Read and mutate the `tasks` table via `repository.go`.
- Append events via `core/orchestration` helpers.
- Call `core/statemachine.CanTransition` for validation.
- Use `run.NewRepository(tx)` and `workunit.NewRepository(tx)` for cascade reads/writes.

Forbidden:
- Direct mutation of `work_units` or `runs` tables through anything other than their repositories.
- Calling `run.Service` or `workunit.Service` methods.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module orchestration belongs ONLY to:
- `internal/core/orchestration`

---

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw database errors leaked outside the module.
- `CodeNotFound` for missing tasks.
- `CodeInvalidTransition` for illegal status changes.
- Validation errors must be deterministic and idempotent.

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
- Hidden side effects (every write emits an event).
- Cross-module mutations via service imports.
- Bypassing the state machine.
- Business logic inside repositories.
- Inline SQL strings.
