# Arquitetura do OrchestraOS

Este documento registra a arquitetura do OrchestraOS, refletindo a **Arquitetura Modular Simplificada** conforme ADR-0019.

## Contexto

O OrchestraOS é um sistema de orquestração de agentes capaz de executar múltiplas tasks em paralelo. Cada agente trabalha com contexto isolado, sandbox próprio e worktree separada por task.

O produto é local-first para desenvolvimento, com desenho pronto para rodar em servidor. A interface inicial é CLI fina, com GitHub como superfície externa principal. O primeiro runtime de agente é Codex/CLI em sandbox.

## Arquitetura Modular Simplificada (ADR-0019)

O OrchestraOS adota uma **Arquitetura Modular Simplificada** conforme ADR-0019, que substituiu a ADR-0015 (Vertical Slice). A arquitetura é organizada em 4 pilares:

### Os 4 Pilares

1. **`internal/domain/`** centraliza **todos** os entity types compartilhados (Task, Run, WorkUnit, Agent, etc.).
2. **Módulos em `internal/modules/`** não importam outros módulos. Zero exceções.
3. **Apenas `internal/bootstrap/` e `internal/modules/orchestrator/`** importam múltiplos módulos.
4. **`repository.go`** é CRUD puro — sem business logic, sem timestamps, sem deduplication.

### Estrutura

- **Domain (`internal/domain/`)**: Todos os entity types compartilhados entre módulos (Task, Run, WorkUnit, Agent, AgentSession, TaskGraph, Trigger, Review, Prompt, etc.).
- **Módulos (`internal/modules/`)**: Cada módulo representa uma entidade de domínio com sua própria lógica, repositório e serviço. Módulos são autônomos e não importam outros módulos.
- **Core (`internal/core/`)**: Infraestrutura compartilhada (apperrors, db, event, eventstore, serialization, statemachine, transition, validation).
- **Bootstrap (`internal/bootstrap/`)**: Injeção de dependências e wiring de serviços com adapters para conectar módulos sem dependências diretas.

## Decisao Arquitetural

A arquitetura inicial sera um **control plane central hibrido com agent workers isolados**.

O OrchestraOS adota uma arquitetura de Orquestracao Hibrida com dois sistemas cooperativos:

1. **Sistema de Orquestracao Inteligente (LLM)**: Um Agente Orquestrador Inteligente que atua como intermediador estrategico. Ele toma decisoes de alto nivel (decomposicao, diagnostico, selecao de perfis, aprovacoes de risco), mas nunca executa codigo nem acessa servicos diretamente.

2. **Sistema de Orquestracao Deterministico (Go)**: O `OrchestratorService` como control plane central e gatekeeper. Ele valida e executa todas as operacoes, transiciona estados, gerencia sandboxes, controla o WebSocket e orquestra a comunicacao cross-module.

Os agentes executores (`code_worker`, `docs_writer`, `reviewer`, etc.) executam trabalho em sandboxes separadas e reportam eventos estruturados ao Orchestrator. Toda comunicacao cross-module passa obrigatoriamente pelo `OrchestratorService`.

Agentes podem solicitar informacoes de outros agentes, mas a comunicacao deve ser mediada pelo Orchestrator para manter auditoria, politicas e controle de contexto.

```mermaid
flowchart TD
    Human["Humano"] --> CLI["CLI local"]
    Human --> GitHub["GitHub: issues, PRs e revisao"]

    GitHub --> Intake
    CLI --> Intake
    Intake --> IOA["Agente Orquestrador Inteligente (LLM)"]

    IOA -->|"comandos estruturados"| OS["OrchestratorService (Go)"]
    OS -->|"Observation API"| IOA

    OS --> TG["Task Graph Planner"]
    TG --> WU["Work Units"]
    WU --> Run["Run / AgentSession"]

    OS --> ES["Event Store (fonte canônica)"]
    ES -.->|"projeções"| Run

    OS --> WSM["Workspace Manager (WSM)"]
    WSM --> Snap["Snapshot Engine"]
    WSM --> OEG["Operation Event Graph"]
    WSM --> MergeO["Merge Orchestrator"]

    WSM -->|"FileAPI / WorkspaceAPI"| Agent["Agent Worker (LLM/CLI)"]
    Agent -->|"checkpoints + ledger"| ES
    Agent <-->|"eventos / comandos"| OS

    OS --> Policy["Policy Engine"]
    Policy --> Approvals["Aprovações de ferramentas"]

    ES --> Observability["Logs, traces e auditoria"]
    ES -.-> Memory["Memória Recursiva"]
    Memory -.-> Prompts["Prompt Composer"]
    Prompts --> Agent

    MergeO --> GitHub
    OS --> CLI
    Chat["Conectores opcionais"] -.-> Intake
    OS -.-> Chat
```

## Principios

