# ORCH-F05-R03-A08 — Migração de Tipos: Trigger Module

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.
> **Branch esperada:** `adr22-a08-trigger-types`
> **Worktree esperada:** `../orchestraos-a08-trigger`
> Se não estiver isolado, execute:
> ```bash
> cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A08 trigger
> ```

---

## Contexto

O módulo `trigger` implementa detecção de anomalias e triggers configuráveis (threshold, anomaly, heartbeat timeout, policy). Atualmente usa **aliases** para múltiplos tipos do domain:

```go
package trigger

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.TriggerStatus
type Type = domain.TriggerType
type Anomaly = domain.AnomalyType
type Resolution = domain.ResolutionAction
```

O módulo consome `Run`, `AgentSession` e `WorkUnit` via interfaces injetadas para avaliação.

**Pré-requisitos:**
- A02 (run) 🟢 — Run já migrado
- A03 (workunit) 🟢 — WorkUnit já migrado
- A05 (agentsession) 🟢 — AgentSession já migrado

---

## Documentação Obrigatória (ler ANTES)

1. `/home/levybonito/Documentos/OrchestraOS/internal/modules/trigger/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/internal/modules/trigger/CONTRACTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0022-llm-optimized-module-architecture.md`
4. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`
5. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-ADR22-MIGRATION/README.md`

---

## Inventário de Imports a Migrar

| Arquivo | Import de domain | O que usa |
|---------|-----------------|-----------|
| `trigger/models.go` | `domain.TriggerStatus`, `domain.TriggerType`, `domain.AnomalyType`, `domain.ResolutionAction` | aliases |
| `trigger/repository.go` | `domain.Trigger` | CRUD |
| `trigger/service.go` | `domain.Trigger`, `domain.TriggerStatus`, `domain.TriggerType`, `domain.AnomalyType`, `domain.ResolutionAction`, `domain.ThresholdConfig`, `domain.EventEnvelope` | service completo |
| `trigger/thresholds.go` | `domain.ThresholdConfig` | `DefaultThresholds()` |
| `trigger/detectors.go` | `domain.Trigger` | detectores de anomalia |
| `trigger/validation.go` | `domain.TriggerType`, `domain.TriggerStatus` | validações |
| `trigger/fetch.go` | `domain.Trigger` | `RequireByID` |
| `trigger/events.go` | `domain.TriggerStatus` | `EventTypeForStatus` |

---

## Estratégia de Migração

### Passo 1: Criar tipos locais em models.go

```go
package trigger

type Status string
const (
    StatusActive    Status = "active"
    StatusTriggered Status = "triggered"
    StatusResolved  Status = "resolved"
    StatusDismissed Status = "dismissed"
)

type Type string
const (
    TypeThreshold        Type = "threshold"
    TypeAnomaly          Type = "anomaly"
    TypeHeartbeatTimeout Type = "heartbeat_timeout"
    TypePolicy           Type = "policy"
)

type AnomalyType string
const (
    AnomalyStall         AnomalyType = "stall"
    AnomalyLoop          AnomalyType = "loop"
    AnomalyDrift         AnomalyType = "drift"
    AnomalyPathViolation AnomalyType = "path_violation"
    AnomalyTokenExceeded AnomalyType = "token_exceeded"
    AnomalyStepsExceeded AnomalyType = "steps_exceeded"
    AnomalyTimeExceeded  AnomalyType = "time_exceeded"
)

type ResolutionAction string
const (
    ResolutionPause    ResolutionAction = "pause"
    ResolutionCancel   ResolutionAction = "cancel"
    ResolutionNotify   ResolutionAction = "notify"
    ResolutionEscalate ResolutionAction = "escalate"
)

type ThresholdConfig struct {
    StallSeconds    int `json:"stall_seconds"`
    LoopRepetitions int `json:"loop_repetitions"`
    TokenMax        int `json:"token_max"`
    StepsMax        int `json:"steps_max"`
    TimeMaxSeconds  int `json:"time_max_seconds"`
}

type Trigger struct {
    ID               string            `json:"id"`
    RunID            *string           `json:"run_id,omitempty"`
    TaskID           *string           `json:"task_id,omitempty"`
    AgentSessionID   *string           `json:"agent_session_id,omitempty"`
    TriggerType      Type              `json:"trigger_type"`
    Status           Status            `json:"status"`
    AnomalyType      *AnomalyType      `json:"anomaly_type,omitempty"`
    ThresholdValue   json.RawMessage   `json:"threshold_value,omitempty"`
    CurrentValue     json.RawMessage   `json:"current_value,omitempty"`
    TriggeredAt      *time.Time        `json:"triggered_at,omitempty"`
    ResolvedAt       *time.Time        `json:"resolved_at,omitempty"`
    ResolutionAction *ResolutionAction `json:"resolution_action,omitempty"`
    CreatedAt        time.Time         `json:"created_at"`
}
```

### Passo 2: Atualizar arquivos internos
- `repository.go`: `*domain.Trigger` → `*Trigger`
- `service.go`: todos os `domain.*` → tipos locais
- `thresholds.go`: `domain.ThresholdConfig` → `ThresholdConfig`
- `detectors.go`: `domain.Trigger` → `Trigger`
- `validation.go`: `domain.TriggerType` → `Type`, `domain.TriggerStatus` → `Status`
- `fetch.go`: `*domain.Trigger` → `*Trigger`
- `events.go`: `domain.TriggerStatus` → `Status`

### Passo 3: Atualizar interfaces de leitura
- `RunReader` interface: se usa `*domain.Run`, atualizar para `*run.Run` (A02 concluído)
- `AgentSessionReader` interface: se usa `*domain.AgentSession`, atualizar para `*agentsession.AgentSession` (A05 concluído)
- `WorkUnitReader` interface: se usa `*domain.WorkUnit`, atualizar para `*workunit.WorkUnit` (A03 concluído)

### Passo 4: Criar Adapters Temporários nos Consumidores

| Consumidor | Motivo | Adapter |
|-----------|--------|---------|
| `internal/modules/orchestrator/models.go` | `TriggerEvaluator` interface | `// TODO[ADR-0022]: migrar para []*trigger.Trigger` |
| `internal/bootstrap/services.go` | trigger adapter | Atualizar com TODO |

### Passo 5: Build + Test + Commit
```bash
go build ./...
go test ./...
./scripts/safe-commit.sh "ADR-0022: migrate Trigger types to modules/trigger"
```

---

## Critérios de Aceitação

- [ ] `internal/modules/trigger/models.go` define `Trigger`, `Status`, `Type`, `AnomalyType`, `ResolutionAction`, `ThresholdConfig` localmente
- [ ] Todos os arquivos em `internal/modules/trigger/` usam tipos locais
- [ ] `RunReader` usa `*run.Run`, `AgentSessionReader` usa `*agentsession.AgentSession`, `WorkUnitReader` usa `*workunit.WorkUnit`
- [ ] Todos os consumidores têm adapters com `// TODO[ADR-0022]: ...`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh` passa
