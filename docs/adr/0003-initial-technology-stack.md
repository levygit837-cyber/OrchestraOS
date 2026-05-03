# ADR 0003: Stack Inicial

## Contexto

O projeto deve iniciar localmente, mas com desenho compativel com servidor. O primeiro runtime de agentes sera Codex/CLI em sandbox. O paralelismo inicial esperado e de 2 a 5 agentes. A operacao inicial sera GitHub-first, com CLI local e worktrees.

O sistema precisa coordenar processos, containers, worktrees, WebSockets, politicas, auditoria e integracoes externas sem introduzir complexidade excessiva no MVP.

## Decisão

A stack inicial sera:

- **Go** para o Orchestrator.
- **Postgres** para estado, auditoria e event store inicial.
- **WebSocket** para comunicacao viva entre Orchestrator e agentes.
- **Postgres outbox** para eventos/filas no MVP.
- **NATS JetStream** como evolucao quando houver necessidade real de filas duraveis e consumo concorrente.
- **Codex/CLI** como primeiro runtime de agentes.
- **Git worktree + Docker** para isolamento inicial por task.
- **gVisor ou Firecracker** como evolucao para sandbox mais forte.
- **GitHub + CLI** como interface operacional inicial.
- **Conectores de chat** como evolucao opcional.
- **TypeScript + React** apenas quando o painel web virar prioridade.
- **JSON Schema/OpenAPI** para contratos.
- **Logs estruturados**, com evolucao para OpenTelemetry.

## Consequências

- Go reduz complexidade no nucleo de concorrencia, rede e processos.
- Postgres permite comecar com menos infraestrutura.
- WebSocket resolve controle em tempo real, mas exige persistencia separada para durabilidade.
- Docker e worktree sao suficientes para validar o fluxo, mas nao encerram o tema de seguranca.
- NATS, Temporal, gVisor, Firecracker e painel web ficam como evolucao, nao como requisito inicial.

## Alternativas Consideradas

- **TypeScript full-stack**: bom para velocidade e painel web, mas menos direto para controle robusto de processos e runtime.
- **Python como nucleo**: forte em bibliotecas de IA, mas menos adequado como base principal de supervisao de processos e concorrencia de sistema.
- **Temporal desde o inicio**: forte para workflows duraveis, mas pesado para validar o primeiro ciclo.
- **Kubernetes desde o inicio**: poderoso para escala, mas desnecessario para 2 a 5 agentes locais.
- **A2A como protocolo interno principal**: util para interoperabilidade futura, mas excessivo para o controle interno inicial.
