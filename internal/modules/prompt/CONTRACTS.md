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
- Append events via `core/orchestration` helpers.
- Read other aggregates via their repositories (not services) for composition context.

Forbidden:
- Direct mutation of `tasks`, `work_units`, `runs`, or `agent_sessions` tables.
- Calling service methods from other modules.
- Inline SQL outside `queries.go`.
- Business logic inside `repository.go`.

Cross-module orchestration belongs ONLY to:
- `internal/core/orchestration`

---

## Error Rules

- All failures must map to `apperrors.Error` with a code and operation.
- No raw database errors leaked outside the module.
- `CodeValidation` for missing required categories or unresolved conflicts.
- `CodeNotFound` for missing runs or sessions referenced in input.

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
- Hidden side effects.
- Cross-module mutations via service imports.
- Business logic inside repositories.
- Inline SQL strings.
