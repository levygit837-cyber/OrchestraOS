# 0025. Module Standardization — Unified Structure, Rules and Contracts

**Data:** 2026-05-16

## 1. Contexto

Após a adoção do ADR-0022 (Vertical Slice Architecture), o projeto migrou de uma arquitetura em camadas para módulos verticais autônomos em `internal/modules/`. Porém, a migração revelou **inconsistências estruturais profundas**:

1. **Regras de `contract.go` são inventadas por módulo** — cada `contract.go` lista regras diferentes sem uma hierarquia clara (globais vs. específicas).
2. **Arquivos obrigatórios faltam em múltiplos módulos** — `validation.go` não existe em `task`, `run`, `agentsession`, `taskgraph`; `models.go` não existe em `prompt`; `events.go` não existe em `orchestrator`, `prompt`, `taskgraph`.
3. **Decomposições `service_*.go` não são documentadas** — `service_checkpoint.go`, `service_heartbeat.go`, `service_create.go`, `service_retry.go` existem sem serem mencionados nos `README.md` ou `CONTRACTS.md`.
4. **Regras de importação de `internal/core/*` são ad hoc** — `orchestrator` importa `core/coordination`, mas `agent/contract.go` proíbe explicitamente isso. Alguns módulos importam `core/serialization`, outros não.
5. **Não há distinção entre regra global, regra de tipo de módulo e regra específica** — um agente que lê `task/contract.go` não sabe quais regras valem só para `task` e quais valem para todos os módulos.

Isso gera **carga cognitiva desnecessária** para LLMs e humanos. Cada módulo é uma "surpresa" estrutural.

## 2. Decisão

Adotaremos um **padrão unificado e hierárquico** para todos os módulos em `internal/modules/*`. A estrutura, os arquivos, as regras de importação e o formato de `contract.go` serão **padronizados e obrigatórios**.

### 2.1 Hierarquia de Regras

Toda regra em `contract.go` deve se enquadrar em um dos três níveis:

```
REGRAS GLOBAIS      → Válidas para TODO módulo em internal/modules/*
REGRAS DE TIPO      → Válidas para um tipo de módulo (domínio, orquestração, infra)
REGRAS ESPECÍFICAS  → Válidas apenas para este módulo
```

### 2.2 Arquivos Obrigatórios (MUST)

Todo módulo **DEVE** conter exatamente estes arquivos:

| Arquivo | Responsabilidade | Pode faltar? |
|---------|------------------|-------------|
| `doc.go` | Package documentation, context briefing | ❌ Nunca |
| `contract.go` | ModuleContract + regras hierárquicas | ❌ Nunca |
| `README.md` | Propósito, file map, allowed dependencies | ❌ Nunca |
| `CONTRACTS.md` | Invariantes, state machine, boundary rules | ❌ Nunca |
| `models.go` | Tipos próprios (structs, enums, constants) | ❌ Nunca |
| `queries.go` | SQL constants (mesmo que placeholder) | ❌ Nunca |
| `repository.go` | CRUD puro, zero business logic | ❌ Nunca |
| `service.go` | Lógica principal, transações, eventos | ❌ Nunca |
| `events.go` | `EventTypeForStatus(status Status) string` | ❌ Nunca |
| `validation.go` | Validações sintáticas de input | ❌ Nunca |

**Exceção:** `orchestrator/` é um módulo de **coordenação**, não de domínio. Ele não possui tabela própria, então `repository.go` pode ser um placeholder. Mas o arquivo **deve existir**.

### 2.3 Arquivos Opcionais (MAY)

| Arquivo | Quando usar | Documentação obrigatória? |
|---------|-------------|--------------------------|
| `fetch.go` | Exportar `RequireByID` como DI para outros módulos | Sim, no README.md |
| `service_<sub>.go` | `service.go` > 300 linhas | **Sim** — README.md + CONTRACTS.md |
| `types.go` | Tipos auxiliares que não cabem em `models.go` | Sim, no README.md |
| `*_test.go` | Testes unitários | Recomendado |
| `<domain>.go` | Lógica específica (planners, detectores, composers) | Sim, no README.md |

**Regra de `service_<sub>.go`:**
1. Só é permitido se `service.go` tiver **> 300 linhas**.
2. O nome deve seguir `service_<verbo>.go` (ex: `service_create.go`, `service_retry.go`).
3. Deve ser listado no `README.md` na seção "File Map".
4. Deve ter uma justificativa de 1 linha no `CONTRACTS.md`.

