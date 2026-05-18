# Audit: ADR-to-Code Alignment — 2026-05-18

> **Branch:** `fix/adr-alignment-and-code-inconsistencies`  
> **Auditor:** Kimi Code CLI (SubAgent 1 + 2)  
> **Scope:** ADR contradictions, missing state machine aggregates, transaction boundary violations, documentation drift  

---

## Summary

This audit documents **7 findings** discovered while aligning ADRs with the current codebase.  
**4 were fixed in this branch** (simple documentation and code standardization).  
**3 were deferred** to future PRs due to complexity (new state machine aggregates, transaction boundary refactor).

---

## Findings Fixed in This Branch

### ✅ FIX-1: Outdated ADR 0022 Section 5.2 (Renaming Proposals)
- **Severity:** Low
- **File:** `docs/adr/0022-vertical-module-architecture.md`
- **Issue:** Section 5.2 proposed renaming `orchestrator/` → `taskflow/` and `event/` → `eventappend/`. These renames were never applied and caused confusion.
- **Fix:** Section removed entirely. Orchestrator remains the single cross-module coordinator.

### ✅ FIX-2: Non-Existent ADR References in Orchestrator Docs
- **Severity:** Low
- **Files:** `internal/modules/orchestrator/doc.go`, `README.md`, `CONTRACTS.md`
- **Issue:** Referenced ADR 0021 and ADR 0027 which do not exist as standalone files (0021 absorbed into 0020; 0027 absorbed into 0022).
- **Fix:** Removed all references. ADR 0020 and 0022 cited instead.

### ✅ FIX-3: Contradictory `FORBIDDEN core/*` Lines in contract.go
- **Severity:** Medium
- **Files:** `internal/modules/*/contract.go` (9 modules)
- **Issue:** Comments said `// FORBIDDEN core/* imports` but modules legitimately import `core/transition`, `core/statemachine`, etc. This misleads LLM agents.
- **Fix:** Removed contradictory lines. Kept accurate allowed/forbidden lists.

### ✅ FIX-4: Inconsistent Event Type Casing
- **Severity:** Medium
- **File:** `internal/modules/agent/gemini_runtime.go`
- **Issue:** Emitted `"agent.Started"` (PascalCase) while `fake_runtime.go` emits `"agent.started"` (lowercase). Event types must be lowercase per convention.
- **Fix:** Changed to `"agent.started"`.

### ✅ FIX-5: Hardcoded Terminal Status Check in Retry Logic
- **Severity:** Medium
- **File:** `internal/modules/run/service_retry.go`
- **Issue:** Manual check `previous.Status != StatusFailed && previous.Status != StatusCancelled` instead of using `transition.IsFinalStatus()`.
- **Fix:** Replaced with `transition.IsFinalStatus(string(previous.Status))`.

### ✅ FIX-6: Canonical Type Aliases Misused
- **Severity:** Low
- **File:** `internal/modules/task/service.go`
- **Issue:** Used `core/event` aliases (`eventmod.Envelope`) instead of canonical `domain.EventEnvelope`.
- **Fix:** Now uses `domain.EventEnvelope` and `domain.EventPriorityNotification` directly.

---

## Findings Deferred to Future PRs

### ⚠️ DEF-1: Missing State Machine Aggregates (CRITICAL)

**Policy:** Every domain module MUST call `core/statemachine.CanTransition` before mutating status.  
**Current Violations:** 4 modules perform manual status validation.

| Module | File | Lines | Current Behavior | Required Aggregate |
|---|---|---|---|---|
| `trigger` | `service.go` | 400–406 | Manual check: `fromStatus == StatusResolved \|\| StatusDismissed` → reject | `AggregateTrigger` |
| `review` | `service.go` | 187–188, 245–250 | Manual checks: `Status != StatusPending`, `isFinalReviewStatus()`, etc. | `AggregateReview` |
| `taskgraph` | `service.go` | 139–152 | Direct SQL `UPDATE ... status = 'superseded'` without transition validation | `AggregateTaskGraph` |
| `agent` | `fake_runtime.go`, `gemini_runtime.go` | 52, 79, 88, 112, 227, 237, 269, 279, 290, 326, 375, 392 | Direct assignment to `status.State` ("running", "stopped", "completed", "failed") | `AggregateAgent` |

**Proposed Aggregates for `internal/core/statemachine/statemachine.go`:**

```go
AggregateTrigger   = "trigger"
AggregateReview    = "review"
AggregateTaskGraph = "taskgraph"
AggregateAgent     = "agent"
```

