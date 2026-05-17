# Decisão de Design: Dual Mode — Command Center + Task Canvas

> Data: 2026-05-17
> Status: Decidido
> Contexto: Definição do componente principal da interface do OrchestraOS

---

## 1. Problema

Ao abrir o OrchestraOS, o usuário precisa de uma interface imediata que responda à pergunta:

> **"O que está acontecendo no sistema e o que eu preciso fazer agora?"**

As alternativas conhecidas não se aplicam ao OrchestraOS:

| Alternativa | Por que não serve |
|---|---|
| **Chat Interface** (Claude Code, Codex) | O OrchestraOS não é conversação com 1 agente. São múltiplos agentes autônomos executando um DAG. |
| **IDE / Editor** | O OrchestraOS não é edição de código. Código é produzido por agentes em sandboxes. |
| **Kanban Board** | Esconde a informação mais crítica: dependências entre work units (`blocks`, `requires_artifact`). |
| **DAG Canvas como tela única** | Visualmente impressionante, mas sobrecarrega o usuário com o grafo completo quando ele só precisa aprovar uma tool ou verificar uma falha. |

**Conclusão:** Não existe referência exata no mercado. Precisamos de um paradigma novo.

---

## 2. Decisão

Adotar arquitetura **Dual Mode** com duas interfaces distintas, hierarquizadas:

### Modo 1: Command Center (Default / Aba Principal)
Interface prioritária que o usuário vê ao abrir o sistema. Foco em **AÇÃO e OBSERVABILIDADE DE ALTO NÍVEL**.

### Modo 2: Task Canvas (Drill-down / Contexto)
Interface secundária acessada ao explorar uma task específica. Foco em **NAVEGAÇÃO ESPACIAL e CONTEXTO COMPLETO**.

---

## 3. Modo 1: Command Center

### Propósito
Responde: *"O que precisa da minha atenção agora? O que está rodando? Qual o estado geral?"*

### Componentes Principais

#### 3.1 Action Required Queue (Prioridade Máxima)
Fila horizontal ou grid de cards do que exige intervenção humana **imediata**:

- **Tool Requests pendentes** (approve/reject/escalate)
- **Falhas que precisam de decisão** (retry/terminate/replan)
- **Warnings que escalaram** (heartbeat lost, loop detected)
- **Reviews pendentes** (PR/diff approval)

Cada card mostra:
- Tipo de ação (file_write, shell.exec, review, etc.)
- Entidade afetada (task, WU, arquivo)
- Nível de risco (safe / guarded / destructive)
- Agente solicitante
- Botões de ação primária

#### 3.2 Active Tasks Overview
Lista/cards de tasks em execução com progresso resumido:

- Título da task
- Progresso (WU completadas / total)
- Estado geral (running / paused / review)
- Agente principal alocado
- Próximo checkpoint ou ação pendente

#### 3.3 System Health
Métricas rápidas em uma linha:

- Agentes ativos / total
- Taxa de sucesso (últimas 24h)
- Throughput de eventos
- Tool requests pendentes

#### 3.4 Recent Events Stream
Stream vertical cronológico dos últimos eventos do Event Store:

- Timestamp
- Tipo de evento (badge colorido)
- Entidade (task, WU, agente)
- Descrição curta

#### 3.5 Quick Actions
Botões de ação global:

- `+ New Task` (criar nova task)
- `Pause All` (pausar todas as runs)
- `View Logs` (logs globais do sistema)
- `System Settings`

### Estados do Command Center

| Estado | Descrição |
|---|---|
| **Vazio / Bootstrap** | Nenhuma task criada ainda. Prompt claro para criar primeira task. |
| **Idle** | Tasks existem, mas nenhuma rodando. Mostra tasks anteriores e quick actions. |
| **Active** | Tasks em execução. Action Queue + Active Tasks + System Health visíveis. |
| **Attention Required** | Action Queue tem itens pendentes. Destaque visual no queue. |
| **System Alert** | Falha crítica ou múltiplos agentes down. Alerta global no topo. |

---

## 4. Modo 2: Task Canvas

### Propósito
Responde: *"Como essa task está estruturada? Quem depende de quem? O que esse agente está fazendo? Qual o contexto completo?"*

