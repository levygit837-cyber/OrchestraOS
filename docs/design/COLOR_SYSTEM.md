# OrchestraOS — Color System

> **Palette Name:** Solar Core (Obsidian & Amber Gold)
>
> **Purpose:** Establish the visual identity for all OrchestraOS graphical interfaces — Canvas, Dashboard, CLI TUI, and any future UI surfaces.
>
> **Status:** Adopted (2026-05-17)

---

## Design Philosophy

The Solar Core palette replaces the generic "tech blue" cliché with a warm, high-impact amber-on-obsidian scheme. Amber/gold is the color of **VIP status**, **controlled energy**, and **elite command systems**. In dark mode, it creates superior emotional contrast and immediately draws the eye to what matters: the active state.

**References:** DaVinci Resolve, high-frequency trading terminals, aerospace control interfaces.

---

## Core Palette

### Backgrounds

| Token | Hex | Usage |
|---|---|---|
| `bg-canvas` | `#070707` | Full-screen canvas background. The deepest layer. |
| `bg-main` | `#0A0A0A` | Application root background. |
| `bg-panel` | `#141414` | Floating panels, HUD elements, cards, inspectors. |
| `bg-hover` | `#1C1C1C` | Hover states on interactive surfaces. |
| `bg-input` | `#0A0A0A` | Input fields, code blocks, terminal surfaces. |

### Text

| Token | Hex | Usage |
|---|---|---|
| `text-primary` | `#E7E5E4` | Headings, active labels, primary content. Warm near-white. |
| `text-secondary` | `#78716C` | Descriptions, hints, disabled text, timestamps. |
| `text-muted` | `#44403C` | Borders, dividers, subtle separators. |

### Accent (Amber)

| Token | Hex | Usage |
|---|---|---|
| `accent` | `#F59E0B` | Primary accent. Active states, selection, primary buttons, live indicators. |
| `accent-light` | `#FBBF24` | Glows, highlights, running animations, pulse rings. |
| `accent-dim` | `rgba(245, 158, 11, 0.15)` | Subtle borders, hover backgrounds on accent elements. |

### Semantic Status Colors

| State | Token | Hex | Visual Treatment |
|---|---|---|---|
| **Success / Done** | `status-success` | `#22C55E` | Solid left border, checkmark icon. No glow. |
| **Running / Active** | `status-running` | `#FBBF24` | Amber glow, pulse animation, dashed animated edges. |
| **Warning / Pending** | `status-warning` | `#F97316` | Dashed border, muted opacity (~70%), clock icon. |
| **Error / Failed** | `status-error` | `#EF4444` | Red border, error icon. Use sparingly. |
| **Info / Idle** | `status-info` | `#A8A29E` | Neutral tone. Used for idle agents and system info. |

---

## Glow & Shadow System

All glows use the amber accent. Avoid cool-toned shadows or blue-tinted blurs.

| Token | Value | Usage |
|---|---|---|
| `glow-soft` | `0 0 20px rgba(245, 158, 11, 0.18)` | Hover glow on nodes, subtle emphasis. |
| `glow-active` | `0 0 35px rgba(245, 158, 11, 0.35)` | Running nodes, selected elements, primary actions. |
| `glow-input` | `0 0 60px rgba(245, 158, 11, 0.12)` | Focus state on input bar / command palette. |
| `shadow-panel` | `0 4px 24px rgba(0,0,0,0.5), 0 0 0 1px rgba(68,64,60,0.08)` | Floating HUD panels, inspector, dock. |
| `shadow-deep` | `0 8px 32px -8px rgba(0,0,0,0.7)` | Modals, dropdowns, elevated surfaces. |
| `inner-border` | `inset 0 0 0 1px rgba(68,64,60,0.12)` | Subtle border inside panels (glass effect). |
| `inner-border-accent` | `inset 0 0 0 1px rgba(245,158,11,0.5)` | Selected state, focused node. |

---

## Visual States

### Node States (Graph Canvas)

| State | Background | Border | Glow | Icon / Indicator |
|---|---|---|---|---|
| **Idle** | `bg-panel` | `text-muted` (`#44403C`) | None | `text-secondary` |
| **Running** | `bg-panel` | `accent-light` | `glow-active` | Pulsing amber dot |
| **Done** | `bg-panel` | `status-success` / `text-muted` | None | Green checkmark |
| **Pending** | `bg-panel` | `text-muted` dashed | None | Clock icon, 70% opacity |
| **Selected** | `bg-panel` | `accent-light` | `glow-active` + inner accent border | — |
| **Error** | `bg-panel` | `status-error` | None | Red alert icon |

### Edge States (Connections)

