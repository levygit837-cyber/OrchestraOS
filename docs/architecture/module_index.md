# Índice de Módulos

| Módulo | Path | Responsabilidade | Depende de |
|--------|------|------------------|------------|
| agent | `internal/modules/agent` | Agent entities, Runtimes (Fake, Gemini, Codex), GeminiPlanner | core/apperrors, core/transition |
| agentsession | `internal/modules/agentsession` | Sessões de agente, checkpoints, heartbeat, timeout | core/apperrors, core/db, core/transition, core/statemachine |
| orchestrator | `internal/modules/orchestrator` | OrchestratorService - coordena fluxo end-to-end de tasks | core/apperrors, core/db, core/transition |
| prompt | `internal/modules/prompt` | Prompt fragments, snapshots, toolsets | core/apperrors, core/db |
| review | `internal/modules/review` | Gates de revisão e validação | core/apperrors, core/db, core/transition |
| run | `internal/modules/run` | Execuções de agente, retry, projeção | core/apperrors, core/db, core/transition |
| task | `internal/modules/task` | Ciclo de vida de tarefas | core/apperrors, core/db, core/transition |
| taskgraph | `internal/modules/taskgraph` | Decomposição e planejamento (local + LLM) | core/apperrors, core/db |
| trigger | `internal/modules/trigger` | Detecção de anomalias (stalls, loops) | core/apperrors, core/db |
| workunit | `internal/modules/workunit` | Work units, dependências, paths | core/apperrors, core/db, core/transition |

## Infraestrutura (core/)

| Pacote | Path | Responsabilidade |
|--------|------|------------------|
| apperrors | `internal/core/apperrors` | Erros padronizados com código e operação |
| db | `internal/core/db` | Conn pool, DBTX interface, tx helpers (BeginTx, CommitTx, RollbackTx, AdvisoryLock) |
| event | `internal/core/event` | EventService wrapper do Event Store |
| eventstore | `internal/core/eventstore` | Store de eventos com validação schema, append e replay |
| transition | `internal/core/transition` | Cross-domain: TransitionInput, OperationResult, AppendTransition, AppendServiceEvent |
| serialization | `internal/core/serialization` | MarshalPayload genérico |
| statemachine | `internal/core/statemachine` | Regras de transição de estado, replay |
| transition | `internal/core/transition` | Payload builders para transições |
| validation | `internal/core/validation` | Validadores genéricos (UUID, texto, priority, risk, runtime) |

## Regras de Navegação para LLMs

- Todo módulo vertical tem: `README.md`, `CONTRACTS.md`, `doc.go`, `models.go`, `service.go`, `repository.go`, `queries.go`, `validation.go`
- **Regra de Ouro (ADR 0022):** Módulos verticais NUNCA importam outros módulos diretamente
- Comunicação cross-module ocorre via `internal/modules/orchestrator/` (camada de orquestração canônica) ou interfaces DI com adapters em `internal/bootstrap/services.go`
- Helpers transacionais: `internal/core/db/` + `BeginTx/CommitTx/RollbackTx/EnsureRowsAffected/AcquireAdvisoryTxLock`
- Validadores genéricos: `internal/core/validation/`
- Erros padronizados: `internal/core/apperrors/`
- Tipos compartilhados entre módulos: `internal/domain/`
