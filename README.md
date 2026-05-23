# OrchestraOS

Sistema de orquestração de agentes de IA. Transforma intenção humana em planejamento, execução e validação contínua via DAG de work units.

## Estado

- Fase atual: Thin Orchestrator — pipeline architecture
- Fonte de verdade: este repositório
- Stack: Go, PostgreSQL (planejado), GitHub
- Autonomia aprovada: Nível 2

## Quickstart

```bash
# Build
go build ./cmd/orchestraos

# Run a task with acceptance criteria
./orchestraos run "Add login feature" \
  "Create login form component" \
  "Add authentication service" \
  "[after: 1,2] Integration tests for login flow"
```

## Arquitetura

Pipeline architecture com 3 regras:

1. `domain/` é puro — zero imports internos
2. Dependências fluem para baixo — nunca entre pacotes irmãos
3. SQL confinado a `store/`

```
cmd/orchestraos/       CLI entrypoint
internal/
├── domain/            Tipos puros (Task, Run, WorkUnit, TaskGraph, EventEnvelope)
├── planner/           Task → DAG de WorkUnits (heuristic planner)
├── executor/          DAG → execução em ordem topológica
├── runtime/           Interface de execução (fake runtime)
├── store/             Persistência unificada (interface + in-memory)
├── event/             Event emitter
├── apperrors/         Erros padronizados
└── orchestrator.go    Composição: planner → executor → runtime (~55 linhas)
tests/architecture/    6 testes de arquitetura via AST
```

Pipeline: `Task → planner.Plan() → []WorkUnit → executor.Execute() → runtime.Execute()`

## Testes de Arquitetura

6 testes automatizados que bloqueiam CI:

| Teste | Regra |
|-------|-------|
| TestDependencyDirection | Grafo de imports validado por package |
| TestDomainPurity | domain/ sem imports internos |
| TestPackageSizeLimit | Nenhum package > 800 linhas |
| TestMaxFunctionComplexity | Nenhuma função > 40 linhas |
| TestSQLConfinement | SQL apenas em store/ |
| TestNoGlobalState | Sem variáveis globais mutáveis |

```bash
make arch    # roda testes de arquitetura
make check   # vet + test + arch + lint + build
```

## Documentação

- [AGENTS.md](AGENTS.md) — regras para agentes
- [docs/canvas/project-canvas.md](docs/canvas/project-canvas.md) — visão do produto
- [docs/adr/](docs/adr/) — decisões arquiteturais
- [docs/development/CODING_STANDARDS.md](docs/development/CODING_STANDARDS.md) — padrões de código
