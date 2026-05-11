# Module: task

## Purpose

This module is responsible for:
- Managing the full lifecycle of a Task from creation through completion.
- Enforcing task status transitions via the state machine.
- Cascading cancellation to dependent WorkUnits and Runs when a task is cancelled.
- Providing read access to tasks for other modules via `RequireByID` (DI interface).

This module DOES NOT:
- Execute work units or runs directly.
- Decompose tasks into graphs (belongs to `taskgraph/`).
- Manage agent sessions or prompts.

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- Task status transitions are atomic and emit exactly one event.
- Terminal statuses (`completed`, `failed`, `cancelled`) are immutable.
- Cancellation cascades to all dependent WorkUnits and Runs in the same transaction.
- Every creation defaults `priority=P2` and `risk_level=low` when omitted.

State Flow:
```
created → triaged → planned → scheduled → sandbox_preparing → running
                                            ↓
                        waiting_approval ←──┴──→ validating → completed
                                            ↓
                                        paused
                                            ↓
                                        failed / cancelled (terminal)
```

---

## File Map

- `doc.go` → package documentation and context briefing
- `models.go` → domain type aliases (`Status`, `Priority`, `RiskLevel`)
- `events.go` → event-type mapping for task status transitions
- `fetch.go` → `RequireByID` exported helper used as `TaskReader` by other modules
- `queries.go` → SQL constants for tasks
- `repository.go` → task CRUD
- `service.go` → task lifecycle logic, cancellation cascade
- `validation_test.go` → input validation tests

---

## Allowed Dependencies

- `internal/core/*` (db, orchestration, statemachine, validation, serialization, apperrors)
- `internal/domain`
- `internal/modules/event` (indirectly via orchestration)
- `internal/modules/run` (repository only for cancellation cascade)
- `internal/modules/workunit` (repository only for cancellation cascade)

Forbidden:
- Direct imports of `run.Service` or `workunit.Service`
- Cross-module mutations outside `core/orchestration`
- `internal/services` or `internal/repository` (removed)

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors — keep changes minimal and localized.
5. State transitions MUST use `core/statemachine.CanTransition`.
6. Every mutation MUST emit an event via `core/orchestration` helpers.
7. SQL belongs only in `queries.go`.
