# Índice de Módulos

| Módulo | Path | Responsabilidade | Depende de |
|--------|------|------------------|------------|
| agent | `internal/modules/agent` | Runtimes de agente (Codex, Gemini, Fake) | core/apperrors |
| event | `internal/core/event` | Event sourcing, EventEnvelope | core/apperrors, core/eventstore, core/db |
| prompt | `internal/modules/prompt` | Prompt fragments, snapshots, toolsets | core/apperrors, core/db |
| task | `internal/modules/task` | Ciclo de vida de tarefas | core/apperrors, core/db |
| taskgraph | `internal/modules/taskgraph` | Decomposição e planejamento | core/apperrors, core/db |
| workunit | `internal/modules/workunit` | Work units, dependências, paths | core/apperrors, core/db |
| run | `internal/modules/run` | Execuções de agente, retry, projeção | core/apperrors, core/db |
| agentsession | `internal/modules/agentsession` | Sessões de agente, checkpoints | core/apperrors, core/db |

## Infraestrutura (core/)

| Pacote | Path | Responsabilidade |
|--------|------|------------------|
| apperrors | `internal/core/apperrors` | Erros padronizados com código e operação |
| db | `internal/core/db` | Conn pool, DBTX interface, tx helpers (BeginTx, CommitTx, RollbackTx, AdvisoryLock) |
| eventstore | `internal/core/eventstore` | Store de eventos com validação schema |
| orchestration | `internal/core/orchestration` | Cross-domain: TransitionInput, OperationResult, AppendTransition, GetTask/GetRun/etc. |
| serialization | `internal/core/serialization` | MarshalPayload genérico |
| statemachine | `internal/core/statemachine` | Regras de transição de estado, replay |
| validation | `internal/core/validation` | Validadores genéricos (UUID, texto, priority, risk, runtime) |

## Regras de Navegação para LLMs
- Todo módulo tem: `models.go`, `repository.go`, `queries.go`
- `common.go` foi eliminado. Services em `internal/services/` importam `core/*` e `modules/*` diretamente
- Regras cross-domain ficam em `internal/core/orchestration/`
- Helpers transacionais: `internal/core/db/txkit.go` + `BeginTx/CommitTx/RollbackTx/EnsureRowsAffected/AcquireAdvisoryTxLock`
- Validadores genéricos: `internal/core/validation/`
- Erros padronizados: `internal/core/apperrors/`
