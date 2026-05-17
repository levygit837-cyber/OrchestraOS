# Coding Standards

## Purpose

This document codifies the rules that every contributor (human or LLM) must follow when modifying the codebase. Violations caught by automated tools (`golangci-lint`, architecture tests) will block CI. Violations not caught by tools should still be treated as blockers during review.

---

## Module Standards

Every module under `internal/modules/` MUST contain:

1. `doc.go` — package documentation and context briefing
2. `README.md` — operational map, responsibilities, file map, allowed dependencies
3. `CONTRACTS.md` — invariants, state machine, boundary rules, error rules
4. `queries.go` — SQL constants only (unless the module is a leaf with no DB access)
5. `models.go` — domain type aliases and constants
6. `repository.go` — pure CRUD, no business logic
7. `service.go` — domain logic, state transitions, event emission

### Forbidden in any module

- `helpers.go` or `utils.go` — move reusable code to `internal/core/`
- Inline SQL strings outside `queries.go`
- Business logic inside `repository.go`
- Direct mutation of another module's tables
- Calling another module's `Service` methods — use DI interfaces or `core/coordination`
- `panic()` — always return `apperrors.Error`
- `fmt.Println` / `fmt.Printf` — use structured logging or return errors

---

## State Machine Rules

1. Every status transition MUST call `core/statemachine.CanTransition` before mutating state.
2. Terminal statuses (`completed`, `failed`, `cancelled`, `stopped`) are immutable.
3. `completed` transitions require `EvidenceRefs`, `ValidationEventID`, or `Justification`.
4. Every mutation MUST emit a domain event in the same transaction.

---

## Error Handling

1. All errors exposed outside a module MUST be `apperrors.Error`.
2. Raw database errors MUST be wrapped with `apperrors.Wrap`.
3. Use `apperrors.CodeNotFound` for missing entities.
4. Use `apperrors.CodeInvalidTransition` for illegal status changes.
5. Use `apperrors.CodeConflict` for concurrency/idempotency violations.
6. Never ignore an error (`_ = someCall()`). If it truly cannot fail, document why with a comment.

---

## Transaction Rules

1. Always use `core/db.BeginTx`, `CommitTx`, `RollbackTx`.
2. Prefer `defer dbcore.RollbackTx(tx)` immediately after `BeginTx`.
3. Verify mutations with `dbcore.EnsureRowsAffected` when updating single rows.
4. Keep transactions as short as possible — do not call external APIs inside a transaction.

---

## Naming Conventions

| Construct | Convention | Example |
|---|---|---|
| Service | `*Service` | `TaskService` |
| Repository | `*Repository` | `TaskRepository` |
| Input struct | `Create*Input`, `Update*Input` | `CreateTaskInput` |
| Result struct | `*Result` | `TaskGraphDecomposeResult` |
| Status constant | `Status*` | `StatusCreated` |
| Event type | `package.status` | `task.created`, `run.started` |
| Operation name | `package.function` | `task_service.create` |

---

## Testing

1. Unit tests belong next to the file they test (`*_test.go`).
2. Architecture tests belong in `tests/architecture/`.
3. Integration tests belong in `tests/integration/`.
4. Every state transition MUST have at least one test path.
5. Mock external dependencies; test business logic in isolation.

---

## Adding a New Module

1. Run `./scripts/new-module.sh <name>`.
2. Fill in `README.md` and `CONTRACTS.md`.
3. Implement `models.go`, `queries.go`, `repository.go`, `service.go`.
4. Add the service factory to `internal/bootstrap/services.go`.
5. Run `go test ./internal/modules/<name>`.
6. Run `./scripts/verify-contracts.sh`.
7. Run `./scripts/lint.sh`.
