# Plano de 5 Variações de Design — OrchestraOS

> **Objetivo:** Criar 5 protótipos visuais distintos, focados em **Chat de Agentes + Humano** dentro de um **Sistema de Orquestração de Múltiplos Agentes**.
>
> **Restrição:** Nenhuma metáfora abstrata (mar, folhas, órbitas, cosmos). Toda direção visual deve ser funcional, pragmática e semanticamente alinhada ao domínio de agentes de IA.
>
> **Créditos disponíveis:** 5

---

## Inventário de Estilos Já Existentes (NÃO repetir)

| Local | Nome | Fundo | Accent | Fonte | Layout |
|---|---|---|---|---|---|
| prototypes/v1 | Glassmorphism Spatial | #070707 | âmbar | Space Grotesk | Canvas + HUD |
| prototypes/v2 | Cyber-Ops Tactical | #050505 | âmbar | Space Grotesk | Canvas denso |
| prototypes/v3 | Swiss Precision | #0A0A0A | âmbar | JetBrains Mono | Canvas + Sidebar |
| prototypes/v4 | Editorial Light | #FFFFFF | navy/coral | Playfair Display | Sidebar + Cards |
| canvas/orbital | Orbital Chat | #050505 | sage/coral/lavender | Outfit | Radial + HUD |
| canvas/garden | Orchestrator Garden | #F5F0E8 | forest/laranja | Satoshi | Orgânico |
| canvas/multi-agent | Multi-Agent Workspace | #0A0A0A | âmbar | Geist | Chat + Sidebar |
| canvas/hub | AI Orchestrator Hub | #0F0F23 | purple/cyan/pink | Syne | Radial neon |
| canvas/dashboard | AI Orchestration Dashboard | #FFFFFF | sky blue | Geist | Grid clínico |
| archive/v4 | Canvas v4 Terminal | #080808 | âmbar | JetBrains Mono | Terminal/logs |

**Regra:** Cada nova variação deve diferir em pelo menos 4 critérios: paleta, tipografia, layout, densidade e metáfora.

---

## As 5 Novas Variações

---

### Variação 1: "Executive Threaded Hub"
**Metáfora:** Sala de comando executiva. Um app de mensagens corporativo de alto nível onde cada agente é um contato em uma thread.

**Foco Funcional:** Múltiplas conversas thread-like. O humano vê todos os agentes como contatos em uma lista, pode entrar em qualquer thread, aprovar ferramentas inline, e ver o status de cada agente no avatar.

| Critério | Especificação |
|---|---|
| **Fundo** | `#141416` (charcoal quente) |
| **Primária** | `#1E1E22` (panels) |
| **Accent** | `#C9A227` (executive gold) — nunca âmbar |
| **Texto** | `#E8E6E1` (warm white) |
| **Sucesso** | `#4CAF7A` (sage money green) |
| **Erro** | `#CF4B4B` (muted red) |
| **Fonte UI** | Plus Jakarta Sans (geometric, corporate, clean) |
| **Fonte Mono** | IBM Plex Mono |
| **Layout** | 3 colunas: Threads 25% + Chat 50% + Contexto 25% |
| **Densidade** | Balanced — legível, espaçamento confortável |
| **Border Radius** | 12px (macio, premium) |
| **Efeitos** | Sem glow. Sombras planas sutis apenas. |
| **Elementos Distintivos** | ① Avatares circulares com anéis de status pulsantes ② Thread badges com contador de aprovações pendentes ③ Botões de aprovação inline no chat (não modal) ④ Typing indicator elegante com nome do agente |

**Cenário no Design:**
- Lista de 4 agentes à esquerda: Codex-Builder (running, gold ring), Review-Bot (idle), Docs-Writer (waiting approval, red badge "2"), Test-Agent (done, green check).
- Chat central mostrando conversa com Codex-Builder. Mensagens do humano à direita, do agente à esquerda com avatar.
- Uma mensagem do agente com um Tool Request inline: "Requesting file_write on events.go" com botões [Approve] [Reject] [View] abaixo da mensagem.
- Painel direito mostra contexto do agente selecionado: Session ID, Sandbox status, último checkpoint, heartbeat.

---

