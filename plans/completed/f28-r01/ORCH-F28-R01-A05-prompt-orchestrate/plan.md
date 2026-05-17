# ORCH-F28-R01-A05: Prompt Module — Cross-Module Orchestration

## Contexto do Projeto

**Projeto:** OrchestraOS (Go monolith modular)  
**Arquitetura:** Vertical Slice (ADR-0022) + Hybrid Intelligent Orchestrator (ADR-0023)  
**Meta:** Mover `PromptOrchestrator` de `internal/core/coordination/` para `internal/modules/prompt/` conforme ADR-0028  
**Risco:** **MÉDIO** — toca em bootstrap, cmd, orchestrator, e múltiplos testes de integração.

---

## Documentação Obrigatória (LEIA ANTES DE COMEÇAR)

1. `docs/adr/0028-core-architecture-and-naming-standards.md` — Seções 2.1.1, 2.2.1
2. `docs/adr/0022-vertical-slice-modules.md` — Regras de módulos
3. `AGENTS.md` — Regras de commits
4. Este arquivo (`plan.md`)

---

## O que Já Existe vs O que Deve Ser Feito

### Arquivo-fonte a mover

| Arquivo Atual (coordination/) | Novo Arquivo (prompt/) | Responsabilidade | Linhas |
|---|---|---|---|
| `prompt_orchestrator.go` | `service_orchestrate.go` | Gather cross-module para preparação de prompt | ~92 |

### Pontos de integração a atualizar

| Arquivo | Uso Atual | Mudança Necessária |
|---|---|---|
| `internal/bootstrap/services.go` | `coordination.NewPromptOrchestrator(db, promptSvc)` | `promptmod.NewPromptOrchestrator(db, promptSvc)` (ou método do PromptService) |
| `internal/bootstrap/services.go` | `*coordination.PromptOrchestrator` em `promptAdapter` | `*promptmod.PromptOrchestrator` |
| `internal/modules/orchestrator/service.go` | `PromptOrchestrator` campo em `Dependencies` | Manter interface `PromptPreparer`, mas implementação vem de `promptmod` |
| `cmd/orchestraos/cmd/run.go` | `coordination.NewPromptOrchestrator(...).PrepareRunPrompt(...)` | `promptmod.NewPromptOrchestrator(...).PrepareRunPrompt(...)` |
| `tests/integration/e2e_orchestration_test.go` | `coordination.NewPromptOrchestrator(...)` | `promptmod.NewPromptOrchestrator(...)` |
| `tests/integration/services_test.go` | `coordination.NewPromptOrchestrator(...)` | `promptmod.NewPromptOrchestrator(...)` |

### Vantagem Arquitetural

O módulo `prompt` JÁ importa `runmod`, `workunitmod`, `taskmod`, `agentsessionmod` em `prompt/service.go` (para `PrepareAndPersistInput`). Portanto, mover `PromptOrchestrator` para `prompt/` não adiciona novos imports cross-module. É a mudança mais limpa de todas.

---

## Fronteiras de Isolamento

**TOCAR:**
- `internal/modules/prompt/service_orchestrate.go` — NOVO arquivo com `PromptOrchestrator` struct e `PrepareRunPrompt`
- `internal/bootstrap/services.go` — ajustar `PromptOrchestrator` instantiation e tipo
- `internal/modules/orchestrator/service.go` — ajustar se necessário (o tipo já é interface `PromptPreparer`)
- `cmd/orchestraos/cmd/run.go` — ajustar import e chamada
- `tests/integration/e2e_orchestration_test.go` — ajustar import e chamada
- `tests/integration/services_test.go` — ajustar import e chamada
- `internal/core/coordination/prompt_orchestrator.go` — REMOVER após mover lógica

**EVITAR:**
- `internal/modules/run/` — não toque (tarefa do A02)
- `internal/modules/task/` — não toque (tarefa do A03)
- `internal/modules/agentsession/` — não toque (tarefa do A04)
- `internal/core/transition/` — não toque exceto se necessário para compilar

---

## Interfaces Contratuais

O `orchestrator` module já define uma interface `PromptPreparer` (verifique em `orchestrator/service.go` ou contratos). O `PromptOrchestrator` deve implementar essa interface. Verifique:

```go
// Em internal/modules/orchestrator/ — já deve existir:
type PromptPreparer interface {
    PrepareRunPrompt(ctx context.Context, input promptmod.PrepareRunPromptInput) (*promptmod.PreparedRunPrompt, error)
}
```

