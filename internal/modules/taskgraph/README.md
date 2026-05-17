# Module: taskgraph

## Purpose

This module is responsible for:
- Decomposing Tasks into directed acyclic graphs (DAGs) of WorkUnits.
- Managing TaskGraph lifecycle: creation, versioning, activation, and supersession.
- Providing planner strategies: local heuristic and Gemini LLM-based decomposition.
- Validating graph plans before persistence.

This module DOES NOT:
- Manage task status transitions (belongs to `task/`).
- Execute work units or runs (belongs to `workunit/` and `run/`).
- Manage agent sessions or prompts.

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- A Task can have at most one `active` TaskGraph at any time.
- Superseding a graph marks the old one `superseded` before activating the new one.
- Graph plans must be validated before persistence (node count, edge count, acyclicity).
- Local heuristic planner enforces `minGraphWorkUnits=2` and `maxGraphWorkUnits=5`.

State Flow:
```
active → superseded
```

---

## File Map

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules
- `models.go` → domain types (`TaskGraph`, `Status`)
- `events.go` → event-type mapping for graph lifecycle
- `queries.go` → SQL constants for task_graphs
- `repository.go` → task-graph CRUD, no business logic
- `service.go` → decomposition orchestration, graph lifecycle, idempotency
- `validation.go` → input validation

### Optional Files
- `planner.go` → Planner interface definition
- `gemini_planner.go` → Gemini LLM planner implementation
- `planner_prompt.go` → prompt rendering for LLM planner
- `planner_validator.go` → plan validation logic
- `heuristic.go` → local heuristic decomposition from acceptance criteria
- `task_graph_service_test.go` → service tests

---

## Allowed Dependencies

- `internal/core/apperrors`, `core/db`, `core/validation`, `core/event`
- `internal/core/statemachine`, `core/transition`, `core/serialization`
- `internal/domain`: ONLY `EventEnvelope` and generic types (never entity structs)
- DI interface types: `task.Task` (for `TaskReader`), `workunit.WorkUnit` (for `WorkUnitWriter`)
  — see ADR-0026: types may be imported ONLY for DI interface return types.

Forbidden:
- `internal/modules/*` services, repositories, or business logic imports
- `internal/core/coordination` (reserved for orchestrator module)
- Direct imports of `task.Service` or `workunit.Service`

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. Every graph creation must emit an event via `core/transition`.
6. SQL belongs only in `queries.go`.
