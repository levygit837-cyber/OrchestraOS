# CHECKLIST — ORCH-F28-R01-A04: AgentSession Timeout

**Agente:** A04 (AgentSession Timeout)  
**Início:** 2026-05-17  
**Status:** in_progress

---

## Checklist de Execução

- [ ] 1. Ler documentação obrigatória (ADR-0028, ADR-0022, AGENTS.md)
- [ ] 2. Analisar `coordination/agentsession_orchestrator.go`
- [ ] 3. Verificar se A02 já entregou (UpdateRunProjection em run/repository.go?)
- [ ] 4. Criar `internal/modules/agentsession/service_timeout.go`
- [ ] 5. Decidir: método do AgentSessionService ou função exportada?
- [ ] 6. Resolver import agentsession → run (allowedModuleImports ou callback)
- [ ] 7. Atualizar `tests/architecture/module_boundaries_test.go`
- [ ] 8. Remover `internal/core/coordination/agentsession_orchestrator.go`
- [ ] 9. Rodar `go build ./...` — deve passar
- [ ] 10. Rodar `go test ./...` — deve passar
- [ ] 11. Rodar `./scripts/verify-contracts.sh` — deve passar
- [ ] 12. Rodar `./scripts/lint.sh` — deve passar
- [ ] 13. Code review auto-critico
- [ ] 14. Commit via `./scripts/safe-commit.sh`
- [ ] 15. Atualizar este checklist como completo
- [ ] 16. Entrega final ao usuário

## Notas de Progresso
<!-- O agente adiciona notas curtas aqui -->
