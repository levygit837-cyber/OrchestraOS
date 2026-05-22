---
tipo: briefing
task-id: 2026-05-21_code-refactor-for-architecture-violations
domain: transversal
affects:
  - internal/modules/*/
  - internal/domain/
  - internal/bootstrap/
  - tests/architecture/
origem: decisao humana
branch: feature/2026-05-21_code-refactor-for-architecture-violations
status: em-andamento
---

# Briefing: Code Refactor for Architecture Violations

## Contexto

A auditoria identificou violações arquiteturais que precisam ser corrigidas para implementar a ADR-0030:

1. **Tipos espalhados pelos módulos:** Entity types (Task, Run, WorkUnit, etc.) estão definidos em `internal/modules/*/models.go` em vez de `internal/domain/`
2. **44 imports cross-module:** Módulos importam tipos uns dos outros via `allowedModuleImports`
3. **`orchestrator/models.go` importa 9 módulos:** Deve usar `internal/domain/` em vez de imports individuais
4. **Business logic em repository.go:** Vários módulos têm lógica de negócio em repositories
5. **`cmd/` bypassa bootstrap DI:** Instancia repositories diretamente

## Motivação

- **Problema:** O código atual viola as regras simplificadas da ADR-0030.
- **Custo:** Sem correção, os novos testes de arquitetura (Task T1) falham permanentemente, bloqueando o CI.

## Escopo

### Dentro do escopo
- Mover shared entity types de `internal/modules/*/models.go` → `internal/domain/`
- Atualizar todos os imports nos 10 módulos
- Atualizar `internal/bootstrap/services.go`
- Atualizar `cmd/*.go` (usar bootstrap para DI)
- Simplificar `internal/modules/orchestrator/models.go`
- Mover business logic de `repository.go` para `service.go` quando detectado
- Atualizar testes unitários que quebrarem
- Garantir que `go build ./...` e `go test ./...` passem

### Fora do escopo
- Criar novos testes de arquitetura (task separada)
- Definir padrões (task separada)
- Scripts e CI/CD (tasks separadas)
- Mudanças funcionais (o comportamento deve permanecer idêntico)

## Arquivos Relevantes
- Todos os `internal/modules/*/*.go`
- `internal/domain/types.go`, `internal/domain/doc.go`
- `internal/bootstrap/services.go`
- `cmd/*.go`
- `tests/architecture/` (testes que devem passar após refatoração)

## Resumo

Refatorar o código para implementar a ADR-0030: centralizar tipos em `internal/domain/`, zerar imports cross-module, e purificar repositories.

## Entradas
- `MIGRATION_MAP.md` da task `2026-05-21_architecture-patterns-and-refactor-mapping`
- ADR-0030
- Testes de arquitetura simplificados

## Saídas Esperadas
- Código refatorado em todos os módulos afetados
- `internal/domain/` contém todos os shared entity types
- `go test ./tests/architecture/...` PASSA
- `go test ./...` PASSA
- `go build ./...` PASSA

## Critérios de Aceitação
- [ ] `TestModuleBoundaries` passa (zero imports cross-module)
- [ ] `TestRepositoryPurity` passa (zero business logic em repository.go)
- [ ] `TestDomainImportIntegrity` passa (entity types em domain/)
- [ ] `TestCodeAnomalies` passa
- [ ] `go test ./...` passa (nenhum teste unitário quebrado)
- [ ] `go build ./...` passa
- [ ] Comportamento funcional inalterado (refatoração pura, sem mudança de lógica)

## Notas Técnicas
- Esta task pode ser grande. Se necessário, dividir em sub-PRs por módulo.
- Priorizar: 1) mover tipos para domain/, 2) atualizar imports, 3) purificar repositories
- `orchestrator/models.go` é o arquivo mais impactado (9 imports cross-module)
- `bootstrap/services.go` provavelmente precisará de ajustes nos adapters
