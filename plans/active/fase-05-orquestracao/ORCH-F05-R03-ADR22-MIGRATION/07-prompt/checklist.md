# Checklist — ORCH-F05-R03-A07: Prompt Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- **Política:** Ver ADR-0026 (Module Import Policy). `internal/domain` só pode ser usado para `EventEnvelope`, `EventPriority`, e tipos genéricos. NUNCA para entity structs.
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisitos: A01 (task) 🟢, A02 (run) 🟢, A03 (workunit) 🟢, A05 (agentsession) 🟢.

---

## Setup
- [ ] Branch `adr22-a07-prompt-types` criada e checkout feito
- [ ] Worktree `../orchestraos-a07-prompt` ativa
- [ ] `internal/modules/prompt/README.md` lido
- [ ] `internal/modules/prompt/CONTRACTS.md` lido
- [ ] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [ ] A01, A02, A03, A05 estão no estado 🟢

---

## Passo 1 — models.go (novo arquivo)
- [ ] `PromptSnapshot` struct definida localmente
- [ ] `PromptFragment` struct definida localmente
- [ ] `PromptFragmentRef` struct definida localmente
- [ ] `ToolsetSnapshot` struct definida localmente
- [ ] `ToolsetTool` struct definida localmente
- [ ] Tags JSON mantidas idênticas
- [ ] `import "internal/domain"` removido (ou nunca adicionado)

---

## Passo 2 — Arquivos Internos
- [ ] `repository.go`: `domain.PromptFragment` → `PromptFragment`, `domain.PromptSnapshot` → `PromptSnapshot`
- [ ] `repository_snapshot.go`: `domain.PromptSnapshot` → `PromptSnapshot`, `domain.ToolsetSnapshot` → `ToolsetSnapshot`
- [ ] `service.go`: `PrepareAndPersistInput` usa tipos locais de task/run/workunit/agentsession
- [ ] `service.go`: `PreparedRunPrompt` usa `*PromptSnapshot`, `*ToolsetSnapshot`
- [ ] `service.go`: funções `*ToDomain()` removidas
- [ ] `types.go`: verificado (já local)
- [ ] `composer.go`: verificado (já local)
- [ ] Testes do módulo atualizados

---

## Passo 3 — Adapters Temporários nos Consumidores
- [ ] `internal/modules/orchestrator/models.go` — `PreparedPrompt` anotado
- [ ] `internal/core/coordination/prompt_orchestrator.go` — adapters criados
- [ ] `internal/bootstrap/services.go` — adapter atualizado

---

## Passo 4 — Validação
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Prompt types to modules/prompt"` passa

---

## Status
- **Agente:** A07
- **Módulo:** prompt
- **Início:** ___
- **Término:** ___
- **Build:** ___
- **Testes:** ___
- **Commit:** ___
