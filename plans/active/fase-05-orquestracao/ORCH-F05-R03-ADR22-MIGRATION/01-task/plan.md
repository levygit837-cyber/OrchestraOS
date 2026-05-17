# ORCH-F05-R03-A01 — Migração de Tipos: Task Module

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.
> **Branch esperada:** `adr22-a01-task-types`
> **Worktree esperada:** `../orchestraos-a01-task`
> Se não estiver isolado, execute:
> ```bash
> cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A01 task
> ```

---

## Contexto

O módulo `task` já foi parcialmente migrado em sessões anteriores:
- ✅ `Task`, `Status`, `Priority`, `RiskLevel` já estão definidos localmente em `models.go`
- ✅ `repository.go`, `service.go`, `fetch.go`, `events.go` já usam tipos locais
- ❌ `models.go` ainda mantém funções `ToDomain()` e `FromDomain()` que importam `internal/domain`
- ❌ Consumidores cruzados ainda usam `domain.Task`, exigindo adapters

Este plano **finaliza** a migração do módulo `task`, removendo a última ponte com `internal/domain` e criando adapters temporários nos consumidores.

---

## Documentação Obrigatória (ler ANTES)

1. `/home/levybonito/Documentos/OrchestraOS/internal/modules/task/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/internal/modules/task/CONTRACTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0022-llm-optimized-module-architecture.md`
4. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`
5. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-ADR22-MIGRATION/README.md`

---

## Inventário de Imports a Migrar

| Arquivo | Import de domain | O que usa |
|---------|-----------------|-----------|
| `internal/modules/task/models.go` | `domain.Task`, `domain.TaskStatus`, `domain.Priority`, `domain.RiskLevel` | `ToDomain()`, `FromDomain()` |

**Nota:** Todos os outros arquivos do módulo task já estão limpos. Este plano é focado em remover `ToDomain()`/`FromDomain()` e atualizar consumidores.

---

## Estratégia de Migração

### Passo 1: Remover ToDomain/FromDomain de models.go
- Apagar `func ToDomain(t *Task) *domain.Task`
- Apagar `func FromDomain(t *domain.Task) *Task`
- Remover `import "github.com/levygit837-cyber/OrchestraOS/internal/domain"` de `models.go`
- Verificar se há outros imports de domain no módulo task

### Passo 2: Criar Adapters Temporários nos Consumidores
Nos arquivos que ainda precisam de `domain.Task`, criar adapters com TODO:

```go
// TODO[ADR-0022]: remover quando orchestrator/models.go for desacoplado de domain.Task
func taskToDomain(t *task.Task) *domain.Task {
    if t == nil { return nil }
    return &domain.Task{
        ID:                   t.ID,
        Title:                t.Title,
        Description:          t.Description,
        Status:               domain.TaskStatus(t.Status),
        Priority:             domain.Priority(t.Priority),
        RiskLevel:            domain.RiskLevel(t.RiskLevel),
        CreatedFromMessageID: t.CreatedFromMessageID,
        AcceptanceCriteria:   t.AcceptanceCriteria,
        CreatedAt:            t.CreatedAt,
        UpdatedAt:            t.UpdatedAt,
    }
}
```

**Consumidores que precisam de adapters:**

| Consumidor | Motivo | Adapter |
|-----------|--------|---------|
| `internal/bootstrap/services.go:33-49` | `taskToDomain()` usado por `orchestrator.TaskServiceReader`, `run.TaskReader`, `workunit.TaskReader` | Manter adapter, atualizar para usar `task.Task` quando interfaces forem migradas |
| `internal/core/coordination/prompt_orchestrator.go:86-97` | Inline `&domain.Task{}` passado para `prompt.PrepareAndPersistInput` | Criar `taskToDomain()` adapter |
| `internal/modules/orchestrator/models.go` | `TaskServiceReader` interface retorna `*domain.Task` | Manter adapter até rodada do orchestrator |
| `internal/modules/run/service.go` | `TaskReader` interface retorna `*domain.Task` | Manter adapter até rodada do run |
| `internal/modules/workunit/service.go` | `TaskReader` interface retorna `*domain.Task` | Manter adapter até rodada do workunit |
| `internal/modules/taskgraph/planner_prompt.go` | `PlannerPrompt(task *domain.Task)` | Manter adapter até rodada do taskgraph |

**Regra:** Os adapters vivem nos **consumidores**, não no módulo `task`.

### Passo 3: Atualizar interfaces dos consumidores (se possível nesta rodada)
- `internal/modules/orchestrator/models.go` — `TaskServiceReader.GetByID` pode receber `*task.Task` se orchestrator já aceita (não, orchestrator será depois)
- **NÃO** alterar interfaces de outros módulos nesta rodada — isso é responsabilidade dos planos individuais de cada módulo.
- **Apenas** criar adapters nos consumidores para manter build passando.

### Passo 4: Build + Test + Commit
```bash
go build ./...
go test ./...
./scripts/safe-commit.sh "ADR-0022: finalize Task module type isolation"
```

---

## Critérios de Aceitação

- [ ] `internal/modules/task/models.go` NÃO contém `ToDomain()` nem `FromDomain()`
- [ ] `internal/modules/task/models.go` NÃO importa `internal/domain`
- [ ] NENHUM arquivo em `internal/modules/task/` importa `internal/domain`
- [ ] Todos os consumidores que precisam de `domain.Task` têm adapter com `// TODO[ADR-0022]: ...`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh` passa
