# Protocolo de Intervencao do Orchestrator

## Visao Geral

O **Protocolo de Intervencao** define como o Orchestrator (tanto o sistema Go quanto o Agente Inteligente) atua sobre agentes executores em execucao.

A intervencao e um mecanismo de **controle cooperativo**: o Orchestrator nao mata processos arbitrariamente, mas envia sinais estruturados que o agente deve respeitar em pontos seguros.

## Principios

1. **Nao-interrupcao brutal.** Prefere-se checkpoint + intervencao a `kill -9`.
2. **Gradualidade.** A escala de intervencao vai da mais leve (dica) a mais forte (terminacao).
3. **Auditabilidade.** Toda intervencao e um evento persistido.
4. **Reversibilidade.** Intervencoes de media intensidade (pause) devem permitir retomada.
5. **Contexto preservado.** Ao intervir, o Orchestrator fornece razao e proximo passo sugerido.

## Escala de Intervencao

```
Nivel 1: HINT          (leve)     → Agente continua, recebe informacao util
Nivel 2: WARNING       (moderado) → Agente e alertado, deve confirmar recebimento
Nivel 3: INTERRUPT     (medio)    → Agente deve parar no proximo checkpoint
Nivel 4: PAUSE         (alto)     → Agente para imediatamente, estado congelado
Nivel 5: RESTART       (critico)  → Run encerrada, nova run iniciada do ultimo checkpoint
Nivel 6: TERMINATE     (fatal)    → Run encerrada, work unit marcada como falha
Nivel 7: ESCALATE      (human)    → Human assume controle
```

### Nivel 1: HINT (Dica)

**Uso**: Agente esta no caminho certo, mas pode se beneficiar de informacao extra.

**Comportamento do Agente**:
- Continua execucao normalmente
- Incorpora a dica no proximo ciclo de raciocinio
- Nao precisa confirmar recebimento (embora seja recomendado)

**Exemplo**:
```json
{
  "type": "message.notify",
  "priority": "notification",
  "payload": {
    "level": "hint",
    "message": "O arquivo internal/auth/middleware.go ja existe no repo. Verifique antes de criar um novo.",
    "context": "file_avoid_duplicate"
  }
}
```

### Nivel 2: WARNING (Alerta)

**Uso**: Detectou comportamento subotimo ou risco iminente.

**Comportamento do Agente**:
- Deve emitir `agent.warning_acknowledged`
- Deve reavaliar proxima acao
- Pode continuar, mas com cautela

**Exemplo**:
```json
{
  "type": "message.notify",
  "priority": "checkpoint",
  "payload": {
    "level": "warning",
    "message": "Voce esta editando um arquivo fora do seu owned_paths. Isso pode causar conflito com wu_003.",
    "context": "path_violation_risk",
    "required_ack": true
  }
}
```

### Nivel 3: INTERRUPT (Interrupcao)

**Uso**: Necessidade imediata de atencao, mas nao e emergencia.

**Comportamento do Agente**:
- Deve parar execucao atual assim que seguro
- Deve processar a mensagem no proximo checkpoint
- Deve emitir `agent.interrupt_processed`

**Exemplo**:
```json
{
  "type": "message.interrupt",
  "priority": "interrupt",
  "payload": {
    "level": "interrupt",
    "message": "wu_001 completou e alterou a interface que voce esta usando. Releia o arquivo antes de continuar.",
    "context": "dependency_updated",
    "required_ack": true,
    "auto_resume": true
  }
}
```

### Nivel 4: PAUSE (Pausa)

**Uso**: Pausa controlada para inspecao, aprovacao ou replanejamento.

**Comportamento do Agente**:
- Pausa imediata (guarda estado atual)
- Estado da run muda para `paused`
- Aguarda `task.resume`

**Exemplo**:
```json
{
  "type": "task.pause",
  "priority": "interrupt",
  "payload": {
    "reason": "Aguardando aprovacao de tool request: shell.exec com acesso a rede",
    "context": "tool_approval_pending",
    "estimated_wait": "5m"
  }
}
```

### Nivel 5: RESTART (Reinicio)

**Uso**: Run corrompida, loop detectado, ou necessidade de recomecar do ultimo checkpoint.

**Comportamento**:
- Run atual e marcada como `failed` ou `cancelled`
- Nova run e criada a partir do ultimo checkpoint valido
- AgentSession pode ser reutilizada ou recriada

