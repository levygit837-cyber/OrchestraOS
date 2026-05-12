# Plano de Implementação — OrchestraOS

**Versão:** 2.0 (Reavaliação 2026-05-11)
**Estratégia:** Walking skeleton → Integrar antes de expandir → Segurança antes de paralelismo

---

## Resumo Executivo

Este roadmap organiza o fluxo de implementação em **fases coesas**, onde cada fase valida o que foi construído antes de avançar. A ordem foi revista para corrigir implementação antecipada de componentes avançados (Prompt Composer, GeminiRuntime) antes de existir um fluxo E2E funcional.

**Regra de ouro:** nenhuma fase é considerada completa sem um teste E2E que a exercite de ponta a ponta.

---

## Fluxo Alvo do MVP

```text
UserMessage/CLI
  -> OrchestratorService.RunTask()
    -> TaskGraphService.Decompose() [DAG]
    -> Para cada WorkUnit (topológica):
      -> SandboxService.Setup() [branch + worktree]
      -> RunService.Create()
      -> AgentService.FindOrCreate()
      -> AgentSessionService.Create()
      -> PromptService.PrepareRunPrompt()
      -> Runtime.Start() [Fake / Gemini / Codex]
      -> Runtime Relay [eventos -> serviços de domínio]
      -> Policy Engine [tool requests]
      -> AgentSessionService.Checkpoint()
      -> RunService.Complete()
      -> SandboxService.Collect() [diff]
    -> TaskService.Complete()
  -> CLI exibe status + diff + evidências
```

---

## Estado Atual Honesto

| Fase | Milestone | Status Real |
|---|---|---|
| Fase 1 | M0-M1: Fundação | ✅ Completo |
| Fase 2 | M2-M3: Planejamento Manual | ⚠️ Funcional mas não integrado |
| Fase 3 | M4: Runtime Isolado | ⚠️ Existe mas é ilha (sem relay) |
| Fase 4 | M4.5: Integração E2E | ❌ Inexistente — **PRÓXIMO PASSO CRÍTICO** |
| Fase 5 | M5: Orquestração | ❌ Inexistente |
| Fase 6 | M7: Sandbox | ❌ Inexistente |
| Fase 7 | M8: Policy Engine | ❌ Inexistente |
| Fase 8 | M6+M9: Comunicação + Runtime Real | ❌ Inexistente |
| Fase 9 | M10: Review e Merge | ❌ Inexistente |
| Fase 10 | M11: GitHub Integration | ❌ Inexistente |
| Fase 11 | M12+M8.5: Memória e Autonomia | ❌ Inexistente |
| Fase 12 | ADR 0022: Arquitetura Vertical | ❌ Adiado pós-MVP |

---

## Fases de Implementação

---

### Fase 1: Fundação Técnica ✅

**Milestones:** M0 (Contratos), M1 (Event Store e State Machine)

**Objetivo:** Ter um repositório com tipos, persistência, event store e CLI mínima operacional.

**Entregáveis:**
- Tipos de domínio (`Task`, `Run`, `Event`, `WorkUnit`, `AgentSession`) em `internal/domain/`.
- JSON Schemas versionados em `contracts/schemas/`.
- Migrations SQL incrementais.
- Event Store com idempotência por `event_id`.
- State Machine testável para todas as entidades.
- CLI mínima: `task create`, `task list`, `task get`.

**Critérios de Aceite:**
- `task create` persiste task e emite `task.created`.
- Evento duplicado com mesmo `event_id` é rejeitado ou retornado como duplicate.
- Transição inválida de status é bloqueada pela state machine.
- Replay reconstroi estado esperado.

---

### Fase 2: Planejamento Manual ✅

**Milestones:** M2 (Task Graph), M3 (Prompt Composer)

**Objetivo:** Decompor tasks em DAG de work units e montar prompts auditáveis.

