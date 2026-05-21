# Coordenacao Multi-Agente

## Visao Geral

O OrchestraOS executa multiplos agentes em paralelo sobre o mesmo projeto. A **Coordenacao Multi-Agente** define os mecanismos atraves dos quais o Orchestrator garante que esses agentes colaborem eficientemente, sem conflitos, deadlocks ou desperdicio de recursos.

Esta documentacao trata da coordenacao **entre agentes executores**. O Agente Orquestrador Inteligente e o mediador estrategico; os agentes executores sao os workers que precisam ser coordenados.

## Principios

1. **Nenhum agente e autonomo em relacao aos outros.** Todo agente opera dentro de um plano global controlado.
2. **Comunicacao mediada obrigatoria.** Agentes nunca se comunicam diretamente.
3. **Dependencias explicitas.** O Task Graph declara todas as dependencias antes da execucao.
4. **Recursos compartilhados controlados.** Paths, ferramentas e dados compartilhados sao gerenciados pelo Orchestrator.
5. **Progresso observavel.** O estado de todo agente e visivel ao Orchestrator em tempo real.

## Modelo de Coordenacao

O OrchestraOS usa um modelo hibrido de coordenacao:

| Mecanismo | Tipo | Uso |
|-----------|------|-----|
| **Task Graph (DAG)** | Estatico / Declarativo | Define dependencias, ordem e ownership antes da execucao |
| **Barrier Synchronization** | Dinamico / Temporal | Sincroniza fases de execucao entre agentes |
| **Validation Gate** | Dinamico / Qualidade | Review-Session obrigatoria antes de liberar dependencias |
| **Shared Information Board** | Dinamico / Semantico | Compartilha descobertas relevantes entre agentes |
| **Contract Net** | Dinamico / Negociacao | Atribui work units a agentes com base em capacidade |
| **Deadlock Detection** | Reativo / Protetivo | Detecta e resolve esperas circulares em runtime |

## 1. Task Graph como Plano Global

O Task Graph e a base estatica da coordenacao. Ele e criado antes do spawn de qualquer agente e define:

- **Nodes**: WorkUnits com objetivo, escopo e criterios de aceite.
- **Edges**: Dependencias do tipo:
  - `blocks`: WU-B so pode comecar apos WU-A completar
  - `requires`: WU-B precisa de artefato produzido por WU-A
  - `conflicts`: WU-A e WU-B nao podem executar em paralelo (shared resource)
  - `informs`: WU-B se beneficia de informacao de WU-A, mas nao depende

### Grafo de Espera (Wait-For Graph)

Em runtime, o Orchestrator mantem um **Wait-For Graph** derivado do Task Graph:

```
WU-A --blocks--> WU-B --requires--> WU-C
  |
  +--conflicts--> WU-D
```

Este grafo e usado para:
- Determinar quais WUs estao prontas para execucao
- Detectar deadlocks (ciclos no grafo de espera)
- Calcular caminho critico

### Caminho Critico

O Orchestrator calcula o caminho critico do Task Graph para:
- Priorizar WUs que bloqueiam o maior numero de outras WUs
- Estimar tempo total de execucao
- Identificar gargalos

```go
type CriticalPath struct {
    Path          []string       // ordem de WUs no caminho critico
    TotalDuration time.Duration  // soma estimada
    Bottlenecks   []string       // WUs com maior slack negativo
}
```

## 2. Barrier Synchronization

Barreiras sao pontos no plano onde o Orchestrator **bloqueia a progressao ate que todas as WUs de uma fase completem**.

### Tipos de Barreira

#### Barreira Dura (Hard Barrier)

Nenhuma WU da proxima fase pode comecar ate que **todas** as WUs da fase atual completem.

**Uso**: Validacoes que precisam do estado completo, integracoes que dependem de todos os modulos.

```text
Fase 1: [WU-schema] [WU-repository] [WU-service]
           |              |              |
           +--------------+--------------+
                          |
                    HARD BARRIER
                          |
                          v
Fase 2: [WU-integration-test]
```