### 2.4 Regras de Importação Padronizadas

#### 2.4.1 Regras Globais (todos os módulos)

```
1. NEVER import internal/modules/* directly.
2. NEVER import internal/domain for entity structs (ADR-0022).
3. NEVER write SQL strings outside queries.go.
4. NEVER call panic() — always return apperrors.Error.
5. NEVER put business logic inside repository.go.
6. NEVER call another module's Service methods — use DI interfaces or core/coordination.
7. NEVER ignore errors (_ = someCall()) without a documented reason.
8. ALWAYS emit a domain event inside the same transaction for every mutation.
```

#### 2.4.2 Regras de Tipo de Módulo

**Módulos de Domínio** (`task`, `run`, `workunit`, `taskgraph`, `agentsession`, `agent`, `prompt`, `trigger`, `review`):
```
1. Status transitions MUST call core/statemachine.CanTransition before mutating.
2. Terminal statuses (completed, failed, cancelled, stopped) are immutable.
3. Every mutation MUST emit exactly one domain event.
4. Input validation MUST use core/validation at module boundaries.
```

**Módulo de Coordenação** (`orchestrator`):
```
1. OrchestratorService is the ONLY component that may coordinate cross-module operations.
2. NEVER import repositories directly — use only domain services injected via Dependencies.
3. NEVER bypass core/statemachine rules of individual modules.
4. Work units are executed sequentially in the first cut (parallelism is future work).
```

#### 2.4.3 Matriz de Importação de `internal/core/*`

| Pacote | O que é | Pode ser importado por | Regra |
|--------|---------|----------------------|-------|
| `core/apperrors` | Erros tipados | **Todos** | Sempre permitido |
| `core/db` | Transações, helpers SQL | **Todos** | Sempre permitido |
| `core/validation` | Validações sintáticas | **Todos** | Sempre permitido |
| `core/event` | Event envelope genérico | **Todos** | Sempre permitido |
| `core/serialization` | Marshal de payloads | Módulos que emitem eventos | Se precisar de payload |
| `core/statemachine` | Regras de transição | Módulos com state machine | Se tiver status transitions |
| `core/transition` | Helpers de transição | Módulos com state machine | Se tiver status transitions |
| `core/eventstore` | Leitura de eventos | Módulos que leem event store | Se precisar de deduplicação |
| `core/coordination` | Coordenação cross-module | **APENAS `orchestrator/`** | Proibido para todos os outros |

**Nota:** `orchestrator/` é o **único** módulo que pode importar `core/coordination`. A regra atual em `agent/contract.go` que proíbe `core/coordination` será **generalizada** para todos os módulos de domínio.

### 2.5 Template Unificado de `contract.go`

Todo `contract.go` deve seguir **exatamente** esta estrutura de comentários:

```go
package <modulo>

import _ "embed"

// ============================================================================
// GLOBAL RULES — apply to ALL modules in internal/modules/*
// ============================================================================
//   1. NEVER import internal/modules/* directly.
//   2. NEVER import internal/domain for entity structs (ADR-0022).
//   3. NEVER write SQL strings outside queries.go.
//   4. NEVER call panic() — always return apperrors.Error.
//   5. NEVER put business logic inside repository.go.
//   6. NEVER call another module's Service methods — use DI interfaces or core/coordination.
//   7. NEVER ignore errors without a documented reason.
//   8. ALWAYS emit a domain event inside the same transaction for every mutation.
//
// ============================================================================
// MODULE-TYPE RULES — apply to ALL <domain|orchestration> modules
// ============================================================================
//   1. Status transitions MUST call core/statemachine.CanTransition before mutating.
//   2. Terminal statuses (completed, failed, cancelled, stopped) are immutable.
//   3. Every mutation MUST emit exactly one domain event.
//   4. Input validation MUST use core/validation at module boundaries.
//
// ============================================================================
// MODULE-SPECIFIC RULES — apply only to <modulo>
// ============================================================================
//   1. [Regra específica 1]
//   2. [Regra específica 2]
//   3. [Regra específica 3]
//
// ============================================================================
// ALLOWED / FORBIDDEN core/* IMPORTS
// ============================================================================
// ALLOWED: core/apperrors, core/db, core/validation, core/event
// ALLOWED: core/statemachine, core/transition, core/serialization (if emitting events)
// FORBIDDEN: core/coordination (except for orchestrator module)
//
// For full contracts and state machine, read CONTRACTS.md in this directory.
// For purpose and dependencies, read README.md in this directory.
// ============================================================================

//go:embed README.md
var _readme string

//go:embed CONTRACTS.md
var _contracts string

// ModuleContract marks this file as the entry point for LLM agents.
var ModuleContract = struct {
    Name    string
    Purpose string
}{
    Name:    "<modulo>",
    Purpose: "<descrição de uma linha>",
}
```

