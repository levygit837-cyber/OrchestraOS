# OrchestraOS — Visual Orchestration Flow

> **Status:** Design Spec — Draft
>
> **Scope:** How the real OrchestraOS orchestration flow should be represented visually
>
> **Purpose:** Bridge the gap between the domain model (ADRs) and the visual interface. Every entity in the flow must have a visual counterpart.

---

## 1. The Problem

The current Canvas shows a static DAG with nodes that "just exist." This is misleading because:

1. **Tasks don't appear from nowhere** — they originate from a user message
2. **Decomposition is invisible** — the user never sees Task → TaskGraph → WorkUnits happening
3. **The Orchestrator is invisible** — the control plane that schedules, monitors, and intervenes has no visual presence
4. **WorkUnits are not directly executed** — they become Runs, which spawn AgentSessions, which run in Sandboxes. These layers are collapsed into one
5. **Retries are hidden** — when a Run fails, the retry loop is invisible until you open an inspector
6. **Tool requests feel disconnected** — they appear in a modal instead of as part of the execution timeline

The Canvas must tell the **true story** of what OrchestraOS does.

---

## 2. The Real Flow (Synthesized from ADRs)

### 2.1 Actors and Entities

| Entity | Role | When It Appears |
|--------|------|-----------------|
| **UserMessage** | Human input (text, command, attachment) | T=0 — starting point |
| **Intake** | Normalizes and routes the message | T=0 — immediate |
| **OrchestratorService (Go)** | Tactical control plane — approves, creates, schedules | T+0 — always active |
| **Intelligent Orchestrator Agent (LLM)** | Interprets intent, suggests Tasks, diagnoses anomalies | T+1s — when strategic decisions needed |
| **Task** | Product work unit | T+2s — after intent validation |
| **TaskGraph (DAG)** | Acyclic decomposition plan | T+3s — after decomposition |
| **WorkUnit** | Executable node in the DAG | T+3s — with the graph |
| **Run** | Execution attempt of a WorkUnit | T+N — when WU is ready |
| **Agent** | Registered worker with profile | T+N — when Run is created |
| **AgentSession** | Agent work session inside sandbox | T+N — with the Run |
| **Sandbox** | Isolated container + worktree | T+N — with the session |
| **Checkpoint** | Persistent progress mark | During execution |
| **ToolRequest** | Tool request by agent | During execution |
| **Review / ValidationGate** | Quality gate after completion | After WU completion |
| **Event** | Everything that happens becomes an event | Continuous |

### 2.2 The Chronological Flow

```
[User types message]
        ↓
[Intake] normalizes → "create event envelope system"
        ↓
[OrchestratorService] receives
        ↓
┌─────────────────────────────────────────────────────┐
│ [Intelligent Agent] interprets intent (optional)    │
│ → Suggests: Task{title:"Event Envelope", priority:P0}│
└─────────────────────────────────────────────────────┘
        ↓
[OrchestratorService] validates and creates Task
        ↓
[TaskGraphService.Decompose()] creates DAG
        ↓
[OrchestratorService] walks DAG topologically
        ↓
For each WorkUnit ready:
    [RunService.Create()] → Run
    [AgentService.FindOrCreate()] → Agent (by profile)
    [AgentSessionService.Create()] → AgentSession
    [PromptService.PrepareRunPrompt()] → PromptSnapshot
    [Runtime.Start()] → Sandbox active
        ↓
    [Agent executes in sandbox]
        ↓
    [Checkpoints emitted during work]
        ↓
    [ToolRequests → approval if risk > low]
        ↓
    [Run completes or fails → retry if needed]
        ↓
Next WorkUnit...
        ↓
[All WUs complete → Task complete]
        ↓
[Reviews / ValidationGates executed]
        ↓
[Task marked done → artifacts collected]
```

---

## 3. Visual Representation by Phase

### Phase 0: Input — The Origin Story

**What happens:** User sends a message.

**What should appear:**
- A **UserMessage node** at the top-left or top-center of the canvas
- The raw message text (truncated if long, expandable on click)
- Attachments indicator if present
- Timestamp

**Visual:**
```
┌─────────────────────────────────────────┐
│ 👤 User — 10:23:45                       │
│                                         │
│ "Implementar sistema de event envelope  │
│  para padronizar mensagens entre agentes│
│  com metadata e correlation IDs"        │
│                                         │
│ 📎 1 attachment                         │
└─────────────────────────────────────────┘
         ↓
```

**Why it matters:** Without this, Tasks appear magically. The user needs to see their own message as the root of everything that follows.

---

### Phase 1: Intake & Routing

