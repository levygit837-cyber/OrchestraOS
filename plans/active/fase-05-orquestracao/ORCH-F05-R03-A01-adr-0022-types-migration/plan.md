# ORCH-F05-R03-A01 — Correção ADR-0022: Migração de Tipos de Domínio para Módulos Verticais

## Contexto

O `internal/domain/types.go` concentra TODAS as structs de entidade (`Task`, `Run`, `WorkUnit`, `AgentSession`, `PromptSnapshot`, etc.) e seus tipos enumerados. Isso viola o ADR-0022 (Vertical Slice Architecture), que exige que cada módulo em `internal/modules/*` seja **autônomo** e **não importe outros módulos**.

A dependência unidirecional `modules/* → domain/types.go` cria:
- **Acoplamento global**: alterar um campo em `Task` expõe código de `Run`, `AgentSession`, etc.
- **Ruído de contexto para LLMs**: ao trabalhar em um módulo, o agente carrega tipos de todas as entidades.
- **Quebra de modularidade**: módulos não são "fatias verticais" isoladas.

## Objetivo

Migrar **cada struct de entidade e seus tipos associados** de `internal/domain/types.go` para o respectivo módulo em `internal/modules/<entidade>/models.go`, garantindo que:

1. `internal/domain/` contenha **apenas contratos compartilhados** (`EventEnvelope`, `EventPriority`, tipos genéricos).
2. Cada módulo seja **100% autônomo** em seus tipos.
3. O build (`go build ./...`) e os testes (`go test ./...`) passem **a cada sessão**.
4. Commits sejam feitos via `./scripts/safe-commit.sh` a cada módulo concluído.

---

## Inventário de Entidades em `internal/domain/types.go`

| Entidade | Módulo Destino | Structs + Tipos | Consumidores Cruzados |
|----------|---------------|-----------------|----------------------|
| `Task` | `internal/modules/task/` | `Task`, `TaskStatus`, `Priority`, `RiskLevel` | orchestrator, bootstrap, cmd, core/orchestration, tests |
| `Run` | `internal/modules/run/` | `Run`, `RunStatus`, `RunResult` | orchestrator, bootstrap, core/orchestration, tests |
| `WorkUnit` | `internal/modules/workunit/` | `WorkUnit`, `WorkUnitStatus` | orchestrator, bootstrap, core/orchestration, tests |
| `TaskGraph` | `internal/modules/taskgraph/` | `TaskGraph`, `TaskGraphStatus` | orchestrator, bootstrap, tests |
| `AgentSession` | `internal/modules/agentsession/` | `AgentSession`, `AgentSessionStatus` | orchestrator, bootstrap, core/orchestration, tests |
| `Agent` | `internal/modules/agent/` | `Agent`, `AgentRuntimeType` | orchestrator, bootstrap, tests |
| `PromptFragment`, `PromptFragmentRef`, `PromptSnapshot` | `internal/modules/prompt/` | (já parcialmente local) | core/orchestration, tests |
| `ToolsetTool`, `ToolsetSnapshot` | `internal/modules/prompt/` | (já parcialmente local) | core/orchestration, tests |
| `Trigger` | `internal/modules/trigger/` | `Trigger`, `TriggerType`, `TriggerStatus`, `AnomalyType`, `ResolutionAction`, `ThresholdConfig` | orchestrator, tests |
| `Review` | `internal/modules/review/` | `Review`, `ReviewStatus`, `ReviewDecision`, `ValidationGate`, `ReviewCriteriaChecked` | orchestrator, tests |

**Permanecem em `internal/domain/`:**
- `EventEnvelope`, `EventPriority` e constantes (contrato do event store)
- Tipos genéricos não vinculados a uma entidade específica

---

## Estratégia de Migração: "Inside-Out + Adapter Bridge"

Em vez de migrar tudo de uma vez (que quebra o build em cascata), usamos uma abordagem **incremental por módulo**, com **pontes temporárias** (type aliases ou adapters) nos consumidores cruzados.

### Regra de Ouro por Sessão

> **Cada sessão migra UM módulo completo, atualiza seus consumidores imediatos, faz build+test passar, e commita.**

### Padrão de Migração por Módulo

Para cada entidade `X`:

1. **Criar tipos locais** em `internal/modules/x/models.go`
   - Copiar a struct e seus tipos associados do `domain/types.go`
   - Remover o prefixo do pacote dos nomes de tipo (ex: `TaskStatus` → `Status`)
   - Manter as tags JSON idênticas

2. **Atualizar arquivos internos do módulo**
   - `repository.go`: substituir `*domain.X` → `*x.X`
   - `service.go`: substituir `domain.XStatus` → `x.Status`
   - `fetch.go`: substituir retornos `*domain.X` → `*x.X`
   - `events.go`: atualizar funções auxiliares de event types
   - `*_test.go`: atualizar construção de structs e constantes