**Transition Rules:**

- **`AggregateTrigger`:** `active` → `triggered` → `resolved` OR `dismissed`
- **`AggregateReview`:** `pending` → `in_progress` → `approved` / `changes_requested` / `needs_discussion`
- **`AggregateTaskGraph`:** `active` → `superseded` (one-way, irreversible)
- **`AggregateAgent`:** `active` ↔ `inactive` (bidirectional), terminal: `failed`

**Impact if Not Fixed:**
- Status transitions bypass central rules → risk of invalid states in production.
- Replay and audit trails become unreliable.
- Architecture tests cannot enforce the rule because the aggregate definitions do not exist.

**Estimated Effort:** Medium (1–2 days). Requires:
1. Add aggregates and transitions to `statemachine.go`
2. Refactor 4 modules to call `CanTransition`
3. Add unit tests for each new aggregate
4. Update architecture tests to verify new aggregates

---

### ⚠️ DEF-2: Prompt Service Breaks Caller Transaction Control (HIGH)

**Severity:** High  
**File:** `internal/modules/prompt/service.go:242`  
**Issue:** `PrepareAndPersistPrompt` receives `*sql.Tx` from the caller (orchestrator) but internally calls `dbcore.CommitTx(tx)` at line 242. This commits the transaction **inside** the service, violating the rule that the caller owns the transaction lifecycle.

**Code:**
```go
func (s *PromptService) PrepareAndPersistPrompt(ctx context.Context, tx *sql.Tx, input PrepareAndPersistInput) (*PreparedRunPrompt, error) {
    // ... operations ...
    if err := dbcore.CommitTx(tx, "prompt_service.commit_prepare"); err != nil {  // ❌ WRONG
        return nil, err
    }
    return &PreparedRunPrompt{...}, nil
}
```

**Why This Is Dangerous:**
- Orchestrator may want to append additional events after prompt preparation in the **same** transaction.
- Committing inside the service makes rollback impossible if downstream operations fail.
- Violates the pattern used by all other services (`task`, `run`, `workunit`) where the caller commits.

**Fix Strategy:**
1. Remove `dbcore.CommitTx` from `PrepareAndPersistPrompt`.
2. Return `*PreparedRunPrompt` without committing.
3. Update all callers (orchestrator service, tests) to commit after the call.
4. Ensure `defer dbcore.RollbackTx(tx)` is present for error paths.

**Estimated Effort:** Low (2–4 hours). Requires updating callers in `orchestrator/` and tests.

---

### ⚠️ DEF-3: Integration Tests Bypass Services and State Machine (ACCEPTED RISK)

**Severity:** Low (documented as accepted)  
**Files:** `tests/integration/interaction_test.go`, `internal/modules/agent/fake_runtime_test.go`  
**Issue:** Tests call `repo.UpdateStatus()` directly instead of going through service methods. This bypasses:
- `CanTransition` validation
- Event emission
- Business logic guards

**Why Accepted:**
- These are **persistence-layer tests**, not service-layer tests.
- They verify that repositories correctly read/write to Postgres.
- Service-layer coverage exists in `tests/unit/`.

**Recommendation:** Add a code comment in each test file clarifying this intent so future developers do not mistake it for a bug.

---

## Additional Notes

### Missing Milestone Definitions
ADRs reference M0 and M3 (e.g., ADR 0013, ADR 0005, ADR 0007, ADR 0016) but no central document defines milestone scopes.  
**Recommendation:** Create `docs/milestones.md` or add an appendix to ADR 0013 defining M0, M1, M2, M3 scopes.

### Workspace vs Worktree Terminology
ADR 0029 introduced "Workspace Manager (WSM)" but older docs still say "worktree".  
**Status:** `project-canvas.md` was updated in this branch. Other docs may still need updates.

---

## Checklist for Next Sprint

- [ ] Add `AggregateTrigger`, `AggregateReview`, `AggregateTaskGraph`, `AggregateAgent` to `statemachine.go`
- [ ] Refactor `trigger/service.go` to call `CanTransition`
- [ ] Refactor `review/service.go` to call `CanTransition`
- [ ] Refactor `taskgraph/service.go` to call `CanTransition` before superseding
- [ ] Refactor agent runtime status assignments to use state machine (or document why not)
- [ ] Fix `prompt/service.go` transaction boundary (remove internal `CommitTx`)
- [ ] Add code comments to integration tests explaining direct repo access
- [ ] Create milestone definitions document
