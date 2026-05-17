# Module: orchestrator

> **⚠️ NOME FUTURO:** Este módulo será renomeado para `runner/` ou `taskflow/` em uma futura refatoração. O nome `orchestrator` será reservado para um módulo de orquestração de agentes (`director/`). Ver `docs/adr/0027-orchestrator-module-naming.md`.

## Purpose

This module is a **Task Execution Workflow Engine**. It is responsible for:
- Executing a task from start to finish by orchestrating calls to domain services in the correct sequence.
- Managing the sequential (and eventually parallel) execution of work units via Runtime + EventRelay.

This is **NOT** an "Agent Orchestrator" (a future `director/` module). It does not decide which task to run, allocate resources, or prioritize work. It simply executes a task that has already been scheduled.

This module DOES NOT:
- Decide which task to execute or when (belongs to future `director/`).
- Implement runtime execution logic (belongs to `agent/`).
- Manage persistent agent or session state directly (belongs to `agent/` and `agentsession/`).
- Compose prompts directly (belongs to `prompt/`).
- Perform low-level transaction coordination (belongs to `core/coordination/`).

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- `RunTask` is the sole entry point for automated task execution.
- Work units are executed in topological order respecting the DAG.
- Every state transition emits an event.
- Timeouts and heartbeat monitoring are enforced per work unit.
- This module does not mutate state directly; it delegates to domain services.

State Flow:
```
RunTask → GetTask → DecomposeGraph → ForEachWU:
  CreateRun → StartRun → FindOrCreateAgent → CreateSession → PreparePrompt → StartRuntime → RelayEvents → CompleteRun
→ CompleteTask
```

---

## File Map

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules
- `models.go` → RunTaskOptions, RunTaskResult, and auxiliary types
- `events.go` → event-type mapping for orchestrator lifecycle
- `queries.go` → SQL constants (if needed for orchestrator-specific tables)
- `service.go` → OrchestratorService with RunTask and dependencies
- `validation.go` → input validation for options

### Optional Files
- None at this time.

---

## Allowed Dependencies

- `internal/core/apperrors`, `core/db`, `core/validation`, `core/event`
- `internal/core/statemachine`, `core/transition`, `core/serialization`
- `internal/core/coordination` (ONLY module allowed to import this package)
- `internal/domain`: ONLY `EventEnvelope` and generic types (never entity structs)
- `internal/modules/review`: DI interface return type `*review.Review` in ReviewManager (ADR-0026)
- `internal/modules/taskgraph`: DI interface return type `*taskgraph.TaskGraph` in TaskGraphManager (ADR-0026)
- `internal/modules/workunit`: DI interface return type `[]workunit.WorkUnit` in WorkUnitLister (ADR-0026)
- `internal/modules/trigger`: DI interface return type `[]*trigger.Trigger` in TriggerEvaluator (ADR-0026)
- `internal/modules/prompt`: DI interface return type `*prompt.PromptSnapshot` and `*prompt.ToolsetSnapshot` in PreparedPrompt (ADR-0026)
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
