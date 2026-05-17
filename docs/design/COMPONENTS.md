# OrchestraOS — Component Specifications

> **Status:** Draft
>
> **Principle:** Each component is defined by its purpose, anatomy, states, and behavior — not by code.

---

## Button

### Variants

| Variant | Background | Border | Text | Usage |
|---|---|---|---|---|
| **Primary** | `accent` (#F59E0B) | none | White, weight 500 | Main action in a context (Run, Confirm, Send) |
| **Secondary** | `bg-panel` | `text-muted` @ 20% | `text-primary` | Alternative action (Cancel, Logs, Back) |
| **Ghost** | transparent | none | `text-secondary` | Low-emphasis action (Close, Dismiss) |
| **Danger** | `status-error` @ 10% | `status-error` @ 30% | `status-error` | Destructive action (Delete, Stop, Kill) |

### Anatomy
```
Height: 32px - 36px
Padding: 0 16px (compact: 0 12px)
Border radius: radius-md (8px) or radius-sm (4px) for compact
Font: text-sm, weight 500
Icon + Text gap: 8px
```

### States
- **Default:** As defined above
- **Hover:** Lighten background 10%, cursor pointer
- **Active:** Scale 0.98, darken background slightly
- **Disabled:** Opacity 50%, cursor not-allowed, no hover effect
- **Loading:** Spinner replaces icon/text, disabled interaction

---

## Input / Text Field

### Anatomy
```
Background: bg-input
Border: 1px solid text-muted @ 20%
Border radius: radius-md (8px)
Padding: 10px 14px
Font: text-sm, text-primary
Placeholder: text-secondary @ 50%
```

### States
- **Default:** Border `text-muted` @ 20%
- **Focus:** Border `accent` @ 40%, subtle glow
- **Error:** Border `status-error`, error icon right
- **Disabled:** Background `bg-panel`, text `text-secondary` @ 50%
- **Filled:** Border `text-muted` @ 40%

### Input Bar (Special Variant)
The floating command palette at the bottom of the canvas:
```
Background: bg-main @ 92% + blur(20px)
Border: 1px solid accent @ 15%
Border radius: radius-xl (16px)
Padding: 12px 16px
Left icon: Accent-colored bolt (context indicator)
Right actions: Attach, Voice, Send
Send button: Accent background, white icon
Focus: Border accent @ 40%, glow-input
```

---

## Card (Node)

### Agent Node Card
```
Background: bg-panel @ 90% + backdrop-blur
Border: 1px solid (varies by state)
Border radius: radius-lg (12px)
Padding: 16px
Shadow: none (default), glow-active (running/selected)
Left accent bar: 4px solid (running=accent, done=success, error=error)

Content:
  - Icon area: 40px x 40px, rounded-lg, tinted background
  - Title: text-sm, weight 600, text-primary
  - Status: text-2xs, mono, colored by state
```

### Work Unit Node Card
```
Background: bg-panel @ 95%
Border: 1px solid text-muted @ 20%
Border radius: radius-md (8px)
Padding: 12px
Left accent bar: 2px solid (done=success, running=accent)

Content:
  - ID badge: text-2xs, mono, in subtle pill
  - Title: text-xs, weight 500, text-primary
  - Status icon: right side (check, clock, spinner)
```

### Orchestrator Node
```
Shape: Diamond (square rotated 45deg), 128px x 128px
Background: bg-panel @ 90% + backdrop-blur
Border: 1px solid accent @ 50%
Border radius: radius-2xl (24px) on the container
Shadow: glow-active (always, it's the center)
Content: Icon + "ORCH" label, both rotated back -45deg
Label below: text-2xs, mono, accent, in pill badge
```

---

## Badge / Tag / Pill

### Status Badge
```
Background: varies by state @ 10%
Border: 1px solid (same color @ 30%)
Border radius: radius-sm (4px)
Padding: 2px 8px
Font: text-3xs, weight 700, uppercase, tracking-wider
```

### ID Tag
```
Background: bg-main
Border: 1px solid text-muted @ 30%
Border radius: radius-sm (4px)
Padding: 2px 6px
Font: text-2xs, mono, text-secondary
```

### Capability Tag
```
Background: bg-main
Border: 1px solid text-muted @ 20%
Border radius: radius-sm (4px)
Padding: 2px 8px
Font: text-2xs, mono, text-secondary
```

---

## Inspector (Floating Panel)

### Anatomy
```
Width: 320px (fixed, max 360px)
Max height: 420px
Background: bg-panel @ 88% + backdrop-blur(16px)
Border: 1px solid text-muted @ 15%
Border radius: radius-xl (16px)
Shadow: shadow-panel

Header:
  - Height: 40px
  - Border bottom: 1px solid text-muted @ 10%
  - Title: text-xs, weight 600 + icon
  - Live indicator: green dot + "Live" label
  - Close button: X icon, top-right

Content:
  - Padding: 14px
  - Scrollable if content exceeds max-height
  - No scrollbars visible until hover (macOS-style)
```

### Behavior
- Appears when a node is clicked
- Positioned intelligently: right of node by default, left if near edge
- Closes on: X click, Escape key, clicking canvas background
- Content changes instantly when switching between nodes

---

## HUD Panel

Generic floating panel used for top bar, dock, zoom controls:
```
Background: bg-panel @ 88% + backdrop-blur(16px)
Border: 1px solid text-muted @ 15%
Border radius: radius-lg (12px)
Shadow: shadow-panel
Inset highlight: inset 0 1px 0 rgba(255,255,255,0.04)
```

---

## Tooltip

```
Background: bg-panel
Border: 1px solid text-muted @ 20%
Border radius: radius-md (8px)
Padding: 8px 12px
Font: text-xs, text-primary
Max width: 240px
Arrow: 6px triangle, same background/border
Delay: 300ms before appearing
Duration: 150ms fade
```

---

## Modal / Dialog

```
Overlay: bg-main @ 70% + backdrop-blur(4px)
Panel:
  - Background: bg-panel
  - Border: 1px solid text-muted @ 20%
  - Border radius: radius-xl (16px)
  - Shadow: shadow-deep
  - Padding: 24px
  - Max width: 480px

Header: text-xl, weight 700, text-primary
Body: text-sm, text-primary, line-height 1.5
Footer: flex row, gap 12px, right-aligned buttons
```

---

## Toast / Notification

```
Background: bg-panel
Border: 1px solid (varies by type)
Border radius: radius-lg (12px)
Padding: 12px 16px
Shadow: shadow-deep
Position: top-right, stacked vertically, gap 8px
Duration: 4s auto-dismiss

Types:
  - Success: left border success
  - Error: left border error
  - Warning: left border warning
  - Info: left border info
```

---

## Mini-Map

```
Size: 160px x 112px
Background: bg-main @ 85% + backdrop-blur(4px)
Border: 1px solid text-muted @ 20%
Border radius: radius-lg (12px)

Content:
  - Grid: same as main canvas but 10% opacity
  - Nodes: simplified rectangles, colored by state
  - Viewport indicator: 1px amber border, 5% amber fill
```

---

## Divider / Separator

```
Color: text-muted @ 15%
Height: 1px (horizontal) or width: 1px (vertical)
Margin: 8px 0 (compact) or 16px 0 (spacious)
```

---

## Scrollbar

```
Width: 6px
Track: transparent
Thumb: bg-hover (#1C1C1C), radius 3px
Thumb hover: accent (#F59E0B)
```

---

## Anti-Patterns

1. **Never use a sidebar.** All navigation is HUD-overlay or dock. The canvas must remain visible.
2. **Never use a modal for non-blocking info.** Use inspector or toast instead.
3. **Never show more than 3 toasts.** Queue and batch if necessary.
4. **Never make inspector wider than 360px.** It should feel lightweight.
5. **Don't use cards without purpose.** If it doesn't need elevation, use a flat panel.
