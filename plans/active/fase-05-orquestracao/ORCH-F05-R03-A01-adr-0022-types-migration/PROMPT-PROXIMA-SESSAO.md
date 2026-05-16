# Prompt para Próxima Sessão — ADR-0022 Type Migration

> **Copie e cole este prompt no chat para continuar o trabalho.** Nada depende de contexto de sessões anteriores.

---

## 1. Objetivo

Migrar **todos os tipos de entidade concretos** de `internal/domain/types.go` para seus respectivos módulos verticais em `internal/modules/<entidade>/models.go`, conforme **ADR-0022** (Vertical Slice Architecture).

O `internal/domain/` deve ficar apenas com contratos genéricos (`EventEnvelope`, `EventPriority`, etc.).

**Branch:** `fix/adr-0022-module-isolation`

---

## 2. Estratégia: Alias Bridge (NÃO Big Bang)

Para garantir que `go build ./...` **nunca quebre** durante a migração:

1. Criar tipos locais no módulo (ex: `task.Task`, `task.Status`).
2. **Manter um alias temporário** em `internal/domain/types.go`:
   ```go
   type Task = task.Task
   type TaskStatus = task.Status
   ```
3. Migrar os arquivos **internos** do módulo para usar tipos locais.
4. `go build ./...` + `go test ./...` → passam (alias mantém compatibilidade).
5. `./scripts/safe-commit.sh`.
6. **Repetir para cada módulo.**
7. **Sessão final:** remover TODOS os aliases de `domain/types.go` e atualizar os consumidores remanescentes.

**Regra de ouro:** Um módulo por sessão. Build verde antes de commit.

---

## 3. Inventário de Entidades (em `internal/domain/types.go`)

| # | Entidade | Módulo Destino | Structs + Tipos |
|---|----------|---------------|-----------------|
| 1 | **Task** | `internal/modules/task/` | `Task`, `TaskStatus` → `Status`, `Priority`, `RiskLevel` |
| 2 | **Run** | `internal/modules/run/` | `Run`, `RunStatus` → `Status`, `RunResult` → `Result` |
| 3 | **WorkUnit** | `internal/modules/workunit/` | `WorkUnit`, `WorkUnitStatus` → `Status` |
| 4 | **TaskGraph** | `internal/modules/taskgraph/` | `TaskGraph`, `TaskGraphStatus` → `Status` |
| 5 | **AgentSession** | `internal/modules/agentsession/` | `AgentSession`, `AgentSessionStatus` → `Status` |
| 6 | **Agent** | `internal/modules/agent/` | `Agent`, `AgentRuntimeType` → `RuntimeType` |
| 7 | **PromptSnapshot** | `internal/modules/prompt/` | `PromptSnapshot`, `PromptFragment`, `ToolsetSnapshot` (já parcialmente local) |
| 8 | **Trigger** | `internal/modules/trigger/` | `Trigger`, `TriggerType`, `TriggerStatus`, `AnomalyType`, `ResolutionAction`, `ThresholdConfig` |
| 9 | **Review** | `internal/modules/review/` | `Review`, `ReviewStatus`, `ReviewDecision`, `ValidationGate`, `ReviewCriteriaChecked` |

**Permanecem em `domain/`:** `EventEnvelope`, `EventPriority`, constantes genéricas.

---

## 4. Convenções

- **Nomenclatura no módulo:** Remover prefixo do pacote. Ex: `domain.TaskStatus` → `task.Status`.
- **Constantes:** Manter prefixo descritivo. Ex: `StatusCreated`, `PriorityP2`.
- **Tags JSON:** Idênticas às originais (para não quebrar Event Store).
- **Aliases:** `type X = x.X` em `domain/types.go`. Remover na Sessão 10.
- **Commit:** `./scripts/safe-commit.sh "ADR-0022: migrate <Entity> types to modules/<entity>"`

---

## 5. Arquivos Essenciais para Ler ANTES de Começar

Leia **na ordem** — são a fonte de verdade:

1. **`/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`** — Roadmap completo com 11 sessões.
2. **`/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/checklist.md`** — Checklist item-a-item por sessão.
3. **`/home/levybonito/Documentos/OrchestraOS/docs/adr/0022-llm-optimized-module-architecture.md`** — Fundamento do ADR-0022.
4. **`/home/levybonito/Documentos/OrchestraOS/internal/domain/types.go`** — Inventário atual de todos os tipos a migrar.
5. **`/home/levybonito/Documentos/OrchestraOS/docs/agents.md`** — Regras obrigatórias para agentes (ler antes de editar).

---

## 6. Como Continuar

### Se nenhuma sessão foi iniciada ainda:

> "Execute **Sessão 1 — Task Module**. Leia o plan.md e checklist.md, depois siga passo a passo. Commit via `./scripts/safe-commit.sh` ao final."

### Se a sessão anterior foi concluída:

> "Continue ADR-0022 migration. Última sessão concluída: **[Task/Run/WorkUnit/etc]**. Próxima sessão: **[nome]**. Leia checklist.md para marcar progresso."

---

## 7. Estado Atual desta Sessão

- ✅ Plano criado em `plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`
- ✅ Checklist criado em `.../checklist.md`
- ✅ Estratégia definida: Alias Bridge (não Big Bang, não adapters espalhados)
- ❌ Nenhuma sessão executada ainda (nenhum tipo migrado)
- ❌ Branch `fix/adr-0022-module-isolation` está no commit original (nenhuma mudança)

---

## 8. Anti-Padrões (não faça)

- ❌ Não migre mais de um módulo por sessão.
- ❌ Não use `type X = domain.X` no módulo (isso é o oposto do objetivo).
- ❌ Não altere `internal/domain/types.go` antes de ter os tipos locais no módulo prontos.
- ❌ Não deixe de rodar `go build ./...` e `go test ./...` antes de commitar.
- ❌ Nunca commit direto em `main`. Use `./scripts/safe-commit.sh`.

---

## 9. Comando Rápido de Validação

```bash
cd /home/levybonito/Documentos/OrchestraOS
go build ./...
go test ./...
```

---

*Prompt criado em 2026-05-15. Atualize esta seção se o estado mudar.*
