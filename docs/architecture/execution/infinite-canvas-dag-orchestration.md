# Infinite Canvas: Grafo DAG de Orquestração e Navegação de Agentes

## Visão

O **Infinite Canvas** é a interface operacional unificada do OrchestraOS. É um espaço contínuo, navegável e persistente onde o estado vivo do sistema se manifesta como um grafo direcionado acíclico (DAG). Nele, humanos e agentes observam, interagem e navegam por tasks, work units, agentes ativos, barreiras de sincronização e — fundamentalmente — **artefatos persistentes** gerados durante a execução de workflows.

O Canvas não é apenas uma visualização. É uma **superfície de operação**: o que está no Canvas é o que existe no sistema. Se um agente está ativo, ele aparece como um nó no grafo. Se uma aplicação foi construída e dockerizada, ela aparece como um artefato navegável, com preview real e endpoints acessíveis.

## Conceitos Fundamentais

### 1. O Canvas como Fonte de Verdade Visual

O Canvas reflete o estado atual do `OrchestratorService` em tempo real:

- Cada `Task` vira uma região ou camada no canvas.
- Cada `WorkUnit` vira um nó no DAG.
- Cada `AgentSession` vira um avatar ou nó ativo conectado à sua `WorkUnit`.
- Cada `Artifact` vira um nó persistente, reutilizável e inspecionável.
- Edges representam dependências (`blocks`, `requires_artifact`, `conflicts_with`, `informs`) e fluxo de dados.

O Canvas é infinito no sentido de que regiões podem ser criadas, zoom out mostra múltiplas tasks, e zoom in revela o interior de uma `WorkUnit` (checkpoints, logs, diffs, prompts).

### 2. Grafo DAG de Work Units

O grafo principal é o `TaskGraph` (DAG) já definido no domínio:

```text
[Task: Feature Event Envelope]
  |
  +-- [WU-001: Criar schema base]
  |       ^
  |       | requires_artifact
  |       v
  +-- [WU-002: Implementar envelope]
  |       ^
  |       | blocks
  |       v
  +-- [WU-003: Adicionar middleware]
          |
          +-- [Gate: Review obrigatório]
          |
          v
  +-- [WU-004: Validar E2E]
```

No Canvas, isso é representado visualmente com:

- **Nós de WorkUnit**: caixas com status (idle, running, completed, failed).
- **Edges coloridos**: verde para `requires_artifact`, azul para `blocks`, amarelo para `informs`, vermelho para `conflicts_with`.
- **Barreiras**: linhas pontilhadas ou gates entre fases.
- **Caminho crítico**: edges destacados que representam o caminho mais longo.

### 3. Artefatos como Cidadãos de Primeira Classe

Além de agentes e work units, o Canvas hospeda **artefatos persistentes**. Eles não são metadados secundários; são elementos navegáveis do grafo.

#### Tipos de Artefato no Canvas

| Artefato | Tipo | Descrição | Interação no Canvas |
|----------|-----------|---------------------|
| `dockerized_app` | Aplicação completa rodando em container | Preview real (iframe), logs, métricas, acesso a endpoints |
| `backend_service` | API ou serviço backend dockerizado | Testar endpoints, ver Swagger/OpenAPI, logs estruturados |
| `frontend_preview` | Build de frontend servido estaticamente | Preview interativo, inspeção de DOM, responsive testing |
| `database_schema` | Schema SQL ou migrations | Visualizar ER diagram, executar queries de teste |
| `test_report` | Relatório de testes (unit, E2E) | Ver cobertura, falhas, replay de testes |
| `diff_review` | Diff de código produzido | Review inline, comentários, aprovação |
| `prompt_snapshot` | Prompt completo usado em uma run | Inspeção, comparação entre versões |
| `workflow_definition` | Workflow reutilizável que agentes podem seguir | Aplicar em outras tasks, versionar, fork |

#### Ciclo de Vida de um Artefato

