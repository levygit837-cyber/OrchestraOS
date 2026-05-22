---
tipo: spec
task-id: 2026-05-21_code-refactor-for-architecture-violations
domain: transversal
status: em-andamento
---

# Spec: Code Refactor for Architecture Violations

## Fases da Migração

### Fase 1: Mover Shared Entity Types

Para cada tipo classificado como "shared" no `MIGRATION_MAP.md`:
1. Copiar struct/enum de `internal/modules/X/models.go` para `internal/domain/types.go`
2. Renomear se necessário (ex: `task.Status` → `domain.TaskStatus`)
3. Manter type alias no módulo durante transição (opcional)

### Fase 2: Atualizar Imports nos Módulos

Para cada módulo:
1. Substituir imports de outros módulos por `internal/domain`
2. Atualizar referências de tipo (ex: `taskmod.Task` → `domain.Task`)
3. Remover aliases de módulos importados que não são mais necessários

### Fase 3: Atualizar bootstrap/services.go

1. Atualizar adapters para usar `domain.*` types
2. Verificar se interfaces DI precisam mudar de assinatura

### Fase 4: Atualizar orchestrator/models.go

1. Substituir 9 imports de módulos por `internal/domain`
2. Verificar se structs locais do orchestrator precisam de ajustes

### Fase 5: Purificar Repositories

Para cada `repository.go` com business logic:
1. Extrair lógica de status para `service.go`
2. Extrair `time.Now()` para ser passado como parâmetro
3. Extrair deduplication/upsert para `service.go`

### Fase 6: Validar

```bash
go build ./...
go test ./...
go test ./tests/architecture/...
```

## Critérios de Aceitação
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `go test ./tests/architecture/...` passa
- [ ] Comportamento funcional inalterado
