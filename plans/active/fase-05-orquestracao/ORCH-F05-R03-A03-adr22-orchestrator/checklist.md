# Checklist — ORCH-F05-R03-A03: ADR-0022 Orchestrator Migration

## Instruções

- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Cada fase deve terminar com `./scripts/safe-commit.sh` passando.

---

### Pré-requisitos

- [ ] Branch `feat/adr22-orchestrator-migration` criada a partir de `master` atualizado.
- [ ] `go build ./...` passa no master antes de iniciar.
- [ ] `go test ./...` passa no master antes de iniciar.

### Fase 1 — Coordination Layer

- [ ] `internal/core/coordination/helpers.go`: `UpdateRunProjection` usa `runmod.Status` e `runmod.Result`
- [ ] `internal/core/coordination/cascade.go`: usa `runmod.StatusCancelled`, `runmod.ResultForStatus()`, sem adapter `domainResult`
- [ ] `internal/core/coordination/run_workunit_sync.go`: `TransitionRunWithWorkUnit` recebe `*runmod.Run` e `runmod.Status`
- [ ] `internal/core/coordination/agentsession_orchestrator.go`: `AgentSessionTimeout` recebe `*agentsessionmod.AgentSession`
- [ ] `internal/core/coordination/runtime_relay.go`: retorna `runmod.Status`, TODOs removidos
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate coordination layer to module types"` passa

### Fase 2 — Orchestrator Module

- [ ] `internal/modules/orchestrator/models.go`: `TaskServiceReader` retorna `*taskmod.Task`
- [ ] `internal/modules/orchestrator/models.go`: `RunLifecycleManager` retorna `*runmod.Run`
- [ ] `internal/modules/orchestrator/models.go`: `AgentManager` usa `agentmod.RuntimeType` e retorna `*agentmod.Agent`
- [ ] `internal/modules/orchestrator/models.go`: `SessionManager` retorna `*agentsessionmod.AgentSession`
- [ ] `internal/modules/orchestrator/models.go`: `WorkUnitLister` retorna `[]workunitmod.WorkUnit`
- [ ] `internal/modules/orchestrator/models.go`: import de `internal/domain` removido (exceto `EventEnvelope` se necessário)
- [ ] `internal/modules/orchestrator/service.go`: `executeWorkUnit` usa `*workunitmod.WorkUnit` e `*taskmod.Task`
- [ ] `internal/modules/orchestrator/service.go`: `topologicalSort` usa `[]workunitmod.WorkUnit`
- [ ] `internal/modules/orchestrator/service.go`: `listWorkUnitsByGraph` retorna `[]workunitmod.WorkUnit`
- [ ] `internal/modules/orchestrator/service.go`: `result.Success` compara com `runmod.StatusCompleted`
- [ ] `internal/modules/orchestrator/validation.go`: `ConvertRuntimeType` retorna `agentmod.RuntimeType`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate orchestrator module to module types"` passa

### Fase 3 — Bootstrap Cleanup

- [ ] `internal/bootstrap/services.go`: `taskToDomain` removido
- [ ] `internal/bootstrap/services.go`: `runToDomain` removido
- [ ] `internal/bootstrap/services.go`: `workunitToDomain` removido
- [ ] `internal/bootstrap/services.go`: `agentToDomain` removido
- [ ] `internal/bootstrap/services.go`: `agentSessionToDomain` removido
- [ ] `internal/bootstrap/services.go`: `taskAdapter` removido
- [ ] `internal/bootstrap/services.go`: `runAdapter` removido
- [ ] `internal/bootstrap/services.go`: `sessionAdapter` removido
- [ ] `internal/bootstrap/services.go`: `agentManagerAdapter` removido
- [ ] `internal/bootstrap/services.go`: `wuListerAdapter` removido
- [ ] `internal/bootstrap/services.go`: `OrchestratorService` faz wire direto sem adapters de entidade
- [ ] `internal/bootstrap/services.go`: `RunService` factory passa `coordination.TransitionRunWithWorkUnit` direto
- [ ] `internal/bootstrap/services.go`: nenhum TODO[ADR-0022] restante
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: remove bootstrap adapters and simplify wiring"` passa

### Fase 4 — Architecture Tests

- [ ] `tests/architecture/module_boundaries_test.go`: `orchestrator` inclui `task`, `run`, `agent`, `agentsession` em `allowedModuleImports`
- [ ] `go test ./tests/architecture/...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: update architecture tests for orchestrator imports"` passa

### Fase 5 — Validação Final

- [ ] `grep -rn "TODO\[ADR-0022\]" internal/` retorna ZERO resultados
- [ ] `go build ./...` passa (todos os pacotes)
- [ ] `go test ./...` passa (todos os pacotes)
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] Criar Pull Request para `master`
- [ ] CI passa (GitHub Actions)
