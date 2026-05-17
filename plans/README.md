# Plans — Execution Plan Structure

This directory contains execution plans for agents. Plans are decomposed into independent, small, verifiable units of work. Each plan follows a standardized serialization convention that enables traceability, parallelization, and controlled re-execution.

## Structure

```text
plans/
├── README.md                  # This file
├── active/                    # Plans in execution
│   └── {ID}-{task-name}/
│       ├── plan.md            # Context, scope, acceptance criteria
│       └── checklist.md       # Execution tracker + delivery artifact
├── completed/                 # Finished plans
│   └── {ID}-{task-name}/
│       ├── plan.md
│       └── checklist.md       # Completed with all items checked
├── templates/                 # Reusable templates
│   ├── plan.md
│   ├── checklist.md
│   └── modulo-go-completo.md
└── archive/                   # Old or superseded plans
```

> **Flat structure:** Prefer `plans/active/{ID}-{task}/` over deep nesting.

## Serialization Convention

Every plan gets a unique, immutable identifier.

Format:
```text
ORCH-{PHASE}-{ROUND}-{AGENT}-{task-name}
```

Example:
```text
ORCH-F05-R01-A01-agentservice
```

| Segment | Meaning | Example |
|---|---|---|
| `ORCH` | Fixed prefix (Orchestrator) | — |
| `{PHASE}` | Phase identifier (F05, F28...) | `F05` = Phase 5 |
| `{ROUND}` | Iteration within phase (R01, R02...) | `R01` = first round |
| `{AGENT}` | Executor agent identifier | `A01` = Agent 1 |
| `{task-name}` | Short scope description (kebab-case) | `agentservice` |

**Full example:** `ORCH-F05-R01-A01-agentservice` = Orchestrator, Phase 5, Round 1, Agent 1, Task: implement AgentService.

## Principle — Decomposed Plans

Each plan represents **a single unit of work** that one agent can execute end-to-end without external coordination. Do not group multiple responsibilities into one plan.

### Characteristics of a valid plan

- **Small scope** — Can be completed in a focused session
- **Independent** — Does not block other plans (or declares explicit dependencies)
- **Verifiable** — Has clear, testable acceptance criteria
- **Isolated** — Defines code boundaries (TOUCH / AVOID)
- **Serialized** — Unique ID enables tracking in canvas, ADRs, and commits

## Creating a New Plan

Use the scaffold script:

```bash
./scripts/new-plan.sh {ID} {task-name} [agent-id]
```

Example:
```bash
./scripts/new-plan.sh ORCH-F05-R01-A01 agentservice agent-1
```

This generates:
- `plans/active/ORCH-F05-R01-A01-agentservice/plan.md`
- `plans/active/ORCH-F05-R01-A01-agentservice/checklist.md`

Then edit both files with task-specific details.

## Plan Lifecycle

### 1. Creation (Orchestrator)

```text
plans/active/{ID}-{task}/
  → plan.md      (context, scope, boundaries, acceptance criteria)
  → checklist.md (pending items, unmarked)
```

### 2. Execution (Executor Agent)

1. Read the plan
2. Locate or create the checklist
3. Execute iteratively: read → execute → validate → update checklist
4. Add annotations with timestamps as needed

### 3. Completion

When all checklist items are checked and annotations are up to date:

- Move the directory from `plans/active/` to `plans/completed/`
- The checklist remains in final state (all `[x]`) with delivery section filled

## Rules

1. **1 Plan = 1 Agent = 1 Micro-task** — Never put prompts for multiple agents in the same plan
2. **Checklist accompanies the plan** — Every plan has its `checklist.md` in the same directory
3. **Move to completed when done** — Transfer the entire directory to `plans/completed/`
4. **Do not edit plans in execution** — If changes are needed, create a new round (R02, R03...)
5. **Commit via safe-commit** — Use `./scripts/safe-commit.sh` after significant cycles