#### Barreira Suave (Soft Barrier)

WUs da proxima fase podem comecar quando **N de M** WUs da fase anterior completam, desde que as WUs restantes nao sejam no caminho critico.

**Uso**: Execucao otimista onde algumas WUs sao independentes.

#### Barreira de Checkpoint (Checkpoint Barrier)

Todos os agentes devem atingir um checkpoint sincronizado antes de prosseguir.

**Uso**: Revisao de progresso, injecao de memoria compartilhada, decisao de replanejamento.

### Implementacao de Barreira

```go
type Barrier struct {
    ID            string
    TaskID        string
    Phase         int
    Type          string   // hard, soft, checkpoint
    RequiredWUs   []string
    CompletedWUs  []string
    SoftThreshold *int     // para soft barrier: minimo de WUs
    Timeout       time.Duration
    CreatedAt     time.Time
}

func (b *Barrier) IsSatisfied() bool {
    if b.Type == "hard" || b.Type == "checkpoint" {
        return len(b.CompletedWUs) == len(b.RequiredWUs)
    }
    if b.Type == "soft" {
        return len(b.CompletedWUs) >= *b.SoftThreshold
    }
    return false
}
```

### Protocolo de Barreira

```text
1. Orchestrator define barreira no Task Graph (ou adiciona dinamicamente)
2. WUs da fase anterior executam
3. Ao completar, cada WU emite evento `work_unit.completed`
4. Orchestrator verifica se barreira e satisfeita
5. Se sim: libera WUs da proxima fase
6. Se timeout: decide entre liberar parcial (soft) ou falhar (hard)
```

## 3. Shared Information Board (Quadro de Informacoes Compartilhadas)

Um mecanismo controlado pelo Orchestrator para compartilhar **descobertas** entre agentes sem violar isolamento.

### Conceito

Quando um agente descobre algo que pode ser util a outro, ele publica no Board. O Orchestrator decide:
- Se a informacao e relevante para outros agentes
- Quais agentes devem receber
- Quando injetar (checkpoint, notificacao, etc.)

### Tipos de Informacao Compartilhavel

| Tipo | Exemplo | Quem recebe |
|------|---------|-------------|
| `api_change` | "Nova interface adicionada em auth.go" | Agentes que usam auth |
| `dependency_update` | "Pacote X atualizado para v2" | Agentes que importam X |
| `pattern_discovered` | "Funcao util Y disponivel em utils.go" | Agentes na mesma task |
| `risk_alert` | "Teste Z e instavel" | Todos os agentes da task |
| `opportunity` | "Codigo duplicado detectado" | Agente `reviewer` |

### Ciclo de Vida de uma Informacao

```text
1. Agente A descobre informacao relevante
2. Agente A emite evento `agent.insight.discovered`
3. Orchestrator avalia:
   a. Relevancia para outros agentes?
   b. Violacao de isolamento?
   c. Conflito com informacao existente?
4. Se aprovado, publica no Board com `insight_id`
5. Orchestrator notifica agentes relevantes no proximo checkpoint
6. Agentes que recebem incorporam ou ignoram
7. Insight expira quando task completa ou quando supersedido
```

### Estrutura de um Insight

```go
type Insight struct {
    ID          string
    TaskID      string
    SourceRunID string
    SourceWU    string

    Type        string   // api_change, dependency_update, pattern, risk, opportunity
    Category    string   // technical, architectural, operational
    Severity    string   // info, warning, critical

    Title       string
    Description string
    Evidence    []string // arquivos, linhas, diffs

    RelevantTo  []string // work_unit_ids que devem receber
    SharedAt    *time.Time
    ExpiresAt   *time.Time

    Status      string   // pending_review, published, delivered, stale
}
```

### Filtros de Relevancia

O Orchestrator aplica filtros antes de compartilhar:

