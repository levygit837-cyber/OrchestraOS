---
tipo: plan
task-id: 2026-05-21_ci-cd-architecture-gates
domain: transversal
status: em-andamento
---

# Plan: CI/CD Architecture Gates

## Tipo: Faseado

## Fase 1: Análise dos Workflows Atuais
- [ ] Ler `.github/workflows/ci.yml` completo
- [ ] Ler `.github/workflows/pr-gate.yml` completo
- [ ] Mapear jobs existentes e suas categorias
- [ ] Decidir onde inserir novos jobs e quais remover/simplificar

## Fase 2: Simplificar ci.yml
- [ ] Simplificar job `architecture` existente
- [ ] Adicionar job `module-boundaries`
- [ ] Adicionar job `repository-purity`
- [ ] Adicionar job `domain-integrity`
- [ ] Adicionar job `ignored-errors`
- [ ] Adicionar job `bootstrap-di-check`
- [ ] Garantir que cada job tenha `actions/checkout@v4` e `actions/setup-go@v5`
- [ ] Garantir mensagens de falha descritivas

## Fase 3: Atualizar pr-gate.yml
- [ ] Adicionar jobs correspondentes com prefixo `PR Gate /`
- [ ] Garantir que jobs estão na seção de gates que bloqueiam merge
- [ ] Garantir consistência com `ci.yml`

## Fase 4: bootstrap-di-check
- [ ] Definir regex/grep exato para detectar instanciação direta em `cmd/`
- [ ] Testar regex localmente: `grep -rn "..." cmd/`
- [ ] Implementar job no workflow

## Fase 5: Validacao
- [ ] Validar YAML com `actionlint` (se instalado) ou revisão manual
- [ ] Verificar que jobs novos usam `go-version-file: go.mod` consistentemente
- [ ] Commit na branch
