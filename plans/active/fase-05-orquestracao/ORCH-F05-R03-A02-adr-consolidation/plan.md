# Plano de Consolidação de ADRs — OrchestraOS

**ID:** ORCH-F05-R03-A02-adr-consolidation  
**Fase:** 5 — Orquestração & Arquitetura  
**Tipo:** Documentação / Refatoração de ADRs  
**Risco:** Baixo (nenhuma mudança de código)  
**Autonomia Aprovada:** Nível 2 — IA implementa com revisão humana

---

## 1. Contexto

O projeto possui 28 arquivos de ADR (27 únicos + 1 duplicado). Análise semântica identificou que vários ADRs tratam do mesmo tema arquitetural, foram escritos como evoluções sequenciais da mesma decisão, ou são especificações de implementação disfarçadas de decisões arquiteturais.

Este plano define como consolidar os ADRs de **27 → 14 documentos**, preservando **100% do conteúdo factual**, removendo apenas duplicação e redirecionamento.

---

## 2. Princípios da Consolidação

1. **Zero perda de conteúdo**: Todo fato, regra, consequência e alternativa considerada de um ADR original deve estar presente no consolidado.
2. **Linha evolutiva preservada**: Datas, contexto de mudança e motivos de evolução são mantidos como seções ou apêndices.
3. **Referências atualizadas**: ADRs que referenciam números antigos devem ter seus links atualizados.
4. **Git history mantida**: Usar `git mv` quando possível; arquivos removidos são commitados como tal.
5. **Status claro no cabeçalho**: ADRs consolidados recebem tag `Consolidated` com lista de ADRs absorvidos.

---

## 3. Mapeamento de Merge — Grupos e Destinos

### Grupo A: Arquitetura de Módulos (5 → 1)

| ADR Original | Conteúdo Chave | Destino no Consolidado |
|--------------|---------------|------------------------|
| **0022** — LLM-Optimized Module Architecture | Decisão de migrar de layered para vertical slices; motivos de eficiência para LLMs | Seção 1: Decisão Principal |
| **0024** — Deprecation of ADR 0017 | Mapeamento de `internal/services/` → `internal/modules/*` | Seção 2: Evolução da Decisão (apêndice de 5 linhas) |
| **0025** — Module Standardization | Template de `contract.go`, arquivos obrigatórios, regras hierárquicas | Seção 3: Padronização de Estrutura |
| **0026** — Module Import Policy | 3 pilares (domain=infra, imports DI permitidos, exceções orchestrator/coordination) | Seção 4: Política de Importação |
| **0027a** — Directory Semantic Renaming | `orchestrator/` → `taskflow/`, `event/` → `eventappend/` | Seção 5: Renomeação Semântica |
| **0027b** — Orchestrator Module Naming | Mesmo tema de 0027a, status Accepted | **Absorvido integralmente em 0027a** |

**Arquivo destino:** `docs/adr/0022-vertical-module-architecture.md`
**Novo título:** "0022. Arquitetura de Módulos Verticais — Decisão, Padronização e Políticas"
**Status:** Consolidated (absorve 0024, 0025, 0026, 0027a, 0027b)

**Conteúdo que NÃO pode faltar:**
- Tabela comparativa de arquiteturas (0022, seção 4)
- Mapeamento completo de serviços para módulos (0024)
- Template unificado de `contract.go` com 3 níveis de regras (0025)
- Matriz de importação `internal/core/*` (0025)
- 3 pilares de importação com exemplos válidos/inválidos em Go (0026)
- Tabela de renomeação de diretórios com alternativas e veredictos (0027a)
- Checklist de implementação combinado (merge de todos os checklists)

---

### Grupo B: Observabilidade e Memória (2 → 1)

| ADR Original | Conteúdo Chave | Destino no Consolidado |
|--------------|---------------|------------------------|
| **0009** — Event Store e Tracing | Event Store como fonte canônica; normalização por IDs; fronteira com memória | Seção 1: Event Store e Tracing |
| **0012** — Recursive Memory | Sistema de memória derivada; classificação, deduplicação, ingestão, retrieval | Seção 2: Memória Recursiva (Derivada) |

**Arquivo destino:** `docs/adr/0009-observability-and-memory.md`
**Novo título:** "0009. Observabilidade: Event Store, Tracing e Memória Recursiva"
**Status:** Consolidated (absorve 0012)

