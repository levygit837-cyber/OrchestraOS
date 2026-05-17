# Checklist — ORCH-F05-R03-A05: AgentSession Module Type Migration

## Instruções
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- **Política:** Ver ADR-0026 (Module Import Policy). `internal/domain` só pode ser usado para `EventEnvelope`, `EventPriority`, e tipos genéricos. NUNCA para entity structs.
- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Pré-requisitos: A02 (run) 🟢 e A06 (agent) 🟢.

---

## Setup
- [ ] Branch `adr22-a05-agentsession-types` criada e checkout feito
- [ ] Worktree `../orchestraos-a05-agentsession` ativa
- [ ] `internal/modules/agentsession/README.md` lido
- [ ] `internal/modules/agentsession/CONTRACTS.md` lido
- [ ] `docs/adr/0022-llm-optimized-module-architecture.md` lido
- [ ] A02 (run) e A06 (agent) estão no estado 🟢

---

## Passo 1 — models.go
- [ ] `AgentSession` struct definida localmente
- [ ] `Status` enum definido localmente
- [ ] Tags JSON mantidas idênticas
- [ ] Prefixo removido (`domain.AgentSessionStatus` → `Status`)
- [ ] `import "internal/domain"` removido de `models.go`

---

## Passo 2 — Arquivos Internos
- [ ] `repository.go`: `*domain.AgentSession` → `*AgentSession`
- [ ] `service.go`: `domain.AgentSession` → `AgentSession`, `domain.AgentSessionStatus` → `Status`, `domain.Agent` → `agent.Agent`
- [ ] `service_checkpoint.go`: `domain.AgentSession` → `AgentSession`
- [ ] `service_heartbeat.go`: `domain.AgentSession` → `AgentSession`
- [ ] `fetch.go`: `*domain.AgentSession` → `*AgentSession`
- [ ] `events.go`: `domain.AgentSessionStatus` → `Status`
- [ ] `checkpoint_policy.go`: `domain.AgentSession` → `AgentSession`
- [ ] Testes do módulo atualizados

---

## Passo 3 — Interfaces Cruzadas
- [ ] `AgentReader` interface usa `*agent.Agent` (A06 concluído)
- [ ] Verificar `Run` references — se A02 concluído, usar `*run.Run`; senão, adapter

---

## Passo 4 — Adapters Temporários nos Consumidores
- [ ] `internal/modules/orchestrator/models.go` — `SessionManager` anotado
- [ ] `internal/core/coordination/prompt_orchestrator.go` — adapter criado
- [ ] `internal/modules/trigger/service.go` — `AgentSessionReader` anotado
- [ ] `internal/bootstrap/services.go` — adapter atualizado

---

## Passo 5 — Validação
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate AgentSession types to modules/agentsession"` passa

---

## Status
- **Agente:** A05
- **Módulo:** agentsession
- **Início:** ___
- **Término:** ___
- **Build:** ___
- **Testes:** ___
- **Commit:** ___
