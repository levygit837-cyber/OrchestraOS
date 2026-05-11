# Contracts: run

## Invariants

- Run status transitions are atomic and emit exactly one domain event.
- Terminal statuses (`completed`, `failed`, `cancelled`) cannot transition to any other status.
- `completed` transitions require evidence, validation event, or justification.
- `ResultForStatus` is deterministic: `completed`→`succeeded`, `failed`→`failed`, `cancelled`→`cancelled`.
- A Run can only be created for a WorkUnit that is not in a terminal status.
- Retry policy must respect `max_attempts` and exponential backoff.
- Transitioning a Run to terminal updates the related WorkUnit projection atomically.

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

Valid transitions:

| From | To |
|---|---|
| created | running, failed, cancelled |
| running | waiting_approval, paused, validating, failed, cancelled |
| waiting_approval | running, paused, failed, cancelled |
| paused | running, failed, cancelled |
| validating | completed, running, failed, cancelled |

Invalid transitions:
- Any transition from `completed`, `failed`, or `cancelled`.
- Any transition not listed in the valid table.
- `completed` without completion evidence.

---

## Execution Rules

- Always validate state before mutation (`CanTransition`).
- Never bypass the state machine.
- Never update a run without emitting a domain event.
- State transitions must be atomic (single transaction).
- Terminal transitions must update the WorkUnit projection in the same tx.
- Idempotency: duplicate event append returns the existing envelope without error.

---

## Boundary Rules

Allowed:
- Read and mutate the `runs` table via `repository.go`.
- Append events via `core/orchestration` helpers.
- Call `core/statemachine.CanTransition` for validation.
- Use DI interfaces (`TaskReader`, `WorkUnitReader`) for cross-module reads.
- Reference `workunit.EventTypeForStatus` for event naming.

Forbidden:
- Direct mutation of `tasks` or `work_units` tables.
- Calling `task.Service` or `workunit.Service` methods.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module orchestration belongs ONLY to:
- `internal/core/orchestration`

---

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw database errors leaked outside the module.
- `CodeNotFound` for missing runs.
- `CodeInvalidTransition` for illegal status changes.
- Retryable errors must be marked explicitly.

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