**Fluxo**:
```text
1. Orchestrator envia task.cancel para run atual
2. RunService.Fail() ou Cancel()
3. Agente emite checkpoint final
4. Orchestrator cria nova Run via RunService.Create()
5. Nova AgentSession com estado do ultimo checkpoint
6. PromptService prepara prompt de continuacao
7. Runtime reiniciado
```

### Nivel 6: TERMINATE (Terminacao)

**Uso**: Work unit e irrecuperavel. Falha definitiva.

**Comportamento**:
- Run marcada como `failed`
- Work unit marcada como `failed`
- Task pode ser replanejada ou marcada como `failed`
- Sandbox e preservado para diagnostico (conforme politica)

### Nivel 7: ESCALATE (Escalonamento Humano)

**Uso**: Situacao que exige julgamento humano.

**Comportamento**:
- Run pausada
- Notificacao enviada a CLI/GitHub/humano
- Human pode: aprovar, negar, replanejar, ou assumir controle
- Todo input humano vira evento auditavel

### Intervencao de Qualidade: Review-Session

**Uso**: Validar o trabalho de uma work unit antes de liberar dependencias ou permitir continuacao.

**Comportamento**:
- Run do agente executor pode continuar ou ser pausada (depende do gate)
- Spawna sessao dedicada do agente `reviewer`
- Reviewer analisa diff, testes, sintaxe, criterios de aceite
- Emite veredicto: `approved`, `changes_requested`, `needs_discussion`
- Se `approved`: libera proximos passos
- Se `changes_requested`: WU vai para retry com feedback
- Se `needs_discussion`: escalona para humano ou Orquestrador LLM

**Exemplo**:
```json
{
  "type": "orchestrator.review_requested",
  "priority": "checkpoint",
  "payload": {
    "review_session_id": "rev_789",
    "work_unit_id": "wu_001",
    "run_id": "run_456",
    "review_type": "code_review",
    "focus": ["syntax", "tests", "pattern_consistency"],
    "blocking": true
  }
}
```

**Nota**: A Review-Session nao substitui os niveis 1-7. Ela e um mecanismo complementar de qualidade que pode ser acionado por qualquer nivel (ex: apos PAUSE, antes de RESUME) ou por Validation Gate no Task Graph.

## Maquina de Estados da Intervencao

```
                    +---------+
                    |  IDLE   |
                    +----+----+
                         |
         hint/warning    |    interrupt
              +----------+----------+
              |                     |
              v                     v
       +-------------+       +-------------+
       |  INFORMED   |       | INTERRUPTED |
       +-------------+       +------+------+
                                      |
                    pause             |         restart
              +-----------------------+------------------+
              |                                          |
              v                                          v
       +-------------+                            +-------------+
       |   PAUSED    |                            |  RESTARTING |
       +------+------+                            +------+------+
              |                                          |
    resume    |    escalate                    complete  |
       +------+------+                            +------+
       |             |                                  |
       v             v                                  v
+-------------+  +----------+                    +-------------+
|   RUNNING   |  | HUMAN_CTL|                    |   RUNNING   |
+-------------+  +----------+                    +-------------+
                                                         |
                                            terminate    |
                                                 +-------+
                                                 |
                                                 v
                                          +-------------+
                                          |   FAILED    |
                                          +-------------+
```

## Decisao de Intervencao

### Quem decide intervir?

| Fonte da Decisao | Quando | Niveis tipicos |
|------------------|--------|----------------|
| **Go Orchestrator** | Regras deterministicas violadas | PAUSE, TERMINATE |
| **Agente Inteligente** | Diagnostico estrategico | HINT, WARNING, INTERRUPT, RESTART |
| **Humano** | Revisao ou emergencia | Qualquer nivel |
| **Policy Engine** | Violacao de politica | WARNING, PAUSE |
| **Agente Executor** | Auto-detecao de problema | ESCALATE (pede ajuda) |

### Matriz de Decisao do Go Orchestrator

| Condicao | Acao Automatica | Notifica Agente Inteligente? |
|----------|-----------------|------------------------------|
| Heartbeat ausente > 2x intervalo | PAUSE + diagnostico | Sim |
| Stall detectado > threshold | WARNING (1x) → PAUSE | Sim |
| Loop detectado | INTERRUPT + analise | Sim |
| Tool `safe` solicitada | AUTO-APROVE | Nao |
| Tool `guarded` solicitada | PAUSE + notificacao | Sim (se Nivel 2) |
| Tool `destructive` solicitada | PAUSE + ESCALATE humano | Sim |
| owned_paths violado | WARNING → PAUSE | Sim |
| Budget excedido | PAUSE + ESCALATE | Sim |
| Checkpoint ausente > threshold | WARNING | Sim |

