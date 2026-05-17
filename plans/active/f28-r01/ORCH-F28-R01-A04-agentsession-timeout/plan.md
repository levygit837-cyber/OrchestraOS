# ORCH-F28-R01-A04: AgentSession Module — Timeout Coordination

## Contexto do Projeto

**Projeto:** OrchestraOS (Go monolith modular)  
**Arquitetura:** Vertical Slice (ADR-0022) + Hybrid Intelligent Orchestrator (ADR-0023)  
**Meta:** Mover `ValidateRunForSessionCreation` e `AgentSessionTimeout` de `internal/core/coordination/` para `internal/modules/agentsession/` conforme ADR-0028  
**Risco:** **MÉDIO** — envolve cross-module (agentsession → run) e o `AgentSessionTimeout` usa `UpdateRunProjection`.

---

## Documentação Obrigatória (LEIA ANTES DE COMEÇAR)

1. `docs/adr/0028-core-architecture-and-naming-standards.md` — Seções 2.1.1, 2.4
2. `docs/adr/0022-vertical-slice-modules.md` — Regras de módulos e imports
3. `AGENTS.md` — Regras de commits
4. Este arquivo (`plan.md`)

---

## O que Já Existe vs O que Deve Ser Feito

### Arquivo-fonte a mover

| Arquivo Atual (coordination/) | Novo Arquivo (agentsession/) | Responsabilidade | Linhas |
|---|---|---|---|
| `agentsession_orchestrator.go` | `service_timeout.go` | Timeout de session + pause de run | ~52 |

### Pontos de integração a atualizar

| Arquivo | Uso Atual | Mudança Necessária |
|---|---|---|
| `internal/bootstrap/services.go` | Não usa diretamente (AgentSessionTimeout é chamado internamente pelo agentsession service) | Verificar se há referência em bootstrap |

### Desafio Arquitetural

`agentsession_orchestrator.go` importa `runmod` para:
- `runmod.RequireByID` (ler run pelo ID)
- `runmod.StatusRunning`, `runmod.StatusWaitingApproval`, `runmod.StatusPaused` (constantes de status)
- `UpdateRunProjection` (atualizar projeção do run)

O `allowedModuleImports` atual para `agentsession`:
```go
"agentsession": {"agent": true},
```

Isso significa que `agentsession` só pode importar `agent`. Importar `run` seria uma violação.

**Opções para resolver:**

1. **Adicionar `"run": true` ao `allowedModuleImports["agentsession"]`** — Justificativa: `AgentSessionTimeout` é um fluxo cross-module legítimo onde `agentsession` possui o processo (timeout da session afeta o run). Segue o ADR-0028.

2. **Usar callback/função injetada** — O `AgentSessionService` poderia receber uma função `pauseRun func(...)` como dependência, evitando importar `runmod` diretamente. Isso manteria `agentsession` como não-importador de `run`.

**RECOMENDAÇÃO DO PLANO:** Opção 1 é mais simples e alinhada com o ADR-0028. Adicione `"run": true` ao `allowedModuleImports["agentsession"]`. Documente: "agentsession imports run for session-timeout-to-run-pause coordination flow (ADR-0028)."

**Alternativa:** Se preferir manter a pureza do módulo, use Opção 2 (callback injetado). Ambas são aceitáveis — escolha uma e documente.

---

## Fronteiras de Isolamento

**TOCAR:**
- `internal/modules/agentsession/service_timeout.go` — NOVO arquivo com `ValidateRunForSessionCreation` e `AgentSessionTimeout`
- `internal/modules/agentsession/service.go` — possivelmente adicionar dependência se usar callback approach
- `tests/architecture/module_boundaries_test.go` — atualizar `allowedModuleImports["agentsession"]`
- `internal/core/coordination/agentsession_orchestrator.go` — REMOVER após mover lógica

**EVITAR:**
- `internal/modules/run/` — não toque (tarefa do A02)
- `internal/modules/task/` — não toque (tarefa do A03)
- `internal/modules/prompt/` — não toque (tarefa do A05)
- `internal/bootstrap/services.go` — não toque exceto se necessário para compilar
- `internal/modules/orchestrator/` — não toque

