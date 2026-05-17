# ORCH-F05-R03-A05 — Migração de Tipos: AgentSession Module

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.
> **Branch esperada:** `adr22-a05-agentsession-types`
> **Worktree esperada:** `../orchestraos-a05-agentsession`
> Se não estiver isolado, execute:
> ```bash
> cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A05 agentsession
> ```

---

## Contexto

O módulo `agentsession` gerencia sessões de agentes — o binding entre um agent runtime e uma execução (Run). Atualmente usa **alias** para o status:

```go
package agentsession

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.AgentSessionStatus
```

O módulo consome `Agent` (do módulo agent) e `Run` (do módulo run) via interfaces injetadas.

**Pré-requisitos:**
- A02 (run) 🟢 — Run já migrado, usar `run.Run`
- A06 (agent) 🟢 — Agent já migrado, usar `agent.Agent`

---

## Documentação Obrigatória (ler ANTES)

1. `/home/levybonito/Documentos/OrchestraOS/internal/modules/agentsession/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/internal/modules/agentsession/CONTRACTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0022-llm-optimized-module-architecture.md`
4. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`
5. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-ADR22-MIGRATION/README.md`

---

## Inventário de Imports a Migrar

| Arquivo | Import de domain | O que usa |
|---------|-----------------|-----------|
| `agentsession/models.go` | `domain.AgentSessionStatus` | alias `Status` |
| `agentsession/repository.go` | `domain.AgentSession` | CRUD |
| `agentsession/service.go` | `domain.AgentSession`, `domain.AgentSessionStatus`, `domain.Agent` | `AgentReader` interface, transitions |
| `agentsession/service_checkpoint.go` | `domain.AgentSession` | checkpoint logic |
| `agentsession/service_heartbeat.go` | `domain.AgentSession` | heartbeat logic |
| `agentsession/fetch.go` | `domain.AgentSession` | `RequireByID` |
| `agentsession/events.go` | `domain.AgentSessionStatus` | `EventTypeForStatus` |
| `agentsession/checkpoint_policy.go` | `domain.AgentSession` | policy logic |

---

## Estratégia de Migração

### Passo 1: Criar tipos locais em models.go

```go
package agentsession

type Status string
const (
    StatusStarting        Status = "starting"
    StatusRunning         Status = "running"
    StatusWaitingApproval Status = "waiting_approval"
    StatusPaused          Status = "paused"
    StatusStopping        Status = "stopping"
    StatusStopped         Status = "stopped"
    StatusDisconnected    Status = "disconnected"
    StatusFailed          Status = "failed"
)

type AgentSession struct {
    ID               string          `json:"id"`
    AgentID          string          `json:"agent_id"`
    RunID            string          `json:"run_id"`
    TaskID           string          `json:"task_id"`
    WorkUnitID       string          `json:"work_unit_id"`
    SandboxID        string          `json:"sandbox_id"`
    ConnectionID     string          `json:"connection_id"`
    Status           Status          `json:"status"`
    LastHeartbeatAt  *time.Time      `json:"last_heartbeat_at,omitempty"`
    LastCheckpointAt *time.Time      `json:"last_checkpoint_at,omitempty"`
    LastSeenEventID  string          `json:"last_seen_event_id,omitempty"`
    RecoverableState json.RawMessage `json:"recoverable_state,omitempty"`
}
```

### Passo 2: Atualizar arquivos internos
- `repository.go`: `*domain.AgentSession` → `*AgentSession`
- `service.go`: `domain.AgentSession` → `AgentSession`, `domain.AgentSessionStatus` → `Status`, `domain.Agent` → `agent.Agent`
- `service_checkpoint.go`: `domain.AgentSession` → `AgentSession`
- `service_heartbeat.go`: `domain.AgentSession` → `AgentSession`
- `fetch.go`: `*domain.AgentSession` → `*AgentSession`
- `events.go`: `domain.AgentSessionStatus` → `Status`
- `checkpoint_policy.go`: `domain.AgentSession` → `AgentSession`

### Passo 3: Atualizar interfaces cruzadas
- `AgentReader` interface: se usa `*domain.Agent`, atualizar para `*agent.Agent` (agent já migrado em A06)
- Verificar `Run` references — se A02 concluído, usar `run.Run`; senão, adapter

### Passo 4: Criar Adapters Temporários nos Consumidores

| Consumidor | Motivo | Adapter |
|-----------|--------|---------|
| `internal/modules/orchestrator/models.go` | `SessionManager` interface | `// TODO[ADR-0022]: migrar para *agentsession.AgentSession` |
| `internal/core/coordination/prompt_orchestrator.go` | usa `domain.AgentSession` | Criar `agentsessionToDomain()` |
| `internal/modules/trigger/service.go` | `AgentSessionReader` interface | `// TODO[ADR-0022]: migrar para *agentsession.AgentSession` |
| `internal/bootstrap/services.go` | agentsession adapter | Atualizar com TODO |

### Passo 5: Build + Test + Commit
```bash
go build ./...
go test ./...
./scripts/safe-commit.sh "ADR-0022: migrate AgentSession types to modules/agentsession"
```

---

## Critérios de Aceitação

- [ ] `internal/modules/agentsession/models.go` define `AgentSession`, `Status` localmente
- [ ] Todos os arquivos em `internal/modules/agentsession/` usam tipos locais
- [ ] `AgentReader` interface usa `*agent.Agent` (agent já migrado)
- [ ] Todos os consumidores têm adapters com `// TODO[ADR-0022]: ...`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh` passa
