# Checklist — ORCH-F05-R03-A09: Review Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- **Política:** Ver ADR-0026 (Module Import Policy). `internal/domain` só pode ser usado para `EventEnvelope`, `EventPriority`, e tipos genéricos. NUNCA para entity structs.
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisitos: A01 (task) 🟢, A02 (run) 🟢, A03 (workunit) 🟢.

---

## Setup
- [ ] Branch `adr22-a09-review-types` criada e checkout feito
- [ ] Worktree `../orchestraos-a09-review` ativa
- [ ] `internal/modules/review/README.md` lido
- [ ] `internal/modules/review/CONTRACTS.md` lido
- [ ] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [ ] A01, A02, A03 estão no estado 🟢

---

## Passo 1 — models.go
- [ ] `Review` struct definida localmente
- [ ] `Status` enum definido localmente
- [ ] `ValidationGate` enum definido localmente
- [ ] `Decision` alias para `Status`
- [ ] `CriteriaChecked` struct definida localmente
- [ ] Tags JSON mantidas idênticas
- [ ] Prefixos removidos (`domain.ReviewStatus` → `Status`, etc.)
- [ ] `import "internal/domain"` removido de `models.go`

---

## Passo 2 — Arquivos Internos
- [ ] `repository.go`: `*domain.Review` → `*Review`
- [ ] `service.go`: `domain.Review` → `Review`, `domain.ReviewStatus` → `Status`, `domain.ValidationGate` → `ValidationGate`
- [ ] `service.go`: `domain.ReviewCriteriaChecked` → `CriteriaChecked`, `domain.ReviewDecision` → `Decision`
- [ ] `validation.go`: valida `Status` e `ValidationGate` locais
- [ ] Testes do módulo atualizados

---

## Passo 3 — Adapters Temporários nos Consumidores
- [ ] `internal/modules/orchestrator/models.go` — `ReviewManager` anotado
- [ ] `internal/bootstrap/services.go` — adapter atualizado

---

## Passo 4 — Validação
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Review types to modules/review"` passa

---

## Status
- **Agente:** A09
- **Módulo:** review
- **Início:** ___
- **Término:** ___
- **Build:** ___
- **Testes:** ___
- **Commit:** ___
