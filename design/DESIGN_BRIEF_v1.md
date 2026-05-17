# OrchestraOS — Design System Brief v1

> Documento de especificação visual e de interação. Deve ser usado como contexto completo antes de qualquer geração de design no AIDesigner.

---

## 1. O QUE É O OrchestraOS (Contexto Pesado)

O OrchestraOS **NÃO É**:
- Um chatbot com interface conversacional
- Uma IDE ou editor de código
- Um dashboard genérico de "IA agents"
- Um sistema sci-fi/cyberpunk retrô

O OrchestraOS **É**:
Um **sistema operacional de orquestração de agentes** onde:
1. Humanos criam **Tasks** (unidades de trabalho)
2. O Orchestrator decompõe cada Task em um **DAG (Task Graph)** de **Work Units**
3. Cada Work Unit tem dependências reais (`blocks`, `requires_artifact`, `conflicts_with`)
4. **Agentes** são spawnados para executar Work Units em **sandboxes isoladas** (worktree + container)
5. Agentes executam em **loops**, registram **checkpoints**, e pedem **ferramentas** que podem exigir aprovação humana
6. Todo evento é persistido em um **Event Store** com trilha de auditoria completa
7. Humanos intervêm em **7 níveis de intensidade** (hint → warning → interrupt → pause → restart → terminate → escalate)
8. Artefatos (código, apps dockerizadas, schemas, test reports) são **cidadãos de primeira classe**

**Fluxo principal:**
```
UserMessage → Task Creation → Task Graph (DAG) → Work Units →
Runs → Agent Sessions → Checkpoints → Tool Requests →
Approvals/Rejections → Artifacts → Review → Merge
```

---

## 2. REFERÊNCIAS VISUAIS (Inspirações Reais)

### 2.1 Estilo Visual: Profissional, Denso, Moderno

**Inspirações principais:**
- **Linear** (linear.app): densidade de informação, cinzas sofisticados, acentos sutis, tipografia clean
- **Vercel Dashboard**: dashboard técnico com cards, métricas, estados de deploy
- **GitHub Actions**: visualização de pipeline/grafo de jobs com dependências
- **Railway**: observabilidade de serviços, logs, métricas em tempo real
- **Temporal UI**: visualização de workflows como DAGs
- **Datadog**: densidade de dados técnicos sem ser caótica
- **Figma (canvas)**: navegação espacial, zoom/pan, elementos posicionados livremente

**O que absorver de cada um:**
- Linear: paleta, tipografia, espaçamento, "quiet luxury" de interface técnica
- GitHub Actions: como mostrar DAG de execução com estados coloridos
- Railway: cards de serviço com métricas ao vivo
- Temporal: visualização de workflow como grafo
- Datadog: densidade de informação técnica organizada

### 2.2 O que EVITAR (Anti-referências)
- VS Code / IDE: não somos editor de código
- ChatGPT / Claude interface: não somos chat
- Sci-fi HUDs (Mission Control antigo): não somos filme dos anos 90
- Kanban boards (Trello, Jira): perdemos a informação de dependência que é o core do sistema
- Dashboards de "AI Agents" genéricos: muito vazio, muito marketing

---

## 3. ARQUITETURA DE INFORMAÇÃO (O que deve aparecer)

### 3.1 Visão Macro (Dashboard/Orchestrator)

**Elemento central: Task Graph (DAG)**
- Nós = Work Units (com ID, título, estado, agente alocado)
- Arestas = Dependências (com tipo e direção)
- Cores de estado: idle, running, completed, failed, blocked
- Cores de aresta: blocks, requires_artifact, conflicts_with, informs
- Caminho crítico destacado
- Zoom out mostra múltiplas Tasks; zoom in mostra interior de uma Task

**Painéis auxiliares:**
- **Agentes Ativos**: lista de agentes em execução com heartbeat, sandbox, WU alocada
- **Fila de Aprovações**: Tool requests pendentes de intervenção humana (com risco, contexto, ações)
- **Timeline de Eventos**: stream cronológico de eventos do Event Store
- **Métricas do Sistema**: agentes ativos, tasks em andamento, taxa de sucesso, latência

### 3.2 Visão de Task Detail

**Quando clicar em uma Task:**
- DAG expandido da task (maior, interativo)
- Cada Work Unit expansível para mostrar:
  - Checkpoints (sequência, resumo, decisões, riscos)
  - Runs e tentativas (attempt 1, 2, 3...)
  - Eventos relacionados
  - Artefatos produzidos
  - Tool requests e suas respostas
- Abas: Overview | Graph | Events | Artifacts | Checkpoints | Logs

### 3.3 Visão de Work Unit (Zoom In)

**Quando clicar em uma Work Unit:**
- Objetivo, critérios de aceite, owned_paths
- Agente alocado e sessão ativa
- Terminal/sandbox ao vivo (se running)
- Checkpoints como timeline vertical
- Artefatos produzidos (com preview se possível)
- Tool requests (pendentes, aprovados, rejeitados)
- Input area para intervenção humana direta

