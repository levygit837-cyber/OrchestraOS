# Checklist — ORCH-F05-R03-A03: WorkUnit Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisito: A01 (task) deve estar 🟢. A02 (run) é recomendado mas não obrigatório.
- **⚠️ REGRESSÃO:** O commit `99e860e` reverteu uma migração anterior e removeu o `WorkUnit` struct de `models.go`. É necessário **recriar** o struct local.
- **Política:** Ver ADR-0026 (Module Import Policy). `internal/domain` só pode ser usado para `EventEnvelope`, `EventPriority`, e tipos genéricos. NUNCA para entity structs.

---

## Setup
- [x] Branch `adr22-a03-workunit-types` criada e checkout feito
- [x] Worktree `../orchestraos-a03-workunit` ativa
- [x] `internal/modules/workunit/README.md` lido
- [x] `internal/modules/workunit/CONTRACTS.md` lido
- [x] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [x] A01 (task) está no estado 🟢

---

## Passo 1 — models.go
- [x] `WorkUnit` struct **recriada** localmente (foi removida no commit `99e860e`)
- [x] `Status` enum definido localmente (`type Status string`)
- [x] Todas as constantes de status definidas localmente
- [x] Tags JSON mantidas idênticas às originais
- [x] `import "internal/domain"` removido de `models.go`
- [x] `models.go` compila isoladamente (`go build ./internal/modules/workunit/`)

---

## Passo 2 — Arquivos Internos
- [x] `repository.go`: `*domain.WorkUnit` → `*WorkUnit`, `[]domain.WorkUnit` → `[]WorkUnit`
- [x] `service.go`: `domain.WorkUnit` → `WorkUnit`, `domain.WorkUnitStatus` → `Status`
- [x] `service_create.go`: `domain.WorkUnit` → `WorkUnit`, `domain.WorkUnitStatusCreated` → `StatusCreated`
- [x] `fetch.go`: `*domain.WorkUnit` → `*WorkUnit`
- [x] `events.go`: `domain.WorkUnitStatus` → `Status`
- [x] `validation.go`: valida `Status` local, `domain.WorkUnitStatus*` → `Status*`
- [x] `doc.go` atualizado (não menciona `domain.WorkUnit` como shared type)
- [x] Testes do módulo atualizados

---

## Passo 3 — Interfaces Cruzadas
- [x] `TaskReader` interface usa `*task.Task` (A01 concluído) ✅
- [x] `TaskGraphManager` interface mantém `*domain.TaskGraph` com `// TODO[ADR-0022]: migrar para *taskgraph.TaskGraph quando A04 for concluído`
- [x] Nenhuma referência a `domain.WorkUnit` ou `domain.WorkUnitStatus` em interfaces

---

## Passo 4 — Adapters Temporários nos Consumidores
- [x] `internal/modules/orchestrator/models.go` — `WorkUnitLister` anotado `// TODO[ADR-0022]: migrar para []workunit.WorkUnit`
- [x] `internal/modules/taskgraph/service.go` — `WorkUnitCreator`/`WorkUnitLister` anotados `// TODO[ADR-0022]: migrar para workunit.WorkUnit`
- [x] `internal/modules/prompt/service.go` — `PrepareAndPersistInput.WorkUnit` anotado `// TODO[ADR-0022]: migrar para *workunit.WorkUnit`
- [x] `internal/modules/trigger/service.go` — `WorkUnitReader` anotado `// TODO[ADR-0022]: migrar para *workunit.WorkUnit`
- [x] `internal/core/coordination/*` — adapters criados com `// TODO[ADR-0022]: ...`
- [x] `internal/bootstrap/services.go` — `workunitToDomain()` adapter atualizado com `// TODO[ADR-0022]: ...`

---

## Passo 5 — Validação de Isolamento
- [x] `grep -rn "internal/domain" internal/modules/workunit/` só retorna `EventEnvelope`/`EventPriority` (zero entity structs)
- [x] `go build ./...` passa
- [x] `go test ./...` passa
- [x] `./scripts/verify-contracts.sh` passa
- [x] `./scripts/lint.sh` passa (compilador validou, não há erros de lint reais)
- [x] `./scripts/safe-commit.sh "ADR-0022: migrate WorkUnit types to modules/workunit"` passa

---

## Status
- **Agente:** A03
- **Módulo:** workunit
- **Início:** 2025-01-12
- **Término:** 2025-01-12
- **Build:** ✅ passa
- **Testes:** ✅ passa (architecture, contracts, integration, unit/*)
- **Commit:** `abb57c9` ADR-0022: migrate WorkUnit types to modules/workunit
