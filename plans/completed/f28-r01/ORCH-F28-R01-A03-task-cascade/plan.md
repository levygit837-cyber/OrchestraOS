# ORCH-F28-R01-A03: Task Module — Cascade Cancellation

## Contexto do Projeto

**Projeto:** OrchestraOS (Go monolith modular)  
**Arquitetura:** Vertical Slice (ADR-0022) + Hybrid Intelligent Orchestrator (ADR-0023)  
**Meta:** Mover `CancelTaskDependents` de `internal/core/coordination/` para `internal/modules/task/` conforme ADR-0028  
**Risco:** **MÉDIO-ALTO** — envolve cross-module logic e o módulo `task` é atualmente um `leafModule` que não pode importar outros módulos.

---

## Documentação Obrigatória (LEIA ANTES DE COMEÇAR)

1. `docs/adr/0028-core-architecture-and-naming-standards.md` — Seções 2.1.1, 2.4
2. `docs/adr/0022-vertical-slice-modules.md` — Regras de módulos, leaf modules, DI
3. `AGENTS.md` — Regras de commits
4. Este arquivo (`plan.md`)

---

## O que Já Existe vs O que Deve Ser Feito

### Arquivo-fonte a mover

| Arquivo Atual (coordination/) | Novo Arquivo (task/) | Responsabilidade | Linhas |
|---|---|---|---|
| `cascade.go` | `service_cascade.go` | Cancelar runs e WUs não-terminais de uma task | ~63 |

### Pontos de integração a atualizar

| Arquivo | Uso Atual | Mudança Necessária |
|---|---|---|
| `internal/bootstrap/services.go` | `coordination.CancelTaskDependents` callback para `taskmod.NewTaskService` | `taskmod.CancelTaskDependents` (ou método interno) |
| `internal/modules/task/service.go` | Recebe `onCancel func(...)` no construtor | Se `CancelTaskDependents` virar método do TaskService, remover callback do construtor |

### Desafio Arquitetural Crítico

O módulo `task` é definido como `leafModule` em `tests/architecture/module_boundaries_test.go`:

```go
var leafModules = map[string]bool{
    "agent": true,
    "task":  true,
}
```

E `allowedModuleImports` NÃO tem entrada para `task`, o que significa que `task` não pode importar NENHUM outro módulo.

Porém, `cascade.go` importa:
- `runmod` (para `NewRepository`, `ListByTask`, `StatusCancelled`, `ResultForStatus`)
- `workunitmod` (para `NewRepository`, `ListByTask`, `StatusCancelled`, `QueryUpdateStatus`)

**Opções para resolver:**

1. **Mover `CancelTaskDependents` para o `orchestrator` module** (que já pode importar todos os módulos) e manter `task` como leaf. O `orchestrator` já é o módulo que coordena cross-module. A implementação ficaria em `internal/modules/orchestrator/service_cascade.go`. O `bootstrap` passaria o callback do `orchestrator` para o `task`. **NOTA:** Esta opção DIVERGE do ADR-0028, mas preserva a arquitetura de leaf modules. Se escolher esta, DOCUMENTE a divergência em uma nota no código.

2. **Tornar `task` um non-leaf module** permitindo imports de `run` e `workunit`. Adicione à `allowedModuleImports`:
   ```go
   "task": {"run": true, "workunit": true},
   ```
   E remova `task` de `leafModules`. **NOTA:** Esta opção segue o ADR-0028 literalmente.

3. **Usar callbacks/funções injetadas** em vez de imports diretos. O `TaskService` já recebe `onCancel` como callback. A implementação do callback pode continuar fora do módulo `task`. Mas o ADR-0028 diz para mover para `task`.

**RECOMENDAÇÃO DO PLANO:** Siga o ADR-0028 (Opção 2). Mova para `task/service_cascade.go`, remova `task` de `leafModules`, adicione `"task": {"run": true, "workunit": true}` em `allowedModuleImports`, e documente a justificativa: "Task é o aggregate raiz do fluxo de cancelamento; imports de run e workunit são para DI de repositórios no fluxo de cascade."

---

## Fronteiras de Isolamento

