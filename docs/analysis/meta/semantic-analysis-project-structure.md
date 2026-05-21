# Análise Semântica: Estrutura do Projeto OrchestraOS

> Data: 2026-05-17
> Escopo: Todos os diretórios e módulos do projeto
> Objetivo: Avaliar se os nomes descrevem corretamente o propósito, e propor renomeações para longo prazo

---

## 1. A Dilema: Renomear Primeiro ou Migrar Domain Primeiro?

### Cenário A: Migrar Domain Primeiro, Renomear Depois

**Vantagens:**
- Resolve bugs reais imediatamente (regressão de WorkUnit, adapters quebrados)
- A migração de domain já está 20% completa (task e agent prontos)
- Renomeação mecânica pode ser automatizada com `gofmt -r` e `sed`
- Menos risco de conflitos em PRs

**Desvantagens:**
- A migração será feita em nomes "errados" (ex: `orchestrator` em vez de `taskflow`)
- Plans/checklists usarão nomes que vão mudar
- Agentes futuros aprendem nomes temporários

### Cenário B: Renomear Primeiro, Migrar Domain Depois

**Vantagens:**
- Migração de domain é feita "no nome certo" desde o início
- Documentação fica consistente
- Agentes futuros aprendem os nomes finais

**Desvantagens:**
- Adiciona 2-3 horas de trabalho mecânico antes de resolver bugs reais
- A migração de domain tem dependências entre módulos (A03 → A02 → A05...)
- Renomear `orchestrator` afeta ~30 arquivos em 7 diretórios diferentes
- Maior risco de introduzir erros mecânicos (imports quebrados) antes da migração funcional

### Decisão Recomendada: **Migrar Domain Primeiro, Renomear Depois**

**Justificativa:**

1. **A migração de domain é funcional; a renomeação é cosmética.** A regressão de WorkUnit quebra builds e confunde agentes. Um nome errado não quebra nada.

2. **A renomeação é 100% automatizável.** Depois que a migração terminar, um único script pode:
   ```bash
   gofmt -r 'orchestrator -> taskflow' -w .
   find . -type d -name orchestrator -exec rename 's/orchestrator/taskflow/' {} \;
   sed -i 's/orchestratormod/taskflowmod/g' internal/bootstrap/*.go
   ```

3. **A ordem de dependências da migração é crítica.** A03 (WorkUnit) desbloqueia A02 (Run), que desbloqueia A05 (AgentSession), que desbloqueia A07 (Prompt) e A08 (Trigger). Adicionar renomeação nessa cadeia atrasa tudo.

4. **A ADR-0027 já documenta a intenção.** Agentes futuros que leem `orchestrator/README.md` veem o alerta: "será renomeado para taskflow/". A informação está preservada.

**Quando renomear:** Após a conclusão de A09 (Review), o último módulo da migração ADR-0022.

---

## 2. Mapa Semântico Atual vs. Ideal

### Legenda

| Emoji | Significado |
|-------|-------------|
| ✅ | Nome correto e descritivo. Manter. |
| 🟡 | Nome aceitável mas com ressalvas. Pode manter. |
| 🔴 | Nome confuso, ambíguo ou não descritivo. Renomear. |

---

### 2.1 Root e Configuração

| Atual | Avaliação | Proposta | Motivo |
|-------|-----------|----------|--------|
| `cmd/orchestraos/` | ✅ | Manter | CLI entry point. Nome claro. |
| `contracts/` | 🟡 | Manter ou `schemas/` | Tem schemas JSON. "contracts" é genérico mas funciona. |
| `migrations/` | ✅ | Manter | Migrations SQL. Padrão do mercado. |
| `scripts/` | ✅ | Manter | Scripts utilitários. Padrão. |

### 2.2 `internal/` — Infraestrutura e Wiring

