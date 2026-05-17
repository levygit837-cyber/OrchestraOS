# Checklist — ORCH-F05-R03-A03: WorkUnit Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisito: A01 (task) deve estar 🟢. A02 (run) é recomendado mas não obrigatório.
- **⚠️ REGRESSÃO:** O commit `99e860e` reverteu uma migração anterior e removeu o `WorkUnit` struct de `models.go`. É necessário **recriar** o struct local.
- **Política:** Ver ADR-0026 (Module Import Policy). `internal/domain` só pode ser usado para `EventEnvelope`, `EventPriority`, e tipos genéricos. NUNCA para entity structs.

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
- [ ] `WorkUnit` struct **recriada** localmente (foi removida no commit `99e860e`)
- [ ] `Status` enum definido localmente (`type Status string`)
- [ ] Todas as constantes de status definidas localmente
- [ ] Tags JSON mantidas idênticas às originais
- [ ] `import "internal/domain"` removido de `models.go`
- [ ] `models.go` compila isoladamente (`go build ./internal/modules/workunit/`)

---

## Passo 2 — Arquivos Internos
- [ ] `repository.go`: `*domain.WorkUnit` → `*WorkUnit`, `[]domain.WorkUnit` → `[]WorkUnit`
- [ ] `service.go`: `domain.WorkUnit` → `WorkUnit`, `domain.WorkUnitStatus` → `Status`
- [ ] `service_create.go`: `domain.WorkUnit` → `WorkUnit`, `domain.WorkUnitStatusCreated` → `StatusCreated`
- [ ] `fetch.go`: `*domain.WorkUnit` → `*WorkUnit`
- [ ] `events.go`: `domain.WorkUnitStatus` → `Status`
- [ ] `validation.go`: valida `Status` local, `domain.WorkUnitStatus*` → `Status*`
- [ ] `doc.go` atualizado (não menciona `domain.WorkUnit` como shared type)
- [ ] Testes do módulo atualizados

---

## Passo 3 — Interfaces Cruzadas
- [ ] `TaskReader` interface usa `*task.Task` (A01 concluído) ✅
- [ ] `TaskGraphManager` interface mantém `*domain.TaskGraph` com `// TODO[ADR-0022]: migrar para *taskgraph.TaskGraph quando A04 for concluído`
- [ ] Nenhuma referência a `domain.WorkUnit` ou `domain.WorkUnitStatus` em interfaces

---

## Passo 4 — Adapters Temporários nos Consumidores
- [ ] `internal/modules/orchestrator/models.go` — `WorkUnitLister` anotado `// TODO[ADR-0022]: migrar para []workunit.WorkUnit`
- [ ] `internal/modules/taskgraph/service.go` — `WorkUnitCreator`/`WorkUnitLister` anotados `// TODO[ADR-0022]: migrar para workunit.WorkUnit`
- [ ] `internal/modules/prompt/service.go` — `PrepareAndPersistInput.WorkUnit` anotado `// TODO[ADR-0022]: migrar para *workunit.WorkUnit`
- [ ] `internal/modules/trigger/service.go` — `WorkUnitReader` anotado `// TODO[ADR-0022]: migrar para *workunit.WorkUnit`
- [ ] `internal/core/coordination/*` — adapters criados com `// TODO[ADR-0022]: ...`
- [ ] `internal/bootstrap/services.go` — `workunitToDomain()` adapter atualizado com `// TODO[ADR-0022]: ...`

---

## Passo 5 — Validação de Isolamento
- [ ] `grep -rn "internal/domain" internal/modules/workunit/` só retorna `EventEnvelope`/`EventPriority` (zero entity structs)
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