**What happens:** The Intake service normalizes the message, extracts intent, and routes to the Orchestrator.

**What should appear:**
- An **Intake node** connected below the UserMessage
- Normalized intent displayed
- Routing decision (which Orchestrator handler)
- If ambiguous: a **branching indicator** showing that the Intelligent Agent may be consulted

**Visual:**
```
┌─────────────────────────────────────────┐
│ 🔀 Intake — 10:23:46                     │
│                                         │
│ Normalized: "create EventEnvelope with  │
│  metadata, timestamps, correlation IDs" │
│                                         │
│ Routing: orchestrator.primary           │
│ Confidence: 0.94                        │
└─────────────────────────────────────────┘
         ↓
```

**Why it matters:** Shows that the system understood the request before acting on it.

---

### Phase 2: Orchestrator Hub — The Control Plane

**What happens:** OrchestratorService receives the normalized intent, optionally consults the Intelligent Agent, and creates a Task.

**What should appear:**
- A prominent **Orchestrator Hub node** — larger than other nodes, visually distinct
- Status indicator: "Analyzing", "Decomposing", "Scheduling"
- Connection lines to all active WorkUnits
- **Live activity indicator** (amber pulse) when orchestrating

**Visual:**
```
         ┌─────────────────────────────┐
         │     ⚙️ ORCHESTRATOR         │
         │                             │
         │  Status: SCHEDULING         │
         │  Active: 2 runs             │
         │  Pending: 1 approval        │
         │                             │
         │  [👁️ View Events]           │
         └─────────────────────────────┘
              ↙          ↓          ↘
```

**Why it matters:** The Orchestrator is the protagonist of the system. It must be visible, not hidden.

---

### Phase 3: Task Creation

**What happens:** A Task entity is created with title, priority, risk level, and acceptance criteria.

**What should appear:**
- A **Task node** connected to the Orchestrator Hub
- Title, priority badge (P0/P1/P2/P3), risk indicator
- Acceptance criteria (collapsed by default, expandable)
- Progress ring that fills as WorkUnits complete

**Visual:**
```
┌─────────────────────────────────────────┐
│ 📋 task-099 — Event Envelope            │
│                                         │
│ Priority: 🔴 P0    Risk: 🟡 Medium      │
│                                         │
│ Criteria: 6 defined                     │
│ Progress: ██░░░░ 3/6 WUs (50%)          │
│                                         │
│ [Expand criteria] [View graph history]  │
└─────────────────────────────────────────┘
```

**Why it matters:** The Task is the product-level unit. The user cares about task completion, not individual WUs.

---

### Phase 4: TaskGraph Decomposition

**What happens:** TaskGraphService decomposes the Task into a DAG of WorkUnits.

**What should appear:**
- **Animation:** The Task node "expands" or a new DAG area appears below it
- WorkUnit nodes appear with dependency edges
- **Decomposition strategy label** (semantic, heuristic, etc.)
- Nodes at the same topological level are aligned horizontally (shows parallelism)

**Visual:**
```
                    ┌─────────┐
                    │ WU-001  │ ← Criar schema base
                    └────┬────┘
                         │ blocks
           ┌─────────────┼─────────────┐
           ↓             ↓             ↓
      ┌─────────┐  ┌─────────┐  ┌─────────┐
      │ WU-002  │  │ WU-003  │  │ WU-005  │ ← Parallel!
      │Implement│  │Middleware│  │CI Setup │
      └────┬────┘  └────┬────┘  └─────────┘
           │            │
           ↓            ↓
      ┌─────────┐  ┌─────────┐
      │ WU-004  │  │ WU-006  │
      │ Validate│  │ Document│
      └─────────┘  └─────────┘
```

**Edge types (visualized):**
| Type | Visual |
|------|--------|
| `blocks` | Solid line → |
| `requires_artifact` | Dashed line with 📄 icon |
| `requires_review` | Dotted line with 👁️ icon |
| `conflicts_with` | Red zigzag line |

**Why it matters:** The DAG is the plan. Users need to see parallelism, dependencies, and bottlenecks at a glance.

---

### Phase 5: WorkUnit → Run → AgentSession → Sandbox

**What happens:** When a WorkUnit is ready (dependencies satisfied), the Orchestrator creates a Run, finds/creates an Agent, spawns a Session, prepares prompts, and starts a Sandbox.

**What should appear (CRITICAL CHANGE from current design):**

Instead of collapsing these into one node, show them as a **vertical stack** or **timeline strip** connected to the WorkUnit:

