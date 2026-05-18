# Integração de Runtime com Serviços de Domínio

> **Migrado de ADR 0019 em 2026-05-17.** Este documento descreve o padrão de integração entre runtimes de agente e serviços de domínio. Não é uma decisão arquitetural, mas uma especificação de implementação derivada das ADRs 0011, 0016, 0020 e 0022.

---

## Contexto

O OrchestraOS possui dois componentes de inferência LLM implementados que operam de forma isolada:

- `GeminiRuntime` (`internal/agent/gemini_runtime.go`): executa work units individuais via inference loop multi-turn com function calling. Emite eventos como `agent.started`, `agent.heartbeat`, `agent.checkpoint_reached`, `agent.tool_requested` e `agent.completed` em um canal interno.

- `GeminiPlanner` (`internal/modules/taskgraph/gemini_planner.go`): decompõe tasks em grafos de work units usando a API Gemini com structured generation. Implementa a interface `Planner` e produz `GraphPlan` validado.

O problema central é que nenhum dos dois está integrado ao fluxo operacional:

- O `GeminiRuntime` emite eventos via `emitEvent()` em um canal Go interno, mas esses eventos nunca passam pelo `AgentSessionService.Checkpoint()`. Isso viola a ADR 0007, que define `AgentSessionService` como a fronteira canônica para persistir checkpoints. A sessão fica com `last_checkpoint_at = NULL` e sem `recoverable_state`.

- O `GeminiPlanner` existe como planner strategy `llm_gemini_v1` dentro do `TaskGraphService.buildPlan()`, mas o CLI `task graph create` não expõe `PlannerStrategy` como flag. O default é `local_heuristic_v1`, então o planner LLM nunca é ativado em operação normal.

- Não existe componente que conecte: `TaskGraphService` → `PromptService` → `RunService` → `AgentSessionService` → `Runtime` → relay de eventos. Cada passo requer intervenção manual via CLI.

---

## Especificação: Relay de Eventos do Runtime para Serviços

Eventos emitidos pelo runtime devem ser roteados para o serviço de domínio correspondente:

- `agent.checkpoint_reached` deve ser encaminhado para `AgentSessionService.Checkpoint()`, que persiste o checkpoint, atualiza `last_checkpoint_at`, `last_seen_event_id` e `recoverable_state` na mesma transação do evento.
- `agent.heartbeat` deve ser encaminhado para `AgentSessionService.Heartbeat()`, que atualiza `last_heartbeat_at`.
- `agent.tool_requested` deve ser persistido via `EventService` e, quando houver Policy Engine, deve passar por validação antes da resposta.
- `agent.completed` deve disparar `RunService.Validate()` ou `RunService.Complete()` conforme política.
- `agent.failed` deve disparar `RunService.Fail()` com motivo do runtime.

O relay deve ser implementado como uma goroutine consumidora do canal `ReceiveEvent()` do runtime, coordenada pelo componente que iniciou a run.

---

## Especificação: Ativação Configurável do Planner LLM

O `TaskGraphService.Decompose()` deve aceitar `PlannerStrategy` de três fontes, em ordem de prioridade:

1. Valor explícito no `DecomposeTaskGraphInput.PlannerStrategy`.
2. Variável de ambiente `ORCHESTRAOS_PLANNER_STRATEGY`.
3. Default: `local_heuristic_v1`.

A CLI `task graph create` deve expor `--planner` como flag opcional com valores aceitos: `local_heuristic_v1` e `llm_gemini_v1`.

A seleção automática por `risk_level` é uma direção futura que não entra neste corte. Quando implementada, tasks com risco `high` ou `critical` poderiam usar LLM por política.

---

## Especificação: Fallback Obrigatório

Quando `llm_gemini_v1` for selecionado e falhar (erro de API, timeout, JSON inválido, validação de DAG falha), o sistema deve:

1. Registrar o motivo da falha no `rationale` do graph.
2. Executar fallback para `local_heuristic_v1`.
3. O `planner_strategy` persistido no graph indica qual estratégia produziu o plano final, não qual foi solicitada.

Esse comportamento já existe em `TaskGraphService.buildPlan()` e deve ser preservado.

---

## Especificação: Unificação do Commander com Domain Services

O pacote `internal/core/transition/` contém helpers cross-module (TransitionInput, OperationResult, AppendTransition, AppendServiceEvent). O cancelamento em cascata e a prompt orchestration foram migrados para `internal/modules/orchestrator/` (camada de orquestração canônica). O antigo `Commander` em `internal/orchestration/commands.go` foi removido quando os serviços de domínio passaram a ser a fronteira obrigatória para transições de estado (ADR 0020). Serviços de domínio em `internal/modules/*/service.go` agora implementam:

- idempotência por `event_id`;
- retry com backoff;
- validação de `owned_paths`;
- atualização de work units relacionadas ao transicionar runs.

Transições de estado devem usar exclusivamente os serviços de domínio nos módulos verticais. A CLI e qualquer futuro componente devem chamar `RunService`, `TaskService`, `WorkUnitService` e `AgentSessionService` para comandos operacionais.

---

## Consequências

- Eventos do runtime passam a ser persistidos e auditáveis pela mesma cadeia dos demais serviços.
- Checkpoints do `GeminiRuntime` atualizam a sessão conforme ADR 0007.
- O planner LLM pode ser ativado via CLI ou variável de ambiente sem mudança de código.
- O fallback automático garante que falhas do LLM não bloqueiem o fluxo de decomposição.
- O `Commander` foi removido. Código existente deve chamar serviços de domínio em `internal/modules/*/service.go`.
- Testes de integração que validem o fluxo completo (Task → Graph → Prompt → Run → AgentSession → Runtime → Checkpoint → Complete) passam a ser viáveis.

---

## Referências

- ADR 0007 — Ciclo Operacional do Agente (Prompts, Ledger, Checkpoints)
- ADR 0013 — Fundação Técnica M0
- ADR 0016 — State Machine Event-Sourced
- ADR 0020 — Serviços de Orquestração
- ADR 0022 — Arquitetura de Módulos Verticais
