---
tipo: briefing
task-id: 2026-05-21_architecture-patterns-and-refactor-mapping
domain: transversal
affects:
  - internal/modules/*/models.go
  - internal/domain/
  - internal/bootstrap/services.go
  - internal/modules/orchestrator/models.go
origem: decisao humana
branch: feature/2026-05-21_architecture-patterns-and-refactor-mapping
status: em-andamento
---

# Briefing: Architecture Patterns and Refactor Mapping

## Contexto

A ADR-0030 decide que `internal/domain/` centraliza TODOS os tipos compartilhados. Atualmente, cada módulo em `internal/modules/*` define seus próprios tipos em `models.go` (778 linhas total). Isso cria:
- Duplicação implícita (muitos módulos referenciam os mesmos conceitos)
- Imports cross-module (44 imports permitidos pela whitelist)
- Dificuldade para agentes entenderem onde os tipos "oficiais" residem

## Motivação

- **Problema:** Não há mapa claro de quais tipos devem migrar para `internal/domain/` e quais devem ficar nos módulos.
- **Custo:** Sem um plano de migração detalhado, a refatoração pode mover tipos errados, quebrar interfaces e criar regressões.

## Escopo

### Dentro do escopo
- Inventariar TODOS os structs e enums em cada `internal/modules/*/models.go`
- Determinar quais são "shared entity types" (usados por 2+ módulos) vs "module-local types"
- Criar `MIGRATION_MAP.md` documentando o que move para `internal/domain/`
- Mapear imports impactados em `internal/bootstrap/services.go`
- Mapear imports no `internal/modules/orchestrator/models.go` (que importa 9 módulos)
- Documentar padrões de conversão (ex: `task.Task` → `domain.Task` em X lugares)
- **NÃO implementar código** — apenas definir o que deve ser feito

### Fora do escopo
- Implementação da migração (task separada: `2026-05-21_code-refactor-for-architecture-violations`)
- Testes de arquitetura (task separada)
- Scripts e CI/CD (tasks separadas)

## Arquivos Relevantes
- `internal/modules/*/models.go` (10 arquivos)
- `internal/domain/types.go`, `internal/domain/doc.go`
- `internal/bootstrap/services.go`
- `internal/modules/orchestrator/models.go`
- `docs/adr/0030-simplified-modular-architecture.md`

## Resumo

Mapear todos os tipos dos 10 módulos, determinar o que vai para `internal/domain/`, e criar um plano de migração detalhado.

## Entradas
- Código fonte atual em `internal/modules/*/models.go`
- ADR-0030 (regras de centralização de tipos)
- `internal/bootstrap/services.go` (dependências)

## Saídas Esperadas

1. `MIGRATION_MAP.md` em `docs/agent/tasks/2026-05-21_architecture-patterns-and-refactor-mapping/` com:
   - Lista completa de structs/enums por módulo
   - Classificação: shared (vai para domain/) vs local (fica no módulo)
   - Lista de arquivos que precisam de atualização de imports
   - Notas sobre conversões especiais (ex: tipos com nomes conflitantes)

## Critérios de Aceitação
- [ ] Todos os 10 `models.go` foram inventariados
- [ ] Cada tipo foi classificado como shared ou local
- [ ] `MIGRATION_MAP.md` está completo e revisável
- [ ] Imports impactados em `bootstrap/services.go` estão mapeados
- [ ] Imports impactados em `orchestrator/models.go` estão mapeados
- [ ] Documento é suficiente para que outro agente execute a migração sem dúvidas
