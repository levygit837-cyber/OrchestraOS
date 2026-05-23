# Contracts: workunit

## Invariants

- WorkUnit status transitions are atomic and emit exactly one domain event.
- Terminal statuses (`completed`, `failed`, `cancelled`) cannot transition to any other status.
- `completed` transitions require evidence, validation event, or justification.
- `ValidateWorkUnitDependencies` must verify that every ID in `depends_on` exists and is not `failed` or `cancelled`.
- `ValidateOwnedPathAvailability` must reject creation if any `owned_paths` overlap with an existing active work unit.
- Creation requires the parent Task to have an active TaskGraph.
- `ValidateDependenciesCompleted` and `ValidateOwnedPathAvailability` are exported for use by `run/`.

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

Valid transitions:

| From | To |
|---|---|
| created | planned, scheduled, blocked, running, cancelled |
| planned | scheduled, blocked, failed, cancelled |
| scheduled | running, blocked, paused, cancelled |
| blocked | scheduled, running, failed, cancelled |
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
- Never update a work unit without emitting a domain event.
- State transitions must be atomic (single transaction).
- Dependency validation must run before scheduling or running.
- Idempotency: duplicate event append returns the existing envelope without error.

---

## Boundary Rules

Allowed:
- Read and mutate the `work_units` table via `repository.go`.
- Append events via `core/transition` helpers.
- Call `core/statemachine.CanTransition` for validation.
- Use DI interfaces (`TaskReader`, `TaskGraphManager`) for cross-module reads.
- Export `ValidateDependenciesCompleted` and `ValidateOwnedPathAvailability` for `run/`.

Forbidden:
- Direct mutation of `tasks` or `runs` tables.
- Calling `task.Service` or `run.Service` methods.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module orchestration belongs ONLY to:
- `internal/modules/orchestrator`

---

## Error Rules

| Code | When to Use |
|------|-------------|
| `CodeValidation` | Invalid input syntax |
| `CodeInvalidInput` | Semantically invalid input |
| `CodeNotFound` | WorkUnit does not exist |
| `CodeInvalidTransition` | State machine violation |
| `CodeConflict` | Path collisions or invalid dependencies |
| `CodePersistence` | Database errors |

---

## Persistence Rules

- All writes must go through `repository.go`.
- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD + row-scanning.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.

---

## File Decomposition

### `service_create.go`
Created because `service.go` exceeded 300 lines. Extracted batch creation of work units and manual task-graph activation logic.

No further decomposition at this time.

---

## Related ADRs

- ADR-0015: Vertical Slice Architecture
- ADR-0019: Module Standardization
