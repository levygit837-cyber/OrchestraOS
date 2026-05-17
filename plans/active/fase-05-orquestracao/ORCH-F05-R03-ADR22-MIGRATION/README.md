# Épico: ADR-0022 — Migração de Tipos de Domínio para Módulos Verticais

## Visão Geral

Este épico contém **planos individuais, um por módulo**, para migrar todos os tipos de entidade de `internal/domain/types.go` para seus respectivos módulos verticais em `internal/modules/*`.

Cada plano é autônomo, com seu próprio agente, branch e worktree, seguindo o padrão Ralph Loop (plan.md + checklist.md).

---

## Índice dos Planos

| # | Agente | Módulo | Status | Depende de |
|---|--------|--------|--------|------------|
| 01 | A01 | [task](01-task/plan.md) | 🔵 Pronto | — |
| 02 | A02 | [run](02-run/plan.md) | 🔵 Pronto | #01 task |
| 03 | A03 | [workunit](03-workunit/plan.md) | 🔵 Pronto | #01 task |
| 04 | A04 | [taskgraph](04-taskgraph/plan.md) | 🔵 Pronto | #01 task, #03 workunit |
| 05 | A05 | [agentsession](05-agentsession/plan.md) | 🔵 Pronto | #02 run, #06 agent |
| 06 | A06 | [agent](06-agent/plan.md) | 🔵 Pronto | — |
| 07 | A07 | [prompt](07-prompt/plan.md) | 🔵 Pronto | #01 task, #02 run, #03 workunit, #05 agentsession |
| 08 | A08 | [trigger](08-trigger/plan.md) | 🔵 Pronto | #02 run, #05 agentsession, #03 workunit |
| 09 | A09 | [review](09-review/plan.md) | 🔵 Pronto | #01 task, #02 run, #03 workunit |

### Legenda
- 🔵 Pronto = plano criado, aguardando execução
- 🟡 Em execução = agente trabalhando
- 🟢 Concluído = build+test+commit passaram
- 🔴 Bloqueado = dependência não satisfeita

---

## Ordem de Execução e Paralelização

```
Fase 1 (Sequencial):
  01-task ───────────────────────────────────────────────►

Fase 2 (Paralelo após #01):
  02-run ────────────────►
  03-workunit ───────────►
  04-taskgraph ──────────►
  06-agent ──────────────►

Fase 3 (Paralelo após #02, #03, #05, #06):
  05-agentsession ───────► (depende #02 run + #06 agent)

Fase 4 (Paralelo após #02, #03, #05):
  07-prompt ─────────────►
  08-trigger ────────────►
  09-review ─────────────►
```

**Regra de ouro:** Nenhum agente inicia antes de TODAS as suas dependências estarem no estado 🟢.

---

## Convenções do Épico

### Nomenclatura de Branches
```
adr22-a0{n}-{modulo}-types
```

Exemplo: `adr22-a01-task-types`, `adr22-a02-run-types`

### Nomenclatura de Worktrees
```
../orchestraos-a0{n}-{modulo}
```

Exemplo: `../orchestraos-a01-task`, `../orchestraos-a02-run`

### Padrão de Commits
```bash
./scripts/safe-commit.sh "ADR-0022: migrate {Modulo} types to modules/{modulo}"
```

### Adapters Temporários
Todo adapter temporário criado durante a migração DEVE seguir o padrão:

```go
// TODO[ADR-0022]: remover quando {consumidor} for desacoplado de domain.{Struct}
func toDomain{Struct}(local *{modulo}.{Struct}) *domain.{Struct} { ... }
```

Para encontrar todos os adapters pendentes:
```bash
grep -rn "TODO\[ADR-0022\]" internal/
```

---

## Documentação Base

Todo agente deve ler antes de executar:
1. `docs/adr/0022-llm-optimized-module-architecture.md` — A decisão arquitetural
2. `plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md` — Plano geral monolítico (contexto histórico)
3. `internal/modules/{modulo}/README.md` — Documentação do módulo
4. `internal/modules/{modulo}/CONTRACTS.md` — Contratos e invariantes

---

## Critérios de Aceitação do Épico

- [ ] Todos os 9 módulos estão no estado 🟢
- [ ] `internal/domain/types.go` contém APENAS `EventEnvelope`, `EventPriority` e tipos genéricos
- [ ] `grep -rn "TODO\[ADR-0022\]" internal/` retorna zero resultados (adapters removidos)
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] Teste de arquitetura: `internal/modules/*` NÃO importa `internal/domain` para structs de entidade

---

## Status do Épico

- **Início:** ___
- **Módulos concluídos:** 0/9
- **Adapters pendentes:** TBD
- **Build:** ___
- **Testes:** ___
