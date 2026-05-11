# ADR 0019: Integracao de Runtime com Servicos de Dominio

## Contexto

O OrchestraOS possui dois componentes de inferencia LLM implementados que operam de forma isolada:

- `GeminiRuntime` (`internal/agent/gemini_runtime.go`): executa work units individuais via inference loop multi-turn com function calling. Emite eventos como `agent.started`, `agent.heartbeat`, `agent.checkpoint_reached`, `agent.tool_requested` e `agent.completed` em um canal interno.

- `GeminiPlanner` (`internal/services/gemini_planner.go`): decompoe tasks em grafos de work units usando a API Gemini com structured generation. Implementa a interface `Planner` e produz `GraphPlan` validado.

O problema central e que nenhum dos dois esta integrado ao fluxo operacional:

- O `GeminiRuntime` emite eventos via `emitEvent()` em um canal Go interno, mas esses eventos nunca passam pelo `AgentSessionService.Checkpoint()`. Isso viola a ADR 0011, que define `AgentSessionService` como a fronteira canonica para persistir checkpoints. A sessao fica com `last_checkpoint_at = NULL` e sem `recoverable_state`.

- O `GeminiPlanner` existe como planner strategy `llm_gemini_v1` dentro do `TaskGraphService.buildPlan()`, mas o CLI `task graph create` nao expoe `PlannerStrategy` como flag. O default e `local_heuristic_v1`, entao o planner LLM nunca e ativado em operacao normal.

- Nao existe componente que conecte: `TaskGraphService` → `PromptService` → `RunService` → `AgentSessionService` → `Runtime` → relay de eventos. Cada passo requer intervencao manual via CLI.

A ADR 0017 definiu que servicos de dominio sao a fronteira obrigatoria para comandos que alteram estado. A ADR 0016 definiu que transicoes de status devem passar por servicos internos. A ADR 0011 definiu que `AgentSessionService` e a unidade canonica de checkpoints. Essas decisoes precisam ser aplicadas ao fluxo de execucao com runtime real.

## Decisao

O OrchestraOS adotara um padrao de integracao entre runtimes de agente e servicos de dominio, resolvendo o isolamento atual.

### Relay de eventos do Runtime para servicos

Eventos emitidos pelo runtime devem ser roteados para o servico de dominio correspondente:

- `agent.checkpoint_reached` deve ser encaminhado para `AgentSessionService.Checkpoint()`, que persiste o checkpoint, atualiza `last_checkpoint_at`, `last_seen_event_id` e `recoverable_state` na mesma transacao do evento.
- `agent.heartbeat` deve ser encaminhado para `AgentSessionService.Heartbeat()`, que atualiza `last_heartbeat_at`.
- `agent.tool_requested` deve ser persistido via `EventService` e, quando houver Policy Engine, deve passar por validacao antes da resposta.
- `agent.completed` deve disparar `RunService.Validate()` ou `RunService.Complete()` conforme politica.
- `agent.failed` deve disparar `RunService.Fail()` com motivo do runtime.

O relay deve ser implementado como uma goroutine consumidora do canal `ReceiveEvent()` do runtime, coordenada pelo componente que iniciou a run.

### Ativacao configuravel do Planner LLM

O `TaskGraphService.Decompose()` deve aceitar `PlannerStrategy` de tres fontes, em ordem de prioridade:

1. Valor explicito no `DecomposeTaskGraphInput.PlannerStrategy`.
2. Variavel de ambiente `ORCHESTRAOS_PLANNER_STRATEGY`.
3. Default: `local_heuristic_v1`.

A CLI `task graph create` deve expor `--planner` como flag opcional com valores aceitos: `local_heuristic_v1` e `llm_gemini_v1`.

A selecao automatica por `risk_level` e uma direcao futura que nao entra neste corte. Quando implementada, tasks com risco `high` ou `critical` poderiam usar LLM por politica.

### Fallback obrigatorio

Quando `llm_gemini_v1` for selecionado e falhar (erro de API, timeout, JSON invalido, validacao de DAG falha), o sistema deve:

1. Registrar o motivo da falha no `rationale` do graph.
2. Executar fallback para `local_heuristic_v1`.
3. O `planner_strategy` persistido no graph indica qual estrategia produziu o plano final, nao qual foi solicitada.

Esse comportamento ja existe em `TaskGraphService.buildPlan()` e deve ser preservado.

### Unificacao do Commander com Domain Services

O pacote `internal/orchestration/commands.go` (`Commander`) foi criado antes da camada de servicos de dominio (ADR 0017). Ambos fazem transicoes atomicas com eventos, mas o `Commander` nao implementa:

- idempotencia por `event_id`;
- retry com backoff;
- validacao de `owned_paths`;
- atualizacao de work units relacionadas ao transicionar runs.

O `Commander` deve ser depreciado. Transicoes de estado devem usar exclusivamente os servicos de dominio. A CLI e qualquer futuro componente devem chamar `RunService`, `TaskService`, `WorkUnitService` e `AgentSessionService` para comandos operacionais.

## Consequencias

- Eventos do runtime passam a ser persistidos e auditaveis pela mesma cadeia dos demais servicos.
- Checkpoints do `GeminiRuntime` atualizam a sessao conforme ADR 0011.
- O planner LLM pode ser ativado via CLI ou variavel de ambiente sem mudanca de codigo.
- O fallback automatico garante que falhas do LLM nao bloqueiem o fluxo de decomposicao.
- O `Commander` nao deve receber novas funcionalidades. Codigo existente que o utilize deve migrar para servicos de dominio.
- Testes de integracao que validem o fluxo completo (Task → Graph → Prompt → Run → AgentSession → Runtime → Checkpoint → Complete) passam a ser viaveis.

## Alternativas consideradas

- **Manter runtime isolado e persistir eventos manualmente no teste**: simples, mas perpetua a violacao da ADR 0011 e impede testes E2E confiáveis.
- **Criar um servico intermediario `RuntimeBridge`**: adiciona indirection sem beneficio claro; o relay pode ser feito por quem inicia a run.
- **Unificar Commander e Domain Services em um unico pacote**: reduziria arquivos, mas misturaria responsabilidades de transicao pura (Commander) com logica de dominio complexa (servicos).
- **Depender de selecao automatica por risk_level agora**: aumentaria escopo antes de validar que o planner LLM funciona de forma confiavel no fluxo completo.
