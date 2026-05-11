# Contracts: taskgraph

## Invariants

- A Task can have at most one `active` TaskGraph at any time.
- Superseding a graph must atomically mark the old graph `superseded` and the new graph `active`.
- Graph plans must pass validation before persistence (node count, edge count, cycle detection).
- Local heuristic planner respects `minGraphWorkUnits=2` and `maxGraphWorkUnits=5`.
- GeminiPlanner must return a valid `GraphPlan` or an error — no partial plans.
- Every graph mutation emits a domain event.

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

Valid transitions:

| From | To |
|---|---|
| active | superseded |

Invalid transitions:
- `superseded` → any status.
- Any status → `active` without superseding the previous active graph.

---

## Execution Rules

- Always validate the plan before persisting nodes, edges, and work units.
- Never create a graph without associated WorkUnits.
- Graph creation must be atomic (graph + work units in one transaction).
- Idempotency: duplicate event append returns the existing envelope without error.

---

## Boundary Rules

Allowed:
- Read and mutate the `task_graphs` table via `repository.go`.
- Append events via `core/orchestration` helpers.
- Use `task.RequireByID` for cross-module reads.
- Use `workunit.NewRepository(tx)` for work-unit creation during decomposition.

Forbidden:
- Direct mutation of `tasks`, `runs`, or `agent_sessions` tables.
- Calling `task.Service` or `workunit.Service` methods.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module orchestration belongs ONLY to:
- `internal/core/orchestration`

---

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw database errors leaked outside the module.
- `CodeValidation` for invalid graph plans.
- `CodeConflict` for concurrent active graph creation.

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
- Hidden side effects (every write emits an event).
- Cross-module mutations via service imports.
- Business logic inside repositories.
- Inline SQL strings.
