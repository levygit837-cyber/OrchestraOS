# ORCH-F05-R03-A03 — ADR-0022: Migração do Módulo Orchestrator e Coordination

## Contexto

Após a migração dos 9 módulos verticais (A01–A09) e o cleanup de tipos mortos (PR #25), restam **22 TODOs ADR-0022** concentrados em três áreas:

1. `internal/modules/orchestrator/` — Interfaces DI ainda consomem `domain.Task`, `domain.Run`, `domain.AgentSession`, `domain.WorkUnit`, `domain.Agent`, `domain.AgentRuntimeType`.
2. `internal/bootstrap/services.go` — 11 adapters (`taskToDomain`, `runToDomain`, `workunitToDomain`, `agentToDomain`, `agentSessionToDomain`) e 6 struct adapters existem apenas para converter tipos de módulo → `domain`.
3. `internal/core/coordination/` — 5 TODOs em helpers, cascade, run_workunit_sync, agentsession_orchestrator, runtime_relay ainda usam `domain.Run`, `domain.RunStatus`, `domain.AgentSession`.

O `statemachine/replay.go` também usa `domain.*Status`, mas não possui TODOs ADR-0022 porque opera sobre eventos genéricos. Essa migração **não remove** os tipos de `domain/types.go`; ela elimina os adapters desnecessários.

## Objetivo

1. Migrar o módulo `orchestrator` para consumir tipos locais dos módulos verticais.
2. Migrar `core/coordination` para usar tipos de módulo onde houver TODOs ADR-0022.
3. Remover todos os adapters `XxxToDomain` e struct adapters do `bootstrap/services.go`.
4. Fazer `go build ./...` e `go test ./...` passar a cada fase.

## NÃO está no escopo

- Remover tipos de `internal/domain/types.go` (statemachine/replay.go e event payloads ainda os necessitam).
- Alterar lógica de negócio do orchestrator (apenas troca de tipos).
- Alterar módulos já migrados (task, run, workunit, etc.).

---

## Estratégia de Migração: Coordination → Orchestrator → Bootstrap

A ordem é crítica: `bootstrap` depende de `orchestrator` e `coordination`, então deve ser a última a mudar.

### Fase 1: Coordination Layer (pré-requisito)

Atualiza `core/coordination/*` para aceitar/receber tipos de módulo. Isso libera `bootstrap` e `orchestrator` para usar tipos de módulo diretamente.

**Arquivos:**
- `internal/core/coordination/helpers.go`
- `internal/core/coordination/cascade.go`
- `internal/core/coordination/run_workunit_sync.go`
- `internal/core/coordination/agentsession_orchestrator.go`
- `internal/core/coordination/runtime_relay.go`

### Fase 2: Orchestrator Module

Atualiza interfaces e implementação do orchestrator para usar tipos de módulo.

**Arquivos:**
- `internal/modules/orchestrator/models.go`
- `internal/modules/orchestrator/service.go`
- `internal/modules/orchestrator/validation.go`

### Fase 3: Bootstrap Cleanup

Remove todos os adapters desnecessários e simplifica o wiring.

**Arquivos:**
- `internal/bootstrap/services.go`

### Fase 4: Testes de Arquitetura

Atualiza `tests/architecture/module_boundaries_test.go` se novos imports forem adicionados.

### Fase 5: Build + Test + Commit

---

## Branch

```
feat/adr22-orchestrator-migration
```

---

## Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação |
|-------|--------------|---------|-----------|
| `transition.OperationResult[T]` genérico quebra com mudança de tipo | Baixa | Alto | O tipo `T` é genérico; Go aceita qualquer tipo. Testar build imediatamente. |
| `core/coordination` usado por múltiplos consumidores | Alta | Médio | Mudar coordination primeiro (Fase 1), build+test, depois prosseguir. |
| `WorkUnitLister` muda de `[]domain.WorkUnit` para `[]workunit.WorkUnit` | Média | Médio | `workunitmod.NewRepository(db).ListByTaskGraph()` já retorna `[]workunitmod.WorkUnit`. Wire direto. |
| `AgentManager` muda `domain.AgentRuntimeType` → `agentmod.RuntimeType` | Média | Médio | `agentmod.RuntimeType` é string-based alias; `agentmod.RuntimeType(runtimeType)` funciona. |
| Testes de integração quebram | Média | Baixo | Build+test a cada fase. Tipos têm mesmos campos/JSON tags. |

---

## Commits

Usar `./scripts/safe-commit.sh` a cada fase:

1. `ADR-0022: migrate coordination layer to module types`
2. `ADR-0022: migrate orchestrator module to module types`
3. `ADR-0022: remove bootstrap adapters and simplify wiring`
4. `ADR-0022: update architecture tests for orchestrator imports`