| State | Stroke | Animation |
|---|---|---|
| **Standard** | `rgba(120,113,108,0.2)` | None |
| **Running** | `#FBBF24` | `stroke-dasharray: 4; dash 1s linear infinite` |
| **Success** | `rgba(34,197,94,0.5)` | None |
| **Pending/Dashed** | `rgba(120,113,108,0.2)` | `stroke-dasharray: 4 4` |

---

## Typography Colors

| Element | Color | Weight | Notes |
|---|---|---|---|
| Headings (H1-H3) | `text-primary` | 600-700 | — |
| Body text | `text-primary` | 400 | — |
| Captions, metadata | `text-secondary` | 400 | Mono font preferred |
| Labels, tags | `accent` or `text-secondary` | 500 | Uppercase + tracking-wider |
| Code / Logs | `text-secondary` | 400 | `JetBrains Mono`. Accent keywords in `accent-light`. |
| Links | `accent-light` | 500 | Underline on hover |
| Buttons (Primary) | White text on `accent` bg | 500 | Hover: lighten bg |
| Buttons (Secondary) | `text-primary` on `bg-panel` | 500 | Border `text-muted` |

---

## Component Examples

### Input Bar (Command Palette)

```
Background: bg-main @ 92% opacity + blur(20px)
Border: 1px solid rgba(245, 158, 11, 0.15)
Focus Border: 1px solid rgba(245, 158, 11, 0.4)
Focus Glow: glow-input
Text: text-primary
Placeholder: text-secondary @ 50%
Send Button: accent bg, white icon
```

### Floating Inspector

```
Background: bg-panel @ 88% opacity + blur(16px)
Border: 1px solid rgba(68, 64, 60, 0.2)
Shadow: shadow-panel
Header Border Bottom: text-muted @ 20%
Live Indicator: status-success pulsing dot + "Live" label
```

### HUD Panel (Top Bar / Dock)

```
Background: bg-panel @ 88% opacity + blur(16px)
Border: 1px solid rgba(68, 64, 60, 0.2)
Shadow: shadow-panel
Inset Highlight: inset 0 1px 0 rgba(255,255,255,0.04)
```

### Node Card (Agent / Work Unit)

```
Background: #1C1C1C @ 90-95% opacity + blur
Border: varies by state (see Visual States above)
Border Radius: 12px (lg) for agents, 8px (md) for WUs
Shadow: none by default, glow-active when running/selected
Left Accent Bar: 2-4px solid color for running/done states
```

---

## Tailwind Config Reference

```js
colors: {
  bgMain: '#0A0A0A',
  bgPanel: '#141414',
  bgHover: '#1C1C1C',
  bgCanvas: '#070707',
  textPri: '#E7E5E4',
  textSec: '#78716C',
  accent: '#F59E0B',
  accentLight: '#FBBF24',
  semantic: {
    success: '#22C55E',
    error: '#EF4444',
    warning: '#F97316',
    info: '#A8A29E'
  }
},
boxShadow: {
  'glow': '0 0 20px rgba(245,158,11,0.18)',
  'glow-active': '0 0 35px rgba(245,158,11,0.35)',
  'glow-panel': '0 8px 32px -8px rgba(0,0,0,0.7)',
  'inner-border': 'inset 0 0 0 1px rgba(68,64,60,0.12)',
  'inner-border-accent': 'inset 0 0 0 1px rgba(245,158,11,0.5)',
  'float': '0 4px 24px rgba(0,0,0,0.5), 0 0 0 1px rgba(68,64,60,0.08)'
}
```

---

## Anti-Patterns (Do NOT)

1. **Never use blue/cyan as primary accent.** The amber identity is intentional. Blue is reserved for third-party integrations or external links only.
2. **Never use pure white (`#FFFFFF`) for text.** Always use `text-primary` (`#E7E5E4`) — pure white is too harsh on dark backgrounds.
3. **Never use pure black (`#000000`) for backgrounds.** Use `bg-canvas` (`#070707`) or `bg-main` (`#0A0A0A`).
4. **Avoid glowing non-active elements.** Glow is reserved for `running`, `selected`, and primary focus states. Overuse dilutes impact.
5. **Do not mix warm and cool grays.** Stick to the stone-based grays (`#44403C`, `#78716C`, `#E7E5E4`). No blue-gray or cool-gray tints.
6. **Do not use semantic colors for decorative purposes.** Green is strictly success/done. Red is strictly error. Orange is strictly warning/pending.

---

## Accessibility Notes

- The amber-on-obsidian combination provides excellent contrast for dark mode usage (WCAG AA for large text, AAA for UI components).
- Running animations (pulse, dash) should respect `prefers-reduced-motion`.
- Color alone should never indicate state — always pair with icon or text label.

---

## Changelog

| Date | Change | Author |
|---|---|---|
| 2026-05-17 | Adopted Solar Core palette (amber/obsidian). Replaced previous cyan/blue scheme. | — |
