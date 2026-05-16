# Checklist — ORCH-F05-R03-A01: ADR-0022 Type Migration

## Instruções

- Marque `[x]` apenas quando o item estiver 100% concluído e verificado.
- Se um item falhar em `go build` ou `go test`, volte para `[ ]` e corrija antes de prosseguir.
- Cada sessão deve terminar com `./scripts/safe-commit.sh` passando.

---

## Sessão 1 — Task Module

- [ ] `internal/modules/task/models.go` define `Task`, `Status`, `Priority`, `RiskLevel` localmente (sem import de `domain`)
- [ ] `internal/modules/task/repository.go` usa `*Task`, `[]Task`, `Task` (não `domain.Task`)
- [ ] `internal/modules/task/service.go` usa `Status`, `Priority`, `RiskLevel` locais
- [ ] `internal/modules/task/fetch.go` retorna `*Task`
- [ ] `internal/modules/task/events.go` usa `Status` local
- [ ] `internal/modules/task/validation_test.go` usa `Priority`, `RiskLevel` locais
- [ ] `internal/core/orchestration/prompt_orchestrator.go` tem adapter `toDomainTask()`
- [ ] `cmd/orchestraos/cmd/task.go` usa `task.Priority` / `task.RiskLevel`
- [ ] `tests/integration/*` substituíram `domain.TaskStatusX` → `task.StatusX`, `domain.PriorityP2` → `task.PriorityP2`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Task types to modules/task"` passa

---

## Sessão 2 — Run Module

- [ ] `internal/modules/run/models.go` define `Run`, `Status`, `Result` localmente
- [ ] `internal/modules/run/repository.go` usa `*Run`, `[]Run`
- [ ] `internal/modules/run/service.go` usa `Status`, `Result` locais
- [ ] `internal/modules/run/service_retry.go` usa `Status`, `Result` locais
- [ ] `internal/modules/run/fetch.go` retorna `*Run`
- [ ] `internal/modules/run/events.go` usa `Status` e `Result` locais
- [ ] `internal/core/orchestration/agentsession_orchestrator.go` usa `runmod.StatusX`
- [ ] `internal/core/orchestration/cascade.go` usa `runmod.StatusCancelled`, `runmod.ResultForStatus()`
- [ ] `internal/core/orchestration/helpers.go` usa `runmod.Status`, `runmod.Result`
- [ ] `tests/integration/*` substituíram `domain.RunStatusX` → `run.StatusX`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Run types to modules/run"` passa

---

## Sessão 3 — WorkUnit Module

- [ ] `internal/modules/workunit/models.go` define `WorkUnit`, `Status` localmente
- [ ] `internal/modules/workunit/repository.go` usa `*WorkUnit`, `[]WorkUnit`
- [ ] `internal/modules/workunit/service.go` usa `Status` local
- [ ] `internal/modules/workunit/service_create.go` usa `WorkUnit` local
- [ ] `internal/modules/workunit/fetch.go` retorna `*WorkUnit`
- [ ] `internal/modules/workunit/validation.go` usa `Status` local
- [ ] `internal/modules/orchestrator/service.go` usa `workunitmod.WorkUnit`
- [ ] `internal/modules/orchestrator/models.go` interface `WorkUnitLister` usa `[]workunitmod.WorkUnit`
- [ ] `tests/integration/*` substituíram `domain.WorkUnitStatusX` → `workunit.StatusX`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate WorkUnit types to modules/workunit"` passa

---

## Sessão 4 — TaskGraph Module

- [ ] `internal/modules/taskgraph/models.go` define `TaskGraph`, `Status` localmente
- [ ] `internal/modules/taskgraph/repository.go` usa `*TaskGraph`, `[]TaskGraph`
- [ ] `internal/modules/taskgraph/service.go` usa `Status` local
- [ ] `internal/modules/orchestrator/models.go` interface `TaskGraphManager` usa `*taskgraphmod.TaskGraph`
- [ ] `tests/integration/*` ajustados
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate TaskGraph types to modules/taskgraph"` passa

---

## Sessão 5 — AgentSession Module

- [ ] `internal/modules/agentsession/models.go` define `AgentSession`, `Status` localmente
- [ ] `internal/modules/agentsession/repository.go` usa `*AgentSession`
- [ ] `internal/modules/agentsession/service.go` usa `Status` local
- [ ] `internal/modules/agentsession/service_checkpoint.go` usa `Status` local
- [ ] `internal/modules/agentsession/service_heartbeat.go` usa `Status` local
- [ ] `internal/modules/agentsession/checkpoint_policy.go` usa `AgentSession` local
- [ ] `internal/core/orchestration/prompt_orchestrator.go` tem adapter `toDomainAgentSession()`
- [ ] `internal/modules/orchestrator/models.go` interface `SessionManager` usa `*agentsessionmod.AgentSession`
- [ ] `tests/integration/*` ajustados
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate AgentSession types to modules/agentsession"` passa