### Matriz de Decisao do Agente Inteligente

| Condicao | Acao Sugerida | Requer Validacao Go? |
|----------|---------------|----------------------|
| Decomposicao necessaria | Sugere replan | Sim |
| Perfil incorreto detectado | Sugere mudanca | Sim |
| Agente com dificuldade | HINT ou WARNING | Nao (hint), Sim (warning) |
| Falha provavel | INTERRUPT ou RESTART | Sim |
| Conflito entre agentes | PAUSE + analise | Sim |
| Decisao ambigua | ESCALATE humano | Sim |

## Protocolo de Comunicacao

### Evento de Intervencao

Todo comando de intervencao e um evento no Event Store:

```json
{
  "id": "evt_01HXINTERVENTION",
  "type": "orchestrator.intervention",
  "task_id": "task_123",
  "run_id": "run_456",
  "agent_id": "agent_789",
  "priority": "interrupt",
  "requires_ack": true,
  "created_at": "2026-05-11T10:00:00Z",
  "payload": {
    "intervention_id": "int_001",
    "level": "interrupt",
    "source": "intelligent_orchestrator",
    "reason": "Detectado loop na execucao",
    "context": {
      "detected_loop": ["tool.read_file", "tool.search", "tool.read_file"],
      "repetitions": 3
    },
    "suggested_action": "Reavalie o goal atual. Voce ja leu os arquivos necessarios.",
    "auto_resume": false,
    "timeout_for_ack": "30s"
  }
}
```

### Confirmacao do Agente

```json
{
  "id": "evt_01HXACK",
  "type": "agent.intervention_acknowledged",
  "task_id": "task_123",
  "run_id": "run_456",
  "agent_id": "agent_789",
  "priority": "checkpoint",
  "payload": {
    "intervention_id": "int_001",
    "acknowledged_at": "2026-05-11T10:00:15Z",
    "agent_response": "Entendido. Vou reavaliar o goal e pular para implementacao.",
    "next_action": "implement_solution"
  }
}
```

### Timeout de Confirmacao

Se o agente nao confirmar dentro do prazo:

| Nivel | Timeout | Acao se nao confirmar |
|-------|---------|----------------------|
| HINT | N/A | Nenhuma |
| WARNING | 30s | Escalar para INTERRUPT |
| INTERRUPT | 30s | Escalar para PAUSE |
| PAUSE | 10s | Escalar para TERMINATE |
| RESTART | 30s | TERMINATE + nova run |

## Recuperacao e Rollback

### Após PAUSE

```text
1. Causa resolvida (aprovacao dada, bug corrigido, etc.)
2. Orchestrator envia task.resume
3. Agente retoma do ultimo checkpoint
4. Se checkpoint for muito antigo, Agente Inteligente pode sugerir RESTART
```

### Após RESTART

```text
1. Nova run criada
2. Estado recuperado do ultimo checkpoint
3. Contexto injetado explica o reinicio
4. Agente continua a partir do checkpoint
```

### Após TERMINATE

```text
1. Work unit marcada como failed
2. Agente Inteligente analisa causa
3. Sugere replanejamento ou retry
4. Se retry: nova work unit ou nova run
5. Se falha definitiva: task marcada como failed
```

## Configuracao

Os thresholds de intervencao devem ser configuraveis por task, perfil e nivel de risco:

```go
type InterventionConfig struct {
    StallThreshold         time.Duration  // default: 10m
    LoopRepetitions        int            // default: 3
    DriftThreshold         float64        // default: 0.6
    HeartbeatMultiplier    int            // default: 2
    MaxTokensWithoutCheckpoint int64      // default: 20000
    AutoApproveSafeTools   bool           // default: true
    EscalateOnRepeatedRejection int       // default: 3
}
```

## Dependencias

- `OrchestratorService` para envio de comandos
- `RunService` para transicoes de estado
- `AgentSessionService` para checkpoints
- `EventService` para persistencia
- `Observation API` para deteccao de anomalias

## Referencias

- `docs/adr/0023-hybrid-intelligent-orchestrator.md`
- `docs/architecture/agents/intelligent-orchestrator-agent.md`
- `docs/architecture/observability/orchestrator-observation-api.md`
- `docs/architecture/agents/multi-agent-coordination.md`
- `docs/architecture/protocols/communication-protocol.md`
- `docs/architecture/project/permissions.md`