### Componentes Principais

#### 4.1 DAG Visualization (Centro)
Canvas navegável (pan/zoom) mostrando o Task Graph:

- **Nós**: Work Units (retângulos com ID, título, estado, agente)
- **Arestas**: Dependências (`blocks`, `requires_artifact`, `conflicts_with`, `informs`)
- **Cores de estado**: idle (cinza), running (azul), completed (verde), failed (vermelho), blocked (roxo)
- **Cores de aresta**: verde (requires_artifact), azul (blocks), amarelo (informs), vermelho (conflicts_with)
- **Caminho crítico**: destacado automaticamente
- **Agentes**: avatares conectados às WUs em execução
- **Artefatos**: nós adicionais conectados por `produced_by`

#### 4.2 Detail Panel (Lateral ou Inferior)
Painel que abre ao clicar em uma WU:

- Abas: Overview | Checkpoints | Events | Artifacts | Terminal | Tool Requests
- Checkpoints: timeline vertical com resumos
- Terminal: sandbox ao vivo (se running)
- Tool Requests: lista de pedidos com ações

#### 4.3 Context Bar (Topo)
Breadcrumb + metadados da task:

- `OrchestraOS > task-099: Event Envelope`
- Status da task
- Botão para voltar ao Command Center

### Interações

| Ação | Resultado |
|---|---|
| Pan (drag no canvas) | Move a visão do grafo |
| Zoom (scroll) | Aumenta/diminui zoom |
| Click em WU | Abre Detail Panel |
| Hover em WU | Tooltip com resumo rápido |
| Drag de WU | Reorganiza posição (persistido) |
| Click em Artefato | Abre preview (app, schema, diff) |

---

## 5. Fluxo de Navegação

```
[Abre OrchestraOS]
        ↓
[Command Center] ←────────────────────────┐
        ↓                                  │
[Clique em Task]                           │
        ↓                                  │
[Task Canvas]                              │
        ↓                                  │
[Clique em WU → Detail Panel]              │
        ↓                                  │
[Aprova/Rejeita Tool Request]              │
        ↓                                  │
[Volta para Command Center] ───────────────┘
```

**Regra:** Qualquer ação de aprovação/rejeição pode ser feita tanto no Command Center (rápido) quanto no Task Canvas (com contexto completo).

---

## 6. Frequência de Uso Estimada

| Interface | Frequência | Quando |
|---|---|---|
| **Command Center** | 80% do tempo | Abrir sistema, aprovações rápidas, verificar estado, criar tasks |
| **Task Canvas** | 20% do tempo | Investigar falhas, revisar contexto completo, entender dependências, debug |

---

## 7. Anti-Patterns Evitados

1. **NUNCA** abrir direto no DAG/Canvas — sobrecarrega o usuário
2. **NUNCA** exigir que o usuário navegue pelo grafo para fazer aprovações simples
3. **NUNCA** mostrar chat como interface primária
4. **NUNCA** forçar kanban como visualização única
5. **NUNCA** misturar observabilidade de alto nível com navegação espacial no mesmo modo

---

## 8. Referências Parciais

Nenhuma referência é exata, mas conceitos são inspirados em:

- **Vercel Dashboard** → Command Center (cards de projeto, deploys, métricas)
- **GitHub Actions** → Task Canvas (pipeline visualization, job dependencies)
- **Linear** → Command Center (inbox, issues, quick actions)
- **Temporal UI** → Task Canvas (workflow DAG visualization)
- **Datadog** → Command Center (métricas, event stream, alertas)

**Diferença fundamental:** Nenhum desses sistemas tem agentes autônomos solicitando aprovações em tempo real dentro de um DAG executável.

---

## 9. Próximos Passos (Pendentes de Decisão)

- [ ] Aprofundar design do Command Center (layout exato, componentes, estados)
- [ ] Aprofundar design do Task Canvas (DAG visualization, zoom levels, node types)
- [ ] Definir transições entre modos (animações, persistência de estado)
- [ ] Gerar mockups/protótipos no AIDesigner
- [ ] Definir responsividade (mobile/tablet)