---

## Interfaces Contratuais

Nenhuma nova interface necessária. As funções mantêm suas assinaturas:

```go
func ValidateRunForSessionCreation(ctx context.Context, tx *sql.Tx, runID string) error
func AgentSessionTimeout(ctx context.Context, tx *sql.Tx, session *agentsessionmod.AgentSession, recoverableState json.RawMessage, input transition.TransitionInput) (*domain.EventEnvelope, bool, error)
```

Se `AgentSessionTimeout` virar um método do `AgentSessionService`:

```go
func (s *AgentSessionService) Timeout(ctx context.Context, tx *sql.Tx, session *AgentSession, recoverableState json.RawMessage, input transition.TransitionInput) (*domain.EventEnvelope, bool, error)
```

Neste caso, `ValidateRunForSessionCreation` pode ser privada (`validateRunForSessionCreation`) se só for usada internamente.

---

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

**Caminho do checklist:** `plans/active/f28-r01/ORCH-F28-R01-A04-agentsession-timeout/checklist.md`

**A cada iteração:**
1. **LER** o checklist para identificar o próximo item pendente
2. **EXECUTAR** o item (código, teste, refactor)
3. **VALIDAR** o item (`go build ./...` no escopo afetado)
4. **ATUALIZAR** o checklist marcando o item como concluído
5. **CONTINUAR** para o próximo item

---

## Regras de Implementação

1. **Mover funções:**
   - Copiar `ValidateRunForSessionCreation` e `AgentSessionTimeout` de `coordination/agentsession_orchestrator.go` para `agentsession/service_timeout.go`
   - Ajustar package de `coordination` para `agentsession`
   - Ajustar imports (remove import de `agentsessionmod` já que estamos no mesmo pacote)

2. **IMPORTANTE — Dependência do A02:**
   - `AgentSessionTimeout` chama `coordination.UpdateRunProjection`. Se A02 já entregou e moveu para `run/repository.go`, use `runmod.NewRepository(tx).UpdateProjection(...)`.
   - Se A02 ainda não entregou, mantenha import temporário de `coordination` com TODO.

3. **Atualizar `allowedModuleImports`:**
   - Adicionar `"run": true` à entrada `"agentsession"`
   - Documentar no comentário do `allowedModuleImports`

4. **Atualizar `AgentSessionService` (opcional):**
   - Se `AgentSessionTimeout` virar método do `AgentSessionService`, ajuste a struct e o construtor
   - Verifique se `AgentSessionService` já tem métodos similares (em `service_heartbeat.go`, `service_checkpoint.go`)

---

## Regras Rígidas de Testes

1. Rode `go test ./internal/modules/agentsession/...` — deve passar
2. Rode `go test ./tests/architecture/...` — deve passar
3. Rode `go test ./...` ao final — deve passar
4. Rode `./scripts/verify-contracts.sh` — deve passar
5. Rode `./scripts/lint.sh` — deve passar

---

## Code Review Auto-Crítico Obrigatório

- [ ] `internal/core/coordination/agentsession_orchestrator.go` foi removido?
- [ ] `internal/modules/agentsession/service_timeout.go` existe?
- [ ] `allowedModuleImports["agentsession"]` inclui `run`?
- [ ] `go build ./...` passa?
- [ ] `go test ./...` passa?
- [ ] Architecture tests passam?

---

## Critérios de Aceite Verificáveis

1. `internal/modules/agentsession/service_timeout.go` existe
2. `internal/core/coordination/agentsession_orchestrator.go` NÃO existe mais
3. `allowedModuleImports["agentsession"]` inclui `run`
4. `go build ./...` passa
5. `go test ./...` passa
6. `./scripts/verify-contracts.sh` passa
7. `./scripts/lint.sh` passa
8. Architecture tests passam

---

## Entrega Final

1. Commit com `./scripts/safe-commit.sh "refactor(agentsession): move timeout coordination from core per ADR-0028"`
2. Push da feature branch
3. Reportar ao usuário: "AgentSession timeout migration completa. Build verde."
