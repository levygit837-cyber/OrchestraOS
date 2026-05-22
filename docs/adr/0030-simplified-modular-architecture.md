# 0030. Arquitetura Modular Simplificada

**Status:** Accepted  
**Data:** 2026-05-21  
**Supersedes:** ADR-0022

---

## 1. Contexto

A ADR-0022 introduziu uma Arquitetura de Módulos Verticais (Vertical Slice) com regras rigorosas:
- Cada módulo define seus próprios tipos em `models.go`
- `internal/domain/` contém apenas infraestrutura (EventEnvelope, checkpoints)
- Uma whitelist de 44 imports cross-module permitidos via DI
- 10 arquivos obrigatórios por módulo
- 10+ testes de arquitetura verificando estrutura

### 1.1 Falhas Observadas

Após 10 dias de operação, uma auditoria de confiabilidade arquitetural (2026-05-21) revelou:

| Problema | Evidência |
|----------|-----------|
| **Testes passam, violações existem** | 83+ violações reais apesar de todos os testes verdes |
| **Whitelist é uma fachada** | 44 imports cross-module permitidos mascaram acoplamento real |
| **Testes verificam estrutura, não comportamento** | `TestModuleBoundaries` verifica se import está na lista, não como é usado |
| **Complexidade excessiva** | 10 arquivos obrigatórios, 10+ testes de arquitetura, 83+ regras implícitas |
| **Agentes não conseguem seguir** | LLMs ignoram `CONTRACTS.md` e `contract.go` porque são muito longos |
| **Duplicação de tipos** | Cada módulo redefine conceitos que referenciam os mesmos dados (Task, Run, WorkUnit) |

### 1.2 O Problema Raiz

A arquitetura Vertical Slice funciona para **humanos experientes** em projetos maduros, mas falha para **LLMs operando em código desconhecido** porque:

1. **Contexto fragmentado:** Um tipo `Task` é definido em `task/models.go`, mas referenciado em 9 outros módulos. O LLM precisa ler 10 arquivos para entender um tipo.
2. **Regras implícitas demais:** A whitelist de imports é um mapa de `string→bool` que nenhum LLM entende sem explicação extensa.
3. **Testes de estrutura são inúteis:** Verificar que `contract.go` existe não impede que o módulo viole contratos.

### 1.3 Necessidade de Simplificação

Precisamos de uma arquitetura que um LLM possa **entender em 30 segundos** e **verificar em 3 testes**:

| Requisito | Justificativa |
|-----------|---------------|
| Regras contáveis em uma mão | LLMs não leem listas longas |
| Testes que falham com violações reais | Fachada de segurança é pior que nenhuma segurança |
| Tipos compartilhados em um só lugar | Reduz fragmentação de contexto |
| Módulos isolados sem exceções | Zero ambiguidade sobre o que é permitido |

---

## 2. Decisão

### 2.1 Decisão Principal

Simplificaremos a arquitetura para um **Modular Monolith** com regras mínimas:

> **Pilar 1:** `internal/domain/` centraliza **TODOS** os tipos compartilhados (entity structs e enums).  
> **Pilar 2:** Módulos em `internal/modules/` **NÃO importam outros módulos**. Zero exceções.  
> **Pilar 3:** Apenas `internal/bootstrap/` e `internal/modules/orchestrator/` importam múltiplos módulos.  
> **Pilar 4:** `repository.go` é CRUD puro — sem business logic, sem timestamps, sem deduplication.

### 2.2 O que muda em relação à ADR-0022

| Aspecto | ADR-0022 (Antigo) | ADR-0030 (Novo) |
|---------|-------------------|-----------------|
| `internal/domain/` | Apenas infraestrutura (EventEnvelope) | **Todos os entity types** (Task, Run, WorkUnit, etc.) |
| Imports cross-module | 44 permitidos via whitelist | **Zero permitidos** |
| Arquivos obrigatórios por módulo | 10 | **~4-5** (doc.go, models.go, repository.go, service.go, README.md) |
| `contract.go` + `CONTRACTS.md` | Obrigatórios | **Opcionais** (mantidos se úteis para LLMs) |
| Testes de arquitetura | 10+ testes complexos | **3-4 testes simples** |
| Regras totais | ~83 implícitas | **~15 explícitas** |

### 2.3 Estrutura de Diretórios

