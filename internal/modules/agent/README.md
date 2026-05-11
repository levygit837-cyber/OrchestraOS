# Module: agent

## Purpose

This module is responsible for:
- Defining the Runtime interface for agent execution.
- Providing concrete runtime implementations: Fake (testing), Gemini (LLM inference), Codex CLI, and External.
- Implementing the GeminiPlanner used by `taskgraph/` for LLM-based task decomposition.

This module DOES NOT:
- Manage task, run, or work-unit lifecycle.
- Manage agent sessions (belongs to `agentsession/`).
- Compose prompts (belongs to `prompt/`).

---

## Contract Summary

This module is governed by CONTRACTS.md.
You MUST read it before making any modification.

Critical invariants:
- Every runtime must implement the `Runtime` interface.
- `FakeRuntime` is deterministic and safe for parallel tests.
- `GeminiPlanner` must return a valid `GraphPlan` or a typed error — no partial results.
- Runtime configuration must be validated before execution.

State Flow:
```
RuntimeConfig → Runtime.Start → Runtime.Execute → Runtime.Stop
```

---

## File Map

- `doc.go` → package documentation and context briefing
- `models.go` → `Agent`, `RuntimeType` definitions
- `runtime.go` → `Runtime` interface and `RuntimeConfig`
- `fake_runtime.go` → deterministic test double
- `gemini_runtime.go` → Gemini inference runtime
- `gemini_inference_test.go` → Gemini runtime tests
- `gemini_runtime_test.go` → additional Gemini tests

---

## Allowed Dependencies

- `internal/core/apperrors`
- `internal/domain` (minimal — EventEnvelope is temporary)

Forbidden:
- Imports of `internal/modules/*`.
- Imports of `internal/core/db`, `internal/core/orchestration`.
- Business logic beyond runtime execution and planning.

---

## Notes for LLM Executors

1. Read `CONTRACTS.md` before editing.
2. Modify only files related to the assigned task.
3. Preserve all invariants listed above.
4. Avoid architectural refactors.
5. New runtimes must implement the full `Runtime` interface.
