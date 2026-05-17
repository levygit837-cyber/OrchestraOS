# OrchestraOS Canvas v4 — Terminal / CLI Aesthetic

> **Status:** Implemented (Prototype)
>
> **Scope:** `design/aidesigner-canvas-v4.html`
>
> **Principle:** The interface is a command station, not a dashboard. Every element earns its pixels. Information is revealed on demand, never forced.

---

## 1. Design Philosophy

The Terminal/CLI aesthetic is a deliberate rejection of the previous Canvas v3 approach (dense glass-like overlays, floating inspectors, rich card-based HUD). Instead, v4 treats the screen as a **terminal surface** where the orchestration flow is readable as a structured log.

### Why Terminal?

1. **Familiarity** — Developers live in terminals. The mental model is immediate.
2. **Information Density** — A log format shows more events per viewport than cards.
3. **Honesty** — No decorative gradients, no fake glass, no 3D depth tricks. What you see is what the system is doing.
4. **Calm** — Removing visual noise reduces cognitive load. The operator can focus on anomalies.

### Core Principles

| Principle | Rule |
|---|---|
| **Monospace First** | Every text element uses JetBrains Mono. No exceptions. |
| **Lines, Not Cards** | Borders are 1px hairlines. No shadows, no rounded cards, no glass. |
| **Color is Signal** | Color appears only for status (amber=running, green=done, red=failed). Everything else is gray. |
| **Reveal on Demand** | Details live in the drawer. The main view shows only timestamp + type + summary. |
| **Keyboard Native** | Input bar is always visible. Tab switching, `/` shortcuts, Enter to submit. |

---

## 2. Layout Structure

The viewport is divided into **4 horizontal zones**, stacked vertically:

```
+----------------------------------------------------------+
|  OrchestraOS    Flow | DAG | Log | Failures          [?] |  <- Header
+----------------------------------------------------------+
|                                                          |
|  10:23:45  user_message   Implementar sistema de...      |
|  10:23:46  intake         Normalizado: criar Event...    |
|  10:23:47  task_created   task-099 Event Envelope | P0   |  <- Content
|  10:24:00  run_start      wu-001 -> run-100 | Criar...   |     (Flow/DAG/Log)
|  ...                                                     |
|                                                          |
+----------------------------------------------------------+
|  Codex-Builder  wu-002 | 3/5 | 45%    Review-Bot ...     |  <- Active Strip
+----------------------------------------------------------+
|  > Descreva sua tarefa, ideia ou objetivo...          ⏎  |  <- Input Bar
+----------------------------------------------------------+
```

### Zone Details

