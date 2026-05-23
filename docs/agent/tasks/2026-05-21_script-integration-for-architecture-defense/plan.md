---
tipo: plan
task-id: 2026-05-21_script-integration-for-architecture-defense
domain: transversal
status: concluido
---

# Plan: Script Integration for Architecture Defense

## Tipo: Faseado

## Fase 1: Análise dos Scripts Atuais
- [x] 1.1 Ler `scripts/git/pre-commit.sh`, `scripts/go/lint.sh`, `scripts/go/verify-contracts.sh`, `scripts/go/verify-module-structure.sh`
- [x] 1.2 Identificar gaps exatos de integração
- [x] 1.3 Verificar dependências entre scripts (safe-commit.sh -> pre-commit.sh -> ...)

**Findings:**
- `scripts/git/safe-commit.sh`, `scripts/git/pre-commit.sh`, `scripts/git/pre-push.sh` existem em `scripts/git/` (não em `scripts/` raiz como assumido anteriormente).
- `scripts/go/lint.sh` não chamava `verify-module-structure.sh`.
- `scripts/git/pre-commit.sh` não chamava `verify-module-structure.sh`.
- `scripts/go/verify-contracts.sh` já reflete a suite completa (alias para `go test ./tests/architecture/...`).

## Fase 2: Simplificar verify-module-structure.sh
- [x] 2.1 Reduzir `MANDATORY_FILES` de 10 para 5
- [x] 2.2 Manter: doc.go, README.md, models.go, repository.go, service.go
- [x] 2.3 Remover obrigatoriedade: contract.go, CONTRACTS.md, events.go, queries.go, validation.go
- [x] 2.4 Testar script manualmente: `./scripts/go/verify-module-structure.sh`

**Nota:** O script `verify-module-structure.sh` já foi simplificado em commit anterior (fevereiro/2026) para 5 arquivos. Nenhuma alteração necessária nesta task.

## Fase 3: Atualizar pre-commit.sh
- [x] 3.1 Adicionar chamada a `./scripts/go/verify-module-structure.sh` após `go test ./tests/architecture/...`
- [x] 3.2 Garantir `set -euo pipefail` está presente
- [x] 3.3 Testar script manualmente: `./scripts/git/pre-commit.sh`

## Fase 4: Atualizar lint.sh
- [x] 4.1 Adicionar chamada a `./scripts/go/verify-module-structure.sh` antes de `golangci-lint`
- [x] 4.2 Garantir que `go test ./tests/architecture/...` ainda está presente
- [x] 4.3 Testar script manualmente: `./scripts/go/lint.sh`

## Fase 5: Atualizar verify-contracts.sh
- [x] 5.1 Atualizar comentário do cabeçalho para refletir suite completa
- [x] 5.2 Garantir que comando roda `go test ./tests/architecture/... -count=1`

**Nota:** O script já estava correto. Nenhuma alteração necessária.

## Fase 6: Atualizar AGENTS.md
- [x] 6.1 Verificar seção "Fluxo Obrigatório"
- [x] 6.2 Verificar seção "Commits e Branches"
- [x] 6.3 Corrigir `main` → `master` (o repo usa `master`, não `main`)
- [x] 6.4 Adicionar instrução explicita sobre execução de `lint.sh` + `verify-contracts.sh`
- [x] 6.5 Adicionar nota de alerta sobre não confiar apenas em `go test ./...`

## Fase 7: Validação
- [x] 7.1 Rodar `./scripts/git/pre-commit.sh` e garantir que passa (com código atual)
- [x] 7.2 Rodar `./scripts/go/lint.sh` e garantir que passa
- [x] 7.3 Simular falha em `verify-module-structure.sh` e garantir que `pre-commit.sh` falha
- [x] 7.4 Commit na branch
