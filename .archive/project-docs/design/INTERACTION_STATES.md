# OrchestraOS — Interaction States

> **Status:** Draft
>
> **Principle:** Every interactive element has a clear, predictable state transition. Users always know what will happen before they click.

---

## Universal State Definitions

### Default
The resting state. No user interaction.

### Hover
Mouse is over the element.
- **Cursor:** Pointer for clickable, grab for canvas, text for input
- **Visual:** Subtle background lighten, border brighten, or scale up
- **Duration:** 100ms
- **Exception:** Touch devices skip hover entirely

### Focus
Element has keyboard focus (Tab navigation).
- **Visual:** `inner-border-accent` (inset amber border) OR 2px outline
- **Rule:** Focus ring must always be visible. Never hide it.
- **Color:** `accent` (#F59E0B) — same as selection

### Active / Pressed
Element is being clicked/tapped.
- **Visual:** Scale 0.98, background darkens slightly
- **Duration:** Instant (0ms) to 100ms

### Selected
Element is chosen/checked in a group.
- **Visual:** `node-selected` treatment — inset accent border + glow-active
- **Example:** Selected node in graph, active tab, checked checkbox

### Disabled
Element cannot be interacted with.
- **Visual:** Opacity 50%, cursor not-allowed, no hover effects
- **Color:** All colors shift to `text-secondary` @ 50%
- **Rule:** Disabled elements must not respond to any input

### Loading / Processing
Action is in progress.
- **Visual:** Spinner replaces content, or skeleton placeholder
- **Spinner:** 16px-20px, `accent` color, 1s linear rotation
- **Button:** Content hidden, spinner centered, remains clickable-area sized

---

## State Matrix by Component

### Button

| State | Background | Border | Text | Cursor | Shadow/Glow |
|---|---|---|---|---|---|
| Default | As per variant | As per variant | As per variant | pointer | none |
| Hover | Lighten 10% | Lighten 10% | unchanged | pointer | none |
| Active | Darken 5%, scale 0.98 | unchanged | unchanged | pointer | none |
| Focus | unchanged | accent @ 50% | unchanged | pointer | none |
| Disabled | Base @ 50% | Base @ 30% | Base @ 50% | not-allowed | none |
| Loading | Base @ 80% | Base @ 50% | hidden | wait | none |

### Input Field

| State | Border | Background | Text | Glow |
|---|---|---|---|---|
| Default | text-muted @ 20% | bg-input | text-primary | none |
| Hover | text-muted @ 30% | bg-input | text-primary | none |
| Focus | accent @ 40% | bg-input | text-primary | glow-input |
| Filled | text-muted @ 40% | bg-input | text-primary | none |
| Error | status-error | bg-input | text-primary | none |
| Disabled | text-muted @ 10% | bg-panel | text-secondary @ 50% | none |

### Node Card (Graph)

| State | Border | Glow | Scale | Z-Index | Left Bar |
|---|---|---|---|---|---|
| Default | text-muted @ 20-30% | none | 1 | 10 | none or state color |
| Hover | text-muted @ 50% | glow-soft | 1.02 | 20 | unchanged |
| Selected | accent @ 80% | glow-active | 1 | 20 | accent |
| Running | accent-light @ 60% | glow-active | 1 (pulse) | 10 | accent-light |
| Done | text-muted @ 20% | none | 1 | 10 | success |
| Pending | text-muted @ 20% dashed | none | 1 | 10 | none |
| Error | status-error @ 60% | none | 1 | 10 | error |

### Inspector Panel

| State | Opacity | Transform | Pointer Events |
|---|---|---|---|
| Hidden | 0 | scale 0.96, translateY(8px) | none |
| Visible | 1 | scale 1, translateY(0) | auto |
| Transition | 250ms | ease-bounce | — |

### Dock Icon

| State | Background | Text/Icon | Border |
|---|---|---|---|
| Default | transparent | text-secondary | none |
| Hover | bg-hover @ 50% | text-primary | none |
| Active (current page) | accent @ 10% | accent-light | 1px solid accent @ 20% |
| Pressed | accent @ 15% | accent-light | 1px solid accent @ 30% |

---

## Feedback Patterns

### Success Feedback
- **Visual:** Brief green flash or checkmark icon
- **Audio:** Optional subtle chime (if sound enabled)
- **Duration:** 500ms-1s, then returns to default

### Error Feedback
- **Visual:** Red border pulse, error icon, tooltip with message
- **Audio:** Optional subtle alert tone
- **Behavior:** Prevents action completion, returns to editable state

### Loading Feedback
- **Visual:** Spinner in button, skeleton in content area, or progress bar
- **Text:** Action verb changes to present continuous ("Run" → "Running...")
- **Cancellation:** Long operations (>3s) should show cancel button

---

## Keyboard Navigation

### Focus Order
1. Top HUD actions (left to right)
2. Canvas nodes (spatial, nearest first)
3. Dock (top to bottom)
4. Input bar
5. Zoom controls

### Shortcuts (Global)

| Shortcut | Action |
|---|---|
| `Tab` | Next focusable element |
| `Shift + Tab` | Previous focusable element |
| `Enter` | Activate focused element |
| `Escape` | Close inspector, dismiss modal, blur input |
| `Cmd/Ctrl + K` | Focus input bar |
| `+` / `-` | Zoom in / out (when canvas focused) |
| `0` | Reset zoom |
| `?` | Show keyboard shortcuts help |

### Canvas Navigation

| Input | Action |
|---|---|
| `Click node` | Select + open inspector |
| `Click canvas` | Deselect all |
| `Drag canvas` | Pan |
| `Scroll + Ctrl` | Zoom |
| `Double-click node` | Zoom to fit node |
| `Space + drag` | Pan (alternative) |

---

## Touch / Mobile States

- **No hover state.** Use active/pressed instead.
- **Tap = click.** Same behavior as mouse click.
- **Long press** on node = context menu
- **Pinch** = zoom canvas
- **Two-finger drag** = pan canvas
- **Tap empty space** = deselect

---

## Anti-Patterns

1. **Never remove focus ring.** Accessibility is non-negotiable.
2. **Never use color alone for state.** Always pair with icon, text, or shape change.
3. **Don't animate disabled states.** Disabled should feel "frozen," not "broken."
4. **Avoid hover-dependent reveals on touch.** Use tap instead.
5. **Don't show loading on non-actionable elements.** Only buttons and forms should have loading states.