**Entregáveis:**
- `TaskGraphService.Decompose()` com planner heurístico (`local_heuristic_v1`).
- `GeminiPlanner` (`llm_gemini_v1`) com fallback automático para heurístico.
- Validação de ciclos e conflitos de `owned_paths`.
- `PromptService.PrepareRunPrompt()` com fragmentos versionados.
- `PromptSnapshot` e `ToolsetSnapshot` imutáveis.
- CLI: `task graph create`, `task graph list`.

**Gaps conhecidos (não bloqueantes para esta fase):**
- CLI `task graph create` não expõe flag `--planner`.
- GeminiPlanner nunca é ativado em operação normal.

**Critérios de Aceite:**
- Task com critérios de aceite vira grafo acíclico com work units.
- Ciclo é rejeitado na validação.
- PromptSnapshot tem hash, referências de fragmentos e é persistido.
- ToolsetSnapshot reflete o perfil do agente atribuído.

---

### Fase 3: Runtime Isolado ✅/⚠️

**Milestone:** M4 (Agent Runtime)

**Objetivo:** Ter runtimes que emitam eventos estruturados em um canal.

**Entregáveis:**
- `FakeRuntime` com simulação determinística de heartbeat, checkpoint, tool request e completion.
- `GeminiRuntime` com inference loop multi-turn e function calling via API Gemini.
- Interface `Runtime` unificada (`Start`, `Stop`, `SendEvent`, `ReceiveEvent`, `Status`).

**Gap crítico:**
- Eventos dos runtimes não passam pelos serviços de domínio. `AgentSessionService.Checkpoint()` e `Heartbeat()` **nunca são chamados** durante a execução do runtime.
- Esse gap é resolvido na Fase 4.

**Critérios de Aceite:**
- FakeRuntime conecta, envia eventos e conclui em menos de 5 segundos.
- GeminiRuntime executa inference real quando `GEMINI_API_KEY` está disponível.
- Max steps e timeout funcionam em ambos.

---

### Fase 4: Integração E2E e Relay de Eventos 🔥 PRÓXIMO PASSO

**ADR de referência:** ADR 0019 (Integração de Runtime com Serviços de Domínio), ADR 0011 (Checkpoints)

**Objetivo:** Conectar o runtime ao resto do sistema para que o fluxo Task→Graph→Run→Session→Runtime→Complete funcione de ponta a ponta.

**Por que é crítico:** Sem esta fase, o OrchestraOS é uma coleção de componentes isolados, não um sistema. Todas as fases seguintes dependem dela.

**O que precisa ser feito:**

1. **Relay de Eventos do Runtime**
   - Criar componente `RuntimeEventRelay` que consome `Runtime.ReceiveEvent()` e roteia:
     - `agent.checkpoint_reached` → `AgentSessionService.Checkpoint()`
     - `agent.heartbeat` → `AgentSessionService.Heartbeat()`
     - `agent.completed` → `RunService.Complete()`
     - `agent.failed` → `RunService.Fail()`
     - `agent.tool_requested` → `EventService.Append()` (e futuramente PolicyEngine)
   - Localização: `internal/services/runtime_relay.go` ou `internal/orchestration/runtime_relay.go`
   - Deve rodar em goroutine durante a vida da run.

2. **Atualizar CLI `run start`**
   - Migrar de usar `Commander` para usar exclusivamente `RunService`, `AgentSessionService`, `PromptService`.
   - `run start` deve:
     a. Criar `Run` via `RunService.Create()`
     b. Criar `AgentSession` via `AgentSessionService.Create()`
     c. Preparar prompt via `PromptService.PrepareRunPrompt()`
     d. Iniciar `FakeRuntime` com o prompt
     e. Iniciar relay de eventos
     f. Aguardar conclusão
     g. Exibir status final

3. **Flag `--planner` na CLI**
   - Adicionar ao `task graph create`: `--planner` com valores `local_heuristic_v1` e `llm_gemini_v1`.
   - Ler `ORCHESTRAOS_PLANNER_STRATEGY` como fallback.