1. **Geração**: Durante a execução de uma `WorkUnit`, o agente pode emitir um evento `artifact.generated`. O Orchestrator captura, valida e persiste.
2. **Materialização**: O Orchestrator constrói/implementa o artefato. Para aplicações, isso significa `docker build` e `docker run`.
3. **Aparecimento no Canvas**: O artefato surge como um nó conectado à `WorkUnit` que o gerou, com uma edge do tipo `produced_by`.
4. **Navegação**: Humanos e agentes podem "abrir" o artefato. Uma aplicação dockerizada abre um preview real; um schema abre um diagrama.
5. **Reutilização**: Outros agentes podem "referenciar" o artefato. O Canvas mostra edges `consumes` ou `extends`.
6. **Arquivamento**: Quando a task completa, artefatos podem ser arquivados (parados mas persistidos) ou mantidos ativos conforme política.

### 4. WorkUnits como Ambientes Isolados de Desenvolvimento

Cada `WorkUnit` no Canvas representa não apenas uma tarefa, mas um **ambiente isolado**:

- **Worktree Git dedicada**: `~/.local/share/orchestraos/worktrees/{task_id}/{wu_id}`
- **Sandbox/container**: Docker container com o código da worktree montado
- **Filesystem isolado**: O agente só enxerga o que foi explicitamente permitido
- **Rede controlada**: `--network none` por padrão, ou bridge dedicada quando necessário
- **Variáveis de ambiente**: segredos injetados temporariamente, não persistidos

No Canvas, clicar em uma `WorkUnit` expande seu interior:

- Abas: Código, Terminal, Logs, Checkpoints, Artefatos
- Status em tempo real: CPU, memória, rede do container
- Histórico de eventos: lista auditável de tudo que aconteceu

### 5. Agentes no Canvas

Agentes são representados como "entidades" que **se movem** no Canvas:

- **Spawn**: O Orchestrator cria um `AgentSession`, e um avatar aparece conectado à `WorkUnit` atribuída.
- **Execução**: O avatar pulsa ou brilha quando a sessão está ativa. Linhas tracejadas indicam comunicação WebSocket com o Orchestrator.
- **Checkpoint**: Quando o agente registra um checkpoint, o Canvas mostra um "marco" na trilha do agente.
- **Tool Request**: Pedidos de ferramenta aparecem como notificações pendentes no avatar, com ícones de risco (🟢 safe, 🟡 medium, 🔴 high).
- **Reconfiguração**: Se o agente for reiniciado com novo prompt, o Canvas mostra uma "transição" preservando o ledger.
- **Encerramento**: Ao completar, o avatar se torna um nó histórico, conectado aos artefatos que produziu.

### 6. Workflows como Artefatos de Processo

Um dos artefatos mais poderosos é o **Workflow Definition**. Ele define uma sequência de `WorkUnits` com perfis de agente, gates, barreiras e critérios de aceite que podem ser **reutilizados**.

Exemplo de Workflow no Canvas:

```text
[Workflow: "Feature Completa Backend"]
  |
  +-- [Step 1: Schema] -> agent: code_worker, profile: backend
  +-- [Step 2: Repository] -> depends_on: Step 1
  +-- [Step 3: Service] -> depends_on: Step 2
  +-- [Gate: Review] -> mode: hard, focus: [tests, patterns]
  +-- [Step 4: E2E] -> depends_on: Gate
  +-- [Step 5: Documentar] -> depends_on: Step 4
```

Workflows são:

- **Versionados**: cada mudança gera uma nova versão no Canvas.
- **Instanciáveis**: aplicar um workflow a uma nova task cria um TaskGraph baseado no template.
- **Forkáveis**: agentes podem sugerir melhorias em um workflow, gerando um fork.
- **Executáveis**: o Orchestrator pode executar um workflow como um "macro".

## Interações no Canvas

### Humanos

- **Zoom/Pan**: navegar pelo espaço infinito.
- **Clique em WorkUnit**: abrir painel de detalhes, terminal, logs.
- **Clique em Artefato**: abrir preview (app), diagrama (schema), ou editor (diff).
- **Arrastar**: reorganizar o layout do grafo (posições são persistidas por usuário).
- **Intervir**: pausar um agente, aprovar uma tool, enviar mensagem a uma sessão.
- **Criar Task**: desenhar uma nova região, o Orchestrator sugere decomposição.

### Agentes

Agentes "navegam" o Canvas através da **Observation API** (ADR 0023), mas a visão conceitual é:

