# Subpackage Candidates Audit — 2026-05-23

**Scope:** All modules under `internal/modules/`
**Goal:** Identify modules or logic blocks that merit extraction into subpackages for better cohesion and maintainability.

---

## Current Module Sizes (`.go` files only)

| Module | Files | Lines of .go | Assessment |
|---|---|---|---|
| `agent` | 11 | ~488 | Cohesive ✅ |
| `agentsession` | 12 | ~870 | Fragmented — 3 service files ⚠️ |
| `orchestrator` | 8 | ~1,400 | Naturally large (orchestrates all) ⚠️ |
| `prompt` | 11 | ~1,100 | `catalog/` moved out; now cohesive ✅ |
| `review` | 8 | ~1,020 | Cohesive but large ✅ |
| `run` | 12 | ~850 | Fragmented — 3 service files ⚠️ |
| `task` | 9 | ~490 | Cohesive ✅ |
| `taskgraph` | 12 | ~850 | Fragmented — 5 planner files ⚠️ |
| `trigger` | 11 | ~1,300 | Very large service ⚠️ |
| `workunit` | 9 | ~1,020 | Cohesive ✅ |

---

## Candidates for Subpackage Extraction

### 1. `run/` → `run/relay/` (Medium Priority)

**Files:** `service_relay.go`, `retry.go`

**Rationale:** The `RuntimeEventRelay` and retry logic are orthogonal to the core run lifecycle. They handle external runtime communication, not run state management.

**Proposed structure:**
```
run/
├── service.go          # Run lifecycle (Create, Start, Complete, Cancel)
├── repository.go
├── ...
└── relay/
    ├── relay.go        # RuntimeEventRelay
    └── retry.go        # Retry policy/config
```

**Impact:** `orchestrator` would import `run/relay` types for DI if needed, or the relay could be wired in `bootstrap/`.

---

### 2. `taskgraph/` → `taskgraph/planner/` (Medium Priority)

**Files:** `planner.go`, `planner_prompt.go`, `planner_validator.go`, `heuristic.go`, `gemini_planner.go`

**Rationale:** 5 files (~40% of the module) are dedicated to planning strategies. As new planners are added (OpenAI, local LLM, rule-based), this will grow.

**Proposed structure:**
```
taskgraph/
├── service.go          # TaskGraph lifecycle
├── repository.go
├── models.go
├── ...
└── planner/
    ├── planner.go          # Interface + orchestration
    ├── heuristic.go        # Heuristic implementation
    ├── gemini.go           # Gemini implementation
    ├── prompt.go           # Prompt builder
    └── validator.go        # Plan validation
```

**Impact:** Clean separation between "graph storage" and "graph planning".

---

### 3. `agentsession/` → `agentsession/checkpoint/` + `agentsession/heartbeat/` (Low Priority)

**Files:** `service_checkpoint.go`, `service_heartbeat.go`

**Rationale:** Checkpoint and heartbeat are specialized subsystems. However, they are tightly coupled to `AgentSession` state. Unless they grow significantly, keeping them as separate files within the module is acceptable.

**Verdict:** Keep as-is for now. Revisit if either file exceeds 200 lines.

---

### 4. `trigger/` → `trigger/evaluator/` + `trigger/detector/` (Medium Priority)

**Rationale:** `trigger/service.go` is 655 lines — the largest service. It likely mixes detection logic, threshold evaluation, and action dispatch.

**Files already present:** `detectors.go`, `thresholds.go` suggest some separation exists but may not be complete.

**Proposed structure:**
```
trigger/
├── service.go          # Orchestration + dispatch
├── repository.go
├── ...
├── detector/
│   └── detector.go     # Detection strategies
└── evaluator/
    └── evaluator.go    # Threshold evaluation engine
```

**Verdict:** Recommended for Batch 2 after analyzing `trigger/service.go` line-by-line.

---

### 5. `prompt/` — `catalog` already externalized ✅

With `configs/prompts/` now holding the fragment catalog, `prompt/` is well-structured. No further subpackages needed unless a new concern emerges (e.g., `prompt/render/` for template engines).

---

## Modules That Should Stay Flat

| Module | Reason |
|---|---|
| `agent` | Small, focused. `fake_runtime.go` and `gemini_runtime.go` are strategy implementations that belong inside the module. |
| `task` | Small, single responsibility. |
| `review` | Large but cohesive — all logic revolves around review lifecycle. |
| `workunit` | Cohesive, single responsibility. |

---

## Summary Table

| Candidate | Priority | Effort | Benefit |
|---|---|---|---|
| `run/relay/` | Medium | Low | High — separates external communication |
| `taskgraph/planner/` | Medium | Medium | High — isolates planning strategies |
| `trigger/evaluator/` + `detector/` | Medium | Medium | High — breaks up 655-line service |
| `agentsession/checkpoint/` | Low | Low | Low — premature |