4. **Depreciar Commander**
   - Marcar `internal/core/orchestration/commands.go` como `// Deprecated: use domain services instead`.
   - Migrar qualquer uso restante na CLI para serviços de domínio.
   - Não adicionar novas funcionalidades ao Commander.

5. **Teste E2E Integrado**
   - Localização: `tests/integration/e2e_orchestration_test.go`
   - Fluxo validado:
     ```
     TaskService.Create()
     -> TaskGraphService.Decompose()
     -> RunService.Create()
     -> AgentSessionService.Create()
     -> PromptService.PrepareRunPrompt()
     -> FakeRuntime.Start()
     -> RuntimeEventRelay
     -> AgentSessionService.Checkpoint() [atualiza last_checkpoint_at]
     -> RunService.Complete()
     -> SandboxService.Collect() [quando existir]
     ```
   - Variante A: FakeRuntime (sem dependência externa)
   - Variante B: GeminiRuntime (requer `GEMINI_API_KEY`, skip se ausente)

**Critérios de Aceite:**
- [ ] Eventos do FakeRuntime são persistidos via serviços de domínio (verificar no banco).
- [ ] `AgentSession.last_checkpoint_at` é atualizado durante a execução.
- [ ] `Run` transita de `created` → `running` → `completed` automaticamente.
- [ ] Planner LLM pode ser ativado via `--planner llm_gemini_v1` ou env var.
- [ ] Teste E2E com FakeRuntime passa sem dependência externa.
- [ ] Teste E2E com GeminiRuntime passa quando `GEMINI_API_KEY` está disponível.
- [ ] Commander não é usado em nenhum fluxo novo.
- [ ] CLI `run start` funciona como ponto de entrada manual único.

---

### Fase 5: Orquestração Automatizada

**ADR de referência:** ADR 0020 (Orchestrator Service), ADR 0021 (Agent Service), ADR 0023 (Hybrid Intelligent Orchestrator)

**Objetivo:** Automatizar o fluxo manual da Fase 4 com um único comando, estabelecendo a arquitetura híbrida de orquestração (Go determinístico + Agente Inteligente LLM).

**Arquitetura desta fase:**
- **OrchestratorService (Go)**: loop de orquestração deterministico que coordena serviços de domínio.
- **Intelligent Orchestrator Agent (LLM)**: ativado sob demanda para decisões estratégicas (decomposição inteligente, seleção de perfis, diagnóstico).
- **Observation API**: fronteira controlada entre o Agente Inteligente e o Go.
- **Review-Session**: sessão dedicada do agente `reviewer` para validar work units em gates de qualidade.
- **Regra de ouro**: módulos nunca conversam diretamente; cross-module obrigatoriamente via OrchestratorService.

**O que precisa ser feito:**

1. **AgentService**
   - `Create(ctx, input) -> Agent`
   - `FindOrCreate(ctx, profile, runtimeType) -> Agent`
   - `GetByID(ctx, id) -> Agent`
   - Localização: `internal/modules/agent/service.go`
   - Validar perfis: `code_worker`, `docs_writer`, `reviewer`, `debugger`, `default`
   - Validar runtime types: `fake`, `gemini`, `codex_cli`, `external`

2. **Validar AgentID em AgentSessionService.Create()**
   - Verificar que `AgentID` referencia um agente existente no banco.
   - Atualizar testes existentes que usam IDs arbitrários para criar agentes via serviço.

3. **OrchestratorService.RunTask()**
   - Recebe `taskID` e `options` (runtime type, planner strategy, max steps).
   - Fluxo:
     1. Obtém task via `TaskService.GetByID()`
     2. Chama `TaskGraphService.Decompose()` (se não houver grafo ativo)
     3. Resolve ordem topológica do DAG
     4. Para cada work unit executável:
        a. `RunService.Create()`
        b. `AgentService.FindOrCreate()`
        c. `AgentSessionService.Create()`
        d. `PromptService.PrepareRunPrompt()`
        e. Instancia runtime conforme `options.RuntimeType`
        f. Inicia runtime + relay de eventos
        g. Aguarda conclusão ou falha
     5. Quando todas as WUs completas, `TaskService.Complete()`
   - Primeiro corte: execução **sequencial** (1 work unit por vez).
   - Localização: `internal/services/orchestrator_service.go`

