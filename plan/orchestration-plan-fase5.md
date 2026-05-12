# 🎼 Plano de Orquestração — Fase 5: Orquestração Automatizada

**ID:** ORCH-F05-R01  
**Data:** 2026-05-11  
**Orquestrador:** Kimi-CLI  
**Agentes:** 3 (Windsurf, Kimi Code Extension, Kimi-CLI)  
**Fase Alvo:** Fase 5 — Orquestração Automatizada (M5)  
**Padrão:** Agent Checklist Pattern (ACP) + Ralph Loop  

---

## Índice de Planos Individuais

Este plano foi decomposto em **3 planos individuais**, um por agente:

| Agente | ID | Tarefa | Plano | Checklist | Ferramenta |
|--------|-----|--------|-------|-----------|------------|
| 1 | ORCH-F05-R01-A01 | AgentService + Validação AgentID | [`plan.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A01-agentservice/plan.md) | [`checklist.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A01-agentservice/checklist.md) | Windsurf |
| 2 | ORCH-F05-R01-A02 | Review Service + Validation Gate | [`plan.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A02-review/plan.md) | [`checklist.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A02-review/checklist.md) | Kimi Code Ext |
| 3 | ORCH-F05-R01-A03 | Triggers Configuráveis | [`plan.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A03-triggers/plan.md) | [`checklist.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A03-triggers/checklist.md) | Kimi-CLI |

> **Regra:** Cada agente deve ler APENAS seu próprio plano e checklist. Não leia os planos dos outros agentes para evitar conflito de contexto.

---

## Estado Atual (Triagem)

### ✅ Completo
- Event Store, State Machine, CLI mínima
- TaskService, WorkUnitService, RunService, AgentSessionService, TaskGraphService, PromptService
- FakeRuntime, GeminiRuntime, RuntimeEventRelay
- Testes E2E integrados
- Prompt Composer com catálogo de fragmentos versionados
- Migrations 001-011

### ⚠️ Parcial / Gaps
- `internal/modules/agent/` tem apenas **runtime**. Não tem service, repository, queries.
- `AgentSessionService.Create()` aceita qualquer `AgentID` sem validar existência.
- Não existe tabela `agents` no banco.

### ❌ Inexistente
- `internal/modules/review/`
- `internal/modules/trigger/`
- `OrchestratorService`
- CLI `task run`
- Perfil `reviewer` no catálogo

---

## Próximo Passo Crítico

**Fase 5 — Orquestração Automatizada** (ADR 0020, 0021, 0023)

Esta rodada (R01) foca nos 3 módulos independentes:
1. **AgentService** + validação de AgentID
2. **Review Service** + Validation Gate
3. **Triggers Configuráveis**

**Rodada 2 (futura):** OrchestratorService.RunTask(), CLI `task run`, Integração E2E.

---

## Decomposição e Isolamento

| Agente | Independência | TOCAR | EVITAR |
|--------|--------------|-------|--------|
| A01 | Total | `internal/modules/agent/*`, `agentsession/service.go`, `migrations/012_agents.sql` | `review/`, `trigger/`, `cmd/` |
| A02 | Total | `internal/modules/review/*`, `domain/types.go`, `prompt/catalog/`, `migrations/013_reviews.sql` | `agent/`, `agentsession/`, `trigger/`, `cmd/` |
| A03 | Total | `internal/modules/trigger/*`, `domain/types.go`, `migrations/014_triggers.sql` | `agent/`, `agentsession/`, `review/`, `cmd/` |

**Risco de conflito:** Baixo. Único ponto compartilhado é `internal/domain/types.go` (adições de types).

---

## Interfaces Contratuais

Nenhuma interface cruzada necessária nesta rodada. Os 3 módulos são independentes.

---

## Ordem de Integração / Merge

1. **A01** → Merge primeiro (cria base de agentes)
2. **A02** → Merge segundo (review é independente)
3. **A03** → Merge terceiro (trigger é independente)

Se houver conflito em `internal/domain/types.go`, resolva aceitando as adições dos 3.

---

## Como Executar

### Para o Orquestrador:
```
Distribua cada plano para o agente correspondente:
- Agente 1 (Windsurf): "Use a skill execute. Leia o plano: plans/active/fase-05-orquestracao/ORCH-F05-R01-A01-agentservice/plan.md"
- Agente 2 (Kimi Code): "Use a skill execute. Leia o plano: plans/active/fase-05-orquestracao/ORCH-F05-R01-A02-review/plan.md"
- Agente 3 (Kimi-CLI): "Use a skill execute. Leia o plano: plans/active/fase-05-orquestracao/ORCH-F05-R01-A03-triggers/plan.md"
```

### Para o Agente Executor:
1. Ative a skill `execute`
2. Leia o plano indicado
3. Leia o checklist associado
4. Siga o Ralph Loop (ler → executar → validar → marcar → repetir)
5. Entregue ao usuário

---

## Checklist Final do Orquestrador

- [x] Estado atual explorado e documentado
- [x] Próximo passo crítico identificado: Fase 5
- [x] Número de agentes confirmado: 3
- [x] Tarefas decompostas em 3 unidades independentes
- [x] Fronteiras TOUCH/EVITAR definidas para cada agente
- [x] Interfaces contratuais definidas (nenhuma necessária nesta rodada)
- [x] Prompts individuais criados (1 por agente)
- [x] Checklists persistentes gerados (1 por agente)
- [x] Ralph Loop documentado e incluído em cada prompt
- [x] Estrutura de planos categorizada criada (`plans/active/`, `archive/`, `templates/`)
- [x] Plano master atualizado como índice
- [x] Skills `orchestrate` e `execute` criadas

---

## Documentação Relacionada

- `docs/development/agent-checklist-pattern.md` — Padrão de checklist e Ralph Loop
- `~/.claude/skills/orchestrate/SKILL.md` — Skill de orquestração (planejador)
- `~/.claude/skills/execute/SKILL.md` — Skill de execução (codificador)
- `plans/README.md` — Estrutura e convenções de planos