- Repositorio continua sendo a fonte de verdade.
- GitHub e CLI sao as interfaces operacionais iniciais.
- Chat e outras interfaces conversacionais sao conectores opcionais futuros, nao memoria definitiva.
- CLI e a primeira interface oficial do MVP; scripts sao bootstrap interno.
- Cada task deve ter workspace isolado (via WSM), branch, estado e trilha de auditoria.
- Cada task complexa deve ser decomposta em Task Graph aciclico.
- Prompts devem ser montados por fragmentos versionados e registrados em snapshot.
- Memoria recursiva deve ser camada derivada de eventos, checkpoints, ledger, artefatos e documentos versionados, nunca fonte canonica paralela.
- Toda acao relevante do agente deve gerar evento estruturado.
- Comunicacao entre agentes deve ser registrada e mediada.
- Permissoes de ferramentas devem seguir politica explicita.
- O sistema deve comecar pequeno, suportando 2 a 5 agentes paralelos.
- O desenho deve permitir evolucao para servidor sem reescrever o dominio.

## Documentos Relacionados

- [Stack inicial](core/stack.md)
- [Orquestracao de agentes](orchestration.md)
- [Agente Orquestrador Inteligente](agents/intelligent-orchestrator-agent.md)
- [Observation API](observability/orchestrator-observation-api.md)
- [Protocolo de Intervencao](protocols/orchestrator-intervention-protocol.md)
- [Coordenacao Multi-Agente](agents/multi-agent-coordination.md)
- [Modelo de dominio](core/domain-model.md)
- [Estrategia de interface](interface/interface-strategy.md)
- [Decomposicao de tasks](execution/task-decomposition.md)
- [Sistema de prompts](interface/prompt-system.md)
- [Sistema de memoria recursiva](observability/memory-system.md)
- [Protocolo de comunicacao](protocols/communication-protocol.md)
- [Estrutura inicial do repositorio](core/repo-structure.md)
- [JSON Schemas](../contracts/json-schemas.md)
- [Permissoes e ferramentas](project/permissions.md)
- [Sandbox e autonomia](agents/sandbox-and-autonomy.md)
- [Estrategia de testes](project/testing-strategy.md)
- [Falhas e rollback](execution/failures-and-rollback.md)
- [MVP local-first](project/mvp.md)
- [Proposta futura: Massive Agents System](agents/massive-agents-system.md)
- [Plano de implementacao](../implementation/roadmap.md)
- [ADR 0002: Orchestrator como control plane](../adr/0002-orchestrator-control-plane.md)
- [ADR 0003: Stack inicial](../adr/0003-initial-technology-stack.md)
- [ADR 0004: Sandbox e autonomia inicial](../adr/0004-sandbox-and-autonomy.md)
- [ADR 0005: Interface inicial do MVP](../adr/0005-mvp-interface-strategy.md)
- [ADR 0006: Decomposicao de tasks e intervencao em agentes](../adr/0006-task-graph-and-agent-intervention.md)
- [ADR 0007: Sistema de composicao de prompts](../adr/0007-prompt-composition-system.md)
- [ADR 0008: Ledger persistente de progresso](../adr/0008-agent-task-ledger.md)
- [ADR 0009: Normalizacao de historico e tracing](../adr/0009-trace-history-normalization.md)
- [ADR 0010: Operacao GitHub-first e chat opcional](../adr/0010-github-first-operations.md)
- [ADR 0011: Agent Checkpoints](../adr/0011-agent-checkpoints.md)
- [ADR 0012: Sistema de memoria recursiva](../adr/0012-recursive-memory-system.md)
- [ADR 0013: Escopo M0 de schemas e tipos de dominio](../adr/0013-m0-domain-contract-scope.md)
- [ADR 0014: Persistencia M0, CLI minima e testes](../adr/0014-m0-cli-persistence-and-integration-tests.md)
- [ADR 0015: TUI como interface local primaria](../adr/0015-tui-as-primary-local-interface.md)
- [ADR 0016: State Machine event-sourced](../adr/0016-event-sourced-state-machine.md)
- [ADR 0015: LLM-Optimized Module Architecture (superseded)](../adr/0015-vertical-module-architecture.md)
- [ADR 0016: Hybrid Intelligent Orchestrator Architecture](../adr/0016-hybrid-intelligent-orchestrator.md)
- [ADR 0019: Arquitetura Modular Simplificada (vigente)](../adr/0019-simplified-modular-architecture.md)

## Referencias Tecnicas

- OpenAI Agents SDK: https://developers.openai.com/api/docs/guides/agents
- OpenAI Agent orchestration: https://openai.github.io/openai-agents-python/multi_agent/
- Model Context Protocol: https://modelcontextprotocol.io/docs/learn/architecture
- Agent2Agent Protocol: https://github.com/a2aproject/A2A
- Temporal: https://docs.temporal.io/
- NATS JetStream: https://docs.nats.io/nats-concepts/jetstream
- Git worktree: https://git-scm.com/docs/git-worktree.html
- Docker security: https://docs.docker.com/engine/security/
- gVisor: https://gvisor.dev/docs/
- OpenTelemetry: https://opentelemetry.io/docs/