4. **CLI `task run`**
   - Flags: `--task-id`, `--runtime`, `--planner`, `--max-steps`
   - Delega diretamente a `OrchestratorService.RunTask()`
   - Exibe progresso das work units no terminal.

5. **Review-Session e Validation Gate**
   - Perfil `reviewer` no catálogo de prompts e toolset.
   - `ValidationGate` declarável no Task Graph (hard, soft, por política).
   - Review-Session spawnada automaticamente quando gate é ativado.
   - Veredicto estruturado: `approved`, `changes_requested`, `needs_discussion`.
   - Se `approved`: libera WUs dependentes.
   - Se `changes_requested`: marca WU original para retry com feedback.
   - Localização: `internal/modules/review/` ou integrado ao `OrchestratorService`.

6. **Triggers Configuráveis (Camada 1)**
   - Thresholds de tokens, steps, tempo, heartbeats no `OrchestratorService`.
   - Detecção determinística de stall, loop, drift, violação de paths.
   - Disparo automático de triggers para o Agente Inteligente.

7. **Testes de integração**
   - `tests/integration/orchestrator_test.go`
   - Valida fluxo completo com FakeRuntime
   - Valida que work units são executadas na ordem topológica correta
   - Valida que agentes são registrados com perfil e runtime type
   - Valida Review-Session como gate entre WUs dependentes
   - Valida triggers de anomalia disparando ativação do Agente Inteligente

**Critérios de Aceite:**
- [ ] `OrchestratorService.RunTask()` executa fluxo completo de uma task com múltiplas work units.
- [ ] Agentes são registrados no banco com perfil e runtime type.
- [ ] `AgentSession.AgentID` referencia agente existente.
- [ ] Work units são executadas na ordem topológica correta.
- [ ] CLI `task run` funciona como ponto de entrada único.
- [ ] Teste E2E com FakeRuntime valida todo o fluxo sem dependência externa.
- [ ] Agente Inteligente pode ser ativado para decompor task em linguagem natural.
- [ ] Observation API expõe resumos estruturados do estado do sistema.
- [ ] Sugestões do Agente Inteligente são validadas pelo Go antes de aplicar.
- [ ] Review-Session valida work unit e emite veredicto estruturado.
- [ ] Validation Gate bloqueia/libera WUs dependentes conforme veredicto.
- [ ] Triggers de anomalia (stall, loop, threshold) disparam ativação do Agente Inteligente.

---

### Fase 6: Sandbox Manager

**ADR de referência:** ADR 0004 (Sandbox e Autonomia Inicial)

**Objetivo:** Isolar filesystem do agente por work unit usando branch e worktree.

**O que precisa ser feito:**

1. **SandboxService**
   - `Setup(ctx, runID, taskID) -> Sandbox`
     - Cria branch: `codex/task-{task_id}-{slug}`
     - Cria worktree em: `~/.local/share/orchestraos/worktrees/{repo_id}/{task_id}/`
     - Persiste sandbox no banco
   - `Collect(ctx, sandboxID) -> Artifact`
     - Coleta diff entre worktree e branch base
     - Cria artifact do tipo `diff`
   - `Teardown(ctx, sandboxID)`
     - Remove worktree mas preserva branch (evidência)
   - Localização: `internal/modules/sandbox/` (novo pacote)

2. **Integração com OrchestratorService**
   - `RunTask()` deve chamar `SandboxService.Setup()` antes de iniciar runtime.
   - `RunTask()` deve chamar `SandboxService.Collect()` ao final da run.
   - `RunTask()` deve chamar `SandboxService.Teardown()` após coleta.

