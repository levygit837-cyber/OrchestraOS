# Prompt V2: Command Center do OrchestraOS — Nível Grande (~600 linhas)

## 1. Contexto do Sistema

O OrchestraOS é um sistema operacional de orquestração de agentes de IA. Quando um humano cria uma Task, o sistema decompõe automaticamente essa Task em um grafo direcionado acíclico (DAG) de Work Units. Cada Work Unit é executada por um agente especializado em uma sandbox isolada (worktree Git + container Docker). Agentes executam em loops, registram checkpoints, e pedem ferramentas (file_write, shell.exec, etc.) que podem exigir aprovação humana. Todo evento é persistido em um Event Store com trilha de auditoria completa. O sistema suporta intervenção humana em 7 níveis de intensidade (hint → warning → interrupt → pause → restart → terminate → escalate).

**O OrchestraOS NÃO É:** um chatbot, uma IDE, um editor de código, um quadro kanban, um dashboard sci-fi, uma ferramenta de deploy.

**O OrchestraOS É:** um sistema de controle e observabilidade de workflows executados por múltiplos agentes autônomos, com humanos supervisionando em pontos de decisão.

## 2. Propósito desta Tela

Esta é a TELA PRINCIPAL — o default quando o usuário abre o OrchestraOS. Em 80% do tempo, o usuário está nesta tela. Ela deve responder imediatamente:

1. "O que precisa da minha atenção AGORA?" (aprovações pendentes, falhas, alertas)
2. "O que está rodando no momento?" (tasks ativas, progresso)
3. "Qual a saúde geral do sistema?" (métricas, eventos)

## 3. Dados de Exemplo Concretos

Use estes dados reais no design, não invente genéricos:

**Task ativa 1:**
- task-099: "Implementar Event Envelope System"
- 4 Work Units: wu-001 (Schema SQL), wu-002 (Repository Go), wu-003 (Middleware), wu-004 (E2E Tests)
- Estado: running
- 2 de 4 WUs completadas (wu-001 done, wu-002 running, wu-003 e wu-004 pending)
- Agente ativo: codex-builder
- Checkpoint atual: checkpoint-003
- Tempo decorrido: 12m 34s

**Task ativa 2:**
- task-100: "Sistema de Autenticação JWT"
- 5 Work Units: wu-005 (Schema), wu-006 (Service), wu-007 (Middleware), wu-008 (E2E), wu-009 (Docs)
- Estado: paused
- 1 de 5 completadas
- Agente ativo: auth-agent
- Motivo da pausa: aguardando aprovação de tool request

**Aprovações pendentes:**
- Tool request #1: file_write em internal/events/envelope.go, risco: medium, agente: codex-builder, razão: "Adicionar struct EventEnvelope com campos ID, Type, Payload, Timestamp"
- Tool request #2: shell.exec "go test -race ./...", risco: high, agente: codex-builder, razão: "Validar race conditions após alteração no envelope"
- Review request #1: PR-102 para wu-001, agente: review-bot, status: 48 linhas adicionadas, 12 removidas, coverage 94%

**Eventos recentes (últimos 5 minutos):**
- 10:24:15 — codex-builder atingiu checkpoint-003 em wu-002
- 10:24:10 — Tool request enviado por codex-builder (file_write)
- 10:23:50 — wu-001 completada por schema-builder
- 10:23:45 — Orchestrator criou task-099
- 10:23:30 — codex-builder iniciou execução de wu-002
- 10:23:15 — schema-builder completou checkpoint-002 em wu-001
- 10:23:00 — task-099 decomposta em 4 work units

## 4. Hierarquia Espacial do Layout

A tela tem 100vh total. Distribuição vertical em desktop (≥1280px):

- **Top Navigation:** 48px fixo (5% da altura)
- **Action Required Queue:** 180px fixo (25% da altura) — PRIORIDADE MÁXIMA
- **Active Tasks Grid:** 280px (35% da altura)
- **System Health Bar:** 60px (8% da altura)
- **Event Stream:** scrollável, ocupa o restante (27% da altura)

Largura total: 100vw. Padding lateral: 24px. Gap entre seções: 16px.

