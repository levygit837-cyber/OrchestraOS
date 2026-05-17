# Plano de Atualização e Consolidação da Documentação

Este plano identifica documentações desatualizadas, duplicatas e inconsistências, e define ações para consolidar e atualizar toda a documentação do OrchestraOS para refletir o estado atual da arquitetura de módulos verticais (ADR 0022).

## Problema Principal

A documentação do OrchestraOS contém referências desatualizadas a `internal/services/` (que foi migrado para `internal/modules/` conforme ADR 0022), duplicatas de conteúdo entre documentos de análise, e inconsistências entre a documentação de arquitetura e o código atual implementado.

## Inventário de Documentos Desatualizados

### 1. ADRs que referenciam `internal/services/` (CRÍTICO)

| ADR | Linhas | Problema | Ação |
|-----|--------|----------|------|
| 0017-Domain Services | Tudo | Documento inteiro descreve `internal/services/` que não existe mais | **DEPRECATED** - Criar novo ADR 0024 marcando 0017 como obsoleto |
| 0019-Runtime Service Integration | 9 | Referencia `internal/services/gemini_planner.go` | Atualizar para `internal/modules/taskgraph/gemini_planner.go` |
| 0020-Orchestrator Service | 85 | Referencia `internal/services/orchestrator_service.go` | Atualizar para `internal/modules/orchestrator/service.go` |
| 0021-Agent Service | 63 | Referencia `internal/services/agent_service.go` | Atualizar para `internal/modules/agent/service.go` |
| 0023-Hybrid Intelligent Orchestrator | 54 | Referencia `internal/services/orchestrator_service.go` | Atualizar para `internal/modules/orchestrator/service.go` |

### 2. Documentação de Arquitetura

| Arquivo | Linhas | Problema | Ação |
|--------|--------|----------|------|
| architecture/orchestration.md | 181 | Referencia `internal/services` como fronteira de comando | Atualizar para `internal/modules/*` e `internal/core/coordination` |
| architecture/repo-structure.md | 12, 25 | Descreve estrutura inicial com `internal/services/` | **REESCREVER** para refletir estrutura atual com módulos verticais |
| architecture/module_index.md | 28 | Menciona `common.go` foi eliminado e `internal/services` | Atualizar para refletir estrutura atual de módulos |
| architecture/orchestrator-observation-api.md | 425 | Exemplo de código com caminho antigo | Atualizar exemplo de código |
| architecture/README.md | - | Diagrama não reflete módulos verticais | Atualizar diagrama e descrição |

### 3. Roadmap e Implementação

| Arquivo | Linhas | Problema | Ação |
|--------|--------|----------|------|
| implementation/roadmap.md | 150, 246, 396-398, 406, 502-503 | Múltiplas referências a `internal/services/` | Atualizar todas as referências para `internal/modules/` |
| templates/MODULE_README.md | 61 | Menciona `internal/services (removed)` | Remover ou atualizar contexto |

### 4. Análises (docs/analysis/)

| Arquivo | Problema | Ação |
|--------|----------|------|
| orchestrator-agent-gap-analysis.md | Linhas 396-398, 406, 502-503 referenciam `internal/services/` | Atualizar referências |
| roadmap-reassessment.md | Discussão sobre ADR 0022 como "adiado pós-MVP" mas código JÁ migrou | Atualizar para refletir estado atual |
| sistema-orquestracao-analise-critica.md | Possível duplicata de conteúdo com outros arquivos de análise | **CONSOLIDAR** - verificar sobreposição |
| melhorias-sistema-avaliacao.md | Possível duplicata de conteúdo | **CONSOLIDAR** - verificar sobreposição |
| isolamento-worktree-analise.md | Possível duplicata de conteúdo | **CONSOLIDAR** - verificar sobreposição |
| arquitetura-escalavel-analise-completa.md | Recém-criado, pode conter conteúdo duplicado | **MANTER** como documento consolidado principal |

## Plano de Ação

### Fase 1: Atualizar Referências Críticas (PRIORIDADE ALTA)

**Objetivo:** Corrigir todas as referências a `internal/services/` para `internal/modules/` ou `internal/core/coordination/`.

#### 1.1 ADRs que precisam de atualização

**ADR 0019 - Runtime Service Integration**
- Substituir: `internal/services/gemini_planner.go` → `internal/modules/taskgraph/gemini_planner.go`
- Substituir: `internal/services/task_graph_service.go` → `internal/modules/taskgraph/service.go`
- Substituir: `internal/services/prompt_service.go` → `internal/modules/prompt/service.go`

