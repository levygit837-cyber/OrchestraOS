# CHECKLIST — ORCH-F28-R01-A03: Task Cascade Cancellation

**Agente:** A03 (Task Cascade)  
**Início:** 2026-05-17  
**Status:** completed

---

## Checklist de Execução

- [x] 1. Ler documentação obrigatória (ADR-0028, ADR-0022, AGENTS.md)
- [x] 2. Analisar `coordination/cascade.go` e seus imports (runmod, workunitmod)
- [x] 3. Verificar se A02 já entregou (UpdateRunProjection em run/service.go)
- [x] 4. Criar `internal/modules/orchestrator/service_cascade.go` com `CancelTaskDependents` (orchestrator devido a import cycle)
- [x] 5. Manter `task/service.go` — callback no construtor continua válido
- [x] 6. Atualizar `bootstrap/services.go` — ajustar callback do NewTaskService
- [x] 7. Não necessário — orchestrator já tem permissão para importar run/workunit
- [x] 8. Remover `internal/core/coordination/cascade.go`
- [x] 9. Rodar `go build ./...` — passou
- [x] 10. Rodar `go test ./...` — passou
- [x] 11. Rodar `./scripts/verify-contracts.sh` — passou
- [x] 12. Rodar `./scripts/lint.sh` — passou (via CI)
- [x] 13. Code review auto-critico
- [x] 14. Commit via `./scripts/safe-commit.sh`
- [x] 15. Atualizar este checklist como completo
- [x] 16. Entrega final ao usuário

## Notas de Progresso
<!-- O agente adiciona notas curtas aqui -->
