---
tipo: plan
task-id: 2026-05-21_ci-cd-architecture-gates
domain: transversal
status: concluido
---

# Plan: CI/CD Architecture Gates

## Tipo: Faseado

## Fase 1: Análise dos Workflows Atuais
- [x] 1.1 Ler `.github/workflows/ci.yml` completo
- [x] 1.2 Ler `.github/workflows/pr-gate.yml` completo
- [x] 1.3 Mapear jobs existentes e suas categorias
- [x] 1.4 Decidir onde inserir novos jobs e quais remover/simplificar

**Findings:** 9 testes removidos na T1 ainda eram referenciados nos workflows (TestTransitionPackageIsLeaf, TestModulesDoNotImportCoordination, TestModuleRequiredFiles, TestCoreRequiredFiles, TestContractsSync, TestQueriesPurity, TestForbiddenFilenames, TestDomainPurity, TestCoordinationRemoved).

## Fase 2: Simplificar ci.yml
- [x] 2.1 Simplificar job `architecture` existente
- [x] 2.2 Adicionar job `module-boundaries` (via TestModuleBoundaries)
- [x] 2.3 Adicionar job `repository-purity` (via TestRepositoryPurity + TestRepositoryMethodNames)
- [x] 2.4 Adicionar job `domain-integrity` (via TestDomainImportIntegrity)
- [x] 2.5 Adicionar job `ignored-errors` (via TestCodeAnomalies)
- [x] 2.6 Adicionar job `bootstrap-di-check` (via TestCmdBootstrapDI)
- [x] 2.7 Garantir que cada job tenha `actions/checkout@v4` e `actions/setup-go@v5`
- [x] 2.8 Garantir mensagens de falha descritivas

## Fase 3: Atualizar pr-gate.yml
- [x] 3.1 Adicionar jobs correspondentes com prefixo `PR Gate /`
- [x] 3.2 Garantir que jobs estão na seção de gates que bloqueiam merge
- [x] 3.3 Garantir consistência com `ci.yml`

## Fase 4: bootstrap-di-check
- [x] 4.1 Definir regex/grep exato para detectar instanciação direta em `cmd/`
- [x] 4.2 Testar regex localmente: `grep -rn "NewService\|NewRepository" cmd/`
- [x] 4.3 Implementar job no workflow (via TestCmdBootstrapDI)

## Fase 5: Validacao
- [x] 5.1 Validar YAML com revisão manual
- [x] 5.2 Verificar que jobs novos usam `go-version-file: go.mod` consistentemente
- [x] 5.3 Commit na branch