**Conteúdo que NÃO pode faltar:**
- Tabela de correlação de eventos (task_id, run_id, agent_id, etc.) (0009)
- Regra de fronteira: "Event Store = o que aconteceu; Tracing = como evoluiu; Memória = o que lembrar" (0009)
- Tabela comparativa Event Store vs Recursive Memory (0012)
- 11 princípios da memória recursiva (0012)
- Fluxo alvo de 11 passos (Memory Intake → Retriever → Prompt Composer) (0012)
- Escopo inicial, primeiro e segundo corte futuro (0012)
- Toda a seção "Alternativas Consideradas" de ambos

---

### Grupo C: Ciclo Operacional do Agente (3 → 1)

| ADR Original | Conteúdo Chave | Destino no Consolidado |
|--------------|---------------|------------------------|
| **0007** — Prompt Composition System | SystemPrompt + TaskPrompt; fragmentos versionados; toolset por run; reconfiguração | Seção 1: Composição de Prompts |
| **0008** — Agent Task Ledger | Estrutura do ledger; checkpoint vs ledger; não substitui ADR/issue/PR | Seção 2: Ledger de Progresso |
| **0011** — Agent Checkpoints | Pontos seguros de progresso; política de emissão; `AgentSessionService.Checkpoint()` | Seção 3: Checkpoints |

**Arquivo destino:** `docs/adr/0007-agent-operational-cycle.md`
**Novo título:** "0007. Ciclo Operacional do Agente — Prompts, Ledger e Checkpoints"
**Status:** Consolidated (absorve 0008, 0011)

**Conteúdo que NÃO pode faltar:**
- Estrutura dos 2 artefatos principais (SystemPrompt, TaskPrompt) (0007)
- 10 tipos de fragmentos de prompt (0007)
- Regras de especialização dinâmica (4 regras) (0007)
- Regras de toolset por run (4 objetivos) (0007)
- Regras de reconfiguração de AgentSession (7 regras) (0007)
- 9 campos do ledger (0008)
- Distinção ledger (memória viva) vs checkpoint (snapshot) (0008)
- 10 campos de checkpoint (0011)
- 8 momentos naturais de emissão + 9 pontos seguros de checkpoint automático (0011)
- Toda seção "Fora Desta Decisão" (0011)
- "Alternativas Consideradas" de todos os três

---

### Grupo D: Serviços de Orquestração (2 → 1)

| ADR Original | Conteúdo Chave | Destino no Consolidado |
|--------------|---------------|------------------------|
| **0020** — Orchestrator Service | `OrchestratorService.RunTask()`; decisões táticas vs estratégicas; paralelismo | Seção 1: OrchestratorService |
| **0021** — Agent Service | `AgentService.Create/GetByID/FindOrCreate`; perfis válidos; runtime types | Seção 2: AgentService |

**Arquivo destino:** `docs/adr/0020-orchestration-services.md`
**Novo título:** "0020. Serviços de Orquestração — OrchestratorService e AgentService"
**Status:** Consolidated (absorve 0021)

**Conteúdo que NÃO pode faltar:**
- 7 responsabilidades do OrchestratorService (0020)
- Fluxo `RunTask()` em 6 passos (0020)
- 5 decisões táticas (Go) + 4 decisões estratégicas futuras (LLM) (0020)
- Regra de paralelismo: sequencial, limite 1 (0020)
- 3 métodos do AgentService (Create, GetByID, FindOrCreate) (0021)
- 5 perfis de agente válidos (0021)
- 4 runtime types (0021)
- Regra de validação de `AgentID` em `AgentSessionService.Create()` (0021)

**Nota:** ADR 0019 (Runtime Service Integration) vira documento de implementação em `docs/implementation/runtime-integration.md`, não é mergeado aqui.

---

### Grupo E: Fundação M0 (2 → 1)

| ADR Original | Conteúdo Chave | Destino no Consolidado |
|--------------|---------------|------------------------|
| **0013** — M0 Domain Contract Scope | 6 entidades com schema; 3 entidades adiadas (Orchestrator, CommunicationProtocol, Session) | Seção 1: Escopo de Entidades M0 |
| **0014** — M0 CLI Persistence & Integration Tests | `cmd/orchestraos`; `internal/repository`; `internal/eventstore`; migrations goose; FakeRuntime | Seção 2: Persistência, CLI e Testes |

**Arquivo destino:** `docs/adr/0013-m0-foundation.md`
**Novo título:** "0013. Fundação Técnica M0 — Schemas, Persistência e Testes"
**Status:** Consolidated (absorve 0014)

**Conteúdo que NÃO pode faltar:**
- Lista de 6 entidades com schema + 3 adiadas com justificativa (0013)
- 7 componentes introduzidos (0014)
- 5 riscos conhecidos (0014)
- 5 atualizações de implementação (0014)

