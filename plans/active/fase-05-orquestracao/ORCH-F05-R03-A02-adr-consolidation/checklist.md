# Checklist — Consolidação de ADRs

**Plano:** ORCH-F05-R03-A02-adr-consolidation  
**Agente:** Documentação / Arquitetura  
**Status:** Em execução

---

## Fase 1: Preparação e Verificação

- [ ] 1.1. Ler e confirmar que TODOS os 28 arquivos em `docs/adr/` foram analisados
- [ ] 1.2. Executar `grep -rn "docs/adr/" docs/ --include="*.md"` para mapear todas as referências cruzadas
- [ ] 1.3. Executar `grep -rn "docs/adr/" AGENTS.md` para verificar links
- [ ] 1.4. Executar `grep -rn "ADR-[0-9]" plans/active/ --include="*.md"` para mapear referências em plans
- [ ] 1.5. Criar branch: `git checkout -b adr-consolidation`
- [ ] 1.6. Confirmar que `./scripts/safe-commit.sh` está funcionando

---

## Fase 2: Consolidação dos Grupos (merge de conteúdo)

### Grupo A: Arquitetura de Módulos (0022 + 0024 + 0025 + 0026 + 0027a + 0027b)

- [ ] 2A.1. Criar `docs/adr/0022-vertical-module-architecture.md` com estrutura do template
- [ ] 2A.2. Copiar conteúdo completo do ADR 0022 (seção 1: Decisão Principal)
- [ ] 2A.3. Adicionar apêndice de evolução: mapeamento 0017→módulos (conteúdo do 0024)
- [ ] 2A.4. Adicionar seção de padronização: template contract.go, 10 arquivos obrigatórios (0025)
- [ ] 2A.5. Adicionar seção de política de importação: 3 pilares, exemplos Go válidos/inválidos (0026)
- [ ] 2A.6. Adicionar seção de renomeação: tabela de diretórios, alternativas, veredictos (0027a)
- [ ] 2A.7. Adicionar Apêndice A: Histórico de Evolução com datas
- [ ] 2A.8. Adicionar Apêndice B: Alternativas Consideradas (merge sem duplicação)
- [ ] 2A.9. Verificar: checklist de implementação combinado está presente

### Grupo B: Observabilidade e Memória (0009 + 0012)

- [ ] 2B.1. Renomear `git mv docs/adr/0009-trace-history-normalization.md docs/adr/0009-observability-and-memory.md`
- [ ] 2B.2. Manter seção 1 (Event Store e Tracing) do 0009 original
- [ ] 2B.3. Adicionar seção 2: Memória Recursiva (conteúdo completo do 0012)
- [ ] 2B.4. Manter tabela comparativa Event Store vs Recursive Memory
- [ ] 2B.5. Manter regra de fronteira: "Event Store = o que aconteceu..."
- [ ] 2B.6. Manter fluxo alvo de 11 passos de ingestão/recuperação
- [ ] 2B.7. Adicionar Apêndice A: Histórico de Evolução
- [ ] 2B.8. Adicionar Apêndice B: Alternativas Consideradas (merge)

### Grupo C: Ciclo Operacional do Agente (0007 + 0008 + 0011)

- [ ] 2C.1. Criar `docs/adr/0007-agent-operational-cycle.md`
- [ ] 2C.2. Copiar conteúdo completo do ADR 0007 (seção 1: Prompts)
- [ ] 2C.3. Adicionar seção 2: Ledger (conteúdo completo do 0008)
- [ ] 2C.4. Adicionar seção 3: Checkpoints (conteúdo completo do 0011)
- [ ] 2C.5. Manter distinção ledger vs checkpoint
- [ ] 2C.6. Manter seção "Fora Desta Decisão" do 0011
- [ ] 2C.7. Adicionar Apêndice A: Histórico de Evolução
- [ ] 2C.8. Adicionar Apêndice B: Alternativas Consideradas (merge)

### Grupo D: Serviços de Orquestração (0020 + 0021)

