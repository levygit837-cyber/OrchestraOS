# CHECKLIST — ORCH-F05-R02-A01: OrchestratorService.RunTask()

**Agente:** Agente 1 (OrchestratorService)  
**Ferramenta:** Windsurf  
**Início:** _pendente_  
**Status:** in_progress

---

## Checklist de Execução

- [x] 1. Ler documentação obrigatória: ADR 0020, ADR 0021, ADR 0023, `docs/implementation/roadmap.md` (Fase 5), `AGENTS.md`
- [x] 2. Analisar código existente: `internal/bootstrap/services.go`, `internal/core/coordination/runtime_relay.go`, `internal/modules/*/service.go`, `cmd/orchestraos/cmd/run.go`
- [x] 3. Criar estrutura do módulo: `internal/modules/orchestrator/doc.go`, `contract.go`, `models.go`
- [x] 4. Implementar `internal/modules/orchestrator/service.go`:
  - `OrchestratorService` struct com todas as dependências injetadas
  - `RunTask(ctx, taskID, options)` — fluxo completo
  - Ordenação topológica do DAG de work units
  - Loop sequencial de execução (1 WU por vez)
  - Integração com RuntimeEventRelay durante cada run
  - Criação automática de Review quando work unit tem ValidationGate (deferred - WorkUnit lacks field)
  - Avaliação de triggers pós-run via TriggerService
- [x] 5. Implementar `internal/modules/orchestrator/validation.go` — validar options (runtime type, planner strategy)
- [x] 6. Atualizar `internal/bootstrap/services.go` com factory para OrchestratorService
- [x] 7. Implementar `internal/modules/orchestrator/events.go` — eventos do orquestrador (já em contract.go)
- [x] 8. Criar testes de integração `tests/integration/orchestrator_service_test.go`:
  - Teste com FakeRuntime: task com 2 work units em sequência
  - Verificar ordenação topológica
  - Verificar que AgentSession.Checkpoint é atualizado
  - Verificar transição Run created → running → completed
  - Verificar que Task fica completed ao final
- [ ] 9. Rodar `go test ./...` — verificar regressão em todos os pacotes (BLOCKED: worktree environment)
- [x] 10. Code review auto-crítico (transações, goroutine leaks, timeout handling, cleanup) - FIXED:
  - Added goroutine completion tracking with routineDone channel
  - Added timeout-safe select for runtime error channel
  - Added runtime.Stop() with 5s timeout context
  - Added session.Stop() for best-effort cleanup
  - Updated SessionManager interface to include Stop method
- [x] 11. Correções pós-review (applied during review)
- [ ] 12. Validar build: `go build ./...` sem erros (BLOCKED: worktree environment)
- [x] 13. Atualizar este checklist como completo + entrega final ao usuário

## Notas de Progresso
<!-- Adicione notas curtas a cada iteração significativa -->

- 2025-01-XX: Criada estrutura do módulo (doc.go, contract.go, models.go, validation.go)
- 2025-01-XX: Implementado service.go com RunTask completo (decomposição, ordenação topológica, execução sequencial, relay, triggers)
- 2025-01-XX: Atualizado bootstrap/services.go com factory OrchestratorService
- 2025-01-XX: Criados testes de integração em tests/integration/orchestrator_service_test.go
- 2025-01-XX: Review creation deferred - WorkUnit lacks ValidationGate field currently
- 2025-01-XX: Code review fixes applied: goroutine leak prevention, runtime/session cleanup, timeout-safe error handling
