# Checklist — ORCH-F05-R03-A07: Prompt Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- **Política:** Ver ADR-0026 (Module Import Policy). `internal/domain` só pode ser usado para `EventEnvelope`, `EventPriority`, e tipos genéricos. NUNCA para entity structs.
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisitos: A01 (task) 🟢, A02 (run) 🟢, A03 (workunit) 🟢, A05 (agentsession) 🟢.

---

## Setup
- [x] Branch `adr22-a07-prompt-types` criada e checkout feito
- [x] Worktree `../orchestraos-a07-prompt` ativa
- [x] `internal/modules/prompt/README.md` lido
- [x] `internal/modules/prompt/CONTRACTS.md` lido
- [x] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [x] A01, A02, A03, A05 estão no estado 🟢

---

## Passo 1 — models.go (novo arquivo)
- [x] `PromptSnapshot` struct definida localmente
- [x] `PromptFragment` struct definida localmente
- [x] `PromptFragmentRef` struct definida localmente
- [x] `ToolsetSnapshot` struct definida localmente
- [x] `ToolsetTool` struct definida localmente
- [x] Tags JSON mantidas idênticas
- [x] `import "internal/domain"` removido (ou nunca adicionado)

---

## Passo 2 — Arquivos Internos
- [x] `repository.go`: `domain.PromptFragment` → `PromptFragment`, `domain.PromptSnapshot` → `PromptSnapshot`
- [x] `repository_snapshot.go`: `domain.PromptSnapshot` → `PromptSnapshot`, `domain.ToolsetSnapshot` → `ToolsetSnapshot`
- [x] `service.go`: `PrepareAndPersistInput` usa tipos locais de task/run/workunit/agentsession
- [x] `service.go`: `PreparedRunPrompt` usa `*PromptSnapshot`, `*ToolsetSnapshot`
- [x] `service.go`: funções `*ToDomain()` removidas
- [x] `types.go`: verificado (já local)
- [x] `composer.go`: verificado (já local)
- [x] Testes do módulo atualizados (nenhum teste existente no módulo)

---

## Passo 3 — Adapters Temporários nos Consumidores
- [ ] `internal/modules/orchestrator/models.go` — `PreparedPrompt` anotado (fora do escopo: não tocar em orchestrator)
- [ ] `internal/core/coordination/prompt_orchestrator.go` — adapters criados (fora do escopo: não tocar em coordination)
- [ ] `internal/bootstrap/services.go` — adapter atualizado (fora do escopo: não tocar em bootstrap)

---

## Passo 4 — Validação
- [ ] `go build ./...` passa (falha em consumers externos fora do escopo)
- [x] `go test ./internal/modules/prompt/...` passa
- [ ] `./scripts/verify-contracts.sh` passa (bloqueado por build externo)
- [ ] `./scripts/lint.sh` passa (bloqueado por build externo)
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Prompt types to modules/prompt"` passa (bloqueado por build externo)

---

## Status
- **Agente:** A07
- **Módulo:** prompt
- **Início:** ___
- **Término:** ___
- **Build:** ___
- **Testes:** ___
- **Commit:** ___
