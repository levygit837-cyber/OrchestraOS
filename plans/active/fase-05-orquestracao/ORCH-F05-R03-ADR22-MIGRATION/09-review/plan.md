# ORCH-F05-R03-A09 — Migração de Tipos: Review Module

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.
> **Branch esperada:** `adr22-a09-review-types`
> **Worktree esperada:** `../orchestraos-a09-review`
> Se não estiver isolado, execute:
> ```bash
> cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A09 review
> ```

---

## Contexto

O módulo `review` gerencia revisões ligadas a runs, work units ou tasks. Atualmente usa **aliases** para tipos do domain:

```go
package review

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Status = domain.ReviewStatus
type Gate = domain.ValidationGate
```

O módulo consome `Run`, `WorkUnit` e `Task` via IDs (string), não via structs. Isso simplifica a migração.

**Pré-requisitos:**
- A01 (task) 🟢 — Task já migrado
- A02 (run) 🟢 — Run já migrado
- A03 (workunit) 🟢 — WorkUnit já migrado

---

## Documentação Obrigatória (ler ANTES)

1. `/home/levybonito/Documentos/OrchestraOS/internal/modules/review/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/internal/modules/review/CONTRACTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0022-llm-optimized-module-architecture.md`
4. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`
5. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-ADR22-MIGRATION/README.md`

---

## Inventário de Imports a Migrar

| Arquivo | Import de domain | O que usa |
|---------|-----------------|-----------|
| `review/models.go` | `domain.ReviewStatus`, `domain.ValidationGate` | aliases `Status`, `Gate` |
| `review/repository.go` | `domain.Review` | CRUD |
| `review/service.go` | `domain.Review`, `domain.ReviewStatus`, `domain.ValidationGate`, `domain.ReviewCriteriaChecked`, `domain.ReviewDecision` | service completo |
| `review/validation.go` | `domain.ReviewStatus`, `domain.ValidationGate` | validações |

---

## Estratégia de Migração

### Passo 1: Criar tipos locais em models.go

```go
package review

type Status string
const (
    StatusPending          Status = "pending"
    StatusInProgress       Status = "in_progress"
    StatusApproved         Status = "approved"
    StatusChangesRequested Status = "changes_requested"
    StatusNeedsDiscussion  Status = "needs_discussion"
)

type ValidationGate string
const (
    GateHard   ValidationGate = "hard"
    GateSoft   ValidationGate = "soft"
    GatePolicy ValidationGate = "policy"
)

type Decision = Status

type CriteriaChecked struct {
    Criterion string `json:"criterion"`
    Passed    bool   `json:"passed"`
    Reason    string `json:"reason,omitempty"`
}

type Review struct {
    ID              string            `json:"id"`
    RunID           *string           `json:"run_id,omitempty"`
    WorkUnitID      *string           `json:"work_unit_id,omitempty"`
    TaskID          *string           `json:"task_id,omitempty"`
    AgentSessionID  *string           `json:"agent_session_id,omitempty"`
    ReviewerAgentID *string           `json:"reviewer_agent_id,omitempty"`
    GateType        ValidationGate    `json:"gate_type"`
    Status          Status            `json:"status"`
    VerdictReason   string            `json:"verdict_reason,omitempty"`
    EvidenceRefs    []string          `json:"evidence_refs,omitempty"`
    CriteriaChecked []CriteriaChecked `json:"criteria_checked,omitempty"`
    CreatedAt       time.Time         `json:"created_at"`
    UpdatedAt       time.Time         `json:"updated_at"`
    CompletedAt     *time.Time        `json:"completed_at,omitempty"`
}
```

### Passo 2: Atualizar arquivos internos
- `repository.go`: `*domain.Review` → `*Review`
- `service.go`: `domain.Review` → `Review`, `domain.ReviewStatus` → `Status`, `domain.ValidationGate` → `ValidationGate`, `domain.ReviewCriteriaChecked` → `CriteriaChecked`, `domain.ReviewDecision` → `Decision`
- `validation.go`: validar `Status` e `ValidationGate` locais

### Passo 3: Criar Adapters Temporários nos Consumidores

| Consumidor | Motivo | Adapter |
|-----------|--------|---------|
| `internal/modules/orchestrator/models.go` | `ReviewManager` interface | `// TODO[ADR-0022]: migrar para *review.Review` |
| `internal/bootstrap/services.go` | review adapter | Atualizar com TODO |

### Passo 4: Build + Test + Commit
```bash
go build ./...
go test ./...
./scripts/safe-commit.sh "ADR-0022: migrate Review types to modules/review"
```

---

## Critérios de Aceitação

- [ ] `internal/modules/review/models.go` define `Review`, `Status`, `ValidationGate`, `Decision`, `CriteriaChecked` localmente
- [ ] Todos os arquivos em `internal/modules/review/` usam tipos locais
- [ ] Todos os consumidores têm adapters com `// TODO[ADR-0022]: ...`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh` passa
