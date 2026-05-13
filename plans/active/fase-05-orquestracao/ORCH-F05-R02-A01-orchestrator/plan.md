# 🎼 PLANO — ORCH-F05-R02-A01: OrchestratorService.RunTask()

**ID:** ORCH-F05-R02-A01  
**Fase:** Fase 5 — Orquestração Automatizada (R02)  
**Agente:** Agente 1 (OrchestratorService Core)  
**Ferramenta:** Windsurf  
**Independência:** Alta — não depende de código do Agente 2 (CLI). A interface contratual é definida abaixo; o Agente 2 a consumirá.  

---

## Contexto do Projeto

Você está trabalhando no **OrchestraOS**, um Sistema de Orquestração de Agentes escrito em Go. O repositório está em `/home/levybonito/Documentos/OrchestraOS`.

A Fase 5 R01 foi concluída: os módulos `agent/`, `review/` e `trigger/` estão implementados e testados. O próximo passo crítico é criar o **OrchestratorService** — o componente central que conecta todos os serviços de domínio em um fluxo automatizado de ponta a ponta.

Sem o OrchestratorService, o sistema exige intervenção manual em cada passo (`task create` → `task graph create` → `run start`). Com ele, um único comando (`task run --id <task_id>`) orquestra tudo.

---

## Documentação Obrigatória (LEIA ANTES DE CODAR)

1. `AGENTS.md` — regras de trabalho no repositório
2. `docs/adr/0020-orchestrator-service.md` — decisão arquitetural do OrchestratorService
3. `docs/adr/0021-agent-service.md` — AgentService e registro de agentes
4. `docs/adr/0023-hybrid-intelligent-orchestrator.md` — arquitetura híbrida
5. `docs/implementation/roadmap.md` — seção "Fase 5: Orquestração Automatizada"
6. `docs/canvas/project-canvas.md` — visão e premissas
7. `internal/bootstrap/services.go` — como os serviços existentes são criados
8. `internal/core/orchestration/runtime_relay.go` — como o relay de eventos funciona

---

## Estado Atual (O que Já Existe)

| Componente | Local | Status |
|---|---|---|
| TaskService | `internal/modules/task/service.go` | ✅ Create, Get, List, Complete |
| TaskGraphService | `internal/modules/taskgraph/service.go` | ✅ Decompose, ListByTask |
| RunService | `internal/modules/run/service.go` | ✅ Create, Start, Complete, Fail, Timeout |
| AgentService | `internal/modules/agent/service.go` | ✅ Create, GetByID, FindOrCreate |
| AgentSessionService | `internal/modules/agentsession/service.go` | ✅ Create (valida AgentID), Connect, Stop, Fail |
| PromptService | `internal/modules/prompt/service.go` | ✅ PrepareRunPrompt |
| PromptOrchestrator | `internal/core/orchestration/prompt_orchestrator.go` | ✅ Cross-module prompt prep |
| ReviewService | `internal/modules/review/service.go` | ✅ Create, Start, SubmitVerdict |
| TriggerService | `internal/modules/trigger/service.go` | ✅ EvaluateRun, EvaluateSession |
| RuntimeEventRelay | `internal/core/orchestration/runtime_relay.go` | ✅ Run(ctx, runtime, config) → status |
| FakeRuntime | `internal/modules/agent/fake_runtime.go` | ✅ Simulação determinística |
| GeminiRuntime | `internal/modules/agent/gemini_runtime.go` | ✅ Inference real |
| Bootstrap | `internal/bootstrap/services.go` | ✅ Factories para todos os serviços acima |

---

## O que Você Deve Implementar

### 1. Módulo `internal/modules/orchestrator/`

Criar a estrutura completa do módulo seguindo o padrão dos outros módulos (`internal/modules/agent/`, `internal/modules/review/`, etc.):

```
internal/modules/orchestrator/
├── doc.go          // Package documentation
├── contract.go     // Constants, event types
├── models.go       // Types específicos do módulo
├── service.go      // OrchestratorService e RunTask()
└── validation.go   // Validações de input
```

#### `service.go` — Requisitos

