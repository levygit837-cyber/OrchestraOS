# ORCH-F05-R03-A04 — Migração de Tipos: TaskGraph Module

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.
> **Branch esperada:** `adr22-a04-taskgraph-types`
> **Worktree esperada:** `../orchestraos-a04-taskgraph`
> Se não estiver isolado, execute:
> ```bash
> cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A04 taskgraph
> ```

---

## Contexto

O módulo `taskgraph` decompõe Tasks em grafos direcionados acíclicos (DAGs) de WorkUnits. Atualmente usa **alias** para o status:

```go
package taskgraph

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.TaskGraphStatus
```

O módulo é um dos mais complexos porque consome **três entidades externas**: `Task` (do módulo task), `WorkUnit` (do módulo workunit) e `TaskGraph` (de si mesmo, mas ainda via domain).

**Pré-requisitos:**
- A01 (task) 🟢 — Task já migrado, usar `*task.Task`
- A03 (workunit) 🟢 — WorkUnit já migrado, usar `workunit.WorkUnit` / `*workunit.WorkUnit`

---

## Documentação Obrigatória (ler ANTES)

1. `/home/levybonito/Documentos/OrchestraOS/internal/modules/taskgraph/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/internal/modules/taskgraph/CONTRACTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0022-llm-optimized-module-architecture.md`
4. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0018-local-heuristic-task-graph-planner.md`
5. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`
6. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-ADR22-MIGRATION/README.md`

---

## Inventário de Imports a Migrar

| Arquivo | Import de domain | O que usa |
|---------|-----------------|-----------|
| `taskgraph/models.go` | `domain.TaskGraphStatus` | alias `Status` |
| `taskgraph/repository.go` | `domain.TaskGraph` | CRUD |
| `taskgraph/service.go` | `domain.TaskGraph`, `domain.Task`, `domain.WorkUnit`, `domain.TaskGraphCreatedPayload` | decompose, duplicate detection |
| `taskgraph/planner.go` | `domain.Task` | interface `Planner` |
| `taskgraph/planner_prompt.go` | `domain.Task` | `PlannerPrompt` function |
| `taskgraph/heuristic.go` | `domain.Task`, `domain.WorkUnit` | `buildLocalHeuristicGraphPlan` |
| `taskgraph/gemini_planner.go` | `domain.Task` | `Plan(ctx, task *domain.Task)` |

---

## Estratégia de Migração

### Passo 1: Criar tipos locais em models.go

```go
package taskgraph

type Status string
const (
    StatusActive     Status = "active"
    StatusSuperseded Status = "superseded"
)

type TaskGraph struct {
    ID              string    `json:"id"`
    TaskID          string    `json:"task_id"`
    Version         int       `json:"version"`
    Status          Status    `json:"status"`
    PlannerStrategy string    `json:"planner_strategy"`
    Rationale       string    `json:"rationale"`
    CreatedBy       string    `json:"created_by"`
    NodeCount       int       `json:"node_count"`
    EdgeCount       int       `json:"edge_count"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

**Nota:** `TaskGraphCreatedPayload` também deve ser definido localmente (hoje é `domain.TaskGraphCreatedPayload`).

### Passo 2: Atualizar arquivos internos
- `repository.go`: `*domain.TaskGraph` → `*TaskGraph`
- `service.go`: `domain.TaskGraph` → `TaskGraph`, `domain.Task` → `task.Task`, `domain.WorkUnit` → `workunit.WorkUnit`
- `planner.go`: `domain.Task` → `task.Task`
- `planner_prompt.go`: `domain.Task` → `task.Task`
- `heuristic.go`: `domain.Task` → `task.Task`, `domain.WorkUnit` → `workunit.WorkUnit`
- `gemini_planner.go`: `domain.Task` → `task.Task`
- `planner_validator.go`: verificar se usa domain types

### Passo 3: Atualizar interfaces de planners
A interface `Planner` atualmente recebe `*domain.Task`. Como Task já foi migrado (A01):

```go
// TODO[ADR-0022]: interface já atualizada para task.Task
type Planner interface {
    Plan(ctx context.Context, task *task.Task) (*GraphPlan, error)
}
```

### Passo 4: Criar Adapters Temporários nos Consumidores

| Consumidor | Motivo | Adapter |
|-----------|--------|---------|
| `internal/modules/orchestrator/models.go` | `TaskGraphManager` interface | `// TODO[ADR-0022]: migrar para *taskgraph.TaskGraph` |
| `internal/modules/workunit/service.go` | `TaskGraphManager` injetado | `// TODO[ADR-0022]: migrar para taskgraph.TaskGraph` |
| `internal/bootstrap/services.go` | taskgraph adapter | Atualizar com TODO |

### Passo 5: Build + Test + Commit
```bash
go build ./...
go test ./...
./scripts/safe-commit.sh "ADR-0022: migrate TaskGraph types to modules/taskgraph"
```

---

## Critérios de Aceitação

- [ ] `internal/modules/taskgraph/models.go` define `TaskGraph`, `Status` localmente
- [ ] `TaskGraphCreatedPayload` definido localmente (não em domain)
- [ ] Todos os arquivos em `internal/modules/taskgraph/` usam tipos locais
- [ ] `Planner` interface usa `*task.Task` (task já migrado)
- [ ] `domain.WorkUnit` substituído por `workunit.WorkUnit` (workunit já migrado)
- [ ] Todos os consumidores têm adapters com `// TODO[ADR-0022]: ...`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh` passa