### Variação 2: "Industrial Kanban Command"
**Metáfora:** Painel de controle industrial / SCADA. Cada Work Unit é uma tarefa em um quadro Kanban, e o chat do agente está integrado no card.

**Foco Funcional:** Work Units como cards em colunas (Pending → Running → Review → Done). O operador clica num card e vê o chat com o agente, logs, e pode intervir. Aprovações aparecem como alertas no topo do card.

| Critério | Especificação |
|---|---|
| **Fundo** | `#2D2D30` (steel gray industrial) |
| **Primária** | `#3E3E42` (panel industrial) |
| **Accent** | `#FF6B35` (safety orange) — usado para RUNNING e ATTENTION |
| **Texto** | `#ECECEC` (off-white industrial) |
| **Sucesso** | `#7CB342` (industrial green) |
| **Erro** | `#E53935` (alarm red) |
| **Fonte UI** | Rajdhani (condensed, industrial, técnica) |
| **Fonte Mono** | Roboto Mono |
| **Layout** | Kanban board full-width. Colunas: PENDING / RUNNING / REVIEW / DONE. Bottom drawer para chat. |
| **Densidade** | Dense — muita informação por pixel, como um painel real |
| **Border Radius** | 4px (quase sharp, funcional) |
| **Efeitos** | Flat, zero blur, zero glow. Bordas grossas de 1-2px. |
| **Elementos Distintivos** | ① LED status indicators (ponto colorido pulsante) em cada card ② Progress bars grossos tipo barra de energia ③ Numeração industrial (WU-001, WU-002) em fonte mono destacada ④ Alert banner laranja no card quando há tool request pendente ⑤ Bottom drawer que sobe com o chat do agente selecionado |

**Cenário no Design:**
- 4 colunas Kanban com ~3 cards cada.
- Card WU-002 "Implementar event envelope" na coluna RUNNING com LED laranja pulsando, barra de progresso em 60%, e alert banner laranja "Tool Request: bash go test".
- Card WU-003 na coluna PENDING com LED amarelo estático.
- Bottom drawer parcialmente aberto mostrando chat com Codex-Builder. Mensagens em bubbles compactos.
- Topo mostra métricas do sistema: 2 Running, 1 Pending Approval, 3 Done, CPU 45%.

---

### Variação 3: "Mission Control Timeline"
**Metáfora:** Console de controle de missão NASA/SpaceX. A execução é uma linha do tempo vertical onde cada evento é um marco, e o chat é contextual ao evento.

**Foco Funcional:** Timeline vertical mostrando todo o fluxo de orquestração: UserMessage → Intake → Task Creation → WU Decomposition → Runs → Sessions → Checkpoints. O operador clica num evento e vê o chat/exchange naquele momento.

| Critério | Especificação |
|---|---|
| **Fundo** | `#0A0E1A` (deep space black-blue) |
| **Primária** | `#111827` (panel escuro) |
| **Accent** | `#00E5CC` (mission cyan) — usado para RUNNING e LIVE |
| **Texto** | `#F0F4F8` (cool white) |
| **Sucesso** | `#00C853` (neon green) |
| **Erro** | `#FF1744` (bright alert red) |
| **Fonte UI** | Exo 2 (futurista técnica, usada em interfaces aeroespaciais) |
| **Fonte Mono** | Source Code Pro |
| **Layout** | Timeline vertical central (60% width). Eventos alternam esquerda/direita. Painel de chat flutua à direita quando evento selecionado. |
| **Densidade** | Dense — dados técnicos visíveis, como um console real |
| **Border Radius** | 2px (quase sharp, técnico) |
| **Efeitos** | Grid de fundo sutil (como console de missão). Cyan glow MÍNIMO apenas em elementos LIVE. |
| **Elementos Distintivos** | ① Timeline central com marcadores circulares coloridos por estado ② T-minus / elapsed time counters (ex: "T+ 02m 14s") ③ Heartbeat line (linha horizontal pulsando quando agente ativo) ④ Evento badges com ícones: [INTAKE], [TASK], [WU], [RUN], [CHK], [TOOL] ⑤ Chat contextual: ao clicar num evento RUN, aparece o exchange humano-agente daquela sessão |