```
internal/
├── domain/           # TODOS os tipos compartilhados
│   ├── types.go      # Entity structs: Task, Run, WorkUnit, Agent, etc.
│   ├── doc.go        # Documentação do pacote
│   └── ...           # EventEnvelope, checkpoint types (mantidos)
├── modules/          # Lógica de negócio isolada
│   ├── task/         # TaskService, TaskRepository (usa domain.Task)
│   ├── run/          # RunService, RunRepository (usa domain.Run)
│   ├── workunit/     # WorkUnitService, WorkUnitRepository (usa domain.WorkUnit)
│   ├── agent/
│   ├── agentsession/
│   ├── taskgraph/
│   ├── prompt/
│   ├── review/
│   ├── trigger/
│   └── orchestrator/ # ÚNICO módulo que importa outros módulos
├── bootstrap/        # Monta DI, importa múltiplos módulos
│   └── services.go
└── core/             # Infraestrutura compartilhada (db, eventstore, transition)
```

### 2.4 Regras Detalhadas

#### Regra 1: Domain Centraliza Tipos

Todos os entity types compartilhados entre 2+ módulos vivem em `internal/domain/`:

```go
// internal/domain/types.go
package domain

type Task struct {
    ID          string
    Title       string
    Status      TaskStatus
    // ...
}

type TaskStatus string
const (
    TaskStatusCreated   TaskStatus = "created"
    TaskStatusRunning   TaskStatus = "running"
    TaskStatusCompleted TaskStatus = "completed"
)
```

Módulos importam `internal/domain` para usar esses tipos:

```go
// internal/modules/task/repository.go
package task

import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

type Repository struct{}

func (r *Repository) GetByID(id string) (*domain.Task, error) { ... }
```

#### Regra 2: Zero Imports Cross-Module

NENHUM módulo em `internal/modules/X/` pode importar `internal/modules/Y/` (onde X ≠ Y).

**ANTES (violando ADR-0030):**
```go
// internal/modules/run/service.go
import taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"

func (s *Service) requireTaskByID(id string) (*taskmod.Task, error) { ... }
```

**DEPOIS (correto ADR-0030):**
```go
// internal/modules/run/service.go
import "github.com/levygit837-cyber/OrchestraOS/internal/domain"

func (s *Service) requireTaskByID(id string) (*domain.Task, error) { ... }
// ou usa DI interface que retorna *domain.Task
```

**Exceções:**
- `internal/modules/orchestrator/` — pode importar múltiplos módulos (é o coordination layer)
- `internal/bootstrap/` — pode importar múltiplos módulos (monta DI)
- `cmd/` — deve usar `internal/bootstrap/` (não instanciar diretamente)

#### Regra 3: Repository é CRUD Puro

`repository.go` contém apenas:
- Queries SQL (em `queries.go` ou inline)
- Execução de INSERT, SELECT, UPDATE, DELETE
- Mapeamento row→struct e struct→row

`repository.go` NÃO contém:
- `if status == StatusRunning` (lógica de status → vai para `service.go`)
- `time.Now()` (timestamp deve ser passado pelo service)
- `ON CONFLICT` (deduplication → vai para `service.go`)
- Validação de campos (`if id == ""` → vai para `service.go` ou `validation.go`)

#### Regra 4: Estrutura de Módulo Simplificada

Cada módulo deve ter:

| Arquivo | Obrigatório | Conteúdo |
|---------|-------------|----------|
| `doc.go` | Sim | Documentação do pacote |
| `README.md` | Sim | Mapa operacional, responsabilidades |
| `models.go` | Sim | Tipos locais do módulo (NÃO entity types compartilhados) |
| `repository.go` | Sim | CRUD puro |
| `service.go` | Sim | Lógica de negócio, state transitions |
| `queries.go` | Opcional | SQL constants (se houver muitas queries) |
| `contract.go` | Opcional | Gateway para README.md/CONTRACTS.md |
| `CONTRACTS.md` | Opcional | Invariantes, state machine |
| `events.go` | Opcional | Event types e payloads |
| `validation.go` | Opcional | Regras de validação |

### 2.5 Tipos que Devem Migrar para `internal/domain/`

