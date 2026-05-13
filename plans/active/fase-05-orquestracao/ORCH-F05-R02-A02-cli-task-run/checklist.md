# CHECKLIST — ORCH-F05-R02-A02: CLI `task run` + Testes E2E

**Agente:** Agente 2 (CLI + E2E)  
**Ferramenta:** Kimi-CLI  
**Início:** 2026-05-13  
**Status:** completed

---

## Checklist de Execução

- [x] 1. Ler documentação obrigatória: ADR 0020, `docs/implementation/roadmap.md` (Fase 5), `AGENTS.md`, `cmd/orchestraos/cmd/task.go`, `cmd/orchestraos/cmd/run.go`
- [x] 2. Analisar a interface contratual do OrchestratorService (definida pelo Orquestrador)
- [x] 3. Criar `cmd/orchestraos/cmd/task_run.go` — comando `task run`:
  - Flags: `--id` (task-id), `--runtime` (fake|gemini|codex_cli), `--planner`, `--max-steps`, `--timeout`
  - Validações de input
  - Chama `OrchestratorService.RunTask()` via bootstrap
  - Exibe progresso das work units no terminal
  - Exibe resultado final (status, runs criados, reviews pendentes)
- [x] 4. Registrar comando `task run` no `taskCmd` em `task.go`
- [x] 5. Refatorar `cmd/orchestraos/cmd/run.go` (`run start`):
  - Migrar geração inline de `AgentID` para usar `bootstrap.AgentService(db).FindOrCreate()`
  - Garantir que perfil do agente corresponde ao work unit
- [x] 6. Criar testes E2E `tests/integration/orchestrator_e2e_test.go`:
  - Stub interface test valida que OrchestratorService é acessível via bootstrap
  - Teste de refatoração valida que FindOrCreate é usado corretamente
  - Teste E2E full flow preparado (skip aguardando implementação do A01)
- [x] 7. Rodar `go test ./...` — verificar regressão em todos os pacotes
- [x] 8. Code review auto-crítico (flags, validações, progress output, test coverage)
- [x] 9. Correções pós-review
- [x] 10. Validar build: `go build ./...` sem erros + entrega final ao usuário

## Notas de Progresso

- Stub do OrchestratorService criado em `internal/modules/orchestrator/` para permitir compilação do CLI e testes enquanto Agente 1 (Windsurf) implementa o serviço real.
- `task run` implementado com todas as flags contratuais e fallback para `ORCHESTRAOS_PLANNER_STRATEGY`.
- `run start` refatorado: remove geração inline de `AgentID`, agora usa `AgentService.FindOrCreate(profile, runtimeType)` conforme ADR 0021.
- Testes de arquitetura passam (inclui contract.go, README.md, CONTRACTS.md, queries.go para o módulo orchestrator).
- Todos os testes do projeto passam sem regressão.
- Commit realizado via `./scripts/safe-commit.sh` na branch `feature/orch-f05-r02-a02-cli-task-run`.
