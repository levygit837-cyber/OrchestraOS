# Plano de Implementacao

Este roadmap organiza o fluxo ideal em fatias tecnicas pequenas e verificaveis.

## Fluxo Alvo

```text
UserMessage
-> Orchestrator
-> Break Tasks into DAG
-> Create Prompts/SystemPrompts
-> Setup Sandbox
-> Spawn Agents
-> Run Task
-> Logs and Events
-> Orchestrator Live View Loop
-> Send Messages to Agents if needed
-> Approve or Deny Tools
-> Agent Loops and Checkpoints
-> Task Complete
-> Approval or Deny Merge
-> Review Diffs
```

## Estrategia

Implementar primeiro um "walking skeleton": uma versao simples que atravessa o fluxo inteiro com agente fake, banco local e CLI. Depois trocar partes por implementacoes reais.

A partir de M4.5, a estrategia muda para "integrar antes de expandir": validar que os componentes existentes funcionam juntos antes de construir novos. Isso evita descobrir incompatibilidades depois de semanas de trabalho em componentes isolados.

## Estado Atual

As milestones M0 a M4 estao completas. O sistema possui:

- Dominio completo: `Task`, `WorkUnit`, `Run`, `AgentSession`, `Event`.
- Event Store com idempotencia por `event_id` e replay estrito.
- State Machine event-sourced para todas as entidades.
- Servicos de dominio: `TaskService`, `RunService`, `WorkUnitService`, `AgentSessionService`, `EventService`, `PromptService`.
- Task Graph com planner heuristico (`local_heuristic_v1`) e planner LLM (`llm_gemini_v1`).
- Prompt Composer com fragmentos versionados, `PromptSnapshot` e `ToolsetSnapshot`.
- FakeRuntime e GeminiRuntime implementados.
- CLI para todas as operacoes de dominio.

Gaps identificados entre M4 e M5:

- GeminiRuntime e GeminiPlanner operam isolados dos servicos de dominio.
- Eventos do runtime nao passam pelo `AgentSessionService.Checkpoint()`.
- Planner LLM nunca e ativado por padrao.
- Commander (`internal/orchestration`) e Domain Services duplicam responsabilidade.
- Nao existe componente que conecte os servicos em fluxo automatizado.

## Milestones

### M0: Contratos e Esqueleto ✅

Objetivo:

- definir estrutura inicial do repo;
- criar tipos principais;
- criar schemas;
- preparar migrations iniciais;
- criar CLI minima.

Aceite:

- `Task`, `Run`, `Event` e `WorkUnit` existem no dominio;
- CLI consegue criar uma task local;
- evento `task.created` e persistido;
- teste de schema falha com payload invalido.

### M1: Event Store e State Machine ✅

Objetivo:

- persistir eventos;
- reconstruir estado de task/run;
- validar transicoes.

Aceite:

- eventos sao idempotentes;
- consulta por `task_id` e `run_id` funciona;
- transicoes invalidas sao bloqueadas;
- replay reconstroi estado esperado.

### M2: Task Graph ✅

Objetivo:

- transformar uma task em DAG de work units;
- validar ciclos e conflitos.

Aceite:

- grafo aciclico valido e persistido;
- ciclo e rejeitado;
- conflito de `owned_paths` bloqueia paralelismo;
- nova versao do grafo pode ser criada em replanejamento.

### M3: Prompt Composer ✅

Objetivo:

- registrar fragmentos;
- montar SystemPrompt e TaskPrompt;
- gerar PromptSnapshot.
- selecionar toolset minimo por AgentSession.

Aceite:

- fragmentos obrigatorios entram no prompt;
- conflito entre fragmentos e detectado;
- snapshot tem hash e versoes;
- task prompt inclui objetivo, criterios, validacao e ledger inicial.
- ToolsetSnapshot e criado para a sessao.

### M4: Agent Runtime ✅

Objetivo:

- simular agente de forma deterministica;
- exercitar heartbeat, checkpoint e conclusao;
- persistir AgentCheckpoint com resumo minimo e evidencias;
- implementar runtime real com GeminiRuntime.

Aceite:

- agente fake conecta, envia eventos, conclui com artifact fake;
- GeminiRuntime executa inference loop multi-turn com function calling;
- max steps e timeout funcionam em ambos os runtimes.

---

### M4.5: Integracao E2E e Relay de Eventos

ADR de referencia: ADR 0019 (Integracao de Runtime com Servicos de Dominio).

Objetivo:

- conectar GeminiRuntime aos servicos de dominio existentes;
- validar que o fluxo completo funciona de ponta a ponta;
- ativar planner LLM como opcao configuravel;
- depreciar Commander em favor de Domain Services.

