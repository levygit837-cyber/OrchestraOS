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

## Error Handling

- Validation errors → return immediately without side effects.
- Runtime errors → fail the current run, record event, continue if possible.
- Timeout errors → mark run and session as timed out, record recoverable state.
- Database errors → wrap with apperrors.CodePersistence and return.

## Testing

- Use `FakeRuntime` for deterministic E2E tests.
- Mock services for unit tests of RunTask logic.
- Verify topological ordering with dependency graphs.