3. **Schema e Migration**
   - Tabela `sandboxes` com: `id`, `run_id`, `task_id`, `branch`, `worktree_path`, `status`, `created_at`.

4. **Testes**
   - Teste de integração valida que worktree é criado fora do repo principal.
   - Teste valida que diff é coletado ao final.
   - Teste valida que teardown não apaga evidências (branch permanece).

**Critérios de Aceite:**
- [ ] Worktree é criado fora do repo principal.
- [ ] Diff é coletado automaticamente ao final da run.
- [ ] Limpeza não apaga evidências (branch persistida).
- [ ] OrchestratorService usa sandbox automaticamente.
- [ ] Teste E2E valida sandbox + runtime + orquestração.

---

### Fase 7: Policy Engine e Tools

**ADR de referência:** ADR 0004 (Autonomia Nível 2)

**Objetivo:** Classificar ferramentas por risco e aplicar decisões de aprovação.

**O que precisa ser feito:**

1. **PolicyService**
   - `EvaluateToolRequest(ctx, request) -> Decision`
   - Regras iniciais:
     - `risk: safe` → auto-aprova (leitura, validação local)
     - `risk: sensitive` → fila de aprovação (rede, segredos, push, PR)
     - `risk: prohibited` → negação imediata
   - Localização: `internal/modules/policy/service.go`

2. **Tool Registry**
   - Definição inicial de ferramentas conhecidas com classificação de risco.
   - Schema de `ToolDefinition` já existe em `domain-model.md`; implementar tabela e repositório.

3. **Integração com Relay de Eventos**
   - O relay (Fase 4) deve passar `agent.tool_requested` pelo `PolicyService` antes de responder ao runtime.
   - Se auto-aprovado: emitir `tool.approved` para runtime.
   - Se fila: pausar runtime (ou deixar pendente) até decisão.
   - Se proibido: emitir `tool.denied` e registrar evento.

4. **Decisões auditáveis**
   - Toda decisão de política vira evento `policy.decision_made`.

5. **Testes**
   - Teste valida auto-aprovação de leitura.
   - Teste valida bloqueio de tool proibida.
   - Teste valida fila de aprovação para tool sensível.

**Critérios de Aceite:**
- [ ] Leitura e validação local são permitidas automaticamente.
- [ ] Rede, segredo, push e PR exigem aprovação (fila).
- [ ] Ação proibida é negada imediatamente.
- [ ] Todas as decisões viram eventos auditáveis.
- [ ] Runtime recebe `tool.approved` ou `tool.denied` conforme política.

---

### Fase 8: Comunicação em Tempo Real e Runtime Real

**Milestones:** M6 (WebSocket), M9 (Codex/CLI Runtime)

**Objetivo:** Ter canal vivo para observação e runtime real que opere em sandbox.

**Por que esta ordem:** WebSocket e runtime real só fazem sentido quando o fluxo E2E já funciona e há sandbox/policy para proteger.

**O que precisa ser feito:**

1. **Servidor WebSocket**
   - Endpoint que publica eventos de uma run em tempo real.
   - Reconexão com `last_seen_event_id` reenvia eventos pendentes.
   - Localização: `internal/websocket/` (novo pacote)

2. **CLI `run watch`**
   - Conecta ao WebSocket e exibe eventos formatados.
   - Suporta `--last-seen-event-id` para retomada.

3. **Interrupt System**
   - Enviar `message.interrupt` que é processado no próximo checkpoint.
   - `OrchestratorService` deve escutar interrupts e decidir pausa ou cancelamento.

4. **Codex/CLI Runtime (ou runtime real)**
   - Implementar `CodexRuntime` que executa `codex` CLI em subprocesso.
   - Capturar stdout/stderr como eventos.
   - Enviar prompts via stdin ou arquivo temporário.
   - Integrar com SandboxService (runtime roda dentro do worktree).
   - Alternativa: aprofundar integração do GeminiRuntime com relay (mais simples se Codex não estiver pronto).

