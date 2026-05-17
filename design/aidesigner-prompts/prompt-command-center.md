# Prompt: OrchestraOS Command Center

## Context
OrchestraOS is an agent orchestration system. Multiple AI agents execute work units in a directed acyclic graph (DAG). Humans supervise, approve tool requests, and intervene when needed. This is NOT a chat app, NOT an IDE, NOT a kanban board.

## Page Purpose
This is the MAIN DASHBOARD — the first screen users see when opening OrchestraOS. It must answer: "What needs my attention right now? What's running? What's the system health?"

## Layout Structure

### Top Navigation Bar (48px height)
- Left: OrchestraOS logo (minimal, text-based) + Environment badge (e.g., "local" / "prod")
- Center: Global search / command palette trigger
- Right: System status indicator (healthy/degraded) + User avatar

### Primary Content Area (below nav)

#### Section 1: Action Required Queue (horizontal scroll, priority)
A row of compact cards showing items needing human intervention:
- Tool request cards: icon, tool name (file_write, shell.exec), target file/command, risk level (safe/guarded/destructive), agent name, two buttons [Approve] [Reject]
- Review request cards: PR/diff preview thumbnail, agent that created it, [Review] button
- Failure cards: failed work unit, error snippet, [Investigate] button

Each card: 280px wide, dark surface background, left border colored by risk (green/yellow/red).

#### Section 2: Active Tasks (2-column grid on desktop)
Cards showing currently executing tasks:
- Task title + ID (mono)
- Progress: X of Y work units completed
- Mini horizontal progress bar
- Current agent avatar + name
- State badge (running / paused / review)
- Next upcoming checkpoint or pending approval

Card: 400px wide, subtle border, hover reveals "Open Canvas" button.

#### Section 3: System Health (single row, 4 metrics)
- Agents Active: 3/5 (with pulsing dot if < 100%)
- Success Rate: 94% (sparkline mini chart)
- Events/min: 12
- Pending Approvals: 2 (clickable, scrolls to Action Queue)

#### Section 4: Recent Event Stream (right sidebar or bottom panel)
Vertical list of latest events:
- Timestamp (mono, 11px)
- Event type badge (small, colored)
- Entity (task-099, wu-002, codex-builder)
- Description
- Auto-scrolls, newest on top

### Empty States
- No active tasks: "No tasks running. Create a new task to start." + prominent "+ New Task" button
- No pending actions: "All clear. System running smoothly." with subtle green check

## Visual Direction
- Dark mode (#1A1A1F background)
- High information density — show lots of data without clutter
- Cards with 6px border radius, 1px subtle borders
- Functional color accents for states only (blue=running, green=success, red=failure, amber=warning, purple=blocked)
- Typography: Inter for UI, Inter Mono for IDs and timestamps
- No gradients, no glows, no decorative elements
- Professional, calm, authoritative — like Linear + Vercel + GitHub Actions

## Key Constraints
- The ACTION QUEUE is the hero — interventions must be impossible to miss
- Active tasks show PROGRESS, not just status
- System health is glanceable — numbers, not charts
- Everything clickable drills down to Task Canvas