3. **Criar/Atualizar adapters nos consumidores cruzados**
   - Se um consumidor (ex: `core/orchestration/prompt_orchestrator.go`) precisa de `*domain.Task` mas o módulo agora retorna `*task.Task`, criar uma **função adapter temporária** no consumidor:
     ```go
     // TODO: remover quando prompt module for totalmente desacoplado de domain.Task
     func toDomainTask(t *task.Task) *domain.Task { ... }
     ```
   - Alternativa: usar **type alias temporário** no módulo se necessário para compatibilidade:
     ```go
     // models.go (compatibilidade temporária)
     type Task = domain.Task  // REMOVER na fase final
     ```

4. **Atualizar `internal/bootstrap/services.go`**
   - Os adapters do bootstrap (`taskAdapter`, `runAdapter`, etc.) devem usar os tipos locais.
   - Se as interfaces do `orchestrator` ainda esperam `*domain.Task`, atualizar as interfaces para `*taskmod.Task`.

5. **Build + Test + Commit**
   - `go build ./...`
   - `go test ./...`
   - `./scripts/safe-commit.sh "ADR-0022: migrate X types to modules/x"`

---

## Roadmap por Sessão

### Sessão 1 — Task Module (Fundação)
**Pré-requisito:** Branch `fix/adr-0022-module-isolation` criada e checkout feito.

**Ações:**
1. `internal/modules/task/models.go`: definir `Task`, `Status`, `Priority`, `RiskLevel` localmente.
2. `internal/modules/task/repository.go`: usar `*Task` e `Task` locais.
3. `internal/modules/task/service.go`: usar `Status`, `Priority`, `RiskLevel` locais.
4. `internal/modules/task/fetch.go`: retornar `*Task`.
5. `internal/modules/task/events.go`: usar `Status` local.
6. `internal/modules/task/validation_test.go`: usar `Priority` e `RiskLevel` locais.
7. `internal/core/orchestration/prompt_orchestrator.go`: criar adapter `toDomainTask()` para o `prompt` module.
8. `cmd/orchestraos/cmd/task.go`: usar `task.Priority` e `task.RiskLevel`.
9. `tests/integration/*`: substituir `domain.PriorityP2` → `task.PriorityP2`, `domain.TaskStatusCreated` → `task.StatusCreated`, etc.
10. Build + Test + Commit.

**Critério de aceitação:** `go build ./...` e `go test ./...` passam. Task module não importa `internal/domain` para tipos de entidade.

---

### Sessão 2 — Run Module
**Pré-requisito:** Sessão 1 concluída e commitada.

**Ações:**
1. `internal/modules/run/models.go`: definir `Run`, `Status`, `Result` localmente.
2. `internal/modules/run/repository.go`, `service.go`, `fetch.go`, `events.go`, `service_retry.go`: usar tipos locais.
3. `internal/core/orchestration/agentsession_orchestrator.go`: usar `runmod.StatusCompleted` etc.
4. `internal/core/orchestration/cascade.go`: usar `runmod.StatusCancelled` e `runmod.ResultForStatus()`.
5. `internal/core/orchestration/helpers.go`: atualizar `UpdateRunProjection` para `runmod.Status` e `runmod.Result`.
6. `tests/integration/*`: substituir `domain.RunStatusRunning` → `run.StatusRunning`, etc.
7. Build + Test + Commit.

**Critério de aceitação:** Build e testes passam. Run module autônomo.

---

### Sessão 3 — WorkUnit Module
**Pré-requisito:** Sessão 2 concluída.

**Ações:**
1. `internal/modules/workunit/models.go`: definir `WorkUnit`, `Status` localmente.
2. `internal/modules/workunit/repository.go`, `service.go`, `fetch.go`, `service_create.go`, `validation.go`: usar tipos locais.
3. `internal/core/orchestration/cascade.go`: atualizar workunit references.
4. `internal/modules/orchestrator/service.go`: atualizar `executeWorkUnit` e `topologicalSort` para `workunitmod.WorkUnit`.
5. `internal/modules/orchestrator/models.go`: atualizar `WorkUnitLister` interface.
6. `tests/integration/*`: ajustar referências.
7. Build + Test + Commit.

**Critério de aceitação:** Build e testes passam.

---

### Sessão 4 — TaskGraph Module
**Pré-requisito:** Sessão 3 concluída.

**Ações:**
1. `internal/modules/taskgraph/models.go`: definir `TaskGraph`, `Status` localmente.
2. Atualizar `repository.go`, `service.go` do taskgraph.
3. `internal/modules/orchestrator/models.go`: `TaskGraphManager` interface.
4. `tests/integration/*`: ajustar.
5. Build + Test + Commit.

---

### Sessão 5 — AgentSession Module
**Pré-requisito:** Sessão 4 concluída.

**Ações:**
1. `internal/modules/agentsession/models.go`: definir `AgentSession`, `Status` localmente.
2. Atualizar `repository.go`, `service.go`, `fetch.go`, `service_checkpoint.go`, `service_heartbeat.go`, `checkpoint_policy.go`.
3. `internal/core/orchestration/prompt_orchestrator.go`: adapter `toDomainAgentSession()`.
4. `internal/modules/orchestrator/models.go`: `SessionManager` interface.
5. Build + Test + Commit.

