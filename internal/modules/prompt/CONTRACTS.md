# Contracts: prompt

## Invariants

- Prompt snapshots are deduplicated by `composition_hash` using UPSERT semantics.
- `MaxAutonomyLevel = 2` is the hard ceiling for any prompt fragment.
- All `RequiredCategories` must be present in a composed prompt.
- Toolset snapshots are immutable after creation.
- Fragment conflicts (`ConflictsWith`) must be resolved at composition time.
- Fragment requirements (`Requires`) must be satisfied at composition time.
- Every prepared prompt must reference valid RunID and AgentSessionID.

Violating invariants is considered a **CRITICAL FAILURE**.

---

## State Machine

This module does not manage a traditional lifecycle state machine.
Snapshots are immutable records with the following flow:

```
PrepareRunPromptInput → composition → snapshot (deduped) → usage tracking
```

---

## Execution Rules

- Always validate input UUIDs before composition.
- Composition must resolve all `Requires` and `ConflictsWith` relationships.
- Snapshots must be persisted before being returned to the caller.
- Deduplication must increment `count_used` and update `last_used_at` atomically.
- Idempotency: identical composition inputs must return the same snapshot.

---

## Boundary Rules

Allowed:
- Read and mutate `prompt_fragments`, `prompt_snapshots`, `toolset_snapshots` via `repository.go`.
- Append events via `core/transition` helpers.
- Compose prompts from a `TaskContext` (pure function, no external data fetching).

Forbidden:
- Direct mutation of `tasks`, `work_units`, `runs`, or `agent_sessions` tables.
- Calling service methods from other modules.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module data gathering and orchestration belongs ONLY to:
- `internal/modules/orchestrator`

This module receives all necessary data via `TaskContext` and `PersistMetadata`; it never fetches from other modules.

---

## Error Rules

| Code | When to Use |
|------|-------------|
| `CodeValidation` | Missing required categories or unresolved conflicts |
| `CodeInvalidInput` | Semantically invalid input |
| `CodeNotFound` | Missing runs or sessions referenced in input |
| `CodeConflict` | Idempotency / concurrency violation |
| `CodePersistence` | Database errors |

---

## Persistence Rules

- All writes must go through `repository.go`.
- SQL belongs only in `queries.go`.
- No business logic inside repositories — pure CRUD + row-scanning.
- Use `core/db.BeginTx` / `CommitTx` / `RollbackTx` for transactions.

---

## File Decomposition

### `composer_render.go`
Created because `composer.go` exceeded 300 lines. Extracted template rendering, system profile building, and formatting helpers.

No further decomposition at this time.

---

## Related ADRs

- ADR-0015: Vertical Slice Architecture
- ADR-0019: Module Standardization
