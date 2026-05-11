# Contracts: agentsession

## Invariants

- Session status transitions are atomic and emit exactly one domain event.
- Terminal statuses (`stopped`, `failed`) cannot transition to any other status.
- Heartbeat updates `last_heartbeat_at` and `last_checkpoint_at` without emitting a status event.
- Checkpoint persists `recoverable_state`, `ledger`, and `evidence_refs` atomically.
- Timeout must transition the session to `failed` AND pause the associated Run in the same transaction.
- `starting` can only transition to `running` or `failed`.
- `stopping` can only transition to `stopped` or `failed`.

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

Valid transitions:

| From | To |
|---|---|
| starting | running, failed |
| running | waiting_approval, paused, disconnected, stopping, failed |
| waiting_approval | running, paused, stopping, failed |
| paused | running, stopping, failed |
| disconnected | running, stopping, failed |
| stopping | stopped, failed |

Invalid transitions:
- Any transition from `stopped` or `failed`.
- Any transition not listed in the valid table.

---

## Execution Rules

- Always validate state before mutation (`CanTransition`).
- Never bypass the state machine.
- Never update a session without emitting a domain event (except heartbeat).
- State transitions must be atomic (single transaction).
- Timeout handling must update both `agent_sessions` and `runs` in the same tx.
- Idempotency: duplicate event append returns the existing envelope without error.

---

## Boundary Rules

Allowed:
- Read and mutate the `agent_sessions` table via `repository.go`.
- Append events via `core/orchestration` helpers.
- Call `core/statemachine.CanTransition` for validation.
- Use `run.NewRepository(tx)` for Run pause on timeout.

Forbidden:
- Direct mutation of `tasks` or `work_units` tables.
- Calling `run.Service` methods.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module orchestration belongs ONLY to:
- `internal/core/orchestration`

---

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw database errors leaked outside the module.
- `CodeNotFound` for missing sessions.
- `CodeInvalidTransition` for illegal status changes.

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
- Hidden side effects (every write emits an event, except heartbeat).
- Cross-module mutations via service imports.
- Bypassing the state machine.
- Business logic inside repositories.
- Inline SQL strings.
