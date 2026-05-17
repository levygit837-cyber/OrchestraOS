# ORCH-F05-R03-A03 — Migração de Tipos: WorkUnit Module

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.
> **Branch esperada:** `adr22-a03-workunit-types`
> **Worktree esperada:** `../orchestraos-a03-workunit`
> Se não estiver isolado, execute:
> ```bash
> ```

---

## Contexto

O módulo `workunit` gerencia unidades de trabalho dentro de um TaskGraph. Atualmente usa **alias** para o status:

```go
package workunit

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.WorkUnitStatus
```

Todos os arquivos do módulo usam `domain.WorkUnit` e `domain.WorkUnitStatus`.

**Pré-requisitos:**
- A01 (task) 🟢 — `TaskReader` interface já usa `*task.Task`
- A02 (run) não é estritamente necessário para iniciar, mas workunit lê Run via callback. Pode usar adapter temporário para Run se A02 não estiver pronto.

---

## Documentação Obrigatória (ler ANTES)

1. `/home/levybonito/Documentos/OrchestraOS/internal/modules/workunit/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/internal/modules/workunit/CONTRACTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0022-llm-optimized-module-architecture.md`
4. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`
5. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-ADR22-MIGRATION/README.md`

---

## Inventário de Imports a Migrar

| Arquivo | Import de domain | O que usa |
|---------|-----------------|-----------|
| `workunit/models.go` | `domain.WorkUnitStatus` | alias `Status` |
| `workunit/repository.go` | `domain.WorkUnit` | CRUD |
| `workunit/service.go` | `domain.WorkUnit`, `domain.WorkUnitStatus`, `domain.Task`, `domain.TaskGraph` | `TaskReader`, `TaskGraphManager` interfaces |
| `workunit/service_create.go` | `domain.WorkUnit` | `CreateMany` |
| `workunit/fetch.go` | `domain.WorkUnit` | `RequireByID` |
| `workunit/events.go` | `domain.WorkUnitStatus` | `EventTypeForStatus` |
| `workunit/validation.go` | `domain.WorkUnitStatus` | validações |

---

## Estratégia de Migração

### Passo 1: Criar tipos locais em models.go

```go
package workunit

type Status string
const (
    StatusCreated         Status = "created"
    StatusPlanned         Status = "planned"
    StatusScheduled       Status = "scheduled"
    StatusBlocked         Status = "blocked"
    StatusRunning         Status = "running"
    StatusWaitingApproval Status = "waiting_approval"
    StatusPaused          Status = "paused"
    StatusValidating      Status = "validating"
    StatusCompleted       Status = "completed"
    StatusFailed          Status = "failed"
    StatusCancelled       Status = "cancelled"
)

type WorkUnit struct {
    ID                   string   `json:"id"`
    TaskID               string   `json:"task_id"`
    TaskGraphID          string   `json:"task_graph_id"`
    Title                string   `json:"title"`
    Objective            string   `json:"objective"`
    AssignedAgentProfile string   `json:"assigned_agent_profile"`
    Status               Status   `json:"status"`
    OwnedPaths           []string `json:"owned_paths"`
    ReadPaths            []string `json:"read_paths"`
    AcceptanceCriteria   []string `json:"acceptance_criteria"`
    ValidationPlan       []string `json:"validation_plan"`
    DependsOn            []string `json:"depends_on"`
}
```

### Passo 2: Atualizar arquivos internos
- `repository.go`: `*domain.WorkUnit` → `*WorkUnit`
- `service.go`: `domain.WorkUnitStatus` → `Status`, `domain.WorkUnit` → `WorkUnit`
- `service_create.go`: `domain.WorkUnit` → `WorkUnit`
- `fetch.go`: `*domain.WorkUnit` → `*WorkUnit`
- `events.go`: `domain.WorkUnitStatus` → `Status`
- `validation.go`: validar `Status` local

### Passo 3: Atualizar interfaces cruzadas
- `TaskReader` interface: se ainda usa `*domain.Task`, atualizar para `*task.Task` (task já migrado em A01)
- `TaskGraphManager` interface: se usa `*domain.TaskGraph`, manter com adapter temporário (taskgraph ainda não migrado)

### Passo 4: Criar Adapters Temporários nos Consumidores

| Consumidor | Motivo | Adapter |
|-----------|--------|---------|
| `internal/modules/orchestrator/models.go` | `WorkUnitLister` interface (`[]domain.WorkUnit`) | `// TODO[ADR-0022]: migrar para []workunit.WorkUnit` |
| `internal/modules/orchestrator/service.go` | `executeWorkUnit`, `topologicalSort` usam `domain.WorkUnit` | Criar `workunitToDomain()` |
| `internal/modules/taskgraph/service.go` | `WorkUnitCreator`, `WorkUnitLister` interfaces | `// TODO[ADR-0022]: migrar para workunit.WorkUnit` |
| `internal/modules/prompt/service.go` | `PrepareAndPersistInput.WorkUnit *domain.WorkUnit` | `// TODO[ADR-0022]: migrar para *workunit.WorkUnit` |
| `internal/modules/trigger/service.go` | `WorkUnitReader` interface | `// TODO[ADR-0022]: migrar para *workunit.WorkUnit` |
| `internal/core/coordination/*` | cascade helpers | Criar adapters com TODO |

### Passo 5: Build + Test + Commit
```bash
go build ./...
go test ./...
./scripts/safe-commit.sh "ADR-0022: migrate WorkUnit types to modules/workunit"
```

---

## Critérios de Aceitação

- [ ] `internal/modules/workunit/models.go` define `WorkUnit`, `Status` localmente
- [ ] Todos os arquivos em `internal/modules/workunit/` usam tipos locais
- [ ] `TaskReader` interface usa `*task.Task` (task já migrado)
- [ ] `TaskGraphManager` interface mantém adapter temporário para `domain.TaskGraph` (ainda não migrado)
- [ ] Todos os consumidores têm adapters com `// TODO[ADR-0022]: ...`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh` passa
