# Stack Inicial

## Recomendacao

| Camada | Escolha inicial | Motivo |
| --- | --- | --- |
| Orchestrator | Go | Concorrencia simples, binario unico, bom controle de processos, rede e WebSocket. |
| Persistencia | Postgres | Estado transacional, auditoria, consultas e base solida para event store inicial. |
| Canal agente-orquestrador | WebSocket | Canal bidirecional vivo para eventos, comandos, checkpoints e interrupcoes. |
| Fila/eventos duraveis | Postgres outbox no MVP; NATS JetStream depois | Reduz complexidade inicial e permite evoluir para filas duraveis quando houver escala. |
| Runtime de agentes | Codex/CLI | Alinha com o uso inicial do projeto e permite execucao tecnica em repositorios reais. |
| Isolamento | Git worktree + Docker | Worktree separa alteracoes por task; Docker limita ambiente e dependencias. |
| Sandbox reforcado futuro | gVisor ou Firecracker | Necessario quando o sistema precisar tratar codigo potencialmente malicioso com mais rigor. |
| Superficie externa inicial | GitHub | Issues, PRs, reviews e checks ja resolvem backlog, revisao e historico. |
| Interface inicial | Scripts de bootstrap + CLI fina | Reduz custo de UI e cria contrato operacional testavel. |
| Conectores de chat futuros | Slack, Discord ou equivalente | Somente quando houver necessidade real de captura, avisos ou rotinas por chat. |
| Painel futuro | TypeScript + React | Bom para uma UI operacional rica quando CLI/GitHub nao forem suficientes. |
| Contratos | JSON Schema e OpenAPI | Mensagens, eventos, tool calls e APIs precisam ser validados nas bordas. |
| Observabilidade | Logs estruturados; OpenTelemetry depois | Comecar simples, mas manter caminho para traces, metricas e logs correlacionados. |

## Go Como Linguagem Principal

Go e a melhor escolha para o nucleo do Orchestrator porque o sistema precisa administrar processos, conexoes WebSocket, timeouts, cancelamentos, concorrencia, health checks, filas e containers.

TypeScript continua recomendado para painel web e integracoes de frontend. Python pode ser usado para adapters ou agentes especificos quando bibliotecas de IA forem mais maduras nesse ecossistema.

## O Que Nao Entra no MVP

- Kubernetes.
- Temporal como dependencia obrigatoria.
- NATS obrigatorio desde o primeiro prototipo.
- Painel web completo.
- Aplicativo desktop.
- Comunicacao peer-to-peer livre entre agentes.
- Suporte completo a runtimes heterogeneos de agentes.
- A2A como protocolo interno principal.

## Evolucao Prevista

1. MVP local com Orchestrator, CLI, Postgres, WebSocket, Docker, worktrees e GitHub.
2. Servidor unico com os mesmos componentes e armazenamento persistente.
3. NATS JetStream quando houver necessidade real de filas duraveis, replay e consumo concorrente.
4. gVisor ou Firecracker quando o risco de codigo malicioso justificar sandbox mais forte.
5. Temporal quando workflows longos, retries, pausas humanas e retomada apos crash virarem dor real.
