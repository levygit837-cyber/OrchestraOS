# Checklist — ORCH-F05-R03-A04: TaskGraph Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- **Política:** Ver ADR-0026 (Module Import Policy). `internal/domain` só pode ser usado para `EventEnvelope`, `EventPriority`, e tipos genéricos. NUNCA para entity structs.
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisitos: A01 (task) 🟢 e A03 (workunit) 🟢.

---

## Setup
- [ ] Branch `adr22-a04-taskgraph-types` criada e checkout feito
- [ ] Worktree `../orchestraos-a04-taskgraph` ativa
- [ ] `internal/modules/taskgraph/README.md` lido
- [ ] `internal/modules/taskgraph/CONTRACTS.md` lido
- [ ] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [ ] A01 (task) e A03 (workunit) estão no estado 🟢

---

## Passo 1 — models.go
- [ ] `TaskGraph` struct definida localmente
- [ ] `Status` enum definido localmente
- [ ] `TaskGraphCreatedPayload` struct definida localmente (antes em domain)
- [ ] Tags JSON mantidas idênticas
- [ ] Prefixo removido (`domain.TaskGraphStatus` → `Status`)
- [ ] `import "internal/domain"` removido de `models.go`

---

## Passo 2 — Arquivos Internos
- [ ] `repository.go`: `*domain.TaskGraph` → `*TaskGraph`
- [ ] `service.go`: `domain.TaskGraph` → `TaskGraph`, `domain.Task` → `task.Task`, `domain.WorkUnit` → `workunit.WorkUnit`
- [ ] `planner.go`: `domain.Task` → `task.Task`
- [ ] `planner_prompt.go`: `domain.Task` → `task.Task`
- [ ] `heuristic.go`: `domain.Task` → `task.Task`, `domain.WorkUnit` → `workunit.WorkUnit`
- [ ] `gemini_planner.go`: `domain.Task` → `task.Task`
- [ ] `planner_validator.go`: verificado e atualizado se necessário
- [ ] Testes do módulo atualizados

---

## Passo 3 — Interfaces de Planner
- [ ] Interface `Planner.Plan()` usa `*task.Task` (não `*domain.Task`)
- [ ] `heuristic.go` — `buildLocalHeuristicGraphPlan` usa `*task.Task`
- [ ] `gemini_planner.go` — `Plan` usa `*task.Task`

---

## Passo 4 — Adapters Temporários nos Consumidores
- [ ] `internal/modules/orchestrator/models.go` — `TaskGraphManager` anotado
- [ ] `internal/modules/workunit/service.go` — `TaskGraphManager` anotado
- [ ] `internal/bootstrap/services.go` — adapter atualizado

---

## Passo 5 — Validação
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate TaskGraph types to modules/taskgraph"` passa

---

## Status
- **Agente:** A04
- **Módulo:** taskgraph
- **Início:** ___
- **Término:** ___
- **Build:** ___
- **Testes:** ___
- **Commit:** ___
