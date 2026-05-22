---
tipo: briefing
task-id: 2026-05-21_architecture-test-suite-hardening
domain: transversal
affects:
  - tests/architecture/
  - internal/modules/*
  - internal/core/*
origem: decisao humana
branch: feature/2026-05-21_architecture-test-suite-hardening
status: em-andamento
---

# Briefing: Architecture Test Suite Simplification

## Contexto

A auditoria de confiabilidade arquitetural (2026-05-21) revelou que **todos os testes de arquitetura passam** enquanto **83+ violações reais existem no código**. Os testes atuais verificam *estrutura* (presença de arquivos, existência de imports na whitelist) mas não verificam *comportamento*.

A decisão arquitetural **ADR-0030** simplifica a arquitetura de Vertical Slice para Modular Monolith:
- `internal/domain/` centraliza TODOS os tipos compartilhados
- Módulos NÃO importam outros módulos (zero exceções)
- `repository.go` é CRUD puro
- Testes de arquitetura reduzidos de ~10 para ~3-4 testes simples

## Motivação

- **Problema:** Testes de arquitetura são uma fachada de segurança. A lista `allowedModuleImports` em `module_boundaries_test.go` permite 44 imports cross-module.
- **Custo:** A cada novo PR, violações se acumulam. A arquitetura perde valor. Extração de módulos para microsserviços se torna impossível.

## Escopo

### Dentro do escopo
- Implementar `TestModuleBoundaries` simplificado — verifica que **nenhum** módulo importa outro módulo (sem whitelist)
- Implementar `TestRepositoryPurity` — detecta business logic em `repository.go` via heurísticas AST
- Implementar `TestDomainImportIntegrity` — verifica que entity types compartilhados estão em `internal/domain/`
- Manter `TestCodeAnomalies` — detectar `_ = variável`, SQL inline, `panic()`, `fmt.Println`
- Garantir que TODOS os novos testes falhem com as violações ATUAIS (prova de que detectam o problema real)

### Fora do escopo
- Correção das violações detectadas (task separada: `2026-05-21_code-refactor-for-architecture-violations`)
- Mudanças em scripts ou CI/CD (task separada)
- Definição de novos padrões arquiteturais (task separada)

## Arquivos Relevantes
- `tests/architecture/module_boundaries_test.go` (teste atual a ser simplificado)
- `tests/architecture/code_anomalies_test.go` (teste atual a ser mantido)
- `tests/architecture/domain_purity_test.go` (teste a ser invertido)
- `tests/architecture/module_contract_test.go` (teste a ser removido)
- `internal/modules/*/*.go` (código alvo dos testes)
- `docs/adr/0030-simplified-modular-architecture.md` (regras de referência)
- `docs/development/CODING_STANDARDS.md`

## Resumo

Simplificar os testes de arquitetura para refletir a ADR-0030: 3-4 testes que detectam violações reais de comportamento, não apenas estrutura.

## Entradas
- Código fonte atual em `internal/modules/*` e `internal/core/*`
- ADR-0030 (regras simplificadas)
- CODING_STANDARDS.md
- Lista de violações conhecidas do relatório `docs/audit/architecture-reliability-audit-2026-05-21.md`

## Saídas Esperadas

Arquivos de teste em `tests/architecture/`:
1. `module_boundaries_test.go` (simplificado — zero whitelist)
2. `repository_purity_test.go` (novo ou atualizado)
3. `domain_import_integrity_test.go` (novo — inverte lógica de domain_purity_test.go)
4. `code_anomalies_test.go` (atualizado)

## Critérios de Aceitação
- [ ] `TestModuleBoundaries` falha com as violações atuais (44 imports cross-module)
- [ ] `TestRepositoryPurity` falha com as violações atuais (business logic em repository.go)
- [ ] `TestDomainImportIntegrity` falha se entity types não estiverem em `internal/domain/`
- [ ] `TestCodeAnomalies` detecta `_ = variável` sem comentário, SQL inline, panic, fmt.Println
- [ ] Todos os novos testes usam mensagens de erro claras com arquivo, linha e descrição da regra quebrada
- [ ] `go test ./tests/architecture/...` compila e roda sem panic
- [ ] Documentação dos testes explica a heurística e como adicionar exceções se necessário

## Notas Técnicas
- Usar `go/ast`, `go/parser`, `go/token` para inspeção estática
- `orchestrator` e `bootstrap` são os únicos pacotes que podem importar múltiplos módulos
- `internal/domain/` deve conter todos os entity types compartilhados (Task, Run, WorkUnit, Agent, AgentSession, TaskGraph, PromptSnapshot, ToolsetSnapshot, Trigger, Review)
