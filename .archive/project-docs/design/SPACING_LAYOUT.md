# OrchestraOS — Spacing & Layout System

> **Status:** Draft
>
> **Principle:** Dense but breathable. Information-rich without feeling cramped. Every pixel has purpose.

---

## Spacing Scale

Base unit: **4px** (`0.25rem`)

| Token | Value | Usage |
|---|---|---|
| `space-px` | 1px | Hairline borders, dividers |
| `space-0.5` | 2px | Icon gaps, tight internal spacing |
| `space-1` | 4px | Default gap unit, icon padding |
| `space-1.5` | 6px | Small component padding |
| `space-2` | 8px | Compact padding, small gaps |
| `space-2.5` | 10px | Tag padding, small buttons |
| `space-3` | 12px | Standard component padding |
| `space-4` | 16px | Card padding, section gaps |
| `space-5` | 20px | Medium section spacing |
| `space-6` | 24px | Large component padding |
| `space-8` | 32px | Section separation |
| `space-10` | 40px | Major section gaps |
| `space-12` | 48px | Page-level spacing |
| `space-16` | 64px | Hero / canvas margins |

### Practical Examples

```
Node Card (Agent):     padding: 16px (space-4), gap: 12px (space-3)
Node Card (WU):        padding: 12px (space-3), gap: 8px (space-2)
HUD Panel:             padding: 12px - 16px
Inspector:             padding: 14px (space-3.5)
Input Bar:             padding: 12px 16px, gap: 12px
Dock Icon:             padding: 8px, gap: 4px between icons
```

---

## Border Radius Scale

| Token | Value | Usage |
|---|---|---|
| `radius-sm` | 4px | Tags, badges, small buttons, status indicators |
| `radius-md` | 8px | Input fields, small cards, WU nodes |
| `radius-lg` | 12px | Agent nodes, panels, inspector |
| `radius-xl` | 16px | Modals, large panels |
| `radius-2xl` | 24px | Orchestrator node (diamond shape container) |
| `radius-full` | 9999px | Avatar circles, pills, dot indicators |

---

## Z-Index Layers

| Layer | Z-Index | Elements |
|---|---|---|
| Canvas Background | 0 | Grid, graph edges |
| Canvas Nodes | 10 | Node cards |
| Canvas Connections | 0 (behind nodes) | SVG paths |
| Floating UI | 40 | Input bar, zoom controls, dock |
| HUD Overlays | 40-50 | Top bar, notifications |
| Inspector | 50 | Floating inspector panel |
| Modals / Overlays | 60 | Dialogs, confirmations |
| Toasts | 70 | Notifications, alerts |
| Tooltip | 80 | Hover tooltips |
| Context Menu | 90 | Right-click menus |

---

## Canvas Layout

### Viewport Behavior

```
Canvas:           100vw x 100vh, overflow: hidden
Inner Graph:      1400px x 900px (virtual canvas), centered
Minimap:          160px x 112px, bottom-right, 16px from edges
Input Bar:        max-width: 672px, bottom-center, 24px from bottom
Dock:             bottom-left, 20px from edges
Top HUD:          full-width, 56px height, 20px horizontal padding
```

### Safe Zones

Elements must not overlap these zones:

```
Top:      56px (HUD bar height + 8px padding)
Bottom:   80px (input bar + hints + margin)
Left:     56px (dock width + margin)
Right:    180px (minimap + zoom controls + margin)
```

---

## Grid System

### Canvas Grid
- **Size:** 48px cells
- **Color:** `rgba(120,113,108,0.03)` (barely visible warm gray)
- **Purpose:** Spatial orientation, subtle depth
- **Vignette:** Radial gradient from transparent center to `bg-canvas` at edges

### UI Grid
- **No rigid column grid.** The HUD paradigm uses floating panels positioned by context.
- **Spacing between HUD elements:** 8px to 12px
- **Panel internal spacing:** 12px to 16px

---

## Responsive Breakpoints

| Breakpoint | Width | Behavior |
|---|---|---|
| `sm` | 640px | Minimum supported width. Canvas still full-screen. Inspector may overlay center. |
| `md` | 768px | Tablet. Dock collapses to icons-only. Input bar shrinks. |
| `lg` | 1024px | Desktop. Full layout as designed. |
| `xl` | 1280px | Large desktop. Comfortable spacing. |
| `2xl` | 1536px | Ultra-wide. Canvas centered, more whitespace on sides. |

---

## Sizing Reference

### Common Component Sizes

| Component | Width | Height |
|---|---|---|
| Top HUD bar | 100% | 56px |
| Dock button | 36px | 36px |
| Zoom button | 32px | 32px |
| Input bar | 100% (max 672px) | auto (min 48px) |
| Inspector | 320px | auto (max 420px) |
| Mini-map | 160px | 112px |
| Agent node | 192px | auto |
| WU node | 200px | auto |
| Orchestrator node | 128px x 128px | — |
| Task chip | auto | 36px |
| Icon button (HUD) | 36px | 36px |
| Status badge | auto | 20px |

---

## Anti-Patterns

1. **Never use magic numbers.** Always use the spacing scale tokens.
2. **Never let floating panels touch screen edges.** Minimum 16px-20px margin.
3. **Never make the inspector wider than 360px.** It should feel like a tooltip, not a sidebar.
4. **Don't center-align text in technical UI.** Left-align everything except numbers in tables.
5. **Avoid excessive whitespace.** This is a dense tool, not a landing page.
