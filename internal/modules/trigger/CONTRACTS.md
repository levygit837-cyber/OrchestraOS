# Contracts: trigger

## Invariants

- Trigger status transitions are atomic and emit exactly one domain event.
- Detectors are deterministic: same input always produces the same output.
- Detectors have no side effects; they only analyze and return triggers.
- `TriggerService` is the only component that persists triggers and emits events.
- Terminal statuses (`resolved`, `dismissed`) cannot transition to any other status.

## State Machine

Valid transitions:

| From | To |
|---|---|
| active | triggered, dismissed |
| triggered | resolved, dismissed |

Invalid transitions:
- Any transition from `resolved` or `dismissed`.
- Any transition not listed in the valid table.

## Execution Rules

- Always validate inputs before mutation.
- Never bypass the state machine.
- Never update a trigger without emitting a domain event.
- State transitions must be atomic (single transaction).
- Detectors must be pure functions (no I/O, no mutation, no randomness).

## Boundary Rules

Allowed:
- Read and mutate the `triggers` table via `repository.go`.
- Append events via `core/transition` helpers.
- Read from event store for replay-based detection.

Forbidden:
- Direct mutation of `runs`, `agent_sessions`, or `work_units` tables.
- Calling `run.Service` or `agentsession.Service` methods.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw database errors leaked outside the module.
- `CodeNotFound` for missing triggers.
- `CodeInvalidTransition` for illegal status changes.

## Persistence Rules

- All writes must go through `repository.go`.
- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD + row-scanning.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.

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

## Forbidden Patterns

- Shared helpers inside the module (move to `core/` if reusable).
- Hidden side effects (every write emits an event).
- Cross-module mutations via service imports.
- Non-deterministic detectors (time.Now without injection, randomness).
- Business logic inside repositories.
- Inline SQL strings.