- [ ] 2D.1. Renomear `git mv docs/adr/0020-orchestrator-service.md docs/adr/0020-orchestration-services.md`
- [ ] 2D.2. Manter seção 1: OrchestratorService (conteúdo do 0020)
- [ ] 2D.3. Adicionar seção 2: AgentService (conteúdo do 0021)
- [ ] 2D.4. Manter perfis de agente e runtime types
- [ ] 2D.5. Adicionar Apêndice A: Histórico de Evolução
- [ ] 2D.6. Adicionar Apêndice B: Alternativas Consideradas (merge)

### Grupo E: Fundação M0 (0013 + 0014)

- [ ] 2E.1. Renomear `git mv docs/adr/0013-m0-domain-contract-scope.md docs/adr/0013-m0-foundation.md`
- [ ] 2E.2. Manter seção 1: Escopo de Entidades (conteúdo do 0013)
- [ ] 2E.3. Adicionar seção 2: Persistência, CLI e Testes (conteúdo do 0014)
- [ ] 2E.4. Manter riscos conhecidos e atualizações de implementação
- [ ] 2E.5. Adicionar Apêndice A: Histórico de Evolução
- [ ] 2E.6. Adicionar Apêndice B: Alternativas Consideradas (merge)

### Grupo F: Interface do MVP (0005 + 0015)

- [ ] 2F.1. Manter `docs/adr/0005-mvp-interface-strategy.md` como base
- [ ] 2F.2. Adicionar seção 2: Evolução para TUI (conteúdo do 0015)
- [ ] 2F.3. Manter framework Bubble Tea + 5 motivos
- [ ] 2F.4. Manter 5 critérios de prototipagem TUI
- [ ] 2F.5. Adicionar Apêndice A: Histórico de Evolução
- [ ] 2F.6. Adicionar Apêndice B: Alternativas Consideradas (merge)

### Grupo G: Novo artefato de implementação (0019)

- [ ] 2G.1. Criar `docs/implementation/runtime-integration.md`
- [ ] 2G.2. Copiar conteúdo completo do ADR 0019
- [ ] 2G.3. Adaptar formato: remover estrutura ADR, usar documentação técnica direta
- [ ] 2G.4. Adicionar nota no topo: "Migrado de ADR 0019 em {data}"

---

## Fase 3: Remoção de Arquivos Legados

- [ ] 3.1. `git rm docs/adr/0007-prompt-composition-system.md`
- [ ] 3.2. `git rm docs/adr/0008-agent-task-ledger.md`
- [ ] 3.3. `git rm docs/adr/0011-agent-checkpoints.md`
- [ ] 3.4. `git rm docs/adr/0012-recursive-memory-system.md`
- [ ] 3.5. `git rm docs/adr/0014-m0-cli-persistence-and-integration-tests.md`
- [ ] 3.6. `git rm docs/adr/0015-tui-as-primary-local-interface.md`
- [ ] 3.7. `git rm docs/adr/0021-agent-service.md`
- [ ] 3.8. `git rm docs/adr/0024-deprecation-of-adr-0017.md`
- [ ] 3.9. `git rm docs/adr/0025-module-standardization.md`
- [ ] 3.10. `git rm docs/adr/0026-module-import-policy.md`
- [ ] 3.11. `git rm docs/adr/0027-directory-semantic-renaming.md`
- [ ] 3.12. `git rm docs/adr/0027-orchestrator-module-naming.md`
- [ ] 3.13. Commit da remoção: `./scripts/safe-commit.sh "docs: remove ADRs consolidados em ORCH-F05-R03-A02"`

---

## Fase 4: Atualização de Referências Cruzadas

### Em ADRs que permanecem

