# OrchestraOS — Typography System

> **Status:** Draft
>
> **Principle:** Readable, technical, dense. The UI feels like a professional command station — not a marketing page.

---

## Typefaces

### Primary: Inter
- **Usage:** All UI text — headings, body, labels, buttons
- **Weights used:** 400 (Regular), 500 (Medium), 600 (Semibold), 700 (Bold)
- **Why:** Excellent legibility at small sizes, neutral personality, great for dense UIs

### Monospace: JetBrains Mono
- **Usage:** Code, logs, IDs, timestamps, metrics, technical labels
- **Weights used:** 400, 500, 700
- **Why:** Distinct zeros/Os and ones/ls, ligatures optional, professional dev tool aesthetic

---

## Type Scale

All sizes are in `rem` (root = 16px) for accessibility.

| Token | Size | Line Height | Letter Spacing | Weight | Usage |
|---|---|---|---|---|---|
| `text-2xl` | 1.5rem (24px) | 1.2 | -0.02em | 700 | Page titles, agent names in inspector |
| `text-xl` | 1.25rem (20px) | 1.3 | -0.01em | 700 | Modal headings, section titles |
| `text-lg` | 1.125rem (18px) | 1.4 | 0 | 600 | Card titles, node headers |
| `text-base` | 1rem (16px) | 1.5 | 0 | 400-500 | Body text, descriptions |
| `text-sm` | 0.875rem (14px) | 1.5 | 0 | 400-500 | Labels, menu items, button text |
| `text-xs` | 0.75rem (12px) | 1.4 | 0.01em | 400-500 | Captions, metadata, secondary info |
| `text-2xs` | 0.625rem (10px) | 1.3 | 0.02em | 500 | Tags, badges, timestamps, IDs |
| `text-3xs` | 0.5625rem (9px) | 1.2 | 0.03em | 500 | Mini labels, status indicators |

### Monospace Scale

| Token | Size | Line Height | Weight | Usage |
|---|---|---|---|---|
| `mono-lg` | 1.25rem (20px) | 1.2 | 500 | Large metric values |
| `mono-base` | 1rem (16px) | 1.4 | 400 | Code blocks, logs |
| `mono-sm` | 0.875rem (14px) | 1.4 | 400 | Inline code, IDs |
| `mono-xs` | 0.75rem (12px) | 1.3 | 400 | Timestamps, small IDs |
| `mono-2xs` | 0.625rem (10px) | 1.2 | 500 | Compact labels, status tags |

---

## Hierarchy Patterns

### Page Header (HUD Top Bar)
```
Logo:        text-sm, weight 700, text-primary
Context:     text-2xs, weight 500, text-secondary, uppercase, tracking-widest
```

### Inspector Header
```
Title:       text-xs, weight 600, text-primary
Subtitle:    text-2xs, weight 400, text-secondary
Status:      text-3xs, weight 700, uppercase, tracking-wider
```

### Node Card (Agent)
```
Name:        text-sm, weight 600, text-primary
Status:      text-2xs, weight 500, mono, accent-light (if running)
Meta:        text-2xs, weight 400, text-secondary
```

### Node Card (Work Unit)
```
ID:          text-2xs, weight 500, mono, text-secondary, in badge
Title:       text-xs, weight 500, text-primary
```

### Input Bar
```
Placeholder: text-sm, weight 400, text-secondary @ 50% opacity
Input text:  text-sm, weight 400, text-primary
Hints:       text-2xs, weight 400, mono, text-secondary @ 70%
```

---

## Letter Spacing Rules

| Context | Letter Spacing |
|---|---|
| Uppercase labels / tags | `0.05em` to `0.1em` (tracking-wider / tracking-widest) |
| Headings | `-0.02em` to `-0.01em` (tighter for impact) |
| Body text | `0` (normal) |
| Monospace IDs / codes | `0` to `0.02em` |

---

## Text Colors (by context)

| Context | Color Token |
|---|---|
| Primary content | `text-primary` (#E7E5E4) |
| Secondary / descriptions | `text-secondary` (#78716C) |
| Disabled / inactive | `text-secondary` @ 50% opacity |
| Active / highlighted | `accent-light` (#FBBF24) |
| Links | `accent-light` with underline on hover |
| Error text | `status-error` (#EF4444) |
| Success text | `status-success` (#22C55E) |

---

## Monospace Usage Rules

**Always use monospace for:**
- IDs (task-099, wu-002, run-101)
- Timestamps (10:30:14 UTC)
- Metric values (2 Agents, 7 Work Units)
- Code snippets, logs, stack traces
- Status durations (02h 14m)

**Never use monospace for:**
- Descriptions or prose
- Button labels (except technical actions)
- Navigation items

---

## Anti-Patterns

1. **Never use pure white (#FFFFFF) for text.** Always use `text-primary` (#E7E5E4).
2. **Never use font-size smaller than 9px (text-3xs).** If it doesn't fit, redesign the layout.
3. **Never mix more than 2 font families on the same screen.** Inter + JetBrains Mono only.
4. **Avoid italic text.** The UI is technical, not editorial. Use weight or color for emphasis.
5. **Don't underline text except links.** Underline means "clickable".
