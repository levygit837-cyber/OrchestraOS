# Prompt: OrchestraOS Task Canvas

## Context
OrchestraOS is an agent orchestration system. A Task is decomposed into a directed acyclic graph (DAG) of Work Units. Each Work Unit has dependencies on others (blocks, requires_artifact, conflicts_with). Agents execute Work Units in isolated sandboxes. This view shows ONE task in full detail.

## Page Purpose
This is the DRILL-DOWN view for a specific task. It answers: "How is this task structured? Who depends on whom? What is each agent doing? What artifacts were produced?"

## Layout Structure

### Header Bar
- Back arrow ← to Command Center
- Task title + ID (mono)
- Task state badge (running / paused / completed / failed)
- Breadcrumb: OrchestraOS / task-099
- Action buttons: [Pause Task] [Cancel] [Replan]

### Main Canvas Area (center, 70% width)
An INFINITE CANVAS showing the Task Graph as a directed acyclic graph:

#### Work Unit Nodes
Rectangular nodes (~220px × 90px) containing:
- Top: WU ID in mono (wu-001) + state badge (small dot + label)
- Middle: Work unit title (1-2 lines max)
- Bottom: Agent avatar (24px circle) + agent name OR "Unassigned"

Node border color = state color:
- Gray (#5C5C66): idle / pending
- Blue (#007AFF): running
- Green (#00CC66): completed
- Red (#FF3B30): failed
- Purple (#9047FF): blocked (waiting dependency)
- Amber (#FF9500): paused / waiting approval

#### Dependency Edges
Curved bezier lines connecting nodes:
- Solid line: `blocks` (WU A must complete before B starts)
- Dashed line: `requires_artifact` (B needs artifact from A)
- Dotted line: `informs` (A notifies B, no hard block)
- Red line: `conflicts_with` (potential conflict, needs review)

Edges have arrowheads showing direction. Hovering an edge shows tooltip with dependency type.

#### Artifact Nodes
Smaller nodes (~140px × 50px) connected to WUs that produced them:
- Icon representing artifact type (container, schema, diff, test_report)
- Artifact name
- Connected via "produced_by" edge (thin line)

#### Agent Avatars
Circular avatars (28px) positioned ON or NEAR running WUs:
- Pulsing ring animation when agent is executing
- Connected to WU by subtle dashed line
- Hover shows agent name + session status

#### Caminho Crítico
The longest dependency path is subtly highlighted (slightly thicker edges, nodes have subtle glow) so users see what blocks overall completion.

### Detail Panel (right sidebar, 30% width)
Clicking a Work Unit opens this panel with tabs:

**Tab: Overview**
- Objective description
- Acceptance criteria (checklist)
- Owned paths (files/modules)
- Assigned agent + capabilities
- Estimated duration vs actual

**Tab: Checkpoints**
Vertical timeline of checkpoints:
- Checkpoint number + timestamp
- Summary text
- Files read / modified
- Decisions made
- Risks identified
- Next goal suggestion

**Tab: Events**
Filtered event stream for this WU only:
- agent.checkpoint_reached
- tool.requested / tool.approved / tool.executed
- orchestrator.intervention
- All with timestamps and payloads

**Tab: Artifacts**
List of artifacts produced:
- Type icon + name
- Created at
- [Preview] button for viewable artifacts
- [Download] for files

**Tab: Terminal**
Live terminal view of the agent's sandbox (if running):
- Scrollable output
- Command prompt at bottom
- Auto-scroll to newest output

### Bottom Panel (collapsible)
**Intervention Input**
When a WU has a pending tool request:
- Shows request details: tool name, input parameters, risk assessment
- Context: why the agent needs this, what will be affected
- [Approve] [Reject] [Modify] [Add Context] buttons
- Text area for human feedback

## Zoom Levels
- **Zoom 100%**: Individual WUs fully readable, good for 4-8 WUs
- **Zoom 75%**: Compact view, good for 8-15 WUs
- **Zoom 50%**: Bird's eye, shows entire task structure, WUs become color-coded squares
- **Fit to Screen**: Auto-zoom to show all WUs

## Visual Direction
- Dark canvas (#1A1A1F background)
- Nodes float on canvas with subtle drop shadow (no heavy shadows)
- Grid dots (very subtle, #2A2A36) for spatial reference — NOT a technical blueprint grid
- Pan: click-drag on empty canvas
- Zoom: mouse wheel or pinch
- Mini-map in bottom-right corner showing viewport rectangle
- Typography: Inter for labels, Inter Mono for IDs
- No sci-fi elements, no neon, no HUD overlays

## Key Constraints
- The DAG is the HERO — everything else is secondary
- Dependencies MUST be visible — this is the core value prop
- Blocked WUs must clearly show WHY they're blocked (which upstream WU)
- Running WUs must feel ALIVE (pulsing agent, animated edges)
- Failed WUs must draw attention without being alarming
- Artifact nodes must not clutter — collapse when zoomed out