**ADR 0020 - Orchestrator Service**
- Substituir: `internal/services/orchestrator_service.go` → `internal/modules/orchestrator/service.go`
- Adicionar nota: "Implementado em `internal/modules/orchestrator/` com adapters via bootstrap/services.go"

**ADR 0021 - Agent Service**
- Substituir: `internal/services/agent_service.go` → `internal/modules/agent/service.go`
- Adicionar nota: "Implementado em `internal/modules/agent/`"

**ADR 0023 - Hybrid Intelligent Orchestrator**
- Substituir: `internal/services/orchestrator_service.go` → `internal/modules/orchestrator/service.go`
- Adicionar nota: "OrchestratorService implementado em `internal/modules/orchestrator/`"

**ADR 0017 - Domain Services for Operational Dependencies**
- **AÇÃO ESPECIAL:** Este ADR está completamente obsoleto pois descreve uma camada que não existe mais
- Criar novo ADR 0024: "Deprecation of ADR 0017 - Domain Services Layer"
- Marcar ADR 0017 como DEPRECATED no topo do arquivo
- Adicionar redirecionamento para ADR 0022 (Vertical Slice Architecture)

#### 1.2 Documentação de Arquitetura

**architecture/orchestration.md**
- Linha 181: Substituir `internal/services` → `internal/modules/*` e `internal/core/coordination`
- Adicionar seção explicando a mudança para Vertical Slices

**architecture/repo-structure.md**
- **REESCREVER COMPLETAMENTE** para refletir estrutura atual:
  ```
  cmd/orchestraos/
  internal/
    bootstrap/          # DI e wiring de serviços
    core/               # Componentes compartilhados
      apperrors/
      db/
      event/
      eventstore/
      orchestration/    # Cross-domain helpers
      serialization/
      statemachine/
      transition/
      validation/
    domain/             # Tipos compartilhados
    modules/            # Módulos verticais autônomos
      agent/
      agentsession/
      orchestrator/
      prompt/
      review/
      run/
      task/
      taskgraph/
      trigger/
      workunit/
  contracts/
  migrations/
  tests/
  docs/
  ```

**architecture/module_index.md**
- Atualizar tabela de módulos para refletir os 10 módulos atuais
- Remover referência a `common.go` e `internal/services`
- Adicionar `orchestrator`, `review`, `trigger` à lista

**architecture/orchestrator-observation-api.md**
- Linha 425: Atualizar exemplo de código

**architecture/README.md**
- Atualizar diagrama para mostrar módulos verticais
- Adicionar seção sobre ADR 0022 e migração para Vertical Slices

#### 1.3 Roadmap e Implementação

**implementation/roadmap.md**
- Linha 150: `internal/services/runtime_relay.go` → `internal/core/coordination/runtime_event_relay.go`
- Linha 246: `internal/services/orchestrator_service.go` → `internal/modules/orchestrator/service.go`
- Linhas 396-398, 406: Atualizar referências a services
- Linhas 502-503: Atualizar referências a services
- Linha 562: Atualizar nota sobre ADR 0022 - já está implementado, não é mais "adiado pós-MVP"

**templates/MODULE_README.md**
- Linha 61: Remover ou atualizar menção a `internal/services`

#### 1.4 Análises

**docs/analysis/orchestrator-agent-gap-analysis.md**
- Linhas 396-398, 406, 502-503: Atualizar referências a `internal/services/`

**docs/analysis/roadmap-reassessment.md**
- Atualizar seção sobre ADR 0022: já está implementado, não é mais "adiado pós-MVP"
- Adicionar nota: "Migração para Vertical Slices concluída conforme ADR 0022"

### Fase 2: Consolidar Documentos de Análise (PRIORIDADE MÉDIA)

**Objetivo:** Eliminar duplicatas e consolidar conteúdo sobreposto em docs/analysis/.

#### 2.1 Análise de Sobreposição

Ler e comparar os seguintes arquivos para identificar conteúdo duplicado:
- docs/analysis/sistema-orquestracao-analise-critica.md
- docs/analysis/melhorias-sistema-avaliacao.md
- docs/analysis/isolamento-worktree-analise.md

#### 2.2 Ações de Consolidação

**Se houver sobreposição significativa:**
- Manter o arquivo mais completo e atualizado
- Adicionar notas nos outros arquivos redirecionando para o principal
- Arquivar ou marcar como "superseded by [arquivo principal]"

**Se houver conteúdo único em cada:**
- Renomear arquivos para serem mais específicos sobre seu escopo
- Adicionar introduções claras explicando o foco de cada análise