**TOCAR:**
- `internal/modules/task/service_cascade.go` — NOVO arquivo com `CancelTaskDependents`
- `internal/modules/task/service.go` — possivelmente remover `onCancel` callback do construtor se `CancelTaskDependents` virar método interno
- `internal/bootstrap/services.go` — ajustar callback de `taskmod.NewTaskService`
- `tests/architecture/module_boundaries_test.go` — atualizar `leafModules` e `allowedModuleImports`
- `internal/core/coordination/cascade.go` — REMOVER após mover lógica

**EVITAR:**
- `internal/modules/run/` — não toque (tarefa do A02, exceto se A02 já entregou)
- `internal/modules/agentsession/` — não toque (tarefa do A04)
- `internal/modules/prompt/` — não toque (tarefa do A05)
- `internal/modules/orchestrator/service.go` — não altere lógica do orchestrator, apenas bootstrap
- `internal/core/transition/` — não toque exceto se necessário para compilar

---

## Interfaces Contratuais

Nenhuma nova interface necessária. `CancelTaskDependents` mantém a mesma assinatura:

```go
func CancelTaskDependents(ctx context.Context, tx *sql.Tx, taskID string, input transition.TransitionInput) error
```

Se `CancelTaskDependents` virar um método do `TaskService` (recomendado para coesão):

```go
func (s *TaskService) CancelDependents(ctx context.Context, tx *sql.Tx, taskID string, input transition.TransitionInput) error
```

Neste caso, o construtor `NewTaskService` pode perder o parâmetro `onCancel`.

---

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

**Caminho do checklist:** `plans/active/f28-r01/ORCH-F28-R01-A03-task-cascade/checklist.md`

**A cada iteração:**
1. **LER** o checklist para identificar o próximo item pendente
2. **EXECUTAR** o item (código, teste, refactor)
3. **VALIDAR** o item (`go build ./...` no escopo afetado)
4. **ATUALIZAR** o checklist marcando o item como concluído
5. **CONTINUAR** para o próximo item

---

## Regras de Implementação

1. **Mover `CancelTaskDependents`:**
   - Copiar conteúdo de `coordination/cascade.go` para `task/service_cascade.go`
   - Ajustar package de `coordination` para `task`
   - Ajustar imports (runmod e workunitmod continuam sendo importados, agora pelo módulo task)
   - `UpdateRunProjection` é chamado em `cascade.go` — verifique se A02 já moveu isso para `run/repository.go`. Se sim, importe `runmod` e use `runmod.NewRepository(tx).UpdateProjection(...)`. Se não, você precisa chamar via callback ou esperar A02.

2. **IMPORTANTE — Dependência do A02:**
   - `cascade.go` chama `coordination.UpdateRunProjection`. Se o A02 já entregou e moveu `UpdateRunProjection` para `run/repository.go`, use `runmod.NewRepository(tx).UpdateProjection(...)`.
   - Se A02 ainda não entregou, seu código não compilará. Neste caso:
     a) Aguarde A02, OU
     b) Mantenha o import de `coordination` temporariamente APENAS para `UpdateRunProjection`, com um TODO: "Remover quando A02 entregar"
   - **O plano assume que você pode usar `runmod.UpdateProjection` ou similar.**

3. **Atualizar `task/service.go`:**
   - Se `CancelTaskDependents` virar método do `TaskService`, remova o parâmetro `onCancel` do construtor `NewTaskService`
   - Atualize todos os callers de `NewTaskService` (apenas `bootstrap/services.go` e possivelmente testes)

4. **Atualizar `bootstrap/services.go`:**
   - `taskmod.NewTaskService(db, coordination.CancelTaskDependents)` → `taskmod.NewTaskService(db)` (se virou método) ou `taskmod.NewTaskService(db, taskmod.CancelTaskDependents)` (se permaneceu função)

5. **Atualizar architecture test:**
   - Em `tests/architecture/module_boundaries_test.go`:
     - Remover `"task": true` de `leafModules`
     - Adicionar `"task": {"run": true, "workunit": true}` em `allowedModuleImports`
     - Atualizar o comentário documentando o novo import permitido

---

## Regras Rígidas de Testes

