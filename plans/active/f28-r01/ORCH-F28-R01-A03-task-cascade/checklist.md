# CHECKLIST — ORCH-F28-R01-A03: Task Cascade Cancellation

**Agente:** A03 (Task Cascade)  
**Início:** 2026-05-17  
**Status:** in_progress

---

## Checklist de Execução

- [ ] 1. Ler documentação obrigatória (ADR-0028, ADR-0022, AGENTS.md)
- [ ] 2. Analisar `coordination/cascade.go` e seus imports (runmod, workunitmod)
- [ ] 3. Verificar se A02 já entregou (UpdateRunProjection em run/repository.go?)
- [ ] 4. Criar `internal/modules/task/service_cascade.go` com `CancelTaskDependents`
- [ ] 5. Ajustar `task/service.go` — simplificar construtor se CancelTaskDependents virar método
- [ ] 6. Atualizar `bootstrap/services.go` — ajustar callback do NewTaskService
- [ ] 7. Atualizar `tests/architecture/module_boundaries_test.go` — remover task de leafModules, adicionar allowedModuleImports
- [ ] 8. Remover `internal/core/coordination/cascade.go`
- [ ] 9. Rodar `go build ./...` — deve passar
- [ ] 10. Rodar `go test ./...` — deve passar
- [ ] 11. Rodar `./scripts/verify-contracts.sh` — deve passar
- [ ] 12. Rodar `./scripts/lint.sh` — deve passar
- [ ] 13. Code review auto-critico (responder perguntas do plan.md)
- [ ] 14. Commit via `./scripts/safe-commit.sh`
- [ ] 15. Atualizar este checklist como completo
- [ ] 16. Entrega final ao usuário

## Notas de Progresso
<!-- O agente adiciona notas curtas aqui -->