```go
func ShouldShare(insight Insight, targetWU WorkUnit) bool {
    // Mesma task?
    if insight.TaskID != targetWU.TaskID {
        return false
    }

    // Overlap de paths?
    if hasPathOverlap(insight.Evidence, targetWU.OwnedPaths, targetWU.ReadPaths) {
        return true
    }

    // Overlap de dominio?
    if hasDomainOverlap(insight.Category, targetWU.Objective) {
        return true
    }

    // Dependencia direta?
    if dependsOn(targetWU, insight.SourceWU) {
        return true
    }

    return false
}
```

## 4. Contract Net (Rede de Contratos)

Mecanismo para atribuicao dinamica de work units a agentes com base em capacidade e disponibilidade.

### Fases

#### Fase 1: Anuncio (Announcement)

O Orchestrator anuncia uma WU disponivel:

```json
{
  "type": "orchestrator.task_announce",
  "payload": {
    "work_unit_id": "wu_005",
    "objective": "Revisar codigo de autenticacao",
    "required_profile": "reviewer",
    "required_capabilities": ["go", "security_review"],
    "estimated_effort": "30m",
    "deadline": "2026-05-11T12:00:00Z"
  }
}
```

#### Fase 2: Proposta (Bidding)

Agentes disponiveis (ou o Agente Inteligente) avaliam e propoem:

```json
{
  "type": "agent.bid",
  "payload": {
    "work_unit_id": "wu_005",
    "agent_id": "agent_reviewer_01",
    "confidence": 0.95,
    "estimated_duration": "25m",
    "rationale": "Tenho contexto da WU anterior de auth"
  }
}
```

#### Fase 3: Selecao (Award)

O Orchestrator seleciona o melhor candidato:

```json
{
  "type": "orchestrator.bid_award",
  "payload": {
    "work_unit_id": "wu_005",
    "winner_agent_id": "agent_reviewer_01",
    "reason": "Maior confianca e contexto relevante"
  }
}
```

### Simplificacao para o MVP

No MVP inicial, o Contract Net e simplificado:
- O Orchestrator (Go) atribui WUs com base no perfil definido no Task Graph
- Nao ha "bidding" real de agentes
- O Agente Inteligente pode sugerir reatribuicao em runtime
- A versao completa de Contract Net entra em fases futuras

## 5. Deadlock e Livelock Detection

### Deadlock por Recurso

Ocorre quando duas ou mais WUs precisam de recursos que estao sob ownership uma da outra.

**Exemplo**:
```
WU-A possui auth.go, precisa de user.go
WU-B possui user.go, precisa de auth.go
```

**Deteccao**:
- Grafo de espera de recursos
- Ciclo no grafo = deadlock

**Prevencao**:
- O Task Graph deve declarar todos os `read_paths` e `owned_paths`
- O `WorkUnitService` valida conflitos ANTES de iniciar runs
- Se conflito detectado: serializar ou replanejar

**Resolucao em runtime**:
1. Detecta ciclo no grafo de espera
2. PAUSE em todas as WUs do ciclo
3. Agente Inteligente analisa e sugere:
   - a. Serializar (uma WU executa primeiro)
   - b. Replan (dividir work units para eliminar conflito)
   - c. Escalonar para humano

### Deadlock por Dependencia

Ocorre quando WUs esperam uma pela outra em cadeia circular.

**Prevencao**:
- Task Graph deve ser aciclico (DAG)
- Validacao antes da persistencia do grafo

**Deteccao em runtime**:
- Se uma WU nao avanca porque depende de outra que tambem nao avanca
- Timeout de dependencia excedido

### Livelock

Agentes estao ativos, mas nao produzem progresso util.

**Exemplos**:
- Dois agentes ficam se notificando sobre mudancas sem parar
- Agente fica alternando entre duas abordagens

**Deteccao**:
- `loop_detected` pelo Observation API
- Checkpoint repetido com mesmo estado

**Resolucao**:
- INTERRUPT com instrucao para parar e reavaliar
- Se persistir: RESTART com prompt corrigido