---

### Grupo F: Interface do MVP (2 → 1)

| ADR Original | Conteúdo Chave | Destino no Consolidado |
|--------------|---------------|------------------------|
| **0005** — MVP Interface Strategy | Scripts → CLI → GitHub; Desktop/Web adiados | Seção 1: Interface Inicial |
| **0015** — TUI as Primary Local Interface | Bubble Tea; CLI permanece como headless; 5 critérios de prototipagem | Seção 2: Evolução para TUI |

**Arquivo destino:** `docs/adr/0005-interface-strategy.md`
**Novo título:** "0005. Estratégia de Interface — CLI, TUI e GitHub-First"
**Status:** Consolidated (absorve 0015)

**Conteúdo que NÃO pode faltar:**
- 3 camadas de progressão (scripts, CLI, GitHub) (0005)
- 6 consequências da decisão (0005)
- Framework Bubble Tea + 5 motivos (0015)
- 5 critérios do protótipo TUI (0015)
- Nota de que CLI não é removida, apenas deixa de ser interface humana principal (0015)

---

### Grupo G: ADRs que permanecem isolados

| # | Arquivo | Motivo |
|---|---------|--------|
| 0001 | `0001-repository-source-of-truth.md` | Decisão raiz do projeto |
| 0002 | `0002-orchestrator-control-plane.md` | Conceito arquitetural central |
| 0003 | `0003-technology-stack.md` | Decisão estratégica de longo prazo |
| 0004 | `0004-sandbox-autonomy.md` | Política de segurança independente |
| 0006 | `0006-task-graph-and-agent-intervention.md` | Conceito de domínio independente |
| 0010 | `0010-github-first-operations.md` | *Ver nota abaixo* |
| 0016 | `0016-event-sourced-state-machine.md` | Padrão técnico profundo |
| 0018 | `0018-heuristic-task-planner.md` | Decisão algorítmica específica |
| 0023 | `0023-hybrid-intelligent-orchestrator.md` | Macro-arquitetura híbrida |

**Nota sobre 0010:** ADR 0010 é semanticamente um reforço de 0005 e 0001. Como é curto (39 linhas) e já está bem delimitado, pode permanecer isolado OU ser absorvido no grupo F (0005). Decisão: **permanece isolado** por ser a formalização do "GitHub-first" como política operacional distinta da estratégia de interface.

---

## 4. Destino dos Arquivos Legados

### Arquivos a serem removidos (conteúdo absorvido)

```
docs/adr/0007-prompt-composition-system.md         → absorvido em 0007-agent-operational-cycle.md
docs/adr/0008-agent-task-ledger.md                  → absorvido em 0007-agent-operational-cycle.md
docs/adr/0009-trace-history-normalization.md        → renomeado para 0009-observability-and-memory.md
docs/adr/0011-agent-checkpoints.md                  → absorvido em 0007-agent-operational-cycle.md
docs/adr/0012-recursive-memory-system.md           → absorvido em 0009-observability-and-memory.md
docs/adr/0014-m0-cli-persistence-and-integration-tests.md → absorvido em 0013-m0-foundation.md
docs/adr/0015-tui-as-primary-local-interface.md     → absorvido em 0005-interface-strategy.md
docs/adr/0021-agent-service.md                      → absorvido em 0020-orchestration-services.md
docs/adr/0024-deprecation-of-adr-0017.md            → absorvido em 0022-vertical-module-architecture.md
docs/adr/0025-module-standardization.md             → absorvido em 0022-vertical-module-architecture.md
docs/adr/0026-module-import-policy.md               → absorvido em 0022-vertical-module-architecture.md
docs/adr/0027-directory-semantic-renaming.md       → absorvido em 0022-vertical-module-architecture.md
docs/adr/0027-orchestrator-module-naming.md       → absorvido em 0022-vertical-module-architecture.md
```

### Arquivos a serem renomeados (`git mv`)

```
docs/adr/0009-trace-history-normalization.md → docs/adr/0009-observability-and-memory.md
docs/adr/0013-m0-domain-contract-scope.md    → docs/adr/0013-m0-foundation.md
docs/adr/0020-orchestrator-service.md        → docs/adr/0020-orchestration-services.md
docs/adr/0022-llm-optimized-module-architecture.md → docs/adr/0022-vertical-module-architecture.md
```

### Novo artefato (documentação de implementação)

```
docs/implementation/runtime-integration.md   ← conteúdo do ADR 0019
```

---

## 5. Estrutura do ADR Consolidado — Template

