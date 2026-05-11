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

- `doc.go` → package documentation and context briefing
- `models.go` → domain type aliases (`Status`)
- `planner.go` → Planner interface definition
- `gemini_planner.go` → Gemini LLM planner implementation
- `planner_prompt.go` → prompt rendering for LLM planner
- `planner_validator.go` → plan validation logic
- `queries.go` → SQL constants for task_graphs
- `repository.go` → task-graph CRUD
- `service.go` → decomposition orchestration, graph lifecycle, idempotency
- `heuristic.go` → local heuristic decomposition from acceptance criteria
- `task_graph_service_test.go` → service tests

---

## Allowed Dependencies

- `internal/core/*` (db, eventstore, orchestration, validation, serialization, apperrors)
- `internal/domain`
- `internal/modules/task` (for `RequireByID` / task reads)
- `internal/modules/workunit` (for work-unit creation during decomposition)
- `internal/modules/agent` (for `GeminiPlanner` runtime)

Forbidden:
- Direct imports of `task.Service` or `workunit.Service`
- Cross-module mutations outside `core/orchestration`

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. Every graph creation must emit an event via `core/orchestration`.
6. SQL belongs only in `queries.go`.
