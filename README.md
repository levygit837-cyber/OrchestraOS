# Sistema de Orquestração de Agentes

Projeto inicial para construir um sistema em que ideias, decisões, execução de código, automações e operação sejam coordenadas por agentes de IA com supervisão humana progressivamente menor.

## Estado

- Nome final: em aberto
- Fase atual: fundação e integração E2E
- Fonte de verdade: este repositório
- Comunicação operacional sugerida: CLI + GitHub
- Execução técnica sugerida: Codex + GitHub + worktrees + automações futuras
- Arquitetura: Arquitetura Modular Simplificada (ADR-0030) com Orchestrator central e agentes isolados

## Documentos Principais

- [AGENTS.md](AGENTS.md): regras que agentes devem seguir ao trabalhar no repositório.
- [docs/canvas/project-canvas.md](docs/canvas/project-canvas.md): canvas textual do produto, legível por humanos e por IA.
- [docs/canvas/system-map.mmd](docs/canvas/system-map.mmd): mapa Mermaid da arquitetura atual.
- [docs/architecture/README.md](docs/architecture/README.md): visão geral da arquitetura do OrchestraOS.
- [docs/architecture/core/stack.md](docs/architecture/core/stack.md): stack técnica recomendada e evolução.
- [docs/architecture/orchestration.md](docs/architecture/orchestration.md): modelo de orquestração de agentes.
- [docs/architecture/core/domain-model.md](docs/architecture/core/domain-model.md): modelos de domínio do sistema.
- [docs/architecture/interface/interface-strategy.md](docs/architecture/interface/interface-strategy.md): estratégia CLI, GitHub, Desktop, Web e conectores opcionais.
- [docs/architecture/execution/task-decomposition.md](docs/architecture/execution/task-decomposition.md): decomposição de tasks em DAG.
- [docs/architecture/interface/prompt-system.md](docs/architecture/interface/prompt-system.md): composição de SystemPrompts e TaskPrompts.
- [docs/architecture/observability/memory-system.md](docs/architecture/observability/memory-system.md): desenho da memória recursiva derivada de eventos, checkpoints e documentação.
- [docs/architecture/protocols/communication-protocol.md](docs/architecture/protocols/communication-protocol.md): contrato inicial de eventos e comandos.
- [docs/architecture/core/repo-structure.md](docs/architecture/core/repo-structure.md): estrutura inicial de codigo, contratos e testes.
- [docs/contracts/json-schemas.md](docs/contracts/json-schemas.md): indice dos schemas executaveis de dominio, eventos e comandos.
- [docs/architecture/project/permissions.md](docs/architecture/project/permissions.md): matriz de ferramentas, riscos e aprovações.
- [docs/architecture/agents/sandbox-and-autonomy.md](docs/architecture/agents/sandbox-and-autonomy.md): política inicial de sandbox e autonomia.
- [docs/architecture/project/testing-strategy.md](docs/architecture/project/testing-strategy.md): estratégia de testes por domínio.
- [docs/architecture/execution/failures-and-rollback.md](docs/architecture/execution/failures-and-rollback.md): falhas, rollback e recuperação.
- [docs/architecture/project/mvp.md](docs/architecture/project/mvp.md): escopo do MVP local-first.
- [docs/implementation/roadmap.md](docs/implementation/roadmap.md): plano técnico de implementação.
- [docs/management/operating-model.md](docs/management/operating-model.md): modelo de gestão do projeto.
- [docs/slack/slack-setup.md](docs/slack/slack-setup.md): configuração opcional de Slack para integração futura.
- [docs/naming.md](docs/naming.md): opções de nomes para produto/empresa.

## Decisões Arquiteturais Atuais

- Orchestrator como control plane central.
- Agentes como workers isolados por task.
- Comunicação agente-orquestrador por WebSocket com eventos persistidos.
- Comunicação entre agentes mediada pelo Orchestrator.
- Task Graph acíclico para decomposição de trabalho; loops ficam dentro das runs dos agentes.
- Sistema de composição de prompts por fragmentos versionados.
- Ledger persistente de progresso por work unit.
- Histórico operacional normalizado no Event Store.
- Memória recursiva futura como camada derivada e auditável, não como fonte de verdade.
- Interface inicial: scripts de bootstrap, CLI fina e GitHub.
- Operação inicial GitHub-first: issues, branches, worktrees, pull requests, reviews e checks.
- Chat, incluindo Slack, fica como conector opcional futuro.
- Stack inicial: Go, Postgres, Codex/CLI, Git worktree, Docker e GitHub.
- Autonomia inicial aprovada: Nível 2.

## Estrutura do Repositório

- `cmd/orchestraos/`: entrada da CLI local.
- `internal/domain/`: todos os entity types compartilhados (ADR-0030 Pilar 1).
- `internal/modules/`: módulos autônomos (agent, agentsession, orchestrator, prompt, review, run, task, taskgraph, trigger, workunit).
- `internal/core/`: infraestrutura compartilhada (apperrors, db, eventstore, statemachine, transition, validation).
- `internal/bootstrap/`: injeção de dependências e wiring de serviços.
- `contracts/schemas/`: JSON Schemas versionados.
- `tests/architecture/`: testes automáticos de regras de arquitetura.

Para mais detalhes, consulte [docs/architecture/core/repo-structure.md](docs/architecture/core/repo-structure.md).

## Regra de Trabalho

Antes de implementar qualquer funcionalidade, transformar a ideia em um item pequeno com objetivo, escopo, critérios de aceite e teste esperado. O projeto deve crescer por decisões registradas, não por improviso acumulado.