O que precisa ser feito:

1. **Relay de eventos do Runtime**: criar goroutine que consome `Runtime.ReceiveEvent()` e roteia:
   - `agent.checkpoint_reached` → `AgentSessionService.Checkpoint()`
   - `agent.heartbeat` → `AgentSessionService.Heartbeat()`
   - `agent.completed` → `RunService.Complete()`
   - `agent.failed` → `RunService.Fail()`
   - `agent.tool_requested` → `EventService.Append()` (e futuramente PolicyEngine)
   - Localização: `internal/services/runtime_relay.go` (nova).

2. **Flag `--planner` na CLI**: adicionar flag ao comando `task graph create` com valores `local_heuristic_v1` e `llm_gemini_v1`. Ler `ORCHESTRAOS_PLANNER_STRATEGY` como fallback.
   - Localização: `cmd/orchestraos/cmd/task.go`.

3. **Depreciacao do Commander**: marcar `internal/orchestration/commands.go` como depreciado. Migrar qualquer uso restante para servicos de dominio.

4. **Teste E2E integrado**: criar teste que executa o caminho completo:
   - `TaskService.Create()` → `TaskGraphService.Decompose()` → para cada WorkUnit: `RunService.Create()` → `AgentSessionService.Create()` → `PromptService.PrepareRunPrompt()` → `GeminiRuntime.Start()` → relay de eventos → `AgentSessionService.Checkpoint()` → `RunService.Complete()`.
   - Localização: `tests/integration/e2e_orchestration_test.go`.
   - Variante com FakeRuntime (sem API key) e variante com GeminiRuntime (requer `GEMINI_API_KEY`).

Aceite:

- eventos do GeminiRuntime sao persistidos via servicos de dominio;
- `AgentSession.last_checkpoint_at` e atualizado durante execucao real;
- planner LLM pode ser ativado via `--planner llm_gemini_v1` ou env var;
- teste E2E passa com FakeRuntime sem dependencia externa;
- teste E2E com GeminiRuntime passa quando `GEMINI_API_KEY` esta disponivel;
- Commander nao e usado em nenhum fluxo novo.

---

### M5: Orchestrator Service

ADR de referencia: ADR 0020 (Orchestrator Service e Loop de Orquestracao), ADR 0021 (Agent Service).

Objetivo:

- implementar servico que coordena o fluxo completo de uma task;
- registrar agentes como entidades de dominio;
- executar work units na ordem topologica do DAG automaticamente.

O que precisa ser feito:

1. **AgentService**: CRUD de agentes com perfil, runtime type e status.
   - `Create(ctx, input) -> Agent`
   - `FindOrCreate(ctx, profile, runtimeType) -> Agent`
   - Validacao de `AgentID` em `AgentSessionService.Create()`.
   - Localização: `internal/services/agent_service.go`.

2. **OrchestratorService.RunTask()**: metodo principal que executa o fluxo end-to-end.
   - Recebe `taskID` e `options` (runtime type, planner strategy, max steps).
   - Resolve ordem topologica do DAG de work units.
   - Para cada work unit executavel: cria Agent, Run, AgentSession, prepara prompt, inicia runtime, consome eventos via relay.
   - Primeiro corte: execucao sequencial (1 work unit por vez).
   - Localização: `internal/services/orchestrator_service.go`.

3. **CLI `task run`**: novo comando que delega ao `OrchestratorService.RunTask()`.
   - Flags: `--task-id`, `--runtime`, `--planner`, `--max-steps`.
   - Localização: `cmd/orchestraos/cmd/task.go`.

4. **Testes de integracao**: testar `OrchestratorService.RunTask()` com FakeRuntime e com GeminiRuntime.
   - Localização: `tests/integration/orchestrator_test.go`.

Aceite:

- `OrchestratorService.RunTask()` executa fluxo completo de uma task com multiplas work units;
- agentes sao registrados no banco com perfil e runtime type;
- `AgentSession.AgentID` referencia agente existente;
- work units sao executadas na ordem topologica correta;
- CLI `task run` funciona como ponto de entrada unico;
- teste E2E com FakeRuntime valida todo o fluxo sem dependencia externa.

---

### M6: WebSocket e Live View

Objetivo:

- manter canal vivo agente-orchestrator;
- enviar comandos pendentes;
- suportar reconexao.

O que precisa ser feito:

1. **Servidor WebSocket**: endpoint que publica eventos de uma run em tempo real.
   - Localização: `internal/websocket/` (novo pacote).

2. **CLI `run watch`**: comando que conecta ao WebSocket e exibe eventos.

