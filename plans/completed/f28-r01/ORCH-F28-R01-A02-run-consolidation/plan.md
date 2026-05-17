# ORCH-F28-R01-A02: Run Module — Consolidation (Projection + WorkUnit Sync + Relay)

## Contexto do Projeto

**Projeto:** OrchestraOS (Go monolith modular)  
**Arquitetura:** Vertical Slice (ADR-0022) + Hybrid Intelligent Orchestrator (ADR-0023)  
**Meta:** Consolidar em `internal/modules/run/` toda a lógica relacionada a runs que atualmente está em `internal/core/coordination/` conforme ADR-0028  
**Risco:** **MÉDIO** — toca em 3 arquivos de coordination, bootstrap, orchestrator, cmd e testes.

---

## Documentação Obrigatória (LEIA ANTES DE COMEÇAR)

1. `docs/adr/0028-core-architecture-and-naming-standards.md` — Seções 2.1.1, 2.2.1, 4 Fase 2
2. `docs/adr/0022-vertical-slice-modules.md` — Regras de módulos e imports
3. `AGENTS.md` — Regras de commits
4. Este arquivo (`plan.md`)

---

## O que Já Existe vs O que Deve Ser Feito

### Arquivos-fonte do coordination/ a mover

| Arquivo Atual (coordination/) | Novo Arquivo (run/) | Responsabilidade | Linhas |
|---|---|---|---|
| `helpers.go` | `repository.go` (método `UpdateProjection`) | Atualizar projeção da tabela runs | ~36 |
| `run_workunit_sync.go` | `service_workunit.go` | Sincronização run ↔ work unit | ~78 |
| `runtime_relay.go` | `service_relay.go` | Consumo de eventos de runtime | ~375 |
| `queries.go` (parte) | `queries.go` (constante `QueryUpdateStatus`) | SQL UPDATE da tabela runs | ~8 |

### Pontos de integração a atualizar

| Arquivo | Uso Atual | Mudança Necessária |
|---|---|---|
| `internal/bootstrap/services.go` | `coordination.TransitionRunWithWorkUnit` callback | `runmod.TransitionRunWithWorkUnit` |
| `internal/bootstrap/services.go` | `coordination.NewRuntimeEventRelay(...)` | `runmod.NewRuntimeEventRelay(...)` |
| `internal/bootstrap/services.go` | `coordination.RuntimeEventRelay` tipo | `runmod.RuntimeEventRelay` tipo |
| `internal/modules/orchestrator/service.go` | `*coordination.RuntimeEventRelay` em Dependencies | `*runmod.RuntimeEventRelay` |
| `internal/modules/orchestrator/service.go` | `coordination.RelayConfig{}` | `runmod.RelayConfig{}` |
| `cmd/orchestraos/cmd/run.go` | `coordination.RelayConfig{}` | `runmod.RelayConfig{}` |
| `tests/unit/core/coordination/runtime_relay_test.go` | Testa `coordination.CheckpointTriggerForRuntimeEvent` | Mover para `tests/unit/modules/run/` e testar `runmod.CheckpointTriggerForRuntimeEvent` |

---

## Fronteiras de Isolamento

**TOCAR:**
- `internal/modules/run/repository.go` — adicionar método `UpdateProjection`
- `internal/modules/run/queries.go` — adicionar `QueryUpdateStatus` (ou similar) se não existir
- `internal/modules/run/service_workunit.go` — NOVO arquivo com `TransitionRunWithWorkUnit`
- `internal/modules/run/service_relay.go` — NOVO arquivo com `RuntimeEventRelay` + `EventSource` + interfaces + helpers
- `internal/bootstrap/services.go` — ajustar imports e referências
- `internal/modules/orchestrator/service.go` — ajustar tipo de `RuntimeEventRelay` e `RelayConfig`
- `cmd/orchestraos/cmd/run.go` — ajustar `RelayConfig`
- `tests/unit/core/coordination/runtime_relay_test.go` → mover para `tests/unit/modules/run/service_relay_test.go`
- `internal/core/coordination/helpers.go` — REMOVER após mover lógica
- `internal/core/coordination/run_workunit_sync.go` — REMOVER após mover lógica
- `internal/core/coordination/runtime_relay.go` — REMOVER após mover lógica
- `internal/core/coordination/queries.go` — REMOVER query de runs

