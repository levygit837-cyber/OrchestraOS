# Padrão de Adapters Temporários para Migração ADR-0022

## Contexto

Durante a migração de tipos de `internal/domain/` para `internal/modules/<entidade>/`, cada módulo passa a ter seus próprios structs e enums. Isso quebra as interfaces de consumidores que ainda esperam `domain.Task`, `domain.Run`, etc.

## Solução: Adapter Temporário (Bridge Pattern)

Em vez de migrar todos os consumidores simultaneamente (impossível sem quebrar o build), usamos **adapters de conversão** nos callers permitidos (`bootstrap/`, `core/orchestration/`, `cmd/`) até que as interfaces dos consumidores também sejam migradas.

### Por que "temporários"?

Quando **todos** os consumidores de um tipo forem migrados para usar o tipo local do módulo, o adapter pode ser removido. É um ponte, não uma solução permanente.

---

## Regra de Ouro

> **Apenas pacotes permitidos podem conter adapters:**
> - `internal/bootstrap/` — adapters para injeção de dependências
> - `internal/core/orchestration/` — adapters para coordenação cross-module
> - `cmd/` — adapters para entrypoints CLI
>
> **Módulos em `internal/modules/*` NUNCA devem ter adapters.** Eles devem usar seus próprios tipos nativamente.

---

## Padrão de Implementação

### 1. No módulo migrado (ex: `internal/modules/task/`)

```go
// models.go — tipos locais, SEM import de domain
package task

type Status string
type Priority string
type RiskLevel string

type Task struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Status    Status    `json:"status"`
    Priority  Priority  `json:"priority"`
    RiskLevel RiskLevel `json:"risk_level"`
    // ...
}
```

**NENHUMA função de conversão no módulo.** O módulo é 100% independente.

### 2. No caller permitido (ex: `internal/bootstrap/`)

```go
// services.go — adapter inline ou helper
package bootstrap

// TODO: remover quando orchestrator.TaskServiceReader usar *task.Task
func taskToDomain(t *taskmod.Task) *domain.Task {
    if t == nil {
        return nil
    }
    return &domain.Task{
        ID:        t.ID,
        Title:     t.Title,
        Status:    domain.TaskStatus(t.Status),
        Priority:  domain.Priority(t.Priority),
        RiskLevel: domain.RiskLevel(t.RiskLevel),
        // ...
    }
}

type taskAdapter struct {
    db  *sql.DB
    svc *taskmod.TaskService
}

func (a *taskAdapter) GetByID(ctx context.Context, id string) (*domain.Task, error) {
    t, err := taskmod.NewRepository(a.db).GetByID(id)
    if err != nil {
        return nil, err
    }
    return taskToDomain(t), nil  // adapter em ação
}
```

### 3. No orchestrator (coordenação cross-module)

```go
// prompt_orchestrator.go
package orchestration

func (o *PromptOrchestrator) PrepareRunPrompt(...) (...) {
    task, _ := taskmod.RequireByID(ctx, tx, run.TaskID)
    // ...
    return o.promptService.PrepareAndPersistPrompt(ctx, tx, promptmod.PrepareAndPersistInput{
        // TODO: remover quando prompt usar *task.Task
        Task: &domain.Task{
            ID:        task.ID,
            Title:     task.Title,
            Status:    domain.TaskStatus(task.Status),
            // ...
        },
    })
}
```

---

## Quando remover os adapters?

Os adapters são removidos na **Sessão Final (Cleanup)** quando:

1. Todas as interfaces de consumidores foram migradas para usar tipos locais.
2. Nenhum código em `bootstrap/`, `orchestration/` ou `cmd/` referencia `domain.Task`.
3. `go build ./...` e `go test ./...` passam sem adapters.

### Sequência de remoção

```
Sessão 1: Task module → adapters criados em bootstrap/orchestration
Sessão 2: Run module  → adapters criados em bootstrap/orchestration
...
Sessão 10: Cleanup domain/types.go
Sessão 11: Remover TODOS os adapters, migrar interfaces dos consumidores
```

---

## Aplicabilidade a Todos os Módulos

Sim, **todos os módulos** devem seguir esse padrão durante a migração:

| Módulo | Adapter em | Converte para |
|--------|-----------|---------------|
| `task` | `bootstrap/services.go`, `orchestration/prompt_orchestrator.go` | `domain.Task` |
| `run` | `bootstrap/services.go`, `orchestration/*` | `domain.Run` |
| `workunit` | `bootstrap/services.go`, `orchestration/*` | `domain.WorkUnit` |
| `agentsession` | `bootstrap/services.go`, `orchestration/*` | `domain.AgentSession` |
| `taskgraph` | `bootstrap/services.go` | `domain.TaskGraph` |
| `agent` | `bootstrap/services.go` | `domain.Agent` |
| `prompt` | `orchestration/prompt_orchestrator.go` | `domain.PromptSnapshot` |
| `trigger` | `bootstrap/services.go` | `domain.Trigger` |
| `review` | `bootstrap/services.go` | `domain.Review` |

### Após migração completa

Quando todos os módulos tiverem tipos locais e todas as interfaces forem atualizadas:

- `bootstrap/` injeta `*taskmod.TaskService` diretamente no `orchestrator`
- `orchestrator` recebe `*task.Task`, `*run.Run`, etc.
- `domain/types.go` contém apenas `EventEnvelope`, `EventPriority`
- **Zero adapters**

---

## Vantagens do padrão

1. **Build verde a cada sessão** — não quebra o compilador.
2. **Migração incremental** — um módulo por vez.
3. **Rastreabilidade** — TODOs marcam onde a dívida técnica está.
4. **Reversível** — se algo quebrar, o adapter pode ser ajustado sem reverter o módulo.