3. **Reconexao com `last_seen_event_id`**: reenvia eventos pendentes apos reconexao.

4. **Interrupt system**: enviar `message.interrupt` que chega no proximo checkpoint.

Aceite:

- CLI consegue assistir eventos de uma run em tempo real;
- `message.interrupt` chega no proximo checkpoint;
- reconexao com `last_seen_event_id` reenvia pendencias;
- timeout marca sessao como desconectada.

---

### M7: Sandbox Manager

ADR de referencia: ADR 0004 (Sandbox e Autonomia Inicial).

Objetivo:

- criar branch e worktree por work unit;
- isolar filesystem do agente;
- preservar artefatos e coletar diff.

O que precisa ser feito:

1. **SandboxService**: cria branch + worktree por work unit.
   - Diretorio: `~/.local/share/orchestraos/worktrees/{repo_id}/{task_id}/`.
   - Branch: `codex/task-{task_id}-{slug}`.
   - Localização: `internal/sandbox/` (novo pacote).

2. **Integracao com OrchestratorService**: `RunTask()` deve chamar `SandboxService.Setup()` antes de iniciar runtime e `SandboxService.Collect()` ao final.

3. **Coleta de diff**: diff coletado automaticamente ao final da run.

4. **Limpeza controlada**: limpeza nao apaga evidencias.

Aceite:

- worktree fica fora do repo principal;
- diff e coletado ao final;
- limpeza nao apaga evidencias;
- OrchestratorService usa sandbox automaticamente.

---

### M8: Policy Engine e Tools

ADR de referencia: ADR 0004 (autonomia nivel 2).

Objetivo:

- classificar ferramentas por risco;
- autoaprovar acoes seguras;
- exigir aprovacao para acoes sensiveis.

O que precisa ser feito:

1. **PolicyService**: avalia cada `tool.requested` contra politicas configuradas.
   - Auto-aprovacao para tools com `risk: safe` (leitura, validacao local).
   - Queue de aprovacao para tools com `risk: sensitive` (rede, segredos, push, PR).
   - Negacao para tools com `risk: prohibited`.
   - Localização: `internal/services/policy_service.go`.

2. **Integracao com relay de eventos**: o relay (M4.5) deve passar `tool.requested` pelo PolicyService antes de responder ao runtime.

3. **Registro de decisoes**: todas as decisoes de politica viram eventos auditaveis.

Aceite:

- leitura e validacao local sao permitidas automaticamente;
- rede, segredo, push e PR exigem aprovacao;
- acao proibida e negada;
- todas as decisoes viram eventos.

---

### M8.5: Especializacao Dinamica Controlada

Objetivo:

- permitir `DynamicPromptFragment` temporario;
- permitir solicitacao de ferramenta ausente;
- reconfigurar AgentSession com novo PromptSnapshot e ToolsetSnapshot quando aprovado.

Aceite:

- fragmento dinamico nao pode sobrescrever politica global;
- expansao de toolset exige evento e decisao do Orchestrator;
- ledger e preservado entre sessoes;
- historico da sessao anterior continua consultavel.

Observacao: esta milestone e futura. Nao bloqueia o MVP.

---

### M9: Codex/CLI Runtime

Objetivo:

- substituir ou complementar GeminiRuntime com Codex/CLI em sandbox;
- controlar prompt, eventos e tool requests.

Aceite:

- agente Codex executa work unit simples em sandbox;
- eventos principais sao emitidos;
- max steps e timeout funcionam;
- resultado inclui diff, validacao e resumo.

---

### M10: Review e Merge Gate

Objetivo:

- revisar diff;
- aprovar ou negar integracao;
- registrar decisao.

O que precisa ser feito:

1. **ReviewService**: coleta diff do sandbox, apresenta evidencias, registra decisao.
2. **CLI `task review`**: mostra diff, evidencias e aceita aprovacao ou rejeicao.
3. **Merge automatico**: quando aprovado, aplica diff na branch principal.

Aceite:

- CLI mostra diff e evidencias;
- merge exige aprovacao;
- negacao preserva branch e artefatos;
- conclusao publica resumo na CLI/GitHub quando configurado.

---

### M11: GitHub Complementar

Objetivo:

- receber pedidos por issue quando aplicavel;
- publicar status em issue/PR;
- criar PR quando aprovado.

Aceite:

- GitHub Issue cria ou referencia task;
- status final e enviado para CLI/GitHub;
- GitHub PR so e aberto com aprovacao;
- falha de conector usa outbox e retry.

---

### M12: Memoria Recursiva

ADR de referencia: ADR 0012 (Sistema de Memoria Recursiva).