### 2.6 Seção Obrigatória em `README.md`

Todo `README.md` deve conter uma seção "File Map" que lista **todos** os arquivos Go do módulo:

```markdown
## File Map

### Mandatory Files
- `doc.go` → package documentation and context briefing
- `contract.go` → ModuleContract + hierarchical rules
- `models.go` → domain types (Task, Status, Priority, RiskLevel)
- `events.go` → event-type mapping for status transitions
- `queries.go` → SQL constants
- `repository.go` → CRUD pure
- `service.go` → main domain logic
- `validation.go` → input validation

### Decomposed Files (service.go > 300 lines)
- `service_create.go` → batch creation logic (extracted from service.go: 420 lines)

### Optional Files
- `fetch.go` → RequireByID exported for DI
```

### 2.7 Seção Obrigatória em `CONTRACTS.md`

Todo `CONTRACTS.md` deve conter uma seção "File Decomposition" se houver `service_*.go`:

```markdown
## File Decomposition

### service_create.go
- **Reason:** `service.go` reached 420 lines; creation logic is self-contained.
- **Rules:** Same transaction boundaries as service.go. Must call core/statemachine.CanTransition.
```

## 3. Consequências

### Positivas
- **Carga cognitiva reduzida:** Um LLM que entra em qualquer módulo já sabe exatamente o que esperar.
- **Automação facilitada:** Scripts de lint/arquitetura podem verificar estrutura automaticamente.
- **Onboarding rápido:** Novos módulos seguem um template único.
- **Contratos claros:** A hierarquia (global → tipo → específico) elimina ambiguidade.

### Negativas
- **Refatoração inicial:** Todos os 10 módulos precisam ter seus `contract.go` e `README.md` atualizados.
- **Arquivos faltando precisam ser criados:** `validation.go`, `events.go`, `models.go` em módulos que não têm.
- `orchestrator/` é o **único** módulo permitido a importar `core/coordination` (documentado e verificado por architecture test).

## 4. Alternativas Consideradas

### Alternativa A: Deixar cada módulo livre
- **Como funciona:** Cada módulo define sua própria estrutura e regras.
- **Problema:** É o caos atual. LLMs se perdem. Inconsistências se acumulam.
- **Descartada:** A premissa do projeto é ser operado por LLMs. Liberdade total = ineficiência.

### Alternativa B: Micro-serviços (cada módulo = repo separado)
- **Como funciona:** Cada módulo vive em seu próprio repositório Git.
- **Problema:** Overhead de versionamento de contratos, CI multi-repo, dificuldade de refatoração cross-module.
- **Descartada:** Para o estágio atual do OrchestraOS, é over-engineering. A padronização interna resolve o problema sem a sobrecarga operacional.

### Alternativa Vencedora: Padronização Interna (este ADR)
- **Como funciona:** Mantém o monorepo, mas impõe estrutura rígida via ADR + architecture tests.
- **Vantagem:** Baixo custo de implementação, alto retorno em previsibilidade.

## 5. Checklist de Implementação

- [ ] Criar template `contract.go.tmpl` em `docs/templates/module/`
- [ ] Atualizar `scripts/new-module.sh` para gerar todos os 10 arquivos obrigatórios
- [ ] Atualizar `docs/templates/module/README.md` com seção "File Map" obrigatória
- [ ] Atualizar `docs/templates/module/CONTRACTS.md` com seção "File Decomposition"
- [ ] Refatorar `contract.go` de todos os 10 módulos para seguir o template unificado
- [ ] Criar arquivos faltantes: `validation.go`, `events.go`, `models.go`
- [ ] Documentar decomposições `service_*.go` existentes nos `README.md` e `CONTRACTS.md`
- [ ] Resolver import de `core/coordination` em `orchestrator/` (mover para `internal/orchestration/` ou documentar exceção)
- [ ] Adicionar architecture test: verificar que todo módulo tem os 10 arquivos obrigatórios
- [ ] Adicionar architecture test: verificar que nenhum módulo (exceto orchestrator) importa `core/coordination`
- [x] Executar `go test ./...`, `go build ./...`, `./scripts/lint.sh`
- [x] Commit via `./scripts/safe-commit.sh`
