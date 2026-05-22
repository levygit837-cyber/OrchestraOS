---
tipo: plan
task-id: 2026-05-21_script-integration-for-architecture-defense
domain: transversal
status: em-andamento
---

# Plan: Script Integration for Architecture Defense

## Tipo: Faseado

## Fase 1: Análise dos Scripts Atuais
- [ ] Ler `scripts/git/pre-commit.sh`, `scripts/go/lint.sh`, `scripts/go/verify-contracts.sh`, `scripts/go/verify-module-structure.sh`
- [ ] Identificar gaps exatos de integração
- [ ] Verificar dependências entre scripts (safe-commit.sh -> pre-commit.sh -> ...)

## Fase 2: Simplificar verify-module-structure.sh
- [ ] Reduzir `MANDATORY_FILES` de 10 para 5
- [ ] Manter: doc.go, README.md, models.go, repository.go, service.go
- [ ] Remover obrigatoriedade: contract.go, CONTRACTS.md, events.go, queries.go, validation.go
- [ ] Testar script manualmente: `./scripts/go/verify-module-structure.sh`

## Fase 3: Atualizar pre-commit.sh
- [ ] Adicionar chamada a `./scripts/go/verify-module-structure.sh` após `go test ./tests/architecture/...`
- [ ] Garantir `set -euo pipefail` está presente
- [ ] Testar script manualmente: `./scripts/git/pre-commit.sh`

## Fase 4: Atualizar lint.sh
- [ ] Adicionar chamada a `./scripts/go/verify-module-structure.sh` antes de `golangci-lint`
- [ ] Garantir que `go test ./tests/architecture/...` ainda está presente
- [ ] Testar script manualmente: `./scripts/go/lint.sh`

## Fase 5: Atualizar verify-contracts.sh
- [ ] Atualizar comentário do cabeçalho para refletir suite completa
- [ ] Garantir que comando roda `go test ./tests/architecture/... -count=1`

## Fase 6: Atualizar AGENTS.md
- [ ] Verificar seção "Fluxo Obrigatório"
- [ ] Verificar seção "Commits e Branches"
- [ ] Adicionar instrução explicita sobre execução de `lint.sh` + `verify-contracts.sh`
- [ ] Adicionar nota de alerta sobre não confiar apenas em `go test ./...`

## Fase 7: Validação
- [ ] Rodar `./scripts/git/pre-commit.sh` e garantir que passa (com código atual)
- [ ] Rodar `./scripts/go/lint.sh` e garantir que passa
- [ ] Simular falha em `verify-module-structure.sh` e garantir que `pre-commit.sh` falha
- [ ] Commit na branch
