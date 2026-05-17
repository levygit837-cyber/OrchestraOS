# Module: trigger

## Purpose

This module is responsible for:
- Managing configurable anomaly detection and threshold triggers for runs, agent sessions, and work units.
- Evaluating runs, sessions, and work units against deterministic detectors.
- Creating, resolving, and dismissing triggers with proper event emission.

This module DOES NOT:
- Manage task or run lifecycle directly.
- Execute agent code.
- Modify work units, runs, or sessions (only reads for evaluation).

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- Trigger status transitions are atomic and emit exactly one event.
- Detectors are deterministic: same input always produces the same output.
- Detectors have no side effects; they only analyze and return triggers.
- Terminal statuses (`resolved`, `dismissed`) are immutable.
- Duplicate active/triggered triggers are suppressed.

---

## File Map

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules
- `models.go` → domain type aliases (`Status`, `TriggerType`)
- `events.go` → event-type mapping for trigger lifecycle
- `queries.go` → SQL constants for triggers
- `repository.go` → trigger CRUD, no business logic
- `service.go` → `TriggerService` with Create, EvaluateRun, EvaluateSession, EvaluateWorkUnit, Resolve, Dismiss, ListActive, ListByRun
- `validation.go` → input validation

### Optional Files
- `fetch.go` → `RequireByID` helper
- `detectors.go` → deterministic anomaly detectors
- `thresholds.go` → ThresholdConfig defaults and validation

---

## Allowed Dependencies

- `internal/core/apperrors`, `core/db`, `core/validation`, `core/event`
- `internal/core/statemachine`, `core/transition`, `core/serialization`
- `internal/domain`: ONLY `EventEnvelope` and generic types (never entity structs)
- DI interfaces only: `RunReader` (from `run/`), `AgentSessionReader` (from `agentsession/`), `WorkUnitReader` (from `workunit/`)

Forbidden:
- `internal/modules/*` (direct imports)
- `internal/core/coordination` (reserved for orchestrator module)
- Direct imports of service logic from other modules.

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. State transitions MUST use `core/statemachine.CanTransition`.
6. Every mutation MUST emit an event.
7. SQL belongs only in `queries.go`.
