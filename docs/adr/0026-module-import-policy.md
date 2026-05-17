# 0026. Module Import Policy — O Que Pode Importar Quem

**Data:** 2026-05-17
**Status:** Accepted
**Relacionada:** ADR-0022, ADR-0025

---

## 1. Contexto

A migração ADR-0022 (Vertical Slice Architecture) criou módulos autônomos em `internal/modules/`. A ADR-0025 padronizou a estrutura dos módulos. Porém, **agentes de IA executando a migração encontram contradições**:

- `ADR-0022` diz: *"Um módulo NÃO PODE importar outro módulo"*.
- `ADR-0025` diz: *"NEVER import internal/modules/* directly"*.
- O `module_boundaries_test.go` permite: `run` → `task`.
- O código de `run/service.go` importa `taskmod` para `TaskReader`.
- O código de `run/service.go` ainda usa `domain.WorkUnit` porque não sabe se pode importar `workunit`.
- Os READMEs dizem: `"internal/domain: ONLY EventEnvelope and generic types"`, mas `domain/types.go` ainda tem `WorkUnit`, `Run`, `TaskGraph`.

Isso gera **paralisia de análise** nos agentes. Eles não sabem se devem:
1. Criar adapters `workunitToDomain()` em `bootstrap/`?
2. Definir `type WorkUnitRef struct{...}` local em `run/`?
3. Importar `workunit` diretamente em `run/`?

**A resposta não pode depender de interpretação. Precisa ser uma regra clara, codificada e testada.**

---

## 2. Decisão

Adotaremos uma política de importação **hierárquica e explícita**, com três pilares:

### Pilar 1: `internal/domain` é um pacote de infraestrutura, não de entidades

Após a migração ADR-0022, `internal/domain` **não deve mais conter entity structs** (`Task`, `WorkUnit`, `Run`, etc.). Ele deve conter **apenas tipos genuinamente compartilhados** entre múltiplos módulos e que não pertencem a nenhum módulo específico.

**O que FICA em `domain` (permanente):**

| Tipo | Por que fica? |
|------|---------------|
| `EventEnvelope` | Todos os módulos emitem e consomem eventos via `core/transition` |
| `EventPriority` | Constantes usadas por todos os módulos ao criar eventos |
| `CheckpointTrigger`, `CheckpointInput`, `HeartbeatInput`, `AutoCheckpointInput`, `CheckpointSuggestion` | Usados por `agentsession`, `coordination`, e runtimes. Não pertencem a nenhum módulo único |
| Payloads de eventos genéricos (`TaskGraphCreatedPayload`, `AgentLedgerUpdatedPayload`, etc.) | São "value objects" de eventos, não entidades. Podem ser usados por múltiplos consumidores de eventos |

**O que SAI de `domain` (migrar para módulos):**

| Tipo | Destino |
|------|---------|
| `Task`, `TaskStatus`, `Priority`, `RiskLevel` | `modules/task` |
| `WorkUnit`, `WorkUnitStatus` | `modules/workunit` |
| `Run`, `RunStatus`, `RunResult` | `modules/run` |
| `Agent`, `AgentRuntimeType`, `AgentStatus` | `modules/agent` ✅ já migrado |
| `AgentSession`, `AgentSessionStatus` | `modules/agentsession` |
| `TaskGraph`, `TaskGraphStatus` | `modules/taskgraph` |
| `PromptFragment`, `PromptFragmentRef`, `PromptSnapshot`, `ToolsetSnapshot`, `ToolsetTool` | `modules/prompt` |
| `Trigger`, `TriggerStatus`, `TriggerType`, `AnomalyType`, `ResolutionAction`, `ThresholdConfig` | `modules/trigger` |
| `Review`, `ReviewStatus`, `ValidationGate`, `ReviewCriteriaChecked` | `modules/review` |

