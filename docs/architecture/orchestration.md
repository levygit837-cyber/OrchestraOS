# Orquestracao de Agentes

## Modelo

O OrchestraOS usa um Orchestrator central para transformar pedidos em tasks, decompor trabalho em work units, preparar prompts, criar sandboxes, iniciar agentes, receber eventos, aplicar politicas e devolver evidencias à CLI/GitHub.

Agentes sao workers especializados. Eles nao sao a fonte de verdade do estado da task; eles executam trabalho e reportam eventos estruturados.

## Componentes

### Orchestrator API

Recebe pedidos da CLI, GitHub ou futuramente do painel web/conectores opcionais. Expoe endpoints internos para criar tasks, consultar runs, aprovar ferramentas, pausar, retomar e cancelar execucoes.

### Scheduler

Seleciona quais tasks podem executar, respeitando prioridade, limite de paralelismo, risco, disponibilidade de recursos e dependencias.

No MVP, o limite alvo e de 2 a 5 agentes paralelos.

### Task Graph Planner

Transforma uma task em um DAG de work units. O planner deve declarar dependencias, ownership de caminhos, criterios de aceite, validacoes esperadas e riscos.

O planner nao deve criar dependencias ciclicas. Loops de tentativa e correcao pertencem a runs de agentes, nao ao grafo de tasks.

### Prompt Composer

Monta SystemPrompt e TaskPrompt por work unit usando fragmentos versionados. Cada run deve ter um `PromptSnapshot` para auditoria e reproducibilidade.

### Policy Engine

Decide quais acoes sao permitidas automaticamente, quais exigem aprovacao humana e quais devem ser bloqueadas. A politica deve considerar nivel de autonomia, risco da task, ferramenta solicitada, destino da acao e escopo do sandbox.

### Sandbox Manager

Cria worktree, branch, container, variaveis de ambiente minimas, limites de recursos e diretorios de artefatos. Tambem encerra e limpa ambientes ao fim da task.

### Agent Runtime

Executa o Codex/CLI dentro do sandbox da task. Recebe contexto inicial controlado, instrucoes do projeto, contrato de eventos e limites de permissao.

### Event Store

Persiste eventos de task, run, agente, ferramentas, mensagens, checkpoints, artefatos e decisoes de aprovacao. O Event Store e a base de auditoria e recuperacao.

### Agent Task Ledger

Mantem objetivo, criterios de aceite, todos, bloqueios, riscos, resumo atual e proximo checkpoint da work unit. O agente deve atualizar o ledger em checkpoints.

### Agent Checkpoint

Registra um snapshot estruturado de progresso em ponto seguro da `AgentSession`. Checkpoints devem conter goal atual, goals concluidos, arquivos lidos/modificados, evidencias, bloqueios, riscos, resumo minimo e proximo goal sugerido.

O checkpoint nao cria uma nova sessao por si so. Ele fornece base auditavel para o Orchestrator decidir conclusao, continuacao futura, revisao ou replanejamento.

### Recursive Memory Service

Capacidade futura que deriva memorias de eventos, checkpoints, ledger, artefatos, validacoes e documentos versionados. O servico deve criar e recuperar contexto auxiliar para agentes, sempre com evidencia canonica, deduplicacao e registro de injecao.

Memoria recursiva nao substitui Event Store, ADRs, canvas, ledger ou checkpoints. O Orchestrator deve mediar toda recuperacao e injecao.

### Artifact Manager

Gerencia patches, diffs, logs, resultados de testes, arquivos gerados, links de PR e evidencias de conclusao.

## Ciclo de Vida da Task

Estados recomendados:

| Estado | Descricao |
| --- | --- |
| `created` | Pedido recebido. |
| `triaged` | Escopo, risco e politica avaliados. |
| `planned` | Task graph criado e work units definidas. |
| `scheduled` | Task pronta para execucao. |
| `sandbox_preparing` | Worktree/container sendo preparados. |
| `running` | Agente executando. |
| `waiting_approval` | Execucao pausada aguardando aprovacao de ferramenta ou decisao humana. |
| `paused` | Execucao pausada por comando externo ou politica. |
| `validating` | Validacoes em execucao ou sendo conferidas. |
| `completed` | Task concluida com evidencias. |
| `failed` | Task falhou com erro registrado. |
| `cancelled` | Task cancelada por usuario, politica ou sistema. |

## Worktree Por Task

Cada task deve ter branch e worktree proprios. O worktree nao deve ficar dentro do diretorio versionado do projeto para evitar recursao, sujeira no status e risco de commit acidental.

Diretorios recomendados:

- Local: `~/.local/share/orchestraos/worktrees/{repo_id}/{task_id}`
- Servidor: `/var/lib/orchestraos/worktrees/{repo_id}/{task_id}`

Enquanto o runtime inicial for Codex/CLI, o prefixo de branch recomendado e:

```text
codex/task-{task_id}-{slug}
```

## Comunicacao Entre Agentes

Agentes nao devem abrir canais diretos sem registro. Quando um agente precisar de informacao de outro:

1. O agente cria um evento `agent.query.requested`.
2. O Orchestrator valida politica, escopo e destino.
3. O Orchestrator entrega a pergunta ao agente de destino como comando ou notificacao.
4. A resposta volta como evento estruturado.
5. O Orchestrator registra a troca e entrega ao solicitante.

Isso preserva isolamento de contexto, auditoria e controle de permissao.

## Conclusao de Task

Uma task so deve ser considerada concluida quando houver:

- Diff, commit ou PR associado.
- Validacao executada ou justificativa registrada.
- Agent Task Ledger sem pendencias bloqueantes ou com justificativa registrada.
- Ultimo Agent Checkpoint registrando goal concluido, evidencias e riscos restantes.
- Resumo do que mudou.
- Riscos restantes.
- Evidencia enviada à CLI/GitHub.
