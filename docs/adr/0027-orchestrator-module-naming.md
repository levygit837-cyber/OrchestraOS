# 0027. Orchestrator Module Naming — Future Rename to runner/ or taskflow/

**Data:** 2026-05-17
**Status:** Accepted
**Relacionada:** ADR-0020, ADR-0022, ADR-0023

---

## 1. Contexto

O módulo `internal/modules/orchestrator/` foi criado como parte da ADR-0020 (Orchestrator Service) e implementa a execução end-to-end de uma task. Porém, o nome `orchestrator` gera **ambiguidade crônica**:

1. **OrchestraOS** (o projeto inteiro) é um sistema de orquestração de agentes.
2. **OrchestratorService** soa como o "cérebro" do sistema, mas na verdade é apenas um **executor determinístico de workflow**.
3. **Agentes de IA** confundem o módulo com um "Agente Orquestrador" — uma entidade de IA que decide qual task executar, aloca recursos, e prioriza trabalho.

Análise do código revela que `orchestrator/`:
- Não tem tabela própria no banco (`repository.go` é placeholder)
- Não tem entity structs (`models.go` tem apenas tipos auxiliares e interfaces DI)
- É um **script de execução sequencial** que chama services em ordem: Task → TaskGraph → WorkUnit → Run → Agent → Session → Prompt → Runtime → Trigger

Isso é um **Workflow Engine**, não um **Orquestrador Inteligente**.

Além disso, `core/coordination/` já existe e faz orquestração cross-module de baixo nível (sincronização de transações, cascatas). Ter dois conceitos chamados "orquestração" em camadas diferentes aumenta a confusão.

---

## 2. Decisão

### 2.1 Renomeação Futura

O módulo `internal/modules/orchestrator/` será **renomeado** para um dos seguintes nomes:

| Nome | Argumento a favor | Argumento contra |
|------|-------------------|------------------|
| `runner/` | Curto, descritivo | Pode soar como "test runner" |
| `taskflow/` | Claramente indica "fluxo de execução de task" | Um pouco mais longo |

**Escolha preferida:** `taskflow/` — é o mais descritivo do que o módulo realmente faz.

**Quando:** Após a conclusão da migração ADR-0022 (migração de tipos para módulos). Renomear durante a migração adicionaria churn desnecessário e risco.

### 2.2 Reserva do Nome `orchestrator`

O nome `orchestrator` será **reservado** para um futuro módulo de **Agente Orquestrador** (também chamado de `director/`):

- Decidir qual task executar e quando
- Alocar recursos (agentes, runtimes)
- Priorizar trabalho baseado em critérios de negócio
- Escalar/parar execuções
- Tomar decisões estratégicas (possivelmente com LLM)

Esse módulo ainda **não existe** e não está no roadmap atual. A reserva é apenas para evitar que o nome seja usado para outra coisa.

### 2.3 Distinção Clara de Responsabilidades

```
┌─────────────────────────────────────────────────────────────┐
│                    FUTURO director/                         │
│         (Agent Orchestrator — decisões estratégicas)        │
│  "Qual task executar? Quando? Com qual agente?"             │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              ATUAL orchestrator/ → FUTURO taskflow/         │
│       (Task Execution Workflow Engine — execução)           │
│  "Executar task X: decompose → run WUs → complete"          │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              ATUAL core/coordination/                       │
│    (Transaction Coordination — consistência transacional)   │
│  "Sincronizar Run + WorkUnit atomicamente na mesma tx"      │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Consequências

### Positivas

- **Eliminação da ambiguidade:** Um agente de IA que lê `taskflow/` sabe imediatamente que é um executor de workflow, não um decisor.
- **Liberação do nome `orchestrator`:** Quando um módulo de Agente Orquestrador for necessário, o nome estará disponível.
- **Alinhamento semântico:** `core/coordination` (técnico), `taskflow/` (execução), `director/` (estratégico) — hierarquia clara.

### Negativas

- **Refatoração mecânica:** Todos os imports, plans, checklists, tests, e docs que mencionam `orchestrator` precisarão ser atualizados.
- **Quebra de links:** Links em documentação e comentários de código podem quebrar.

### Mitigação

A renomeação será feita em um **PR isolado**, após a migração ADR-0022, usando `gofmt -r` e `sed` para atualizar imports mecanicamente. Os plans/checklists de migração ADR-0022 continuarão usando `orchestrator/` até que a migração termine.

---

## 4. Checklist de Renomeação (para execução futura)

- [ ] Renomear diretório `internal/modules/orchestrator/` → `internal/modules/taskflow/`
- [ ] Atualizar package name: `package orchestrator` → `package taskflow`
- [ ] Atualizar todos os imports em:
  - [ ] `internal/bootstrap/services.go`
  - [ ] `internal/core/coordination/*.go`
  - [ ] `cmd/orchestraos/*.go`
  - [ ] `tests/**/*.go`
  - [ ] `plans/**/*.md`
- [ ] Atualizar `module_boundaries_test.go` (se referenciar `orchestrator`)
- [ ] Atualizar todos os `README.md`, `CONTRACTS.md`, `contract.go`, `doc.go`
- [ ] Atualizar `docs/adr/` que mencionam `orchestrator`
- [ ] Executar `go test ./...`, `go build ./...`, `./scripts/lint.sh`
- [ ] Commit via `./scripts/safe-commit.sh`

---

## 5. Documentação Interina

Até a renomeação ocorrer, todos os arquivos em `orchestrator/` devem conter uma nota de alerta:

```markdown
> **⚠️ FUTURE RENAME:** This module will be renamed to `runner/` or `taskflow/`.
> The name "orchestrator" will be reserved for a future Agent Orchestrator module.
> See `docs/adr/0027-orchestrator-module-naming.md`.
```

Isso garante que agentes de IA futuros não confundam o propósito do módulo.
