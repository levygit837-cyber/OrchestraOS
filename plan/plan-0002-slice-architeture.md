# Reorganização de `internal/services` em Sub-Pacotes LLM-Friendly

## Contexto

O pacote `internal/services` é um **pacote flat com 15 arquivos e ~4.700 linhas** que agrupa 7 subsistemas distintos. Quando uma LLM precisa trabalhar em um subsistema (ex: Planner), ela acaba carregando para o contexto TODOS os arquivos do pacote, incluindo ~3.500 linhas que não têm relação com a tarefa.

O problema não é só organização para humanos — é **eficiência de contexto para agentes LLM**. O objetivo desta reorganização é criar **ilhas de contexto** onde cada sub-pacote contém tudo que a LLM precisa para entender e operar naquele domínio, sem poluição.

### Princípio: Cada sub-pacote como uma "unidade de contexto" para LLMs

Uma LLM que precisa trabalhar em checkpoints **não precisa ler** o planner.
Uma LLM que precisa modificar o planner **não precisa entender** retry policies.

A estrutura proposta garante que: **abrir todos os arquivos de um sub-pacote = ter 100% do contexto necessário**.

## User Review Required

> [!IMPORTANT]
> Esta reorganização altera o import path de **todos os consumers** de `internal/services` — CLI, testes de integração e pacote `orchestration`. Embora os tipos públicos e assinaturas permaneçam idênticos, será um refactor transversal.

> [!WARNING]
> A ADR 0017 define os 5 serviços canônicos (`TaskService`, `RunService`, `WorkUnitService`, `AgentSessionService`, `EventService`). Esta reorganização **respeita** esses serviços mas os distribui em pacotes próprios. Se a ADR 0017 for considerada normativa quanto ao pacote único, precisamos revisá-la primeiro.

## Open Questions

1. **Preservar `services` como façade?** Podemos criar um `internal/services/services.go` que re-exporta todos os tipos dos sub-pacotes para manter backward compatibility temporária. Isso facilita a migração mas adiciona indireção. Recomendo **não fazer** e atualizar os imports diretamente — é mais limpo e a LLM entende melhor. O que prefere?

2. **Nome do pacote de utilidades compartilhadas:** Proponho `svckit` (service kit). Alternativas: `svcutil`, `servicecore`, `shared`. O nome `common` deve ser evitado pois não carrega semântica.

3. **`checkpoint_policy.go` pertence a `agentsession/` ou merece pacote próprio?** Hoje é um arquivo que estende `AgentSessionService` com métodos de checkpoint. Recomendo manter junto do `agentsession/` porque o `SuggestCheckpoint` e `AutomaticCheckpoint` são métodos do `AgentSessionService`. Concorda?

---

## Análise de Dependências (Grafo Atual)

Antes de propor a estrutura, identifiquei as dependências cruzadas entre os arquivos:

```mermaid
graph TD
    subgraph "core/orchestration/"
        CO[TransitionInput<br>OperationResult<br>AppendTransition<br>GetTask/GetRun/etc.]
    end

    subgraph "core/db/"
        CD[BeginTx, CommitTx, RollbackTx<br>EnsureRowsAffected, AdvisoryLock]
    end

    subgraph "core/validation/"
        CV[RequiredUUID, RequiredText<br>Priority, RiskLevel, Runtime]
    end

    subgraph "core/serialization/"
        CS[MarshalPayload]
    end

    subgraph "modules/run/"
        MR[RetryPolicy, UpdateRunProjection<br>ResultForStatus, EventTypeForStatus]
    end

    subgraph "modules/task/"
        MT[EventTypeForStatus]
    end

    subgraph "modules/workunit/"
        MW[EventTypeForStatus]
    end

    subgraph "modules/agentsession/"
        MA[EventTypeForStatus]
    end

    subgraph "services/"
        TS[TaskService]
        WS[WorkUnitService]
        RS[RunService]
        AS[AgentSessionService]
        ES[EventService]
        TG[TaskGraphService]
        PS[PromptService]
    end

    TS --> CO
    TS --> CD
    TS --> CV
    TS --> CS
    TS --> MR
    TS --> MT
    TS --> MW
    WS --> CO
    WS --> CD
    WS --> CV
    WS --> CS
    RS --> CO
    RS --> CD
    RS --> CV
    RS --> CS
    RS --> MR
    RS --> MW
    AS --> CO
    AS --> CD
    AS --> CV
    AS --> CS
    AS --> MR
    ES --> CV
    TG --> CD
    TG --> CV
    TG --> CS
    PS --> CO
    PS --> CD
    PS --> CV
    PS --> CS
```

> [!NOTE]
> As linhas tracejadas representam **acoplamentos funcionais** — não são imports, porque tudo está no mesmo pacote. Esses são os pontos que precisam ser resolvidos na extração.

