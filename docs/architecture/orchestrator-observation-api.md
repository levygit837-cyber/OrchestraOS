# Observation API

## Visao Geral

A **Observation API** e a fronteira controlada atraves da qual o **Agente Orquestrador Inteligente** (e futuros consumidores autorizados) observa o estado do sistema.

Em vez de expor o Event Store brut — que seria proibitivo em tokens, ruido e risco de vazamento — o `OrchestratorService` produz **resumos estruturados, filtrados e classificados** do estado operacional.

## Principios

1. **Nunca expor eventos brut.** O consumidor recebe derivados, nao fonte.
2. **Filtrar por relevancia.** Apenas informacoes pertinentes a decisao em questao.
3. **Classificar por gravidade.** Alertas, anomalias e bloqueios devem ser destacados.
4. **Respeitar escopo e permissao.** O consumidor so ve o que tem autorizacao para ver.
5. **Sem segredos.** Tokens, chaves e credenciais nunca aparecem na observacao.
6. **Eficiencia de tokens.** O output deve ser conciso o suficiente para consumo por LLM.

## Interface

### Entrada

```go
type ObservationRequest struct {
    // Identificacao do solicitante
    AgentSessionID string
    AgentID        string

    // Escopo da observacao
    TaskID         *string   // se nil, observa todas as tasks ativas
    RunID          *string   // se nil, observa todas as runs da task
    WorkUnitID     *string   // filtro adicional

    // Tipo de informacao desejada
    IncludeSummary       bool
    IncludeActiveRuns    bool
    IncludePendingTools  bool
    IncludeAnomalies     bool
    IncludeCheckpoints   bool
    IncludeMemoryHints   bool
    IncludeBudget        bool
    IncludeSystemAlerts  bool

    // Filtros
    MaxRuns           int      // default: 10
    MaxCheckpoints    int      // default: 5
    MaxToolRequests   int      // default: 10
    TimeWindow        *time.Duration // apenas eventos recentes
    MinSeverity       string   // "info", "warning", "critical"
}
```

### Saida

```go
type Observation struct {
    // Metadados
    GeneratedAt   time.Time
    RequestID     string
    Scope         string   // "task", "run", "system"

    // Resumos
    TaskSummary       *TaskStatusSummary
    ActiveRuns        []RunObservation
    PendingApprovals  []ToolRequestSummary
    Anomalies         []AnomalyReport
    RecentCheckpoints []CheckpointSummary
    MemoryHints       []MemorySummary
    BudgetStatus      *BudgetReport
    SystemAlerts      []SystemAlert

    // Controle
    HasMoreData   bool     // true se ha mais dados que o limite
    NextCursor    *string  // para paginacao se necessario
}
```

## Componentes do Observation

### TaskStatusSummary

```go
type TaskStatusSummary struct {
    TaskID          string
    Title           string
    Status          string   // created, triaged, planned, running, etc.
    RiskLevel       string
    ProgressPercent int      // calculado com base no grafo

    WorkUnits       []WorkUnitSummary
    Dependencies    []DependencyStatus
    Blockers        []BlockerInfo
}

type WorkUnitSummary struct {
    WorkUnitID     string
    Status         string
    Profile        string
    Objective      string
    OwnedPaths     []string
    RunID          *string
    AgentSessionID *string
    CompletionPct  int
    TimeElapsed    time.Duration
}

type DependencyStatus struct {
    From      string
    To        string
    Type      string   // blocks, requires, conflicts
    Satisfied bool
}

type BlockerInfo struct {
    WorkUnitID string
    Reason     string
    Since      time.Time
    Severity   string
}
```

**Exemplo de output:**
```json
{
  "task_summary": {
    "task_id": "task_abc",
    "title": "Implementar autenticacao JWT",
    "status": "running",
    "risk_level": "medium",
    "progress_percent": 45,
    "work_units": [
      {
        "work_unit_id": "wu_001",
        "status": "completed",
        "profile": "code_worker",
        "objective": "Criar schema de usuarios",
        "completion_pct": 100
      },
      {
        "work_unit_id": "wu_002",
        "status": "running",
        "profile": "code_worker",
        "objective": "Implementar middleware JWT",
        "run_id": "run_456",
        "completion_pct": 30,
        "time_elapsed": "12m"
      }
    ],
    "blockers": []
  }
}
```

### RunObservation

```go
type RunObservation struct {
    RunID          string
    WorkUnitID     string
    Status         string
    RuntimeType    string
    AgentSessionID string

    // Progresso
    StartTime      time.Time
    LastEventAt    time.Time
    Duration       time.Duration
    EventsCount    int
    CheckpointsCount int

    // Estado atual do agente
    CurrentGoal    string
    GoalsCompleted []string
    GoalsPending   []string
    FilesModified  []string
    FilesRead      []string

    // Ferramentas
    ToolsRequested int
    ToolsApproved  int
    ToolsDenied    int
    PendingTools   []PendingToolSummary

    // Health
    HeartbeatAge   time.Duration
    LastCheckpointAge time.Duration
    StallDetected  bool
    LoopDetected   bool
}

type PendingToolSummary struct {
    ToolRequestID string
    Tool          string
    Risk          string
    Reason        string
    WaitingSince  time.Time
}
```

### ToolRequestSummary

```go
type ToolRequestSummary struct {
    ToolRequestID string
    RunID         string
    WorkUnitID    string
    AgentID       string
    Tool          string
    Risk          string
    Scope         string
    Reason        string
    RequestedAt   time.Time
    WaitingDuration time.Duration
}
```

### AnomalyReport