## 6. Serializacao e Locking

### Lock de Path

Quando duas WUs precisam acessar o mesmo arquivo:

```go
type PathLock struct {
    Path        string
    WorkUnitID  string
    RunID       string
    LockType    string   // read, write
    AcquiredAt  time.Time
    ExpiresAt   time.Time
}
```

**Regras**:
- Multiplos `read` locks sao permitidos simultaneamente
- `write` lock e exclusivo
- Se WU solicita `write` e outra tem `read`: aguarda ou aborta (depende de politica)
- Locks expiram com o run (ou timeout configurado)

### Serializacao de Execucao

Para WUs conflitantes, o Orchestrator pode optar por serializar:

```text
Opcao A: Paralelo controlado
  WU-A (write auth.go) -> completa
  WU-B (write auth.go) -> espera -> completa

Opcao B: Replan
  WU-A (write auth.go) -> completa
  WU-B (write user.go, read auth.go) -> paralelo
```

A decisao entre serializar ou replanejar pode ser:
- Deterministica (Go): se conflito e simples e previsivel
- Estrategica (LLM): se conflito exige analise semantica

## 7. Validation Gate e Review-Session

O **Validation Gate** e um mecanismo de coordenacao que insere uma **Review-Session** obrigatoria entre a conclusao de uma work unit e o inicio de suas dependentes.

### Conceito

Em vez de liberar uma WU automaticamente ao completar, o Orchestrator pode exigir que um agente `reviewer` valide o trabalho antes de permitir que WUs dependentes iniciem.

```text
WU-001 (Implementar schema) --completa--> [VALIDATION GATE] --approved--> WU-002 (Implementar repository)
                                              |
                                              |--changes_requested--> WU-001 (retry)
```

### Tipos de Gate

#### Gate Obrigatorio (Hard Gate)

WU dependente NUNCA inicia ate que a Review-Session emita `approved`.

**Uso**: WUs de risco alto, contratos publicos, mudancas de API.

#### Gate Flexivel (Soft Gate)

WU dependente pode iniciar em paralelo com a Review-Session, mas nao pode concluir ate que o gate seja satisfeito.

**Uso**: WUs independentes que podem comecar a ler/entender o codigo enquanto reviewer valida.

#### Gate por Politica

Gate ativado apenas se:
- WU tem `risk_level >= medium`
- WU modificou arquivos em `critical_paths`
- Task tem `review_required: true`

### Declaracao no Task Graph

```json
{
  "nodes": [
    {"id": "wu_001", "objective": "Criar schema SQL"},
    {"id": "wu_002", "objective": "Implementar repository"}
  ],
  "edges": [
    {"from": "wu_001", "to": "wu_002", "type": "blocks"}
  ],
  "gates": [
    {
      "id": "gate_001",
      "after_node": "wu_001",
      "type": "review_session",
      "mode": "hard",
      "required_veredict": "approved",
      "review_focus": ["syntax", "tests", "pattern_consistency"],
      "auto_retry_on_changes_requested": true,
      "max_retries": 2
    }
  ]
}
```

### Protocolo de Gate

```text
1. WU-001 emite evento work_unit.completed
2. OrchestratorService detecta gate_001 associado
3. OrchestratorService cria Review-Session:
   a. Coleta diff de wu_001
   b. Coleta criterios de aceite
   c. Spawna agente reviewer com contexto
4. Agente reviewer emite veredicto
5. OrchestratorService processa:
   - approved: marca gate como satisfeito, agenda wu_002
   - changes_requested: marca wu_001 para retry, mantem wu_002 blocked
   - needs_discussion: escalona para humano
6. Toda decisao vira evento auditavel
```

### Review-Session como Coordenacao

A Review-Session nao e apenas validacao; e um **ponto de sincronizacao semantico**:

- **Qualidade**: Garante que codigo ruim nao propaga para WUs dependentes
- **Conhecimento**: O review vira memoria (padroes, erros comuns, decisoes)
- **Continuidade**: Se WU-001 e retry, WU-002 ja tem contexto do que mudou
- **Gate humano**: Human pode substituir o reviewer em casos criticos

