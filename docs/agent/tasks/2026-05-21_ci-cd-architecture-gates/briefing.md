---
tipo: briefing
task-id: 2026-05-21_ci-cd-architecture-gates
domain: transversal
affects:
  - .github/workflows/ci.yml
  - .github/workflows/pr-gate.yml
origem: decisao humana
branch: feature/2026-05-21_ci-cd-architecture-gates
status: em-andamento
---

# Briefing: CI/CD Architecture Gates

## Contexto

Os workflows de CI/CD atualmente rodam testes que não detectam a maioria das violações reais. A auditoria identificou que:
1. `TestModuleBoundaries` usa uma whitelist de 44 imports cross-module permitidos
2. `TestCodeAnomalies` não detecta business logic em repository, `_ = variável`, SQL sem FROM
3. Não existe job para verificar se `cmd/` instancia repositories/services diretamente (bypass DI)
4. A estrutura de jobs reflete a arquitetura antiga (ADR-0022) com muitos checks desnecessários

Com a ADR-0030, os gates devem ser simplificados para refletir as 3-4 regras principais.

## Motivação

- **Problema:** O CI/CD é a última linha de defesa, mas está cego para as violações mais comuns.
- **Custo:** PRs com violações arquiteturais são mergeados, acumulando dívida técnica.

## Escopo

### Dentro do escopo
- Simplificar job `architecture` em `ci.yml`: rodar apenas os ~3-4 testes simplificados
- Adicionar job `module-boundaries` (zero imports cross-module)
- Adicionar job `repository-purity` (business logic em repository.go)
- Adicionar job `domain-integrity` (entity types em domain/)
- Adicionar job `ignored-errors` (code anomalies)
- Adicionar job `bootstrap-di-check` (cmd/ não bypassa bootstrap)
- Garantir que cada job tenha mensagem de falha clara no PR UI
- Garantir que jobs novos bloqueiem merge (PR Gate)

### Fora do escopo
- Implementação dos testes de arquitetura em si (task separada)
- Scripts locais (task separada)
- Refatoração de código (task separada)

## Arquivos Relevantes
- `.github/workflows/ci.yml`
- `.github/workflows/pr-gate.yml`
- `tests/architecture/` (testes que os jobs executarão)

## Resumo

Simplificar e atualizar os workflows de CI/CD para refletir a arquitetura simplificada (ADR-0030), adicionando gates que detectam as violações reais.

## Entradas
- Workflows atuais: `.github/workflows/ci.yml` e `.github/workflows/pr-gate.yml`
- Testes de arquitetura em `tests/architecture/` (simplificados)
- ADR-0030

## Saídas Esperadas

Workflows atualizados com jobs simplificados:
- `module-boundaries`
- `repository-purity`
- `domain-integrity`
- `ignored-errors`
- `bootstrap-di-check`

## Critérios de Aceitação
- [ ] `ci.yml` contém os jobs simplificados
- [ ] `pr-gate.yml` contém os jobs correspondentes (com prefixo `PR Gate /`)
- [ ] Cada job tem mensagem de falha descritiva no `run:`
- [ ] Jobs estão na seção correta (PR Gate jobs bloqueiam merge)
- [ ] Workflows são válidos YAML
- [ ] Documentação dos workflows explica o propósito de cada job