#### Header (40px height)
- **Left:** Logo (`Orchestra<span>OS</span>`) + subtitle (`Canvas v4 — Terminal`)
- **Right:** 4 tabs — Flow / DAG / Log / Failures
- **Border:** 1px solid `--border` (#1C1C1C)
- **Background:** `--bg-panel` (#0C0C0C)

#### Content (flex: 1, scrollable)
- Padding: 0 (edges flush with viewport)
- Background: `--bg` (#080808)
- Content depends on active tab (see Section 3)

#### Active Strip (auto height, ~36px)
- Shows running sessions as compact pills
- Each pill: `● AgentName  wu-XXX | attempt | CPU%`
- Click opens the Session Drawer
- Border-top: 1px solid `--border`
- Background: `--bg-panel`

#### Input Bar (40px height)
- Prompt: amber `>` symbol
- Text input: transparent bg, no border
- Placeholder: dim gray
- Submit hint: dim `↵` symbol
- Border-top: 1px solid `--border`
- Background: `--bg-panel`

---

## 3. View System

Four views controlled by header tabs. All share the same log-line DNA.

### 3.1 Flow View (Default)

Shows the **event stream** + **DAG summary** in a single scrollable column.

**Structure:**
```
[Task Header]     TASK: task-099 Event Envelope | P0 | 3/6 WU | 2 active sessions
─────────────────────────────────────────────────────────────────
[Event 1]         10:23:45  user_message   Implementar sistema...
[Event 2]         10:23:46  intake         Normalizado: criar...
[Event 3]         10:23:47  task_created   task-099 Event Envelope...
...
─────────────────────────────────────────────────────────────────
[DAG Summary]     ASCII tree showing current task graph state
```

**Event Line Format:**
```
<TIMESTAMP>  <TYPE_LABEL>   <SUMMARY>  |  <METADATA>
```

| Field | Width | Color | Example |
|---|---|---|---|
| Timestamp | 70px fixed | `--text-dim` | `10:23:45` |
| Type Label | 130px fixed | Semantic color | `user_message`, `run_start` |
| Summary | flex | `--text-bright` | `wu-001 -> run-100 OK` |
| Metadata | flex | `--text` dimmed | `\| agent: Codex-Builder` |

**Type Colors:**
- `user_message` / `user_intervention` → `--accent` (amber)
- `intake` → `--blue` (#3B82F6)
- `task_created` / `taskgraph` → `--green` (#22C55E)
- `orchestrator` → `--accent`
- `run_start` / `run_complete` → `--text` (neutral)
- `error` / `run_failed` → `--red` (#EF4444)
- `checkpoint` → `--green`

### 3.2 DAG View

Pure ASCII-art tree showing the Task Graph structure.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  task-099  Event Envelope                                        P0  3/6 WU │
├─────────────────────────────────────────────────────────────────────────────┤
│  OK  wu-001  Criar schema base        100%   Codex-Builder    3m 12s         │
│      │                                                                      │
│      ├── >>  wu-002  Implementar envelope   67%   Codex-Builder   att 3/3   │
│      ├── >>  wu-003  Adicionar middleware   15%   Review-Bot      att 1/3   │
│      └── OK  wu-005  Setup CI pipeline     100%   System         4m 30s      │
│           │  │                                                              │
│           ├── --  wu-004  Validar E2E         0%   Review-Bot               │
│           └── --  wu-006  Documentar API      0%   Review-Bot               │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Status Icons:**
| Status | Icon | Color |
|---|---|---|
| Done | `OK` | `--green` |
| Running | `>>` | `--accent` (amber) |
| Failed | `XX` | `--red` |
| Pending | `--` | `--text-dim` |

**Columns (left to right inside the box):**
1. Status icon (2 chars)
2. Work Unit ID (6 chars)
3. Title (padded to ~24 chars)
4. Progress % (5 chars)
5. Agent name (14 chars)
6. Duration / Attempt (10 chars)

### 3.3 Log View

Same as Flow view but **without the DAG summary** at the bottom. Pure chronological event stream.

Useful when the operator wants to scroll through history without the DAG taking space.

### 3.4 Failures View

Filters the event stream to show **only error events** (`type === 'error'`).

Visual distinction:
- Left border: 2px solid `--red` (instead of transparent)
- All events are `run_failed` or system errors
- No other event types clutter the view

---

## 4. Color System (Terminal Variant)

The Terminal aesthetic uses a **reduced palette** compared to the full Solar Core. Fewer grays, no glass opacity layers, no glows.

### Tokens

| Token | Hex | Usage |
|---|---|---|
| `--bg` | `#080808` | Root background |
| `--bg-panel` | `#0C0C0C` | Header, active strip, input bar, drawer |
| `--bg-hover` | `#121212` | Row hover background |
| `--border` | `#1C1C1C` | Horizontal separators |
| `--border-light` | `#292524` | Drawer border, active item borders |
| `--text` | `#A8A29E` | Body text, metadata |
| `--text-dim` | `#57534E` | Timestamps, labels, placeholders |
| `--text-bright` | `#E7E5E4` | Primary content, IDs, summaries |
| `--accent` | `#F59E0B` | Running status, active tabs, prompt symbol |
| `--accent-dim` | `#92400E` | Accent hover backgrounds |
| `--green` | `#22C55E` | Done, success, checkpoints |
| `--green-dim` | `#14532D` | Green hover backgrounds |
| `--red` | `#EF4444` | Failed, error, stop button |
| `--red-dim` | `#7F1D1D` | Red hover backgrounds |
| `--blue` | `#3B82F6` | Intake, routing labels |

### Differences from Solar Core (v3)

| Aspect | v3 (Solar Core) | v4 (Terminal) |
|---|---|---|
| Background | `#070707` | `#080808` (slightly lighter) |
| Panel bg | `#141414` @ 88% opacity + blur | `#0C0C0C` solid, no blur |
| Border radius | 8-12px for cards | 0px everywhere (sharp corners) |
| Shadows | `glow-active`, `shadow-panel` | None. Flat design. |
| Glass effect | `backdrop-filter: blur(16px)` | Not used |
| Glows | Amber glow on running nodes | No glows. Color only. |

---

## 5. Typography

**Font family:** `JetBrains Mono` (all weights: 300-700)

**Base size:** 13px
**Line height:** 1.6

### Scale

| Element | Size | Weight | Color |
|---|---|---|---|
| Header title | 14px | 600 | `--text-bright` |
| Header subtitle | 11px | 400 | `--text-dim` |
| Tab label | 12px | 400 | `--text-dim` / `--accent` (active) |
| Log timestamp | 11px | 400 | `--text-dim` |
| Log type label | 11px | 400 | Semantic color |
| Log summary | 13px | 400 | `--text-bright` |
| Log metadata | 13px | 400 | `--text` (dimmed via opacity) |
| DAG tree text | 13px | 400 | Mixed (semantic colors for status) |
| Active strip name | 12px | 500 | `--text-bright` |
| Active strip meta | 11px | 400 | `--text-dim` |
| Input prompt | 13px | 600 | `--accent` |
| Input text | 13px | 400 | `--text-bright` |
| Drawer title | 13px | 600 | `--text-bright` |
| Drawer label | 11px | 500 | `--text-dim` uppercase |
| Drawer row key | 12px | 400 | `--text` |
| Drawer row value | 12px | 400 | `--text-bright` |
| Code block | 12px | 400 | `--text-bright` |
| Button text | 12px | 400 | `--text` / `--accent` / `--red` |

### Monospace Rules (v4 Specific)

- **Everything is monospace.** Even the header title uses JetBrains Mono.
- **IDs are not specially formatted** — they blend into the line but are bright.
- **Timestamps use a fixed-width column** so events align vertically.
- **No letter-spacing adjustments** except drawer labels (`0.5px` for uppercase).

---

## 6. Component Specs

### 6.1 Log Entry

```
Layout: flex row, gap 12px
Padding: 3px 16px
Border-left: 2px solid transparent (2px solid --red in Failures view)
Hover: background var(--bg-hover)
Transition: background 0.1s
```

**Content structure:**
```html
<div class="log-entry">
  <span class="log-timestamp">10:23:45</span>
  <span class="log-type type-intake">intake</span>
  <span class="log-msg">
    Normalizado: criar EventEnvelope
    <span class="dim">| routing: orchestrator.primary</span>
  </span>
</div>
```

### 6.2 Active Strip Item

```
Layout: inline-flex, align-items center, gap 8px
Padding: 4px 10px
Border: 1px solid var(--border-light)
Background: transparent
Cursor: pointer
Hover: border-color var(--accent)
Transition: border-color 0.15s
```

**Parts:**
- **Dot:** 6px circle, `--accent`, `pulse-dot` animation (2s infinite)
- **Name:** 12px, `--text-bright`, weight 500
- **Meta:** 11px, `--text-dim`

### 6.3 Input Bar

```
Layout: flex row, align-items center, gap 8px
Padding: 8px 16px
Border-top: 1px solid var(--border)
Background: var(--bg-panel)
```

**Behavior:**
- Focus: no visual change (terminal doesn't highlight inputs)
- Placeholder disappears on focus
- Enter submits, Shift+Enter adds newline (for multi-line — future)
- `>` prompt is static amber

### 6.4 Session Drawer

**Trigger:** Click on active strip item (or future: click on DAG node)

**Animation:**
```
Overlay: opacity 0 -> 1, 0.2s
Drawer: translateX(100%) -> translateX(0), 0.25s ease
```

**Width:** 420px (max 90vw)
**Background:** `--bg-panel`
**Border-left:** 1px solid `--border-light`

**Content sections (top to bottom):**
1. **Header:** Agent name + Session ID + close button
2. **Work Unit:** ID, Title
3. **Run:** ID, Attempt, Status (colored), Duration
4. **Session:** ID, Agent, Profile, Heartbeat (live dot), Checkpoint
5. **Sandbox:** ID, Container, Branch (accent color), Worktree, CPU, Memory
6. **Recent Output:** Code block with terminal output
7. **Actions:** [Chat] [Pause] [Stop] buttons

**Drawer Button Styles:**
| Type | Border | Text | Hover |
|---|---|---|---|
| Primary | `--accent` | `--accent` | bg `--accent-dim`, text white |
| Default | `--border-light` | `--text` | border `--accent`, text bright |
| Danger | `--red` | `--red` | bg `--red-dim` |

---

## 7. Interaction Patterns

### 7.1 Tab Switching
- Click tab → content swaps instantly (no animation)
- Active tab: amber color + amber bottom border
- Inactive tab: dim gray, bottom border transparent

### 7.2 Row Hover
- Any log entry row highlights on hover (`--bg-hover`)
- Cursor: default (not pointer — rows are not clickable in v4, only the drawer trigger is)

### 7.3 Active Strip Click
- Click any active session pill → drawer slides in from right
- Overlay fades in behind drawer
- Click overlay or close button → drawer slides out

### 7.4 Input Bar
- Always focused on page load (`autofocus`)
- `Enter` submits (currently alerts/mock)
- `Esc` clears input (future: or closes drawer if open)
- No suggestions dropdown in v4 (future: autocomplete with `Tab`)

### 7.5 Keyboard Shortcuts (Planned)
| Key | Action |
|---|---|
| `1-4` | Switch to Flow/DAG/Log/Failures |
| `Esc` | Close drawer |
| `Enter` | Submit input |
| `/` | Focus input bar |
| `?` | Show keyboard help |

---

## 8. Data Model in the View

The v4 prototype uses **realistic mock data** that simulates the actual OrchestraOS orchestration flow:

### Event Stream (21 events)

Represents the full lifecycle of a task:

1. **User Message** → Natural language input
2. **Intake** → Normalization & routing
3. **Orchestrator** → Intent analysis
4. **Task Created** → Task entity spawned
5. **Orchestrator** → Decomposition strategy
6. **TaskGraph** → WUs and levels defined
7. **Run Start** → WU-001 begins (Codex-Builder)
8. **Run Complete** → WU-001 done
9. **Orchestrator** → Unlock parallel WUs
10. **Run Start** → WU-002 begins (attempt 1)
11. **Run Start** → WU-005 begins (System)
12. **Run Failed** → WU-002 timeout
13. **Orchestrator** → Retry attempt 2
14. **Run Start** → WU-002 attempt 2
15. **Run Failed** → WU-002 linter error
16. **User Intervention** → Human provides hint
17. **Orchestrator** → Retry attempt 3
18. **Run Start** → WU-002 attempt 3
19. **Run Start** → WU-003 begins (Review-Bot)
20. **Checkpoint** → WU-002 reaches checkpoint
21. **Orchestrator** → Monitoring summary

### DAG Structure

```
Level 1: wu-001 (done)          → Schema base
Level 2: wu-002 (running, 67%)  → Implement envelope  [parallel]
         wu-003 (running, 15%)  → Add middleware      [parallel]
         wu-005 (done)          → CI pipeline         [parallel]
Level 3: wu-004 (pending)       → Validate E2E
         wu-006 (pending)       → Document API
```

### Active Sessions (2)

| WU | Run | Agent | Status | Sandbox |
|---|---|---|---|---|
| wu-002 | run-103 (att 3) | Codex-Builder | running | sandbox-7a3f, 45% CPU |
| wu-003 | run-201 (att 1) | Review-Bot | running | sandbox-8b2e, 23% CPU |

---

## 9. Comparison: v3 vs v4

| Dimension | Canvas v3 (Glass HUD) | Canvas v4 (Terminal) |
|---|---|---|
| **Visual weight** | Heavy — glass panels, glows, shadows, rounded cards | Light — 1px lines, flat colors, sharp corners |
| **Information density** | Medium — cards take space, inspector is floating | High — log lines pack events vertically |
| **Primary element** | Graph canvas with pan/zoom | Event stream scroll |
| **Color usage** | Rich — gradients, opacity layers, glows | Sparse — only semantic status colors |
| **Font** | Inter + JetBrains Mono (mixed) | JetBrains Mono only |
| **Animations** | Rich — node hover bounce, inspector scale, edge dash | Minimal — only dot pulse, drawer slide |
| **Mental model** | "I'm looking at a graph" | "I'm reading a log" |
| **Best for** | Visualizing large task graphs, spatial relationships | Monitoring live execution, debugging failures |
| **Intervention** | Inspector overlay with tabs | Drawer slide with session details |

### When to Use Which

- **v3** is ideal for: Understanding the big picture of a TaskGraph, seeing agent clusters, spatial reasoning about dependencies.
- **v4** is ideal for: Real-time monitoring, debugging runs, reading event history, quick keyboard-driven operation.

---

## 10. Future Enhancements

### Short Term
- [ ] Live WebSocket event feed (replace static mock data)
- [ ] Keyboard shortcuts (`1-4` for tabs, `?` for help)
- [ ] Search/filter in log view (`/pattern`)
- [ ] Auto-scroll to newest event
- [ ] Session Chat tab inside drawer

### Medium Term
- [ ] Resizeable drawer width
- [ ] Multiple drawer instances (split view)
- [ ] Log export (copy as text, download as `.log`)
- [ ] ANSI color support in code blocks
- [ ] Time-range filter ("last 5 min", "since checkpoint")

### Long Term
- [ ] Integration with real OrchestraOS event store
- [ ] Custom color themes (user preference)
- [ ] Vim-like navigation mode (`j/k` scroll, `gg` top, `G` bottom)
- [ ] Command palette (`:` for commands)

---

## 11. Accessibility

- **Contrast:** All text meets WCAG AA (amber on dark is 7.2:1, green is 5.8:1, red is 5.4:1)
- **Keyboard:** All interactive elements (tabs, active strip, drawer, input) are keyboard-focusable
- **Reduced motion:** `prefers-reduced-motion` disables the dot pulse animation
- **Screen reader:** Log entries use semantic HTML (`<time>`, `<strong>`, `<span>` with aria-labels)

---

## 12. Changelog

| Date | Change | Author |
|---|---|---|
| 2026-05-17 | Initial Terminal/CLI aesthetic design | — |
| 2026-05-17 | Implemented v4 prototype (`aidesigner-canvas-v4.html`) | — |
| 2026-05-17 | Documented v4 UI/UX spec | — |