### Layout em Tablet (768px - 1279px)
- Top Navigation: 48px
- Action Queue: 160px
- Active Tasks: 1 coluna, scroll vertical
- System Health: 50px, 2 métricas por linha
- Event Stream: colapsado em drawer inferior, mostra últimos 3 eventos + "Show All"

### Layout em Mobile (< 768px)
- Top Navigation: 48px, search vira ícone
- Action Queue: 140px, cards maiores (1 por vez, swipe horizontal)
- Active Tasks: 1 coluna, cards empilhados verticalmente
- System Health: scroll horizontal, 2 métricas visíveis
- Event Stream: aba separada, acessível via bottom nav

## 5. Componentes Detalhados com Todos os Estados

### 5.1 Top Navigation (48px)

**Desktop (≥1280px):**
- Esquerda: "OrchestraOS" em Inter SemiBold 16px cor #FFFFFF + badge "local" em mono 10px, fundo #25252B, cor #8E8E93, padding 2px 8px, border-radius 4px
- Centro: Search bar 320px largura, altura 32px, placeholder "Search tasks, agents, events...", ícone de lupa 16px cor #5C5C66, fundo #25252B, border 1px #3A3A44, border-radius 6px. Foco: borda #007AFF, placeholder desaparece, dropdown de sugestões aparece abaixo com fundo #25252B, borda #3A3A44, shadow 0 4px 12px rgba(0,0,0,0.4)
- Direita: Dot verde pulsante (8px, animação pulse 2s infinite) + "Healthy" em mono 11px cor #00CC66 + avatar usuário 28px, border-radius 50%, border 2px #3A3A44

**Estados da Top Navigation:**
- Normal: fundo #1A1A1F, border-bottom 1px #3A3A44
- Scroll (quando usuário scrolla para baixo): fundo ganha blur backdrop-filter: blur(12px), background rgba(26,26,31,0.9)
- Search focado: dropdown aparece com 5 sugestões recentes, cada item: ícone + texto em Inter 13px + atalho keyboard em mono 10px cor #5C5C66
- Sistema degradado: dot verde vira âmbar pulsante + "Degraded" em cor #FF9500
- Sistema offline: dot vermelho + "Offline" em cor #FF3B30

### 5.2 Action Required Queue (180px altura, scroll horizontal)

**Container:**
- Fundo: #1A1A1F com borda inferior 1px #3A3A44
- Padding: 16px 24px
- Position: relative
- Quando há itens pendentes: borda inferior ganha gradiente sutil left-to-right vermelho→âmbar→verde representando a mix de riscos

**Header:**
- Texto: "Action Required" em Inter SemiBold 14px cor #FFFFFF
- Badge circular: fundo #FF3B30, texto branco mono 11px bold, 20px diâmetro, posicionado -4px acima e à direita do texto
- Quando 0 itens: badge desaparece, texto fica cor #5C5C66
- Ícone de campainha 16px à esquerda do texto, cor #FF3B30 quando há itens, #5C5C66 quando vazio

**Cards (280px largura, 140px altura, gap 12px, scroll snap):**

#### Card de Tool Request — Estados:

**Estado Normal:**
- Fundo: #25252B
- Borda: 1px #3A3A44
- Borda esquerda: 3px solid, cor depende do risco
- Border-radius: 8px
- Padding: 12px
- Layout: flex column, justify space-between

**Conteúdo (de cima para baixo):**
1. Linha superior: Ícone ferramenta 20px (wrench para file_write, terminal para shell.exec) cor #8E8E93 + nome da ferramenta em mono 11px uppercase cor #FFFFFF
2. Target: "internal/events/envelope.go" em mono 10px cor #8E8E93, truncate com ellipsis, max-width 100%
3. Descrição: "Adicionar struct EventEnvelope..." em Inter 12px cor #8E8E93, 2 linhas máximo, line-clamp
4. Linha de metadados: Risk badge + Agente
   - Risk badge: mono 9px uppercase, padding 2px 6px, border-radius 3px
     - SAFE: fundo #00CC6620, texto #00CC66, borda 1px #00CC6640
     - GUARDED: fundo #FF950020, texto #FF9500, borda 1px #FF950040
     - DESTRUCTIVE: fundo #FF3B3020, texto #FF3B30, borda 1px #FF3B3040
   - Agente: avatar 18px + "codex-builder" em mono 10px cor #8E8E93
