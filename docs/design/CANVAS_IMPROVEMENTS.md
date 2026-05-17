# OrchestraOS Canvas — Proposed Improvements

> **Status:** Proposal
>
> **Context:** Based on deep analysis of OrchestraOS domain model (Tasks, TaskGraphs, WorkUnits, Runs, AgentSessions, Events, Triggers, ToolRequests, Reviews, Checkpoints, PromptSnapshots, Sandboxes).

---

## The Gap

The current Canvas variant is a beautiful shell but lacks **domain depth**. It shows a static graph with mock data. The real OrchestraOS is an event-driven, stateful system where:
- Agents request tool approvals (human must intervene)
- Runs have retry attempts with failure reasons
- Sessions heartbeat and create checkpoints
- Anomalies trigger alerts (stall, loop, token exceeded)
- Prompts are composed from versioned fragments
- WorkUnits have validation gates and reviews

The Canvas should make all of this **visible at a glance** and **actionable in one click**.

---

## Improvements by Impact

### 🔴 Critical (High Impact, Core to Domain)

#### 1. Event Stream Overlay
**What:** A retractable panel (bottom or side) showing live events from the Event Store.
**Why:** OrchestraOS is event-driven. The operator needs to see what's happening *right now*.
**Visual:** Scrolling list of event cards:
```
[10:32:14] [RUN] run-101 started for wu-002          → amber
[10:32:15] [TOOL] Codex-Builder requests file_write  → orange (needs approval!)
[10:32:16] [CHK] Checkpoint-003 reached              → green
[10:32:20] [EVT] Heartbeat from Review-Bot           → gray
```
**Actions:** Click event → jump to related node in graph. Filter by type/priority.

#### 2. Tool Request / Approval Queue HUD
**What:** A persistent indicator in the top HUD showing pending tool requests requiring human approval.
**Why:** This is the primary human intervention point in the system.
**Visual:**
```
[🔔 3]  ← pulsing amber badge when >0
```
**Click:** Opens approval queue overlay with:
- Agent name
- Tool requested (file_write, bash, etc.)
- Risk level
- Input preview
- [Approve] [Reject] [View Context] buttons

#### 3. Agent Session Detail in Inspector
**What:** Inspector shows live session data, not just static agent info.
**Why:** AgentSession has heartbeat, checkpoint, sandbox, recoverable_state.
**Fields to show:**
- Session status (starting/running/waiting_approval/paused/stopped)
- Last heartbeat: "2s ago"
- Last checkpoint: checkpoint-003
- Sandbox: container-id, branch, worktree path
- Current toolset (snapshot)
- Prompt snapshot used (with hash)

#### 4. Run Attempt History
**What:** WorkUnit inspector shows all runs (attempts) with their results.
**Why:** Runs have `attempt` number and `failure_reason`.
**Visual:**
```
Attempts:
  ● Attempt 1 — FAILED (timeout after 5m)
  ● Attempt 2 — FAILED (linter error)
  ▸ Attempt 3 — RUNNING (2m elapsed)
```
**Click:** Expand attempt → view logs, artifacts, prompt snapshot.

---

### 🟡 High (Important for Operator Experience)

#### 5. Anomaly / Trigger Alert Banner
**What:** Floating banner when triggers fire (stall, loop, drift, token exceeded).
**Why:** The system auto-detects problems. The operator must know immediately.
**Visual:** Top-center banner, colored by severity:
```
⚠️  Stall detected: Codex-Builder on wu-002 (idle for 5m)
    [Investigate] [Pause Agent] [Force Checkpoint]
```

#### 6. Edge Type Visualization
**What:** Different line styles for dependency types.
**Why:** Dependencies have types: `blocks`, `requires_artifact`, `requires_review`, `conflicts_with`.
**Visual:**
- `blocks`: Solid line (standard)
- `requires_artifact`: Dashed line with artifact icon
- `requires_review`: Dotted line with review icon
- `conflicts_with`: Red zigzag line

#### 7. Task Progress Ring
**What:** Circular progress indicator around the Task node.
**Why:** Tasks have multiple WUs. Operator wants % complete at a glance.
**Visual:** SVG ring around task chip, filling clockwise as WUs complete.

#### 8. "Jump to Active" Button
**What:** Button in HUD that auto-pans/zooms to whatever is currently running.
**Why:** In large graphs, finding the active node is tedious.
**Visual:** Small target icon in zoom controls. Shortcut: `A`.