**Cenário no Design:**
- Topo: "MISSION: task-099 — Event Envelope System" com status "IN PROGRESS" em cyan.
- Timeline vertical com marcadores: 10:23:45 [INTAKE] User message received → 10:23:46 [TASK] Created task-099 → 10:23:50 [WU] Decomposed into 6 work units → 10:24:00 [RUN] run-101 starting wu-002 → 10:24:15 [TOOL] Request approval: file_write → 10:24:20 [CHK] Checkpoint-003 reached.
- Evento [TOOL] destacado com borda ciano e botões [Approve] [Reject].
- Painel direito mostrando chat da sessão run-101: mensagens do Codex-Builder e respostas do humano.

---

### Variação 4: "DevStudio Multi-Agent IDE"
**Metáfora:** IDE profissional (VS Code / JetBrains). Cada agente é tratado como um "arquivo" ou "terminal" em seu próprio pane. O desenvolvedor já conhece esse paradigma.

**Foco Funcional:** Interface de desenvolvedor onde cada agente tem sua própria aba (tab) e painel. O chat é o "console/terminal" do agente. O operador alterna entre agentes como alterna entre arquivos num IDE.

| Critério | Especificação |
|---|---|
| **Fundo** | `#1E1E1E` (VS Code dark) |
| **Primária** | `#252526` (sidebar / panel) |
| **Accent** | `#4EC9B0` (teal — como classes/types no VS Code) |
| **Texto** | `#D4D4D4` (light gray) |
| **Sucesso** | `#B5CEA8` (light green) |
| **Erro** | `#F48771` (salmon red) |
| **String/Mensagem** | `#CE9178` (orange string) |
| **Fonte UI** | Segoe UI (clean, Microsoft-style) |
| **Fonte Mono** | Cascadia Code (fonte de IDE moderna) |
| **Layout** | Activity Bar esquerda (ícones de agentes) + Sidebar (file-tree de tasks/WUs) + Editor area (split panes, cada pane = chat de um agente) + Bottom panel (logs/event stream) |
| **Densidade** | Dense — como um IDE real |
| **Border Radius** | 0px (sharp, funcional) |
| **Efeitos** | Zero glow, zero blur. Bordas de 1px sólidas entre painéis. |
| **Elementos Distintivos** | ① Activity Bar vertical à esquerda com ícones de cada agente + badge de notificação ② Tabs por agente na área de editor, com indicador de status (ponto colorido) ③ Chat estilo terminal: mensagens com syntax highlighting (human = comentário cinza, agente = código branco, tool request = string laranja) ④ Bottom panel com event stream estilo console: timestamps em cinza, tipos coloridos ⑤ Integrated "Problems" panel mostrando tool requests pendentes como warnings |

**Cenário no Design:**
- Activity Bar esquerda: 4 ícones de agentes. Codex-Builder com ponto verde (ativo), Review-Bot com ponto cinza, Docs-Writer com badge "1" (tool pending).
- Editor area com 2 tabs abertas: "Codex-Builder" (ativo) e "Review-Bot". Tab ativo com borda teal embaixo.
- Chat no estilo terminal: linhas de texto com mono font. Human message prefix: "// user:". Agent message prefix: "// codex-builder:". Tool request destacado em laranja: `[TOOL] file_write("events.go")`.
- Sidebar mostra árvore de Tasks > Work Units > Runs.
- Bottom panel: console output com linhas de evento coloridas.

---

### Variação 5: "Zen Focus Mode"
**Metáfora:** Escritório minimalista japonês ( Japandi ). Foco absoluto em um agente por vez. Tudo o mais some da vista. A interação é contemplativa e intencional.

**Foco Funcional:** Um agente por vez em tela cheia. O humano navega entre agentes via gestos ou atalhos. Cada agente ocupa o espaço completo, com mensagens grandes, legíveis, e sem distrações. Aprovações aparecem como interrupções elegantes no centro.