**EVITAR:**
- `internal/modules/task/` — não toque (tarefa do A03)
- `internal/modules/agentsession/` — não toque (tarefa do A04)
- `internal/modules/prompt/` — não toque (tarefa do A05)
- `internal/core/transition/` — não toque exceto se necessário para compilar
- `tests/integration/*` — não toque exceto se necessário para compilar

---

## Interfaces Contratuais

As interfaces `SessionService` e `RunService` atualmente definidas em `coordination/runtime_relay.go` devem ser movidas para `run/service_relay.go` e manter a mesma assinatura. Elas são usadas pelo `RuntimeEventRelay` para evitar imports cíclicos:

```go
// SessionService abstracts agent-session operations needed by the relay.
type SessionService interface {
    Heartbeat(ctx context.Context, sessionID string, input domain.HeartbeatInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
    CheckpointFromEvent(ctx context.Context, sessionID string, event *domain.EventEnvelope) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
    Stop(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
    Fail(ctx context.Context, sessionID string, input transition.TransitionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
    Timeout(ctx context.Context, sessionID string, recoverableState json.RawMessage, input transition.TransitionInput) (*transition.OperationResult[*agentsessionmod.AgentSession], error)
    AutomaticCheckpoint(ctx context.Context, sessionID string, input domain.AutoCheckpointInput) (*transition.OperationResult[*agentsessionmod.AgentSession], *domain.CheckpointSuggestion, error)
}

// RunService abstracts run operations needed by the relay.
type RunService interface {
    Validate(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*runmod.Run], error)
    Complete(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*runmod.Run], error)
    Fail(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*runmod.Run], error)
    Timeout(ctx context.Context, runID string, input transition.TransitionInput) (*transition.OperationResult[*runmod.Run], error)
}
```

**Importante:** `EventSource` interface também deve ser movida:
```go
type EventSource interface {
    ReceiveEvent(ctx context.Context) (*domain.EventEnvelope, error)
}
```

---

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

**Caminho do checklist:** `plans/active/f28-r01/ORCH-F28-R01-A02-run-consolidation/checklist.md`

**A cada iteração:**
1. **LER** o checklist para identificar o próximo item pendente
2. **EXECUTAR** o item (código, teste, refactor)
3. **VALIDAR** o item (`go build ./...` no escopo afetado)
4. **ATUALIZAR** o checklist marcando o item como concluído
5. **CONTINUAR** para o próximo item

**Regras do Ralph Loop:**
- Nunca pule um item sem marcá-lo no checklist
- Se encontrar bloqueio, adicione uma nota na seção "Notas de Progresso"
- Ao final de cada ciclo significativo, faça um commit pequeno via `./scripts/safe-commit.sh`
- O checklist é sua fonte de verdade de progresso

---

## Regras de Implementação

1. **`UpdateProjection` em `repository.go`:**
   - Tornar-se um método de `Repository` (recebe `ctx, runID, status, result, failureReason`)
   - O SQL atualmente usa `coordination.QueryRunUpdateStatus` — mover para `run/queries.go` como `QueryUpdateStatus`
   - Ou inline o SQL no método (se o padrão do módulo for inline)

2. **`TransitionRunWithWorkUnit` em `service_workunit.go`:**
   - Função exportada do pacote `run`
   - Mantém a mesma assinatura
   - Usa `workunitmod.RequireByID`, `workunitmod.QueryUpdateStatus`, etc.
   - A função `workUnitEventTypeForStatus` pode ser privada (`workUnitEventTypeForStatus`) ou movida para `run/events.go`

3. **`RuntimeEventRelay` em `service_relay.go`:**
   - Struct, construtor, método `Run`, e todos os handlers privados
   - `EventSource`, `SessionService`, `RunService`, `RelayConfig` interfaces
   - `CheckpointTriggerForRuntimeEvent` e `decodePayloadMap` funções
   - **NÃO renomeie** `RuntimeEventRelay` — o nome já é descritivo. Mas o arquivo deve ser `service_relay.go`

