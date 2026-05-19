# Prompt V1: Command Center do OrchestraOS — Nível Médio (~300 linhas)

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

**Task ativa:**
- task-099: "Implementar Event Envelope System"
- 4 Work Units: wu-001 (Schema SQL), wu-002 (Repository Go), wu-003 (Middleware), wu-004 (E2E Tests)
- Estado: running
- 2 de 4 WUs completadas
- Agente ativo: codex-builder

**Aprovações pendentes:**
- Tool request #1: file_write em internal/events/envelope.go, risco: medium, agente: codex-builder
- Tool request #2: shell.exec "go test -race ./...", risco: high, agente: codex-builder

**Eventos recentes:**
- 10:24:15 — codex-builder atingiu checkpoint-003 em wu-002
- 10:23:50 — wu-001 completada por schema-builder
- 10:23:45 — Orchestrator criou task-099

## 4. Hierarquia Espacial do Layout

A tela tem 100vh total. Distribuição vertical:

- **Top Navigation:** 48px fixo (5% da altura)
- **Action Required Queue:** 180px fixo (25% da altura) — PRIORIDADE MÁXIMA
- **Active Tasks Grid:** 280px (35% da altura)
- **System Health Bar:** 60px (8% da altura)
- **Event Stream:** scrollável, ocupa o restante (27% da altura)

Largura total: 100vw. Padding lateral: 24px. Gap entre seções: 16px.

## 5. Componentes Detalhados

### 5.1 Top Navigation (48px)
- Esquerda: "OrchestraOS" em Inter SemiBold 16px + badge "local" em mono 10px, fundo #25252B, cor #8E8E93
- Centro: Search bar 320px largura, placeholder "Search tasks, agents, events...", ícone de lupa
- Direita: Dot verde pulsante (8px) + "Healthy" em mono 11px + avatar usuário 28px

### 5.2 Action Required Queue (180px altura, scroll horizontal)
Esta é a seção mais importante. Deve chamar atenção imediatamente.

**Container:** fundo #1A1A1F, borda inferior 1px #3A3A44, padding 16px 24px.
**Header:** "Action Required" em Inter SemiBold 14px + badge circular vermelho com número (ex: "2") em mono 11px.

**Cards (280px largura, 140px altura, gap 12px):**

