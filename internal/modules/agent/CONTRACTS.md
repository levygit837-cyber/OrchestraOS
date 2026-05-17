# Contracts: agent

## Invariants

- Every runtime must implement the `Runtime` interface completely.
- `FakeRuntime` responses are deterministic for the same input configuration.
- `GeminiPlanner` returns either a fully valid `GraphPlan` or an error — no partial plans.
- Runtime configuration must be validated before `Start` is called.
- `RuntimeType` values are fixed constants; adding a new runtime requires updating validation.
- Agent creation emits exactly one `agent.created` domain event.
- Agent profile is validated as non-empty snake_case; the database CHECK constraint enforces the allowed set.
- `FindOrCreate` is atomic (transaction + INSERT) and handles unique-violation races by falling back to SELECT.
- Agent ID must be a valid UUID.
- Empty arrays (capabilities, allowed_tools, default_prompt_fragments) are persisted as empty arrays, not NULL.

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

This module manages two state machines:

### Agent Lifecycle
```
active → inactive
```
Agents are created with status `active` and can be deactivated. Status changes are not yet implemented but reserved for future use.

### Runtime Lifecycle
```
Configured → Starting → Running → Stopping → Stopped
                              ↓
                           Failed
```

---

## Execution Rules

- Always validate `RuntimeConfig` before starting a runtime.
- `Execute` must handle context cancellation gracefully.
- `Stop` must be idempotent and safe to call multiple times.
- `GeminiPlanner` must validate the returned plan before returning it.
- Agent creation must use transactions (`core/db.BeginTx` / `CommitTx` / `RollbackTx`).
- Agent creation must emit an event via `transition.AppendServiceEvent`.
- `GetByID` must validate that the ID is a valid UUID before querying the database.
- `FindOrCreate` must run inside a transaction and handle `23505` unique-violation by re-selecting the existing row.

---

## Boundary Rules

Allowed:
- Implement the `Runtime` interface.
- Call external APIs (Gemini) for inference.
- Read and mutate the `agents` table via `repository.go`.
- Append events via `core/transition` helpers.
- Export `AgentReader` interface for cross-module reads (used by `agentsession/`).

Forbidden:
- Direct mutation of `tasks`, `work_units`, `runs`, or `agent_sessions` tables.
- Calling service methods from other modules.
- Importing `internal/modules/*`.
- Business logic beyond runtime execution, planning, and agent management.
- Inline SQL outside `queries.go`.

Cross-module orchestration belongs ONLY to:
- `internal/modules/orchestrator`

---

## Error Rules

| Code | When to Use |
|------|-------------|
| `CodeValidation` | Invalid runtime configuration or agent input |
| `CodeInvalidInput` | Semantically invalid input |
| `CodeNotFound` | Missing agents |
| `CodeExternal` | Gemini API failures |
| `CodePersistence` | Database operation failures |

---

## Persistence Rules

- All agent writes must go through `repository.go`.
- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD + row-scanning.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.
- Agent creation is atomic (single transaction with event append).
- `FindOrCreate` uses a unique index on `(profile, runtime_type)` to prevent race-condition duplicates.

---

## File Decomposition

No service decomposition at this time. `service.go` is the single file for agent management.

---

## Related ADRs

- ADR-0022: Vertical Slice Architecture
- ADR-0025: Module Standardization
