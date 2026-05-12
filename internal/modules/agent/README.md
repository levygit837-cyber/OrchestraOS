# Module: agent

## Purpose

This module is responsible for:
- Managing Agent entities — persistent records of agent configurations (profile, runtime type, capabilities, tools).
- Providing the Runtime interface for agent execution.
- Providing concrete runtime implementations: Fake (testing), Gemini (LLM inference), Codex CLI, and External.
- Implementing the GeminiPlanner used by `taskgraph/` for LLM-based task decomposition.

This module DOES NOT:
- Manage task, run, or work-unit lifecycle.
- Manage agent sessions (belongs to `agentsession/`).
- Compose prompts (belongs to `prompt/`).

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- Agent creation emits exactly one `agent.created` event.
- Agent profile and runtime_type are validated against fixed allowed values.
- `FindOrCreate` reuses existing active agents when available.
- Every runtime must implement the `Runtime` interface.
- `FakeRuntime` is deterministic and safe for parallel tests.
- `GeminiPlanner` must return a valid `GraphPlan` or a typed error — no partial results.
- Runtime configuration must be validated before execution.

State Flow:
```
AgentService.Create → agent.created event → persisted agent
AgentService.FindOrCreate → existing agent OR Create → agent.created event
RuntimeConfig → Runtime.Start → Runtime.Execute → Runtime.Stop
```

---

## File Map

- `doc.go` → package documentation and context briefing
- `models.go` → `Agent`, `RuntimeType`, `AgentStatus` definitions and converters
- `runtime.go` → `Runtime` interface and `RuntimeConfig`
- `service.go` → AgentService (Create, GetByID, FindOrCreate) with persistence
- `repository.go` → agent CRUD and query operations
- `queries.go` → SQL constants for agents table
- `validation.go` → input validation (profile, runtime_type, name)
- `events.go` → event-type mapping for agent lifecycle
- `contract.go` → module contract and AgentReader interface
- `fake_runtime.go` → deterministic test double
- `gemini_runtime.go` → Gemini inference runtime
- `gemini_inference_test.go` → Gemini runtime tests
- `gemini_runtime_test.go` → additional Gemini tests
- `service_test.go` → AgentService validation tests

---

## Allowed Dependencies

- `internal/core/*` (db, orchestration, statemachine, validation, serialization, apperrors, transition)
- `internal/domain`

Forbidden:
- Direct imports of `internal/modules/*` (except DI interfaces via bootstrap).
- Business logic beyond runtime execution, planning, and agent management.

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. State transitions MUST emit events.
6. SQL belongs only in `queries.go`.
7. All writes must use transactions via `core/db.BeginTx` / `CommitTx` / `RollbackTx`.
