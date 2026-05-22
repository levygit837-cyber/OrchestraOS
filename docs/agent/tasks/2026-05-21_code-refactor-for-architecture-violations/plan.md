---
tipo: plan
task-id: 2026-05-21_code-refactor-for-architecture-violations
domain: transversal
status: em-andamento
---

# Plan: Code Refactor for Architecture Violations

## Tipo: Faseado

## Fase 0: Pré-requisitos
- [ ] Aguardar merge da task `2026-05-21_architecture-patterns-and-refactor-mapping` (MIGRATION_MAP.md)
- [ ] Criar branch `feature/2026-05-21_code-refactor-for-architecture-violations`

## Fase 1: Mover Tipos para domain/
- [ ] Criar/expandir `internal/domain/types.go` com shared entity types
- [ ] Manter type aliases nos módulos durante transição (se necessário)
- [ ] `go build ./...` deve compilar após cada módulo movido

## Fase 2: Atualizar Imports — Módulos de Domínio
- [ ] Atualizar `internal/modules/task/` para usar `domain.Task`, `domain.TaskStatus`
- [ ] Atualizar `internal/modules/run/` para usar `domain.Run`, `domain.RunStatus`
- [ ] Atualizar `internal/modules/workunit/` para usar `domain.WorkUnit`
- [ ] Atualizar `internal/modules/agent/` para usar `domain.Agent`
- [ ] Atualizar `internal/modules/agentsession/` para usar `domain.AgentSession`
- [ ] Atualizar `internal/modules/taskgraph/` para usar `domain.TaskGraph`
- [ ] Atualizar `internal/modules/prompt/` para usar `domain.PromptSnapshot`
- [ ] Atualizar `internal/modules/review/` para usar `domain.Review`
- [ ] Atualizar `internal/modules/trigger/` para usar `domain.Trigger`

## Fase 3: Atualizar Imports — Infraestrutura
- [ ] Atualizar `internal/bootstrap/services.go`
- [ ] Atualizar `internal/modules/orchestrator/models.go`
- [ ] Atualizar `cmd/*.go`

## Fase 4: Purificar Repositories
- [ ] `agent/repository.go` — passar status como parâmetro
- [ ] `agentsession/repository.go` — passar heartbeatAt como parâmetro
- [ ] `prompt/repository.go` — mover dedup para service
- [ ] `run/repository.go` — passar timestamps como parâmetros
- [ ] `core/eventstore/repository.go` — mover validação para service

## Fase 5: Remover Aliases e Limpar
- [ ] Remover type aliases temporários dos módulos
- [ ] Remover imports cross-module que sobraram
- [ ] Verificar `go vet ./...`

## Fase 6: Validar Testes
- [ ] `go test ./tests/architecture/...` — TODOS passam
- [ ] `go test ./...` — TODOS passam
- [ ] `go build ./...` — passa
- [ ] `./scripts/go/lint.sh` — passa

## Fase 7: Entrega
- [ ] Commit na branch
- [ ] Push e abertura de PR