5. Linha de ações (base do card): [Approve] + [Reject]
   - Approve: fundo #00CC66, texto #000000, mono 10px uppercase bold, padding 4px 12px, border-radius 4px, hover: brightness 1.2
   - Reject: fundo transparente, texto #FF3B30, borda 1px #FF3B30, mono 10px uppercase, padding 4px 12px, border-radius 4px, hover: fundo #FF3B3020

**Estado Hover:**
- Card sobe 1px (transform: translateY(-1px))
- Shadow: 0 4px 12px rgba(0,0,0,0.2)
- Borda: #007AFF40 (azul sutil)

**Estado Aprovando (loading):**
- Approve button vira spinner 14px azul + texto "Processing..." em mono 9px
- Reject button desabilitado, opacidade 0.5
- Card não pode ser interagido

**Estado Aprovado (sucesso):**
- Card slide-out para direita com fade (transform: translateX(100px), opacity: 0, transition 0.3s)
- Evento "tool.approved" aparece no Event Stream
- Badge de Action Required decrementa em 1

**Estado Rejeitado:**
- Card slide-out para esquerda com fade
- Evento "tool.rejected" aparece no Event Stream
- Pode disparar notificação para agente

#### Card de Review Request:
- Borda esquerda: 3px #007AFF
- Thumbnail: container 64px altura, fundo #1A1A1F, border-radius 4px, mostra diff colorido s (linhas verdes/vermelhas em mono 8px, sintaxe sutil)
- "PR-102" em mono 12px cor #007AFF
- Stats: "+48 / -12" em mono 10px cor #00CC66 / #FF3B30 + "94% cov" em mono 10px cor #8E8E93
- Agente: "review-bot"
- Botão [Review]: fundo #007AFF20, texto #007AFF, borda 1px #007AFF40, hover: fundo #007AFF40

#### Card de Failure:
- Borda esquerda: 3px #FF3B30
- Ícone: octógono X 20px cor #FF3B30
- "wu-003 failed" em mono 12px cor #FF3B30
- Snippet do erro: 2 linhas, mono 10px cor #FF3B3080, line-clamp
- Botão [Investigate]: fundo transparente, texto #FF3B30, borda 1px #FF3B3040, hover: fundo #FF3B3010

**Empty state:**
- Container centralizado na área de 180px
- Ícone: 3 círculos conectados formando check, 48px, cor #00CC6630
- Texto: "All clear — no actions required" em Inter 16px cor #8E8E93
- Subtexto: "System is running smoothly" em Inter 13px cor #5C5C66

### 5.3 Active Tasks Grid (280px altura)

**Header:**
- Esquerda: "Active Tasks" em Inter SemiBold 14px cor #FFFFFF
- Direita: "2 running" em mono 11px cor #8E8E93 + "1 paused" em mono 11px cor #FF9500
- Border-bottom: 1px #3A3A44, padding-bottom 8px, margin-bottom 12px

**Grid:**
- Desktop: 2 colunas, gap 16px
- Tablet: 1 coluna
- Mobile: 1 coluna, scroll horizontal com snap

**Task Card (altura 240px):**
- Fundo: #25252B
- Border: 1px #3A3A44
- Border-radius: 8px
- Padding: 16px
- Transition: all 0.2s ease

**Card — Task Running:**
1. Header do card:
   - ID: "task-099" em mono 11px cor #5C5C66
   - Badge: "RUNNING" em mono 9px uppercase, fundo #007AFF20, texto #007AFF, padding 2px 8px, border-radius 4px
   - Tempo: "12m 34s" em mono 10px cor #5C5C66, alinhado à direita

2. Título: "Implementar Event Envelope System" em Inter SemiBold 15px cor #FFFFFF, line-height 1.4, max 2 linhas, truncate