### 3.4 Visão de Agente

**Perfil do agente:**
- Nome, profile, capabilities
- Estado atual (idle, running, paused)
- Work Unit atual (se running)
- Histórico de execução (tasks completadas, falhas)
- Checkpoints registrados
- Métricas de performance (tempo médio, taxa de sucesso)

### 3.5 Visão de Aprovação/Intervenção

**Fila de aprovações:**
- Tool request com risco (safe/guarded/destructive)
- Contexto completo (o que o agente quer fazer, por quê, em qual WU)
- Preview do impacto (arquivos afetados, diff preview)
- Ações: Approve | Reject | Modify | Escalate | Add Context
- Histórico de intervenções anteriores

---

## 4. DESIGN SYSTEM ESPECÍFICO

### 4.1 Paleta de Cores

**Base (inspirado em Linear + Vercel):**
- Background principal: `#0F1117` (quase preto, não puro)
- Background secundário: `#161922` (cards, painéis)
- Background terciário: `#1E2028` (hover, estados)
- Borda sutil: `#2A2D37` (divisores, bordas de cards)
- Texto primário: `#E8E8E8` (quase branco)
- Texto secundário: `#8B8F98` (cinza médio)
- Texto terciário: `#5C5F6A` (cinza escuro, metadados)

**Estados (cores funcionais):**
- Idle/Pending: `#6B7280` (cinza)
- Running/Active: `#3B82F6` (azul)
- Completed/Success: `#10B981` (verde)
- Failed/Error: `#EF4444` (vermelho)
- Warning/Paused: `#F59E0B` (âmbar)
- Blocked/Waiting: `#8B5CF6` (roxo)
- Info/Checkpoint: `#06B6D4` (ciano)

**Acentos (uso restrito):**
- Destaque primário: `#3B82F6` (azul)
- Destaque secundário: `#8B5CF6` (roxo)
- Destaque terciário: `#10B981` (verde)
- Risco baixo: `#10B981`
- Risco médio: `#F59E0B`
- Risco alto: `#EF4444`

**Regras de cor:**
- NUNCA usar gradientes chamativos ou neon
- NUNCA usar scanlines, grids técnicos óbvios, ou efeitos sci-fi
- Cores de estado devem ser consistentes em TODO o sistema
- Contraste mínimo WCAG AA em todo texto

### 4.2 Tipografia

**Fonte principal: Inter** (ou similar clean sans-serif)
- Títulos: 20-24px, weight 600-700
- Subtítulos: 14-16px, weight 500-600
- Corpo: 13-14px, weight 400
- Metadados/mono: 11-12px, JetBrains Mono ou SF Mono

**Fonte monoespaçada: JetBrains Mono**
- IDs, timestamps, códigos, métricas
- 11-13px, weight 400-500

**Regras tipográficas:**
- IDs de sistema (task-099, wu-002) sempre em mono
- Timestamps sempre em mono
- Código/diffs em mono com syntax highlighting sutil
- Títulos de WU em sans-serif

### 4.3 Espaçamento e Densidade

**Densidade: ALTA (density = comfortable, não compact)**
- Cards com padding 12-16px
- Gap entre elementos 8-12px
- Painéis laterais 280-320px
- Header 48-56px
- Bordas arredondadas: 6-8px (nada de 24px "chat bubbles")

**Regras:**
- Interface deve mostrar MUITA informação sem parecer bagunçada
- Agrupar relacionados, separar não-relacionados
- Hierarquia visual clara com tamanho, peso e cor

### 4.4 Componentes Visuais

**Work Unit Node (no DAG):**
- Retângulo com borda arredondada (6px)
- Cor de borda = estado
- Conteúdo: ID (mono), título (sans), agente alocado (avatar + nome), estado (badge)
- Tamanho: ~200px × ~80px
- Hover: expande ligeiramente, mostra mais info

**Agente Avatar:**
- Círculo 24-28px com iniciais ou ícone
- Borda/badge indicando estado (verde=ativo, cinza=idle)
- Conectado à WU por linha tracejada animada (se running)

**Artefato Node:**
- Retângulo com ícone de tipo (container, schema, diff, etc.)
- Conectado à WU por edge "produced_by"
- Preview disponível no hover/click

**Edge (dependência):**
- Linha curva (bezier) entre nós
- Cor = tipo de dependência
- Seta indicando direção
- Label com tipo (opcional, no hover)

**Approval Card:**
- Card destacado com borda de risco
- Header: tipo de tool, nível de risco, agente solicitante
- Body: descrição do pedido, preview de impacto
- Footer: botões de ação (Approve, Reject, etc.)

**Event Stream:**
- Lista vertical cronológica
- Cada evento: timestamp (mono), tipo (badge), descrição, entidade relacionada
- Agrupamento por task/run

---

## 5. INTERAÇÕES E ESTADOS

### 5.1 Navegação Espacial (Canvas)

