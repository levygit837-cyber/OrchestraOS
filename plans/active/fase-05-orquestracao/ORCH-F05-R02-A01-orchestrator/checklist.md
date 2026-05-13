# CHECKLIST — ORCH-F05-R02-A01: OrchestratorService.RunTask()

**Agente:** Agente 1 (OrchestratorService)  
**Ferramenta:** Windsurf  
**Início:** _pendente_  
**Status:** in_progress

---

## Checklist de Execução

- [ ] 1. Ler documentação obrigatória: ADR 0020, ADR 0021, ADR 0023, `docs/implementation/roadmap.md` (Fase 5), `AGENTS.md`
- [ ] 2. Analisar código existente: `internal/bootstrap/services.go`, `internal/core/orchestration/runtime_relay.go`, `internal/modules/*/service.go`, `cmd/orchestraos/cmd/run.go`
- [ ] 3. Criar estrutura do módulo: `internal/modules/orchestrator/doc.go`, `contract.go`, `models.go`
- [ ] 4. Implementar `internal/modules/orchestrator/service.go`:
  - `OrchestratorService` struct com todas as dependências injetadas
  - `RunTask(ctx, taskID, options)` — fluxo completo
  - Ordenação topológica do DAG de work units
  - Loop sequencial de execução (1 WU por vez)
  - Integração com RuntimeEventRelay durante cada run
  - Criação automática de Review quando work unit tem ValidationGate
  - Avaliação de triggers pós-run via TriggerService
- [ ] 5. Implementar `internal/modules/orchestrator/validation.go` — validar options (runtime type, planner strategy)
- [ ] 6. Implementar `internal/modules/orchestrator/events.go` — eventos do orquestrador
- [ ] 7. Atualizar `internal/bootstrap/services.go` — adicionar `OrchestratorService(db)` factory
- [ ] 8. Criar testes de integração `tests/integration/orchestrator_service_test.go`:
  - Teste com FakeRuntime: task com 2 work units em sequência
  - Verificar ordenação topológica
  - Verificar que AgentSession.Checkpoint é atualizado
  - Verificar transição Run created → running → completed
  - Verificar que Task fica completed ao final
- [ ] 9. Rodar `go test ./...` — verificar regressão em todos os pacotes
- [ ] 10. Code review auto-crítico (transações, goroutine leaks, timeout handling, cleanup)
- [ ] 11. Correções pós-review
- [ ] 12. Validar build: `go build ./...` sem erros
- [ ] 13. Atualizar este checklist como completo + entrega final ao usuário

## Notas de Progresso
<!-- Adicione notas curtas a cada iteração significativa -->