3. Progress section:
   - Texto: "2 of 4 work units completed" em mono 11px cor #8E8E93
   - Barra: container 100% largura, 4px altura, border-radius 2px, fundo #3A3A44
   - Preenchimento: 50%, fundo #007AFF, border-radius 2px, transition width 0.5s ease
   - Mini-DAG: 4 círculos 12px conectados por linhas 2px
     - Círculo 1: fundo #00CC66 (completado), ícone check 8px branco
     - Círculo 2: fundo #007AFF (running), dot branco 4px pulsante
     - Círculo 3: fundo #3A3A44 (pending), borda 1px #5C5C66
     - Círculo 4: fundo #3A3A44 (pending), borda 1px #5C5C66
     - Conexões: linhas 2px #3A3A44 entre círculos

4. Agente atual:
   - Avatar: 24px, border-radius 50%, border 2px #007AFF
   - Nome: "codex-builder" em mono 12px cor #FFFFFF
   - Status dot: 6px, #007AFF, animação pulse 2s infinite, posicionado -2px abaixo do avatar

5. Próximo passo:
   - Ícone: flag 12px cor #FF9500
   - Texto: "Checkpoint-004 pending" em mono 11px cor #FF9500

6. Hover overlay (aparece suavemente):
   - Fundo: rgba(37,37,43,0.95)
   - Botão: "Open Canvas →" em Inter SemiBold 13px cor #FFFFFF, alinhado centro
   - Seta anima para direita no hover

**Card — Task Paused:**
- Badge: "PAUSED" em mono 9px, fundo #FF950020, texto #FF9500
- Barra de progresso: preenchimento #FF9500
- Mini-DAG: círculo running vira âmbar
- Motivo da pausa: "Aguardando aprovação" em mono 11px cor #FF9500
- Agente: avatar com borda âmbar

**Card — Task Completed:**
- Badge: "COMPLETED" em mono 9px, fundo #00CC6620, texto #00CC66
- Barra: 100%, preenchimento #00CC66
- Mini-DAG: todos os círculos verdes com check
- Não mostra "próximo passo", mostra "Completed at 10:45:22" em mono 11px cor #8E8E93

**Card — Task Failed:**
- Badge: "FAILED" em mono 9px, fundo #FF3B3020, texto #FF3B30
- Barra: preenchimento vermelho até ponto de falha
- Mini-DAG: círculo failed vermelho com X
- Erro: snippet em mono 10px cor #FF3B30, 1 linha

**Card vazio / placeholder:**
- Fundo #25252B com border dashed 1px #3A3A44
- Ícone + "Create new task" em Inter 14px cor #8E8E93
- Botão "+ New Task" centralizado

### 5.4 System Health Bar (60px altura)

Container: flex, justify space-around, align center, padding 0 24px, border-top 1px #3A3A44, border-bottom 1px #3A3A44.

**Métrica 1: Agents Active**
- Label: "AGENTS ACTIVE" em mono 9px uppercase cor #5C5C66, letter-spacing 0.5px
- Valor: "3/5" em Inter Bold 20px cor #FFFFFF
- Sub: dot verde 6px + "healthy" em mono 10px cor #00CC66
- Quando < 100%: valor fica cor #FF9500, dot âmbar pulsante
- Quando 0/5: valor cor #FF3B30, dot vermelho, fundo da métrica ganha #FF3B3005

**Métrica 2: Success Rate**
- Label: "SUCCESS RATE" em mono 9px uppercase cor #5C5C66
- Valor: "94%" em Inter Bold 20px cor #FFFFFF
- Sub: sparkline SVG 48px × 20px, linha #00CC66, stroke 1.5px, fill #00CC6610, mostra tendência dos últimos 20 pontos
- Quando < 90%: valor cor #FF9500, sparkline vira âmbar
- Quando < 80%: valor cor #FF3B30, sparkline vermelho

**Métrica 3: Events/min**
- Label: "EVENTS/MIN" em mono 9px uppercase cor #5C5C66
- Valor: "12" em Inter Bold 20px cor #FFFFFF
- Sub: "stable" em mono 10px cor #8E8E93 + ícone de estabilidade (linha horizontal)
- Quando varia > 50%: sub vira "spiking ↑" em cor #FF9500

