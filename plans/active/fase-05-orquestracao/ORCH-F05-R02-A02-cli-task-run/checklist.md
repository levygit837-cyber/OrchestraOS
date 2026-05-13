# CHECKLIST — ORCH-F05-R02-A02: CLI `task run` + Testes E2E

**Agente:** Agente 2 (CLI + E2E)  
**Ferramenta:** Kimi-CLI  
**Início:** _pendente_  
**Status:** in_progress

---

## Checklist de Execução

- [ ] 1. Ler documentação obrigatória: ADR 0020, `docs/implementation/roadmap.md` (Fase 5), `AGENTS.md`, `cmd/orchestraos/cmd/task.go`, `cmd/orchestraos/cmd/run.go`
- [ ] 2. Analisar a interface contratual do OrchestratorService (definida pelo Orquestrador)
- [ ] 3. Criar `cmd/orchestraos/cmd/task_run.go` — comando `task run`:
  - Flags: `--id` (task-id), `--runtime` (fake|gemini|codex_cli), `--planner`, `--max-steps`, `--timeout`
  - Validações de input
  - Chama `OrchestratorService.RunTask()` via bootstrap
  - Exibe progresso das work units no terminal
  - Exibe resultado final (status, runs criados, reviews pendentes)
- [ ] 4. Registrar comando `task run` no `taskCmd` em `task.go`
- [ ] 5. Refatorar `cmd/orchestraos/cmd/run.go` (`run start`):
  - Migrar geração inline de `AgentID` para usar `bootstrap.AgentService(db).FindOrCreate()`
  - Garantir que perfil do agente corresponde ao work unit
- [ ] 6. Criar testes E2E `tests/integration/orchestrator_e2e_test.go`:
  - Fluxo completo: `task create` → `task graph create` → `task run` (via OrchestratorService)
  - FakeRuntime com 2 work units sequenciais
  - Verificar que task termina como `completed`
  - Verificar que runs foram criados para cada work unit
  - Verificar que agentes foram registrados com perfil correto
- [ ] 7. Rodar `go test ./...` — verificar regressão em todos os pacotes
- [ ] 8. Code review auto-crítico (flags, validações, progress output, test coverage)
- [ ] 9. Correções pós-review
- [ ] 10. Validar build: `go build ./...` sem erros + entrega final ao usuário

## Notas de Progresso
<!-- Adicione notas curtas a cada iteração significativa -->
