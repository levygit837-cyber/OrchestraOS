# Migration Map: Types to internal/domain/

**Task ID:** 2026-05-21_architecture-patterns-and-refactor-mapping  
**ADR Ref:** ADR-0030 — Simplified Modular Architecture (Pilar 1)  
**Date:** 2026-05-21  
**Status:** Ready for Review  

---

## Resumo Executivo

Este documento mapeia TODOS os tipos definidos nos 10 módulos de `internal/modules/*/models.go`, classificando-os como **shared** (deve migrar para `internal/domain/`) ou **local** (permanece no módulo).

| Métrica | Valor |
|---------|-------|
| Módulos inventariados | 10 |
| Tipos shared (migrar para domain) | 26 entity types + 3 structs de payload |
| Tipos local (permanecer no módulo) | ~30+ types |
| Arquivos com imports a atualizar | 15+ |
| Conflitos de nomeação a resolver | 9 (Status em cada módulo) |

---

## Critério de Classificação

### Shared (vai para `internal/domain/`)
Um tipo é **shared** se atender a QUALQUER um dos critérios:
1. Usado como campo ou parâmetro/retorno em **2+ módulos**
2. Referenciado em `internal/bootstrap/services.go`
3. Referenciado em `internal/modules/orchestrator/models.go` (interfaces que expõem o tipo)
4. Listado em `sharedEntityTypes` do teste `TestDomainImportIntegrity`

### Local (fica no módulo)
Um tipo é **local** se:
1. Usado **apenas dentro do próprio módulo**
2. É tipo auxiliar para lógica interna (payload de evento, config, criteria)
3. Não é referenciado por nenhum outro módulo nem por bootstrap/orchestrator

---

## Inventário por Módulo

### Módulo: task

**Arquivo:** `internal/modules/task/models.go` (53 linhas)

| Tipo | Kind | Classificação | Destino | Usado Por | Notas |
|------|------|--------------|---------|-----------|-------|
| **Task** | struct | **shared** | `domain.Task` | run, workunit, taskgraph, orchestrator, bootstrap | Entity principal |
| **Status** | string | **shared** | `domain.TaskStatus` | run, workunit, taskgraph, orchestrator, bootstrap | Renomear para evitar conflito |
| **Priority** | string | **shared** | `domain.TaskPriority` | taskgraph, orchestrator | Usado em queries de planejamento |
| **RiskLevel** | string | **shared** | `domain.TaskRiskLevel` | run (validação de política), orchestrator | Usado em regras de execução |

**Observações:**
- Todos os 4 tipos são shared. O arquivo `models.go` do módulo task será eliminado após a migração (ou convertido em re-exports se necessário para compatibilidade temporária).

---

### Módulo: run

**Arquivo:** `internal/modules/run/models.go` (36 linhas)

| Tipo | Kind | Classificação | Destino | Usado Por | Notas |
|------|------|--------------|---------|-----------|-------|
| **Run** | struct | **shared** | `domain.Run` | orchestrator, trigger, agentsession (relay), bootstrap | Entity principal |
| **Status** | string | **shared** | `domain.RunStatus` | orchestrator, trigger, bootstrap | Renomear para evitar conflito |
| **Result** | string | **shared** | `domain.RunResult` | orchestrator, bootstrap | Usado em projeções e cascade |

**Observações:**
- Todos os 3 tipos são shared. O módulo run usa `taskmod.Task` em `service.go` (via TaskReader interface).

---

### Módulo: workunit

**Arquivo:** `internal/modules/workunit/models.go` (34 linhas)

| Tipo | Kind | Classificação | Destino | Usado Por | Notas |
|------|------|--------------|---------|-----------|-------|
| **WorkUnit** | struct | **shared** | `domain.WorkUnit` | run, orchestrator, trigger, taskgraph, bootstrap | Entity principal |
| **Status** | string | **shared** | `domain.WorkUnitStatus` | run, orchestrator, bootstrap | Renomear para evitar conflito |

**Observações:**
- Todos os 2 tipos são shared. O módulo workunit importa `task` e `taskgraph` em `service.go` e `service_create.go`.
- Comentário no arquivo: "NENHUM import de internal/domain neste arquivo" — será atualizado após migração.

---

### Módulo: agent

**Arquivo:** `internal/modules/agent/models.go` (30 linhas)

