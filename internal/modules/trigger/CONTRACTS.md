# Contracts: trigger

## Invariants

- Trigger status transitions are atomic and emit exactly one domain event.
- Detectors are deterministic: same input always produces the same output.
- Detectors have no side effects; they only analyze and return triggers.
- `TriggerService` is the only component that persists triggers and emits events.
- Terminal statuses (`resolved`, `dismissed`) cannot transition to any other status.
- Duplicate active/triggered triggers are suppressed: `persistDetectedTrigger` checks for an existing similar trigger (same type, run/session, anomaly) before inserting.
- All read operations (`ListActive`, `ListByRun`) accept and propagate `context.Context`.

---

## State Machine

Valid transitions:

| From | To |
|---|---|
| active | triggered, dismissed |
| triggered | resolved, dismissed |

Invalid transitions:
- Any transition from `resolved` or `dismissed`.
- Any transition not listed in the valid table.

---

## Execution Rules

- Always validate inputs before mutation.
- Never bypass the state machine.
- Never update a trigger without emitting a domain event.
- State transitions must be atomic (single transaction).
- Detectors must be pure functions (no I/O, no mutation, no randomness).
- `persistDetectedTrigger` must deduplicate against existing active/triggered triggers before creating a new one.

---

## Boundary Rules

Allowed:
- Read and mutate the `triggers` table via `repository.go`.
- Append events via `core/transition` helpers.
- Read from event store for replay-based detection.

Forbidden:
- Direct mutation of `runs`, `agent_sessions`, or `work_units` tables.
- Calling `run.Service` or `agentsession.Service` methods.
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
| `CodeNotFound` | Missing triggers |
| `CodeInvalidTransition` | State machine violation |
| `CodeConflict` | Idempotency / concurrency violation |
| `CodePersistence` | Database errors |

---

## Persistence Rules

- All writes must go through `repository.go`.
- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD + row-scanning.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.
- All query methods accept `context.Context` and use `QueryContext`.

---

## File Decomposition

No service decomposition at this time. `service.go` is the single file for trigger lifecycle logic.

### `detectors.go`
Deterministic anomaly detectors extracted for reuse and testability.

### `thresholds.go`
ThresholdConfig defaults and validation extracted for reuse across detectors and service.

---

## Related ADRs

- ADR-0022: Vertical Slice Architecture
- ADR-0025: Module Standardization