| Tipo | Origem | Destino |
|------|--------|---------|
| Task, TaskStatus, TaskPriority, TaskRiskLevel | `internal/modules/task/models.go` | `internal/domain/types.go` |
| Run, RunStatus, RunResult | `internal/modules/run/models.go` | `internal/domain/types.go` |
| WorkUnit, WorkUnitStatus | `internal/modules/workunit/models.go` | `internal/domain/types.go` |
| Agent, AgentRuntimeType, AgentStatus | `internal/modules/agent/models.go` | `internal/domain/types.go` |
| AgentSession, AgentSessionStatus | `internal/modules/agentsession/models.go` | `internal/domain/types.go` |
| TaskGraph, TaskGraphStatus | `internal/modules/taskgraph/models.go` | `internal/domain/types.go` |
| PromptFragment, PromptSnapshot, ToolsetSnapshot, ComposedPrompt | `internal/modules/prompt/models.go` | `internal/domain/types.go` |
| Review, ReviewStatus, ReviewValidationGate | `internal/modules/review/models.go` | `internal/domain/types.go` |
| Trigger, TriggerStatus, TriggerType, AnomalyType, ResolutionAction | `internal/modules/trigger/models.go` | `internal/domain/types.go` |

**Tipos que FICAM nos módulos:**
- Payloads específicos de eventos (ex: `TaskGraphCreatedPayload` se usado apenas internamente)
- Types auxiliares para lógica interna
- Input/Result structs de serviços locais

---

## 3. Consequências

### Positivas

1. **Regras contáveis:** 4 pilares em vez de 83+ regras implícitas.
2. **Testes confiáveis:** 3-4 testes simples que falham com violações reais.
3. **Contexto unificado:** Um LLM lê `internal/domain/types.go` e sabe TODOS os tipos do sistema.
4. **Zero ambiguidade:** "Módulos não importam outros módulos" é impossível de mal-interpretar.
5. **Manutenção simplificada:** Adicionar um novo campo a `Task` requer mudar apenas um arquivo.

### Negativas

1. **`internal/domain/` vira um pacote grande:** ~30+ structs e enums. Isso é aceitável para um monolith; se o projeto crescer para microsserviços, `domain/` pode ser extraído como um módulo Go separado.
2. **Mudança massiva de imports:** ~75 imports precisam ser atualizados. Isso é trabalho de uma task dedicada (T5).
3. **`orchestrator` continua complexo:** Como único módulo que importa outros, ele concentra a complexidade de coordenação. Isso é intencional — melhor um módulo complexo do que 10 módulos com acoplamento escondido.
4. **Perda de "Feature Cohesion":** A premissa da ADR-0022 (contexto isolado por feature) é parcialmente abandonada. Isso é aceitável porque a premissa não se materializou na prática — LLMs ainda precisavam ler múltiplos arquivos para entender um tipo.

---

## 4. Alternativas Consideradas

### Alternativa A: Isolamento Total (Zero Shared Types)

Cada módulo define seus próprios tipos. O `orchestrator` faz tradução entre tipos espelhados.

**Rejeitada:** Criaria um "Deus Orchestrator" com 100+ funções de tradução. Duplicação massiva de structs idênticos.

### Alternativa B: Manter ADR-0022 e "Endurecer" Testes

Adicionar mais testes de comportamento (AST inspection) sem mudar a arquitetura.

**Rejeitada:** A auditoria provou que a arquitetura em si é o problema. Mais testes sobre uma fundação quebrada = mais fachada. O custo de manter 83+ regras é maior que o custo de simplificar.

### Alternativa C: Hexagonal / Ports & Adapters

Introduzir interfaces de porta, adapters, domain services, application services.

**Rejeitada:** Para um time de 1 pessoa + LLMs, Hexagonal é over-engineering. A simplificação proposta (4 pilares) atinge 80% do benefício com 20% do custo.

---

## 5. Plano de Migração

### Fase 1: Documentação (Task T1-T3)
- Criar esta ADR-0030
- Atualizar ADR-0022 (marcar como superseded)
- Simplificar testes de arquitetura
- Atualizar scripts e CI/CD

### Fase 2: Mapeamento (Task T4)
- Inventariar todos os types em `internal/modules/*/models.go`
- Criar `MIGRATION_MAP.md`
- Documentar imports impactados

### Fase 3: Migração (Task T5)
- Mover shared entity types para `internal/domain/types.go`
- Atualizar imports em todos os módulos
- Atualizar `bootstrap/services.go`
- Atualizar `orchestrator/models.go`
- Purificar repositories

### Fase 4: Validação
- `go test ./tests/architecture/...` passa
- `go test ./...` passa
- `go build ./...` passa
- `go vet ./...` passa

---

## 6. Referências

- ADR-0022 (superseded): Arquitetura de Módulos Verticais
- `docs/audit/architecture-reliability-audit-2026-05-21.md`
- `docs/agent/tasks/2026-05-21_architecture-test-suite-hardening/`
- `docs/agent/tasks/2026-05-21_code-refactor-for-architecture-violations/`
