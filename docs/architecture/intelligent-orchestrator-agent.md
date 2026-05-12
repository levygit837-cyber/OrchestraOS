# Agente Orquestrador Inteligente

## Visao Geral

O **Intelligent Orchestrator Agent** (ou Agente Orquestrador Inteligente) e a camada de decisao estrategica do OrchestraOS. Ele e um agente LLM especializado que opera como **cliente do control plane**, nunca como controlador direto.

Ele transforma intencao humana em planejamento, diagnostica execucoes, sugere correcoes e toma decisoes de medio/alto nivel — mas **sempre** suas decisoes sao validadas e aplicadas pelo `OrchestratorService` Go deterministico.

## Principios

1. **O Agente Inteligente e um cliente, nao um dono.** Ele sugere; o Go decide se executa.
2. **Zero acesso direto.** Sem DB, sem filesystem, sem chamadas diretas a servicos de dominio.
3. **Observacao controlada.** So ve o que o `OrchestratorService` expoe via Observation API.
4. **Comandos estruturados.** Toda acao e emitida como evento/JSON com schema rigido.
5. **Ativacao sob demanda.** Nao fica rodando continuamente; e acionado por triggers.

## Ciclo de Vida

```text
Trigger (mensagem humana OU anomalia detectada pelo Go OU pending approval)
    |
    v
OrchestratorService (Go)
    |
    |-- Cria AgentSession (perfil: orchestrator)
    |-- Cria Run (orchestrator run)
    |-- Prepara Prompt via PromptService
    |   (contexto: estado do sistema, ADRs, politicas, observacoes)
    |
    v
GeminiRuntime.Start() (ou outro runtime LLM)
    |
    v
Loop de Decisao do Agente Inteligente
    |
    |-- 1. Recebe Observation (resumo do estado atual)
    |-- 2. Analisa com LLM
    |-- 3. Escolhe ferramenta de decisao
    |-- 4. Emite comando estruturado
    |-- 5. Aguarda resultado do OrchestratorService
    |-- 6. Se nova observacao necessaria, volta ao passo 1
    |-- 7. Se nao ha mais decisoes, emite checkpoint e completa
    |
    v
Checkpoint + Run Complete
    |
    v
OrchestratorService processa proximos passos (spawn de agentes, etc.)
```

## Triggers de Ativacao

O Agente Inteligente e ativado em situacoes especificas, nao continuamente:

| Trigger | Motivo | Exemplo |
|---------|--------|---------|
| `human_message` | Usuario enviou mensagem em linguagem natural | "Crie uma API de autenticacao" |
| `anomaly_detected` | Go detectou stall, loop ou comportamento anomalo | Agente sem progresso por 10 min |
| `tool_pending_approval` | Ferramenta de risco medio/alto aguarda decisao | `shell.exec` com rede |
| `work_unit_failed` | Work unit falhou e replanejamento e necessario | Testes falharam 3x |
| `task_graph_ready` | Task criada e precisa de decomposicao inteligente | Criterios de aceite gerados |
| `checkpoint_review` | Checkpoint indica bloqueio ou deriva de objetivo | Agente pediu escopo diferente |
| `policy_violation` | Agente tentou acao fora da politica | Acesso a path bloqueado |

## Perfil e Prompt

O Agente Inteligente usa o perfil `orchestrator` no catalogo de prompts.

### SystemPrompt

```text
Voce e o Intelligent Orchestrator do OrchestraOS.

Sua funcao e tomar decisoes estrategicas de orquestracao.
Voce NAO executa codigo, NAO edita arquivos e NAO acessa sistemas diretamente.

Voce analisa o estado do sistema atraves da Observation API e emite comandos
estruturados que o OrchestratorService (Go) validara e executara.

Regras:
1. Sempre justifique suas decisoes no campo `rationale`.
2. Se nao tiver informacao suficiente, peca mais observacoes.
3. Respeite o nivel de autonomia aprovado (Nivel 2 = revisao humana para acoes irreversiveis).
4. Prefira menos agentes bem coordenados a paralelismo artificial.
5. Todo comando e uma SUGESTAO ate o OrchestratorService aprovar.
6. Se o OrchestratorService rejeitar sua sugestao, analise o motivo e adapte.
```

### TaskPrompt

