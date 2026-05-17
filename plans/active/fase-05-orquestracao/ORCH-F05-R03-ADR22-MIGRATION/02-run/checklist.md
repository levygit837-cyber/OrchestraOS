# Checklist — ORCH-F05-R03-A02: Run Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisito: A01 (task) deve estar 🟢 antes de iniciar este checklist.

---

## Setup
- [x] Branch `agent-a02/run-types-migration` criada e checkout feito
- [x] Worktree `../orchestraos-adr22-a02-run` ativa
- [x] `internal/modules/run/README.md` lido
- [x] `internal/modules/run/CONTRACTS.md` lido
- [x] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [x] A01 (task) está no estado 🟢 (commitado em `main` ou branch base)
- [x] A03 (workunit) está no estado 🟢 (mergeado em `master` via PR #15, commit `8b8e5af`)

---

## Passo 1 — models.go
- [x] `Run` struct definida localmente (copiada de domain/types.go)
- [x] `Status` enum definido localmente (`type Status string` + constantes)
- [x] `Result` enum definido localmente (`type Result string` + constantes)
- [x] Tags JSON mantidas idênticas às originais
- [x] Prefixo do pacote removido (`domain.RunStatus` → `Status`)
- [x] `import "internal/domain"` removido de `models.go`

---

## Passo 2 — Arquivos Internos
- [x] `repository.go`: `*domain.Run` → `*Run`, `[]domain.Run` → `[]Run`
- [x] `service.go`: `domain.RunStatus` → `Status`, `domain.RunResult` → `Result`, `domain.Run` → `Run`
- [x] `service_retry.go`: `domain.RunStatus` → `Status`
- [x] `fetch.go`: retorno `*domain.Run` → `*Run`
- [x] `events.go`: `domain.RunStatus` → `Status`, `domain.RunResult` → `Result`
- [x] `queries.go` verificado (não referencia tipos de domain)
- [x] Testes do módulo run atualizados (`run/*_test.go`) — nenhum teste existente, todos os testes do projeto passam

---

## Passo 3 — Interfaces Cruzadas
- [x] `run/service.go` — `TaskReader` interface usa `*task.Task` (não `*domain.Task`)
- [x] `run/service.go` — `WorkUnitReader` interface usa `*workunit.WorkUnit` (não `*domain.WorkUnit`)
- [x] `tests/architecture/module_boundaries_test.go` atualizado: adicionar `run` → `workunit`
- [x] Nenhum adapter `taskToDomain()` ou `workunitToDomain()` em `run/`

---

## Passo 4 — Adapters Temporários nos Consumidores
- [x] `internal/modules/orchestrator/models.go` — `RunLifecycleManager` anotado `// TODO[ADR-0022]: ...` (já existente)
- [x] `internal/modules/orchestrator/service.go` — adapters inline para `domain.RunStatus` anotados (já existente)
- [x] `internal/modules/trigger/models.go` — `RunReader` anotado `// TODO[ADR-0022]: ...` (já existente)
- [x] `internal/core/coordination/*` — adapters criados com `// TODO[ADR-0022]: ...` (já existente)
- [x] `internal/bootstrap/services.go` — run adapter atualizado com `// TODO[ADR-0022]: ...` — `runWorkUnitReaderAdapter` removido; `workunit.Repository` conectado diretamente

---

## Passo 5 — Validação de Isolamento
- [x] `grep -rn "internal/domain" internal/modules/run/` só retorna `EventEnvelope`/`EventPriority`
- [x] `grep -rn "domain\.WorkUnit" internal/modules/run/` retorna ZERO resultados
- [x] `go build ./...` passa
- [x] `go test ./...` passa
- [x] `./scripts/verify-contracts.sh` passa
- [x] `./scripts/lint.sh` passa
- [x] `./scripts/safe-commit.sh "ADR-0022: migrate Run module WorkUnitReader to use *workunit.WorkUnit directly"` passa

---

## Status
- **Agente:** A02
- **Módulo:** run
- **Início:** 2026-05-17
- **Término:** 2026-05-17
- **Build:** ✅
- **Testes:** ✅
- **Commit:** `3d432cc`
- **PR:** #16 — https://github.com/levygit837-cyber/OrchestraOS/pull/16
