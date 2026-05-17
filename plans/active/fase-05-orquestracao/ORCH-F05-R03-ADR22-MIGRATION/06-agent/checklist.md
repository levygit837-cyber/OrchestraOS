# Checklist — ORCH-F05-R03-A06: Agent Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisitos: Nenhum. Agent é independente.

---

## Setup
- [ ] Branch `adr22-a06-agent-types` criada e checkout feito
- [ ] Worktree `../orchestraos-a06-agent` ativa
- [ ] `internal/modules/agent/README.md` lido
- [ ] `internal/modules/agent/CONTRACTS.md` lido
- [ ] `docs/adr/0022-llm-optimized-module-architecture.md` lido

---

## Passo 1 — models.go
- [ ] `Agent` struct definida localmente
- [ ] `RuntimeType` enum definido localmente
- [ ] `AgentStatus` enum definido localmente
- [ ] Tags JSON mantidas idênticas
- [ ] `ToDomainRuntimeType()` removido
- [ ] `FromDomainRuntimeType()` removido
- [ ] `import "internal/domain"` removido de `models.go`

---

## Passo 2 — Arquivos Internos
- [ ] `repository.go`: `*domain.Agent` → `*Agent`
- [ ] `service.go`: `domain.Agent` → `Agent`, `domain.AgentRuntimeType` → `RuntimeType`
- [ ] `validation.go`: valida `RuntimeType` local
- [ ] `contract.go`: `AgentReader` retorna `*Agent`
- [ ] `runtime.go`: verificado (não usa domain)
- [ ] `fake_runtime.go`: verificado (não usa domain)
- [ ] `gemini_runtime.go`: verificado (não usa domain)
- [ ] Testes do módulo atualizados

---

## Passo 3 — Adapters Temporários nos Consumidores
- [ ] `internal/modules/orchestrator/models.go` — `AgentManager` anotado
- [ ] `internal/modules/agentsession/service.go` — `AgentReader` anotado
- [ ] `internal/bootstrap/services.go` — adapter atualizado

---

## Passo 4 — Validação
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Agent types to modules/agent"` passa

---

## Status
- **Agente:** A06
- **Módulo:** agent
- **Início:** ___
- **Término:** ___
- **Build:** ___
- **Testes:** ___
- **Commit:** ___
