# Module: agentsession

## Purpose

This module is responsible for:
- Managing AgentSessions — bindings between an agent runtime and a Run.
- Handling creation, heartbeat, manual/automatic checkpoint, and timeout.
- Enforcing session status transitions and pausing the associated Run on timeout.

This module DOES NOT:
- Manage task or work-unit lifecycle.
- Execute prompts (belongs to `prompt/`).
- Decompose tasks (belongs to `taskgraph/`).

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- Session status transitions are atomic and emit exactly one event.
- Terminal statuses (`stopped`, `failed`) are immutable.
- Heartbeat updates `last_heartbeat_at` without changing status.
- Checkpoint persists recoverable state and ledger evidence.
- Timeout transitions the session to `failed` and pauses the associated Run atomically.

State Flow:
```
starting → running → stopping → stopped
              ↓
    waiting_approval ←──┘
              ↓
          paused → disconnected → failed (terminal)
              ↓
          failed (terminal)
```

---

## File Map

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules
- `models.go` → domain type aliases (`Status`)
- `events.go` → event-type mapping for session status transitions
- `queries.go` → SQL constants for agent_sessions
- `repository.go` → session CRUD, no business logic
- `service.go` → session lifecycle, creation, connect, disconnect, resume, stop, timeout, fail
- `validation.go` → input validation

### Optional Files
- `fetch.go` → read helpers
- `service_heartbeat.go` → heartbeat event append and projection update
- `service_checkpoint.go` → manual checkpoint with recoverable state persistence
- `checkpoint_policy.go` → automatic and suggested checkpoint logic

---

## Allowed Dependencies

- `internal/core/apperrors`, `core/db`, `core/validation`, `core/event`
- `internal/core/statemachine`, `core/transition`, `core/serialization`
- `internal/domain`: ONLY `EventEnvelope` and generic types (never entity structs)
- `internal/modules/run` (repository only for Run pause on timeout)

Forbidden:
- `internal/modules/*` (direct imports, except `run.Repository` for cascade)
- `internal/core/coordination` (reserved for orchestrator module)
- Direct imports of `run.Service`

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. State transitions MUST use `core/statemachine.CanTransition`.
6. Every mutation MUST emit an event.
7. SQL belongs only in `queries.go`.