| Tipo | Kind | Classificação | Destino | Usado Por | Notas |
|------|------|--------------|---------|-----------|-------|
| **Agent** | struct | **shared** | `domain.Agent` | orchestrator, agentsession, bootstrap | Entity principal |
| **RuntimeType** | string | **shared** | `domain.AgentRuntimeType` | orchestrator, bootstrap | Usado em criação de runtime |
| **AgentStatus** | string | **local** | `agent.AgentStatus` | agent (próprio módulo) | Usado apenas internamente no módulo agent |

**Observações:**
- `AgentStatus` é **local** — não é referenciado por nenhum outro módulo. Apenas o próprio módulo agent usa esse tipo.
- `RuntimeType` é shared porque é parâmetro de `AgentManager.FindOrCreate` no orchestrator.

---

### Módulo: agentsession

**Arquivo:** `internal/modules/agentsession/models.go` (34 linhas)

| Tipo | Kind | Classificação | Destino | Usado Por | Notas |
|------|------|--------------|---------|-----------|-------|
| **AgentSession** | struct | **shared** | `domain.AgentSession` | orchestrator, trigger, run (relay), bootstrap | Entity principal |
| **Status** | string | **shared** | `domain.AgentSessionStatus` | orchestrator, run, bootstrap | Renomear para evitar conflito |

**Observações:**
- Todos os 2 tipos são shared. O módulo agentsession importa `agent` em `service.go` (via AgentReader).

---

### Módulo: taskgraph

**Arquivo:** `internal/modules/taskgraph/models.go` (69 linhas)

| Tipo | Kind | Classificação | Destino | Usado Por | Notas |
|------|------|--------------|---------|-----------|-------|
| **TaskGraph** | struct | **shared** | `domain.TaskGraph` | orchestrator, workunit, bootstrap | Entity principal |
| **Status** | string | **shared** | `domain.TaskGraphStatus` | orchestrator, bootstrap | Renomear para evitar conflito |
| **TaskGraphNodeInfo** | struct | **shared** | `domain.TaskGraphNodeInfo` | taskgraph, domain (já existe em event_payloads.go) | Já existe em `domain/event_payloads.go` |
| **TaskGraphEdgeInfo** | struct | **shared** | `domain.TaskGraphEdgeInfo` | taskgraph, domain (já existe em event_payloads.go) | Já existe em `domain/event_payloads.go` |
| **TaskGraphCreatedPayload** | struct | **shared** | `domain.TaskGraphCreatedPayload` | taskgraph, domain (já existe em event_payloads.go) | Já existe em `domain/event_payloads.go` |
| **PlanWorkUnit** | struct | **local** → **eliminar** | — | taskgraph, bootstrap | Espelha `workunit.WorkUnit`. Será eliminado quando WorkUnit estiver em domain. |

**Observações:**
- `PlanWorkUnit` é um caso especial. Ele foi criado para **evitar import cycle** entre taskgraph e workunit. Quando `WorkUnit` migrar para `domain.WorkUnit`, `PlanWorkUnit` pode ser eliminado e substituído diretamente por `domain.WorkUnit`.
- Os tipos `TaskGraphNodeInfo`, `TaskGraphEdgeInfo`, `TaskGraphCreatedPayload` já existem em `domain/event_payloads.go` — o módulo taskgraph ainda define cópias locais que devem ser removidas.

---

### Módulo: prompt

**Arquivo:** `internal/modules/prompt/models.go` (261 linhas)

| Tipo | Kind | Classificação | Destino | Usado Por | Notas |
|------|------|--------------|---------|-----------|-------|
| **PromptSnapshot** | struct | **shared** | `domain.PromptSnapshot` | orchestrator, bootstrap | Entity principal, usado em PreparedPrompt |
| **PromptFragment** | struct | **shared** | `domain.PromptFragment` | orchestrator (indireto) | Entity principal |
| **ToolsetSnapshot** | struct | **shared** | `domain.ToolsetSnapshot` | orchestrator, bootstrap | Entity principal |
| **ComposedPrompt** | struct | **shared** | `domain.ComposedPrompt` | orchestrator, bootstrap | Usado em PersistComposedPrompt |
| MaxAutonomyLevel | const (int) | local | — | prompt | Constante interna do módulo |
| FragmentKind | string | local | — | prompt | Usado apenas em lógica de composição |
| FragmentCategory | string | local | — | prompt | Usado apenas em lógica de composição |
| RequiredCategories | var ([]FragmentCategory) | local | — | prompt | Configuração interna |
| PromptFragmentRef | struct | local | — | prompt | Referência interna usada apenas no módulo prompt |
| ToolsetTool | struct | local | — | prompt | Tipo auxiliar interno |
| AppliesWhen | struct | local | — | prompt | Tipo auxiliar de composição |
| Fragment | struct | local | — | prompt | Runtime representation usada apenas no módulo |
| FragmentRef | struct | local | — | prompt | Runtime reference usada apenas no módulo |
| ToolRisk | string | local | — | prompt | Enum de risco de tool |
| Tool | struct | local | — | prompt | Tipo auxiliar de tool |
| ToolsetSelection | struct | local | — | prompt | Seleção de toolset |
| TaskContext | struct | local | — | prompt | Contexto de task para composição |
| SystemProfile | struct | local | — | prompt | Perfil do sistema |

