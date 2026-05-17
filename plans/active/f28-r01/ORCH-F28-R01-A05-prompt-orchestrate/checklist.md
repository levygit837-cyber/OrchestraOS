# CHECKLIST — ORCH-F28-R01-A05: Prompt Orchestrate

**Agente:** A05 (Prompt Orchestrate)  
**Início:** 2026-05-17  
**Status:** in_progress

---

## Checklist de Execução

- [x] 1. Ler documentação obrigatória (ADR-0028, ADR-0022, AGENTS.md)
- [x] 2. Analisar `coordination/prompt_orchestrator.go`
- [x] 3. Analisar `prompt/service.go` (verificar se PromptOrchestrator pode ser método do PromptService)
- [x] 4. Criar `internal/modules/prompt/service_orchestrate.go`
- [x] 5. Atualizar `bootstrap/services.go` (PromptOrchestrator instantiation)
- [x] 6. Atualizar `cmd/orchestraos/cmd/run.go`
- [x] 7. Atualizar `tests/integration/e2e_orchestration_test.go`
- [x] 8. Atualizar `tests/integration/services_test.go`
- [x] 9. Remover `internal/core/coordination/prompt_orchestrator.go`
- [x] 10. Rodar `go build ./...` — deve passar
- [x] 11. Rodar `go test ./...` — deve passar
- [x] 12. Rodar `./scripts/verify-contracts.sh` — deve passar
- [x] 13. Rodar `./scripts/lint.sh` — deve passar
- [x] 14. Code review auto-critico
- [ ] 15. Commit via `./scripts/safe-commit.sh`
- [ ] 16. Atualizar este checklist como completo
- [ ] 17. Entrega final ao usuário

## Notas de Progresso
<!-- O agente adiciona notas curtas aqui -->
