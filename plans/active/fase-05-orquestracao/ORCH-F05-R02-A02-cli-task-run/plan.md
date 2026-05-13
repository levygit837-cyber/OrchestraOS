# 🎼 PLANO — ORCH-F05-R02-A02: CLI `task run` + Testes E2E

**ID:** ORCH-F05-R02-A02  
**Fase:** Fase 5 — Orquestração Automatizada (R02)  
**Agente:** Agente 2 (CLI + E2E)  
**Ferramenta:** Kimi-CLI  
**Independência:** Depende da interface contratual do OrchestratorService (definida abaixo). Não depende da implementação interna do Agente 1.  

---

## Contexto do Projeto

Você está trabalhando no **OrchestraOS**, um Sistema de Orquestração de Agentes escrito em Go. O repositório está em `/home/levybonito/Documentos/OrchestraOS`.

A Fase 5 R01 foi concluída (agent, review, trigger modules). A R02 está criando o OrchestratorService (Agente 1) e a CLI `task run` (você).

Sua missão é:
1. Criar o comando `task run` na CLI — o ponto de entrada único para o usuário executar uma task completa
2. Refatorar o `run start` existente para usar `AgentService.FindOrCreate()` (conforme ADR 0021)
3. Criar testes E2E que validam o fluxo completo via OrchestratorService

---

## Documentação Obrigatória (LEIA ANTES DE CODAR)

1. `AGENTS.md` — regras de trabalho no repositório
2. `docs/adr/0020-orchestrator-service.md` — RunTask e CLI
3. `docs/adr/0021-agent-service.md` — FindOrCreate, validação de AgentID
4. `docs/implementation/roadmap.md` — seção "Fase 5" (CLI `task run`)
5. `cmd/orchestraos/cmd/task.go` — estrutura atual dos comandos de task
6. `cmd/orchestraos/cmd/run.go` — `run start` atual (será refatorado)
7. `internal/bootstrap/services.go` — factories existentes

---

## Interface Contratual (Consumida por Você)

O Agente 1 está implementando o `OrchestratorService` nesta interface exata. Você deve programar a CLI assumindo que ela já existe:

```go
package orchestrator

// Service é o ponto de entrada único para orquestração automatizada.
type Service struct { /* implementação do Agente 1 */ }

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

**Como você acessará:**
```go
orchService := bootstrap.OrchestratorService(getDB())
result, err := orchService.RunTask(ctx, taskID, orchestrator.RunTaskOptions{...})
```

Se `bootstrap.OrchestratorService()` ainda não existir (Agente 1 ainda não entregou), crie um **stub temporário** em `internal/bootstrap/services.go` para permitir compilação e testes da CLI. O stub será substituído pela implementação real do Agente 1 no merge.

---

## O que Você Deve Implementar

### 1. CLI `task run`

Adicionar ao `cmd/orchestraos/cmd/task.go` (ou criar `task_run.go` no mesmo pacote):

```go
var taskRunCmd = &cobra.Command{
    Use:   "run",
    Short: "Run a task end-to-end via OrchestratorService",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 1. Ler flags
        // 2. Validar taskID
        // 3. Resolver defaults (planner strategy do env var, runtime fake, etc.)
        // 4. Chamar OrchestratorService.RunTask()
        // 5. Exibir progresso e resultado
    },
}
```

**Flags obrigatórias e opcionais:**

| Flag | Tipo | Obrigatório | Default | Descrição |
|---|---|---|---|---|
| `--id` | string | ✅ | — | Task ID a executar |
| `--runtime` | string | ❌ | `fake` | Tipo de runtime: `fake`, `gemini`, `codex_cli` |
| `--planner` | string | ❌ | `local_heuristic_v1` | Estratégia de planner. Também lê `ORCHESTRAOS_PLANNER_STRATEGY` |
| `--max-steps` | int | ❌ | `10` | Máximo de steps por work unit |
| `--timeout` | int | ❌ | `300` | Timeout em segundos por work unit |

**Comportamento de saída no terminal:**

```
$ orchestraos task run --id abc-123 --runtime fake
Orchestrating task: abc-123
  → Decomposing into work units...
  → Executing 3 work units sequentially
  [1/3] WU: implement-domain-types     RUN: run-xxx  AGENT: code_worker(fake)  → completed
  [2/3] WU: implement-service          RUN: run-yyy  AGENT: code_worker(fake)  → completed
  [3/3] WU: write-tests                RUN: run-zzz  AGENT: code_worker(fake)  → completed
  → Task completed successfully
