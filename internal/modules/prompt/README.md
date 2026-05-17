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

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules
- `models.go` → domain types (PromptSnapshot, ToolsetSnapshot, Fragment)
- `events.go` → event-type mapping for prompt lifecycle
- `queries.go` → SQL constants for prompt_fragments, prompt_snapshots, toolset_snapshots
- `repository.go` → prompt fragment CRUD, no business logic
- `service.go` → prompt preparation service for runs
- `validation.go` → input validation

### Optional Files
- `types.go` → Fragment, Composer, Toolset types and constants (legacy — will merge into models.go)
- `catalog.go` → fragment catalog loader
- `catalog/` → markdown fragment files organized by category
- `composer.go` → fragment selection and validation logic
- `composer_render.go` → template rendering, system profile building, formatting helpers
- `composer_test.go` → composer tests
- `toolset.go` → toolset snapshot logic
- `toolset_test.go` → toolset tests
- `repository_snapshot.go` → prompt snapshot and toolset snapshot CRUD (legacy — will merge into repository.go)

---

## Allowed Dependencies

- `internal/core/apperrors`, `core/db`, `core/validation`, `core/event`
- `internal/core/serialization`
- `internal/domain`: ONLY `EventEnvelope` and generic types (never entity structs)
- DI interface types: `task.Task`, `workunit.WorkUnit`, `run.Run`, `agentsession.AgentSession`
  — see ADR-0026: types may be imported ONLY for DI interface return types.

Forbidden:
- `internal/modules/*` services, repositories, or business logic imports
- `internal/core/coordination` (reserved for orchestrator module)
- Direct imports of service logic from other modules.

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. SQL belongs only in `queries.go`.
