---
tipo: plan
task-id: 2026-05-21_architecture-patterns-and-refactor-mapping
domain: transversal
status: em-andamento
---

# Plan: Architecture Patterns and Refactor Mapping

## Tipo: Por Domínio

## Domínio: Inventário Completo de Tipos
- [ ] Parsear `internal/modules/task/models.go` — listar structs, enums, types
- [ ] Parsear `internal/modules/run/models.go` — listar structs, enums, types
- [ ] Parsear `internal/modules/workunit/models.go` — listar structs, enums, types
- [ ] Parsear `internal/modules/agent/models.go` — listar structs, enums, types
- [ ] Parsear `internal/modules/agentsession/models.go` — listar structs, enums, types
- [ ] Parsear `internal/modules/taskgraph/models.go` — listar structs, enums, types
- [ ] Parsear `internal/modules/prompt/models.go` — listar structs, enums, types
- [ ] Parsear `internal/modules/review/models.go` — listar structs, enums, types
- [ ] Parsear `internal/modules/trigger/models.go` — listar structs, enums, types
- [ ] Parsear `internal/modules/orchestrator/models.go` — listar structs, enums, types

## Domínio: Classificação
- [ ] Definir critério formal: shared = usado em 2+ módulos OU em bootstrap/services.go OU em orchestrator/
- [ ] Classificar cada tipo de cada módulo
- [ ] Documentar raciocínio para tipos na fronteira (ex: PlanWorkUnit em taskgraph)

## Domínio: Mapeamento de Imports Impactados
- [ ] Analisar `internal/bootstrap/services.go` — listar todos os usos de tipos de módulos
- [ ] Analisar `internal/modules/orchestrator/models.go` — listar todos os imports cross-module
- [ ] Analisar `cmd/*.go` — listar usos de tipos de módulos
- [ ] Mapear conversões necessárias (ex: `taskmod.Task` → `domain.Task`)

## Domínio: Documentação
- [ ] Criar `MIGRATION_MAP.md` com formato padronizado
- [ ] Revisar consistência entre classificações
- [ ] Validar que nenhum tipo essencial foi esquecido
- [ ] Commit na branch