Se não existir, o `orchestrator` está usando o tipo concreto `*coordination.PromptOrchestrator`. Neste caso, crie a interface no `orchestrator` e faça o `promptmod.PromptOrchestrator` implementá-la.

---

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

**Caminho do checklist:** `plans/active/f28-r01/ORCH-F28-R01-A05-prompt-orchestrate/checklist.md`

**A cada iteração:**
1. **LER** o checklist para identificar o próximo item pendente
2. **EXECUTAR** o item (código, teste, refactor)
3. **VALIDAR** o item (`go build ./...` no escopo afetado)
4. **ATUALIZAR** o checklist marcando o item como concluído
5. **CONTINUAR** para o próximo item

---

## Regras de Implementação

1. **Mover `PromptOrchestrator`:**
   - Copiar struct, construtor e método `PrepareRunPrompt` de `coordination/prompt_orchestrator.go` para `prompt/service_orchestrate.go`
   - Ajustar package de `coordination` para `prompt`
   - Ajustar imports (remove `promptmod` já que estamos no mesmo pacote; usar tipos diretamente: `PrepareRunPromptInput`, `PreparedRunPrompt`)

2. **Considerar fusão com `PromptService`:**
   - O `PromptOrchestrator` tem um campo `promptService *promptmod.PromptService` e chama `o.promptService.PrepareAndPersistPrompt(...)`
   - Isso é um indireção desnecessária se ambos estão no mesmo pacote
   - **RECOMENDAÇÃO:** Em vez de manter `PromptOrchestrator` como struct separada, considere adicionar `PrepareRunPrompt` como método direto do `PromptService`. Isso elimina a necessidade de uma struct intermediária.
   - Se fizer isso, `NewPromptService` receberia `db` (já recebe) e `PrepareRunPrompt` seria um método de `PromptService`.
   - **ALTERNATIVA:** Se preferir manter a separação, mantenha `PromptOrchestrator` como struct em `prompt/service_orchestrate.go`.

3. **Atualizar `bootstrap/services.go`:**
   - Se `PromptOrchestrator` virou método do `PromptService`: simplifique — `promptAdapter` chama `promptSvc.PrepareRunPrompt(...)` diretamente
   - Se mantém struct separada: `promptmod.NewPromptOrchestrator(db, promptSvc)`
   - Remover import de `internal/core/coordination`

4. **Atualizar `cmd/orchestraos/cmd/run.go`:**
   - Remover import de `internal/core/coordination`
   - Ajustar chamada de `coordination.NewPromptOrchestrator(...)` para `promptmod.NewPromptOrchestrator(...)` ou `bootstrap.PromptService(...).PrepareRunPrompt(...)`

5. **Atualizar testes de integração:**
   - `tests/integration/e2e_orchestration_test.go`
   - `tests/integration/services_test.go`
   - Substituir `coordination.NewPromptOrchestrator` por `promptmod.NewPromptOrchestrator`

---

## Regras Rígidas de Testes

1. Rode `go test ./internal/modules/prompt/...` — deve passar
2. Rode `go test ./tests/integration/...` — deve passar
3. Rode `go test ./...` ao final — deve passar
4. Rode `./scripts/verify-contracts.sh` — deve passar
5. Rode `./scripts/lint.sh` — deve passar

---

## Code Review Auto-Crítico Obrigatório

- [ ] `internal/core/coordination/prompt_orchestrator.go` foi removido?
- [ ] `internal/modules/prompt/service_orchestrate.go` existe?
- [ ] `bootstrap/services.go` não importa mais `internal/core/coordination`?
- [ ] `cmd/orchestraos/cmd/run.go` não usa mais símbolos de `coordination`?
- [ ] Testes de integração foram atualizados?
- [ ] `go build ./...` passa?
- [ ] `go test ./...` passa?

---

## Critérios de Aceite Verificáveis

1. `internal/modules/prompt/service_orchestrate.go` existe
2. `internal/core/coordination/prompt_orchestrator.go` NÃO existe mais
3. `bootstrap/services.go` não usa símbolos de `coordination`
4. `cmd/orchestraos/cmd/run.go` não usa símbolos de `coordination`
5. Testes de integração não importam `internal/core/coordination`
6. `go build ./...` passa
7. `go test ./...` passa
8. `./scripts/verify-contracts.sh` passa
9. `./scripts/lint.sh` passa

---

## Entrega Final

1. Commit com `./scripts/safe-commit.sh "refactor(prompt): move PromptOrchestrator from coordination per ADR-0028"`
2. Push da feature branch
3. Reportar ao usuário: "Prompt orchestration migration completa. Build verde."
