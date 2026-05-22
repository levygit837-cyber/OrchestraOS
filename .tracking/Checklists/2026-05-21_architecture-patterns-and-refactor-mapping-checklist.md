# CHECKLIST — Architecture Patterns and Refactor Mapping

**Agent:** Kimi Code CLI  
**Started:** 2026-05-21T00:24:00-03:00  
**Updated:** 2026-05-21T00:45:00-03:00  
**Status:** completed  
**Branch:** feature/2026-05-21_architecture-patterns-and-refactor-mapping  
**Plan Ref:** docs/agent/tasks/2026-05-21_architecture-patterns-and-refactor-mapping/plan.md

---

## Execution Checklist

### Domínio: Inventário Completo de Tipos
- [x] 1. Parsear `internal/modules/task/models.go` — listar structs, enums, types
- [x] 2. Parsear `internal/modules/run/models.go` — listar structs, enums, types
- [x] 3. Parsear `internal/modules/workunit/models.go` — listar structs, enums, types
- [x] 4. Parsear `internal/modules/agent/models.go` — listar structs, enums, types
- [x] 5. Parsear `internal/modules/agentsession/models.go` — listar structs, enums, types
- [x] 6. Parsear `internal/modules/taskgraph/models.go` — listar structs, enums, types
- [x] 7. Parsear `internal/modules/prompt/models.go` — listar structs, enums, types
- [x] 8. Parsear `internal/modules/review/models.go` — listar structs, enums, types
- [x] 9. Parsear `internal/modules/trigger/models.go` — listar structs, enums, types
- [x] 10. Parsear `internal/modules/orchestrator/models.go` — listar structs, enums, types

### Domínio: Classificação
- [x] 11. Definir critério formal: shared = usado em 2+ módulos OU em bootstrap/services.go OU em orchestrator/
- [x] 12. Classificar cada tipo de cada módulo
- [x] 13. Documentar raciocínio para tipos na fronteira (ex: PlanWorkUnit em taskgraph)

### Domínio: Mapeamento de Imports Impactados
- [x] 14. Analisar `internal/bootstrap/services.go` — listar todos os usos de tipos de módulos
- [x] 15. Analisar `internal/modules/orchestrator/models.go` — listar todos os imports cross-module
- [x] 16. Analisar imports cross-module em todos os módulos
- [x] 17. Mapear conversões necessárias (ex: `taskmod.Task` → `domain.Task`)

### Domínio: Documentação
- [x] 18. Criar `MIGRATION_MAP.md` com formato padronizado
- [x] 19. Revisar consistência entre classificações
- [x] 20. Validar que nenhum tipo essencial foi esquecido
- [x] 21. Commit na branch

---

## Annotations

### 2026-05-21 00:30
- Branch criada: `feature/2026-05-21_architecture-patterns-and-refactor-mapping`
- Inventário completo dos 10 módulos realizado
- Domain atual (`internal/domain/`) já contém: EventEnvelope, EventPriority, checkpoint types, e event payloads (TaskGraphNodeInfo, TaskGraphEdgeInfo, TaskGraphCreatedPayload, etc.)
- Cada módulo tem seu próprio `Status` string type — na migração serão renomeados para evitar conflitos (TaskStatus, RunStatus, etc.)

### 2026-05-21 00:35
- Critério formal definido: shared = usado em 2+ módulos OU bootstrap/services.go OU orchestrator/models.go
- 26 entity types identificados como shared (matching o teste de arquitetura domain_import_integrity_test.go)
- PlanWorkUnit é um caso especial: espelha workunit.WorkUnit para evitar import cycle. Será eliminado quando WorkUnit estiver em domain.
- Tipos de payload/evento (TaskGraphCreatedPayload, etc.) já estão em domain/event_payloads.go

### 2026-05-21 00:45
- MIGRATION_MAP.md criado com 31KB de documentação
- 25 entity types shared mapeados para domain (mais 3 payloads já existentes)
- ~30+ tipos locais identificados
- 15+ arquivos com imports impactados mapeados
- Ordem de migração definida (task → agent → agentsession → run → workunit → taskgraph → prompt → review → trigger → orchestrator → bootstrap)
- Commit realizado na branch

---

## Delivery
**Completed:** 2026-05-21T00:50:00-03:00  
**Status:** completed
