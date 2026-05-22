---
tipo: spec
task-id: 2026-05-21_architecture-patterns-and-refactor-mapping
domain: transversal
status: em-andamento
---

# Spec: Architecture Patterns and Refactor Mapping

## Inventário Completo

Para cada módulo, listar:
1. Todos os structs (com campos)
2. Todos os enums (status, priority, etc.)
3. Todos os types (string aliases)
4. Classificação: shared vs local

## Critério de Classificação

**Shared (vai para internal/domain/):**
- Usado como campo em 2+ módulos
- Usado como parâmetro/retorno em interfaces DI de 2+ módulos
- Referenciado em `internal/bootstrap/services.go`
- Referenciado em `internal/modules/orchestrator/`

**Local (fica no módulo):**
- Usado apenas dentro do próprio módulo
- Tipo auxiliar para lógica interna
- Payload/evento específico do módulo

## Formato do MIGRATION_MAP.md

```markdown
# Migration Map: Types to internal/domain/

## Módulo: task

| Tipo | Kind | Classificação | Destino | Notas |
|------|------|--------------|---------|-------|
| Task | struct | shared | domain.Task | Usado por run, workunit, taskgraph |
| Status | string | shared | domain.TaskStatus | Usado por run, workunit |
| Priority | string | shared | domain.TaskPriority | Usado por taskgraph |
| RiskLevel | string | shared | domain.TaskRiskLevel | Usado por orchestrator |

## Módulo: run
...

## Imports Impactados

### bootstrap/services.go
- task.Task → domain.Task (linha X)
- run.Run → domain.Run (linha Y)
...

### orchestrator/models.go
- taskgraphmod.TaskGraph → domain.TaskGraph
...
```

## Critérios de Aceitação
- [ ] Todos os 10 módulos inventariados
- [ ] Cada tipo classificado como shared ou local
- [ ] `MIGRATION_MAP.md` completo
- [ ] Imports impactados mapeados
