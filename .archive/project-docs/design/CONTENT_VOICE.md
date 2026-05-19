# OrchestraOS — Content Voice & Terminology

> **Status:** Draft
>
> **Principle:** Clear, concise, technical. The system speaks like a senior engineer — precise, no fluff, but helpful.

---

## Voice & Tone

### We Are
- **Precise:** Every word has a purpose. No filler.
- **Technical:** We use domain terms correctly (agent, work unit, orchestrator, run).
- **Confident:** The system knows what it's doing. No apologetic language.
- **Helpful:** When something fails, we explain why and what to do.

### We Are NOT
- Chatty or casual ("Oops! Something went wrong 🙈")
- Marketing-speak ("Unlock the power of AI orchestration")
- Robotic ("ERROR_CODE_404: RESOURCE_NOT_FOUND")
- Apologetic ("Sorry, we couldn't...")

---

## Writing Rules

### 1. Be Concise
```
❌ "Please click on the button below in order to create a new task"
✅ "New Task"

❌ "The agent is currently in the process of executing the work unit"
✅ "Codex-Builder is running wu-002"
```

### 2. Use Active Voice
```
❌ "The task was completed by the agent"
✅ "Codex-Builder completed task-003"
```

### 3. Front-load Important Info
```
❌ "There are 3 work units waiting for approval"
✅ "3 work units pending approval"
```

### 4. Use Sentence Case (Not Title Case)
```
❌ "New Task Wizard"
✅ "New task"

❌ "Atividade Recente"
✅ "Atividade recente"
```

### 5. No Exclamation Points
```
❌ "Task completed successfully!"
✅ "Task completed"
```

### 6. No Please / Thank You
This is a tool, not a conversation.
```
❌ "Please select an agent"
✅ "Select an agent"
```

---

## Domain Terminology

### Core Concepts (Always Use These)

| Term | Definition | Never Call It |
|---|---|---|
| **Agent** | Autonomous worker that executes tasks | Bot, worker, node, AI |
| **Orchestrator** | Central supervisor that routes work | Controller, master, hub |
| **Task** | Top-level objective given to the system | Job, project, request |
| **Work Unit (WU)** | Atomic unit of work within a task | Step, subtask, job |
| **Run** | Single execution instance of a work unit | Execution, attempt, iteration |
| **Session** | Lifecycle of an agent from start to stop | Connection, instance |
| **Canvas** | Main graph visualization interface | Dashboard, view, board |
| **Inspector** | Floating detail panel for selected items | Sidebar, details, info |
| **Dock** | Left-bottom navigation HUD | Sidebar, menu, nav |

### Status Labels (Exact Copy)

| Status | UI Label | Badge Text | Color |
|---|---|---|---|
| `created` | Created | CREATED | info |
| `running` | Running | RUNNING | running (amber) |
| `completed` | Done | DONE | success |
| `failed` | Failed | FAILED | error |
| `pending` | Pending | PENDING | warning |
| `cancelled` | Cancelled | CANCELLED | error |
| `idle` | Idle | IDLE | info |
| `waiting_approval` | Waiting approval | PENDING | warning |

### Actions (Button Labels)

| Action | Button Label | Notes |
|---|---|---|
| Create task | "New task" | Not "Create", not "Add" |
| Start run | "Run" | Verb, imperative |
| Stop agent | "Stop" | Immediate |
| View logs | "Logs" | Noun, concise |
| Approve tool | "Approve" | Single word |
| Reject tool | "Reject" | Single word |
| Cancel operation | "Cancel" | Standard |
| Save changes | "Save" | Standard |
| Delete item | "Delete" | Always with confirmation |

---

## Error Messages

### Structure
```
[What happened] + [Why it matters] + [What to do]
```

### Examples
```
❌ "Error!"
✅ "Run failed: Codex-Builder could not write to events.go. Check file permissions and retry."

❌ "Something went wrong"
✅ "Task-099 blocked: wu-002 depends on wu-001, which failed. Fix wu-001 or remove the dependency."

❌ "Agent disconnected"
✅ "Review-Bot disconnected after 5m timeout. The session will resume when the agent reconnects."
```

### Error Tone Guidelines
- **User errors (invalid input):** Neutral, instructive. "Enter a valid task name."
- **System errors (bugs, crashes):** Transparent, actionable. "Orchestrator could not route task. Retry or check system logs."
- **Agent errors (agent failure):** Specific, contextual. "Codex-Builder: linter failed on events.go. Check syntax and retry."

---

## Empty States

### Structure
```
[Icon] + [Headline] + [Description] + [Action, optional]
```

### Examples
```
Inspector (nothing selected):
  Icon: Target crosshair
  Headline: "Nothing selected"
  Description: "Click a node in the graph to inspect details."

No tasks:
  Icon: Empty box
  Headline: "No tasks"
  Description: "Create a task to start orchestrating."
  Action: "New task" button

No runs for agent:
  Icon: Clock
  Headline: "No runs yet"
  Description: "This agent hasn't executed any work units."
```

---

## Time & Numbers

### Time Format
```
Absolute: 10:30:14 UTC (always UTC, always 24h)
Relative: 2m ago, 14h ago, 3d ago
Duration: 02h 14m, 45s, 1d 06h
```

### Number Format
```
Counts: 1,234 (comma separator)
Percentages: 33% (no space)
File sizes: 1.2 MB, 456 KB
IDs: lowercase with hyphens (task-099, wu-002, run-101)
```

---

## Localization Notes

Primary language: **English** for system internals, IDs, logs.
UI can be **Portuguese** or bilingual, but:
- Domain terms stay in English: Agent, Task, Work Unit, Run, Orchestrator
- Descriptions and actions can be localized
- Never translate IDs: `task-099` is always `task-099`, never `tarefa-099`

---

## Anti-Patterns

1. **No emojis in production UI.** Use icons instead.
2. **No ALL CAPS except badges.** Badges are uppercase; everything else is sentence case.
3. **No vague error messages.** Every error must be actionable.
4. **No marketing copy in tool UI.** The Canvas is a workspace, not a landing page.
5. **Don't personify the system.** "OrchestraOS could not..." not "I couldn't..."