---

## Proposed Changes

### Estrutura Final Implementada

Helpers extraídos para infraestrutura (`core/`) e domínio (`modules/`):

```
internal/core/
├── db/
│   ├── conn.go, dbtx.go        # Conn pool, DBTX interface
│   └── txkit.go                # BeginTx, CommitTx, RollbackTx, EnsureRowsAffected, AdvisoryLock
├── orchestration/
│   ├── commands.go             # Commander com TransitionOptions
│   └── helpers.go              # TransitionInput, OperationResult, AppendTransition, GetTask/GetRun/etc.
├── serialization/
│   └── serialization.go        # MarshalPayload
├── validation/
│   └── validation.go           # RequiredUUID, RequiredText, Priority, RiskLevel, Runtime
├── eventstore/
│   └── store.go                # Event store com schema validation

internal/modules/
├── event/
│   ├── models.go, repository.go, queries.go, service.go
├── task/
│   ├── models.go, repository.go, queries.go, events.go
├── workunit/
│   ├── models.go, repository.go, queries.go, events.go
├── run/
│   ├── models.go, repository.go, queries.go, events.go, retry.go, projection.go
├── agentsession/
│   ├── models.go, repository.go, queries.go, events.go
├── taskgraph/
│   ├── models.go, repository.go, queries.go
├── prompt/
│   ├── models.go, repository.go, queries.go

internal/services/              # REMOVIDO — todos os serviços migrados para modules/
```

### Contagem de contexto por sub-pacote (estimativa)

| Pacote | Arquivos | Linhas | LLM lê tudo? |
|---|---|---|---|
| `core/db/` | 3 | ~150 | Só quando precisa de transações |
| `core/orchestration/` | 3 | ~500 | Só para transições cross-domain |
| `core/validation/` | 1 | ~80 | Só para validação de input |
| `core/serialization/` | 1 | ~20 | Raramente |
| `core/eventstore/` | 4 | ~400 | Só para persistência de eventos |
| `modules/event/` | 4 | ~290 | ✅ auto-contido |
| `modules/task/` | 4 | ~300 | ✅ auto-contido |
| `modules/workunit/` | 4 | ~550 | ✅ auto-contido |
| `modules/run/` | 4 | ~480 | ✅ auto-contido |
| `modules/agentsession/` | 3 | ~510 | ✅ auto-contido |
| `modules/taskgraph/` | 3 | ~650 | ✅ auto-contido |
| `modules/prompt/` | 6 | ~500 | ✅ auto-contido |
| `modules/agent/` | 5 | ~400 | ✅ auto-contido |

**Antes:** LLM carrega ~4.700 linhas para qualquer tarefa (services + common.go).
**Depois:** LLM carrega ~150-500 linhas de core/ quando necessário + módulos de ~300-650 linhas. Redução de **60-85% de contexto irrelevante** para tarefas em módulos.

---

### O papel do `doc.go` (crítico para LLMs)

Cada sub-pacote terá um `doc.go` que serve como **briefing de contexto** para a LLM. Formato padronizado:

```go
// Package planner implements task decomposition into directed acyclic graphs
// (DAGs) of WorkUnits.
//
// # Responsibility
// Transforms a domain.Task into a GraphPlan containing WorkUnits, nodes,
// edges and a rationale. Two strategies are available: GeminiPlanner (LLM)
// and local heuristic (inside taskgraph/).
//
// # Key Types
//   - Planner: interface that any decomposition strategy must implement
//   - GraphPlan: result struct with graph ID, work units, nodes and edges
//   - GeminiPlanner: concrete implementation using Gemini API
//
// # Dependencies
//   - domain: WorkUnit, Task, TaskGraphNodeInfo, TaskGraphEdgeInfo
//   - apperrors: error typing
//   - genai: Google Gemini SDK (only in gemini.go)
//
// # Related Packages
//   - taskgraph/: orchestrates decomposition lifecycle, persists results
//   - svckit/: shared validation and persistence helpers
package planner
```

Este `doc.go` permite que a LLM, ao abrir o pacote, entenda imediatamente:
1. **O que faz** → decide se precisa ler os outros arquivos
2. **Tipos principais** → sabe o que procurar
3. **Dependências** → sabe quais outros pacotes pode precisar consultar
4. **Pacotes relacionados** → navegação intencional, não fishing

---

### Resolução de Acoplamentos Cruzados

#### `RunService` → `WorkUnitService` (transitionRelatedWorkUnit)

A função `transitionRelatedWorkUnit` atualmente chama `validateDependenciesCompleted` e `validateOwnedPathAvailability` que são internas de `workunit_service.go`.

