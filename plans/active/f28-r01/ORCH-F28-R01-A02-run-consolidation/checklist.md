# CHECKLIST — ORCH-F28-R01-A02: Run Module Consolidation

**Agente:** A02 (Run Consolidation)  
**Início:** 2026-05-17  
**Status:** in_progress

---

## Checklist de Execução

- [ ] 1. Ler documentação obrigatória (ADR-0028, ADR-0022, AGENTS.md)
- [ ] 2. Analisar código existente: `coordination/helpers.go`, `run_workunit_sync.go`, `runtime_relay.go`, `queries.go`
- [ ] 3. Analisar código destino: `run/repository.go`, `run/queries.go`, `run/service.go`, `bootstrap/services.go`, `orchestrator/service.go`
- [ ] 4. Mover `UpdateRunProjection` para `run/repository.go` como método `UpdateProjection`
- [ ] 5. Mover SQL de runs de `coordination/queries.go` para `run/queries.go`
- [ ] 6. Criar `run/service_workunit.go` com `TransitionRunWithWorkUnit`
- [ ] 7. Criar `run/service_relay.go` com `RuntimeEventRelay`, interfaces, helpers
- [ ] 8. Resolver imports cíclicos/proibidos (run → agentsession se necessário)
- [ ] 9. Atualizar `bootstrap/services.go` (remover import coordination, usar runmod.*)
- [ ] 10. Atualizar `orchestrator/service.go` (tipo RuntimeEventRelay, RelayConfig)
- [ ] 11. Atualizar `cmd/orchestraos/cmd/run.go` (RelayConfig)
- [ ] 12. Mover `tests/unit/core/coordination/runtime_relay_test.go` → `tests/unit/modules/run/service_relay_test.go`
- [ ] 13. Remover arquivos mortos do `coordination/`: helpers.go, run_workunit_sync.go, runtime_relay.go
- [ ] 14. Limpar `coordination/queries.go` (remover QueryRunUpdateStatus)
- [ ] 15. Rodar `go build ./...` — deve passar
- [ ] 16. Rodar `go test ./...` — deve passar
- [ ] 17. Rodar `./scripts/verify-contracts.sh` — deve passar
- [ ] 18. Rodar `./scripts/lint.sh` — deve passar
- [ ] 19. Code review auto-critico (responder perguntas do plan.md)
- [ ] 20. Commit via `./scripts/safe-commit.sh`
- [ ] 21. Atualizar este checklist como completo
- [ ] 22. Entrega final ao usuário

## Notas de Progresso
<!-- O agente adiciona notas curtas aqui -->
