# Contracts: orchestrator

## Invariants

1. `RunTask` is the only public method of `OrchestratorService`.
2. `RunTask` never accesses repositories directly; it uses only injected domain services.
3. Work units are executed in topological order (DAG dependency order).
4. Each work unit execution is isolated: its own Run, AgentSession, and Runtime instance.
5. If a work unit fails, the orchestrator records the failure but continues with remaining independent work units in the first cut.
6. Every significant step emits an event via the EventStore.
7. Timeouts are enforced per work unit, not per task.
8. The orchestrator never mutates task, run, or session state directly; it delegates to the respective services.

---

## State Machine

```
RunTask called
  → task.fetched
  → graph.decomposed (if needed)
  → work_units.ordered
  → ForEach work_unit:
      → run.created
      → run.started
      → agent.found_or_created
      → session.created
      → session.connected
      → prompt.prepared
      → runtime.started
      → relay.running
      → [events routed]
      → run.completed | run.failed
      → triggers.evaluated
  → task.completed | task.partial | task.failed
```

---

## Execution Rules

- `RunTask` is the only public method — no direct repository access.
- Services are injected via interfaces; never import repositories from other modules.
- Work units are processed in topological order.
- Each work unit gets its own Run + AgentSession.
- Failures are recorded but do not halt the entire task unless all work units fail.
- Every significant step MUST emit an event.

---

## Boundary Rules

Allowed:
- Import `internal/core/coordination` (only module with this permission).
- Call injected service interfaces from all other modules.
- Emit events via `core/transition` helpers.

Forbidden:
- Direct repository imports from other modules.
- Inline SQL outside `queries.go`.
- Business logic beyond task orchestration and coordination.

Cross-module orchestration is the sole responsibility of this module.

---

## Error Rules

| Code | When to Use |
|------|-------------|
| `CodeValidation` | Invalid options or input |
| `CodeInvalidInput` | Semantically invalid input |
| `CodeNotFound` | Task or dependency not found |
| `CodeExternal` | Runtime API failures |
| `CodePersistence` | Database errors |

---

## Persistence Rules

- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.

---

## File Decomposition

No service decomposition at this time. `service.go` is the single file for orchestration logic.

---

## Related ADRs

- ADR-0022: Vertical Slice Architecture
- ADR-0025: Module Standardization