- **Pan**: click-drag no canvas vazio
- **Zoom**: scroll wheel ou pinch
- **Zoom to fit**: botão para ver task inteira
- **Zoom to selection**: duplo-click em WU zooma para ela
- **Minimap**: visão geral do canvas no canto (opcional)

### 5.2 Interações nos Nós

- **Click**: abre painel de detalhes da WU
- **Hover**: tooltip com resumo rápido
- **Drag**: reorganiza posição no canvas (persistido)
- **Context menu**: ações disponíveis (pause, restart, view logs, etc.)

### 5.3 Estados do Sistema

**Task States:**
- planned → graph_created → in_progress → review → completed | failed

**Work Unit States:**
- idle → scheduled → running → checkpoint → paused → completed | failed

**Agent Session States:**
- spawning → running → checkpointing → paused → terminated

**Tool Request States:**
- requested → pending_approval → approved → executing → completed | failed

---

## 6. ANTI-PATTERNS VISUAIS (O que NUNCA fazer)

1. **NUNCA** usar chat bubbles ou interface conversacional como elemento principal
2. **NUNCA** usar linhas de código numeradas como representação de agente
3. **NUNCA** usar tema sci-fi com scanlines, grids, neon, fontes futuristicas
4. **NUNCA** usar kanban como visualização principal (esconde dependências)
5. **NUNCA** usar cards flutuantes desconectos sem relações visuais claras
6. **NUNCA** usar gradientes coloridos ou sombras dramáticas
7. **NUNCA** usar ícones cartoonescos ou ilustrações 3D
8. **NUNCA** tratar artefatos como meros anexos — eles são nós no grafo
9. **NUNCA** mostrar apenas estado atual sem histórico/timeline
10. **NUNCA** usar fontes decorativas, serifadas, ou "tech" (Rajdhani, Exo 2, etc.)

---

## 7. EXEMPLO DE CENÁRIO VISUAL

**Cenário: Task "Implementar Event Envelope"**

**Canvas mostra:**
- Região da Task com título e metadados
- DAG com 4 WUs:
  - WU-001 (completed, verde) "Criar schema SQL"
  - WU-002 (running, azul) "Implementar envelope" — agente Codex-Builder conectado
  - WU-003 (blocked, roxo) "Adicionar middleware" — depende de WU-002
  - WU-004 (idle, cinza) "Validar E2E" — depende de WU-003
- Arestas: WU-001 → WU-002 (requires_artifact), WU-002 → WU-003 (blocks), WU-003 → WU-004 (blocks)
- Artefato: Schema SQL produzido por WU-001 (nó verde conectado)
- Painel direito: Event stream mostrando últimos eventos
- Painel inferior: Tool request pendente de Codex-Builder (file_write em events.go)

**Interação:**
- Usuário clica em WU-002 → painel lateral abre com:
  - Checkpoints (3 registrados)
  - Terminal ao vivo do sandbox
  - Artefatos (nenhum ainda)
  - Tool requests (1 pendente)
- Usuário aprova tool request → evento registrado → agente continua

---

## 8. PROMPT DE GERAÇÃO (Para uso no AIDesigner)

**Quando estivermos prontos para gerar, o prompt deve incluir:**

```
Contexto: Sistema operacional de orquestração de agentes de IA.
NÃO É chat, NÃO É IDE, NÃO É sci-fi retrô.

Elemento central: DAG (grafo direcionado acíclico) mostrando Work Units 
como nós e dependências como arestas coloridas.

Estilo visual: Profissional, moderno, denso. Inspirado em Linear, Vercel, 
GitHub Actions, Railway, Temporal. Paleta escura sofisticada. 
Sem gradientes, sem neon, sem scanlines, sem cartoon.

Cores de estado: cinza=idle, azul=running, verde=completed, vermelho=failed, 
roxo=blocked, âmbar=warning.

Tipografia: Inter (sans) + JetBrains Mono ( IDs, timestamps).

Densidade: Alta. Muita informação organizada. Cards com bordas arredondadas 
sutis (6-8px). Sem "chat bubbles" ou cards flutuantes desconectos.

Componentes: Work Unit Nodes, Agent Avatars, Artifact Nodes, Approval Cards, 
Event Stream, Métricas do Sistema.

Interações: Canvas navegável (pan/zoom). Nós clicáveis para detalhes. 
Drag para reorganizar.

Anti-patterns: NÃO usar kanban, NÃO usar chat, NÃO usar IDE, 
NÃO usar sci-fi, NÃO usar cards desconectos.

[Incluir descrição específica da tela a ser gerada]
```

---

## 9. PRÓXIMOS PASSOS

1. Revisar e aprovar este brief
2. Definir se vamos gerar por tela (dashboard, task detail, WU detail, agent profile) 
   ou um design system unificado primeiro
3. Usar scada-terminal como base estrutural APENAS para layout macro 
   (header, painéis, métricas) — substituir kanban por DAG
4. Refinar/gernar com prompt completo deste brief
5. Iterar com feedback antes de finalizar
