# Checklist — ORCH-F05-R03-A09: Review Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- **Política:** Ver ADR-0026 (Module Import Policy). `internal/domain` só pode ser usado para `EventEnvelope`, `EventPriority`, e tipos genéricos. NUNCA para entity structs.
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisitos: A01 (task) 🟢, A02 (run) 🟢, A03 (workunit) 🟢.

---

## Setup
- [x] Branch `adr22-a09-review-types` criada e checkout feito
- [x] Worktree `../orchestraos-a09-review` ativa
- [x] `internal/modules/review/README.md` lido
- [x] `internal/modules/review/CONTRACTS.md` lido
- [x] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [x] A01, A02, A03 estão no estado 🟢

---

## Passo 1 — models.go
- [x] `Review` struct definida localmente
- [x] `Status` enum definido localmente
- [x] `ValidationGate` enum definido localmente
- [x] `Decision` alias para `Status`
- [x] `CriteriaChecked` struct definida localmente
- [x] Tags JSON mantidas idênticas
- [x] Prefixos removidos (`domain.ReviewStatus` → `Status`, etc.)
- [x] `import "internal/domain"` removido de `models.go`

---

## Passo 2 — Arquivos Internos
- [x] `repository.go`: `*domain.Review` → `*Review`
- [x] `service.go`: `domain.Review` → `Review`, `domain.ReviewStatus` → `Status`, `domain.ValidationGate` → `ValidationGate`
- [x] `service.go`: `domain.ReviewCriteriaChecked` → `CriteriaChecked`, `domain.ReviewDecision` → `Decision`
- [x] `validation.go`: valida `Status` e `ValidationGate` locais
- [x] Testes do módulo atualizados

---

## Passo 3 — Adapters Temporários nos Consumidores
- [x] `internal/modules/orchestrator/models.go` — `ReviewManager` atualizado para usar review module types
- [x] `internal/bootstrap/services.go` — adapter atualizado para pass-through
- [x] `internal/modules/orchestrator/README.md` — review adicionado em Allowed Dependencies
- [x] `tests/architecture/module_boundaries_test.go` — importo orchestrator -> review adicionado

---

## Passo 4 — Validação
- [x] `go build ./...` passa
- [x] `go test ./tests/integration/...` passa
- [x] `./scripts/verify-contracts.sh` passa
- [x] Architecture tests pass (module_boundaries_test.go, contracts_sync_test.go)
- [x] `./scripts/safe-commit.sh "ADR-0022: migrate Review types to modules/review"` passa

---

## Status
- **Agente:** A09
- **Módulo:** review
- **Início:** 2025-01-XX
- **Término:** 2025-01-XX
- **Build:** ✅
- **Testes:** ✅
- **Commit:** ✅ (cascade/voc-est-migrando-o-m-dulo-review-a09-f5edc3)
- **PR:** ⚠️ (erro 403 - criar manualmente via GitHub UI)