1. Rode `go test ./internal/modules/task/...` — deve passar
2. Rode `go test ./tests/architecture/...` — deve passar (especialmente `module_boundaries_test.go`)
3. Rode `go test ./...` ao final — deve passar
4. Rode `./scripts/verify-contracts.sh` — deve passar
5. Rode `./scripts/lint.sh` — deve passar

---

## Code Review Auto-Crítico Obrigatório

- [ ] `internal/core/coordination/cascade.go` foi removido?
- [ ] `internal/modules/task/service_cascade.go` existe e contém `CancelTaskDependents`?
- [ ] `task` foi removido de `leafModules`?
- [ ] `allowedModuleImports` inclui `task` com `run` e `workunit`?
- [ ] Bootstrap não usa mais `coordination.CancelTaskDependents`?
- [ ] Se `CancelTaskDependents` virou método do TaskService, o construtor foi simplificado?
- [ ] `go build ./...` passa?
- [ ] `go test ./...` passa?
- [ ] Architecture tests passam?

---

## Critérios de Aceite Verificáveis

1. `internal/modules/task/service_cascade.go` existe e exporta `CancelTaskDependents`
2. `internal/core/coordination/cascade.go` NÃO existe mais
3. `task` não está mais em `leafModules`
4. `allowedModuleImports["task"]` contém `run` e `workunit`
5. `bootstrap/services.go` não usa símbolos de `coordination` para cancelamento
6. `go build ./...` passa
7. `go test ./...` passa
8. `./scripts/verify-contracts.sh` passa
9. `./scripts/lint.sh` passa
10. Architecture tests passam

---

## Entrega Final

1. Commit com `./scripts/safe-commit.sh "refactor(task): move CancelTaskDependents from coordination per ADR-0028"`
2. Push da feature branch
3. Reportar ao usuário: "Task cascade migration completa. Build verde."
# ORCH-F28-R01-A03: Task Module — Cascade Cancellation

## Contexto do Projeto

**Projeto:** OrchestraOS (Go monolith modular)  
**Arquitetura:** Vertical Slice (ADR-0022) + Hybrid Intelligent Orchestrator (ADR-0023)  
**Meta:** Mover `CancelTaskDependents` de `internal/core/coordination/` para `internal/modules/task/` conforme ADR-0028  
**Risco:** **MÉDIO-ALTO** — envolve cross-module logic e o módulo `task` é atualmente um `leafModule` que não pode importar outros módulos.

---

## Documentação Obrigatória (LEIA ANTES DE COMEÇAR)

1. `docs/adr/0028-core-architecture-and-naming-standards.md` — Seções 2.1.1, 2.4
2. `docs/adr/0022-vertical-slice-modules.md` — Regras de módulos, leaf modules, DI
3. `AGENTS.md` — Regras de commits
4. Este arquivo (`plan.md`)

---

## O que Já Existe vs O que Deve Ser Feito

### Arquivo-fonte a mover

| Arquivo Atual (coordination/) | Novo Arquivo (task/) | Responsabilidade | Linhas |
|---|---|---|---|
| `cascade.go` | `service_cascade.go` | Cancelar runs e WUs não-terminais de uma task | ~63 |

### Pontos de integração a atualizar

| Arquivo | Uso Atual | Mudança Necessária |
|---|---|---|
| `internal/bootstrap/services.go` | `coordination.CancelTaskDependents` callback para `taskmod.NewTaskService` | `taskmod.CancelTaskDependents` (ou método interno) |
| `internal/modules/task/service.go` | Recebe `onCancel func(...)` no construtor | Se `CancelTaskDependents` virar método do TaskService, remover callback do construtor |

### Desafio Arquitetural Crítico

O módulo `task` é definido como `leafModule` em `tests/architecture/module_boundaries_test.go`:

```go
var leafModules = map[string]bool{
    "agent": true,
    "task":  true,
}
```

E `allowedModuleImports` NÃO tem entrada para `task`, o que significa que `task` não pode importar NENHUM outro módulo.

Porém, `cascade.go` importa:
- `runmod` (para `NewRepository`, `ListByTask`, `StatusCancelled`, `ResultForStatus`)
- `workunitmod` (para `NewRepository`, `ListByTask`, `StatusCancelled`, `QueryUpdateStatus`)

**Opções para resolver:**

