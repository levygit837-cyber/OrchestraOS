# OrchestraOS — Iconography System

> **Status:** Draft
>
> **Principle:** Icons are functional, not decorative. Every icon must communicate meaning instantly.

---

## Icon Style

### Stroke-based (Outlined)
- **Source:** Heroicons-style (or any consistent outlined set)
- **Stroke width:** 2px default, 1.5px for small (12px), 3px for emphasis (checkmarks)
- **Stroke linecap:** round
- **Stroke linejoin:** round
- **Fill:** none (outlined style)
- **Why:** Outlined icons are lighter and work better in dense UIs. They don't compete with text for attention.

### Size Scale

| Size | Dimension | Stroke | Usage |
|---|---|---|---|
| `icon-xs` | 12px | 1.5px | Inline in text, compact tags |
| `icon-sm` | 16px | 2px | Buttons, list items, badges |
| `icon-md` | 20px-24px | 2px | Primary icons in cards, nav items |
| `icon-lg` | 32px | 2px | Feature icons, empty states |
| `icon-xl` | 40px-48px | 1.5px | Orchestrator node, hero icons |

---

## Color Rules

| Context | Color |
|---|---|
| Default icon | `text-secondary` (#78716C) |
| Hover icon | `text-primary` (#E7E5E4) |
| Active/Selected icon | `accent-light` (#FBBF24) |
| Success icon | `status-success` (#22C55E) |
| Warning icon | `status-warning` (#F97316) |
| Error icon | `status-error` (#EF4444) |
| Disabled icon | `text-secondary` @ 50% |
| Button icon | Matches button text color |

---

## Icon Catalog

### Navigation

| Icon | Name | Usage |
|---|---|---|
| ◫ | `layout-grid` | Canvas view (active) |
| ☑ | `clipboard-check` | Tasks |
| ▶ | `terminal` | Runs & Logs |
| 🧪 | `beaker` | Agents |
| ✚ | `plus` | New Task |

### Actions

| Icon | Name | Usage |
|---|---|---|
| 🔍 | `search` | Search, find |
| ⚙ | `cog` | Settings |
| 🔔 | `bell` | Notifications |
| ✕ | `x` | Close, dismiss, delete |
| ✓ | `check` | Done, success, confirm |
| ⎘ | `external-link` | Open external |
| 📎 | `paper-clip` | Attach file |
| 🎤 | `microphone` | Voice input |
| 📤 | `paper-airplane` | Send, submit |
| ⎘ | `arrows-expand` | Fullscreen |
| ⤢ | `zoom-in` | Zoom in |
| ⤡ | `zoom-out` | Zoom out |
| ⟲ | `refresh` | Reset, reload |

### Status

| Icon | Name | Usage |
|---|---|---|
| ● | `dot` | Live indicator, status dot |
| ◐ | `clock` | Pending, waiting |
| ⚡ | `bolt` | Active, running, energy |
| ✕ | `x-circle` | Error, failed |
| ! | `exclamation` | Warning, alert |
| ℹ | `information-circle` | Info, help |

### Agent / Work Unit

| Icon | Name | Usage |
|---|---|---|
| 🧑‍💻 | `code` | Codex-Builder (dev agent) |
| 📄 | `document-text` | Review-Bot (doc agent) |
| ⚙ | `cog` | Orchestrator |
| 📦 | `cube` | Work Unit |
| 🏷 | `tag` | Task label |
| 🔗 | `link` | Dependency |

### Canvas Controls

| Icon | Name | Usage |
|---|---|---|
| 🔍➕ | `zoom-in` | Zoom in |
| 🔍➖ | `zoom-out` | Zoom out |
| ⌂ | `home` | Fit view, reset zoom |
| 👁 | `eye` | Show/hide layer |
| 📌 | `pin` | Pin inspector |

---

## Icon + Text Pairing

When icon appears with text:
```
Gap: 8px (space-2)
Alignment: Center vertically
Icon position: Left of text (default), or right for external links
Icon color: Matches text color or one step lighter
```

### Examples
```
[✓] Done           icon=success, text=success
[●] Running        icon=accent, text=accent (with pulse)
[◐] Pending        icon=text-secondary, text=text-secondary
[!] Error          icon=error, text=error
```

---

## Avatar / Agent Icons

Agents have visual identifiers beyond just names:

```
Container: 36px-40px square, radius-lg (12px)
Background: tinted by agent type
  - Dev agent: accent @ 15%
  - Review agent: text-muted @ 15%
  - System: text-secondary @ 15%
Icon: 20px-24px, centered
Border: 1px solid matching color @ 20%
```

---

## Empty States

When no icon is sufficient, use a **dashed circle** as placeholder:
```
Size: 48px-64px
Border: 1px dashed text-secondary
Icon inside: 24px, text-secondary
```

---

## Anti-Patterns

1. **Never use emojis in production.** Use proper SVG icons. Emojis are acceptable only in mockups.
2. **Never use different icon styles.** Don't mix filled and outlined icons.
3. **Never use icons without labels for unfamiliar concepts.** "Settings" is fine without label; "WU Graph Decomposition" is not.
4. **Don't animate icons unnecessarily.** Only spin loading icons and pulse status dots.
5. **Avoid decorative icons.** Every icon must serve a functional purpose.