Inclui dinamicamente:
- Estado atual das tasks/runs relevantes
- Politicas vigentes
- Historico de decisoes recentes
- Memorias recuperadas do projeto
- Instrucoes especificas do trigger atual

## Toolset Exclusivo

O Agente Inteligente recebe apenas ferramentas de **analise e decisao**. Nenhuma ferramenta de execucao.

### Ferramentas de Analise

#### `analyze_task_graph`
Recebe o estado atual do grafo de tasks e retorna analise estruturada.

**Input:**
```json
{
  "task_id": "task_123",
  "focus": "bottlenecks"
}
```

**Output:**
```json
{
  "analysis": {
    "bottlenecks": ["wu_002 waiting for wu_001"],
    "parallelizable": ["wu_003", "wu_004"],
    "risk_nodes": ["wu_005: network access required"]
  }
}
```

#### `compare_checkpoints`
Compara checkpoints de uma run para detectar progresso ou regressao.

**Input:**
```json
{
  "run_id": "run_456",
  "checkpoint_ids": ["cp_001", "cp_002", "cp_003"]
}
```

**Output:**
```json
{
  "delta": {
    "goals_completed": ["schema SQL criado"],
    "goals_added": ["ajustar indices"],
    "files_modified": ["migrations/001.sql"],
    "blocked": false
  }
}
```

#### `detect_anomaly`
Solicita ao OrchestratorService uma analise de anomalias em uma task.

**Input:**
```json
{
  "task_id": "task_123",
  "anomaly_types": ["stall", "loop", "resource_exhaustion"]
}
```

**Output:**
```json
{
  "anomalies": [
    {
      "type": "stall",
      "run_id": "run_456",
      "duration_minutes": 12,
      "last_event": "agent.heartbeat",
      "recommendation": "send_interrupt_or_check_status"
    }
  ]
}
```

### Ferramentas de Decisao

#### `suggest_replan`
Sugere replanejamento de uma task ou work unit.

**Input:**
```json
{
  "task_id": "task_123",
  "reason": "Work unit wu_002 falhou 3x. A abordagem atual nao funciona.",
  "suggestion": {
    "action": "split_work_unit",
    "target": "wu_002",
    "new_units": [
      {"objective": "Criar schema isolado", "profile": "code_worker"},
      {"objective": "Integrar schema existente", "profile": "code_worker"}
    ]
  }
}
```

#### `suggest_profile`
Sugere perfil de agente para uma work unit.

**Input:**
```json
{
  "work_unit_id": "wu_003",
  "context": "Essa work unit envolve escrever documentacao tecnica da API"
}
```

**Output (via Go):**
```json
{
  "suggested_profile": "docs_writer",
  "rationale": "A work unit e puramente documentacao, sem logica de negocio.",
  "confidence": 0.92
}
```

#### `suggest_intervention`
Sugere intervencao em uma run ativa.

**Input:**
```json
{
  "run_id": "run_456",
  "intervention_type": "hint",
  "message": "Voce pode usar o WorkUnitService para validar owned_paths antes de editar.",
  "priority": "checkpoint"
}
```

### Ferramentas de Acao Mediada

#### `request_tool_approval`
Decide sobre uma ferramenta pendente de aprovacao.

**Input:**
```json
{
  "tool_request_id": "tr_789",
  "decision": "approved",
  "rationale": "O comando 'go test ./...' e somente leitura, sem rede, dentro do worktree. Risco baixo.",
  "conditions": ["timeout: 5m"]
}
```

#### `send_agent_message`
Envia mensagem para um agente executor.

**Input:**
```json
{
  "run_id": "run_456",
  "message": "Atenção: o arquivo internal/events/repository.go foi alterado por outro agente. Releia antes de continuar.",
  "priority": "interrupt"
}
```

#### `inject_memory`
Solicita injecao de memoria em um agente.

**Input:**
```json
{
  "run_id": "run_456",
  "memory_ids": ["mem_123", "mem_456"],
  "context": "Regras de validacao de eventos"
}
```

### Ferramentas de Escalonamento

#### `escalate_to_human`
Solicita intervencao humana.

**Input:**
```json
{
  "task_id": "task_123",
  "reason": "Conflito de arquitetura: dois agentes sugerem abordagens incompativeis para autenticacao.",
  "urgency": "high",
  "options": ["JWT com refresh tokens", "Session-based com Redis"]
}
```

