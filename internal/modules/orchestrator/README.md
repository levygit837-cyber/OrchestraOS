# Module: orchestrator

## Purpose

This module is responsible for:
- Coordinating the complete automated execution of a task from creation to completion.
- Orchestrating calls to TaskService, TaskGraphService, RunService, AgentService, AgentSessionService, PromptService, ReviewService, and TriggerService.
- Managing the sequential (and eventually parallel) execution of work units via Runtime + EventRelay.

This module DOES NOT:
- Implement runtime execution logic (belongs to `agent/`).
- Manage persistent agent or session state directly (belongs to `agent/` and `agentsession/`).
- Compose prompts directly (belongs to `prompt/`).

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- `RunTask` is the sole entry point for automated task execution.
- Cross-module communication is strictly mediated by this service; modules never talk directly.
- Work units are executed in topological order respecting the DAG.
- Every state transition emits an event.
- Timeouts and heartbeat monitoring are enforced per work unit.

State Flow:
```
RunTask → GetTask → DecomposeGraph → ForEachWU:
  CreateRun → StartRun → FindOrCreateAgent → CreateSession → PreparePrompt → StartRuntime → RelayEvents → CompleteRun
→ CompleteTask
```

---

## File Map

- `doc.go` → package documentation and context briefing
- `service.go` → OrchestratorService with RunTask and dependencies
- `models.go` → RunTaskOptions, RunTaskResult, and auxiliary types
- `validation.go` → input validation for options
- `events.go` → event-type mapping for orchestrator lifecycle
- `contract.go` → module contract
- `queries.go` → SQL constants (if needed for orchestrator-specific tables)

---

## Allowed Dependencies

- `internal/core/*` (db, event, orchestration, statemachine, validation, serialization, apperrors, transition)
- `internal/domain`
- `internal/modules/*` services (via DI interfaces, never direct repository imports)

Forbidden:
- Direct imports of `internal/modules/*/repository.go`.
- Business logic beyond task orchestration and coordination.

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. State transitions MUST emit events.
6. SQL belongs only in `queries.go`.
7. All writes must use transactions via `core/db.BeginTx` / `CommitTx` / `RollbackTx`.