```
WU Node:
┌─────────────────────────────────────────┐
│ wu-002 — Implementar envelope           │
│ Status: RUNNING  ⏱️ 2m 14s              │
│ Attempt: 3/3                            │
└─────────────────────────────────────────┘
              ↓
Run Strip (appears when WU starts):
┌─────────────────────────────────────────┐
│ ▶ run-103 — Attempt 3                   │
│ Started: 10:32:16  Duration: 2m 14s     │
│ Status: RUNNING                         │
│ [View logs] [View artifacts]            │
└─────────────────────────────────────────┘
              ↓
Session Strip:
┌─────────────────────────────────────────┐
│ 🤖 session-042 — Codex-Builder          │
│ Profile: code_worker                    │
│ Heartbeat: ● 2s ago                     │
│ Checkpoint: cp-003                      │
│ [Chat] [Pause] [Stop]                   │
└─────────────────────────────────────────┘
              ↓
Sandbox Strip:
┌─────────────────────────────────────────┐
│ 📦 sandbox-7a3f                         │
│ Container: orch-sandbox-7a3f            │
│ Branch: feat/t-099-event-envelope       │
│ CPU: 45%  Memory: 128MB                 │
│ [View filesystem] [View terminal]       │
└─────────────────────────────────────────┘
```

**Visual rules:**
- These strips are **collapsed by default** (show only one line: Run ID + status)
- **Expand on click** or when something important happens (failure, tool request)
- The WorkUnit node shows a **mini-indicator**: `▶ 1 active session`
- Multiple attempts stack vertically with clear visual separation

**Why it matters:** This is the core execution chain. Collapsing it hides the reality of how OrchestraOS works.

---

### Phase 6: Execution — Checkpoints, ToolRequests, Events

**What happens:** The agent works inside the sandbox, emitting checkpoints, requesting tools, producing artifacts.

**What should appear:**
- **Event stream attached to the Session strip** — scrollable, timestamped
- **Checkpoint markers** on the Session timeline
- **ToolRequest badges** that appear on the Session strip and pulse if pending approval
- **Artifact thumbnails** or links when produced

**Visual (Session strip expanded):**
```
┌─────────────────────────────────────────┐
│ 🤖 session-042 — Codex-Builder          │
│                                         │
│ ─── Event Stream ───                    │
│ 10:32:16 ▶ Run started                  │
│ 10:32:20 📝 Writing events.go...        │
│ 10:32:45 🧪 Running tests...            │
│ 10:33:10 ✅ Checkpoint: cp-003          │
│ 10:33:30 🔧 Tool: file_write (auto)     │
│ 10:33:45 ⚠️ Tool: bash "go test"        │
│          [Approve] [Reject] [View]      │ ← Pending approval!
│ 10:34:00 📄 Artifact: test-output.log   │
│                                         │
│ [View full transcript]                  │
└─────────────────────────────────────────┘
```

**Why it matters:** The user needs to see what the agent is doing **right now** without opening a separate inspector.

---

### Phase 7: Retry Loop

**What happens:** When a Run fails, the Orchestrator may retry (up to N attempts).

**What should appear:**
- **Failed Run strip** turns red and collapses
- **New Run strip** appears below it with attempt number
- **Failure reason** visible on the failed strip (hover or expand)
- **Retry count indicator** on the WorkUnit node

**Visual:**
```
┌─────────────────────────────────────────┐
│ wu-002 — Implementar envelope           │
│ Status: RUNNING (attempt 3/3)           │
└─────────────────────────────────────────┘
              ↓
┌─────────────────────────────────────────┐
│ ▶ run-101 — Attempt 1                   │
│ Status: ❌ FAILED — timeout after 5m    │
│ [View logs]                             │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│ ▶ run-102 — Attempt 2                   │
│ Status: ❌ FAILED — linter error        │
│ Reason: undefined: uuid (events.go:42)  │
│ [View logs]                             │
└─────────────────────────────────────────┘
┌─────────────────────────────────────────┐
│ ▶ run-103 — Attempt 3 ← CURRENT         │
│ Status: 🟡 RUNNING                      │
│ Duration: 2m 14s                        │
└─────────────────────────────────────────┘
```

**Why it matters:** Retries are a core feature. Hiding them makes failures feel like black boxes.

---

### Phase 8: Human Intervention

**What happens:** User sends a message to a specific agent or approves/denies a tool.

**What should appear:**
- **Intervention marker** on the Session timeline
- The user's message appears as a distinct event (different color/icon)
- Tool approval buttons appear inline in the event stream
- **Intervention badge** on the WorkUnit node

**Visual:**
```
│ 10:31:10 👤 User: "Try using crypto/rand │
│          instead of uuid package"        │
│          → Forwarded to Codex-Builder   │
```

