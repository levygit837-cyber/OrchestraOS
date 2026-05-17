# Checklist — ORCH-F05-R03-A03: WorkUnit Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisito: A01 (task) deve estar 🟢. A02 (run) é recomendado mas não obrigatório.

---

## Setup
- [ ] Branch `adr22-a03-workunit-types` criada e checkout feito
- [ ] Worktree `../orchestraos-a03-workunit` ativa
- [ ] `internal/modules/workunit/README.md` lido
- [ ] `internal/modules/workunit/CONTRACTS.md` lido
- [ ] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [ ] A01 (task) está no estado 🟢

---

## Passo 1 — models.go
- [ ] `WorkUnit` struct definida localmente
- [ ] `Status` enum definido localmente
- [ ] Tags JSON mantidas idênticas
- [ ] Prefixo removido (`domain.WorkUnitStatus` → `Status`)
- [ ] `import "internal/domain"` removido de `models.go`

---

## Passo 2 — Arquivos Internos
- [ ] `repository.go`: `*domain.WorkUnit` → `*WorkUnit`
- [ ] `service.go`: `domain.WorkUnit` → `WorkUnit`, `domain.WorkUnitStatus` → `Status`
- [ ] `service_create.go`: `domain.WorkUnit` → `WorkUnit`
- [ ] `fetch.go`: `*domain.WorkUnit` → `*WorkUnit`
- [ ] `events.go`: `domain.WorkUnitStatus` → `Status`
- [ ] `validation.go`: valida `Status` local
- [ ] Testes do módulo atualizados

---

## Passo 3 — Interfaces Cruzadas
- [ ] `TaskReader` interface atualizada para `*task.Task` (A01 concluído)
- [ ] `TaskGraphManager` interface mantém `domain.TaskGraph` com adapter/anotação
- [ ] Verificar `Run` references — se A02 concluído, usar `*run.Run`; senão, adapter

---

## Passo 4 — Adapters Temporários nos Consumidores
- [ ] `internal/modules/orchestrator/models.go` — `WorkUnitLister` anotado
- [ ] `internal/modules/orchestrator/service.go` — adapters criados
- [ ] `internal/modules/taskgraph/service.go` — `WorkUnitCreator`/`WorkUnitLister` anotados
- [ ] `internal/modules/prompt/service.go` — `PrepareAndPersistInput.WorkUnit` anotado
- [ ] `internal/modules/trigger/service.go` — `WorkUnitReader` anotado
- [ ] `internal/core/coordination/*` — adapters criados
- [ ] `internal/bootstrap/services.go` — adapter atualizado

---

## Passo 5 — Validação
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate WorkUnit types to modules/workunit"` passa

---

## Status
- **Agente:** A03
- **Módulo:** workunit
- **Início:** ___
- **Término:** ___
- **Build:** ___
- **Testes:** ___
- **Commit:** ___