**Critérios de Aceite:**
- [ ] CLI consegue assistir eventos de uma run em tempo real via WebSocket.
- [ ] Reconexão com `last_seen_event_id` reenvia pendências.
- [ ] `message.interrupt` chega no próximo checkpoint e pode pausar a run.
- [ ] Timeout marca sessão como desconectada.
- [ ] Runtime real (Codex ou Gemini integrado) executa work unit simples em sandbox.
- [ ] Resultado inclui diff, validação e resumo.

---

### Fase 9: Review e Merge Gate

**Milestone:** M10

**Objetivo:** Revisar evidências antes de integrar alterações.

**O que precisa ser feito:**

1. **ReviewService**
   - `CollectEvidence(ctx, runID) -> ReviewBundle`
     - Coleta diff do sandbox
     - Coleta logs e artifacts
     - Apresenta critérios de aceite da work unit
   - `SubmitDecision(ctx, reviewID, decision, reason) -> Review`
     - `approved` ou `rejected`
   - Localização: `internal/modules/review/service.go`

2. **CLI `task review`**
   - Mostra diff, evidências e critérios de aceite.
   - Aceita aprovação (`--approve`) ou rejeição (`--reject --reason "..."`).

3. **Merge Automatizado**
   - Quando aprovado, aplica diff na branch principal via git.
   - Registra `MergeDecision` como evento.
   - Se rejeitado, preserva branch e artifacts para análise.

**Critérios de Aceite:**
- [ ] CLI mostra diff e evidências de forma legível.
- [ ] Merge exige aprovação explícita.
- [ ] Rejeição preserva branch e artifacts.
- [ ] Conclusão publica resumo na CLI.

---

### Fase 10: GitHub Integration

**Milestone:** M11

**Objetivo:** Conectar o OrchestraOS ao GitHub como superfície externa.

**O que precisa ser feito:**

1. **GitHub Connector**
   - Criar ou referenciar task a partir de GitHub Issue.
   - Publicar status final em issue/PR quando configurado.
   - Criar PR quando aprovado (ReviewService).
   - Localização: `internal/connectors/github/` (novo pacote)

2. **Outbox Pattern**
   - Operações GitHub são assíncronas e podem falhar.
   - Usar tabela `outbox` para garantir entrega com retry.
   - Processador de outbox em background (ou trigger simples).

3. **Testes**
   - Mock do cliente GitHub para testes.
   - Valida que falha de GitHub não quebra o fluxo principal.

**Critérios de Aceite:**
- [ ] GitHub Issue pode criar ou referenciar task.
- [ ] Status final é enviado para CLI e, quando configurado, para GitHub.
- [ ] GitHub PR só é aberto com aprovação.
- [ ] Falha de conector usa outbox e retry.

---

### Fase 11: Memória Recursiva e Autonomia Avançada

**Milestones:** M12 (Memória Recursiva), M8.5 (Especialização Dinâmica)

**Objetivo:** Adicionar camadas de inteligência operacional sem comprometer fonte de verdade.

**Pré-requisitos:** Event Store ✅, Checkpoints ✅, PromptSnapshot ✅, Artifact Manager (M7 ✅).

**O que precisa ser feito:**

1. **MemoryService**
   - Ingestão de fontes canônicas: ADRs, canvas, checkpoints, artifacts.
   - Deduplicação por `content_hash` e `semantic_key`.
   - Geração de `MemoryRecord` com referências a evidências.
   - Localização: `internal/modules/memory/` (novo pacote)

2. **Retrieval e Injeção**
   - `RetrieveForSession(ctx, query) -> RetrievedMemoryBundle`
   - Respeitar escopo, domínio, paths e autonomia.
   - Injetar bundles no prompt sem exceder token budget.
   - Registrar bundles injetados para evitar repetição.

