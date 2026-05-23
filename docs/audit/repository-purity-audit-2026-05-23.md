# Repository Purity Audit — 2026-05-23

**Scope:** All `repository.go` files under `internal/modules/`
**Standard:** ADR-0030 Pilar 4 — `repository.go` is pure CRUD: no business logic, no timestamps, no deduplication.

---

## Summary

| Module | `time.Now()` | `ON CONFLICT` / Upsert | Field Defaults/Validation | `ORDER BY` / `LIMIT` in Queries | Status |
|---|---|---|---|---|---|
| `agent` | ✅ Yes | ❌ No | ✅ Yes (ID generation) | ✅ Yes | ❌ Violations |
| `agentsession` | ✅ Yes (6 locations) | ❌ No | ❌ No | ✅ Yes | ❌ Violations |
| `orchestrator` | N/A (no table) | N/A | N/A | N/A | ✅ N/A |
| `prompt` | ❌ No | ✅ Yes | ✅ Yes (nil-checks, defaults) | ✅ Yes | ❌ Violations |
| `review` | ✅ Yes | ❌ No | ❌ No | ✅ Yes | ❌ Violations |
| `run` | ✅ Yes | ❌ No | ✅ Yes (Attempt default) | ✅ Yes | ❌ Violations |
| `task` | ✅ Yes | ❌ No | ❌ No | ✅ Yes | ❌ Violations |
| `taskgraph` | ❌ No | ❌ No | ❌ No | ✅ Yes | ⚠️ Minor |
| `trigger` | ❌ No | ❌ No | ❌ No | ✅ Yes | ⚠️ Minor |
| `workunit` | ✅ Yes | ❌ No | ❌ No | ✅ Yes | ❌ Violations |

**None are 100% pure.** The closest to clean are `taskgraph` and `trigger` (only `ORDER BY` violations in queries).

---

## Violation Details

### 1. `agent/repository.go`
- **Line ~29:** `now := time.Now().UTC()` in `Create`
- **Line ~25-27:** `if agent.ID == "" { agent.ID = uuid.New().String() }` — field default before insert

### 2. `agentsession/repository.go`
- **Line ~34:** `now := time.Now()` in `Create`
- **Line ~74:** `now := time.Now()` in `UpdateStatus`
- **Line ~93:** `now := time.Now()` in `UpdateHeartbeat`
- **Line ~108:** `now := time.Now().UTC()` in `UpdateHeartbeatWithEvent`
- **Line ~124:** `now := time.Now()` in `UpdateCheckpoint`
- **Line ~139:** `now := time.Now().UTC()` in `UpdateCheckpointWithEvent`
- **Line ~155:** `now := time.Now().UTC()` in `UpdateRecoverableState`

### 3. `prompt/repository.go`
- **Line ~145-167:** `CreateOrReferencePromptSnapshot` contains deduplication logic (`FirstUsedAt`, `LastUsedAt`, nil-checks, conditional marshalling) — **business logic in repository**
- **Line ~257-265:** `CreateToolsetSnapshot` does nil-check and defaults to `[]ToolsetTool{}` — business logic
- **`queries.go` line ~12-18:** `ON CONFLICT (composition_hash) DO UPDATE` — **upsert in queries**

### 4. `review/repository.go`
- **Line ~30:** `now := time.Now().UTC()` in `Create`
- **Line ~140:** `now := time.Now().UTC()` in `UpdateStatus`

### 5. `run/repository.go`
- **Line ~36:** `now := time.Now()` in `Create`
- **Line ~117:** `now := time.Now()` in `UpdateStatus`
- **Line ~32-34:** `if run.Attempt == 0 { run.Attempt = 1 }` — field default before insert

### 6. `task/repository.go`
- **Line ~88:** `task.UpdatedAt = time.Now()` in `Update`

### 7. `workunit/repository.go`
- **Line ~34:** `now := time.Now()` in `Create`

---

## Recommendations (Batch 2)

1. **Remove all `time.Now()` calls from repositories.** Pass timestamps from `service.go` or use database `DEFAULT now()`.
2. **Remove field defaults/validation from repositories.** `service.go` should prepare complete structs before calling repository.
3. **Move `prompt` upsert logic to `service.go`.** The repository should only `INSERT`; deduplication/checking existence belongs to the service layer.
4. **Move `ORDER BY` / `LIMIT` semantics to `service.go` or accept them as parameters.** Queries should be plain SQL constants without business ordering.