| Critério | Especificação |
|---|---|
| **Fundo** | `#F7F5F0` (washi paper white) |
| **Primária** | `#FFFFFF` (cards flutuantes) |
| **Accent** | `#D4734F` (terracotta) — quente, terroso |
| **Texto** | `#2C2C2C` (sumi ink) |
| **Sucesso** | `#81B29A` (sage green) |
| **Erro** | `#C75146` (muted terracotta red) |
| **Fonte UI** | Noto Sans JP (japonês clean, geometric, legível) |
| **Fonte Mono** | JetBrains Mono |
| **Layout** | Single pane 100% width. Mensagens centralizadas com largura máxima de 720px. Navegação por dots no topo ou swipe. |
| **Densidade** | Airy — muito espaço em branco, respiração generosa |
| **Border Radius** | 24px (orgânico, suave) |
| **Efeitos** | Sombras suaves e difusas (nunca duras). Transições lentas (400ms). |
| **Elementos Distintivos** | ① Um agente por vez — avatar grande no topo com nome e status ② Mensagens enormes e legíveis (18px body) ③ Aprovações como um "card flutuante" elegante no centro da tela, desfocando o fundo ④ Breathing animation sutil no avatar quando agente está pensando ⑤ Navegação por 4 dots no topo representando os 4 agentes, com cor indicando status |

**Cenário no Design:**
- Tela cheia. Topo: 4 dots. Dot 1 (Codex-Builder) terracotta pulsante = running. Dots 2-4 cinza = idle.
- Centro: Avatar grande circular do Codex-Builder (64px) com nome e status "Running · 2m 14s".
- Conversa em coluna central estreita. Mensagens do humano alinhadas à direita com fundo terracotta claro. Mensagens do agente à esquerda com fundo branco e borda sutil.
- No meio da conversa, um card flutuante de Tool Request: "Codex-Builder requests: bash 'go test ./...'" com botões grandes [Approve] [Reject] e fundo desfocado.
- Rodapé minimalista: input bar larga e limpa com placeholder "Intervene or message agent..."

---

## Checklist de Divergência

| Variação | Fundo | Primária | Accent | Fonte UI | Fonte Mono | Layout | Densidade |
|---|---|---|---|---|---|---|---|
| Executive Threaded | #141416 | #1E1E22 | Gold #C9A227 | Plus Jakarta Sans | IBM Plex Mono | 3-col chat | Balanced |
| Industrial Kanban | #2D2D30 | #3E3E42 | Safety Orange #FF6B35 | Rajdhani | Roboto Mono | Kanban board | Dense |
| Mission Timeline | #0A0E1A | #111827 | Cyan #00E5CC | Exo 2 | Source Code Pro | Timeline vertical | Dense |
| DevStudio IDE | #1E1E1E | #252526 | Teal #4EC9B0 | Segoe UI | Cascadia Code | IDE split-pane | Dense |
| Zen Focus | #F7F5F0 | #FFFFFF | Terracotta #D4734F | Noto Sans JP | JetBrains Mono | Single pane | Airy |

---

## Mapeamento do Fluxo Real para Elementos Visuais

Baseado em `docs/design/ORCHESTRATION_FLOW_VISUAL.md` e ADRs 0002, 0006, 0007, 0023. **Todo protótipo deve conseguir representar (mesmo que de formas diferentes) todas estas entidades e estados.**

### Entidades Obrigatórias em Cada Design