### Fase 3: Criar Documento de Migração (PRIORIDADE MÉDIA)

**Objetivo:** Documentar a migração de `internal/services/` para `internal/modules/` para contexto histórico.

**Criar:** docs/architecture/migration-vertical-slices.md

Conteúdo:
- Contexto: Por que a migração foi necessária (ADR 0022)
- O que mudou: De layered architecture para vertical slice architecture
- Quando ocorreu: Data aproximada da migração
- Como funciona agora: Padrão de módulos verticais, comunicação via core/coordination
- Benefícios: Isolamento de contexto para LLMs, escalabilidade, testabilidade
- Referências: ADR 0022, docs/development/CODING_STANDARDS.md

### Fase 4: Atualizar Índices e Referências Cruzadas (PRIORIDADE BAIXA)

**Objetivo:** Garantir que todos os índices e referências cruzadas estejam atualizados.

#### 4.1 docs/architecture/README.md
- Atualizar lista de documentos de arquitetura
- Adicionar link para migration-vertical-slices.md
- Remover referências a documentos obsoletos

#### 4.2 docs/README.md (se existir)
- Criar ou atualizar índice principal da documentação
- Organizar por categoria: ADRs, Arquitetura, Análises, Implementação

#### 4.3 docs/adr/README.md (se existir)
- Criar índice de ADRs com status
- Marcar ADR 0017 como DEPRECATED
- Adicionar ADR 0024 como referência

### Fase 5: Validação Final

**Objetivo:** Garantir que não restem referências desatualizadas.

#### 5.1 Busca Global
```bash
# Buscar por referências a internal/services
grep -r "internal/services" docs/

# Buscar por referências a common.go
grep -r "common.go" docs/

# Buscar por referências a domain services layer
grep -r "domain services" docs/ -i
```

#### 5.2 Validação de Links
- Verificar que todos os links internos ainda funcionam
- Atualizar links quebrados

#### 5.3 Consistência de Terminologia
- Garantir uso consistente de "módulo vertical", "vertical slice", "domain module"
- Garantir uso consistente de "core/coordination" para helpers cross-domain

## Ordem de Execução

1. **Fase 1.1:** Atualizar ADRs 0019, 0020, 0021, 0023
2. **Fase 1.1 (especial):** Criar ADR 0024 e marcar 0017 como DEPRECATED
3. **Fase 1.2:** Atualizar docs/architecture/orchestration.md, repo-structure.md, module_index.md
4. **Fase 1.3:** Atualizar docs/implementation/roadmap.md
5. **Fase 1.4:** Atualizar docs/analysis/orchestrator-agent-gap-analysis.md, roadmap-reassessment.md
6. **Fase 3:** Criar docs/architecture/migration-vertical-slices.md
7. **Fase 2:** Consolidar documentos de análise (após leitura e comparação)
8. **Fase 4:** Atualizar índices e referências cruzadas
9. **Fase 5:** Validação final com busca global

## Critérios de Aceite

- [ ] Nenhuma referência a `internal/services/` permanece em docs/
- [ ] ADR 0017 está marcado como DEPRECATED com redirecionamento para ADR 0022
- [ ] ADR 0024 criado explicando depreciação de ADR 0017
- [ ] docs/architecture/repo-structure.md reflete estrutura atual com módulos verticais
- [ ] docs/architecture/module_index.md lista todos os 10 módulos atuais
- [ ] docs/implementation/roadmap.md não contém referências a internal/services
- [ ] Documentos de análise consolidados (sem duplicatas significativas)
- [ ] docs/architecture/migration-vertical-slices.md criado
- [ ] Busca global por "internal/services" retorna zero resultados em docs/
- [ ] Todos os links internos funcionam
- [ ] Terminologia consistente em toda a documentação

## Riscos e Mitigações

**Risco:** Perda de contexto histórico ao atualizar documentos
- **Mitigação:** Criar documento de migração (Fase 3) que preserva histórico

**Risco:** Quebra de links externos que referenciam documentos atualizados
- **Mitigação:** Usar redirects ou manter aliases se necessário

**Risco:** Consolidação de análises pode remover conteúdo útil
- **Mitigação:** Ler completamente antes de consolidar; manter conteúdo único; arquivar ao invés de deletar

## Tempo Estimado

- Fase 1: 2-3 horas (atualizações diretas de texto)
- Fase 2: 1-2 horas (leitura e consolidação)
- Fase 3: 1 hora (criação de documento)
- Fase 4: 1 hora (atualização de índices)
- Fase 5: 0.5 hora (validação)

**Total:** 5.5-7.5 horas