```go
type AnomalyReport struct {
    Type        string   // stall, loop, resource_exhaustion, drift, conflict
    Severity    string   // info, warning, critical
    RunID       *string
    WorkUnitID  *string
    TaskID      string
    Description string
    DetectedAt  time.Time
    SuggestedAction string
    Metrics     map[string]interface{}
}
```

**Tipos de anomalia detectados:**

| Tipo | Descricao | Metricas |
|------|-----------|----------|
| `stall` | Agente sem progresso por tempo excessivo | `minutes_without_checkpoint`, `last_event_type` |
| `loop` | Agente repetindo acoes sem avancar | `repeated_tools`, `repeated_goals`, `cycle_length` |
| `resource_exhaustion` | Consumo excessivo de tokens/tempo | `tokens_used`, `tokens_budget`, `time_elapsed` |
| `drift` | Agente desviando do objetivo original | `original_goal`, `current_goal`, `similarity_score` |
| `conflict` | Dois agentes com resultados contraditorios | `run_ids`, `conflict_type`, `severity` |
| `heartbeat_missing` | Heartbeat nao recebido no prazo | `expected_interval`, `actual_interval` |
| `dependency_timeout` | Dependencia nao satisfeita no prazo | `dependency`, `waiting_since`, `timeout_configured` |

### CheckpointSummary

```go
type CheckpointSummary struct {
    CheckpointID   string
    RunID          string
    Sequence       int
    Timestamp      time.Time
    Goal           string
    GoalsCompleted []string
    GoalsPending   []string
    Blockers       []string
    Risks          []string
    EvidenceCount  int
}
```

### MemorySummary

```go
type MemorySummary struct {
    MemoryID    string
    Type        string
    Domain      string
    Title       string
    Content     string
    Confidence  float64
    EvidenceRef string
    Relevance   float64   // score de matching com o contexto atual
}
```

### BudgetReport

```go
type BudgetReport struct {
    TaskID           string
    TotalRuns        int
    ActiveRuns       int

    TokensUsed       int64
    TokensBudget     int64
    TokensRemaining  int64

    TimeElapsed      time.Duration
    TimeBudget       time.Duration
    TimeRemaining    time.Duration

    CostEstimate     float64   // em moeda, se configurado
    CostBudget       float64
}
```

### SystemAlert

```go
type SystemAlert struct {
    Level       string   // info, warning, error, critical
    Source      string   // scheduler, policy_engine, sandbox_manager, etc.
    Message     string
    Timestamp   time.Time
    RelatedTask *string
    RelatedRun  *string
}
```

## Algoritmos de Deteccao de Anomalia

### Stall Detection

```
stall_detected = true se:
  (now - last_checkpoint_at) > stall_threshold
  E (now - last_non_heartbeat_event) > stall_threshold
  E run.status == "running"

stall_threshold = max(5 min, 2 * avg_checkpoint_interval)
```

### Loop Detection

```
loop_detected = true se:
  sequencia de N eventos repete em ciclo
  OU mesmo goal aparece em checkpoints consecutivos sem progresso
  OU mesma ferramenta chamada com mesmos argumentos > M vezes

N = 3 (configuravel)
M = 2 (configuravel)
```

### Drift Detection

```
drift_detected = true se:
  similaridade_semantica(original_goal, current_goal) < drift_threshold
  E agente nao registrou justificativa para mudanca de escopo

drift_threshold = 0.6 (configuravel)
```

## Ciclo de Atualizacao

A Observation API pode operar em dois modos:

### Modo Pull (Sincrono)

O Agente Inteligente solicita observacao explicitamente via ferramenta.

```go
func (s *OrchestratorService) GetObservation(req ObservationRequest) (*Observation, error)
```

**Uso**: Inicio de ativacao, entre decisoes, quando mais contexto e necessario.

### Modo Push (Assincrono)

O OrchestratorService notifica o Agente Inteligente quando anomalias criticas sao detectadas.

```go
type ObservationPush struct {
    Trigger     string   // anomaly_detected, tool_pending, task_failed
    Urgency     string   // low, medium, high
    Snapshot    Observation
}
```

**Uso**: Ativacao automatica do Agente Inteligente sem polling.

## Seguranca e Privacidade

1. **Sanitizacao**: Toda string e verificada para nao conter segredos antes de retornar.
2. **Escopo**: Solicitante so ve tasks/runs que tem permissao para observar.
3. **Rate Limiting**: Maximo de N observacoes por minuto por sessao.
4. **Audit Log**: Toda observacao solicitada e registrada com timestamp, escopo e tamanho.

## Implementacao

A Observation API deve ser implementada como metodo do `OrchestratorService`:

```go
// internal/services/orchestrator_service.go

func (s *OrchestratorService) GetObservation(ctx context.Context, req ObservationRequest) (*Observation, error) {
    // 1. Validar permissao da sessao
    // 2. Coletar dados dos servicos de dominio
    // 3. Detectar anomalias
    // 4. Classificar e resumir
    // 5. Sanitizar
    // 6. Retornar estrutura
}
```

## Dependencias

- `TaskService` / `WorkUnitService` — estado do grafo
- `RunService` — progresso das runs
- `AgentSessionService` — checkpoints e heartbeats
- `EventService` — eventos recentes
- `PolicyService` (futuro) — pendencias de aprovacao
- `MemoryService` (futuro) — memorias relevantes

## Referencias

- `docs/adr/0023-hybrid-intelligent-orchestrator.md`
- `docs/architecture/intelligent-orchestrator-agent.md`
- `docs/architecture/orchestrator-intervention-protocol.md`
- `docs/architecture/multi-agent-coordination.md`
- `docs/architecture/communication-protocol.md`