## 8. Coordenacao Temporal

### Timeouts Global e Local

| Timeout | Escopo | Responsavel |
|---------|--------|-------------|
| Task timeout | Toda a task | OrchestratorService |
| Work unit timeout | WU individual | OrchestratorService |
| Run timeout | Run individual | RunService |
| Tool timeout | Chamada de ferramenta | Runtime / Policy |
| Heartbeat timeout | Sinal de vida | AgentSessionService |
| Checkpoint timeout | Tempo entre checkpoints | OrchestratorService |
| Barrier timeout | Espera por barreira | OrchestratorService |

### Politica de Timeout

```go
type TimeoutPolicy struct {
    TaskTimeout        time.Duration
    WorkUnitTimeout    time.Duration
    RunTimeout         time.Duration
    ToolTimeout        time.Duration
    HeartbeatInterval  time.Duration
    CheckpointInterval time.Duration
    BarrierTimeout     time.Duration

    // Comportamento ao expirar
    OnTaskTimeout      string   // fail, escalate
    OnWorkUnitTimeout  string   // retry, replan, fail
    OnRunTimeout       string   // retry, fail
    OnBarrierTimeout   string   // fail_hard, fail_soft, escalate
}
```

## 9. Rebalanceamento Dinamico

### Split de Work Unit

Se uma WU esta muito grande ou lenta:

```text
1. Agente Inteligente ou Go detecta que WU extrapolou estimativa
2. Sugere split em 2+ WUs menores
3. OrchestratorService valida:
   - As novas WUs mantem dependencias validas?
   - Nao criam ciclos?
   - Paths sao particionados sem conflito?
4. Se validado:
   - Marca WU original como `split`
   - Cria novas WUs
   - Atualiza Task Graph
   - Redireciona runs se necessario
```

### Merge de Work Unit

Se duas WUs sao muito pequenas e uma ja terminou:

```text
1. WU-A completa rapidamente
2. WU-B ainda nao comecou e e similar
3. Agente Inteligente sugere merge
4. OrchestratorService valida e aplica
```

## 10. Métricas de Coordenacao

O Orchestrator deve coletar metricas para avaliar a eficiencia da coordenacao:

```go
type CoordinationMetrics struct {
    TaskID                string
    TotalWorkUnits        int
    CompletedWorkUnits    int
    FailedWorkUnits       int
    ParallelMax           int       // maximo de agentes simultaneos
    TotalDuration         time.Duration
    CriticalPathDuration  time.Duration
    EfficiencyRatio       float64   // critical_path / total_duration
    DeadlockCount         int
    BarrierWaitTime       time.Duration
    ReplanCount           int
    SplitCount            int
    InsightSharedCount    int
    InterventionsRequired int
}
```

**Efficiency Ratio**:
- `1.0` = execucao perfeitamente paralela (impossivel na pratica)
- `0.5` = metade do tempo foi gasto em espera/serializacao
- Quanto mais proximo de 1.0, melhor a coordenacao

## Dependencias

- `OrchestratorService` — coordenacao central
- `TaskGraphService` — grafo estatico
- `WorkUnitService` — validacao de paths e dependencias
- `RunService` — ciclo de vida de runs
- `Observation API` — deteccao de anomalias
- `AgentSessionService` — checkpoints e estado
- `EventService` — persistencia de eventos de coordenacao

## Referencias

- `docs/adr/0023-hybrid-intelligent-orchestrator.md`
- `docs/architecture/agents/intelligent-orchestrator-agent.md`
- `docs/architecture/observability/orchestrator-observation-api.md`
- `docs/architecture/protocols/orchestrator-intervention-protocol.md`
- `docs/architecture/protocols/communication-protocol.md`
- `docs/architecture/orchestration.md`
- `docs/adr/0006-task-graph-and-agent-intervention.md`
