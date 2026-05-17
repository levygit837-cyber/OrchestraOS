# ORCH-F05-R03-A06 — Migração de Tipos: Agent Module

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.
> **Branch esperada:** `adr22-a06-agent-types`
> **Worktree esperada:** `../orchestraos-a06-agent`
> Se não estiver isolado, execute:
> ```bash
> cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A06 agent
> ```

---

## Contexto

O módulo `agent` define interfaces e implementações de runtime para agentes de IA (fake, gemini, codex-cli, external). Atualmente tem `RuntimeType` local, mas ainda usa `domain.Agent` e `domain.AgentRuntimeType` para persistência:

```go
package agent

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// ToDomainRuntimeType converte local para domain
type RuntimeType string

func ToDomainRuntimeType(rt RuntimeType) domain.AgentRuntimeType { ... }
func FromDomainRuntimeType(rt domain.AgentRuntimeType) RuntimeType { ... }
```

**Pré-requisitos:** Nenhum. Agent é independente. Pode rodar em paralelo com A02, A03, A04.

---

## Documentação Obrigatória (ler ANTES)

1. `/home/levybonito/Documentos/OrchestraOS/internal/modules/agent/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/internal/modules/agent/CONTRACTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0022-llm-optimized-module-architecture.md`
4. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-A01-adr-0022-types-migration/plan.md`
5. `/home/levybonito/Documentos/OrchestraOS/plans/active/fase-05-orquestracao/ORCH-F05-R03-ADR22-MIGRATION/README.md`

---

## Inventário de Imports a Migrar

| Arquivo | Import de domain | O que usa |
|---------|-----------------|-----------|
| `agent/models.go` | `domain.AgentRuntimeType` | `ToDomainRuntimeType()`, `FromDomainRuntimeType()` |
| `agent/repository.go` | `domain.Agent` | CRUD |
| `agent/service.go` | `domain.Agent`, `domain.AgentRuntimeType` | `AgentService`, `FindOrCreate` |
| `agent/validation.go` | `domain.AgentRuntimeType` | `ValidateRuntimeType` |
| `agent/contract.go` | `domain.Agent` | `AgentReader` interface (`GetByID`) |

---

## Estratégia de Migração

### Passo 1: Criar tipos locais em models.go

```go
package agent

type RuntimeType string
const (
    RuntimeTypeCodexCLI RuntimeType = "codex_cli"
    RuntimeTypeFake     RuntimeType = "fake"
    RuntimeTypeExternal RuntimeType = "external"
    RuntimeTypeGemini   RuntimeType = "gemini"
)

type AgentStatus string
const (
    AgentStatusActive   AgentStatus = "active"
    AgentStatusInactive AgentStatus = "inactive"
)

type Agent struct {
    ID                     string       `json:"id"`
    Name                   string       `json:"name"`
    Profile                string       `json:"profile"`
    Capabilities           []string     `json:"capabilities"`
    AllowedTools           []string     `json:"allowed_tools"`
    DefaultPromptFragments []string     `json:"default_prompt_fragments"`
    RuntimeType            RuntimeType  `json:"runtime_type"`
    Status                 AgentStatus  `json:"status"`
    CreatedAt              time.Time    `json:"created_at"`
    UpdatedAt              time.Time    `json:"updated_at"`
}
```

### Passo 2: Remover funções de conversão
- Apagar `ToDomainRuntimeType()`
- Apagar `FromDomainRuntimeType()`
- O módulo agent deve ser 100% limpo de `domain`

### Passo 3: Atualizar arquivos internos
- `repository.go`: `*domain.Agent` → `*Agent`
- `service.go`: `domain.Agent` → `Agent`, `domain.AgentRuntimeType` → `RuntimeType`
- `validation.go`: validar `RuntimeType` local
- `contract.go`: `AgentReader` interface retorna `*Agent` (não `*domain.Agent`)

### Passo 4: Criar Adapters Temporários nos Consumidores

| Consumidor | Motivo | Adapter |
|-----------|--------|---------|
| `internal/modules/orchestrator/models.go` | `AgentManager` interface | `// TODO[ADR-0022]: migrar para *agent.Agent` |
| `internal/modules/agentsession/service.go` | `AgentReader` interface | `// TODO[ADR-0022]: migrar para *agent.Agent` |
| `internal/bootstrap/services.go` | agent adapter | Atualizar com TODO |

### Passo 5: Build + Test + Commit
```bash
go build ./...
go test ./...
./scripts/safe-commit.sh "ADR-0022: migrate Agent types to modules/agent"
```

---

## Critérios de Aceitação

- [ ] `internal/modules/agent/models.go` define `Agent`, `RuntimeType`, `AgentStatus` localmente
- [ ] `ToDomainRuntimeType()` e `FromDomainRuntimeType()` removidos
- [ ] Todos os arquivos em `internal/modules/agent/` usam tipos locais
- [ ] `AgentReader` interface retorna `*Agent` (não `*domain.Agent`)
- [ ] Todos os consumidores têm adapters com `// TODO[ADR-0022]: ...`
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/safe-commit.sh` passa