| Atual | Avaliação | Proposta | Motivo |
|-------|-----------|----------|--------|
| `internal/bootstrap/` | 🔴 | `internal/wire/` ou `internal/di/` | "Bootstrap" sugere inicialização de sistema. O que faz é **Dependency Injection wiring**. `wire/` é o termo padrão (usado pelo Google Wire). |
| `internal/domain/` | 🟡 | Manter | O nome "domain" sugere o centro do domínio, mas será esvaziado de entity structs. No entanto, renomear é MUITO disruptivo (afeta todos os módulos). Melhor manter o nome e mudar o conteúdo + documentação. |
| `internal/migrations/` | ✅ | Manter | Migrations SQL. Claro. |

### 2.3 `internal/core/` — Pacotes de Infraestrutura

| Atual | Avaliação | Proposta | Motivo |
|-------|-----------|----------|--------|
| `core/apperrors/` | ✅ | Manter | Erros tipados. Nome claro. |
| `core/db/` | ✅ | Manter | Database helpers. Padrão. |
| `core/event/` | 🟡 | Manter | "Event" é genérico, mas o package doc explica que é "event-append service". Aceitável. |
| `core/eventstore/` | ✅ | Manter | Persistência de eventos. Nome claro. |
| `core/coordination/` | 🟡 | Manter | "Coordination" é abstrato, mas o doc explica que é "shared transition helpers for cross-module synchronization". O nome já está enraizado. Renomear para `sync/` ou `cascade/` não agrega valor suficiente para justificar o churn. |
| `core/serialization/` | ✅ | Manter | Marshalling de payloads. Claro. |
| `core/statemachine/` | ✅ | Manter | Regras de transição de estado. Claro. |
| `core/transition/` | 🟡 | Manter | "Transition" como substantivo é estranho, mas o package só tem tipos puros. Aceitável. |
| `core/validation/` | ✅ | Manter | Validação de input. Claro. |

### 2.4 `internal/modules/` — Módulos de Domínio

| Atual | Avaliação | Proposta | Motivo |
|-------|-----------|----------|--------|
| `modules/task/` | ✅ | Manter | Task lifecycle. Nome claro e conciso. |
| `modules/workunit/` | 🟡 | Manter | "WorkUnit" é um nome composto sem hífen (convenção Go). É um pouco longo, mas alternativas (`unit/`, `work/`) são muito genéricas. Aceitável. |
| `modules/run/` | 🟡 | Manter | "Run" é um verbo comum que funciona como substantivo técnico (CI/CD "runs"). Alternativas (`execution/`, `job/`, `attempt/`) não são significativamente melhores. Aceitável. |
| `modules/taskgraph/` | 🟡 | Manter | "TaskGraph" é longo, mas `graph/` sozinho é muito genérico. Aceitável. |
| `modules/agentsession/` | 🟡 | Manter | Longo, mas `session/` sozinho é genérico. Aceitável. |
| `modules/agent/` | ✅ | Manter | Agente e runtime. Nome perfeito. |
| `modules/prompt/` | ✅ | Manter | Prompt engineering. Nome claro. |
| `modules/trigger/` | ✅ | Manter | Anomaly detection e thresholds. Nome claro. |
| `modules/review/` | ✅ | Manter | Reviews e validation gates. Nome claro. |
| `modules/orchestrator/` | 🔴 | `modules/taskflow/` | **Ambiguidade crônica.** Soa como "o cérebro do sistema", mas é apenas um executor de workflow. O nome "orchestrator" será necessário para um futuro módulo de Agente Orquestrador. |

### 2.5 `tests/` — Estrutura de Testes

| Atual | Avaliação | Proposta | Motivo |
|-------|-----------|----------|--------|
| `tests/architecture/` | ✅ | Manter | Testes de arquitetura. Claro. |
| `tests/contracts/` | ✅ | Manter | Testes de contratos. Claro. |
| `tests/integration/` | ✅ | Manter | Testes de integração. Claro. |
| `tests/unit/` | ✅ | Manter | Testes unitários. Claro. |

---

## 3. Renomeações que JUSTIFICAM o Churn

### 3.1 `modules/orchestrator` → `modules/taskflow` (ou `modules/runner`)

