# Contracts: agent

## Invariants

- Every runtime must implement the `Runtime` interface completely.
- `FakeRuntime` responses are deterministic for the same input configuration.
- `GeminiPlanner` returns either a fully valid `GraphPlan` or an error — no partial plans.
- Runtime configuration must be validated before `Start` is called.
- `RuntimeType` values are fixed constants; adding a new runtime requires updating validation.
- Agent creation emits exactly one `agent.created` domain event.
- Agent profile and runtime_type are validated against fixed allowed values.
- `FindOrCreate` reuses existing active agents when available.
- Agent ID must be a valid UUID.

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

---

## Boundary Rules

Allowed:
- Implement the `Runtime` interface.
- Call external APIs (Gemini) for inference.
- Use `internal/domain` types for input/output.
- Read and mutate the `agents` table via `repository.go`.
- Append events via `core/transition` helpers.
- Export `AgentReader` interface for cross-module reads (used by `agentsession/`).

Forbidden:
- Direct mutation of `tasks`, `work_units`, `runs`, or `agent_sessions` tables.
- Calling service methods from other modules.
- Importing `internal/modules/*` (except DI interfaces via bootstrap).
- Business logic beyond runtime execution, planning, and agent management.
- Inline SQL outside `queries.go`.

Cross-module orchestration belongs ONLY to:
- `internal/core/orchestration`

---

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw HTTP errors leaked outside the module.
- No raw database errors leaked outside the module.
- `CodeValidation` for invalid runtime configuration or agent input.
- `CodeExternal` for Gemini API failures.
- `CodeNotFound` for missing agents.
- `CodePersistence` for database operation failures.

---

## Persistence Rules

- All agent writes must go through `repository.go`.
- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD + row-scanning.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.
- Agent creation is atomic (single transaction with event append).

---

## LLM Execution Rules

LLM executors MUST:

1. Read `README.md` first.
2. Read `CONTRACTS.md` before editing.
3. Modify only files related to the assigned task.
4. Preserve all invariants.
5. Avoid speculative refactors.
6. Avoid introducing new abstractions unless required.
7. Keep implementations deterministic.
8. Preserve module boundaries.
9. Every mutation MUST emit an event.

---

## Forbidden Patterns

- Adding database or orchestration imports (except via core/*).
- Cross-module mutations.
- Business logic inside runtime implementations.
- Partial plan returns from `GeminiPlanner`.
- Inline SQL strings.
- Bypassing transaction boundaries.
- Silent errors (always wrap with `apperrors.Wrap`).
