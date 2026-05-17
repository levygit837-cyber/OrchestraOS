# ORCH-F05-R03-A07 — Migração de Tipos: Prompt Module

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.
> **Branch esperada:** `adr22-a07-prompt-types`
> **Worktree esperada:** `../orchestraos-a07-prompt`
> Se não estiver isolado, execute:
> ```bash
> cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A07 prompt
> ```

---

## Contexto

O módulo `prompt` gerencia engenharia de prompts, snapshots e toolsets para execuções de agentes. É um dos módulos **mais complexos** porque consome **múltiplas entidades externas**: `Task`, `Run`, `WorkUnit`, `AgentSession`.

Atualmente não tem `models.go`. Usa diretamente do domain:
- `domain.PromptFragment`
- `domain.PromptFragmentRef`
- `domain.PromptSnapshot`
- `domain.ToolsetSnapshot`
- `domain.Run`, `domain.WorkUnit`, `domain.Task`, `domain.AgentSession` (em `PrepareAndPersistInput`)

**Pré-requisitos:**
- A01 (task) 🟢 — Task já migrado
- A02 (run) 🟢 — Run já migrado
- A03 (workunit) 🟢 — WorkUnit já migrado
- A05 (agentsession) 🟢 — AgentSession já migrado

---

## Documentação Obrigatória (ler ANTES)

1. `/home/levybonito/Documentos/OrchestraOS/internal/modules/prompt/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/internal/modules/prompt/CONTRACTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0022-llm-optimized-module-architecture.md`
4. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0007-prompt-composition-system.md`
5. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`
6. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-ADR22-MIGRATION/README.md`

---

## Inventário de Imports a Migrar

| Arquivo | Import de domain | O que usa |
|---------|-----------------|-----------|
| `prompt/repository.go` | `domain.PromptFragment`, `domain.PromptSnapshot` | CRUD de fragments e snapshots |
| `prompt/repository_snapshot.go` | `domain.PromptSnapshot`, `domain.ToolsetSnapshot` | CRUD de snapshots |
| `prompt/service.go` | `domain.Run`, `domain.WorkUnit`, `domain.Task`, `domain.AgentSession`, `domain.PromptSnapshot`, `domain.ToolsetSnapshot` | `PrepareAndPersistInput`, `PreparedRunPrompt` |

**Nota:** `prompt/types.go` e `prompt/composer.go` já definem tipos locais (`Fragment`, `Toolset`, etc.). O trabalho é migrar os que ainda dependem de `domain`.

---

## Estratégia de Migração

### Passo 1: Criar tipos locais em models.go (novo arquivo)

Criar `internal/modules/prompt/models.go`:

```go
package prompt

type PromptSnapshot struct {
    ID                 string              `json:"id"`
    RunID              string              `json:"run_id"`
    WorkUnitID         string              `json:"work_unit_id"`
    AgentSessionID     string              `json:"agent_session_id"`
    SystemPrompt       string              `json:"system_prompt"`
    TaskPrompt         string              `json:"task_prompt"`
    CombinedPrompt     string              `json:"combined_prompt"`
    SystemPromptHash   string              `json:"system_prompt_hash"`
    TaskPromptHash     string              `json:"task_prompt_hash"`
    CombinedPromptHash string              `json:"combined_prompt_hash"`
    CompositionHash    string              `json:"composition_hash"`
    CategorySignature  string              `json:"category_signature"`
    FragmentRefs       []PromptFragmentRef `json:"fragment_refs"`
    AssemblyOrder      []string            `json:"assembly_order"`
    VariablesApplied   json.RawMessage     `json:"variables_applied"`
    CountUsed          int                 `json:"count_used"`
    FirstUsedAt        time.Time           `json:"first_used_at"`
    LastUsedAt         time.Time           `json:"last_used_at"`
    CreatedAt          time.Time           `json:"created_at"`
}

