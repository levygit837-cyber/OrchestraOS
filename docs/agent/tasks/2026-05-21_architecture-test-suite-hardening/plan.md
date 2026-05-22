---
tipo: plan
task-id: 2026-05-21_architecture-test-suite-hardening
domain: transversal
status: em-andamento
---

# Plan: Architecture Test Suite Simplification

## Tipo: Por Domínio

## Fase 1: Setup e Análise
- [ ] Ler relatório de auditoria: `docs/audit/architecture-reliability-audit-2026-05-21.md`
- [ ] Estudar testes atuais em `tests/architecture/` para entender padrões existentes
- [ ] Decidir quais testes manter, quais remover, quais simplificar
- [ ] Definir estrutura de AST inspection para cada novo teste

## Fase 2: Implementar TestModuleBoundaries (simplificado)
- [ ] Substituir `module_boundaries_test.go` com versão sem whitelist
- [ ] Lógica: qualquer import de `internal/modules/X` em `internal/modules/Y` (onde X≠Y) falha
- [ ] Exceções: `orchestrator/` e `bootstrap/`
- [ ] Rodar teste e verificar que detecta 44 violações atuais
- [ ] Ajustar mensagens de erro

## Fase 3: Implementar TestRepositoryPurity
- [ ] Criar/atualizar `repository_purity_test.go`
- [ ] Implementar detecção de `if status == Status*` em repository.go
- [ ] Implementar detecção de `time.Now()` / `time.Now().UTC()` em repository.go
- [ ] Implementar detecção de `ON CONFLICT` em repository.go
- [ ] Rodar teste e verificar que detecta violações conhecidas

## Fase 4: Implementar TestDomainImportIntegrity
- [ ] Criar `domain_import_integrity_test.go`
- [ ] Definir lista de entity types que DEVEM estar em `internal/domain/`
- [ ] Verificar que cada entity type está definido em `internal/domain/*.go`
- [ ] Verificar que módulos importam `internal/domain` para usar esses tipos
- [ ] Rodar teste (deve falhar inicialmente, pois tipos ainda estão nos módulos)

## Fase 5: Corrigir TestCodeAnomalies
- [ ] Adicionar detecção de `_ = <ident>` (variável, não call)
- [ ] Adicionar detecção de `_ = call()` dentro de `defer func() { ... }()`
- [ ] Expandir regex SQL para `SELECT \w+\(` (sem FROM)
- [ ] Rodar teste e verificar que detecta anomalias atuais

## Fase 6: Remover Testes Obsoletos
- [ ] Remover `module_contract_test.go` (exigência de contract.go com //go:embed)
- [ ] Remover `forbidden_filenames_test.go` se não for mais relevante
- [ ] Remover `contracts_sync_test.go`
- [ ] Remover `queries_purity_test.go` se não for mais relevante
- [ ] Remover `transition_imports_test.go`
- [ ] Remover `coordination_removed_test.go` se não for mais relevante

## Fase 7: Integração e Documentação
- [ ] Garantir que `go test ./tests/architecture/...` compila sem erros
- [ ] Atualizar comentários dos testes explicando heurísticas
- [ ] Criar `tests/architecture/README.md` documentando cada teste

## Fase 8: Validação Final
- [ ] Rodar suite completa: `go test ./tests/architecture/... -v`
- [ ] Verificar que NOVOS testes falham (provando que detectam violações reais)
- [ ] Commit na branch com mensagem descritiva
