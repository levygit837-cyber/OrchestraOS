# Instruções Para Agentes

Este arquivo deve ser lido por qualquer agente antes de editar o projeto.

## O Projeto

O OrchestraOS é um sistema de orquestração de agentes de IA. Ele transforma intenção humana em planejamento, execução, validação e operação contínua. O sistema é local-first, com desenho compatível para servidor. A fonte de verdade é este repositório.

- **Fase atual:** Thin Orchestrator — pipeline architecture
- **Autonomia aprovada:** Nível 2 (IA implementa com revisão humana)
- **Stack:** Go, Postgres, GitHub

## Arquitetura (Thin Orchestrator)

Pipeline architecture com 3 regras:

1. **`internal/domain/`** é puro — zero imports de packages internos, zero lógica de negócio.
2. **Dependências fluem para baixo** — planner, executor, runtime dependem de domain e store; nunca entre si.
3. **SQL vive em `store/queries.go`** — nenhum outro package contém SQL.

```text
internal/
  domain/       # Tipos puros (Task, Run, WorkUnit, TaskGraph, EventEnvelope)
  planner/      # Task → DAG de WorkUnits (heuristic + gemini)
  executor/     # DAG → Execução ordenada
  runtime/      # Interface de execução do agente
  store/        # Persistência unificada
  event/        # Event store simplificado
  apperrors/    # Erros padronizados
```

## Prioridades

1. Preservar a intenção do usuário.
2. Manter o repositório como fonte de verdade.
3. Fazer mudanças pequenas, verificáveis e reversíveis.
4. Nunca tratar conversa solta, chat, comentários avulsos ou memória do agente como fonte definitiva.

## Navegação

| Para... | Consulte |
|---------|----------|
| Entender o produto e visão | `docs/canvas/project-canvas.md` |
| Decisões arquiteturais | `docs/adr/` |
| Regras de código | `docs/development/CODING_STANDARDS.md` |

## Fluxo de Trabalho

1. Entender o item de trabalho.
2. Consultar o canvas e as ADRs relevantes.
3. Implementar a menor mudança suficiente.
4. Rodar validações (`go vet`, `go test ./...`, `go build ./...`).
5. Registrar o que mudou e qualquer risco restante.

## Regras de Commit

- **Nunca** commit ou push diretamente na branch `master`.
- Usar feature branch e abrir Pull Request.
- Rodar `make check` antes de commitar.

## Decisões

Decisões arquiteturais relevantes devem virar ADR em `docs/adr/`.

## Política de Autonomia

Atualmente: **Nível 2** (IA implementa com revisão humana).