| Entidade | O que Representa | Como Deve Aparecer Visualmente |
|---|---|---|
| **UserMessage** | Mensagem/input do humano que inicia tudo | Bubble de chat do humano, ou nó de origem, ou entrada na timeline. Deve ser o **ponto zero** visível. |
| **Intake** | Normaliza e rota a mensagem | Badge [INTAKE] ou linha de evento mostrando intent extraído e rota escolhida. |
| **OrchestratorService** | Control plane central (Go) | Elemento visual central e proeminente — maior que os outros. Mostra status (SCHEDULING, ANALYZING), contagem de runs ativos, aprovações pendentes. **É o protagonista visual.** |
| **Intelligent Orchestrator Agent** | Agente estratégico (LLM) | Distinto dos agentes executores. Pode aparecer como um "advisor" ou "co-pilot" ao lado do Orchestrator Hub. Não executa código — apenas sugere. |
| **Task** | Objetivo de alto nível | Card ou nó com: título, priority badge (P0/P1/P2/P3), risk level, progress ring (X/6 WUs completos). |
| **TaskGraph (DAG)** | Grafo acíclico de decomposição | Nós (WorkUnits) conectados por edges. Deve mostrar paralelismo (nós no mesmo nível horizontal). Edge types visuais: `blocks` (sólida), `requires_artifact` (tracejada + ícone), `requires_review` (pontilhada), `conflicts_with` (vermelha zigue-zague). |
| **WorkUnit (WU)** | Unidade executável atômica | Card menor que Task. Mostra: ID badge (wu-002), título, status border, left accent bar por estado. |
| **Run** | Tentativa de execução de um WU | Strip ou linha abaixo do WU. Mostra: run ID, attempt number (Attempt 2/3), duration, status. Runs falhadas empilham acima da atual. |
| **Agent** | Worker registrado com perfil | Avatar + nome + profile (code_worker, docs_writer, reviewer). Mostra capabilities como tags. |
| **AgentSession** | Sessão de trabalho do agente no sandbox | Strip abaixo do Run. Mostra: session ID, heartbeat ("2s ago"), último checkpoint, botões [Chat] [Pause] [Stop]. |
| **Sandbox** | Container isolado + worktree | Strip abaixo da Session. Mostra: container ID, branch name (feat/t-099-event-envelope), CPU/Memory metrics. Botões [View filesystem] [View terminal]. |
| **Checkpoint** | Marca de progresso persistente | Marcador na timeline ou session strip. Mostra: cp-003, current_goal, completed_goals. |
| **ToolRequest** | Solicitação de ferramenta pelo agente | **Elemento crítico de interação humano.** NUNCA em modal. Deve aparecer: inline no chat, como banner no card, ou como evento destacado na timeline. Sempre com: nome da tool, risco (safe/medium/high), preview do input, botões [Approve] [Reject] [View Context]. |
| **Review / ValidationGate** | Quality gates após completion | Badges no WU: 🛡️ Hard gate passed, ○ Soft gate pending, △ Policy gate under review. |
| **Event** | Tudo que acontece vira evento | Stream de eventos com timestamp, tipo colorido, descrição. Tipos: [RUN], [TOOL], [CHK], [EVT], [ERR]. |
| **Human Intervention** | Mensagem do humano para agente específico | Marcador distinto no chat/timeline: ícone de usuário, mensagem, indicação de qual agente recebeu. Deve parecer **first-class**, não afterthought. |

### Estados de Ciclo de Vida — Cores e Animações

```
Task:     CREATED(gray) → DECOMPOSING(amber pulse) → SCHEDULING(amber) → RUNNING(mixed) → VALIDATING → COMPLETED(green)
WorkUnit: PENDING(dashed border, 70% opacity) → READY(solid border) → RUNNING(amber border + pulse) → COMPLETED(green + checkmark)
Run:      CREATED → STARTING → RUNNING(amber) → COMPLETED(green)  ou  FAILED(red) → RETRYING(new Run abaixo)
```

### Regras Visuais de Ouro

1. **O Orchestrator é o protagonista.** Não pode ser invisível. É o centro de gravidade visual.
2. **WU → Run → Session → Sandbox NUNCA colapsados em um só nó.** Isso esconde a arquitetura real.
3. **Tool Requests são inline, nunca modais.** São parte do fluxo de execução, não interrupções externas.
4. **Retries devem ser visíveis.** Runs falhadas aparecem acima da atual, com failure reason.
5. **Eventos são contextuais à Session**, não um log global genérico.
6. **A mensagem do usuário é a origem de tudo.** Sem ela, Tasks parecem surgir do nada.

### Como Cada Variação Adapta o Fluxo

| Variação | UserMessage | Orchestrator | TaskGraph | WU/Run/Session | ToolRequest | Chat |
|---|---|---|---|---|---|---|
| **Executive Threaded** | Bubble no chat | Status no topo do thread | Resumo no painel direito | Cards expansíveis no contexto | Inline abaixo da mensagem do agente | Thread por agente |
| **Industrial Kanban** | Card na coluna "Origin" | Dashboard metric no topo | Representado implicitamente pelas colunas | Dentro de cada card Kanban | Alert banner no card + drawer chat | Bottom drawer |
| **Mission Timeline** | Primeiro evento na timeline | Evento grande central | Evento de decomposição com sub-nós | Eventos empilhados verticalmente | Evento [TOOL] destacado com botões | Painel direito contextual |
| **DevStudio IDE** | Comentário no terminal | Explorer tree root node | Tree view expandable | Sub-nodes na tree (WU > Run > Session) | Inline no terminal como `[TOOL]` | Terminal por agente (tab) |
| **Zen Focus** | Bubble grande no chat | Nome do agente + status no topo | Navegação por dots indica task progress | Expande ao tocar no status dot | Card flutuante central elegante | Chat full-screen de um agente |