**Observações:**
- Apenas 4 tipos são shared. Os demais são tipos de composição interna, configuração e runtime que não são expostos para fora do módulo.
- `PromptSnapshot`, `ToolsetSnapshot` e `ComposedPrompt` são referenciados diretamente em `orchestrator/models.go` (interfaces `PromptPersistence`, `PreparedPrompt`).

---

### Módulo: review

**Arquivo:** `internal/modules/review/models.go` (46 linhas)

| Tipo | Kind | Classificação | Destino | Usado Por | Notas |
|------|------|--------------|---------|-----------|-------|
| **Review** | struct | **shared** | `domain.Review` | orchestrator, bootstrap | Entity principal |
| **Status** | string | **shared** | `domain.ReviewStatus` | orchestrator, bootstrap | Renomear para evitar conflito |
| **ValidationGate** | string | **shared** | `domain.ReviewValidationGate` | orchestrator, bootstrap | Usado em ReviewManager.Create |
| Decision | alias (= Status) | local | — | review | Alias interno |
| CriteriaChecked | struct | local | — | review | Tipo auxiliar de critérios |

**Observações:**
- `Decision` é um alias para `Status` usado apenas internamente no módulo review.
- `CriteriaChecked` é um tipo auxiliar usado apenas no campo `CriteriaChecked` de `Review`.

---

### Módulo: trigger

**Arquivo:** `internal/modules/trigger/models.go` (69 linhas)

| Tipo | Kind | Classificação | Destino | Usado Por | Notas |
|------|------|--------------|---------|-----------|-------|
| **Trigger** | struct | **shared** | `domain.Trigger` | orchestrator, bootstrap | Entity principal |
| **Status** | string | **shared** | `domain.TriggerStatus` | orchestrator, bootstrap | Renomear para evitar conflito |
| **Type** | string | **shared** | `domain.TriggerType` | orchestrator, bootstrap | Usado em criação e avaliação |
| AnomalyType | string | local | — | trigger | Usado apenas internamente |
| ResolutionAction | string | local | — | trigger | Usado apenas internamente |
| ThresholdConfig | struct | local | — | trigger | Configuração de threshold |

**Observações:**
- `AnomalyType`, `ResolutionAction` e `ThresholdConfig` são tipos de configuração/lógica interna do módulo trigger.

---

### Módulo: orchestrator

**Arquivo:** `internal/modules/orchestrator/models.go` (146 linhas)

| Tipo | Kind | Classificação | Destino | Notas |
|------|------|--------------|---------|-------|
| RunTaskOptions | struct | local | — | Input de execução, usado apenas no orchestrator |
| RunTaskResult | struct | local | — | Output de execução, usado apenas no orchestrator |
| WorkUnitExecutionResult | struct | local | — | Resultado de execução de WU |
| DecomposeInput | struct | local | — | Input de decomposição |
| DecomposeResult | struct | local | — | Output de decomposição (contém pointers para types de outros módulos) |
| CreateRunInput | struct | local | — | Input de criação de run |
| CreateAgentSessionInput | struct | local | — | Input de criação de sessão |
| PreparedPrompt | struct | local | — | Contém `*promptmod.PromptSnapshot` e `*promptmod.ToolsetSnapshot` |
| Runtime | interface | local | — | Interface de runtime |
| RuntimeConfig | struct | local | — | Config de runtime |
| RuntimeStatus | struct | local | — | Status de runtime |
| TaskServiceReader | interface | local | — | Interface que expõe `*taskmod.Task` |
| TaskGraphManager | interface | local | — | Interface que expõe `*taskgraphmod.TaskGraph` |
| RunLifecycleManager | interface | local | — | Interface que expõe `*runmod.Run` |
| AgentManager | interface | local | — | Interface que expõe `*agentmod.Agent` e `agentmod.RuntimeType` |
| SessionManager | interface | local | — | Interface que expõe `*agentsessionmod.AgentSession` |
| PromptPersistence | interface | local | — | Interface que expõe `*promptmod.PromptSnapshot` e `*promptmod.ComposedPrompt` |
| ReviewManager | interface | local | — | Interface que expõe `*reviewmod.Review` e `reviewmod.ValidationGate` |
| TriggerEvaluator | interface | local | — | Interface que expõe `*triggermod.Trigger` |
| WorkUnitLister | interface | local | — | Interface que expõe `[]workunitmod.WorkUnit` |