- **Leitura de Contexto**: antes de iniciar, o agente recebe um "subgraph" do Canvas relevante à sua `WorkUnit` (work units upstream, artefatos disponíveis, insights publicados no Shared Information Board).
- **Publicação de Artefatos**: ao gerar uma aplicação, o agente emite evento e o Canvas atualiza em tempo real.
- **Consulta de Workflows**: agente pode solicitar "qual workflow devo seguir para esta task?" e receber a definição do Canvas.
- **Navegação de Memória**: o Canvas deriva "trilhas de memória" — linhas no grafo conectando checkpoints, decisões e artefatos que um agente deve conhecer.

## Estados Visuais do Canvas

| Estado | Representação |
| ------ | ------------- |
| Task planejada | Região cinza, nós desenhados mas inativos |
| Task em execução | Região com borda pulsante, nós ativos |
| WorkUnit idle | Nó cinza, sem avatar de agente |
| WorkUnit running | Nó azul, avatar do agente conectado, edge WebSocket pulsando |
| WorkUnit completed | Nó verde, conectado aos artefatos produzidos |
| WorkUnit failed | Nó vermelho, stack trace e logs acessíveis |
| Barreira ativa | Linha pontilhada amarela, WUs atrás aguardando |
| Gate de review | Diamante laranja, review session em andamento |
| Artefato ativo | Nó com ícone específico, preview acessível |
| Artefato arquivado | Nó com opacidade reduzida, restaurável |

## Persistência e Modelagem

O Canvas não é um arquivo de design. É uma **projeção ao vivo** dos dados do sistema:

- **Posições e layouts**: persistidos em uma tabela `canvas_layouts` (por usuário/task).
- **Estado do grafo**: derivado do `TaskGraph`, `WorkUnit`, `Run`, `AgentSession` e `Event Store`.
- **Artefatos**: referenciados na tabela `artifacts`, com metadados de acesso (URI, porta, container_id).
- **Workflows**: armazenados como `Artifact` do tipo `workflow_definition`, versionados.
- **Snapshots**: o Canvas permite "salvar uma vista" — um bookmark de posição e zoom para revisitar uma task complexa.

## Exemplo de Cenário Completo

1. Humano cria task: "Implementar sistema de autenticação".
2. Orchestrator decompõe em TaskGraph com 5 WUs.
3. Canvas mostra o DAG, com WU-001 destacada no caminho crítico.
4. Agente `code_worker` é spawnado em WU-001. Avatar aparece, container sobe.
5. Agente gera artefato: schema SQL. Um nó de artefato aparece conectado a WU-001.
6. Agente gera artefato: aplicação backend dockerizada. O Canvas mostra um nó com ícone de container. Humano clica e vê o Swagger UI rodando.
7. WU-001 completa. Barreira de fase ativada. WU-002 pode iniciar.
8. WU-002 é uma "Review-Session". Agente `reviewer` inspeciona o diff e o artefato rodando.
9. Review aprovado. Gate libera WU-003.
10. WU-003 gera frontend dockerizado. Canvas agora mostra dois artefatos: backend e frontend, com edge `depends_on`.
11. Humano clica no artefato frontend, vê a aplicação real funcionando contra o backend.
12. Tudo completa. Canvas oferece "arquivar task" — artefatos param, mas ficam restauráveis.

## Relação com o Domínio Existente

Este conceito estende o domínio atual sem conflitos:

- `Artifact` já existe no modelo de domínio (domain-model.md). Apenas expandimos os tipos e a interatividade.
- `WorkUnit` já tem `owned_paths` e sandbox. O Canvas torna isso visual.
- `TaskGraph` já é um DAG. O Canvas é sua representação.
- `AgentSession` já reporta via WebSocket. O Canvas consome esses eventos.
- `Workflow` como artefato é novo, mas alinhado com a ideia de reutilização e memória operacional.

## Futuro

- **Massive Agents**: quando o sistema suportar dezenas de agentes, o Canvas usará clustering e filtros.
- **Memória Visual**: trilhas de memória derivadas do Recursive Memory System aparecem como "constelações" conectando checkpoints e decisões.
- **Edição Colaborativa**: múltiplos humanos navegam o mesmo Canvas simultaneamente.
- **Modo Offline**: Canvas pode operar em modo de leitura com snapshot do estado.
- **Exportação**: exportar uma vista do Canvas como imagem, vídeo ou apresentação.