Card de Tool Request:
- Borda esquerda 3px: verde (#00CC66) para safe, âmbar (#FF9500) para guarded, vermelho (#FF3B30) para destructive
- Ícone da ferramenta (24px) no topo esquerdo
- Nome da ferramenta em mono 12px: "file_write"
- Target em mono 11px cor #8E8E93: "internal/events/envelope.go"
- Risk badge: "MEDIUM" em mono 10px, fundo #FF950020, texto #FF9500, padding 2px 6px, border-radius 4px
- Agente: avatar 20px + "codex-builder" em mono 11px
- Dois botões na base: [Approve] verde, [Reject] vermelho, ambos 80px largura, 28px altura, mono 10px uppercase

Card de Review Request:
- Borda esquerda 3px azul (#007AFF)
- Thumbnail miniatura do diff (64px altura, código colorido sutil)
- "PR-102" em mono 12px
- Agente que criou: "review-bot"
- Botão [Review] azul

Card de Failure:
- Borda esquerda 3px vermelha (#FF3B30)
- Ícone de erro (x octógono)
- "wu-003 failed" em mono 12px
- Snippet do erro em mono 10px cor #FF3B30, 2 linhas máximo, truncate com ellipsis
- Botão [Investigate] vermelho outline

**Empty state:** Se não há ações, mostrar ícone check verde 32px + "All clear — no actions required" em Inter 14px cor #8E8E93, centralizado verticalmente na seção de 180px.

### 5.3 Active Tasks Grid (280px altura)

**Header:** "Active Tasks" em Inter SemiBold 14px + "2 running" em mono 11px cor #8E8E93.

**Grid:** 2 colunas no desktop (gap 16px), 1 coluna em mobile.

**Task Card (largura flexível, min 400px, altura 240px):**
- Fundo #25252B, border-radius 8px, borda 1px #3A3A44
- Padding: 16px

**Card conteúdo:**
- Topo: ID em mono 11px cor #5C5C66 ("task-099") + badge de estado
- Título: "Implementar Event Envelope System" em Inter SemiBold 15px, cor #FFFFFF, 2 linhas máximo
- Progresso: "2 of 4 work units completed" em mono 11px cor #8E8E93
- Barra de progresso: fundo #3A3A44, preenchimento #007AFF, altura 4px, border-radius 2px, largura 50%
- Mini-DAG horizontal: 4 círculos 12px conectados por linhas 2px. Cores: verde (completado), azul (running), cinza (pending). Mostra visualmente o progresso do grafo.
- Agente atual: avatar 24px + "codex-builder" em mono 12px + dot verde pulsante 6px
- Próximo passo: "Checkpoint-004 pending" em mono 11px cor #FF9500
- Hover: borda muda para #007AFF50, cursor pointer, aparece botão "Open Canvas →" no canto inferior direito

**Badge de estado:**
- Running: fundo #007AFF20, texto #007AFF, mono 10px uppercase, padding 2px 8px, border-radius 4px
- Paused: fundo #FF950020, texto #FF9500
- Review: fundo #9047FF20, texto #9047FF

### 5.4 System Health Bar (60px altura)

Uma linha horizontal com 4 métricas, distribuídas igualmente.

Cada métrica:
- Label em mono 10px uppercase cor #5C5C66 (ex: "AGENTS ACTIVE")
- Valor em Inter Bold 20px cor #FFFFFF (ex: "3/5")
- Indicador secundário: para agentes, dot verde 6px pulsante se < 100%. Para success rate, mini sparkline SVG 40px × 16px mostrando tendência.

Métricas:
1. Agents Active: "3/5" + dot verde
2. Success Rate: "94%" + sparkline verde
3. Events/min: "12" + sinal de estabilidade (—)
4. Pending Approvals: "2" + badge âmbar 16px

Separadores verticais: 1px #3A3A44 entre métricas.

### 5.5 Event Stream (altura flexível, scroll vertical)

**Header:** "Recent Events" em Inter SemiBold 14px + "Live" em mono 10px cor #00CC66 com dot verde pulsante.

**Lista:** padding 0, gap 0. Cada item é uma linha horizontal completa.

**Event Item (altura 40px, padding 0 24px, hover fundo #25252B):**
- Timestamp: "10:24:15" em mono 11px cor #5C5C66, largura fixa 64px
- Badge de tipo: "CHK" em mono 9px uppercase, padding 2px 6px, border-radius 3px. Cores: CHK=#00CC6620 texto #00CC66, EVT=#5AC8FA20 texto #5AC8FA, TOL=#FF950020 texto #FF9500, ERR=#FF3B3020 texto #FF3B30
- Entidade: "wu-002" em mono 11px cor #007AFF, largura fixa 80px
- Agente: "codex-builder" em mono 11px cor #8E8E93, largura fixa 120px
- Descrição: "Checkpoint-003 reached: schema base completed" em Inter 13px cor #FFFFFF, truncate com ellipsis
- Ícone de expansão: "→" no hover, indica que clicar mostra payload completo

Auto-scroll: novos eventos aparecem no topo com fade-in suave. Máximo 50 eventos visíveis, scroll para ver mais.

## 6. Direção Visual

- Fundo: #1A1A1F em toda a tela
- Cards: #25252B, border-radius 8px, borda 1px #3A3A44
- Texto primário: #FFFFFF, Inter, peso 400-600
- Texto secundário: #8E8E93, Inter, peso 400
- Texto discreto: #5C5C66, mono, peso 400
- IDs e timestamps: Inter Mono, 10-12px
- Estados: cinza=#5C5C66 idle, azul=#007AFF running, verde=#00CC66 success, vermelho=#FF3B30 failure, âmbar=#FF9500 warning, roxo=#9047FF blocked
- Sem gradientes. Sem glows. Sem elementos decorativos. Sem sombras dramáticas (apenas sombra sutil 0 1px 3px rgba(0,0,0,0.3) em cards).
- Densidade: alta. Muita informação por pixel².
- Sensação: profissional, calma, autoritária. Como Linear + Vercel + GitHub Actions.

## 7. Interações Principais

- **Card de aprovação:** clicar [Approve] ou [Reject] executa ação inline (sem navegar para outra tela). Card desaparece com animação de slide-out. Evento é registrado no Event Stream.
- **Task Card:** clicar navega para Task Canvas daquela task. Hover revela "Open Canvas →".
- **Event Item:** clicar expande mostrando payload JSON completo em mono 10px, fundo #0F0F13, border-radius 4px.
- **Métrica "Pending Approvals":** clicar scrolla suavemente para a Action Required Queue.
- **Search bar:** foco expande para 400px, mostra sugestões dropdown.

## 8. Estados Vazios

- Nenhuma task ativa: Ilustração mínima (3 círculos conectados, cinza, 48px) + "No tasks running" em Inter 16px + "Create a new task to get started" em Inter 14px cor #8E8E93 + botão "+ New Task" primário azul.
- Nenhuma ação pendente: Check verde 32px + "All clear" em Inter 16px + "System running smoothly" em Inter 14px cor #8E8E93.
- Nenhum evento: Linha pontilhada cinza + "Waiting for events..." em mono 11px cor #5C5C66.

## 9. Anti-Patterns

- NUNCA usar chat bubbles ou interface conversacional
- NUNCA usar linhas numeradas de código como elemento de UI
- NUNCA usar tema sci-fi com scanlines, grids técnicos, neon
- NUNCA usar kanban como visualização principal
- NUNCA usar cards flutuantes desconectos
- NUNCA usar gradientes coloridos
- NUNCA usar ícones cartoonescos ou ilustrações 3D
- NUNCA mostrar apenas estado atual sem histórico