#### 9. Context Menu (Right-click on Node)
**What:** Right-click any node → context menu with actions.
**Why:** Power users need quick actions without navigating away.
**Menu items by node type:**
- **Agent:** View Logs | Pause | Stop | View Sandbox | View Session
- **WU:** Retry Run | View Attempts | View Artifacts | Mark Complete | View Prompt
- **Task:** Cancel | View Graph History | Change Priority | View Reviews

#### 10. Filter Layer
**What:** Toggle chips in top HUD to filter visible nodes.
**Why:** Large projects have dozens of WUs. Filtering reduces noise.
**Filters:**
- Status: Running | Pending | Done | Failed | All
- Agent: Codex-Builder | Review-Bot | All
- Type: My WUs | WUs needing review | WUs with failures

---

### 🟢 Medium (Polish & Depth)

#### 11. Checkpoint Timeline
**What:** In Agent inspector, a mini-timeline of checkpoints reached.
**Why:** Checkpoints have `current_goal`, `completed_goals`, `blockers`, `risks`.
**Visual:** Horizontal timeline with dots. Click dot → expand checkpoint details.

#### 12. Sandbox Status Indicator
**What:** Small icon on agent node showing sandbox health.
**Why:** Sandbox has status, container_id, resource_limits.
**Visual:**
- 🟢 Container healthy
- 🟡 Container throttled
- 🔴 Container crashed

#### 13. Prompt Composition Preview
**What:** In WU inspector, show which prompt fragments were composed.
**Why:** PromptSnapshot has `fragment_refs`, `assembly_order`, hashes.
**Visual:** List of fragment titles with version badges. Expand to see body.

#### 14. Validation Gate Status
**What:** Visual indicator on WU showing if it passed validation gates.
**Why:** Reviews have `ValidationGate` (hard/soft/policy) and `ReviewStatus`.
**Visual:**
- Shield icon: 🛡️ Hard gate passed
- Circle icon: ○ Soft gate pending
- Triangle icon: △ Policy gate under review

#### 15. Priority & Risk Visualization
**What:** Task node shows P0/P1/P2/P3 priority and risk level.
**Why:** Tasks have `Priority` and `RiskLevel` fields.
**Visual:**
- Priority: colored dot (P0=red, P1=orange, P2=amber, P3=gray)
- Risk: subtle flame icon for high/critical

#### 16. Sound / Audio Feedback (Optional)
**What:** Subtle audio cues for critical events.
**Why:** Operator may not be looking at the screen.
**Events with sound:**
- Tool request pending approval (gentle chime)
- Run failed (alert tone)
- Task completed (success chime)
- Anomaly detected (warning tone)
**Respect:** `prefers-reduced-motion` and system volume settings.

#### 17. Keyboard Shortcuts HUD
**What:** `?` key opens overlay with all shortcuts.
**Why:** Power users live on keyboard.
**Visual:** Centered modal, amber accents, categorized shortcuts.

#### 18. Node Grouping / Clustering
**What:** Visually group WUs by agent or by phase.
**Why:** TaskGraphs can have 10+ WUs. Groups reduce visual clutter.
**Visual:** Subtle background rectangles with labels ("Phase 1: Setup", "Codex-Builder's WUs").

---

## Implementation Priority

| Priority | Feature | Effort | Impact |
|---|---|---|---|
| P0 | Tool Request / Approval Queue HUD | Medium | Critical |
| P0 | Agent Session Detail in Inspector | Low | Critical |
| P0 | Run Attempt History | Low | High |
| P1 | Event Stream Overlay | High | Critical |
| P1 | Context Menu | Medium | High |
| P1 | "Jump to Active" Button | Low | High |
| P1 | Edge Type Visualization | Low | Medium |
| P2 | Anomaly Alert Banner | Medium | High |
| P2 | Task Progress Ring | Low | Medium |
| P2 | Filter Layer | Medium | Medium |
| P3 | Checkpoint Timeline | Medium | Medium |
| P3 | Prompt Composition Preview | Medium | Low |
| P3 | Validation Gate Status | Low | Medium |

---

## Design Principles for These Features

1. **Don't block the canvas.** Every new element is HUD overlay. The graph remains king.
2. **Progressive disclosure.** Show summary at a glance, details on demand (inspector).
3. **Amber = requires attention.** Anything that needs human action glows amber.
4. **Green = all good.** Completed, healthy, approved.
5. **Red = problem.** Failed, crashed, anomaly.
6. **Sound is secondary.** Always visual-first. Audio is a backup notification channel.

---

## Next Steps

1. Review this list and select features to implement.
2. For each selected feature, create a detailed component spec in `COMPONENTS.md`.
3. Update the HTML prototype with the selected features.
4. Test interactions and visual hierarchy.