---

## Prompts Prontos para Geração

### Prompt 1 — Executive Threaded Hub
```
A premium executive messaging interface for an AI agent orchestration system called OrchestraOS.

Palette: Warm charcoal #141416 background with panel color #1E1E22, executive gold #C9A227 accent (never amber), warm white #E8E6E1 text, sage green #4CAF7A for success, muted red #CF4B4B for errors.

Typography: Plus Jakarta Sans for all UI text (headings, labels, buttons), IBM Plex Mono for IDs, timestamps, and technical data. Clean geometric corporate feel.

Layout: Three-column messaging layout. Left column 25% shows a thread list of 4 AI agents (Codex-Builder, Review-Bot, Docs-Writer, Test-Agent) with circular avatars, status rings, and pending approval badges. Center column 50% shows an active chat with Codex-Builder — human messages right-aligned, agent messages left-aligned with avatar. Right column 25% shows agent context panel: session ID, sandbox status, last checkpoint, heartbeat timer.

Density: Balanced. Comfortable spacing, 14-16px body text, clear visual hierarchy.

Shapes: 12px border radius on all cards and buttons. Soft, premium feel.

Effects: NO glow effects. Only subtle flat shadows (0 2px 8px rgba(0,0,0,0.15)). No glassmorphism.

Distinctive elements: ① Circular avatars with pulsing status rings (gold for running, green for done, gray for idle). ② Inline tool approval buttons below agent messages — [Approve] [Reject] [View] — not in modals. ③ Thread list item shows "2 pending" badge in red when agent needs approval. ④ Elegant typing indicator with agent name.

Product context: OrchestraOS orchestrates multiple AI agents. Each agent executes work units in sandboxes. Humans chat with agents, approve tools, and intervene. Show a realistic scenario: Codex-Builder is running, sent a message "Implementing event envelope system...", and there's a pending tool request for file_write on events.go with inline approval buttons.
```

### Prompt 2 — Industrial Kanban Command
```
An industrial SCADA-style kanban command interface for an AI agent orchestration system called OrchestraOS.

Palette: Steel gray #2D2D30 background, panel color #3E3E42, safety orange #FF6B35 accent for RUNNING and ATTENTION states, off-white #ECECEC text, industrial green #7CB342 for done, alarm red #E53935 for errors.

Typography: Rajdhani (condensed, technical, industrial) for all UI text. Roboto Mono for IDs, counters, metrics, and timestamps. Feels like a factory control panel.

Layout: Full-width kanban board with 4 columns: PENDING / RUNNING / REVIEW / DONE. Each column contains 2-3 work unit cards. A bottom drawer (30% height) shows the chat with the selected agent. Top bar shows system metrics: 2 Running, 1 Pending Approval, 3 Done, CPU 45%, Memory 128MB.

Density: Dense. Industrial panels show maximum information. Small text (12-13px), tight spacing, thick borders.

Shapes: 4px border radius. Almost sharp corners. Functional, not decorative.

Effects: ZERO blur, ZERO glow, ZERO glassmorphism. Flat design with thick 1-2px borders in #525252. Solid colors only.

Distinctive elements: ① LED status indicators — small colored circles that pulse when running. ② Thick horizontal progress bars inside cards (energy-bar style). ③ Industrial numbering: WU-001, WU-002 in bold mono font. ④ Orange alert banner on cards with pending tool requests: "TOOL REQUEST: bash go test". ⑤ Bottom drawer chat with compact message bubbles and monospace timestamps.

Product context: OrchestraOS decomposes tasks into Work Units (WU). Each WU is executed by an agent in a sandbox. Show WU-002 "Implement event envelope" in RUNNING column with orange LED pulsing, progress bar at 60%, and alert banner. WU-003 in PENDING. WU-001 and WU-004 in DONE with green LED.
```

