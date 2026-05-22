---
tipo: briefing
task-id: 2026-05-21_script-integration-for-architecture-defense
domain: transversal
affects:
  - scripts/git/pre-commit.sh
  - scripts/go/lint.sh
  - scripts/go/verify-contracts.sh
  - scripts/go/verify-module-structure.sh
  - AGENTS.md
origem: decisao humana
branch: feature/2026-05-21_script-integration-for-architecture-defense
status: em-andamento
---

# Briefing: Script Integration for Architecture Defense

## Contexto

Os scripts de validação local têm lacunas críticas:
1. `pre-commit.sh` não chama `verify-module-structure.sh`
2. `lint.sh` não chama `verify-module-structure.sh`
3. `verify-contracts.sh` é apenas um alias para `go test ./tests/architecture/...`
4. `verify-module-structure.sh` exige 10 arquivos obrigatórios por módulo (regra da ADR-0022, agora obsoleta pela ADR-0030)
5. `AGENTS.md` instrui agentes a rodarem `lint.sh`, mas `lint.sh` não cobre verificação de estrutura de módulos

## Motivação

- **Problema:** A camada de defesa local (scripts) está incompleta e desconectada da nova arquitetura simplificada.
- **Custo:** Agentes e humanos não têm feedback local rápido sobre violações arquiteturais.

## Escopo

### Dentro do escopo
- Simplificar `verify-module-structure.sh`: reduzir de 10 para ~4-5 arquivos obrigatórios (doc.go, models.go, repository.go, service.go, README.md)
- Integrar `verify-module-structure.sh` em `pre-commit.sh`
- Integrar `verify-module-structure.sh` em `lint.sh`
- Expandir `verify-contracts.sh` para incluir todos os testes de arquitetura simplificados
- Atualizar `pre-commit.sh` para garantir que rode a suite completa
- Verificar e corrigir `AGENTS.md` se as instruções para agentes estão desatualizadas
- Garantir que `safe-commit.sh` continue funcionando após mudanças

### Fora do escopo
- Implementação dos novos testes de arquitetura (task separada)
- CI/CD workflows (task separada)
- Refatoração de código (task separada)

## Arquivos Relevantes
- `scripts/git/pre-commit.sh`
- `scripts/go/lint.sh`
- `scripts/go/verify-contracts.sh`
- `scripts/go/verify-module-structure.sh`
- `scripts/git/safe-commit.sh`
- `AGENTS.md`

## Resumo

Atualizar scripts de validação local para refletir a arquitetura simplificada (ADR-0030) e garantir que toda a suite de defesa arquitetural rode antes de cada commit.

## Entradas
- Scripts atuais em `scripts/`
- ADR-0030 (novas regras de estrutura de módulo)
- Testes de arquitetura em `tests/architecture/` (existentes + futuros)

## Saídas Esperadas

Scripts atualizados:
- `scripts/git/pre-commit.sh` — roda suite completa
- `scripts/go/lint.sh` — roda suite completa + module structure
- `scripts/go/verify-contracts.sh` — reflete suite completa
- `scripts/go/verify-module-structure.sh` — regras simplificadas
- `AGENTS.md` atualizado com instruções claras

## Critérios de Aceitação
- [ ] `verify-module-structure.sh` exige apenas ~4-5 arquivos por módulo (não 10)
- [ ] `pre-commit.sh` chama `verify-module-structure.sh`
- [ ] `lint.sh` chama `verify-module-structure.sh`
- [ ] `verify-contracts.sh` tem comentário atualizado refletindo suite completa
- [ ] `AGENTS.md` instrui agentes a rodarem `lint.sh` e `verify-contracts.sh` explicitamente
- [ ] Todos os scripts terminam com exit code 0 quando tudo passa
- [ ] Todos os scripts terminam com exit code != 0 quando qualquer verificação falha
- [ ] `safe-commit.sh` funciona sem alterações (ele delega para `pre-commit.sh`)