1. **Mover `CancelTaskDependents` para o `orchestrator` module** (que já pode importar todos os módulos) e manter `task` como leaf. O `orchestrator` já é o módulo que coordena cross-module. A implementação ficaria em `internal/modules/orchestrator/service_cascade.go`. O `bootstrap` passaria o callback do `orchestrator` para o `task`. **NOTA:** Esta opção DIVERGE do ADR-0028, mas preserva a arquitetura de leaf modules. Se escolher esta, DOCUMENTE a divergência em uma nota no código.

2. **Tornar `task` um non-leaf module** permitindo imports de `run` e `workunit`. Adicione à `allowedModuleImports`:
   ```go
   "task": {"run": true, "workunit": true},
   ```
   E remova `task` de `leafModules`. **NOTA:** Esta opção segue o ADR-0028 literalmente.

3. **Usar callbacks/funções injetadas** em vez de imports diretos. O `TaskService` já recebe `onCancel` como callback. A implementação do callback pode continuar fora do módulo `task`. Mas o ADR-0028 diz para mover para `task`.

**RECOMENDAÇÃO DO PLANO:** Siga o ADR-0028 (Opção 2). Mova para `task/service_cascade.go`, remova `task` de `leafModules`, adicione `"task": {"run": true, "workunit": true}` em `allowedModuleImports`, e documente a justificativa: "Task é o aggregate raiz do fluxo de cancelamento; imports de run e workunit são para DI de repositórios no fluxo de cascade."

---

## Fronteiras de Isolamento

**TOCAR:**
- `internal/modules/task/service_cascade.go` — NOVO arquivo com `CancelTaskDependents`
- `internal/modules/task/service.go` — possivelmente remover `onCancel` callback do construtor se `CancelTaskDependents` virar método interno
- `internal/bootstrap/services.go` — ajustar callback de `taskmod.NewTaskService`
- `tests/architecture/module_boundaries_test.go` — atualizar `leafModules` e `allowedModuleImports`
- `internal/core/coordination/cascade.go` — REMOVER após mover lógica

**EVITAR:**
- `internal/modules/run/` — não toque (tarefa do A02, exceto se A02 já entregou)
- `internal/modules/agentsession/` — não toque (tarefa do A04)
- `internal/modules/prompt/` — não toque (tarefa do A05)
- `internal/modules/orchestrator/service.go` — não altere lógica do orchestrator, apenas bootstrap
- `internal/core/transition/` — não toque exceto se necessário para compilar

---

## Interfaces Contratuais

Nenhuma nova interface necessária. `CancelTaskDependents` mantém a mesma assinatura:

```go
func CancelTaskDependents(ctx context.Context, tx *sql.Tx, taskID string, input transition.TransitionInput) error
```

Se `CancelTaskDependents` virar um método do `TaskService` (recomendado para coesão):

```go
func (s *TaskService) CancelDependents(ctx context.Context, tx *sql.Tx, taskID string, input transition.TransitionInput) error
```

Neste caso, o construtor `NewTaskService` pode perder o parâmetro `onCancel`.

---

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

**Caminho do checklist:** `plans/active/f28-r01/ORCH-F28-R01-A03-task-cascade/checklist.md`

**A cada iteração:**
1. **LER** o checklist para identificar o próximo item pendente
2. **EXECUTAR** o item (código, teste, refactor)
3. **VALIDAR** o item (`go build ./...` no escopo afetado)
4. **ATUALIZAR** o checklist marcando o item como concluído
5. **CONTINUAR** para o próximo item

---

## Regras de Implementação

1. **Mover `CancelTaskDependents`:**
   - Copiar conteúdo de `coordination/cascade.go` para `task/service_cascade.go`
   - Ajustar package de `coordination` para `task`
   - Ajustar imports (runmod e workunitmod continuam sendo importados, agora pelo módulo task)
   - `UpdateRunProjection` é chamado em `cascade.go` — verifique se A02 já moveu isso para `run/repository.go`. Se sim, importe `runmod` e use `runmod.NewRepository(tx).UpdateProjection(...)`. Se não, você precisa chamar via callback ou esperar A02.