Objetivo:

- criar memoria operacional derivada de fontes canonicas;
- recuperar contexto util para agentes sem depender de transcript completo;
- deduplicar e auditar memorias criadas e injetadas.

Pre-requisitos: Event Store ✅, Agent Task Ledger, Agent Checkpoints ✅, PromptSnapshot ✅, Artifact Manager.

Aceite:

- `MemoryRecord` sempre referencia evidencia canonica;
- ingestao inicial usa ADRs, canvas e checkpoints;
- memoria repetida nao gera duplicata;
- `RetrievedMemoryBundle` respeita escopo, dominio, paths e autonomia;
- bundles injetados sao registrados para evitar repeticao na mesma run;
- falha do servico de memoria nao interrompe AgentSession.

---

## Propostas Futuras Nao Decididas

Estas propostas registram possibilidades para depois do MVP. Elas nao alteram a ordem recomendada, os limites de autonomia ou o paralelismo inicial.

| Proposta | Resumo |
| --- | --- |
| [Massive Agents System](../architecture/massive-agents-system.md) | Execucao controlada de muitos agentes em paralelo sobre work units semanticamente isoladas, com mapeamento estatico de codigo, escopo minimo, validacao local, limites de custo e revisao humana enquanto a autonomia aprovada for Nivel 2. |

## Ordem Recomendada

1. M0 ✅
2. M1 ✅
3. M2 ✅
4. M3 ✅
5. M4 ✅
6. **M4.5** ← proximo
7. M5
8. M6
9. M7
10. M8
11. M9
12. M10
13. M11
14. M12

## Backlog Atualizado

| Prioridade | Item | Milestone | Status |
| --- | --- | --- | --- |
| P0 | Criar dominio `Task`, `Run`, `Event`, `WorkUnit`. | M0 | ✅ |
| P0 | Criar Event Store local em Postgres. | M1 | ✅ |
| P0 | Criar CLI minima para task/run/event. | M0 | ✅ |
| P0 | Criar validador de JSON Schema. | M0 | ✅ |
| P0 | Criar State Machine event-sourced. | M1 | ✅ |
| P0 | Criar servicos de dominio (TaskService, RunService, etc). | M1 | ✅ |
| P0 | Criar planner DAG heuristico. | M2 | ✅ |
| P0 | Criar planner DAG com LLM (GeminiPlanner). | M2 | ✅ |
| P0 | Criar Prompt Composer com fragmentos estaticos. | M3 | ✅ |
| P0 | Criar ToolsetSnapshot minimo por AgentSession. | M3 | ✅ |
| P0 | Criar FakeRuntime. | M4 | ✅ |
| P0 | Criar GeminiRuntime com inference real. | M4 | ✅ |
| **P0** | **Criar relay de eventos do Runtime para servicos.** | **M4.5** | pendente |
| **P0** | **Adicionar `--planner` flag na CLI.** | **M4.5** | pendente |
| **P0** | **Criar teste E2E integrado (Task → Complete).** | **M4.5** | pendente |
| **P0** | **Depreciar Commander em favor de Domain Services.** | **M4.5** | pendente |
| **P1** | **Migrar camadas técnicas para Módulos Verticais (ADR 0022).** | **Arch** | pendente |
| **P1** | **Criar AgentService.** | **M5** | pendente |
| **P1** | **Criar OrchestratorService.RunTask().** | **M5** | pendente |
| **P1** | **Criar CLI `task run`.** | **M5** | pendente |
| P1 | Criar WebSocket/Live View. | M6 | pendente |
| P1 | Criar Sandbox Manager com worktree. | M7 | pendente |
| P1 | Criar Policy Engine minimo. | M8 | pendente |
| P2 | Criar Codex/CLI Runtime em sandbox. | M9 | pendente |
| P2 | Criar Review e Merge Gate. | M10 | pendente |
| P2 | Adicionar GitHub Issue/PR gate. | M11 | pendente |
| P3 | Adicionar DynamicPromptFragment controlada. | M8.5 | pendente |
| P3 | Adicionar memoria recursiva. | M12 | pendente |
| P3 | Adicionar conector de chat opcional. | futuro | pendente |

## Fora Do Primeiro Corte

- Desktop app.
- Web dashboard.
- NATS.
- Temporal.
- gVisor ou Firecracker.
- memoria vetorial compartilhada antes de memoria estruturada, Event Store, checkpoints e deduplicacao.
- marketplace de agentes.
- autonomia nivel 4 ou 5.
- paralelismo real de agentes antes de sandbox e policy engine.
- Orchestrator LLM inteligente antes do fluxo deterministico estar validado.