**Observações:**
- TODOS os tipos do orchestrator são **locais** — são interfaces e structs de input/output específicos do orchestrator.
- Os interfaces referenciam tipos de outros módulos, mas os próprios tipos do orchestrator não precisam migrar.
- Após a migração, as interfaces usarão `domain.*` em vez de `*mod.*` (ex: `*domain.Task` em vez de `*taskmod.Task`).

---

## Resumo de Migração

### Tipos Shared (migrar para `internal/domain/`)

| # | Tipo Atual | Destino em domain | Módulo Origem |
|---|-----------|-------------------|---------------|
| 1 | `task.Task` | `domain.Task` | task |
| 2 | `task.Status` | `domain.TaskStatus` | task |
| 3 | `task.Priority` | `domain.TaskPriority` | task |
| 4 | `task.RiskLevel` | `domain.TaskRiskLevel` | task |
| 5 | `run.Run` | `domain.Run` | run |
| 6 | `run.Status` | `domain.RunStatus` | run |
| 7 | `run.Result` | `domain.RunResult` | run |
| 8 | `workunit.WorkUnit` | `domain.WorkUnit` | workunit |
| 9 | `workunit.Status` | `domain.WorkUnitStatus` | workunit |
| 10 | `agent.Agent` | `domain.Agent` | agent |
| 11 | `agent.RuntimeType` | `domain.AgentRuntimeType` | agent |
| 12 | `agentsession.AgentSession` | `domain.AgentSession` | agentsession |
| 13 | `agentsession.Status` | `domain.AgentSessionStatus` | agentsession |
| 14 | `taskgraph.TaskGraph` | `domain.TaskGraph` | taskgraph |
| 15 | `taskgraph.Status` | `domain.TaskGraphStatus` | taskgraph |
| 16 | `prompt.PromptSnapshot` | `domain.PromptSnapshot` | prompt |
| 17 | `prompt.PromptFragment` | `domain.PromptFragment` | prompt |
| 18 | `prompt.ToolsetSnapshot` | `domain.ToolsetSnapshot` | prompt |
| 19 | `prompt.ComposedPrompt` | `domain.ComposedPrompt` | prompt |
| 20 | `review.Review` | `domain.Review` | review |
| 21 | `review.Status` | `domain.ReviewStatus` | review |
| 22 | `review.ValidationGate` | `domain.ReviewValidationGate` | review |
| 23 | `trigger.Trigger` | `domain.Trigger` | trigger |
| 24 | `trigger.Status` | `domain.TriggerStatus` | trigger |
| 25 | `trigger.Type` | `domain.TriggerType` | trigger |

**Nota:** Os tipos `TaskGraphNodeInfo`, `TaskGraphEdgeInfo`, `TaskGraphCreatedPayload` já existem em `domain/event_payloads.go` mas ainda têm cópias em `taskgraph/models.go`. As cópias no módulo devem ser removidas.

### Tipos Local (permanecem nos módulos)

| Módulo | Tipos |
|--------|-------|
| task | (nenhum — todos são shared) |
| run | (nenhum — todos são shared) |
| workunit | (nenhum — todos são shared) |
| agent | `AgentStatus` |
| agentsession | (nenhum — todos são shared) |
| taskgraph | `PlanWorkUnit` (será eliminado) |
| prompt | `MaxAutonomyLevel`, `FragmentKind`, `FragmentCategory`, `RequiredCategories`, `PromptFragmentRef`, `ToolsetTool`, `AppliesWhen`, `Fragment`, `FragmentRef`, `ToolRisk`, `Tool`, `ToolsetSelection`, `TaskContext`, `SystemProfile` |
| review | `Decision`, `CriteriaChecked` |
| trigger | `AnomalyType`, `ResolutionAction`, `ThresholdConfig` |
| orchestrator | TODOS (RunTaskOptions, RunTaskResult, WorkUnitExecutionResult, DecomposeInput, DecomposeResult, CreateRunInput, CreateAgentSessionInput, PreparedPrompt, Runtime, RuntimeConfig, RuntimeStatus, e todas as interfaces) |

