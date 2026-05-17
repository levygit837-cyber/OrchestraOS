# Checklist — ORCH-F05-R03-A08: Trigger Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisitos: A02 (run) 🟢, A03 (workunit) 🟢, A05 (agentsession) 🟢.

---

## Setup
- [ ] Branch `adr22-a08-trigger-types` criada e checkout feito
- [ ] Worktree `../orchestraos-a08-trigger` ativa
- [ ] `internal/modules/trigger/README.md` lido
- [ ] `internal/modules/trigger/CONTRACTS.md` lido
- [ ] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [ ] A02, A03, A05 estão no estado 🟢

---

## Passo 1 — models.go
- [ ] `Trigger` struct definida localmente
- [ ] `Status` enum definido localmente
- [ ] `Type` enum definido localmente
- [ ] `AnomalyType` enum definido localmente
- [ ] `ResolutionAction` enum definido localmente
- [ ] `ThresholdConfig` struct definida localmente
- [ ] Tags JSON mantidas idênticas
- [ ] Prefixos removidos (`domain.TriggerStatus` → `Status`, etc.)
- [ ] `import "internal/domain"` removido de `models.go`

---

## Passo 2 — Arquivos Internos
- [ ] `repository.go`: `*domain.Trigger` → `*Trigger`
- [ ] `service.go`: todos os `domain.*` → tipos locais
- [ ] `thresholds.go`: `domain.ThresholdConfig` → `ThresholdConfig`
- [ ] `detectors.go`: `domain.Trigger` → `Trigger`
- [ ] `validation.go`: `domain.TriggerType` → `Type`, `domain.TriggerStatus` → `Status`
- [ ] `fetch.go`: `*domain.Trigger` → `*Trigger`
- [ ] `events.go`: `domain.TriggerStatus` → `Status`
- [ ] Testes do módulo atualizados

---

## Passo 3 — Interfaces Cruzadas
- [ ] `RunReader` usa `*run.Run` (A02 concluído)
- [ ] `AgentSessionReader` usa `*agentsession.AgentSession` (A05 concluído)
- [ ] `WorkUnitReader` usa `*workunit.WorkUnit` (A03 concluído)

---

## Passo 4 — Adapters Temporários nos Consumidores
- [ ] `internal/modules/orchestrator/models.go` — `TriggerEvaluator` anotado
- [ ] `internal/bootstrap/services.go` — adapter atualizado

---

## Passo 5 — Validação
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Trigger types to modules/trigger"` passa

---

## Status
- **Agente:** A08
- **Módulo:** trigger
- **Início:** ___
- **Término:** ___
- **Build:** ___
- **Testes:** ___
- **Commit:** ___