4. **Imports:**
   - `service_relay.go` importará `agentsessionmod` — isso é permitido? Verifique `allowedModuleImports`: `run` pode importar `task` e `workunit`, mas NÃO `agentsession`.
   - **PROBLEMA:** `runtime_relay.go` atualmente importa `agentsessionmod`. Se mover para `run/`, o módulo `run` estará importando `agentsession`.
   - **SOLUÇÃO:** A importação de `agentsessionmod` em `service_relay.go` é APENAS para os tipos de retorno das interfaces `SessionService`. Os tipos `*agentsessionmod.AgentSession` e `domain.HeartbeatInput`, `domain.AutoCheckpointInput` são usados nas assinaturas de interface.
   - Verifique `allowedModuleImports`: `agentsession` pode importar `agent`. `run` pode importar `task` e `workunit`. Não está listado que `run` pode importar `agentsession`.
   - **DECISÃO:** Se `run` importar `agentsession` for proibido pelo teste de arquitetura, você tem duas opções:
     a) Adicionar `"run": {"agentsession": true}` em `allowedModuleImports` e documentar que é para tipos de interface DI do relay
     b) Extrair os tipos necessários para `domain/` e usar `domain.AgentSession` em vez de `*agentsessionmod.AgentSession`
   - **RECOMENDAÇÃO:** Opção (a) é mais simples — adicione a permissão e documente.

5. **Bootstrap:**
   - `RunService(db)` continua retornando `*runmod.RunService` — o callback `TransitionRunWithWorkUnit` agora vem de `runmod`
   - `RuntimeEventRelay(db)` retorna `*runmod.RuntimeEventRelay`
   - Remover import de `internal/core/coordination` do bootstrap

6. **Orchestrator:**
   - `Dependencies.RuntimeEventRelay` tipo muda de `func(db *sql.DB) *coordination.RuntimeEventRelay` para `func(db *sql.DB) *runmod.RuntimeEventRelay`
   - `RelayConfig{}` literal continua funcionando desde que os field names sejam os mesmos
   - Remover import de `internal/core/coordination` do orchestrator

---

## Regras Rígidas de Testes

1. Mova `tests/unit/core/coordination/runtime_relay_test.go` → `tests/unit/modules/run/service_relay_test.go`
2. Ajuste o package de `coordination_test` para `run_test`
3. Ajuste os imports no teste
4. Rode `go test ./internal/modules/run/...` — deve passar
5. Rode `go test ./tests/unit/modules/run/...` — deve passar
6. Rode `go test ./...` ao final — deve passar
7. Rode `./scripts/verify-contracts.sh` — deve passar
8. Rode `./scripts/lint.sh` — deve passar

---

## Code Review Auto-Crítico Obrigatório

Antes de entregar, responda:

- [ ] `internal/core/coordination/helpers.go` foi removido?
- [ ] `internal/core/coordination/run_workunit_sync.go` foi removido?
- [ ] `internal/core/coordination/runtime_relay.go` foi removido?
- [ ] `internal/core/coordination/queries.go` não contém mais SQL de runs?
- [ ] `run/repository.go` tem método `UpdateProjection` funcionando?
- [ ] `run/service_workunit.go` exporta `TransitionRunWithWorkUnit`?
- [ ] `run/service_relay.go` contém `RuntimeEventRelay` + interfaces + helpers?
- [ ] Bootstrap não importa mais `internal/core/coordination`?
- [ ] Orchestrator não importa mais `internal/core/coordination`?
- [ ] `cmd/orchestraos/cmd/run.go` não usa mais `coordination.RelayConfig`?
- [ ] Teste de runtime relay foi movido para `tests/unit/modules/run/`?
- [ ] `go build ./...` passa?
- [ ] `go test ./...` passa?
- [ ] Architecture tests passam?

---

## Critérios de Aceite Verificáveis

1. `internal/modules/run/repository.go` contém método `UpdateProjection`
2. `internal/modules/run/service_workunit.go` existe e exporta `TransitionRunWithWorkUnit`
3. `internal/modules/run/service_relay.go` existe e contém `RuntimeEventRelay` struct
4. `internal/core/coordination/helpers.go` NÃO existe mais
5. `internal/core/coordination/run_workunit_sync.go` NÃO existe mais
6. `internal/core/coordination/runtime_relay.go` NÃO existe mais
7. `tests/unit/modules/run/service_relay_test.go` existe (veio do coordination)
8. `internal/bootstrap/services.go` não importa `internal/core/coordination`
9. `internal/modules/orchestrator/service.go` não importa `internal/core/coordination`
10. `cmd/orchestraos/cmd/run.go` não usa símbolos do pacote `coordination`
11. `go build ./...` passa
12. `go test ./...` passa
13. `./scripts/verify-contracts.sh` passa
14. `./scripts/lint.sh` passa
15. Architecture tests passam (especialmente `module_boundaries_test.go`)

---

## Entrega Final

1. Commit com `./scripts/safe-commit.sh "refactor(run): consolidate coordination logic into run module per ADR-0028"`
2. Push da feature branch
3. Reportar ao usuário: "Run module consolidation completa. Build verde."
