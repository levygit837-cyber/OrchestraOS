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
- Agent profile is validated as non-empty snake_case; the database CHECK constraint enforces the allowed set.
- Agent runtime_type is validated against fixed allowed values.
- `FindOrCreate` is atomic (transaction + INSERT) and handles unique-violation races by falling back to SELECT.
- Every runtime must implement the `Runtime` interface.
- `FakeRuntime` is deterministic and safe for parallel tests.
- `GeminiPlanner` must return a valid `GraphPlan` or a typed error — no partial results.
- Runtime configuration must be validated before execution.

State Flow:
```
AgentService.Create → agent.created event → persisted agent
AgentService.FindOrCreate → SELECT → INSERT (atomic) OR existing agent
RuntimeConfig → Runtime.Start → Runtime.Execute → Runtime.Stop
```

---

## File Map

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules
- `models.go` → `Agent`, `RuntimeType`, `AgentStatus` definitions
- `events.go` → event-type mapping for agent lifecycle
- `queries.go` → SQL constants for agents table
- `repository.go` → agent CRUD, no business logic
- `service.go` → AgentService (Create, GetByID, FindOrCreate)
- `validation.go` → input validation (profile, runtime_type, name)

### Optional Files
- `runtime.go` → `Runtime` interface and `RuntimeConfig`
- `fake_runtime.go` → deterministic test double
- `gemini_runtime.go` → Gemini inference runtime
- `gemini_inference_test.go` → Gemini runtime tests
- `gemini_runtime_test.go` → additional Gemini tests
- `service_test.go` → AgentService validation tests

---

## Allowed Dependencies

- `internal/core/apperrors`, `core/db`, `core/validation`, `core/event`
- `internal/core/statemachine`, `core/transition`, `core/serialization`
- `internal/domain`: ONLY `EventEnvelope` and generic types (never entity structs)

Allowed from `internal/modules/*`:
- NONE — agent is a leaf module with no DI dependencies on other modules.

Forbidden:
- `internal/modules/*` services, repositories, or business logic imports
- `internal/core/coordination` (reserved for orchestrator module)
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
