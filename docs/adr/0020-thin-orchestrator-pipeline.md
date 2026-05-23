# 0020. Thin Orchestrator — Pipeline Architecture

**Status:** Accepted
**Data:** 2026-05-23
**Supersedes:** ADR-0019

---

## 1. Contexto

ADR-0019 simplificou a arquitetura modular, mas o codebase ainda tinha 10 módulos, 22.000 linhas de documentação, e um ratio docs:code de 1.3:1. O sistema nunca executou uma task real. 23% dos commits eram refatorações.

### 1.1 Problemas identificados

| Problema | Evidência |
|----------|-----------|
| Over-engineering | 10 módulos para zero funcionalidade executável |
| God Object | orchestrator/ importava 11 deps em 517 linhas |
| Documentação > Código | 22.000 linhas docs vs 16.700 linhas código |
| Instabilidade | 3 mudanças de arquitetura em 21 dias |

---

## 2. Decisão

Pivot para **Thin Orchestrator** — pipeline architecture com 5+2 packages:

```
internal/
  domain/       # Tipos puros (zero imports internos)
  planner/      # Task → DAG de WorkUnits
  executor/     # DAG → execução topológica
  runtime/      # Interface de execução do agente
  store/        # Persistência unificada
  event/        # Event emitter
  apperrors/    # Erros padronizados
```

### 2.1 Três regras

1. `domain/` é puro — zero imports de packages internos.
2. Dependências fluem para baixo — todos dependem de domain; nunca entre si.
3. SQL confinado a `store/` — nenhum outro package contém SQL.

### 2.2 Pipeline

```
Task → planner.Plan() → []WorkUnit (DAG)
     → executor.Execute() → topological sort → runtime.Execute() per WU
     → store persists all state transitions
```

O orchestrator é ~55 linhas que compõe planner→executor→runtime.

---

## 3. Consequências

### Positivas

- Codebase de ~1200 linhas (era 16.700)
- Documentação de ~500 linhas (era 22.000)
- Ratio docs:code de ~0.4:1 (era 1.3:1)
- Pipeline funcional end-to-end via CLI
- 6 testes de arquitetura que validam as 3 regras

### Negativas

- Perda do código dos módulos antigos (preservado no git history)
- Sem PostgreSQL store ainda (in-memory apenas)
- Sem runtime real (fake runtime que sempre sucede)

---

## 4. Alternativas consideradas

1. **Corrigir ADR-0019** — manter 10 módulos, corrigir violações. Rejeitado: complexidade excessiva para zero funcionalidade.
2. **Rewrite completo** — começar do zero. Rejeitado: domain types e heuristic planner eram bons.
3. **Thin Orchestrator** (escolhido) — preservar domain types, reconstruir como pipeline.

---

## 5. Testes de arquitetura

| Teste | Regra |
|-------|-------|
| TestDependencyDirection | Grafo de imports por package |
| TestDomainPurity | domain/ zero imports internos |
| TestPackageSizeLimit | Max 800 linhas por package |
| TestMaxFunctionComplexity | Max 40 linhas por função |
| TestSQLConfinement | SQL só em store/ |
| TestNoGlobalState | Sem variáveis globais mutáveis |