**Solução:** Extrair essas validações para `workunit/validation.go` e exportá-las. O `RunService` importa `workunit` para chamar validações mas **não** importa o service inteiro — importa funções de validação puras.

#### `TaskService` → `cancelTaskDependents`

A função cascata cancela WorkUnits e Runs. Atualmente usa repos diretamente.

**Solução:** Manter a lógica dentro de `task/service.go`, importando repos diretamente (como já faz). Não precisa chamar `WorkUnitService` ou `RunService` — usa repo + event append diretamente, preservando atomicidade transacional.

#### `AgentSessionService.Timeout` → pausa Run

**Solução:** `agentsession/service.go` importa `svckit` para helpers de transição e usa repo/event append diretamente para pausar o run. Não precisa do `RunService` — faz a operação atômica na mesma transaction.

---

### Infraestrutura Core (Extraída)

Em vez de um único `svckit/`, a infraestrutura foi dividida em pacotes especializados no `core/`:

| Pacote | Responsabilidade |
|---|---|
| `core/db/` | `BeginTx`, `CommitTx`, `RollbackTx`, `EnsureRowsAffected`, `AcquireAdvisoryTxLock`, `DBTX` |
| `core/orchestration/` | `TransitionInput`, `OperationResult[T]`, `TransitionPayload`, `RequireFinalAudit`, `IsFinalStatus`, `TransitionContext`, `AppendServiceEvent` |
| `core/validation/` | `RequiredUUID`, `OptionalUUID`, `RequiredText`, `StringList`, `Priority`, `RiskLevel`, `Runtime` |
| `core/serialization/` | `MarshalPayload` |
| `core/eventstore/` | `Store`, `Repository`, `Validator` — persistência com JSON-Schema |
| `core/statemachine/` | `CanTransition`, regras de transição de status |

---

### Consumers que precisam de atualização de imports

| Consumer | Import Atual | Imports Novos |
|---|---|---|
| `cmd/orchestraos/cmd/task.go` | `services` | `task`, `svckit` |
| `cmd/orchestraos/cmd/workunit.go` | `services` | `workunit`, `svckit` |
| `cmd/orchestraos/cmd/run.go` | `services` | `run`, `svckit` |
| `cmd/orchestraos/cmd/event.go` | `services` | `event` |
| `cmd/orchestraos/cmd/agentsession.go` | `services` | `agentsession`, `svckit` |
| `cmd/orchestraos/cmd/run_test.go` | `services` | `run`, `svckit` |
| `internal/orchestration/commands.go` | `services` | `task`, `workunit`, `run`, `agentsession`, `taskgraph`, `prompt`, `svckit` |
| `tests/integration/*.go` | `services` | Vários, conforme uso |

---

## Estratégia de Migração

A migração deve ser **incremental e verificável** (conforme AGENTS.md: "mudanças pequenas, verificáveis e reversíveis"):

### Fase 1: Extração da infraestrutura core (zero-risk)
- Criar `core/db/`, `core/orchestration/`, `core/validation/`, `core/serialization/`, `core/statemachine/`, `core/eventstore/`
- Manter `services/` original intacto compilando (ambos coexistem temporariamente)
- ✅ `go build ./...` + testes

### Fase 2: Extração dos serviços de lifecycle (`event/`, `task/`, `workunit/`)
- Criar sub-pacotes importando `core/*`
- Atualizar consumers incrementalmente
- ✅ `go build ./...` + testes após cada serviço

### Fase 3: Extração dos serviços de execução (`run/`, `agentsession/`)
- Resolver acoplamentos com DI interfaces (`TaskReader`, `WorkUnitReader`)
- ✅ `go build ./...` + testes

### Fase 4: Extração de `taskgraph/`, `prompt/` e `agent/`
- `taskgraph/` importa `agent/` para GeminiPlanner
- `prompt/` consolida catálogo, composer e toolset
- ✅ `go build ./...` + testes

### Fase 5: Limpeza
- ✅ Remover `internal/services/`, `internal/repository/`, `internal/prompting/`, `internal/db/`, `internal/agent/`
- Verificar que nenhum import antigo sobrevive
- ✅ `go build ./...` + testes de integração completos

### Fase 6: ADR + doc.go
- ✅ Registrar ADR 0022 documentando a decisão
- ✅ Escrever `doc.go` em cada sub-pacote
- ✅ Atualizar roadmap

---

## Verification Plan

### Automated Tests

```bash
# Após cada fase:
go build ./...
go vet ./...
go test ./internal/services/... -v
go test ./tests/integration/... -v -count=1
```

### Manual Verification

- Confirmar que `go doc ./internal/services/planner/` mostra documentação coerente
- Confirmar que `go doc ./internal/services/svckit/` mostra API exportada limpa
- Validar que nenhum import circular existe: `go vet ./...`
