# Contracts: {{MODULE_NAME}}

## Invariants

- {{INVARIANT}}
- {{INVARIANT}}
- {{INVARIANT}}
- {{INVARIANT}}

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

Valid transitions:

{{VALID_TRANSITIONS}}

Invalid transitions:

{{INVALID_TRANSITIONS}}

Rules enforced by `core/statemachine.CanTransition`:
1. Terminal statuses cannot transition to any other status.
2. `completed` transitions require evidence, validation event, or justification.
3. No transition is allowed unless explicitly listed above.

---

## Execution Rules

- Always validate state before mutation (`CanTransition`).
- Never bypass the state machine.
- Never update an entity without emitting a domain event.
- State transitions must be atomic (single transaction).
- Idempotency: duplicate event append returns the existing envelope without error.

---

## Boundary Rules

Allowed:
- Read own tables via `repository.go`.
- Append events via `core/coordination` helpers.
- Call `core/statemachine.CanTransition` for validation.
- Use DI interfaces (`TaskReader`, `WorkUnitReader`, `TaskGraphManager`) to read other aggregates.

Forbidden:
- Direct mutation of another module's tables.
- Calling another module's service methods.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module orchestration belongs ONLY to:
- `internal/core/coordination`

---

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw database errors leaked outside the module.
- Retryable errors must be explicitly marked (`apperrors.CodePersistence`).
- Validation errors must be deterministic and idempotent.

---

## Persistence Rules

- All writes must go through `repository.go`.
- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD + row-scanning.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.

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

- Shared helpers inside the module (move to `core/` if reusable).
- Hidden side effects (every write emits an event).
- Cross-module mutations.
- Bypassing the state machine.
- Business logic inside repositories.
- Inline SQL strings.