**Why it matters:** Human intervention is a first-class concept, not an afterthought.

---

### Phase 9: Completion & Validation

**What happens:** All WorkUnits complete. Reviews and ValidationGates execute.

**What should appear:**
- **Task node** shows 100% progress ring
- **Validation gate indicators** on each WorkUnit
  - 🛡️ Hard gate passed
  - ○ Soft gate pending
  - △ Policy gate under review
- **Review panel** appears when gates are active
- **Artifact collection** — all outputs gathered

**Visual:**
```
┌─────────────────────────────────────────┐
│ 📋 task-099 — Event Envelope            │
│                                         │
│ Status: ✅ COMPLETE                     │
│ Progress: ████████ 6/6 WUs (100%)       │
│                                         │
│ Validation:                             │
│   🛡️ Hard: 6/6 passed                  │
│   ○ Soft: 5/6 pending (wu-004)         │
│                                         │
│ Artifacts: 12 files, 3 logs, 2 diffs    │
│                                         │
│ [View PR] [Download artifacts]          │
└─────────────────────────────────────────┘
```

---

## 4. Layout Proposal — "Flow-DAG Hybrid"

Instead of a pure DAG (current) or a pure terminal log (rejected), the Canvas should use a **hybrid layout** that preserves both the chronological flow and the structural DAG.

### 4.1 Screen Zones

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ TOP HUD — Task context, filters, global actions                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  LEFT PANEL (20%)        │  CENTER CANVAS (55%)      │  RIGHT PANEL (25%) │
│  ─────────────           │  ─────────────────        │  ────────────────  │
│  Event Stream            │  DAG + Execution Strips   │  Inspector         │
│  (chronological)         │  (structural + live)      │  (details on click)│
│                                                                             │
│  • User message          │  • Task node              │  • WU details      │
│  • Intake                │  • Orchestrator hub       │  • Run history     │
│  • Orchestrator events   │  • WU nodes               │  • Session chat    │
│  • Run starts/completes  │  • Run strips (expand)    │  • Sandbox info    │
│  • Checkpoints           │  • Session strips         │  • Prompt preview  │
│  • Tool requests         │  • Sandbox strips         │  • Artifacts       │
│  • User interventions    │  • Dependency edges       │                    │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│ BOTTOM BAR — Input field, quick actions, status                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 4.2 Interaction Model

| Action | Result |
|--------|--------|
| Click WU node | Expand/collapse Run/Session/Sandbox strips |
| Click Session strip | Open right inspector with full details |
| Click ToolRequest badge | Open approval dialog inline |
| Click UserMessage | Expand to full text |
| Click Orchestrator Hub | Filter event stream to orchestrator events |
| Drag canvas | Pan the DAG |
| Scroll | Zoom or scroll event stream (context-dependent) |
| Type in bottom bar | Send message to Orchestrator or selected agent |

### 4.3 Progressive Disclosure

| Level | What's Visible | How to Access |
|-------|---------------|---------------|
| **Glance** | Task status, WU statuses, active sessions count | Default view |
| **Focus** | Run strips expanded, event stream filtered | Click WU node |
| **Detail** | Full session transcript, sandbox metrics, prompt preview | Click Session strip → Inspector |
| **Debug** | Raw events, replay state, checkpoint JSON | "Debug mode" toggle |

---

## 5. State Transitions — What Changes Visually

### 5.1 Task Lifecycle

```
[CREATED] ──→ [DECOMPOSING] ──→ [SCHEDULING] ──→ [RUNNING] ──→ [VALIDATING] ──→ [COMPLETED]
   │                │                  │               │                │
   │                │                  │               │                │
   ▼                ▼                  ▼               ▼                ▼
 Gray           Amber pulse       Amber solid     Mixed colors     Green
 Task node      Task node +       Orchestrator    WUs running      Progress ring
 appears        "expanding"       hub active      / pending /      full
                animation                         done             Validation
                                                                  gates shown
```

### 5.2 WorkUnit Lifecycle

```
[PENDING] ──→ [READY] ──→ [RUNNING] ──→ [COMPLETED]
   │             │             │               │
   ▼             ▼             ▼               ▼
 Dashed        Solid         Amber           Green
 border        border        border +        border +
 70% opacity   100% opacity  pulse ring      checkmark
```

### 5.3 Run Lifecycle

```
[CREATED] ──→ [STARTING] ──→ [RUNNING] ──→ [COMPLETED]
                                     │
                                     ▼
                               [FAILED] ──→ [RETRYING] ──→ (new Run created)
```