---

## Imports Impactados

### `internal/bootstrap/services.go`

**Imports a remover (após migração):**
```go
agentmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
promptmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/prompt"
reviewmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/review"
runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
triggermod "github.com/levygit837-cyber/OrchestraOS/internal/modules/trigger"
workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
```

**Imports a adicionar:**
```go
"github.com/levygit837-cyber/OrchestraOS/internal/domain"
```

**Conversões de tipo necessárias:**

| Linha | De | Para |
|-------|-----|------|
| ~24 | `*taskmod.TaskService` | `*task.TaskService` (service permanece no módulo) |
| ~29 | `*taskgraphmod.PlanWorkUnit` | `*domain.WorkUnit` |
| ~33 | `*workunitmod.WorkUnit` | `*domain.WorkUnit` |
| ~69 | `*agentmod.Agent` | `*domain.Agent` |
| ~73 | `agentmod.RuntimeType` | `domain.AgentRuntimeType` |
| ~82 | `*agentsessionmod.AgentSession` | `*domain.AgentSession` |
| ~91 | `*runmod.Run` | `*domain.Run` |
| ~96 | `*runmod.RunService` | `*run.RunService` (service permanece) |
| ~107 | `*workunitmod.WorkUnitService` | `*workunit.WorkUnitService` (service permanece) |
| ~115 | `*agentsessionmod.AgentSessionService` | `*agentsession.AgentSessionService` (service permanece) |
| ~129 | `*taskgraphmod.TaskGraphService` | `*taskgraph.TaskGraphService` (service permanece) |
| ~149 | `*taskgraphmod.PlanWorkUnit` | `*domain.WorkUnit` |
| ~159 | `[]taskgraphmod.PlanWorkUnit` | `[]domain.WorkUnit` |
| ~172 | `*promptmod.PromptService` | `*prompt.PromptService` (service permanece) |
| ~177 | `*reviewmod.ReviewService` | `*review.ReviewService` (service permanece) |
| ~187 | `*taskgraphmod.GeminiPlanner` | `*taskgraph.GeminiPlanner` (service permanece) |
| ~192 | `*taskmod.Task` | `*domain.Task` |
| ~197 | `*taskgraphmod.GraphPlan` | `*taskgraph.GraphPlan` (tipo local de taskgraph) |
| ~202 | `*triggermod.TriggerService` | `*trigger.TriggerService` (service permanece) |
| ~218 | `*runmod.RuntimeEventRelay` | `*run.RuntimeEventRelay` (service permanece) |
| ~227 | `*orchestratormod.Service` | `*orchestrator.Service` (service permanece) |
| ~262 | `*taskmod.Task` | `*domain.Task` |
| ~276 | `*runmod.Run` | `*domain.Run` |
| ~288 | `*taskgraphmod.TaskGraph` | `*domain.TaskGraph` |
| ~298 | `[]workunitmod.WorkUnit` | `[]domain.WorkUnit` |
| ~309 | `*agentsessionmod.AgentSession` | `*domain.AgentSession` |
| ~324 | `*reviewmod.Review` | `*domain.Review` |
| ~324 | `reviewmod.ValidationGate` | `domain.ReviewValidationGate` |
| ~335 | `*promptmod.ComposedPrompt` | `*domain.ComposedPrompt` |
| ~335 | `promptmod.PersistMetadata` | `prompt.PersistMetadata` (tipo local de prompt) |
| ~335 | `*promptmod.PreparedRunPrompt` | `*prompt.PreparedRunPrompt` (tipo local de prompt) |
| ~342 | `*workunitmod.WorkUnit` | `*domain.WorkUnit` |
| ~366 | `orchestratormod.RuntimeStatus` | Permanece (tipo local do orchestrator) |

**Observação:** Os services (`TaskService`, `RunService`, etc.) permanecem nos módulos. Apenas os **entity types** (structs e enums) migram para `domain`. As interfaces em `bootstrap/services.go` (adapters) continuam existindo, mas referenciam `domain.*` em vez de `*mod.*`.

---

### `internal/modules/orchestrator/models.go`

**Imports a remover:**
```go
agentmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agent"
agentsessionmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/agentsession"
promptmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/prompt"
reviewmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/review"
runmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/run"
taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"
taskgraphmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/taskgraph"
triggermod "github.com/levygit837-cyber/OrchestraOS/internal/modules/trigger"
workunitmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"
```