3. **Especialização Dinâmica (M8.5)**
   - `DynamicPromptFragment` temporário aprovado pelo Orchestrator.
   - Solicitação de tool ausente com decisão explícita.
   - Reconfiguração de AgentSession preservando ledger.

**Critérios de Aceite:**
- [ ] `MemoryRecord` sempre referencia evidência canônica.
- [ ] Memória repetida não gera duplicata.
- [ ] `RetrievedMemoryBundle` respeita escopo e token budget.
- [ ] Falha do serviço de memória não interrompe AgentSession.
- [ ] Especialização dinâmica exige evento e decisão do Orchestrator.

---

### Fase 12: Arquitetura LLM-Optimized (ADR 0022) — Pós-MVP

**Objetivo:** Migrar codebase para módulos verticais 100% isolados.

**Por que pós-MVP:**
- A migração é uma refatoração arquitetural pesada que não adiciona valor funcional ao usuário.
- O código atual funciona e tem testes. Quebrá-lo agora atrasa o MVP.
- Após o MVP, com fluxo validado, a migração pode ser feita por agentes de IA com contexto claro.

**O que precisa ser feito:**
- Quebrar dependências entre `internal/modules/*`.
- Usar interfaces e camada de aplicação (`internal/app/` ou `cmd/`) para orquestração.
- Comunicação entre módulos via eventos (`core/eventstore`) em vez de imports diretos.
- Validar que cada módulo pode ser compreendido isoladamente por um LLM.

---

## Backlog Consolidado

| Prioridade | Item | Fase | Status |
|---|---|---|---|
| P0 | Fundação técnica (M0-M1) | Fase 1 | ✅ |
| P0 | Task Graph + Prompt Composer (M2-M3) | Fase 2 | ✅ |
| P0 | Runtime isolado (M4) | Fase 3 | ⚠️ |
| **P0** | **Relay de eventos Runtime → Serviços** | **Fase 4** | **🔥 PRÓXIMO** |
| **P0** | **Depreciar Commander** | **Fase 4** | **🔥 PRÓXIMO** |
| **P0** | **Teste E2E integrado (Task → Complete)** | **Fase 4** | **🔥 PRÓXIMO** |
| **P0** | **Flag `--planner` na CLI** | **Fase 4** | **🔥 PRÓXIMO** |
| P0 | AgentService | Fase 5 | pendente |
| P0 | OrchestratorService.RunTask() | Fase 5 | pendente |
| P0 | CLI `task run` | Fase 5 | pendente |
| P1 | Sandbox Manager (branch + worktree) | Fase 6 | pendente |
| P1 | Policy Engine mínimo | Fase 7 | pendente |
| P1 | WebSocket / Live View | Fase 8 | pendente |
| P1 | Codex/CLI Runtime ou Gemini integrado | Fase 8 | pendente |
| P2 | Review e Merge Gate | Fase 9 | pendente |
| P2 | GitHub Integration | Fase 10 | pendente |
| P3 | Memória Recursiva | Fase 11 | pendente |
| P3 | Especialização Dinâmica | Fase 11 | pendente |
| P4 | ADR 0022: Módulos Verticais isolados | Fase 12 | pós-MVP |

---

## Decisões Registradas

1. **ADR 0022 é pós-MVP:** A migração para módulos verticais 100% isolados foi adiada para não travar o fluxo E2E. O código atual viola a regra de ouro da ADR 0022, mas funciona. A migração será retomada após o MVP estar operacional.
2. **Sandbox antes de Policy:** Sandbox (Fase 6) vem antes de Policy Engine (Fase 7) porque a política precisa de um ambiente real para proteger. A ordem original (M8 antes de M7) estava invertida.
3. **WebSocket após Runtime integrado:** WebSocket (Fase 8) só é útil quando há eventos reais fluindo. A ordem original (M6 antes de M5) era prematura.
4. **Teste E2E como critério de aceite obrigatório:** Nenhuma fase é considerada completa sem um teste que valide o fluxo inteiro. Isso evita ilhas de código como ocorreu com M4.
