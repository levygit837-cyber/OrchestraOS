# Orquestracao de Agentes

## Modelo

O OrchestraOS usa uma arquitetura de **Orquestracao Hibrida** composta por dois sistemas cooperativos:

1. **Sistema de Orquestracao Inteligente (LLM)**: Agente Orquestrador Inteligente que toma decisoes estrategicas (decomposicao, diagnostico, selecao de perfis, aprovacoes de risco medio/alto).
2. **Sistema de Orquestracao Deterministico (Go)**: `OrchestratorService` que executa decisoes taticas (transicoes de estado, validacao de dependencias, sandbox, WebSocket, auto-aprovacao de tools seguras).

O `OrchestratorService` e o **control plane central e gatekeeper**. Toda interacao cross-module passa obrigatoriamente por ele. O Agente Inteligente e um **cliente** do sistema, com sua propria sessao, prompt e toolset de decisao.

Agentes executores (`code_worker`, `docs_writer`, `reviewer`, etc.) sao workers especializados. Eles nao sao a fonte de verdade do estado; eles executam trabalho e reportam eventos estruturados.

## Arquitetura Hibrida

```text
┌─────────────────────────────────────────────────────────────┐
│  CAMADA DE DECISAO ESTRATEGICA (LLM)                        │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  Intelligent Orchestrator Agent                     │    │
│  │  - Perfil: orchestrator                             │    │
│  │  - NAO executa codigo, NAO edita arquivos           │    │
│  │  - NAO acessa DB ou servicos diretamente            │    │
│  │  - Decide: decompor, replanear, selecionar perfil,  │    │
│  │    diagnosticar, aprovar tools de risco, intervir   │    │
│  └──────────────────────┬──────────────────────────────┘    │
│                         │                                   │
│                         │ Eventos/Comandos estruturados     │
│                         │ (via Observation API / Event Store)│
│                         ▼                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  CAMADA DE CONTROLE TATICO (Go)                     │    │
│  │  ┌─────────────────────────────────────────────┐    │    │
│  │  │  OrchestratorService                        │    │    │
│  │  │  - Control plane central                    │    │    │
│  │  │  - Valida TODAS as decisoes                 │    │    │
│  │  │  - Transiciona state machines               │    │    │
│  │  │  - Persiste eventos no Event Store          │    │    │
│  │  │  - Orquestra cross-module communication     │    │    │
│  │  └──────────────────────┬──────────────────────┘    │    │
│  │                         │                           │    │
│  │          ┌──────────────┼──────────────┐            │    │
│  │          ▼              ▼              ▼            │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │    │
│  │  │  Task    │  │  Prompt  │  │  Policy  │  ...     │    │
│  │  │ Service  │  │ Service  │  │ Service  │          │    │
│  │  └──────────┘  └──────────┘  └──────────┘          │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  CAMADA DE EXECUCAO (Agentes Workers)               │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐               │    │
│  │  │ code_   │ │ docs_   │ │ review_ │               │    │
│  │  │ worker  │ │ writer  │ │   er    │               │    │
│  │  └────┬────┘ └────┬────┘ └────┬────┘               │    │
│  │       │           │           │                      │    │
│  │       └───────────┼───────────┘                      │    │
│  │                   │ WebSocket                        │    │
│  │                   ▼                                  │    │
│  │          OrchestratorService                         │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## Regra de Ouro: Modulos Nunca Conversam Diretamente

**Nenhum modulo do sistema pode comunicar-se diretamente com outro modulo.**

Toda interacao cross-module ocorre exclusivamente atraves do `OrchestratorService`, que:
1. Recebe a solicitacao (de agente, humano, ou sistema).
2. Valida a operacao (permissoes, estados, regras de negocio).
3. Orquestra a chamada aos servicos de dominio necessarios.
4. Persiste o resultado no Event Store.
5. Retorna a resposta controlada.

Isso inclui:
- Agente Executor -> Agente Executor: comunicacao mediada pelo Orchestrator.
- Agente Inteligente -> Servicos de Domínio: **proibido**. Deve passar pelo OrchestratorService.
- Servico de Task -> Servico de Prompt: **proibido**. Deve passar pelo OrchestratorService.
- Qualquer modulo -> Qualquer modulo: **proibido sem intermediacao do Orchestrator**.

## Componentes

### Intelligent Orchestrator Agent

A camada de inteligencia estrategica. Operacao sob demanda, ativada por triggers especificos.

**Responsabilidades:**
- Receber e interpretar mensagens em linguagem natural
- Decompor tasks em work units com semantica e contexto
- Selecionar perfis dinamicos de agente
- Diagnosticar stalls, loops e anomalias
- Decidir sobre replanejamento apos falha
- Aprovar/negar ferramentas de risco medio e alto
- Sugerir intervencoes em agentes (dicas, pausas, reinicios)
- Escalonar para aprovacao humana

**Documentacao detalhada:** `docs/architecture/intelligent-orchestrator-agent.md`

### OrchestratorService (Go)

O control plane deterministico. Unica entidade com acesso direto aos servicos de dominio.

**Responsabilidades:**
- Receber uma task e executar o fluxo completo
- Validar dependencias entre work units antes de iniciar execucao
- Validar conflitos de `owned_paths`
- Aplicar timeout e retry
- Detectar heartbeat ausente
- Auto-aprovar ferramentas seguras (`risk: safe`)
- Escalar ferramentas de risco para o Agente Inteligente ou humano
- Gerenciar ciclo de vida do WebSocket
- Coordenar comunicacao cross-module
- Validar e aplicar (ou rejeitar) sugestoes do Agente Inteligente

**Metodo principal:**
```
OrchestratorService.RunTask(ctx, taskID, options) -> OrchestratorResult
```

### Observation API

Fronteira controlada atraves da qual o Agente Inteligente observa o sistema.

- Nunca expoe eventos brut
- Produz resumos estruturados, filtrados e classificados
- Detecta anomalias (stall, loop, drift, conflict)
- Respeita escopo, permissao e seguranca

**Documentacao detalhada:** `docs/architecture/orchestrator-observation-api.md`

### Orchestrator API

Recebe pedidos da CLI, GitHub ou futuramente do painel web/conectores opcionais. Expoe endpoints internos para criar tasks, consultar runs, aprovar ferramentas, pausar, retomar e cancelar execucoes.

### Scheduler

Seleciona quais tasks podem executar, respeitando prioridade, limite de paralelismo, risco, disponibilidade de recursos e dependencias.

No MVP, o limite alvo e de 2 a 5 agentes paralelos.

### Task Graph Planner

Transforma uma task em um DAG de work units. O planner pode operar em dois modos:
- **Heuristica local (`local_heuristic_v1`)**: deterministico, baseado em criterios de aceite
- **LLM-based (`llm_gemini_v1`)**: inteligente, com fallback para heuristica

O planner deve declarar dependencias, ownership de caminhos, criterios de aceite, validacoes esperadas e riscos.

### Prompt Composer

Monta SystemPrompt e TaskPrompt por WorkUnit usando fragmentos versionados, categorias canonicas e `SystemProfile`. O `TaskPrompt` nasce depois do perfil sistemico e inclui `TaskPromptDecompose` com ownership, dependencias e limites da unidade do TaskGraph.

Cada run referencia um `PromptSnapshot` para auditoria e reproducibilidade; composicoes identicas reutilizam o snapshot e incrementam `count_used`.

### Policy Engine

Decide quais acoes sao permitidas automaticamente, quais exigem aprovacao humana e quais devem ser bloqueadas. A politica deve considerar nivel de autonomia, risco da task, ferramenta solicitada, destino da acao e escopo do sandbox.

No primeiro corte, decisoes `safe` sao autoaprovadas pelo Go; decisoes de risco medio/alto sao escalonadas para o Agente Inteligente ou humano.

### Sandbox Manager

Cria worktree, branch, container, variaveis de ambiente minimas, limites de recursos e diretorios de artefatos. Tambem encerra e limpa ambientes ao fim da task.

### Agent Runtime

Executa o Codex/CLI dentro do sandbox da task. Recebe contexto inicial controlado, instrucoes do projeto, contrato de eventos e limites de permissao.

### Event Store

Persiste eventos de task, run, agente, ferramentas, mensagens, checkpoints, artefatos e decisoes de aprovacao. O Event Store e a base de auditoria e recuperacao.

### Domain Services

`internal/services` e a fronteira de comando para operacoes que alteram estado operacional.

- `TaskService`, `WorkUnitService`, `RunService` e `AgentSessionService` validam entrada, aplicam state machines, gravam eventos canonicos e atualizam projecoes relacionais na mesma transacao.
- `EventService` envolve o Event Store para append validado, idempotencia por `event_id`, compatibilidade de referencias, consultas e replay estrito.
- `WorkUnitService` serializa a checagem de `owned_paths` por task durante agendamento/inicio, evitando que runs concorrentes ativem work units com paths conflitantes.
- `RunService.Retry` exige `event_id` como chave de idempotencia, aplica timeout/backoff da politica de retry e registra a politica no evento da nova tentativa.
- `AgentSessionService` e a unidade canonica de checkpoints.
- CLI, TUI, runtimes de agente e conectores futuros devem chamar servicos quando houver regra de dominio, transicao de estado, retry, timeout, cancelamento ou auditoria obrigatoria.
- Repositorios continuam como primitivas de leitura e escrita, mas nao decidem transicoes, retry, timeout, conclusao ou compensacao.

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

## Protocolo de Intervencao

O Orchestrator pode intervir em agentes executores em sete niveis de intensidade:

| Nivel | Nome | Comportamento |
|-------|------|---------------|
| 1 | **HINT** | Envia informacao util, agente continua |
| 2 | **WARNING** | Alerta que requer confirmacao |
| 3 | **INTERRUPT** | Agente para no proximo checkpoint |
| 4 | **PAUSE** | Pausa imediata, aguarda resume |
| 5 | **RESTART** | Nova run a partir do ultimo checkpoint |
| 6 | **TERMINATE** | Falha definitiva da work unit |
| 7 | **ESCALATE** | Human assume controle |

A decisao de intervir pode vir do Go (regras deterministicas), do Agente Inteligente (diagnostico estrategico), do humano ou do proprio agente executor.

**Documentacao detalhada:** `docs/architecture/orchestrator-intervention-protocol.md`

## Coordenacao Multi-Agente

O Orchestrator coordena multiplos agentes paralelos atraves de:

- **Task Graph (DAG)**: plano estatico de dependencias e ownership
- **Barrier Synchronization**: sincronizacao de fases entre agentes
- **Shared Information Board**: compartilhamento controlado de descobertas
- **Contract Net**: atribuicao dinamica de work units (futuro)
- **Deadlock/Livelock Detection**: deteccao e resolucao de esperas circulares

**Documentacao detalhada:** `docs/architecture/multi-agent-coordination.md`

## Ciclo de Vida da Task

Estados recomendados:

| Estado | Descricao |
| --- | --- |
| `created` | Pedido recebido. |
| `triaged` | Escopo, risco e politica avaliados pelo Go; Agente Inteligente pode revisar. |
| `planned` | Task graph criado e work units definidas. |
| `scheduled` | Task pronta para execucao. |
| `sandbox_preparing` | Worktree/container sendo preparados. |
| `running` | Agente(s) executando. |
| `waiting_approval` | Execucao pausada aguardando aprovacao de ferramenta ou decisao. |
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

## Comunicacao entre Agente Inteligente e OrchestratorService

A comunicacao segue o mesmo protocolo de eventos dos agentes executores:

```text
Agente Inteligente (LLM)
  |
  |-- Emite eventos estruturados
  |   Ex: orchestrator.suggest_replan
  |       orchestrator.suggest_profile
  |       orchestrator.approve_tool
  |       orchestrator.intervene_run
  |
  v
