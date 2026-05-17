# Checklist — ORCH-F05-R03-A02: Run Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisito: A01 (task) deve estar 🟢 antes de iniciar este checklist.

---

## Setup
- [ ] Branch `adr22-a02-run-types` criada e checkout feito
- [ ] Worktree `../orchestraos-a02-run` ativa
- [ ] `internal/modules/run/README.md` lido
- [ ] `internal/modules/run/CONTRACTS.md` lido
- [ ] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [ ] A01 (task) está no estado 🟢 (commitado em `main` ou branch base)

---

## Passo 1 — models.go
- [ ] `Run` struct definida localmente (copiada de domain/types.go)
- [ ] `Status` enum definido localmente (`type Status string` + constantes)
- [ ] `Result` enum definido localmente (`type Result string` + constantes)
- [ ] Tags JSON mantidas idênticas às originais
- [ ] Prefixo do pacote removido (`domain.RunStatus` → `Status`)
- [ ] `import "internal/domain"` removido de `models.go`

---

## Passo 2 — Arquivos Internos
- [ ] `repository.go`: `*domain.Run` → `*Run`, `[]domain.Run` → `[]Run`
- [ ] `service.go`: `domain.RunStatus` → `Status`, `domain.RunResult` → `Result`, `domain.Run` → `Run`
- [ ] `service_retry.go`: `domain.RunStatus` → `Status`
- [ ] `fetch.go`: retorno `*domain.Run` → `*Run`
- [ ] `events.go`: `domain.RunStatus` → `Status`, `domain.RunResult` → `Result`
- [ ] `queries.go` verificado (não referencia tipos de domain)
- [ ] Testes do módulo run atualizados (`run/*_test.go`)

---

## Passo 3 — Interfaces Cruzadas (Task já migrado)
- [ ] `run/service.go` — `TaskReader` interface usa `*task.Task` (não `*domain.Task`)
- [ ] Adapter `taskToDomain()` removido de `run/` se existir (task já é local)

---

## Passo 4 — Adapters Temporários nos Consumidores
- [ ] `internal/modules/orchestrator/models.go` — `RunLifecycleManager` anotado `// TODO[ADR-0022]: ...`
- [ ] `internal/modules/orchestrator/service.go` — adapters inline para `domain.RunStatus` anotados
- [ ] `internal/modules/trigger/models.go` — `RunReader` anotado `// TODO[ADR-0022]: ...`
- [ ] `internal/core/coordination/*` — adapters criados com `// TODO[ADR-0022]: ...`
- [ ] `internal/bootstrap/services.go` — run adapter atualizado com `// TODO[ADR-0022]: ...`

---

## Passo 5 — Validação
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Run types to modules/run"` passa

---

## Status
- **Agente:** A02
- **Módulo:** run
- **Início:** ___
- **Término:** ___
- **Build:** ___
- **Testes:** ___
- **Commit:** ___
