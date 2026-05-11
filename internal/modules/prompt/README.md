# Module: prompt

## Purpose

This module is responsible for:
- Preparing, composing, and storing prompts and toolsets for agent runs.
- Managing prompt snapshots with deduplication by composition hash.
- Managing toolset snapshots bound to agent sessions.
- Assembling prompt fragments from the catalog into system + task prompts.

This module DOES NOT:
- Manage task or run lifecycle.
- Execute agent code.
- Manage agent sessions directly (only snapshots).

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- Prompt snapshots are deduplicated by `composition_hash` (UPSERT semantics).
- `MaxAutonomyLevel = 2` is the highest allowed autonomy level for any prompt.
- All required fragment categories must be present in a composed prompt.
- Toolset snapshots are immutable once created.

State Flow:
```
(fragment catalog) → composition → snapshot → usage
```

---

## File Map

- `doc.go` → package documentation and context briefing
- `types.go` → Fragment, Composer, Toolset types and constants
- `catalog.go` → fragment catalog loader
- `catalog/` → markdown fragment files organized by category
- `composer.go` → fragment selection and validation logic
- `composer_render.go` → template rendering, system profile building, formatting helpers
- `composer_test.go` → composer tests
- `toolset.go` → toolset snapshot logic
- `toolset_test.go` → toolset tests
- `queries.go` → SQL constants for prompt_fragments, prompt_snapshots, toolset_snapshots
- `repository.go` → prompt fragment CRUD
- `repository_snapshot.go` → prompt snapshot and toolset snapshot CRUD
- `service.go` → prompt preparation service for runs

---

## Allowed Dependencies

- `internal/core/*` (db, orchestration, validation, serialization, apperrors)
- `internal/domain`
- `internal/modules/task` (for task reads during prompt composition)
- `internal/modules/workunit` (for work-unit reads during prompt composition)
- `internal/modules/run` (for run reads during prompt preparation)
- `internal/modules/agentsession` (for session reads during prompt preparation)

Forbidden:
- Direct imports of service logic from other modules.
- Cross-module mutations outside `core/orchestration`.

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. SQL belongs only in `queries.go`.
