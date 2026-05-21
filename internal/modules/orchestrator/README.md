# Module: orchestrator

> **Nome do Módulo:** `orchestrator/` é a **Camada de Orquestração Canônica** do OrchestraOS.
> É o único módulo autorizado a coordenar operações cross-module. Todas as interações entre módulos verticais que exigem composição de dados, transações distribuídas ou sequenciamento de operações DEVEM passar pelo `orchestrator/`.

## Purpose

This module is the **Canonical Cross-Module Orchestration Layer**. It is responsible for:
- Executing a task from start to finish by orchestrating calls to domain services in the correct sequence.
- Managing the sequential (and eventually parallel) execution of work units via Runtime + EventRelay.
- **Coordinating all cross-module interactions**: prompt preparation, run-workunit synchronization, cascade cancellation, runtime event relay, and any future coordination logic.

This is **NOT** an "Agent Orchestrator" in the sense of strategic decision-making (a future `director/` module may handle that). It does not decide which task to run, allocate resources, or prioritize work. It simply executes a task that has already been scheduled, and **mediates all cross-module communication**.

This module DOES NOT:
- Decide which task to execute or when (belongs to future `director/`).
- Implement runtime execution logic (belongs to `agent/`).
- Manage persistent agent or session state directly (belongs to `agent/` and `agentsession/`).
- Define prompt fragments or compose prompts in isolation (belongs to `prompt/`).
- Perform low-level transaction coordination for single-module operations (belongs to owner modules via DI interfaces).

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

### Optional / Contextual Files
- `service_<context>.go` → Cross-module coordination logic for a specific domain (e.g., `service_cascade.go`, `service_prompt.go`, `service_run_workunit_sync.go`).
- Subdirectories for complex coordination domains:
  - `execution/` → Runtime execution coordination, event relay
  - `recovery/` → Failover, retry, cascade cancellation
  - `scheduling/` → Future: work unit scheduling and prioritization

---

## Allowed Dependencies

- `internal/core/apperrors`, `core/db`, `core/validation`, `core/event`
- `internal/core/statemachine`, `core/transition`, `core/serialization`
- `internal/domain`: ONLY `EventEnvelope` and generic types (never entity structs)
- `internal/modules/review`: DI interface return type `*review.Review` in ReviewManager (ADR-0026)
- `internal/modules/taskgraph`: DI interface return type `*taskgraph.TaskGraph` in TaskGraphManager (ADR-0026)
- `internal/modules/workunit`: DI interface return type `[]workunit.WorkUnit` in WorkUnitLister (ADR-0026)
- `internal/modules/trigger`: DI interface return type `[]*trigger.Trigger` in TriggerEvaluator (ADR-0026)
- `internal/modules/prompt`: DI interface return type `*prompt.PromptSnapshot` and `*prompt.ToolsetSnapshot` in PreparedPrompt (ADR-0026)
- `internal/modules/*` services (via DI interfaces, never direct repository imports)

Forbidden:
- Direct imports of `internal/modules/*/repository.go` (exception: `service_cascade.go` and similar emergency coordination where cyclic imports would occur).
- Business logic that belongs to a single module (e.g., prompt composition details, task state machine rules).
- Any cross-module coordination outside this module.

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. State transitions MUST emit events.
6. SQL belongs only in `queries.go`.
7. All writes must use transactions via `core/db.BeginTx` / `CommitTx` / `RollbackTx`.
