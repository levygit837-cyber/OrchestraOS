# DELIVERY — Code Refactor for Architecture Violations

**Agent:** Kimi Code CLI  
**Started:** 2026-05-21T00:55:00-03:00  
**Updated:** 2026-05-21T00:55:00-03:00  
**Status:** in_progress  
**Branch:** feature/2026-05-21_architecture-patterns-and-refactor-mapping  
**Plan Ref:** docs/agent/tasks/2026-05-21_code-refactor-for-architecture-violations/plan.md  
**Migration Map Ref:** docs/agent/tasks/2026-05-21_architecture-patterns-and-refactor-mapping/MIGRATION_MAP.md

---

## Fases

### Fase 0: Pré-requisitos
- [x] MIGRATION_MAP.md disponível
- [x] Branch criada (`feature/2026-05-21_architecture-patterns-and-refactor-mapping`)

### Fase 1: Mover Tipos para domain/
- [ ] Criar `internal/domain/entities.go` com shared entity types
- [ ] Criar aliases nos módulos durante transição
- [ ] `go build ./...` compila

### Fase 2: Atualizar Imports — Módulos de Domínio
- [ ] task
- [ ] run
- [ ] workunit
- [ ] agent
- [ ] agentsession
- [ ] taskgraph
- [ ] prompt
- [ ] review
- [ ] trigger

### Fase 3: Atualizar Imports — Infraestrutura
- [ ] bootstrap/services.go
- [ ] orchestrator/models.go
- [ ] cmd/*.go

### Fase 4: Purificar Repositories
- [ ] agent/repository.go
- [ ] agentsession/repository.go
- [ ] prompt/repository.go
- [ ] run/repository.go
- [ ] core/eventstore/repository.go

### Fase 5: Remover Aliases e Limpar
- [ ] Remover aliases temporários
- [ ] Verificar go vet

### Fase 6: Validar Testes
- [ ] `go test ./tests/architecture/...`
- [ ] `go test ./...`
- [ ] `go build ./...`
- [ ] `./scripts/go/lint.sh`

### Fase 7: Entrega
- [ ] Commit
- [ ] Push
- [ ] PR

---

## Decisões e Anotações

### 2026-05-21 00:55 — Início
Task T5 iniciada na branch existente. Usando MIGRATION_MAP.md como guia.