2. **IMPORTANTE — Dependência do A02:**
   - `cascade.go` chama `coordination.UpdateRunProjection`. Se o A02 já entregou e moveu `UpdateRunProjection` para `run/repository.go`, use `runmod.NewRepository(tx).UpdateProjection(...)`.
   - Se A02 ainda não entregou, seu código não compilará. Neste caso:
     a) Aguarde A02, OU
     b) Mantenha o import de `coordination` temporariamente APENAS para `UpdateRunProjection`, com um TODO: "Remover quando A02 entregar"
   - **O plano assume que você pode usar `runmod.UpdateProjection` ou similar.**

3. **Atualizar `task/service.go`:**
   - Se `CancelTaskDependents` virar método do `TaskService`, remova o parâmetro `onCancel` do construtor `NewTaskService`
   - Atualize todos os callers de `NewTaskService` (apenas `bootstrap/services.go` e possivelmente testes)

4. **Atualizar `bootstrap/services.go`:**
   - `taskmod.NewTaskService(db, coordination.CancelTaskDependents)` → `taskmod.NewTaskService(db)` (se virou método) ou `taskmod.NewTaskService(db, taskmod.CancelTaskDependents)` (se permaneceu função)

5. **Atualizar architecture test:**
   - Em `tests/architecture/module_boundaries_test.go`:
     - Remover `"task": true` de `leafModules`
     - Adicionar `"task": {"run": true, "workunit": true}` em `allowedModuleImports`
     - Atualizar o comentário documentando o novo import permitido

---

## Regras Rígidas de Testes

1. Rode `go test ./internal/modules/task/...` — deve passar
2. Rode `go test ./tests/architecture/...` — deve passar (especialmente `module_boundaries_test.go`)
3. Rode `go test ./...` ao final — deve passar
4. Rode `./scripts/verify-contracts.sh` — deve passar
5. Rode `./scripts/lint.sh` — deve passar

---

## Code Review Auto-Crítico Obrigatório

- [ ] `internal/core/coordination/cascade.go` foi removido?
- [ ] `internal/modules/task/service_cascade.go` existe e contém `CancelTaskDependents`?
- [ ] `task` foi removido de `leafModules`?
- [ ] `allowedModuleImports` inclui `task` com `run` e `workunit`?
- [ ] Bootstrap não usa mais `coordination.CancelTaskDependents`?
- [ ] Se `CancelTaskDependents` virou método do TaskService, o construtor foi simplificado?
- [ ] `go build ./...` passa?
- [ ] `go test ./...` passa?
- [ ] Architecture tests passam?

---

## Critérios de Aceite Verificáveis

1. `internal/modules/task/service_cascade.go` existe e exporta `CancelTaskDependents`
2. `internal/core/coordination/cascade.go` NÃO existe mais
3. `task` não está mais em `leafModules`
4. `allowedModuleImports["task"]` contém `run` e `workunit`
5. `bootstrap/services.go` não usa símbolos de `coordination` para cancelamento
6. `go build ./...` passa
7. `go test ./...` passa
8. `./scripts/verify-contracts.sh` passa
9. `./scripts/lint.sh` passa
10. Architecture tests passam

---

## Entrega Final

1. Commit com `./scripts/safe-commit.sh "refactor(task): move CancelTaskDependents from coordination per ADR-0028"`
2. Push da feature branch
3. Reportar ao usuário: "Task cascade migration completa. Build verde."

---

## Resultado da Execução

**Data:** 2026-05-17  
**Branch:** `feat/adr28-a03-task-cascade`  
**PR:** #32

**Divergência do plano original:**
- O `CancelTaskDependents` NÃO foi movido para `task/` como previsto no ADR-0028 literal.
- Foi identificado um **import cycle** impossível de resolver: `run` já importa `task` (via `TaskReader`), logo `task` não pode importar `run`.
- A solução arquiteturalmente correta foi mover `CancelTaskDependents` para **`internal/modules/orchestrator/service_cascade.go`** (orchestrator já importa todos os módulos legitimamente).
- O `task` permaneceu como `leafModule` — nenhuma mudança em `allowedModuleImports` foi necessária.
- O bootstrap foi atualizado para passar `orchestratormod.CancelTaskDependents` como callback para `taskmod.NewTaskService`.
- Todos os checks passaram (build, test, vet, architecture, contracts, lint).
