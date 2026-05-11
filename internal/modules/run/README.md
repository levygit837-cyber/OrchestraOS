# Module: run

## Purpose

This module is responsible for:
- Managing Runs — executions of WorkUnits by agent sessions.
- Enforcing run status transitions and retry policies.
- Projecting run results and cascading status updates to related WorkUnits.
- Mapping terminal statuses to `RunResult` values.

This module DOES NOT:
- Manage task lifecycle (belongs to `task/`).
- Manage work-unit dependencies or path availability (belongs to `workunit/`).
- Manage agent sessions directly (belongs to `agentsession/`).

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- Run status transitions are atomic and emit exactly one event.
- Terminal statuses (`completed`, `failed`, `cancelled`) are immutable.
- `ResultForStatus` maps `completed` → `succeeded`, `failed` → `failed`, `cancelled` → `cancelled`.
- A Run can only be created for a WorkUnit in a non-terminal status.
- Retry policy defines max attempts and backoff strategy.

State Flow:
```
created → running → validating → completed
            ↓
    waiting_approval ←──┘
            ↓
        paused → failed / cancelled (terminal)
```

---

## File Map

- `doc.go` → package documentation and context briefing
- `models.go` → domain type aliases (`Status`, `Result`)
- `events.go` → event-type mapping and `ResultForStatus` helper
- `fetch.go` → read helpers
- `queries.go` → SQL constants for runs
- `repository.go` → run CRUD
- `service.go` → run lifecycle, status transitions, work-unit cascade sync
- `service_retry.go` → retry orchestration with backoff and idempotency
- `retry.go` → retry policy and backoff definitions

---

## Allowed Dependencies

- `internal/core/*` (db, eventstore, orchestration, statemachine, validation, serialization, apperrors)
- `internal/domain`
- `internal/modules/event` (indirectly via orchestration)
- `internal/modules/workunit` (for `EventTypeForStatus` and validation helpers)
- DI interfaces only: `TaskReader` (from `task/`), `WorkUnitReader` (from `workunit/`)

Forbidden:
- Direct imports of `task.Service` or `workunit.Service`
- Cross-module mutations outside `core/orchestration`

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. State transitions MUST use `core/statemachine.CanTransition`.
6. Every mutation MUST emit an event.
7. SQL belongs only in `queries.go`.
