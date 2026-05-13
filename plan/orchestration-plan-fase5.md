# 🎼 Plano de Orquestração — Fase 5: Orquestração Automatizada

**ID:** ORCH-F05  
**Data:** 2026-05-13  
**Orquestrador:** Kimi-CLI  
**Padrão:** Agent Checklist Pattern (ACP) + Ralph Loop  

---

## Índice de Planos Individuais

### Rodada 01 (Concluída)

| Agente | ID | Tarefa | Plano | Checklist | Ferramenta | Status |
|--------|-----|--------|-------|-----------|------------|--------|
| 1 | ORCH-F05-R01-A01 | AgentService + Validação AgentID | [`plan.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A01-agentservice/plan.md) | [`checklist.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A01-agentservice/checklist.md) | Windsurf | ✅ Completo |
| 2 | ORCH-F05-R01-A02 | Review Service + Validation Gate | [`plan.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A02-review/plan.md) | [`checklist.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A02-review/checklist.md) | Kimi Code Ext | ✅ Completo |
| 3 | ORCH-F05-R01-A03 | Triggers Configuráveis | [`plan.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A03-triggers/plan.md) | [`checklist.md`](plans/active/fase-05-orquestracao/ORCH-F05-R01-A03-triggers/checklist.md) | Kimi-CLI | ✅ Completo |

### Rodada 02 (Em Execução)

| Agente | ID | Tarefa | Plano | Checklist | Ferramenta | Status |
|--------|-----|--------|-------|-----------|------------|--------|
| 1 | ORCH-F05-R02-A01 | OrchestratorService.RunTask() | [`plan.md`](plans/active/fase-05-orquestracao/ORCH-F05-R02-A01-orchestrator/plan.md) | [`checklist.md`](plans/active/fase-05-orquestracao/ORCH-F05-R02-A01-orchestrator/checklist.md) | Windsurf | 🔥 Em Execução |
| 2 | ORCH-F05-R02-A02 | CLI `task run` + Testes E2E | [`plan.md`](plans/active/fase-05-orquestracao/ORCH-F05-R02-A02-cli-task-run/plan.md) | [`checklist.md`](plans/active/fase-05-orquestracao/ORCH-F05-R02-A02-cli-task-run/checklist.md) | Kimi-CLI | 🔥 Em Execução |

> **Regra:** Cada agente deve ler APENAS seu próprio plano e checklist. Não leia os planos dos outros agentes para evitar conflito de contexto.

---

## Estado Atual (Triagem)

### ✅ Completo (R01)
- Event Store, State Machine, CLI mínima
- TaskService, WorkUnitService, RunService, AgentSessionService, TaskGraphService, PromptService
- **AgentService** com Create, GetByID, FindOrCreate + migration 012
- **AgentSessionService** valida AgentID via AgentReader
- **ReviewService** com Create, Start, SubmitVerdict + migrations 013-014
- **TriggerService** com EvaluateRun, EvaluateSession, EvaluateWorkUnit + migration 014
- FakeRuntime, GeminiRuntime, RuntimeEventRelay
- Testes E2E integrados
- Prompt Composer com catálogo de fragmentos versionados (inclui perfil `reviewer`)
- Schemas JSON para agent, review, trigger

### ❌ Inexistente / Próximo Passo Crítico
- **OrchestratorService** — componente central que conecta todos os serviços
- **CLI `task run`** — comando único para executar task de ponta a ponta
- Testes E2E do fluxo automatizado completo (Task → Graph → Run → Complete)

---

## Próximo Passo Crítico

**Fase 5 — Rodada 02: Orquestração Automatizada End-to-End** (ADR 0020, 0021, 0023)

A R01 criou os módulos independentes. A **R02 é o passo crítico** que os conecta:

1. **OrchestratorService.RunTask()** — loop de orquestração que coordena Task → Graph → Run → Agent → Session → Prompt → Runtime → Relay → Complete
2. **CLI `task run`** — interface do usuário que delega ao OrchestratorService
3. **Testes E2E** — validam o fluxo automatizado de ponta a ponta

Sem a R02, o OrchestraOS continua sendo uma plataforma de execução manual.

---

## Decomposição e Isolamento (R02)

| Agente | Independência | TOCAR | EVITAR |
|--------|--------------|-------|--------|
| A01 | Alta (define interface) | `internal/modules/orchestrator/`, `internal/bootstrap/services.go`, `tests/integration/orchestrator_service_test.go` | `cmd/orchestraos/cmd/` |
| A02 | Depende da interface do A01 | `cmd/orchestraos/cmd/task.go`, `cmd/orchestraos/cmd/task_run.go`, `cmd/orchestraos/cmd/run.go`, `tests/integration/orchestrator_e2e_test.go` | `internal/modules/orchestrator/` (implementação interna do A01) |

**Risco de conflito:** Baixo. Ponto compartilhado é `internal/bootstrap/services.go` (factory do OrchestratorService). O A02 pode usar um stub temporário se o A01 ainda não mergeou.

---

## Interfaces Contratuais (R02)

Definida pelo Orquestrador. O Agente 1 implementa; o Agente 2 consome.

```go
package orchestrator

func NewService(deps Dependencies) *Service
func (s *Service) RunTask(ctx context.Context, taskID string, options RunTaskOptions) (*RunTaskResult, error)

type RunTaskOptions struct {
    RuntimeType     string // "fake" | "gemini" | "codex_cli"
    PlannerStrategy string // "local_heuristic_v1" | "llm_gemini_v1"
    MaxSteps        int    // padrão: 10
    TimeoutSeconds  int    // padrão: 300
}

type RunTaskResult struct {
    TaskID    string
    RunIDs    []string
    Status    string // "completed" | "failed" | "partial"
    ReviewIDs []string
}
```

---

## Ordem de Integração / Merge (R02)

1. **A01 (OrchestratorService)** → Merge primeiro (cria a API que o CLI consome)
2. **A02 (CLI + E2E)** → Merge segundo (consome o OrchestratorService real)

Se o A02 terminar antes do A01, o stub temporário em `bootstrap/services.go` permite que o A02 compile e teste. O stub é removido no merge do A01.

---

## Como Executar

### Para o Orquestrador:
```
Distribua cada plano para o agente correspondente:
- Agente 1 (Windsurf): "Use a skill execute. Leia o plano: plans/active/fase-05-orquestracao/ORCH-F05-R02-A01-orchestrator/plan.md"
- Agente 2 (Kimi-CLI): "Use a skill execute. Leia o plano: plans/active/fase-05-orquestracao/ORCH-F05-R02-A02-cli-task-run/plan.md"
```

### Para o Agente Executor:
1. Ative a skill `execute`
2. Leia o plano indicado
3. Leia o checklist associado
4. Siga o Ralph Loop (ler → executar → validar → marcar → repetir)
5. Entregue ao usuário

---

## Checklist Final do Orquestrador

### R01
- [x] Estado atual explorado e documentado
- [x] Próximo passo crítico identificado: Fase 5
- [x] Número de agentes confirmado: 3
- [x] Tarefas decompostas em 3 unidades independentes
- [x] Fronteiras TOUCH/EVITAR definidas para cada agente
- [x] Prompts individuais criados (1 por agente)
- [x] Checklists persistentes gerados (1 por agente)
- [x] Ralph Loop documentado e incluído em cada prompt

### R02
- [x] R01 validada e marcada como completa
- [x] Próximo passo crítico identificado: OrchestratorService + CLI task run
- [x] Número de agentes confirmado: 2
- [x] Tarefas decompostas em 2 unidades com interface contratual
- [x] Interface contratual `RunTask()` definida explicitamente
- [x] Fronteiras TOUCH/EVITAR definidas para cada agente
- [x] Ordem de merge definida (A01 → A02)
- [x] Prompts individuais criados (1 por agente)
- [x] Checklists persistentes gerados (1 por agente)
- [x] Ralph Loop documentado e incluído em cada prompt
- [x] Plano master atualizado como índice

---

## Documentação Relacionada

- `docs/development/agent-checklist-pattern.md` — Padrão de checklist e Ralph Loop
- `docs/adr/0020-orchestrator-service.md` — OrchestratorService e loop de orquestração
- `docs/adr/0021-agent-service.md` — AgentService e registro de agentes
- `docs/adr/0023-hybrid-intelligent-orchestrator.md` — Arquitetura híbrida
- `docs/implementation/roadmap.md` — Roadmap técnico completo
- `~/.claude/skills/orchestrate/SKILL.md` — Skill de orquestração (planejador)
- `~/.claude/skills/execute/SKILL.md` — Skill de execução (codificador)
- `plans/README.md` — Estrutura e convenções de planos
