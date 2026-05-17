# OrchestraOS — Design System

> **Status:** In Progress (Fase de Fundação)
>
> **Principle:** Technology-agnostic. These docs define *what* the UI should look and feel like, not *how* to build it. Any framework (React, Vue, Tauri, TUI) can implement from these specs.

---

## Why Document Design Before Tech?

The visual identity of OrchestraOS must remain consistent regardless of:
- Which frontend framework is chosen
- Whether the interface is web, desktop, or terminal
- Who implements the feature (human or AI agent)

These documents are the **source of truth** for all visual decisions.

---

## Document Map

### Foundation (Tokens)
Documents that define the atomic building blocks. Every other doc depends on these.

| Document | Status | Description |
|---|---|---|
| [`COLOR_SYSTEM.md`](./COLOR_SYSTEM.md) | ✅ Adopted | Paleta Solar Core (âmbar/obsidian), glows, sombras, estados semânticos |
| [`TYPOGRAPHY.md`](./TYPOGRAPHY.md) | 📝 Draft | Fontes, hierarquia, tamanhos, pesos, espaçamento entre letras |
| [`SPACING_LAYOUT.md`](./SPACING_LAYOUT.md) | 📝 Draft | Grid, espaçamentos, border-radius, breakpoints, z-index layers |
| [`MOTION.md`](./MOTION.md) | 📝 Draft | Durações, easing curves, regras de animação, reduced-motion |

### Components
Especificações visuais e comportamentais de cada componente. Sem código — apenas *o quê*, *quando* e *como* deve se comportar.

| Document | Status | Description |
|---|---|---|
| [`COMPONENTS.md`](./COMPONENTS.md) | 📝 Draft | Catálogo de componentes: botões, inputs, cards, modais, tooltips, badges |
| [`INTERACTION_STATES.md`](./INTERACTION_STATES.md) | 📝 Draft | Hover, focus, active, disabled, loading, selected, error states |
| [`ICONOGRAPHY.md`](./ICONOGRAPHY.md) | 📝 Draft | Estilo de ícones, tamanhos, stroke, quando usar outlined vs filled |

### Patterns & Guidelines
Regras de alto nível sobre como montar interfaces consistentes.

| Document | Status | Description |
|---|---|---|
| [`CONTENT_VOICE.md`](./CONTENT_VOICE.md) | 📝 Draft | Tom de voz, terminologia do domínio, labels, mensagens de erro |
| [`DATA_VIZ.md`](./DATA_VIZ.md) | ⏳ Future | Como representar métricas, gráficos, logs, timelines |
| [`RESPONSIVE.md`](./RESPONSIVE.md) | ⏳ Future | Comportamento em diferentes viewports, mobile, tablet |
| [`ACCESSIBILITY.md`](./ACCESSIBILITY.md) | ⏳ Future | WCAG targets, keyboard nav, screen readers, focus management |

### Domain-Specific
Regras visuais específicas dos conceitos do OrchestraOS.

| Document | Status | Description |
|---|---|---|
| [`CANVAS_GRAPH.md`](./CANVAS_GRAPH.md) | ⏳ Future | Regras do grafo DAG: nós, edges, zoom, pan, mini-map, layout algorithms |
| [`HUD_OVERLAY.md`](./HUD_OVERLAY.md) | ⏳ Future | Regras de elementos flutuantes: inspector, input bar, dock, notificações |
| [`AGENT_VISUALS.md`](./AGENT_VISUALS.md) | ⏳ Future | Como representar agentes visualmente: avatar, status, capabilities, runs |

---

## Design Principles

1. **The Canvas is King** — The graph canvas is the primary interface. Everything else is HUD overlay.
2. **Amber Means Active** — The amber glow is sacred. Only running/selected/focused elements may glow.
3. **Dark by Default** — OrchestraOS is a dark-mode-first system. Light mode is not a priority.
4. **Information Density** — Show what matters, hide what doesn't. Progressive disclosure via inspector.
5. **Spatial Memory** — Users should always know where they are in the graph. Mini-map + context chips.
6. **Keyboard First** — Every action must be keyboard accessible. Mouse is optional.

---

## How to Use These Docs

- **Implementing a new screen?** Start with `COLOR_SYSTEM.md` + `SPACING_LAYOUT.md` + `COMPONENTS.md`
- **Adding a new component?** Write the spec in `COMPONENTS.md` before writing code
- **Changing a color?** Update `COLOR_SYSTEM.md` first, then propagate
- **Writing error messages?** Check `CONTENT_VOICE.md`
- **Animating something?** Check `MOTION.md` for durations and easing

---

## Decision Log

| Date | Decision | Rationale |
|---|---|---|
| 2026-05-17 | Adopted Solar Core palette (amber/obsidian) | Superior emotional impact vs generic tech blue |
| 2026-05-17 | Canvas-first, HUD-overlay paradigm | Graph is the primary mental model for orchestration |