OrchestratorService (Go)
  |
  |-- Valida comando
  |-- Persiste evento de decisao
  |-- Executa via servico de dominio
  |-- Retorna resultado ou rejeicao
  |
  v
Novo estado no sistema
```

O OrchestratorService pode **rejeitar** uma sugestao do Agente Inteligente se ela violar politicas, estados invalidos ou regras de seguranca. A rejeicao vira evento auditavel.

## Conclusao de Task

Uma task so deve ser considerada concluida quando houver:

- Diff, commit ou PR associado.
- Validacao executada ou justificativa registrada.
- Agent Task Ledger sem pendencias bloqueantes ou com justificativa registrada.
- Ultimo Agent Checkpoint registrando goal concluido, evidencias e riscos restantes.
- Resumo do que mudou.
- Riscos restantes.
- Evidencia enviada a CLI/GitHub.

## Referencias

- `docs/adr/0023-hybrid-intelligent-orchestrator.md`
- `docs/architecture/intelligent-orchestrator-agent.md`
- `docs/architecture/orchestrator-observation-api.md`
- `docs/architecture/orchestrator-intervention-protocol.md`
- `docs/architecture/multi-agent-coordination.md`
- `docs/architecture/communication-protocol.md`
- `docs/architecture/permissions.md`
- `docs/architecture/sandbox-and-autonomy.md`
- `docs/architecture/memory-system.md`