- [ ] 4.1. Atualizar `docs/adr/0006-task-graph-and-agent-intervention.md` se referenciar 0007/0008/0011
- [ ] 4.2. Atualizar `docs/adr/0016-event-sourced-state-machine.md` se referenciar 0009/0012
- [ ] 4.3. Atualizar `docs/adr/0017-domain-services-for-operational-dependencies.md` → já deprecated, verificar se precisa atualizar link para 0022
- [ ] 4.4. Atualizar `docs/adr/0018-local-heuristic-task-graph-planner.md` se referenciar outros ADRs
- [ ] 4.5. Atualizar `docs/adr/0019-runtime-service-integration.md` → será removido, mas antes verificar se outros ADRs o referenciam
- [ ] 4.6. Atualizar `docs/adr/0020-orchestration-services.md` se referenciar 0021
- [ ] 4.7. Atualizar `docs/adr/0022-vertical-module-architecture.md` se referenciar 0024/0025/0026/0027
- [ ] 4.8. Atualizar `docs/adr/0023-hybrid-intelligent-orchestrator.md` se referenciar 0020/0021/0022

### Em documentos de arquitetura

- [ ] 4.9. `docs/architecture/orchestration.md`
- [ ] 4.10. `docs/architecture/communication-protocol.md`
- [ ] 4.11. `docs/architecture/intelligent-orchestrator-agent.md`
- [ ] 4.12. `docs/architecture/orchestrator-observation-api.md`
- [ ] 4.13. `docs/architecture/orchestrator-intervention-protocol.md`
- [ ] 4.14. `docs/architecture/multi-agent-coordination.md`
- [ ] 4.15. `docs/architecture/migration-vertical-slices.md`

### Em AGENTS.md e plans/

- [ ] 4.16. Atualizar `AGENTS.md` se houver links para ADRs removidos
- [ ] 4.17. Atualizar `plans/active/` que referenciam ADRs removidos
- [ ] 4.18. Commit das atualizações: `./scripts/safe-commit.sh "docs: update cross-references after ADR consolidation"`

---

## Fase 5: Validação Final

- [ ] 5.1. Contar arquivos em `docs/adr/`: devem ser exatamente 14
- [ ] 5.2. Listar os 14 arquivos e confirmar que todos têm `Status:` no cabeçalho
- [ ] 5.3. Verificar que nenhum arquivo legado existe em `docs/adr/`
- [ ] 5.4. Rodar `grep -r "docs/adr/" docs/ --include="*.md" | grep -v "Consolidated"` e confirmar zero links quebrados
- [ ] 5.5. Verificar que `docs/implementation/runtime-integration.md` existe
- [ ] 5.6. Verificar que `docs/adr/0017-domain-services-for-operational-dependencies.md` permanece como deprecated com link para 0022
- [ ] 5.7. Revisar manualmente cada ADR consolidado: abrir e confirmar que as seções esperadas existem
- [ ] 5.8. Revisar que todos os "Alternativas Consideradas" estão presentes
- [ ] 5.9. Revisar que todos os "Consequências" estão presentes
- [ ] 5.10. `./scripts/verify-contracts.sh` (se aplicável)
- [ ] 5.11. `./scripts/lint.sh` (se aplicável)
- [ ] 5.12. Commit final: `./scripts/safe-commit.sh "docs: consolidate ADRs — 27 → 14, zero content loss"`

---

## Fase 6: Pós-Implementação

- [ ] 6.1. Mover este plano para `plans/archive/fase-05-orquestracao/ORCH-F05-R03-A02-adr-consolidation/`
- [ ] 6.2. Renomear `checklist.md` para `checklist-completed.md`
- [ ] 6.3. Atualizar `plans/README.md` se houver seção de fases ativas
- [ ] 6.4. Abrir PR para revisão humana
- [ ] 6.5. Aguardar aprovação antes de merge

---

## Resumo Esperado (pós-conclusão)

| Métrica | Antes | Depois |
|---------|-------|--------|
| ADRs em `docs/adr/` | 28 (27 únicos + 1 duplicado) | 14 |
| Duplicados | 1 (0027a + 0027b) | 0 |
| ADRs deprecados isolados | 1 (0017) | 1 (permanece como marcador) |
| Documentos de implementação | 0 | 1 (`runtime-integration.md`) |
| Perda de conteúdo | — | 0% |
