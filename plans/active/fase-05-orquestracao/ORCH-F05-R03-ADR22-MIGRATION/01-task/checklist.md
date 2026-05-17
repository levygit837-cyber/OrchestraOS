# Checklist — ORCH-F05-R03-A01: Task Module Type Migration (Finalização)

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Esta sessão termina com `./scripts/safe-commit.sh`.

---

## Setup
- [x] Branch `adr22-a01-task-types` criada e checkout feito
- [x] Worktree `../orchestraos-a01-task` ativa
- [x] `internal/modules/task/README.md` lido
- [x] `internal/modules/task/CONTRACTS.md` lido
- [x] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [x] Plano geral ORCH-F05-R03-A01 lido

---

## Passo 1 — Limpar models.go
- [x] `ToDomain()` removido de `internal/modules/task/models.go`
- [x] `FromDomain()` removido de `internal/modules/task/models.go`
- [x] `import "github.com/levygit837-cyber/OrchestraOS/internal/domain"` removido de `models.go`
- [x] `models.go` compila isoladamente (`go build ./internal/modules/task/`)

---

## Passo 2 — Verificar módulo task 100% limpo
- [x] `grep -rn "internal/domain" internal/modules/task/` retorna ZERO resultados
- [x] `contract.go` não importa domain
- [x] `doc.go` não importa domain
- [x] `events.go` não importa domain
- [x] `fetch.go` não importa domain
- [x] `queries.go` não importa domain
- [x] `repository.go` não importa domain
- [x] `service.go` não importa domain
- [x] `validation_test.go` não importa domain

---

## Passo 3 — Adapters Temporários nos Consumidores
- [x] `internal/bootstrap/services.go` — adapter `taskToDomain()` tem `// TODO[ADR-0022]: ...`
- [x] `internal/core/coordination/prompt_orchestrator.go` — adapter criado com `// TODO[ADR-0022]: ...`
- [x] `internal/modules/orchestrator/models.go` — interface `TaskServiceReader` anotada com `// TODO[ADR-0022]: migrar para *task.Task`
- [x] `internal/modules/run/service.go` — interface `TaskReader` anotada com `// TODO[ADR-0022]: migrar para *task.Task`
- [x] `internal/modules/workunit/service.go` — interface `TaskReader` anotada com `// TODO[ADR-0022]: migrar para *task.Task`
- [x] `internal/modules/taskgraph/planner_prompt.go` — função `PlannerPrompt` anotada com `// TODO[ADR-0022]: migrar para *task.Task`

---

## Passo 4 — Validação
- [x] `go build ./...` passa
- [x] `go test ./...` passa
- [x] `./scripts/verify-contracts.sh` passa
- [x] `./scripts/lint.sh` passa
- [x] `./scripts/safe-commit.sh "ADR-0022: finalize Task module type isolation"` passa

---

## Status
- **Agente:** A01
- **Módulo:** task
- **Início:** 2026-05-16
- **Término:** 2026-05-16
- **Build:** pass
- **Testes:** pass
- **Commit:** `6e97d31`