### Prompt 3 — Mission Control Timeline
```
A NASA/SpaceX mission control timeline interface for an AI agent orchestration system called OrchestraOS.

Palette: Deep space black-blue #0A0E1A background, panel #111827, mission cyan #00E5CC accent for LIVE and RUNNING states, cool white #F0F4F8 text, neon green #00C853 for success, bright alert red #FF1744 for errors.

Typography: Exo 2 (futuristic, aerospace-style) for UI text. Source Code Pro for all technical data, timestamps, event codes. Feels like a mission console.

Layout: Vertical timeline in the center (60% width). Events alternate left and right of the central timeline line. Each event is a compact card with timestamp, event type badge, and description. Right side panel (30%) shows the chat/context when an event is selected. Top shows mission header: "MISSION: task-099 — Event Envelope System" with status "IN PROGRESS" in cyan.

Density: Dense. Technical data visible at all times. Small but readable text. Console-like information density.

Shapes: 2px border radius. Sharp, technical, precise.

Effects: Subtle grid background pattern (like graph paper) at 5% opacity. MINIMAL cyan glow (0 0 10px rgba(0,229,204,0.15)) ONLY on currently running elements. No other glows.

Distinctive elements: ① Central vertical timeline with circular markers color-coded by state. ② Elapsed time counters: "T+ 02m 14s" next to running events. ③ Event type badges with icons: [INTAKE], [TASK], [WU], [RUN], [CHK], [TOOL]. ④ A [TOOL] event highlighted with cyan border showing inline approval buttons [Approve] [Reject]. ⑤ Heartbeat indicator — a thin horizontal line that pulses when an agent is active.

Product context: OrchestraOS executes tasks through a pipeline: UserMessage → Intake → Task → WorkUnit → Run → Session → Checkpoint. Show this flow in the timeline: 10:23:45 [INTAKE] User message received → 10:23:46 [TASK] task-099 created → 10:23:50 [WU] 6 work units decomposed → 10:24:00 [RUN] run-101 starting wu-002 → 10:24:15 [TOOL] Pending approval: file_write → 10:24:20 [CHK] Checkpoint-003 reached. The [TOOL] event is selected and shows chat panel on the right.
```

### Prompt 4 — DevStudio Multi-Agent IDE
```
A professional IDE-style interface (like VS Code) for an AI agent orchestration system called OrchestraOS.

Palette: VS Code dark #1E1E1E background, sidebar/panel #252526, teal accent #4EC9B0 (like class names in VS Code), light gray text #D4D4D4, light green #B5CEA8 for success, salmon red #F48771 for errors, string orange #CE9178 for tool requests.

Typography: Segoe UI for UI chrome (menus, labels, buttons). Cascadia Code for ALL mono content: chats, logs, IDs, code. True IDE aesthetic.

Layout: Classic IDE layout. Far left: Activity Bar (30px wide) with agent icons as "activities" — Codex-Builder, Review-Bot, Docs-Writer, Test-Agent — each with colored status dot or notification badge. Left sidebar: Explorer tree showing Tasks > Work Units > Runs hierarchy. Center: Editor area with tabs — each tab is an agent's chat/terminal. Bottom panel: Integrated terminal showing event stream/console output.

Density: Dense. IDE-level information density. 13px text, tight spacing, clear panel borders.

Shapes: 0px border radius everywhere. Sharp corners like a real IDE.

Effects: ZERO glow, ZERO blur, ZERO glassmorphism. Only solid 1px borders between panels in #3C3C3C. Flat, functional.

Distinctive elements: ① Activity Bar with agent icons + status dots (green=running, gray=idle, red=error, yellow badge=pending approval). ② Tabs per agent in editor area — active tab has teal bottom border. ③ Chat rendered as terminal output: lines prefixed with "// user:" (gray comment style), "// codex-builder:" (white), and "[TOOL] file_write(...)" (orange string style). ④ Explorer sidebar tree with expandable nodes: task-099 > wu-002 > run-101. ⑤ Bottom terminal panel showing colored event stream: gray timestamps, blue [INTAKE], green [TASK], amber [RUN], red [ERROR].

Product context: OrchestraOS developers manage agents like they manage code. Show Codex-Builder tab active with terminal-style chat. Agent asks for file_write approval. Explorer shows task-099 expanded with 3 work units. Bottom terminal shows recent events.
```

