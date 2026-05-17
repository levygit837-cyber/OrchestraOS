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
- Append events via `core/transition` helpers.
- Call `core/statemachine.CanTransition` for validation.
- Use `run.NewRepository(tx)` for Run pause on timeout.

Forbidden:
- Direct mutation of `tasks` or `work_units` tables.
- Calling `run.Service` methods.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module orchestration belongs ONLY to:
- `internal/core/coordination`
- `internal/modules/orchestrator`

---

## Error Rules

| Code | When to Use |
|------|-------------|
| `CodeValidation` | Invalid input syntax |
| `CodeInvalidInput` | Semantically invalid input |
| `CodeNotFound` | Session does not exist |
| `CodeInvalidTransition` | State machine violation |
| `CodeConflict` | Idempotency / concurrency violation |
| `CodePersistence` | Database errors |

---

## Persistence Rules

- All writes must go through `repository.go`.
- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD + row-scanning.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.

---

## File Decomposition

### `service_heartbeat.go`
Created because `service.go` exceeded 300 lines. Extracted heartbeat event append and projection update logic.

### `service_checkpoint.go`
Created because `service.go` exceeded 300 lines. Extracted manual checkpoint with recoverable state persistence logic.

### `checkpoint_policy.go`
Automatic and suggested checkpoint logic, shared between `service.go` and `service_checkpoint.go`.

No further decomposition at this time.

---

## Related ADRs

- ADR-0022: Vertical Slice Architecture
- ADR-0025: Module Standardization
