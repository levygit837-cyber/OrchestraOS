# CHECKLIST — Architecture Test Suite Hardening

**ID:** ADR-0030-T1-architecture-test-suite-hardening-checklist  
**Referência ao Plano:** docs/agent/tasks/2026-05-21_architecture-test-suite-hardening/plan.md  
**Agente:** Kimi Code CLI  
**Iniciado em:** 2026-05-21T00:30:00-03:00  
**Atualizado em:** 2026-05-21T01:00:00-03:00  
**Status:** em_andamento

---

## Visão Geral

Simplificar os testes de arquitetura para refletir a ADR-0030: de ~10 testes complexos para 4 testes simples que detectam violações reais de comportamento (imports cross-module, business logic em repository, entity types em domain, anomalias de código).

---

## Itens de Execução

### Fase 1: Setup e Análise
- [x] 1.1 Ler relatório de auditoria
- [x] 1.2 Estudar testes atuais em tests/architecture/
- [x] 1.3 Decidir quais testes manter, remover, simplificar
- [x] 1.4 Criar branch feature/2026-05-21_architecture-test-suite-hardening

### Fase 2: Implementar TestModuleBoundaries (simplificado)
- [x] 2.1 Substituir module_boundaries_test.go com versão sem whitelist
- [x] 2.2 Lógica: qualquer import de internal/modules/X em internal/modules/Y (X≠Y) falha
- [x] 2.3 Exceções: orchestrator/ e bootstrap/
- [x] 2.4 Rodar teste e verificar que detecta violações atuais (17 ocorrências, 7 pares únicos)

### Fase 3: Implementar TestRepositoryPurity
- [x] 3.1 Criar repository_purity_test.go
- [x] 3.2 Implementar detecção de if status == Status* em repository.go
- [x] 3.3 Implementar detecção de ON CONFLICT em strings SQL
- [x] 3.4 Implementar detecção de nomes de método não-CRUD
- [x] 3.5 Rodar teste e verificar que detecta violações conhecidas (3 status-branching, 5 non-CRUD methods)

### Fase 4: Implementar TestDomainImportIntegrity
- [x] 4.1 Criar domain_import_integrity_test.go
- [x] 4.2 Definir lista de 26 entity types que DEVEM estar em internal/domain/
- [x] 4.3 Verificar que cada entity type está definido em internal/domain/*.go
- [x] 4.4 Rodar teste (falha como esperado — 26 tipos ainda não migrados)

### Fase 5: Corrigir TestCodeAnomalies
- [x] 5.1 Adicionar detecção de _ = <ident> (variável, não call)
- [x] 5.2 Adicionar detecção de _ = call() dentro de defer func() { ... }()
- [x] 5.3 Expandir regex SQL para SELECT \w+\( (sem FROM)
- [x] 5.4 Rodar teste e verificar que detecta anomalias atuais (19 ignored values + 1 inline SQL)

### Fase 6: Remover Testes Obsoletos
- [x] 6.1 Remover module_contract_test.go
- [x] 6.2 Remover contracts_sync_test.go
- [x] 6.3 Remover queries_purity_test.go
- [x] 6.4 Remover transition_imports_test.go
- [x] 6.5 Remover coordination_removed_test.go
- [x] 6.6 Remover forbidden_filenames_test.go
- [x] 6.7 Remover module_files_test.go
- [x] 6.8 Remover orchestration_imports_test.go
- [x] 6.9 Remover domain_purity_test.go (substituído por domain_import_integrity)

### Fase 7: Integração e Documentação
- [x] 7.1 Garantir que go test ./tests/architecture/... compila sem erros
- [x] 7.2 Atualizar comentários dos testes explicando heurísticas
- [x] 7.3 Criar tests/architecture/README.md documentando cada teste

### Fase 8: Validação Final
- [ ] 8.1 Rodar suite completa: go test ./tests/architecture/... -v
- [ ] 8.2 Verificar que NOVOS testes falham (provando que detectam violações reais)
- [ ] 8.3 Commit na branch com mensagem descritiva
- [ ] 8.4 Push e abertura de PR

---

## Anotações

### 2026-05-21 00:30
**Contexto:** Início da implementação da T1.  
**Decisão/Ação:** Criada branch e checklist. Leitura completa dos 11 testes atuais.  
**Impacto:** Entendimento claro do que manter (code_anomalies), simplificar (module_boundaries), inverter (domain_purity → domain_import_integrity), e remover (todos os outros).

### 2026-05-21 00:40
**Contexto:** Implementando TestModuleBoundaries simplificado.  
**Decisão/Ação:** Substituída whitelist por regra rígida: zero imports cross-module.  
**Impacto:** Teste detecta 17 ocorrências (7 pares únicos: agentsession→agent, run→agentsession/task/workunit, taskgraph→task, trigger→agentsession/run/workunit, workunit→task/taskgraph).

### 2026-05-21 00:45
**Contexto:** Implementando TestRepositoryPurity.  
**Decisão/Ação:** Focado em status-branching e non-CRUD methods. time.Now() e scan* não são flagados (prática comum aceitável).  
**Impacto:** Teste detecta 3 violações de status-branching (agentsession, run) e 5 non-CRUD methods (prompt, taskgraph, eventstore).

### 2026-05-21 00:50
**Contexto:** Implementando TestDomainImportIntegrity.  
**Decisão/Ação:** Lista de 26 tipos compartilhados. Teste verifica presença em internal/domain/.  
**Impacto:** Teste falha como esperado — todos os 26 tipos ainda estão nos módulos (serão migrados na T5).

### 2026-05-21 00:55
**Contexto:** Atualizando TestCodeAnomalies.  
**Decisão/Ação:** Adicionadas 2 novas detecções: ignored variables (não só calls) e ignored errors em defer blocks.  
**Impacto:** Detecta 19 ocorrências de _ = ctx em fetch.go/service.go e 1 inline SQL (pg_advisory_xact_lock).

---

## Bloqueios

Nenhum no momento.

---

## Entrega

**Concluído em:** pendente  
**Resumo:**
- Pendente

**Arquivos Alterados:**
- Pendente

**Status:** ⏳ Em Andamento
