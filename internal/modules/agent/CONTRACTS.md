# Contracts: agent

## Invariants

- Every runtime must implement the `Runtime` interface completely.
- `FakeRuntime` responses are deterministic for the same input configuration.
- `GeminiPlanner` returns either a fully valid `GraphPlan` or an error — no partial plans.
- Runtime configuration must be validated before `Start` is called.
- `RuntimeType` values are fixed constants; adding a new runtime requires updating `validation.Runtime`.

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

This module does not manage a domain lifecycle state machine.
Runtime lifecycle:

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

---

## Boundary Rules

Allowed:
- Implement the `Runtime` interface.
- Call external APIs (Gemini) for inference.
- Use `internal/domain` types for input/output.

Forbidden:
- Direct mutation of `tasks`, `work_units`, `runs`, or `agent_sessions` tables.
- Calling service methods from other modules.
- Importing `internal/modules/*` or `internal/core/orchestration`.
- Business logic beyond runtime execution and planning.

Cross-module orchestration belongs ONLY to:
- `internal/core/orchestration`

---

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw HTTP errors leaked outside the module.
- `CodeValidation` for invalid runtime configuration.
- `CodeExternal` for Gemini API failures.

---

## Persistence Rules

- This module does NOT persist to the database.
- All persistence is handled by consumers (`taskgraph/`, `agentsession/`).

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

- Adding database or orchestration imports.
- Cross-module mutations.
- Business logic inside runtime implementations.
- Partial plan returns from `GeminiPlanner`.