```go
package orchestrator

// Dependencies contém todas as dependências do OrchestratorService.
// Use interfaces quando necessário para evitar importações diretas.
type Dependencies struct {
    DB                  *sql.DB
    TaskService         TaskServiceReader
    TaskGraphService    TaskGraphManager
    RunService          RunLifecycleManager
    AgentService        AgentManager
    AgentSessionService SessionManager
    PromptOrchestrator  PromptPreparer
    ReviewService       ReviewManager
    TriggerService      TriggerEvaluator
    RuntimeEventRelay   func(db *sql.DB) *orchestration.RuntimeEventRelay
    NewFakeRuntime      func() agent.Runtime
    NewGeminiRuntime    func() agent.Runtime
}

type RunTaskOptions struct {
    RuntimeType     string // fake | gemini | codex_cli
    PlannerStrategy string // local_heuristic_v1 | llm_gemini_v1
    MaxSteps        int    // default: 10
    TimeoutSeconds  int    // default: 300
}

type RunTaskResult struct {
    TaskID    string
    RunIDs    []string
    Status    string // completed | failed | partial
    ReviewIDs []string
}

func NewService(deps Dependencies) *Service

func (s *Service) RunTask(ctx context.Context, taskID string, options RunTaskOptions) (*RunTaskResult, error)
```

**Fluxo de `RunTask`:**

1. **Obter task** via `TaskService.GetByID()`
2. **Decompor** via `TaskGraphService.Decompose()` se não houver grafo ativo
   - Usar `options.PlannerStrategy` (default: `local_heuristic_v1`)
3. **Listar work units** do grafo ativo via `WorkUnitService`
4. **Ordenar topologicamente** as work units (respeitar dependências do DAG)
5. **Para cada work unit** (execução sequencial):
   a. `RunService.Create()`
   b. `RunService.Start()`
   c. Determinar `profile` da work unit (default: `code_worker`)
   d. `AgentService.FindOrCreate(profile, options.RuntimeType)`
   e. `AgentSessionService.Create()` (valida AgentID automaticamente)
   f. `AgentSessionService.Connect()`
   g. `PromptOrchestrator.PrepareRunPrompt()`
   h. Instanciar runtime conforme `options.RuntimeType`:
      - `"fake"` → `NewFakeRuntime()`
      - `"gemini"` → `NewGeminiRuntime()`
   i. `runtime.Start(ctx, config)` com prompt preparado
   j. Iniciar `RuntimeEventRelay.Run(ctx, runtime, config)`
   k. Aguardar conclusão (relay retorna status final)
   l. Se work unit tem `ValidationGate` ≠ `none`:
      - `ReviewService.Create()` com gate type
      - NÃO aguardar veredicto no primeiro corte (registrar e continuar)
   m. `TriggerService.EvaluateRun()` para detectar anomalias
   n. Se run falhou → registrar falha e continuar (ou abortar; primeiro corte: continuar)
6. **Quando todas as WUs terminarem**:
   - Se todas completaram → `TaskService.Complete()`
   - Se alguma falhou → `TaskService` permanece no estado atual (ou `failed` se houver política)
7. **Retornar** `RunTaskResult` com todos os IDs

**Regras importantes:**
- Execução **sequencial** no primeiro corte (1 WU por vez)
- Timeout por work unit = `options.TimeoutSeconds`
- Context com timeout para cada runtime
- Todo erro deve ser registrado como evento quando possível
- Nunca acesse repositórios diretamente — use apenas serviços de domínio

#### `validation.go`

- Validar `RuntimeType`: `fake`, `gemini`, `codex_cli`
- Validar `PlannerStrategy`: `local_heuristic_v1`, `llm_gemini_v1`
- `MaxSteps` > 0, `TimeoutSeconds` > 0

#### `events.go`

Definir event types:
- `orchestrator.task_started`
- `orchestrator.work_unit_started`
- `orchestrator.work_unit_completed`
- `orchestrator.work_unit_failed`
- `orchestrator.task_completed`
- `orchestrator.task_failed`

#### `models.go`

Types auxiliares se necessário (ex: `WorkUnitExecutionResult`).

#### `doc.go`

Documentação do pacote seguindo padrão dos outros módulos.

---

### 2. Atualizar `internal/bootstrap/services.go`

Adicionar factory:

```go
func OrchestratorService(db *sql.DB) *orchestrator.Service {
    return orchestrator.NewService(orchestrator.Dependencies{
        DB:                  db,
        TaskService:         TaskService(db),
        TaskGraphService:    TaskGraphService(db),
        RunService:          RunService(db),
        AgentService:        AgentService(db),
        AgentSessionService: AgentSessionService(db),
        PromptOrchestrator:  NewPromptOrchestrator(db, PromptService(db)),
        ReviewService:       ReviewService(db),
        TriggerService:      TriggerService(db),
        RuntimeEventRelay:   RuntimeEventRelay,
        NewFakeRuntime:      func() agent.Runtime { return agent.NewFakeRuntime() },
        NewGeminiRuntime:    func() agent.Runtime { return agent.NewGeminiRuntime() },
    })
}
```

**Cuidado:** Evitar ciclos de importação. Se necessário, use interfaces no pacote `orchestrator` em vez de tipos concretos.