**Impacto:** Alto (~30 arquivos em 7 diretórios)
**Justificativa:** Elimina ambiguidade crônica que confunde agentes de IA.
**Quando:** Após a migração ADR-0022.

**Arquivos afetados:**
- `internal/modules/orchestrator/` → renomear diretório
- `internal/bootstrap/services.go` → atualizar imports
- `cmd/orchestraos/cmd/*.go` → atualizar imports (task_run.go, run.go, etc.)
- `internal/core/coordination/*.go` → verificar imports
- `tests/unit/modules/orchestrator/` → renomear diretório
- `tests/integration/*.go` → atualizar imports
- `plans/active/**/*.md` → atualizar referências
- `docs/adr/0020-orchestrator-service.md` → atualizar
- `docs/adr/0023-hybrid-intelligent-orchestrator.md` → atualizar

### 3.2 `internal/bootstrap` → `internal/wire` (ou `internal/di`)

**Impacto:** Médio (~5-10 arquivos)
**Justificativa:** "Bootstrap" não descreve o propósito real (DI wiring). `wire/` é o termo padrão da indústria.
**Quando:** Qualquer momento. Pode ser feito em paralelo com a migração ADR-0022 porque é independente.

**Arquivos afetados:**
- `internal/bootstrap/` → renomear diretório
- `cmd/orchestraos/cmd/*.go` → atualizar imports
- `tests/integration/*.go` → atualizar imports
- `internal/modules/*/service.go` → verificar se referenciam bootstrap (não deveriam)

---

## 4. Renomeações que NÃO Justificam o Churn

| Nome Atual | Proposta Alternativa | Por que não justifica |
|------------|----------------------|----------------------|
| `agentsession` | `session` | "Session" é muito genérico. Perde o contexto de "agent". |
| `taskgraph` | `graph` | "Graph" é muito genérico. Perde o contexto de "task". |
| `workunit` | `unit` | "Unit" é muito genérico. Perde o contexto de "work". |
| `core/coordination` | `core/sync` | "Sync" sugere sincronização de threads, não transações cross-module. Não é significativamente melhor. |
| `core/event` | `core/eventappend` | Muito específico e feio. "Event" é aceitável com documentação. |
| `core/transition` | `core/txtypes` | Feio e não padronizado. "Transition" é aceitável. |
| `contracts` | `schemas` | Ganho marginal. "Contracts" é genérico mas funciona. |

---

## 5. Recomendação de Roadmap de Renomeação

### Fase 1: Durante a Migração ADR-0022 (agora)
- ❌ NÃO renomear nada. Focar na migração de domain.
- ✅ Adicionar alertas nos `doc.go` e `README.md` dos módulos que serão renomeados.
- ✅ Criar ADRs documentando as intenções (já feito: ADR-0027).

### Fase 2: Após a Migração ADR-0022 (futuro próximo)
- ✅ Renomear `modules/orchestrator` → `modules/taskflow`
- ✅ Renomear `internal/bootstrap` → `internal/wire`
- ✅ Atualizar todos os imports mecanicamente
- ✅ Atualizar plans/checklists/docs

### Fase 3: Futuro Distante (quando necessário)
- ✅ Criar `modules/director/` (Agente Orquestrador de alto nível)
- ✅ O nome `orchestrator` estará livre para uso se desejado (ou usar `director/`)

---

## 6. Resumo Executivo

| Ação | Prioridade | Quando |
|------|-----------|--------|
| Migrar domain/types.go para módulos | 🔴 **Alta** | **Agora** (A03 → A02 → A04...) |
| Renomear `orchestrator` → `taskflow` | 🟡 **Média** | Após A09 |
| Renomear `bootstrap` → `wire` | 🟡 **Média** | Após A09 ou em paralelo |
| Renomear outros módulos/core | 🟢 **Baixa** | Não justifica |

**Mensagem principal:** Não deixe a perfeição semântica atrasar a correção funcional. A migração de domain resolve problemas reais hoje. A renomeação é um luxo que pode esperar — e será mais fácil quando a base de código estiver limpa e isolada.
