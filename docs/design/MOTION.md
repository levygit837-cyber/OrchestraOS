# OrchestraOS — Motion System

> **Status:** Draft
>
> **Principle:** Motion conveys state, not decoration. Animations are fast, purposeful, and never block interaction.

---

## Philosophy

In a system that orchestrates agents in real-time, motion communicates **life**:
- A pulsing node means "this is running right now"
- A dashing edge means "data is flowing"
- A sliding inspector means "context has changed"

Motion is not for delight — it's for **status awareness**.

---

## Duration Scale

| Token | Duration | Usage |
|---|---|---|
| `duration-instant` | 0ms | No animation. State changes immediately. Used for: error states, critical alerts. |
| `duration-fast` | 100ms | Micro-interactions. Hover color changes, opacity shifts. |
| `duration-normal` | 150ms | Standard transitions. Button presses, panel fades, selection changes. |
| `duration-smooth` | 200ms | UI element movements. Inspector open/close, tooltip appear, dropdown. |
| `duration-slow` | 300ms | Larger movements. Modal open, canvas zoom, sidebar collapse. |
| `duration-ambient` | 2s - 3s | Continuous ambient loops. Pulse rings, breathing glows. |

---

## Easing Curves

| Token | Value | Usage |
|---|---|---|
| `ease-default` | `cubic-bezier(0.4, 0, 0.2, 1)` | Default for most transitions. Smooth deceleration. |
| `ease-bounce` | `cubic-bezier(0.175, 0.885, 0.32, 1.275)` | Node hover scale. Playful but quick. |
| `ease-in-out` | `cubic-bezier(0.4, 0, 0.6, 1)` | Symmetric transitions. Inspector slide, panel collapse. |
| `ease-out` | `cubic-bezier(0, 0, 0.2, 1)` | Elements appearing. Modal, toast, tooltip. |
| `ease-linear` | `linear` | Continuous loops. Dashing edges, spinner rotation. |

---

## Animation Patterns

### Node Hover
```
Transform: translateY(-4px) scale(1.02)
Duration: 300ms
Easing: ease-bounce
Z-index: elevate to 20
```

### Node Selection
```
Border: inset accent border appears
Glow: glow-active activates
Duration: 200ms
Easing: ease-default
```

### Inspector Open
```
Initial: opacity 0, scale 0.96, translateY(8px)
Final: opacity 1, scale 1, translateY(0)
Duration: 250ms
Easing: ease-bounce
```

### Inspector Close
```
Final: opacity 0, scale 0.96
Duration: 200ms
Easing: ease-default
```

### Input Bar Focus
```
Border: accent opacity 0.15 → 0.4
Glow: expands from 40px to 60px radius
Duration: 300ms
Easing: ease-default
```

### Canvas Zoom
```
Transform: scale()
Duration: 150ms
Easing: ease-out
```

### Running Edge (Data Flow)
```
Stroke-dashoffset: animate continuously
Duration: 1s per cycle
Easing: linear
```

### Pulse Ring (Running Node)
```
Scale: 1 → 1.02 → 1
Box-shadow: accent glow expands and fades
Duration: 2s
Easing: ease-default
Iteration: infinite
```

### Ping Indicator (Live Status)
```
Outer ring: scale 1 → 1.5, opacity 0.75 → 0
Duration: 1s
Easing: ease-out
Iteration: infinite
Inner dot: static
```

---

## Reduced Motion

All animations MUST respect `prefers-reduced-motion: reduce`:

```css
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

When reduced motion is active:
- Pulse rings become static glows
- Dashing edges become solid lines
- Hover transforms become color changes only
- Inspector appears instantly (no slide/scale)

---

## Stagger Patterns

When multiple elements animate together:

| Scenario | Delay between items | Example |
|---|---|---|
| Node group appear | 50ms | WUs appearing under an agent |
| List items | 30ms | Inspector content loading |
| Toast queue | 150ms | Multiple notifications |
| Graph initial load | 100ms | Nodes appearing layer by layer |

---

## What NOT to Animate

1. **Background colors** — Never animate the canvas background or grid.
2. **Text color** — Avoid color transitions on text. It feels laggy.
3. **Width/height** — Use transform: scale instead. Performance is better.
4. **Blur filters** — Never animate backdrop-filter blur. It causes repaint storms.
5. **Scroll position** — Let the browser handle scrolling natively.

---

## Performance Rules

1. **Only animate `transform` and `opacity`.** These are GPU-composited.
2. **Use `will-change` sparingly.** Only on elements that animate frequently (running nodes, edges).
3. **Limit concurrent animations.** Max ~10 simultaneous animations on screen.
4. **Pause off-screen animations.** Nodes outside viewport should not pulse.
