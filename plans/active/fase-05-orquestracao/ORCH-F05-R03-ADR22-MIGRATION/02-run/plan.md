# ORCH-F05-R03-A02 — Migração de Tipos: Run Module

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.
> **Branch esperada:** `adr22-a02-run-types`
> **Worktree esperada:** `../orchestraos-a02-run`
> Se não estiver isolado, execute:
> ```bash
> cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A02 run
> ```

---

## Contexto

O módulo `run` gerencia execuções de WorkUnits por agent sessions. Atualmente usa **aliases** para tipos do `domain`:

```go
package run

type Status = domain.RunStatus
type Result = domain.RunResult
```

Isso é uma migração incompleta. O módulo ainda importa `internal/domain` e todos os seus arquivos (repository, service, fetch, events) usam `domain.Run`, `domain.RunStatus`, `domain.RunResult`.

**Pré-requisito:** O módulo `task` (A01) deve estar concluído (🟢), pois `run` consome `task.Task` via `TaskReader` interface.

---

## Documentação Obrigatória (ler ANTES)

1. `/home/levybonito/Documentos/OrchestraOS/internal/modules/run/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/internal/modules/run/CONTRACTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0022-llm-optimized-module-architecture.md`
4. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`
5. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-ADR22-MIGRATION/README.md`

---

## Inventário de Imports a Migrar

| Arquivo | Import de domain | O que usa |
|---------|-----------------|-----------|
| `run/models.go` | `domain.RunStatus`, `domain.RunResult` | aliases `Status`, `Result` |
| `run/repository.go` | `domain.Run` | `Create(run *domain.Run)`, `GetByID`, `scanRun` |
| `run/service.go` | `domain.Run`, `domain.RunStatus`, `domain.RunResult` | struct `RunService`, `CreateRunInput`, transitions |
| `run/service_retry.go` | `domain.RunStatus` | retry policies |
| `run/fetch.go` | `domain.Run` | `RequireByID` retorna `*domain.Run` |
| `run/events.go` | `domain.RunStatus`, `domain.RunResult` | `EventTypeForStatus`, `ResultForStatus` |

---

## Estratégia de Migração

### Passo 1: Criar tipos locais em models.go
Substituir aliases por structs e enums próprios:

```go
package run

type Status string
const (
    StatusCreated         Status = "created"
    StatusRunning         Status = "running"
    StatusWaitingApproval Status = "waiting_approval"
    StatusPaused          Status = "paused"
    StatusValidating      Status = "validating"
    StatusCompleted       Status = "completed"
    StatusFailed          Status = "failed"
    StatusCancelled       Status = "cancelled"
)

type Result string
const (
    ResultSucceeded Result = "succeeded"
    ResultFailed    Result = "failed"
    ResultCancelled Result = "cancelled"
)

type Run struct {
    ID         string     `json:"id"`
    TaskID     string     `json:"task_id"`
    WorkUnitID string     `json:"work_unit_id"`
    Status     Status     `json:"status"`
    Attempt    int        `json:"attempt"`
    StartedAt  time.Time  `json:"started_at"`
    FinishedAt *time.Time `json:"finished_at,omitempty"`
    Result     *Result    `json:"result,omitempty"`
    FailureReason *string `json:"failure_reason,omitempty"`
}
```

### Passo 2: Atualizar arquivos internos
- `repository.go`: `*domain.Run` → `*Run`, `[]domain.Run` → `[]Run`
- `service.go`: todas as referências a `domain.Run`, `domain.RunStatus`, `domain.RunResult`
- `service_retry.go`: `domain.RunStatus` → `Status`
- `fetch.go`: `*domain.Run` → `*Run`
- `events.go`: `domain.RunStatus` → `Status`, `domain.RunResult` → `Result`

### Passo 3: Atualizar interfaces dependentes
- `internal/modules/run/service.go` — `TaskReader` interface usa `*domain.Task`. Como task já foi migrado (A01), atualizar para `*task.Task` ou manter adapter temporário.
- `internal/modules/workunit/service.go` — `TaskReader` interface (se referencia run, verificar)

### Passo 4: Criar Adapters Temporários nos Consumidores

| Consumidor | Motivo | Adapter |
|-----------|--------|---------|
| `internal/modules/orchestrator/models.go` | `RunLifecycleManager` interface retorna `*domain.Run` | `// TODO[ADR-0022]: migrar para *run.Run` |
| `internal/modules/orchestrator/service.go` | `executeWorkUnit` usa `domain.RunStatusCompleted` | Adapter inline ou anotação |
| `internal/modules/trigger/models.go` | `RunReader` interface retorna `*domain.Run` | `// TODO[ADR-0022]: migrar para *run.Run` |
| `internal/core/coordination/*` | helpers usam `domain.Run` | Criar `runToDomain()` se necessário |
| `internal/bootstrap/services.go` | run adapter | Atualizar ou manter com TODO |

### Passo 5: Build + Test + Commit
```bash
go build ./...
go test ./...
./scripts/safe-commit.sh "ADR-0022: migrate Run types to modules/run"
```

---

## Critérios de Aceitação

- [ ] `internal/modules/run/models.go` define `Run`, `Status`, `Result` localmente (sem import de domain)
- [ ] Todos os arquivos em `internal/modules/run/` usam tipos locais
- [ ] `TaskReader` interface usa `*task.Task` (não `*domain.Task`), pois task já foi migrado
- [ ] Todos os consumidores que precisam de `domain.Run` têm adapter com `// TODO[ADR-0022]: ...`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh` passa