**Regra absoluta:** Um módulo em `internal/modules/*` NUNCA importa `internal/domain` para obter uma **entity struct** ou **enum de entity**. Pode importar apenas os tipos listados na coluna "FICA" acima.

### Pilar 2: Imports entre módulos são permitidos para Interfaces DI

Um módulo `A` pode importar um módulo `B` **se e somente se**:

1. O import é usado **exclusivamente** em uma **interface de Injeção de Dependência (DI)**.
2. O tipo importado é usado **apenas como tipo de retorno** da interface (ex: `GetByID() (*b.SomeType, error)`).
3. `A` **nunca** chama `b.Service`, `b.Repository`, ou qualquer função/lógica de negócio de `B`.
4. A implementação da interface é **injeta em `internal/bootstrap/`**.

**Exemplo válido (já usado no código):**

```go
// internal/modules/run/service.go
import taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"

// TaskReader é uma interface DI. run NUNCA chama task.Service diretamente.
type TaskReader interface {
    GetByID(id string) (*taskmod.Task, error)
}
```

**Exemplo inválido (proibido):**

```go
// internal/modules/run/service.go
import "github.com/levygit837-cyber/OrchestraOS/internal/modules/workunit"

// ❌ PROIBIDO: chamar service/repository de outro módulo
func (s *RunService) Create(...) {
    wu, _ := workunit.NewRepository(tx).GetByID(id)  // ❌
    workunit.NewWorkUnitService(...).Block(...)       // ❌
}
```

**Exemplo inválido (proibido — entity struct de domain):**

```go
// internal/modules/run/service.go
import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

// ❌ PROIBIDO: usar entity struct de domain
type WorkUnitReader interface {
    GetByID(id string) (*domain.WorkUnit, error)  // ❌ domain.WorkUnit é entity
}
```

### Pilar 3: `core/coordination` e `orchestrator` são exceções arquiteturais

- `internal/core/coordination`: Pode importar **qualquer módulo** e `internal/domain`. É a camada de orquestração cross-module.
- `internal/modules/orchestrator`: Pode importar **qualquer módulo**, `core/coordination`, e `internal/domain`. É o módulo de orquestração de alto nível.
- `internal/bootstrap`: Pode importar **tudo**. É a camada de wiring/DI.
- `cmd/orchestraos`: Pode importar **tudo**. É a camada de entrada.

---

## 3. Regras por Pacote

### `internal/modules/*` (módulos de domínio)

| Pode importar | Condição |
|---------------|----------|
| `internal/domain` | Apenas `EventEnvelope`, `EventPriority`, `Checkpoint*`, `Heartbeat*`, payloads de eventos genéricos |
| `internal/modules/B` | Apenas para tipos de retorno em interfaces DI. Nunca para services/repositories |
| `internal/core/*` | Conforme matriz da ADR-0025 |
| `internal/bootstrap` | ❌ NUNCA |
| `internal/modules/orchestrator` | ❌ NUNCA |

### `internal/core/coordination`

| Pode importar | Condição |
|---------------|----------|
| `internal/modules/*` | ✅ SIM — é a função do pacote |
| `internal/domain` | ✅ SIM — ainda necessário durante transição |
| `internal/modules/orchestrator` | ❌ NUNCA (evita ciclo) |

### `internal/modules/orchestrator`

| Pode importar | Condição |
|---------------|----------|
| `internal/modules/*` | ✅ SIM |
| `internal/core/coordination` | ✅ SIM — único módulo permitido |
| `internal/domain` | ✅ SIM — durante transição |

### `internal/bootstrap`

| Pode importar | Condição |
|---------------|----------|
| Tudo | ✅ SIM — é a camada de wiring |

---

## 4. Como Isso Resolve a Confusão dos Agentes

Antes desta ADR, um agente migrando `run` via de regra fazia:

1. "Não posso importar `workunit`" → cria adapter em `bootstrap/services.go`
2. "Mas `service.go` ainda precisa de `*domain.WorkUnit`" → deixa `domain.WorkUnit`
3. Resultado: migração parcial, `domain` ainda usado para entity.