**Conversões de tipo nas interfaces:**

| Interface | Campo/Método | De | Para |
|-----------|-------------|-----|------|
| TaskServiceReader | GetByID retorno | `*taskmod.Task` | `*domain.Task` |
| TaskServiceReader | Complete retorno | `*taskmod.Task` | `*domain.Task` |
| TaskServiceReader | Fail retorno | `*taskmod.Task` | `*domain.Task` |
| TaskGraphManager | GetActiveByTask retorno | `*taskgraphmod.TaskGraph` | `*domain.TaskGraph` |
| RunLifecycleManager | Create retorno | `*runmod.Run` | `*domain.Run` |
| RunLifecycleManager | Start retorno | `*runmod.Run` | `*domain.Run` |
| AgentManager | FindOrCreate param/retorno | `agentmod.RuntimeType`, `*agentmod.Agent` | `domain.AgentRuntimeType`, `*domain.Agent` |
| SessionManager | Create/Connect/Stop retorno | `*agentsessionmod.AgentSession` | `*domain.AgentSession` |
| PromptPersistence | PersistComposedPrompt params | `*promptmod.ComposedPrompt`, `promptmod.PersistMetadata` | `*domain.ComposedPrompt`, `prompt.PersistMetadata` |
| PromptPersistence | PersistComposedPrompt retorno | `*promptmod.PreparedRunPrompt` | `*prompt.PreparedRunPrompt` |
| ReviewManager | Create param | `reviewmod.ValidationGate` | `domain.ReviewValidationGate` |
| ReviewManager | Create retorno | `*reviewmod.Review` | `*domain.Review` |
| TriggerEvaluator | EvaluateRun retorno | `*triggermod.Trigger` | `*domain.Trigger` |
| WorkUnitLister | ListByTaskGraph retorno | `[]workunitmod.WorkUnit` | `[]domain.WorkUnit` |
| DecomposeResult | Graph campo | `*taskgraphmod.TaskGraph` | `*domain.TaskGraph` |
| DecomposeResult | WorkUnits campo | `[]workunitmod.WorkUnit` | `[]domain.WorkUnit` |
| PreparedPrompt | PromptSnapshot campo | `*promptmod.PromptSnapshot` | `*domain.PromptSnapshot` |
| PreparedPrompt | ToolsetSnapshot campo | `*promptmod.ToolsetSnapshot` | `*domain.ToolsetSnapshot` |

---

### Outros Módulos com Imports Cross-Module

#### `internal/modules/agentsession/service.go`
- Importa `agentmod` (AgentReader interface usa `*agentmod.Agent`)
- **Conversão:** `*agentmod.Agent` → `*domain.Agent`

#### `internal/modules/run/service.go`
- Importa `taskmod` (TaskReader interface usa `*taskmod.Task`)
- Importa `workunitmod` (WorkUnitReader interface usa `*workunitmod.WorkUnit`)
- **Conversão:** `*taskmod.Task` → `*domain.Task`, `*workunitmod.WorkUnit` → `*domain.WorkUnit`

#### `internal/modules/run/service_relay.go`
- Importa `agentsessionmod` (SessionManager interface)
- **Conversão:** `*agentsessionmod.AgentSession` → `*domain.AgentSession`

#### `internal/modules/run/service_workunit.go`
- Importa `workunitmod` (WorkUnitReader e Status)
- **Conversão:** `workunitmod.Status` → `domain.WorkUnitStatus`, `*workunitmod.WorkUnit` → `*domain.WorkUnit`

#### `internal/modules/taskgraph/service.go`
- Importa `task` (TaskReader usa `*task.Task`)
- **Conversão:** `*task.Task` → `*domain.Task`

#### `internal/modules/trigger/service.go`
- Importa `runmod` (RunReader usa `*runmod.Run`)
- Importa `agentsessionmod` (AgentSessionReader usa `*agentsessionmod.AgentSession`)
- Importa `workunitmod` (WorkUnitReader usa `*workunitmod.WorkUnit`)
- **Conversão:** `*runmod.Run` → `*domain.Run`, `*agentsessionmod.AgentSession` → `*domain.AgentSession`, `*workunitmod.WorkUnit` → `*domain.WorkUnit`

