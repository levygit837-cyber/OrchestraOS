# Instruções Para Agentes

Este arquivo deve ser lido por qualquer agente antes de editar o projeto.

## O Projeto

O OrchestraOS é um sistema de orquestração de agentes de IA. Ele transforma intenção humana em planejamento, execução, validação e operação contínua. O sistema é local-first, com desenho compatível para servidor. A fonte de verdade é este repositório.

- **Fase atual:** fundação e integração E2E
- **Autonomia aprovada:** Nível 2 (IA implementa com revisão humana)
- **Stack:** Go, Postgres, WebSocket, Docker, GitHub

## Arquitetura (ADR-0030)

A arquitetura vigente é a **ADR-0030** (Arquitetura Modular Simplificada). Os 4 pilares são:

1. **`internal/domain/`** centraliza todos os entity types compartilhados (Task, Run, WorkUnit, Agent, etc.).
2. **Módulos em `internal/modules/`** não importam outros módulos. Zero exceções.
3. **Apenas `internal/bootstrap/` e `internal/modules/orchestrator/`** importam múltiplos módulos.
4. **`repository.go`** é CRUD puro — sem business logic, sem timestamps, sem deduplication.

```text
internal/
  bootstrap/        # DI e wiring de serviços
  core/             # Infraestrutura compartilhada (apperrors, db, eventstore, statemachine, transition, validation)
  domain/           # Todos os entity types compartilhados
  modules/          # Módulos verticais autônomos
    agent/  agentsession/  orchestrator/  prompt/  review/
    run/    task/           taskgraph/     trigger/ workunit/
```

## Prioridades

1. Preservar a intenção do usuário.
2. Manter o repositório como fonte de verdade.
3. Fazer mudanças pequenas, verificáveis e reversíveis.
4. Atualizar documentação quando a mudança alterar comportamento, arquitetura ou processo.
5. Nunca tratar conversa solta, chat, comentários avulsos ou memória do agente como fonte definitiva.

## Navegação

| Para... | Consulte |
|---------|----------|
| Entender o produto e visão | `docs/canvas/project-canvas.md` |
| Decisões arquiteturais | `docs/adr/` (vigente: `0030-simplified-modular-architecture.md`) |
| Regras de código, naming, erros, testes | `docs/development/CODING_STANDARDS.md` |
| Fluxo de trabalho do agente | `docs/agent/PLAYBOOK.md` |
| Tipos de plano | `docs/development/plan-types.md` |
| Estrutura do repositório | `docs/architecture/core/repo-structure.md` |
| Visão geral da arquitetura | `docs/architecture/README.md` |

## Fluxo de Trabalho

1. Entender o item de trabalho.
2. Consultar o canvas e as ADRs relevantes.
3. Seguir o `docs/agent/PLAYBOOK.md` para gerar artefatos necessários (BRIEFING, SPEC, PLAN quando aplicável).
4. Implementar a menor mudança suficiente.
5. Rodar validações (`go vet`, testes de arquitetura, contracts).
6. Registrar o que mudou e qualquer risco restante.

## Regras de Commit

- **Nunca** commit ou push diretamente na branch `master`.
- Usar feature branch e abrir Pull Request.
- Rodar validações antes de commitar. Consulte `docs/development/CODING_STANDARDS.md` para detalhes.

## Decisões

Decisões arquiteturais relevantes devem virar ADR em `docs/adr/` com: Contexto, Decisão, Consequências e Alternativas consideradas.

## Política de Autonomia

O projeto ganha autonomia por níveis (0 a 5). Nenhum agente deve assumir autonomia maior que a aprovada nos documentos do projeto. Atualmente: **Nível 2**.
