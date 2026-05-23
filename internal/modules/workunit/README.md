# Module: workunit

## Purpose

This module is responsible for:
- Managing WorkUnits — the smallest assignable units of work within a TaskGraph.
- Enforcing status transitions and dependency validation.
- Checking owned-path availability and validating work-unit dependencies.
- Activating manual task graphs when no active graph exists.

This module DOES NOT:
- Manage task lifecycle (belongs to `task/`).
- Execute runs directly (belongs to `run/`).
- Decompose tasks into graphs (belongs to `taskgraph/`).

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- WorkUnit status transitions are atomic and emit exactly one event.
- Terminal statuses (`completed`, `failed`, `cancelled`) are immutable.
- `ValidateWorkUnitDependencies` ensures all `depends_on` IDs exist and are in valid states.
- `ValidateOwnedPathAvailability` prevents path collisions between work units.
- A WorkUnit can only be created if its Task has an active TaskGraph.

State Flow:
```
created → planned → scheduled → blocked → running → validating → completed
            ↓           ↓           ↓         ↓           ↓
        cancelled   paused    failed    waiting_approval  failed
                                                    ↓
                                                cancelled
```

---

## File Map

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules
- `models.go` → domain types (`WorkUnit`, `Status`)
- `events.go` → event-type mapping for work-unit status transitions
- `queries.go` → SQL constants for work_units
- `repository.go` → work-unit CRUD, no business logic
- `service.go` → work-unit lifecycle (assign, block, schedule, start, validate, complete, fail, cancel)
- `validation.go` → dependency validation (DAG acyclicity), owned-path availability checks

### Optional Files
- `fetch.go` → read helpers
- `service_create.go` → batch creation of work units and manual task-graph activation
- `validation_test.go` → invariant and rule tests

---

## Allowed Dependencies

- `internal/core/apperrors`, `core/db`, `core/validation`, `core/event`
- `internal/core/statemachine`, `core/transition`, `core/serialization`
- `internal/domain`: ONLY `EventEnvelope` and generic types (never entity structs)
- DI interface types: `task.Task` (for `TaskReader`), `taskgraph.TaskGraph` (for `TaskGraphManager`)
  — see ADR-0019: types may be imported ONLY for DI interface return types.

Forbidden:
- `internal/modules/*` services, repositories, or business logic imports
- Direct imports of `task.Service` or `taskgraph.Service`
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
