# Reavaliação do Roadmap de Implementação

**Data:** 2026-05-11
**Autor:** Análise de Arquitetura do OrchestraOS
**Status:** Proposta para aprovação

---

## 1. Diagnóstico do Estado Atual

### 1.1 O que está realmente implementado

Após auditoria do código, a situação é diferente do que o roadmap atual (M0-M4 ✅) sugere:

| Milestone | Status Real | Observação Crítica |
|---|---|---|
| M0 Contratos | ✅ Funcional | Tipos, schemas, migrations, CLI mínima operam. |
| M1 Event Store | ✅ Funcional | Idempotência, replay, state machine validam transições. |
| M2 Task Graph | ⚠️ Parcial | Planner heurístico funciona. GeminiPlanner existe mas **nunca é ativado** por padrão. CLI `task graph create` não expõe `--planner`. |
| M3 Prompt Composer | ✅ Funcional | Fragmentos, snapshot, toolset funcionam, mas são acionados **manualmente** via CLI. |
| M4 Agent Runtime | ⚠️ Parcial | FakeRuntime e GeminiRuntime existem e rodam, mas seus eventos **nunca passam pelos serviços de domínio**. `AgentSession.Checkpoint()` e `Heartbeat()` não são chamados pelo runtime. |
| M4.5 Integração | ❌ Inexistente | Não existe relay de eventos. Não existe teste E2E que cruze Task→Graph→Run→Session→Runtime→Complete. |
| M5 Orchestrator | ❌ Inexistente | Não existe `OrchestratorService`. Não existe `AgentService`. Fluxo requer 3+ comandos CLI manuais. |
| M6 WebSocket | ❌ Inexistente | Zero código. |
| M7 Sandbox | ❌ Inexistente | Zero código. Nenhuma branch ou worktree é criada. |
| M8 Policy Engine | ❌ Inexistente | Zero código. Todas as tools são "aprovadas" por ausência de política. |
| M9 Codex/CLI Runtime | ❌ Inexistente | Apenas Fake e Gemini. |
| M10-M12 | ❌ Inexistente | Memória, GitHub gate, review — tudo futuro. |

### 1.2 Problemas estruturais identificados

#### Problema A: Implementação antecipada de componentes avançados
O projeto construiu **Prompt Composer com 15+ fragmentos versionados** (M3) e **GeminiRuntime com function calling** (M4) antes de ter um **fluxo mínimo end-to-end que funcione**. O resultado é código sofisticado que não tem como ser exercitado de forma integrada.

#### Problema B: Runtime isolado do domínio
`GeminiRuntime` emite eventos em um `chan` Go interno. Nenhuma goroutine consome esse canal e roteia para `AgentSessionService.Checkpoint()`. Isso viola a ADR 0011 (checkpoints como fronteira canônica) e impede auditoria real.

#### Problema C: Commander legado coexistindo com Domain Services
`internal/core/orchestration/commands.go` (Commander) e os serviços de domínio (`internal/modules/*/service.go`) fazem a mesma coisa: transições atômicas com eventos. O Commander não tem idempotência, retry, nem validação de `owned_paths`. Sua existência cria duas fontes de verdade para transições.

#### Problema D: ADR 0022 aprovada mas não implementada (e talvez prematura)
A ADR 0022 (Módulos Verticais LLM-Optimized) estabeleceu que módulos em `internal/modules/*` **não podem se importar**. O código atual viola essa regra massivamente:
- `task` importa `run`, `workunit`
- `prompt` importa `task`, `run`, `workunit`, `agentsession`
- `run` importa `workunit`
- `workunit` importa `task`, `taskgraph`

Migrar para módulos 100% isolados exigiria:
- Inversão de dependências em todos os serviços
- Camada de orquestração no `cmd/` ou `internal/orchestration/` que monta tudo
- Event-driven communication entre módulos

Isso é um trabalho de **refatoração arquitetural pesada** que não deveria bloquear o MVP.

#### Problema E: Ordem do roadmap original ignora dependências reais
O roadmap original coloca WebSocket (M6) antes de Sandbox (M7) e Policy (M8). Mas:
- WebSocket só é útil quando há runtime real produzindo eventos (M4.5 integrado)
- Policy Engine só faz sentido quando há ferramentas reais sendo executadas em sandbox (M7 + M9)
- Paralelismo de agentes (canvas menciona 2-5) só é seguro com sandbox + policy

### 1.3 Riscos do caminho atual

1. **Complexidade acumulada sem validação**: Temos 3.500+ linhas de código de runtime e prompt que nunca rodaram em um teste E2E.
2. **Falta de testes em serviços críticos**: `agentsession`, `run`, `workunit` não têm testes de unidade.
3. **Arquitetura quebrada**: ADR 0022 criou uma regra que o código atual viola, gerando inconsistência entre documentação e realidade.
4. **Sem fluxo automatizado**: O sistema ainda é uma plataforma de execução manual, não um orquestrador.

---

## 2. Princípios para o Novo Roadmap