type PromptFragment struct {
    ID               string          `json:"id"`
    Version          string          `json:"version"`
    Category         string          `json:"category"`
    Kind             string          `json:"kind"`
    Title            string          `json:"title"`
    Priority         int             `json:"priority"`
    ExclusiveGroup   string          `json:"exclusive_group"`
    BodyHash         string          `json:"body_hash"`
    MetadataHash     string          `json:"metadata_hash"`
    Body             string          `json:"body"`
    AppliesWhen      json.RawMessage `json:"applies_when,omitempty"`
    Requires         []string        `json:"requires,omitempty"`
    ConflictsWith    []string        `json:"conflicts_with,omitempty"`
    Allows           []string        `json:"allows,omitempty"`
    Denies           []string        `json:"denies,omitempty"`
    ApprovalRequired []string        `json:"approval_required,omitempty"`
    AutonomyLevel    int             `json:"autonomy_level,omitempty"`
    CreatedAt        time.Time       `json:"created_at"`
    UpdatedAt        time.Time       `json:"updated_at"`
}

type PromptFragmentRef struct {
    ID           string `json:"id"`
    Version      string `json:"version"`
    Category     string `json:"category"`
    Kind         string `json:"kind"`
    Order        int    `json:"order"`
    BodyHash     string `json:"body_hash"`
    MetadataHash string `json:"metadata_hash"`
    Title        string `json:"title"`
}

type ToolsetSnapshot struct {
    ID             string        `json:"id"`
    RunID          string        `json:"run_id"`
    AgentSessionID string        `json:"agent_session_id"`
    Tools          []ToolsetTool `json:"tools"`
    CreatedReason  string        `json:"created_reason"`
    CreatedAt      time.Time     `json:"created_at"`
}

type ToolsetTool struct {
    Name   string `json:"name"`
    Scope  string `json:"scope"`
    Risk   string `json:"risk"`
    Reason string `json:"reason,omitempty"`
}
```

### Passo 2: Atualizar arquivos internos
- `repository.go`: `domain.PromptFragment` → `PromptFragment`, `domain.PromptSnapshot` → `PromptSnapshot`
- `repository_snapshot.go`: `domain.PromptSnapshot` → `PromptSnapshot`, `domain.ToolsetSnapshot` → `ToolsetSnapshot`
- `service.go`: 
  - `PrepareAndPersistInput` usa `*run.Run`, `*workunit.WorkUnit`, `*task.Task`, `*agentsession.AgentSession`
  - `PreparedRunPrompt` usa `*PromptSnapshot`, `*ToolsetSnapshot`
  - Remover `promptFragmentToDomain()`, `promptFragmentRefsToDomain()`, `toolsetToolsToDomain()` (não serão mais necessários)

### Passo 3: Atualizar structs de input

```go
type PrepareAndPersistInput struct {
    Run                     *run.Run
    WorkUnit                *workunit.WorkUnit
    Task                    *task.Task
    Session                 *agentsession.AgentSession
    PromptSnapshotID        string
    ToolsetSnapshotID       string
    PromptSnapshotEventID   string
    ToolsetSnapshotEventID  string
}
```

### Passo 4: Criar Adapters Temporários nos Consumidores

| Consumidor | Motivo | Adapter |
|-----------|--------|---------|
| `internal/modules/orchestrator/models.go` | `PreparedPrompt` struct (`*domain.PromptSnapshot`, `*domain.ToolsetSnapshot`) | `// TODO[ADR-0022]: migrar para prompt.PromptSnapshot` |
| `internal/core/coordination/prompt_orchestrator.go` | Usa todas as entidades domain | Múltiplos adapters necessários |
| `internal/bootstrap/services.go` | prompt adapter | Atualizar com TODO |

### Passo 5: Build + Test + Commit
```bash
go build ./...
go test ./...
./scripts/safe-commit.sh "ADR-0022: migrate Prompt types to modules/prompt"
```

---

## Critérios de Aceitação

- [ ] `internal/modules/prompt/models.go` criado com `PromptSnapshot`, `PromptFragment`, `PromptFragmentRef`, `ToolsetSnapshot`, `ToolsetTool`
- [ ] Todos os arquivos em `internal/modules/prompt/` usam tipos locais
- [ ] `PrepareAndPersistInput` usa `*task.Task`, `*run.Run`, `*workunit.WorkUnit`, `*agentsession.AgentSession`
- [ ] Funções de conversão `*ToDomain()` removidas
- [ ] Todos os consumidores têm adapters com `// TODO[ADR-0022]: ...`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh` passa