Runs: 3 | Agents: 3 | Reviews: 0 | Time: 2.3s
```

**Regras:**
- Se task não existir → erro claro
- Se task já estiver `completed` ou `failed` → avisar e confirmar (ou erro; primeiro corte: erro)
- Se `--planner` não for informado, ler `ORCHESTRAOS_PLANNER_STRATEGY`
- Progresso deve ser exibido em tempo real (via callback `OnEvent` do relay, se possível; primeiro corte: status final de cada WU)

### 2. Refatorar `run start` para usar `AgentService.FindOrCreate()`

No arquivo `cmd/orchestraos/cmd/run.go`, o `run start` atual gera `AgentID` inline:

```go
agentID := fmt.Sprintf("agent-%s", uuid.New().String()[:8])
```

**Migre para:**
```go
agentService := bootstrap.AgentService(getDB())
agent, err := agentService.FindOrCreate(ctx, profile, domain.AgentRuntimeType(runtimeType))
if err != nil { ... }
agentID := agent.ID
```

Onde `profile` deve ser determinado a partir do work unit (default: `code_worker`). Verifique se `WorkUnit` tem campo `AssignedAgentProfile`.

### 3. Testes E2E

Criar `tests/integration/orchestrator_e2e_test.go`:

```go
func TestOrchestrator_RunTask_FullFlow(t *testing.T)
```

**Cenário:**
1. Criar task via `TaskService.Create()`
2. Criar task graph via `TaskGraphService.Decompose()`
3. Chamar `OrchestratorService.RunTask()` com `RuntimeType: "fake"`
4. Verificar:
   - `result.Status == "completed"`
   - `len(result.RunIDs) == len(workUnits)`
   - Task status no banco é `completed`
   - Cada run existe e tem status `completed`
   - Cada agent session referencia um agente válido

**Cenário de ordenação topológica:**
1. Criar task graph com dependências (WU B depende de WU A)
2. Executar via OrchestratorService
3. Verificar que runs foram criados na ordem A → B

**Cenário com Review Gate:**
1. Criar task graph com work unit que tem `ValidationGate: hard`
2. Executar via OrchestratorService
3. Verificar que `result.ReviewIDs` não está vazio

---

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

Você deve executar esta tarefa em ciclos curtos usando o arquivo de checklist persistente.

**Caminho do checklist:** `plans/active/fase-05-orquestracao/ORCH-F05-R02-A02-cli-task-run/checklist.md`

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
- `cmd/orchestraos/cmd/task.go` (adicionar subcomando `run`)
- `cmd/orchestraos/cmd/task_run.go` (novo arquivo, recomendado)
- `cmd/orchestraos/cmd/run.go` (refatorar `run start` para FindOrCreate)
- `tests/integration/orchestrator_e2e_test.go` (novo)
- `internal/bootstrap/services.go` (apenas se precisar do stub de OrchestratorService)

### EVITAR
- `internal/modules/orchestrator/` — é responsabilidade do Agente 1
- `internal/modules/agent/service.go` — não modifique
- `internal/modules/review/service.go` — não modifique
- `internal/modules/trigger/service.go` — não modifique
- Qualquer outro `internal/modules/*/` — não modifique services existentes

---

## Critérios de Aceite

- [ ] Comando `orchestraos task run --id <task_id>` funciona e delega a `OrchestratorService.RunTask()`
- [ ] Flags `--runtime`, `--planner`, `--max-steps`, `--timeout` funcionam corretamente
- [ ] Flag `--planner` usa `ORCHESTRAOS_PLANNER_STRATEGY` como fallback
- [ ] `run start` usa `AgentService.FindOrCreate()` em vez de gerar AgentID inline
- [ ] Teste E2E valida fluxo completo: task → graph → runs → complete
- [ ] Teste E2E valida ordenação topológica de work units
- [ ] Teste E2E valida que agentes são registrados com perfil correto
- [ ] `go test ./...` passa sem regressões
- [ ] `go build ./...` compila sem erros
- [ ] CLI exibe progresso legível das work units

---

## Entrega Final

Ao concluir:
1. Atualize o checklist marcando todos os itens
2. Execute `go test ./...` e `go build ./...`
3. Execute `./scripts/safe-commit.sh "ORCH-F05-R02-A02: Implement CLI task run + E2E tests"`
4. Informe ao usuário o nome da branch criada e o estado dos testes

---

## Stub Temporário (se necessário)

Se o Agente 1 ainda não entregou o `OrchestratorService`, adicione temporariamente em `internal/bootstrap/services.go`:

```go
// OrchestratorService stub — será substituído pela implementação real do Agente 1.
func OrchestratorService(db *sql.DB) *orchestrator.Service {
    // TODO: substituir por implementação real quando ORCH-F05-R02-A01 for mergeado
    return nil
}
```

**Atenção:** Marque com `// TODO(ORCH-F05-R02-A01)` para facilitar o merge posterior.