#### `internal/modules/workunit/service.go`
- Importa `task` (TaskReader usa `*task.Task`)
- Importa `taskgraph` (TaskGraphManager usa `*taskgraph.TaskGraph`)
- **Conversão:** `*task.Task` → `*domain.Task`, `*taskgraph.TaskGraph` → `*domain.TaskGraph`

#### `internal/modules/workunit/service_create.go`
- Importa `task` e `taskgraph`
- **Conversão:** `*task.Task` → `*domain.Task`, `*taskgraph.TaskGraph` → `*domain.TaskGraph`

---

## Conflitos de Nomeação

### Problema: Múltiplos `Status` types

Cada módulo define seu próprio `Status` como `type Status string`. Na migração para `domain`, todos coexistirão no mesmo package, exigindo renomeação:

| Módulo | Tipo Atual | Novo Nome em domain |
|--------|-----------|---------------------|
| task | `Status` | `TaskStatus` |
| run | `Status` | `RunStatus` |
| workunit | `Status` | `WorkUnitStatus` |
| agentsession | `Status` | `AgentSessionStatus` |
| taskgraph | `Status` | `TaskGraphStatus` |
| review | `Status` | `ReviewStatus` |
| trigger | `Status` | `TriggerStatus` |

**Observação:** O teste `TestDomainImportIntegrity` já espera esses nomes (`TaskStatus`, `RunStatus`, etc.).

### Problema: `taskgraph.PlanWorkUnit` vs `workunit.WorkUnit`

- `PlanWorkUnit` foi criado em `taskgraph` para evitar import cycle com `workunit`.
- Quando `WorkUnit` migrar para `domain.WorkUnit`, `PlanWorkUnit` pode ser eliminado.
- Em `bootstrap/services.go`, as funções `planWorkUnitToDomain` e `workUnitToPlan` podem ser eliminadas (não haverá mais conversão necessária).
- Em `taskgraph/models.go`, remover `PlanWorkUnit` e usar `domain.WorkUnit` diretamente.

---

## Ordem de Migração Recomendada

Para minimizar dependências circulares e quebra incremental, recomenda-se a seguinte ordem:

1. **task** (`Task`, `TaskStatus`, `TaskPriority`, `TaskRiskLevel`)
   - É a raiz da hierarquia — outros módulos referenciam `TaskID` mas não o tipo `Task` diretamente em muitos lugares.

2. **agent** (`Agent`, `AgentRuntimeType`)
   - Poucas dependências — apenas `agentsession` e `orchestrator` usam.

3. **agentsession** (`AgentSession`, `AgentSessionStatus`)
   - Depende de `Agent` já migrado.

4. **run** (`Run`, `RunStatus`, `RunResult`)
   - Depende de `Task` e `WorkUnit` (indiretamente via campos TaskID/WorkUnitID, não pelos tipos).

5. **workunit** (`WorkUnit`, `WorkUnitStatus`)
   - Depende de `Task` e `TaskGraph` (via campos TaskID/TaskGraphID).

6. **taskgraph** (`TaskGraph`, `TaskGraphStatus`)
   - Depende de `WorkUnit` (via PlanWorkUnit que será eliminado).
   - Remover `TaskGraphNodeInfo`, `TaskGraphEdgeInfo`, `TaskGraphCreatedPayload` (já existem em domain).

7. **prompt** (`PromptSnapshot`, `PromptFragment`, `ToolsetSnapshot`, `ComposedPrompt`)
   - Poucas dependências externas.

8. **review** (`Review`, `ReviewStatus`, `ReviewValidationGate`)
   - Poucas dependências externas.

9. **trigger** (`Trigger`, `TriggerStatus`, `TriggerType`)
   - Depende de `Run` e `AgentSession` (via campos RunID/AgentSessionID, não pelos tipos).

10. **orchestrator** (atualizar interfaces para usar `domain.*`)
    - Depende de TODOS os tipos acima.

11. **bootstrap** (atualizar adapters para usar `domain.*`)
    - Depende de TODOS os tipos acima.

---

## Arquivos a Modificar na Task T5 (Code Refactor)

