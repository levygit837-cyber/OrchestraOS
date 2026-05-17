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

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules
- `models.go` → domain types (`Run`, `Status`, `Result`)
- `events.go` → event-type mapping and `ResultForStatus` helper
- `queries.go` → SQL constants for runs
- `repository.go` → run CRUD, no business logic
- `service.go` → run lifecycle, status transitions, work-unit cascade sync
- `validation.go` → input validation

### Optional Files
- `fetch.go` → read helpers
- `retry.go` → retry policy and backoff definitions
- `service_retry.go` → retry orchestration with backoff and idempotency

---

## Allowed Dependencies

- `internal/core/apperrors`, `core/db`, `core/validation`, `core/event`
- `internal/core/statemachine`, `core/transition`, `core/serialization`
- `internal/domain`: ONLY `EventEnvelope` and generic types (never entity structs)
- DI interface types: `task.Task` (for `TaskReader`), `workunit.WorkUnit` (for `WorkUnitReader`)
  — see ADR-0026: types may be imported ONLY for DI interface return types.

Forbidden:
- `internal/modules/*` services, repositories, or business logic imports
- Direct imports of `task.Service` or `workunit.Service`
- Cross-module mutations outside `orchestrator/` module

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. State transitions MUST use `core/statemachine.CanTransition`.
6. Every mutation MUST emit an event.
7. SQL belongs only in `queries.go`.