### Prompt 5 — Zen Focus Mode
```
A minimal Japandi-style focus interface for an AI agent orchestration system called OrchestraOS.

Palette: Washi paper white #F7F5F0 background, pure white #FFFFFF for floating cards, terracotta #D4734F accent (warm, earthy), sumi ink #2C2C2C text, sage green #81B29A for success, muted terracotta red #C75146 for errors.

Typography: Noto Sans JP (clean Japanese geometric sans) for all UI text. JetBrains Mono for IDs and timestamps only. Extremely legible, peaceful.

Layout: Single-pane full-screen layout. One agent occupies the entire view. Messages centered in a narrow column (max-width 720px). Top navigation: 4 small dots representing 4 agents — colored by status. No sidebars, no panels, no clutter.

Density: Airy. Generous whitespace. 18px body text. Large touch targets. Breathing room between every element.

Shapes: 24px border radius. Organic, soft, approachable.

Effects: Soft diffused shadows only (0 8px 32px rgba(0,0,0,0.06)). Slow 400ms transitions. NO glow. NO sharp shadows.

Distinctive elements: ① Large circular avatar (64px) at top center with agent name and breathing pulsing animation when thinking. ② Messages are large and readable — human messages right-aligned with soft terracotta tint background, agent messages left-aligned with white card and subtle border. ③ Tool request approval appears as a floating center card that blurs the background behind it — elegant, non-intrusive. ④ 4 status dots at very top: terracotta = running, sage = done, gray = idle, red = error. ⑤ Minimal input bar at bottom: wide, clean, single-line with subtle placeholder "Message Codex-Builder or type /command..."

Product context: OrchestraOS where the human focuses on one agent at a time. Show Codex-Builder active (dot pulsing terracotta). Conversation shows: agent message "Implementing the event envelope schema...", then a floating tool request card "Codex-Builder requests: bash 'go test ./...' — Risk: medium" with large [Approve] and [Reject] buttons. Human message below: "Use table-driven tests". Clean, calm, intentional.
```

---

## Ordem de Geração Recomendada

1. **Zen Focus** (mais diferente — único light mode airy)
2. **Industrial Kanban** (único flat denso)
3. **DevStudio IDE** (paradigma familiar para devs)
4. **Mission Timeline** (visualmente impactante)
5. **Executive Threaded** (mais conservador, bom para comparar)

---

## Ferramentas do AiDesigner a Considerar

| Ferramenta | Uso neste projeto |
|---|---|
| `generate_design` | Principal — geração dos 5 protótipos. Usar modo desktop. |
| `generate_website_design_image` | Alternativa — se `generate_design` não capturar bem layouts densos |
| `generate_branding_kit_variations` | **Útil** — gerar um 3x3 de identidade visual baseada em uma direção antes de detalhar o protótipo |
| `create_brand_kit_from_variation` | **Útil** — salvar uma identidade escolhida como kit para manter consistência numa variação |
| `generate_media_asset_kit` | Gerar ícones/assets isolados se necessário (logo, avatares de agente) |
| `generate_image` | Se precisar de assets específicos (ex: avatares de agente estilizados) |

**Recomendação:** Como temos apenas 5 créditos de design (`generate_design`), devemos usá-los diretamente nos 5 protótipos. O `branding_kit` gera imagens (custa 1 crédito por variação 3x3) — seria útil para explorar, mas consome créditos. **Priorizar geração direta dos 5 designs.**

---

## Critérios de Aceite para Validação

- [ ] Cada protótipo tem paleta, tipografia e layout distintos dos existentes
- [ ] Cada protótipo mostra Chat de Agentes + Humano como elemento central
- [ ] Cada protótipo inclui mecanismo de aprovação de tool (human-in-the-loop)
- [ ] Cada protótipo mostra múltiplos agentes ou a navegação entre eles
- [ ] Cada protótipo usa terminologia do domínio OrchestraOS (Task, Work Unit, Run, Session, Orchestrator)
- [ ] Nenhum protótipo usa metáforas abstratas (mar, folhas, órbitas, cosmos)
- [ ] Todos os arquivos salvos em `design/prototypes/`