### Entity Types (migração de types)
- `internal/domain/types.go` — adicionar todos os entity types
- `internal/domain/doc.go` — atualizar documentação
- `internal/modules/task/models.go` — REMOVER (todos os tipos são shared)
- `internal/modules/run/models.go` — REMOVER
- `internal/modules/workunit/models.go` — REMOVER
- `internal/modules/agent/models.go` — manter `AgentStatus` (local), remover `Agent` e `RuntimeType`
- `internal/modules/agentsession/models.go` — REMOVER
- `internal/modules/taskgraph/models.go` — manter tipos locais, remover `TaskGraph`, `Status`, `TaskGraphNodeInfo`, `TaskGraphEdgeInfo`, `TaskGraphCreatedPayload`, `PlanWorkUnit`
- `internal/modules/prompt/models.go` — manter tipos locais, remover `PromptSnapshot`, `PromptFragment`, `ToolsetSnapshot`, `ComposedPrompt`
- `internal/modules/review/models.go` — manter `Decision`, `CriteriaChecked`, remover `Review`, `Status`, `ValidationGate`
- `internal/modules/trigger/models.go` — manter `AnomalyType`, `ResolutionAction`, `ThresholdConfig`, remover `Trigger`, `Status`, `Type`
- `internal/modules/orchestrator/models.go` — atualizar interfaces para `domain.*`

### Services (atualização de imports e types)
- `internal/modules/task/service.go` — atualizar para `domain.Task`, `domain.TaskStatus`, etc.
- `internal/modules/task/repository.go` — atualizar para `domain.Task`
- `internal/modules/run/service.go` — atualizar para `domain.Run`, `domain.Task`, `domain.WorkUnit`
- `internal/modules/run/service_workunit.go` — atualizar para `domain.WorkUnitStatus`
- `internal/modules/run/service_relay.go` — atualizar para `domain.AgentSession`
- `internal/modules/workunit/service.go` — atualizar para `domain.WorkUnit`, `domain.Task`, `domain.TaskGraph`
- `internal/modules/workunit/service_create.go` — atualizar para `domain.WorkUnit`, `domain.Task`, `domain.TaskGraph`
- `internal/modules/agentsession/service.go` — atualizar para `domain.AgentSession`, `domain.Agent`
- `internal/modules/taskgraph/service.go` — atualizar para `domain.TaskGraph`, `domain.Task`, `domain.WorkUnit`
- `internal/modules/taskgraph/gemini_planner.go` — atualizar para `domain.Task`
- `internal/modules/prompt/service.go` — atualizar para `domain.PromptSnapshot`, etc.
- `internal/modules/review/service.go` — atualizar para `domain.Review`
- `internal/modules/trigger/service.go` — atualizar para `domain.Trigger`, `domain.Run`, `domain.AgentSession`, `domain.WorkUnit`
- `internal/modules/orchestrator/service.go` — atualizar para `domain.*` em TODOS os métodos
- `internal/modules/orchestrator/service_cascade.go` — atualizar para `domain.RunStatus`, `domain.WorkUnitStatus`
- `internal/modules/orchestrator/validation.go` — atualizar para `domain.Agent`

### Bootstrap
- `internal/bootstrap/services.go` — atualizar TODOS os adapters para `domain.*`

### Tests
- `tests/architecture/domain_import_integrity_test.go` — tipos devem passar a existir em domain
- `tests/architecture/module_boundaries_test.go` — imports cross-module devem desaparecer
- Todos os testes unitários dos módulos devem ser atualizados

---

## Validação

### Checklist de Completude
- [x] Todos os 10 `models.go` foram inventariados
- [x] Cada tipo foi classificado como shared ou local
- [x] Todos os 26 entity types shared estão mapeados
- [x] Todos os 9 `Status` types têm nome único em domain
- [x] Imports em `bootstrap/services.go` estão mapeados
- [x] Imports em `orchestrator/models.go` estão mapeados
- [x] Imports cross-module em outros módulos estão mapeados
- [x] Ordem de migração foi definida para minimizar dependências
- [x] Conflitos de nomeação foram identificados e resolvidos
- [x] Documento é suficiente para que outro agente execute a migração

---

## Notas para o Executor da Task T5

1. **Não implementar nesta task** — este documento é APENAS o mapa.
2. **Usar renomeação gradual** — renomear `Status` → `TaskStatus` etc. pode ser feito módulo por módulo.
3. **Eliminar PlanWorkUnit** — quando `WorkUnit` estiver em domain, `PlanWorkUnit` deve ser eliminado e substituído por `domain.WorkUnit`.
4. **Atualizar interfaces primeiro** — os interfaces em `orchestrator/models.go` e `bootstrap/services.go` são os pontos de integração mais críticos.
5. **Testar a cada módulo** — após migrar cada módulo, rodar `go build ./...` e `go test ./tests/architecture/...`.
6. **Remover models.go vazios** — se `models.go` ficar sem nenhum tipo após a migração, remova o arquivo.
