# Organizacao de Artefatos de Agente

Este documento define onde, como e sob que convenção os artefatos gerados por agentes sao armazenados no repositório.

---

## 1. Hierarquia de Diretórios

```
docs/agent/
├── README.md                  # Índice do diretório
├── PLAYBOOK.md                # Fluxo obrigatório de execução
├── ARTIFACT_ORGANIZATION.md   # Este documento
├── templates/                 # Templates padrão (5 arquivos fixos)
│   ├── BRIEFING.md
│   ├── SPEC.md
│   ├── PLAN.md
│   ├── PROGRESS.md
│   └── REVIEW.md
├── domains/                   # Contexto persistente por domínio
│   ├── task/
│   ├── run/
│   ├── agentsession/
│   ├── orchestrator/
│   ├── prompt/
│   ├── review/
│   ├── workunit/
│   ├── taskgraph/
│   ├── agent/
│   └── trigger/
└── tasks/                     # Tarefas transversais ou pequenas
    └── YYYY-MM-DD_<slug>/
```

---

## 2. Conceitos

| Conceito | Definição |
|---|---|
| **Domain** | Área de negócio do sistema com fronteiras claras (ex: `task`, `run`). Cada domain mapeia diretamente para um módulo em `internal/modules/`. |
| **Task** | Unidade de trabalho executada por um agente. Pode ser simples (1 arquivo) ou complexa (multi-arquivo, multi-módulo). |
| **Artefato** | Documento gerado durante o ciclo de vida de uma task: briefing, spec, plan, progress, review. |
| **Slug** | Identificador curto da task em kebab-case: 3-5 palavras descritivas. |

---

## 3. Convenção de Naming

### 3.1 Identificador da Task

```
<YYYY-MM-DD>_<slug-descritivo>
```

| Campo | Regra | Exemplo |
|---|---|---|
| `YYYY-MM-DD` | Data de início da task | `2026-05-21` |
| `slug` | 3-5 palavras, kebab-case, sem acentos | `auth-middleware-refactor` |

**Exemplos válidos:**
- `2026-05-21_add-session-timeout`
- `2026-05-22_fix-event-store-race`
- `2026-05-23_orchestrator-graceful-shutdown`

### 3.2 Arquivos de Artefato

Todos em **minúsculas**, sem prefixos:

| Artefato | Nome do arquivo | Sempre gerado? |
|---|---|---|
| Briefing | `briefing.md` | ✅ Sim |
| Spec | `spec.md` | 🟡 Apenas se altera comportamento |
| Plan | `plan.md` | 🟡 Apenas se task é complexa |
| Progress | `progress.md` | 🟡 Opcional (efêmero) |
| Review | `review.md` | ✅ Sim (antes do PR) |

---

## 4. Onde Guardar

### 4.1 Regra de Decisão

```
A task toca um único módulo de domínio claro?
├── Sim → docs/agent/domains/<modulo>/<task-id>/
└── Não → docs/agent/tasks/<task-id>/
```

### 4.2 Domains (`docs/agent/domains/`)

Cada domain possui um `README.md` com contexto persistente: responsabilidade, arquivos-chave, dependências, regras específicas.

```
docs/agent/domains/task/
├── README.md                    # Contexto persistente do módulo task
├── CONTRACTS.md                 # Regras de fronteira (opcional)
└── 2026-05-21_add-auto-archive/ # Task específica
    ├── briefing.md
    ├── spec.md
    └── review.md
```

**Domains existentes (mapeados de `internal/modules/`):**

| Domain | Módulo | Responsabilidade |
|---|---|---|
| `task` | `internal/modules/task/` | CRUD, lifecycle e status de tarefas |
| `workunit` | `internal/modules/workunit/` | Unidades de trabalho dentro de uma task |
| `run` | `internal/modules/run/` | Execução de work units em runtimes |
| `agentsession` | `internal/modules/agentsession/` | Sessão de agente executor (checkpoints, ledger) |
| `agent` | `internal/modules/agent/` | Configuração e metadados de agentes |
| `prompt` | `internal/modules/prompt/` | Composição e snapshot de prompts |
| `trigger` | `internal/modules/trigger/` | Gatilhos e automações |
| `review` | `internal/modules/review/` | Validação e gate de artefatos |
| `taskgraph` | `internal/modules/taskgraph/` | Decomposição de tasks em grafos acíclicos |
| `orchestrator` | `internal/modules/orchestrator/` | Coordenação cross-module (único que pode importar todos) |

### 4.3 Tasks Transversais (`docs/agent/tasks/`)

Para tarefas que:
- Tocam múltiplos módulos sem domínio claro (ex: refatoração global)
- São pequenas demais para justificar um contexto de domain
- São de infraestrutura/core (ex: atualizar `internal/core/eventstore/`)

```
docs/agent/tasks/
└── 2026-05-21_remove-coordination-package/
    ├── briefing.md
    ├── plan.md
    └── review.md
```

---

## 5. Ciclo de Vida dos Artefatos

### 5.1 Durante Execução (Worktree)

Artefatos vivos ficam no worktree temporário do agente, **não no repo principal**:

```
.orchestraos/
└── artifacts/
    └── <task-id>/
        ├── briefing.md
        ├── plan.md
        └── progress.md   ← atualizado em tempo real
```

### 5.2 Após Entrega (Repo Principal)

Ao concluir, artefatos relevantes são movidos para `docs/agent/`:

| Artefato | Destino | Guardar? |
|---|---|---|
| `briefing.md` | `docs/agent/{domains,tasks}/<task-id>/` | ✅ Sim |
| `spec.md` | `docs/agent/{domains,tasks}/<task-id>/` | ✅ Sim (se gerado) |
| `plan.md` | `docs/agent/{domains,tasks}/<task-id>/` | ✅ Sim (se gerado) |
| `review.md` | `docs/agent/{domains,tasks}/<task-id>/` | ✅ Sim |
| `progress.md` | Descartado | ❌ Não (ruído operacional) |

---

## 6. Front Matter Obrigatório

Todo artefato deve incluir metadados no topo:

```yaml
---
tipo: briefing    # briefing | spec | plan | progress | review
task-id: 2026-05-21_add-session-timeout
domain: agentsession   # ou "transversal"
origem: issue #42 | comando CLI | decisao humana
branch: feature/2026-05-21_add-session-timeout
status: em-andamento | concluido | cancelado
---
```

---

## 7. Scripts Automatizados

Use os scripts em `scripts/` para criar estruturas sem erro manual:

| Script | Função |
|---|---|
| `./scripts/new-task.sh` | Cria estrutura de task com templates preenchidos |
| `./scripts/new-domain.sh` | Registra novo domínio em `docs/agent/domains/` |

Consulte `--help` em cada script para uso.

---

## 8. Busca e Descoberta

Para encontrar artefatos rapidamente:

```bash
# Todas tasks de um domínio
fd . docs/agent/domains/task/

# Tasks recentes (últimos 30 dias)
fd . docs/agent/tasks/ | grep "$(date +%Y-%m)" | head -20

# Tasks que mencionam "event store"
rg "event store" docs/agent/ -l

# Briefings de tasks concluídas
fd "briefing.md" docs/agent/ | xargs grep -l "status: concluido"
```
