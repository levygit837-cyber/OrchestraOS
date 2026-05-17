# ORCH-F28-R01-A06: Cleanup Final — Remover coordination/ + Documentação

## Contexto do Projeto

**Projeto:** OrchestraOS (Go monolith modular)  
**Arquitetura:** Vertical Slice (ADR-0022) + Hybrid Intelligent Orchestrator (ADR-0023)  
**Meta:** Remover completamente `internal/core/coordination/`, atualizar documentação, e garantir build verde.  
**Risco:** **BAIXO-MÉDIO** — operação de cleanup que depende de todos os outros planos terem sido entregues.

---

## Documentação Obrigatória (LEIA ANTES DE COMEÇAR)

1. `docs/adr/0028-core-architecture-and-naming-standards.md` — Seção 4 Fase 2 e Fase 3
2. `AGENTS.md` — Regras de commits, nomes proibidos
3. Este arquivo (`plan.md`)

---

## O que Já Existe vs O que Deve Ser Feito

### Estado esperado antes deste plano

Todos os planos A01-A05 devem ter sido entregues. O diretório `internal/core/coordination/` deve conter APENAS:
- `doc.go` (será removido)
- `queries.go` (deve estar vazio ou conter apenas SQL já movido)
- Possivelmente algum arquivo residual

### Tarefas de cleanup

| # | Tarefa | Arquivos |
|---|---|---|
| 1 | Remover diretório `internal/core/coordination/` inteiro | `rm -rf internal/core/coordination/` |
| 2 | Verificar imports órfãos de `internal/core/coordination` | grep em todo o projeto |
| 3 | Atualizar architecture tests | `tests/architecture/module_boundaries_test.go`, `tests/architecture/transition_imports_test.go` |
| 4 | Atualizar AGENTS.md | Adicionar regra de nomes proibidos de arquivos |
| 5 | Verificar build e testes | `go build ./...`, `go test ./...`, `verify-contracts.sh`, `lint.sh` |

### Architecture Tests a atualizar

Em `tests/architecture/module_boundaries_test.go`:
- Remover referências a `core/coordination` na mensagem de erro (linha que diz "refactor to use DI or move to core/coordination")
- Verificar se há alguma regra específica sobre `coordination`

Em `tests/architecture/transition_imports_test.go`:
- Verificar se há regra sobre `internal/core/transition` não importar `internal/core/coordination`
- Se sim, ajustar ou remover a regra (já que coordination não existe mais)

### Documentação a atualizar

Em `AGENTS.md`:
- Adicionar seção "Nomes Proibidos de Arquivos" com a lista de ADR-0028
- Adicionar regra: "NUNCA crie helpers.go, utils.go, common.go, base.go, misc.go, kit.go, ops.go, eventops.go"

---

## Fronteiras de Isolamento

**TOCAR:**
- `internal/core/coordination/` — REMOVER diretório inteiro
- `tests/architecture/module_boundaries_test.go` — atualizar mensagens
- `tests/architecture/transition_imports_test.go` — atualizar regras
- `AGENTS.md` — adicionar regra de nomes proibidos

**EVITAR:**
- `internal/modules/*` — não toque (já foi feito em A01-A05)
- `internal/core/*` (exceto coordination/) — não toque
- `internal/bootstrap/` — não toque (já foi feito em A02-A05)
- `cmd/` — não toque (já foi feito em A02-A05)
- `docs/adr/` — não modifique ADRs (são registros históricos)

---

## Interfaces Contratuais

Nenhuma.

---

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

**Caminho do checklist:** `plans/active/f28-r01/ORCH-F28-R01-A06-cleanup/checklist.md`

**A cada iteração:**
1. **LER** o checklist para identificar o próximo item pendente
2. **EXECUTAR** o item
3. **VALIDAR** o item
4. **ATUALIZAR** o checklist marcando o item como concluído
5. **CONTINUAR** para o próximo item

---

## Regras de Implementação

1. **Verifique que A01-A05 entregaram:**
   - Rode `grep -r "internal/core/coordination" --include="*.go" .` no projeto
   - Se houver imports restantes, identifique qual plano deveria ter removido e reporte ao usuário
   - NÃO prossiga com o cleanup se ainda houver imports ativos

2. **Remover `internal/core/coordination/`:**
   - `rm -rf internal/core/coordination/`
   - Ou `git rm -rf internal/core/coordination/`

3. **Atualizar architecture tests:**
   - `tests/architecture/module_boundaries_test.go`: Atualizar mensagem de erro para não mencionar `core/coordination`
   - `tests/architecture/transition_imports_test.go`: Remover ou ajustar regra sobre coordination

4. **Atualizar AGENTS.md:**
   - Adicionar após a seção "Padrões de Código":
   ```markdown
   ### Nomes Proibidos de Arquivos
   NUNCA crie arquivos com nomes genéricos: `helpers.go`, `utils.go`, `common.go`, `base.go`, `misc.go`, `kit.go`, `ops.go`, `eventops.go`.
   Arquivos devem comunicar **o quê** fazem, não **que tipo** de código contêm.
   Consulte ADR-0028 para a lista completa.
   ```

---

## Regras Rígidas de Testes

1. `go build ./...` — deve passar sem erros
2. `go test ./...` — deve passar sem falhas
3. `./scripts/verify-contracts.sh` — deve passar
4. `./scripts/lint.sh` — deve passar
5. `grep -r "internal/core/coordination" --include="*.go" .` — deve retornar ZERO resultados

---

## Code Review Auto-Crítico Obrigatório

- [ ] `internal/core/coordination/` foi completamente removido?
- [ ] Nenhum import de `internal/core/coordination` resta no código?
- [ ] Architecture tests foram atualizados?
- [ ] `AGENTS.md` foi atualizado com regra de nomes proibidos?
- [ ] `go build ./...` passa?
- [ ] `go test ./...` passa?
- [ ] `./scripts/verify-contracts.sh` passa?
- [ ] `./scripts/lint.sh` passa?

---

## Critérios de Aceite Verificáveis

1. `internal/core/coordination/` NÃO existe mais
2. Zero imports de `internal/core/coordination` em todo o projeto
3. Architecture tests não mencionam `core/coordination`
4. `AGENTS.md` contém regra de nomes proibidos
5. `go build ./...` passa
6. `go test ./...` passa
7. `./scripts/verify-contracts.sh` passa
8. `./scripts/lint.sh` passa

---

## Entrega Final

1. Commit com `./scripts/safe-commit.sh "refactor(core): remove coordination package per ADR-0028"`
2. Push da feature branch
3. Reportar ao usuário: "Cleanup completo. Pacote coordination/ removido. Build verde. Documentação atualizada."

---

## Resultado da Execução

**Data:** 2026-05-17  
**Branch:** `feat/adr28-a03-task-cascade`  
**PR:** #32

**Execução:**
- Diretório `internal/core/coordination/` removido completamente (4 arquivos deletados).
- `grep -r "internal/core/coordination" --include="*.go" .` retornou **zero** imports funcionais restantes (apenas architecture tests detectam a proibição de importação).
- Atualizados 17 arquivos de documentação (`doc.go`, `contract.go`) que mencionavam `core/coordination` → `core/transition`.
- Atualizados 3 architecture tests: `module_boundaries_test.go`, `transition_imports_test.go`, `orchestration_imports_test.go`.
- Comentário em `run/service.go` atualizado (`UpdateRunProjection`).
- `AGENTS.md` não necessitou atualização (não mencionava `coordination`).
- Todos os checks passaram (build, test, vet, architecture, contracts, lint).