---

### Sessão 6 — Agent Module
**Pré-requisito:** Sessão 5 concluída.

**Ações:**
1. `internal/modules/agent/models.go`: definir `Agent`, `RuntimeType` localmente.
2. Atualizar `repository.go`, `service.go` do agent.
3. `internal/modules/orchestrator/models.go`: `AgentManager` interface.
4. Build + Test + Commit.

---

### Sessão 7 — Prompt Module (PromptSnapshot, ToolsetSnapshot)
**Pré-requisito:** Sessão 6 concluída.

**Ações:**
1. `internal/modules/prompt/types.go` já tem tipos locais, mas verificar se ainda há aliases para `domain`.
2. Se houver referências a `domain.PromptSnapshot` ou `domain.ToolsetSnapshot` em consumidores, criar adapters.
3. Build + Test + Commit.

---

### Sessão 8 — Trigger Module
**Pré-requisito:** Sessão 7 concluída.

**Ações:**
1. `internal/modules/trigger/models.go`: definir `Trigger`, `TriggerType`, `TriggerStatus`, `AnomalyType`, `ResolutionAction`, `ThresholdConfig` localmente.
2. Atualizar `repository.go`, `service.go`.
3. `internal/modules/orchestrator/models.go`: `TriggerEvaluator` interface.
4. Build + Test + Commit.

---

### Sessão 9 — Review Module
**Pré-requisito:** Sessão 8 concluída.

**Ações:**
1. `internal/modules/review/models.go`: definir `Review`, `ReviewStatus`, `ValidationGate`, `ReviewCriteriaChecked` localmente.
2. Atualizar `repository.go`, `service.go`.
3. `internal/modules/orchestrator/models.go`: `ReviewManager` interface.
4. Build + Test + Commit.

---

### Sessão 10 — Cleanup de `internal/domain/types.go`
**Pré-requisito:** Todas as sessões 1-9 concluídas.

**Ações:**
1. Remover TODAS as structs de entidade migradas de `internal/domain/types.go`.
2. Remover TODOS os tipos enumerados de entidade (ex: `TaskStatus`, `RunStatus`, etc.).
3. Manter apenas:
   - `EventEnvelope`
   - `EventPriority` e constantes
   - Tipos genéricos realmente compartilhados (se houver)
4. Opcional: renomear `types.go` para `events.go` ou `contracts.go`.
5. Rodar `./scripts/verify-contracts.sh` e `./scripts/lint.sh`.
6. Build + Test + Commit.

---

### Sessão 11 — Adapters Finais e Testes de Arquitetura
**Pré-requisito:** Sessão 10 concluída.

**Ações:**
1. Identificar e remover todos os adapters temporários `toDomainXxx()` em `core/orchestration/` e outros lugares.
2. Se um consumidor ainda precisa de `domain.Task`, reavaliar: ele deveria usar `taskmod.Task` diretamente ou o contrato deveria ser redefinido.
3. Adicionar teste de arquitetura: garantir que `internal/modules/*` NÃO importe `internal/domain` para structs de entidade (exceto `EventEnvelope`).
4. Adicionar teste de arquitetura: garantir que `internal/domain/types.go` NÃO contenha structs de entidade concretas.
5. Build + Test + `./scripts/safe-commit.sh`.
6. Criar Pull Request.

---

## Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação |
|-------|--------------|---------|-----------|
| Quebra de event payloads serializados | Média | Alto | Manter tags JSON idênticas; testes de integração cobrem serialization |
| Dependência circular entre módulos | Baixa | Alto | Seguir ordem do roadmap; nunca adicionar import `A → B` se `B → A` existe |
| Bootstrap incompatível | Alta | Médio | Atualizar interfaces do orchestrator e adapters do bootstrap a cada módulo |
| Tests de integração quebram | Alta | Baixo | Atualizar tests imediatamente a cada módulo; manter constantes com valores idênticos |
| Go exige type casting explícito | Média | Médio | Usar adapters temporários; Go compila string-based enums sem casting se forem alias |

---

## Convenções de Código para a Migração

1. **Nomenclatura de tipos no módulo:** Remova o prefixo do pacote. Ex: `domain.TaskStatus` → `task.Status`.
2. **Nomenclatura de constantes:** Mantenha o prefixo descritivo. Ex: `StatusCreated`, `PriorityP2`.
3. **Tags JSON:** Mantenha **idênticas** às originais para não quebrar serialization/event store.
4. **Adapters temporários:** Prefixe com `toDomain` e adicione `// TODO: remover quando [módulo] for desacoplado`.
5. **Commits:** Use `./scripts/safe-commit.sh` sempre. Mensagem padrão: `ADR-0022: migrate <Entity> types to modules/<entity>`.

---

## Checklist Geral do Roadmap

Veja arquivo companion: `checklist.md`