## Estados do Agente Inteligente

| Estado | Descricao |
|--------|-----------|
| `idle` | Aguardando trigger. Nao consome recursos. |
| `observing` | Coletando observacoes do sistema. |
| `analyzing` | Processando informacoes com LLM. |
| `deciding` | Selecionando e emitindo comando estruturado. |
| `waiting` | Aguardando resposta do OrchestratorService. |
| `adapting` | Recebendo rejeicao; reavaliando decisao. |
| `completing` | Emitindo checkpoint final e concluindo run. |

## Limites e Guardrails

1. **Maximo de iteracoes**: 10 ciclos de decisao por ativacao (evita loops de sugestao).
2. **Timeout por decisao**: 60s para cada chamada de ferramenta.
3. **Orcamento de tokens**: Limite configuravel por trigger (ex: 50k tokens para decomposicao).
4. **Rejeicao acumulada**: Se 3 sugestoes consecutivas forem rejeitadas, escalar para humano.
5. **Sem acesso a segredos**: Observation API nunca expoe tokens, chaves ou credenciais.
6. **Sem acesso a codigo fonte**: O agente ve resumos, nao arquivos inteiros.

## Erros e Recuperacao

| Cenario | Comportamento |
|---------|---------------|
| LLM retorna JSON invalido | Retry com prompt corrigido (max 2x); se falhar, escalar para humano. |
| LLM sugere acao proibida | OrchestratorService rejeita com `reason: policy_violation`. |
| LLM entra em loop de sugestoes | OrchestratorService interrompe apos maximo de iteracoes. |
| Observation API indisponivel | Agente nao e ativado; decisoes pendentes vao para fila humana. |
| Custo excede orcamento | Pausa run do orquestrador; notifica humano. |

## Integracao com Memoria Recursiva

O Agente Inteligente e um consumidor importante da Recursive Memory:

- **Antes de ativar**: MemoryRetriever busca memorias relevantes para o contexto do trigger.
- **Durante analise**: Agente pode solicitar memorias adicionais via `query_memory`.
- **Ao concluir**: Checkpoints do agente orquestrador geram novas memorias (ex: "decomposicao de task de API funcionou bem com 4 work units").

## Exemplo de Fluxo Completo

```text
1. Usuario envia: "Implemente autenticacao JWT com refresh tokens"
   |
2. OrchestratorService (Go) detecta mensagem humana
   |-- Cria Task basica
   |-- Cria Run para Agente Inteligente
   |-- Prepara Prompt com contexto do projeto
   |
3. Agente Inteligente ativa
   |-- analyze_task_graph: ve que e task nova, sem grafos
   |-- suggest_replan: propoe decomposicao em 4 work units
   |   (schema, handler, middleware, tests)
   |-- suggest_profile: atribui perfis a cada WU
   |
4. OrchestratorService valida
   |-- Verifica se grafos sao aciclicos
   |-- Verifica se perfis existem
   |-- Aplica decomposicao
   |-- Spawna agentes executores
   |
5. Durante execucao
   |-- wu_003 (middleware) falha 2x
   |-- Go detecta falha e ativa Agente Inteligente
   |-- Agente analisa checkpoints, ve que o erro e import errado
   |-- suggest_intervention: envia dica para agente executor
   |-- Agente executor corrige e completa
   |
6. Ao final
   |-- Agente Inteligente verifica evidencias
   |-- Escalona para humano: "Task pronta para review. Diffs em 4 arquivos."
```

## Dependencias

- `OrchestratorService` (Go) funcionando
- `Observation API` implementada
- `PromptService` com suporte a perfil `orchestrator`
- `Event Store` para persistencia de decisoes
- `Recursive Memory` (opcional no primeiro corte)
- Runtime LLM (Gemini ou similar)

## Referencias

- `docs/adr/0023-hybrid-intelligent-orchestrator.md`
- `docs/architecture/orchestrator-observation-api.md`
- `docs/architecture/orchestrator-intervention-protocol.md`
- `docs/architecture/multi-agent-coordination.md`
- `docs/architecture/communication-protocol.md`
- `docs/architecture/memory-system.md`
