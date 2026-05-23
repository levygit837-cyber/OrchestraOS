# Índice de Módulos

| Módulo | Path | Responsabilidade | Depende de |
|--------|------|------------------|------------|
| agent | `internal/modules/agent` | Agent entities, Runtimes (Fake, Gemini, Codex), GeminiPlanner | core/apperrors, domain |
| agentsession | `internal/modules/agentsession` | Sessões de agente, checkpoints, heartbeat, timeout | core/apperrors, core/db, core/statemachine, core/transition, domain |
| orchestrator | `internal/modules/orchestrator` | OrchestratorService - coordena fluxo end-to-end de tasks | core/apperrors, core/db, core/statemachine, core/transition, domain, **múltiplos módulos** |
| prompt | `internal/modules/prompt` | Prompt fragments, snapshots, toolsets | core/apperrors, core/db, domain |
| review | `internal/modules/review` | Gates de revisão e validação | core/apperrors, core/db, core/statemachine, core/transition, domain |
| run | `internal/modules/run` | Execuções de agente, retry, projeção | core/apperrors, core/db, core/statemachine, core/transition, domain |
| task | `internal/modules/task` | Ciclo de vida de tarefas | core/apperrors, core/db, core/statemachine, core/transition, domain |
| taskgraph | `internal/modules/taskgraph` | Decomposição e planejamento (local + LLM) | core/apperrors, core/db, domain |
| trigger | `internal/modules/trigger` | Detecção de anomalias (stalls, loops) | core/apperrors, core/db, domain |
| workunit | `internal/modules/workunit` | Work units, dependências, paths | core/apperrors, core/db, core/statemachine, core/transition, domain |

## Infraestrutura (core/)

| Pacote | Path | Responsabilidade |
|--------|------|------------------|
| apperrors | `internal/core/apperrors` | Erros padronizados com código e operação |
| db | `internal/core/db` | Conn pool, DBTX interface, tx helpers (BeginTx, CommitTx, RollbackTx, AdvisoryLock) |
| event | `internal/core/event` | EventService wrapper do Event Store |
| eventstore | `internal/core/eventstore` | Store de eventos com validação schema, append e replay |
| serialization | `internal/core/serialization` | MarshalPayload genérico |
| statemachine | `internal/core/statemachine` | Regras de transição de estado, replay |
| transition | `internal/core/transition` | Payload builders e helpers para transições cross-module |
| validation | `internal/core/validation` | Validadores genéricos (UUID, texto, priority, risk, runtime) |

## Regras de Navegação para LLMs

- Todo módulo deve ter no mínimo: `doc.go`, `README.md`, `models.go`, `repository.go`, `service.go`
- **Regra de Isolamento (ADR-0019):** Módulos NUNCA importam outros módulos diretamente. Apenas `orchestrator/` e `bootstrap/` importam múltiplos módulos.
- Dependências cross-module são resolvidas via interfaces DI com adapters em `internal/bootstrap/services.go`
- Helpers transacionais: `internal/core/db/` + `BeginTx/CommitTx/RollbackTx/EnsureRowsAffected/AcquireAdvisoryTxLock`
- Validadores genéricos: `internal/core/validation/`
- Erros padronizados: `internal/core/apperrors/`
- Todos os entity types compartilhados: `internal/domain/`