1. **Walking Skeleton primeiro**: O fluxo mínimo deve funcionar com código simples antes de adicionar sofisticação.
2. **Integrar antes de expandir**: Cada nova feature deve ser demonstrada funcionando com o que já existe.
3. **Segurança antes de paralelismo**: Sandbox e Policy vêm antes de executar múltiplos agentes reais.
4. **Manual antes de automático**: CLI com comandos manuais funcionando antes de `OrchestratorService`.
5. **Fake antes de real**: FakeRuntime valida o fluxo; Gemini/Codex entram depois.
6. **Arquitetura depois do MVP**: ADR 0022 é importante mas deve ser pós-MVP para não travar entrega.

---

## 3. Proposta de Reordenação

### Fase 1 — Fundação (M0-M1) ✅
Manter como está. Event Store, State Machine, CLI mínima.

### Fase 2 — Planejamento Manual (M2-M3) ✅
Manter como está. Task Graph e Prompt Composer existem.

### Fase 3 — Runtime Isolado (M4) ⚠️
Manter mas marcar como **incompleto sem integração**. O runtime existe mas é ilha.

### Fase 4 — Integração E2E (NOVO — crítico)
**Este é o próximo passo obrigatório.**
- Relay de eventos FakeRuntime → serviços de domínio
- Depreciar Commander
- Teste E2E: Task → Graph → Run → Session → FakeRuntime → Complete
- CLI `run start` deve usar serviços de domínio (não Commander)

### Fase 5 — Orquestração Automatizada
- `AgentService` (registro de agentes)
- `OrchestratorService.RunTask()` (sequencial, FakeRuntime)
- CLI `task run`
- Teste E2E automatizado

### Fase 6 — Sandbox e Isolamento
- `SandboxService`: branch + worktree por work unit
- Integração com `OrchestratorService`
- Coleta de diff

### Fase 7 — Segurança e Política
- `PolicyService`: classificação de risco de ferramentas
- Auto-aprovação / fila de aprovação
- Integração no relay de eventos

### Fase 8 — Runtime Real e Comunicação
- `Codex/CLI Runtime` (ou integração Gemini com relay)
- WebSocket server + CLI `run watch`
- Reconexão com `last_seen_event_id`

### Fase 9 — Revisão e Merge
- `ReviewService`
- CLI `task review`
- Merge gate

### Fase 10 — GitHub Integration
- GitHub Issue/PR connector
- Outbox + retry

### Fase 11 — Memória e Autonomia Avançada
- Memória Recursiva (M12)
- Especialização Dinâmica (M8.5)

### Fase 12 — Arquitetura LLM-Optimized (ADR 0022)
- Migração completa para módulos verticais isolados
- **Pós-MVP**

---

## 4. Mudanças Específicas no Roadmap

| Item Antigo | Problema | Mudança |
|---|---|---|
| M4.5 como "próximo passo" | Era um mini-épico de integração, mas o roadmap o tratava como trivial. | Eleva M4.5 a **Fase 4 obrigatória** com critérios de aceite rígidos. |
| M6 (WebSocket) antes de M7 (Sandbox) | WebSocket sem runtime integrado não transmite nada útil. | M6 deslocado para **Fase 8**, após Sandbox e Policy. |
| M9 (Codex Runtime) após M10/M11 | Runtime real deveria vir antes de Review/Merge para gerar diffs reais. | M9 deslocado para **Fase 8**, junto com WebSocket. |
| ADR 0022 como P1 no backlog | Migração arquitetural pesada durante construção do MVP. | Rebaixado para **Fase 12 (pós-MVP)**. |
| Sem fase de testes E2E explícita | Testes existem isolados, mas não há validação do fluxo completo. | Cada fase agora exige **teste E2E** como critério de aceite. |

---

## 5. Critério de Sucesso do MVP Revisitado

O MVP do OrchestraOS estará validado quando:

1. Uma task pode ser criada pela CLI e decomposta em work units (grafo acíclico).
2. Uma work unit pode ser executada em worktree isolado por um agente FakeRuntime.
3. O Orchestrator consegue montar prompts e registrar PromptSnapshot.
4. O Agent Task Ledger é atualizado em checkpoints (via relay de eventos).
5. O Orchestrator consegue pausar ou negar uma ferramenta solicitada (Policy Engine).
6. Eventos principais ficam persistidos e consultáveis.
7. O resultado tem diff, validação e resumo.
8. CLI recebe status final com evidências.

**Nota:** GitHub Issue/PR, WebSocket live view, Memória Recursiva e Autonomia Nível 4+ são **fora do MVP**, conforme `docs/architecture/mvp.md`.

---

## 6. Recomendação Imediata

1. **Congelar novas features** até a Fase 4 (Integração E2E) estar completa.
2. **Remover ou depreciar** o Commander em favor dos Domain Services.
3. **Escrever o teste E2E** que falha hoje (Task→Complete com FakeRuntime).
4. **Adiar** a migração ADR 0022 para pós-MVP.
5. **Atualizar** `docs/implementation/roadmap.md` com a nova estrutura de fases.