**Métrica 4: Pending Approvals**
- Label: "PENDING" em mono 9px uppercase cor #5C5C66
- Valor: "2" em Inter Bold 20px cor #FF9500
- Sub: badge âmbar 16px + "need action" em mono 10px cor #FF9500
- Clicável: hover underline, cursor pointer, rola para Action Queue
- Quando 0: valor "0" cor #00CC66, sub "all clear" cor #00CC66

Separadores: 1px #3A3A44, altura 24px, entre cada métrica.

### 5.5 Event Stream (altura flexível)

**Header (sticky, fundo #1A1A1F, z-index 10):**
- Esquerda: "Recent Events" em Inter SemiBold 14px cor #FFFFFF
- Direita: dot verde 6px pulsante + "Live" em mono 10px cor #00CC66 + timestamp "10:24:15 UTC" em mono 10px cor #5C5C66
- Border-bottom: 1px #3A3A44
- Padding: 12px 24px

**Lista:**
- Padding: 0
- Gap: 0
- Cada item: altura 44px, padding 0 24px, display flex, align center, gap 12px
- Hover: fundo #25252B, cursor pointer
- Alt items (zebra): fundo #1E1E24 sutil

**Event Item — Tipos:**

Evento de Checkpoint (CHK):
- Timestamp: "10:24:15" mono 11px #5C5C66, width 56px
- Badge: "CHK" mono 9px, fundo #00CC6620, texto #00CC66, padding 2px 6px, border-radius 3px
- Entidade: "wu-002" mono 11px #007AFF, width 72px
- Agente: "codex-builder" mono 11px #8E8E93, width 120px
- Descrição: "Checkpoint-003: schema base completed" Inter 13px #FFFFFF, flex 1, truncate
- Payload indicador: "{}" mono 10px #5C5C66 no hover

Evento de Tool Request (TOL):
- Badge: "TOL" mono 9px, fundo #FF950020, texto #FF9500
- Entidade: "file_write" mono 11px #FF9500
- Agente: "codex-builder"
- Descrição: "Requesting write to internal/events/envelope.go"

Evento de Erro (ERR):
- Badge: "ERR" mono 9px, fundo #FF3B3020, texto #FF3B30
- Entidade: "wu-003" mono 11px #FF3B30
- Agente: "schema-builder"
- Descrição: "Stat failed: file not found /root/events.go"
- Fundo do item: #FF3B3005

Evento de System (SYS):
- Badge: "SYS" mono 9px, fundo #5C5C6620, texto #5C5C66
- Entidade: "orchestrator" mono 11px #8E8E93
- Agente: "—"
- Descrição: "Created task-099 with 4 work units"

**Expansão de evento (clique):**
- Item expande para 200px altura
- Mostra payload JSON formatado em mono 10px
- Fundo #0F0F13, padding 12px, border-radius 4px
- Syntax highlight sutil: keys em #007AFF, strings em #00CC66, numbers em #FF9500
- Botão "Copy" no canto superior direito

**Auto-scroll:**
- Novos eventos: fade-in 0.3s, slide-down 0.3s
- Máximo 50 visíveis
- Scroll infinito: carrega mais 50 ao scrollar para baixo
- Indicador de novos eventos: badge "3 new" no topo, clicável para scrollar ao topo

## 6. Direção Visual Completa

### Paleta
- Background: #1A1A1F
- Surface: #25252B
- Surface hover: #2E2E36
- Border: #3A3A44
- Border hover: #007AFF40
- Text primary: #FFFFFF
- Text secondary: #8E8E93
- Text muted: #5C5C66
- Running: #007AFF
- Success: #00CC66
- Failure: #FF3B30
- Warning: #FF9500
- Blocked: #9047FF
- Info: #5AC8FA

### Tipografia
- Títulos de seção: Inter SemiBold 14px, cor #FFFFFF
- Labels: Inter Medium 12px, cor #8E8E93
- Corpo: Inter Regular 13px, cor #FFFFFF
- IDs/Timestamps/Metrics: Inter Mono Regular 10-12px
- Badges: Inter Mono Medium 9-10px uppercase
- Botões: Inter Mono Bold 10px uppercase

### Espaçamento
- Padding lateral global: 24px
- Gap entre seções: 16px
- Gap entre cards: 12px
- Padding interno cards: 12-16px
- Border-radius padrão: 8px (cards), 4px (badges), 6px (inputs)

### Sombras
- Cards: 0 1px 3px rgba(0,0,0,0.2)
- Cards hover: 0 4px 12px rgba(0,0,0,0.3)
- Dropdowns: 0 4px 12px rgba(0,0,0,0.4)
- Nenhuma sombra em elementos flat (métricas, event items)

### Animações
- Transitions padrão: all 0.2s ease
- Hover cards: translateY(-1px) + shadow, 0.2s
- Button hover: brightness(1.1), 0.15s
- Badge pulse: opacity 0.5→1, 2s infinite
- Event fade-in: opacity 0→1 + translateY(-4px→0), 0.3s
- Card slide-out: translateX(0→100px) + opacity 1→0, 0.3s
- Progress bar: width transition 0.5s ease

## 7. Fluxos de Navegação

**Fluxo 1: Aprovação Rápida**
1. Usuário abre Command Center
2. Vê Action Required Queue com 2 cards
3. Lê card de file_write, entende o contexto
4. Clica [Approve]
5. Card slide-out para direita
6. Badge decrementa de 2→1
7. Evento "tool.approved" aparece no Event Stream
8. Task Card atualiza (próximo passo muda)

**Fluxo 2: Investigar Falha**
1. Usuário vê card de falha na Action Queue
2. Clica [Investigate]
3. Navega para Task Canvas da task afetada
4. Task Canvas abre com a WU falhada selecionada
5. Painel de detalhes mostra logs de erro

**Fluxo 3: Criar Nova Task**
1. Usuário clica "+ New Task" ou usa search
2. Modal/drawer abre com formulário
3. Usuário descreve a task
4. Orchestrator decompõe em DAG
5. Nova Task Card aparece no grid
6. Evento "task.created" no Event Stream

**Fluxo 4: Monitorar Progresso**
1. Usuário observa Task Cards
2. Vê mini-DAG animando (círculos mudando de cor)
3. Barra de progresso aumentando
4. Event Stream mostrando checkpoints em tempo real
5. Clica "Open Canvas" para ver detalhes completos

## 8. Estados de Erro e Loading

**Erro de conexão com agente:**
- Task Card: avatar fica cinza, dot vermelho, mensagem "Agent offline" em mono 10px #FF3B30
- Ação automática: Orchestrator tenta reconectar após 30s

**Erro de sistema (Orchestrator down):**
- Top Navigation: dot vermelho + "Offline"
- System Health: todas as métricas em vermelho
- Action Queue: congelado, timestamp do último evento
- Banner global: "System connection lost. Retrying..." em Inter 14px, fundo #FF3B3010, borda #FF3B30, padding 12px 24px

**Loading inicial:**
- Skeleton screens: retângulos cinzas animados (#25252B → #2E2E36 → #25252B, 1.5s infinite)
- 4 skeletons na Action Queue
- 2 skeletons no Active Tasks
- 4 skeletons no System Health
- 10 skeletons no Event Stream

**Loading de ação:**
- Botão vira spinner 14px + texto
- Card não interagível
- Timeout após 10s: mostra "Retry" ou "Cancel"

## 9. Anti-Patterns

- NUNCA usar chat bubbles ou interface conversacional
- NUNCA usar linhas numeradas de código como elemento de UI
- NUNCA usar tema sci-fi com scanlines, grids técnicos, neon
- NUNCA usar kanban como visualização principal
- NUNCA usar cards flutuantes desconectos
- NUNCA usar gradientes coloridos
- NUNCA usar ícones cartoonescos ou ilustrações 3D
- NUNCA mostrar apenas estado atual sem histórico
- NUNCA usar fontes decorativas (serifadas, futuristicas)
- NUNCA usar sombras coloridas ou glows