---

### 3. Testes de Integração

Criar `tests/integration/orchestrator_service_test.go`:

```go
func TestOrchestratorService_RunTask_WithFakeRuntime(t *testing.T)
```

**Cenário:**
1. Criar task com 2 critérios de aceite
2. Chamar `OrchestratorService.RunTask()` com `RuntimeType: "fake"`
3. Verificar:
   - Task transita para `completed`
   - 2 runs foram criados (1 por WU)
   - 2 agent sessions foram criadas
   - Work units foram executadas na ordem correta
   - Eventos `orchestrator.*` foram emitidos

**Cenário de falha:**
1. Criar task
2. Simular falha em uma work unit
3. Verificar que o resultado reflete a falha

---

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

Você deve executar esta tarefa em ciclos curtos usando o arquivo de checklist persistente.

**Caminho do checklist:** `plans/active/fase-05-orquestracao/ORCH-F05-R02-A01-orchestrator/checklist.md`

**A cada iteração:**
1. **LER** o checklist para identificar o próximo item pendente
2. **EXECUTAR** o item (código, teste, refactor)
3. **VALIDAR** o item (testes passam? comportamento correto?)
4. **ATUALIZAR** o checklist marcando o item como concluído
5. **CONTINUAR** para o próximo item

**Regras do Ralph Loop:**
- Nunca pule um item sem marcá-lo no checklist
- Se encontrar bloqueio, adicione uma nota na seção "Notas de Progresso"
- Se precisar adicionar itens ao checklist, faça-o (são raras exceções)
- Ao final de cada ciclo significativo, faça um commit pequeno
- O checklist é sua fonte de verdade de progresso

---

## Fronteiras de Isolamento

### TOCAR
- `internal/modules/orchestrator/` (novo diretório)
- `internal/bootstrap/services.go` (adicionar factory)
- `tests/integration/orchestrator_service_test.go` (novo)
- `internal/domain/types.go` (se precisar adicionar types)

### EVITAR
- `cmd/orchestraos/cmd/` (é responsabilidade do Agente 2)
- `internal/modules/agent/` (não modifique runtime existente)
- `internal/modules/review/` (não modifique service existente)
- `internal/modules/trigger/` (não modifique service existente)
- `internal/modules/task/`, `run/`, `workunit/`, `agentsession/`, `prompt/` (não modifique services existentes)

---

## Interface Contratual (Definida pelo Orquestrador)

Esta interface é consumida pelo Agente 2 (CLI). Você é responsável por implementá-la exatamente assim:

```go
package orchestrator

// Service é o ponto de entrada único para orquestração automatizada.
type Service struct { /* implementação interna */ }

func NewService(deps Dependencies) *Service

// RunTask executa uma task completa de ponta a ponta.
func (s *Service) RunTask(ctx context.Context, taskID string, options RunTaskOptions) (*RunTaskResult, error)

type RunTaskOptions struct {
    RuntimeType     string // "fake" | "gemini" | "codex_cli"
    PlannerStrategy string // "local_heuristic_v1" | "llm_gemini_v1"
    MaxSteps        int    // padrão: 10
    TimeoutSeconds  int    // padrão: 300
}

type RunTaskResult struct {
    TaskID    string
    RunIDs    []string // runs criados, na ordem de execução
    Status    string   // "completed" | "failed" | "partial"
    ReviewIDs []string // reviews criados para gates
}
```

---

## Critérios de Aceite

- [ ] `OrchestratorService.RunTask()` executa fluxo completo de uma task com múltiplas work units
- [ ] Agentes são registrados no banco com perfil e runtime type (via `AgentService.FindOrCreate`)
- [ ] `AgentSession.AgentID` referencia agente existente (validação já existe)
- [ ] Work units são executadas na ordem topológica correta
- [ ] Eventos do runtime são roteados via `RuntimeEventRelay` para serviços de domínio
- [ ] `Run` transita de `created` → `running` → `completed` automaticamente
- [ ] `Task` transita para `completed` quando todas as WUs terminam
- [ ] Teste E2E com FakeRuntime valida todo o fluxo sem dependência externa
- [ ] `go test ./...` passa sem regressões
- [ ] `go build ./...` compila sem erros
- [ ] Nenhum import direto de repositórios cruzados (apenas serviços)

---

## Entrega Final

Ao concluir:
1. Atualize o checklist marcando todos os itens
2. Execute `go test ./...` e `go build ./...`
3. Execute `./scripts/safe-commit.sh "ORCH-F05-R02-A01: Implement OrchestratorService.RunTask"`
4. Informe ao usuário o nome da branch criada e o estado dos testes