Com esta ADR, o agente segue um fluxo determinístico:

1. **Migra `workunit` primeiro** (A03): cria `WorkUnit` local em `workunit/models.go`.
2. **Migra `run` depois** (A02): atualiza `WorkUnitReader` para `GetByID() (*workunitmod.WorkUnit, error)`.
3. **Remove `domain.WorkUnit`** de `run/service.go`.
4. **Adiciona `run` → `workunit`** em `module_boundaries_test.go`.

Zero ambiguidade. Zero adapters desnecessários. O compilador é o guardião.

---

## 5. Consequências

### Positivas

- **Eliminação do monólito `domain/types.go`**: Entity structs migram para seus módulos. `domain` fica pequeno e estável.
- **Tipos fortes**: `run` recebe `*workunit.WorkUnit`, não `*domain.WorkUnit` nem `*run.WorkUnitRef` (tipo anêmico).
- **Compilador como aliado**: Se `workunit.WorkUnit` muda, `run` quebra em compile-time. Não há divergência silenciosa.
- **Menos código**: Adapters `XToDomain()` em `bootstrap/services.go` são eliminados ou minimizados.
- **LLMs eficientes**: Um agente tocando `workunit/models.go` sabe que `run` será atualizado automaticamente pelo compilador. Não precisa tocar 3 arquivos.

### Negativas

- **Acoplamento compile-time leve**: `run` depende de `workunit` em tempo de compilação. Mas é apenas um tipo, não lógica.
- **Matriz `allowedModuleImports` cresce**: O teste `module_boundaries_test.go` precisa ser atualizado a cada novo import permitido.
- **Refatoração inicial de `core/coordination`**: `coordination` ainda usa `domain.Run`, `domain.WorkUnit`. Será o último a ser limpo, pois precisa coordenar entre módulos que ainda não migraram.

---

## 6. Alternativas Consideradas

### Alternativa A: Regra Estrita — zero imports entre módulos

- **Como funciona:** Cada módulo define tipos anêmicos locais (`WorkUnitRef`) ou usa `domain.X` como lingua franca.
- **Problema:** `domain/types.go` vira monólito eterno. Adapters em `bootstrap` explodem. Tipos anêmicos divergem silenciosamente.
- **Descartada:** Preserva o problema que ADR-0022 tenta resolver.

### Alternativa B: Domain Bridge — manter `domain` como lingua franca

- **Como funciona:** Todos os módulos continuam usando `domain.WorkUnit`, `domain.Run`, etc.
- **Problema:** `domain/types.go` tem 390 linhas. Qualquer mudança afeta 9 módulos. Viola explicitamente os READMEs que dizem "NEVER import domain for entity structs".
- **Descartada:** É o status quo que estamos destruindo.

### Alternativa Vencedora: Import de Tipos para DI (esta ADR)

- **Como funciona:** Módulos importam apenas **tipos** de outros módulos para **interfaces DI**.
- **Vantagem:** Preserva isolamento de lógica de negócio, elimina `domain` como gargalo, mantém tipos fortes.

---

## 7. Checklist de Implementação

- [ ] Atualizar `module_boundaries_test.go` para codificar a nova regra (distinguir "import de tipo para DI" de "import de service/repository")
- [ ] Atualizar `contract.go` de todos os módulos — substituir "NEVER import internal/modules/*" pela regra refinada
- [ ] Atualizar `README.md` de todos os módulos — seção "Allowed Dependencies" refletindo a nova política
- [ ] Executar migrações pendentes (A02-A09) com a nova regra
- [ ] Remover entity structs de `internal/domain/types.go` conforme migrações avançam
- [ ] Atualizar `core/coordination` para usar tipos locais dos módulos (última etapa)
- [ ] Executar `go test ./...`, `go build ./...`, `./scripts/lint.sh`
- [ ] Commit via `./scripts/safe-commit.sh`