Todo ADR consolidado deve seguir esta estrutura:

```markdown
# {NNNN}. {Título Consolidado}

**Status:** Consolidated (absorve: ADR XXXX, ADR YYYY, ADR ZZZZ)  
**Data original:** {data do ADR mais antigo}  
**Última atualização:** {data do merge}  

---

## Contexto (integrado)

[Mesclar os contextos dos ADRs originais, mantendo a linha do tempo]

---

## 1. {Tema Principal do ADR-base}

[Conteúdo do ADR original principal]

---

## 2. {Tema do ADR absorvido 1}

### 2.1 Contexto adicional
[Por que este tema surgiu como decisão separada]

### 2.2 Decisão

### 2.3 Consequências

---

## 3. {Tema do ADR absorvido 2}

[Mesma estrutura]

---

## Apêndice A: Histórico de Evolução

| Data | Evento | ADR Original |
|------|--------|--------------|
| YYYY-MM-DD | Decisão inicial | XXXX |
| YYYY-MM-DD | Refinamento de regra | YYYY |

---

## Apêndice B: Alternativas Consideradas (consolidado)

[Merge de todas as alternativas dos ADRs originais, sem duplicação]
```

---

## 6. Referências Cruzadas a Atualizar

Após a consolidação, os seguintes arquivos precisam ter seus links atualizados:

### Em ADRs que permanecem (não consolidados)

| Referência Atual | Novo Destino |
|------------------|--------------|
| `docs/adr/0007-prompt-composition-system.md` | `docs/adr/0007-agent-operational-cycle.md` |
| `docs/adr/0009-trace-history-normalization.md` | `docs/adr/0009-observability-and-memory.md` |
| `docs/adr/0011-agent-checkpoints.md` | `docs/adr/0007-agent-operational-cycle.md#3-checkpoints` |
| `docs/adr/0012-recursive-memory-system.md` | `docs/adr/0009-observability-and-memory.md#2-memoria-recursiva` |
| `docs/adr/0014-m0-cli-persistence-and-integration-tests.md` | `docs/adr/0013-m0-foundation.md#2-persistencia-cli-e-testes` |
| `docs/adr/0015-tui-as-primary-local-interface.md` | `docs/adr/0005-interface-strategy.md#2-evolucao-para-tui` |
| `docs/adr/0021-agent-service.md` | `docs/adr/0020-orchestration-services.md#2-agentservice` |
| `docs/adr/0024-deprecation-of-adr-0017.md` | `docs/adr/0022-vertical-module-architecture.md` |
| `docs/adr/0025-module-standardization.md` | `docs/adr/0022-vertical-module-architecture.md#3-padronizacao` |
| `docs/adr/0026-module-import-policy.md` | `docs/adr/0022-vertical-module-architecture.md#4-politica-de-importacao` |
| `docs/adr/0027-directory-semantic-renaming.md` | `docs/adr/0022-vertical-module-architecture.md#5-renomeacao-semantica` |
| `docs/adr/0027-orchestrator-module-naming.md` | `docs/adr/0022-vertical-module-architecture.md#5-renomeacao-semantica` |

### Em documentos de arquitetura

- `docs/architecture/orchestration.md`
- `docs/architecture/communication-protocol.md`
- `docs/architecture/intelligent-orchestrator-agent.md`
- `docs/architecture/orchestrator-observation-api.md`
- `docs/architecture/orchestrator-intervention-protocol.md`
- `docs/architecture/multi-agent-coordination.md`
- `docs/architecture/migration-vertical-slices.md`

### Em AGENTS.md

O cabeçalho de `AGENTS.md` referencia ADRs. Deve ser verificado se há links diretos.

### Em plans/

Todos os planos em `plans/active/` que mencionam números de ADR devem ser auditados.

---

## 7. Critérios de Aceite

- [ ] 14 ADRs finais existem em `docs/adr/` com conteúdo completo
- [ ] 13 arquivos legados foram removidos via commit (não apenas deletados do filesystem)
- [ ] 4 arquivos renomeados preservaram history via `git mv`
- [ ] Todo fato, regra, consequência e alternativa dos ADRs originais está presente em algum consolidado
- [ ] Nenhum link quebrado em `docs/adr/` (verificado com `grep -r "docs/adr/" docs/`)
- [ ] `AGENTS.md` não referencia ADRs removidos
- [ ] `plans/active/` atualizado se necessário
- [ ] `docs/implementation/runtime-integration.md` criado com conteúdo do ADR 0019
- [ ] Commits via `./scripts/safe-commit.sh`