**Visual for retry:**
- Failed Run strip: red, collapsed, failure reason on hover
- New Run strip: amber, appears below with attempt counter
- WorkUnit node: shows `⚠️ Attempt 2/3`

---

## 6. Critical UI/UX Improvements

### 6.1 Must-Have (P0)

| # | Improvement | Why |
|---|-------------|-----|
| 1 | **UserMessage as origin node** | Tasks must have a visible parent |
| 2 | **Orchestrator Hub as central node** | The control plane must be visible |
| 3 | **WU → Run → Session → Sandbox as stacked strips** | Show the real execution chain |
| 4 | **Retry visualization** | Failed attempts must be visible |
| 5 | **Inline tool approval** | Tool requests should not be in a modal |
| 6 | **Event stream connected to sessions** | Live activity must be contextual |

### 6.2 Should-Have (P1)

| # | Improvement | Why |
|---|-------------|-----|
| 7 | **Checkpoint timeline on session strip** | Progress visibility |
| 8 | **Validation gate badges on WU nodes** | Quality visibility |
| 9 | **Artifact thumbnails on session strip** | Output visibility |
| 10 | **Intervention markers in event stream** | Human agency visibility |
| 11 | **Parallel execution animation** | When 2+ WUs run simultaneously |
| 12 | **Sandbox resource metrics (CPU/Memory)** | Operational awareness |

### 6.3 Nice-to-Have (P2)

| # | Improvement | Why |
|---|-------------|-----|
| 13 | **Decomposition animation** | Watch the DAG being built |
| 14 | **Agent "typing" indicator** | When agent is thinking |
| 15 | **Sound for critical events** | Attention when away |
| 16 | **Vim-like keyboard nav** | Power user efficiency |

---

## 7. Anti-Patterns to Avoid

1. **Don't collapse WU/Run/Session/Sandbox into one node.** This hides the architecture.
2. **Don't show the DAG without the Orchestrator.** The Orchestrator is the protagonist.
3. **Don't make tool requests appear in modals.** They are part of the execution timeline.
4. **Don't hide failed attempts.** Retries are a feature, not a bug.
5. **Don't make the event stream a separate panel without context.** Events belong to sessions.
6. **Don't use pure terminal aesthetic.** The DAG needs visual structure, not just text.

---

## 8. Implementation Checklist

### Phase A: Structure (Foundation)
- [ ] Add UserMessage node type
- [ ] Add Intake node type
- [ ] Add Orchestrator Hub node type (larger, distinct)
- [ ] Add Task node with progress ring
- [ ] Add WorkUnit node with status border
- [ ] Add Run/Session/Sandbox strip components

### Phase B: Data Flow (Connections)
- [ ] Draw edges: UserMessage → Intake → Orchestrator → Task
- [ ] Draw edges: Task → TaskGraph → WorkUnits
- [ ] Draw edges: WorkUnit → Run → Session → Sandbox
- [ ] Draw dependency edges between WorkUnits
- [ ] Animate edge state (pending → active → done)

### Phase C: Interactivity (Behavior)
- [ ] Click WU → expand/collapse strips
- [ ] Click Session → open inspector
- [ ] Click ToolRequest → inline approval
- [ ] Bottom bar → send message (to Orchestrator or agent)
- [ ] Event stream → filter by entity

### Phase D: Live Data (Real-time)
- [ ] WebSocket connection for events
- [ ] Real-time status updates
- [ ] Heartbeat indicators
- [ ] Checkpoint markers appearing live
- [ ] Tool request badges appearing live

---

## 9. Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-05-17 | Rejected Terminal/CLI aesthetic | Too sparse; the DAG needs visual structure |
| 2026-05-17 | Adopted Flow-DAG Hybrid layout | Preserves both chronology and structure |
| 2026-05-17 | WU/Run/Session/Sandbox as stacked strips | Shows real execution chain without clutter |
| 2026-05-17 | Orchestrator Hub as central visual element | The control plane is the protagonist |
| 2026-05-17 | Event stream attached to sessions, not global | Contextual events reduce cognitive load |

---

## 10. References

- ADR 0002: Orchestrator Como Control Plane
- ADR 0004: Sandbox e Autonomia Inicial
- ADR 0006: Decomposição de Tasks e Intervenção em Agentes
- ADR 0007: Ciclo Operacional do Agente
- ADR 0009: Observabilidade
- ADR 0016: State Machine Event-Sourced
- ADR 0020: Serviços de Orquestração
- ADR 0023: Hybrid Intelligent Orchestrator Architecture
- `docs/design/CANVAS_IMPROVEMENTS.md`