---

## Sessão 6 — Agent Module

- [ ] `internal/modules/agent/models.go` define `Agent`, `RuntimeType` localmente
- [ ] `internal/modules/agent/repository.go` usa `*Agent`
- [ ] `internal/modules/agent/service.go` usa `RuntimeType` local
- [ ] `internal/modules/orchestrator/models.go` interface `AgentManager` usa `*agentmod.Agent`, `agentmod.RuntimeType`
- [ ] `tests/integration/*` ajustados
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Agent types to modules/agent"` passa

---

## Sessão 7 — Prompt Module

- [ ] Verificar que `internal/modules/prompt/types.go` NÃO tem aliases para `domain.PromptSnapshot` / `domain.ToolsetSnapshot`
- [ ] Se houver aliases, convertê-los para structs locais
- [ ] `internal/core/orchestration/prompt_orchestrator.go`: adapters para prompt types se necessário
- [ ] `tests/integration/*` ajustados
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: decouple Prompt types from domain"` passa

---

## Sessão 8 — Trigger Module

- [ ] `internal/modules/trigger/models.go` define `Trigger`, `TriggerType`, `TriggerStatus`, `AnomalyType`, `ResolutionAction`, `ThresholdConfig` localmente
- [ ] `internal/modules/trigger/repository.go` usa tipos locais
- [ ] `internal/modules/trigger/service.go` usa tipos locais
- [ ] `internal/modules/orchestrator/models.go` interface `TriggerEvaluator` usa `[]*triggermod.Trigger`
- [ ] `tests/integration/*` ajustados
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Trigger types to modules/trigger"` passa

---

## Sessão 9 — Review Module

- [ ] `internal/modules/review/models.go` define `Review`, `ReviewStatus`, `ValidationGate`, `ReviewCriteriaChecked` localmente
- [ ] `internal/modules/review/repository.go` usa tipos locais
- [ ] `internal/modules/review/service.go` usa tipos locais
- [ ] `internal/modules/orchestrator/models.go` interface `ReviewManager` usa `*reviewmod.Review`, `reviewmod.ValidationGate`
- [ ] `tests/integration/*` ajustados
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: migrate Review types to modules/review"` passa

---

## Sessão 10 — Cleanup `internal/domain/types.go`

- [ ] `internal/domain/types.go` NÃO contém `Task`, `TaskStatus`, `Priority`, `RiskLevel`
- [ ] `internal/domain/types.go` NÃO contém `Run`, `RunStatus`, `RunResult`
- [ ] `internal/domain/types.go` NÃO contém `WorkUnit`, `WorkUnitStatus`
- [ ] `internal/domain/types.go` NÃO contém `TaskGraph`, `TaskGraphStatus`
- [ ] `internal/domain/types.go` NÃO contém `Agent`, `AgentRuntimeType`
- [ ] `internal/domain/types.go` NÃO contém `AgentSession`, `AgentSessionStatus`
- [ ] `internal/domain/types.go` NÃO contém `PromptSnapshot`, `PromptFragment`, `ToolsetSnapshot`
- [ ] `internal/domain/types.go` NÃO contém `Trigger`, `Review` e tipos associados
- [ ] `internal/domain/types.go` contém APENAS `EventEnvelope`, `EventPriority`, constantes, e tipos genéricos
- [ ] Opcional: renomear `types.go` → `contracts.go` ou `events.go`
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: remove migrated entity types from domain package"` passa

---

## Sessão 11 — Finalização e Testes de Arquitetura

- [ ] Remover adapters temporários `toDomainXxx()` de `internal/core/orchestration/*`
- [ ] Adicionar teste: `internal/modules/*` não importa `internal/domain` para structs de entidade (exceto `EventEnvelope`)
- [ ] Adicionar teste: `internal/domain/types.go` não contém structs de entidade concretas
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] `go test ./...` passa (todos os pacotes)
- [ ] `go build ./...` passa
- [ ] `./scripts/safe-commit.sh "ADR-0022: add architecture tests for module isolation"` passa
- [ ] Criar Pull Request para `fix/adr-0022-module-isolation`
- [ ] CI passa (GitHub Actions)

---

## Status Geral

- **Sessão atual:** ___
- **Último commit:** ___
- **Build status:** ___
- **Test status:** ___
