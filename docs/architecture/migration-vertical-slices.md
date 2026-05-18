# Migração para Vertical Slice Architecture

Este documento descreve a migração da arquitetura do OrchestraOS de uma arquitetura em camadas com `internal/services/` para uma arquitetura de **Módulos Verticais (Vertical Slice Architecture)** conforme ADR 0022.

## Contexto

A ADR 0017 (Domain Services for Operational Dependencies) estabeleceu inicialmente uma camada de serviços de domínio em `internal/services/` como fronteira obrigatória para comandos que alteram estado operacional.

No entanto, a ADR 0022 (LLM-Optimized Module Architecture) foi aprovada posteriormente, estabelecendo uma arquitetura de **Módulos Verticais** para otimizar o sistema para operação por agentes de IA (LLMs). Esta arquitetura reduz contexto desnecessário, carga cognitiva e dependências cross-module.

## O Que Mudou

### Estrutura Antiga (Layered Architecture)

```
internal/
  domain/          # Tipos centrais do domínio
  services/        # Serviços de domínio (TaskService, RunService, etc.)
    task_service.go
    run_service.go
    workunit_service.go
    agentsession_service.go
    event_service.go
    orchestration/  # Helpers de orquestração
  core/            # Componentes compartilhados
```

### Estrutura Atual (Vertical Slice Architecture)

```
internal/
  bootstrap/       # DI e wiring de serviços com adapters
  core/            # Componentes compartilhados
    apperrors/
    db/
    event/
    eventstore/
    transition/     # Helpers cross-domain
    serialization/
    statemachine/
    transition/
    validation/
  domain/          # Tipos compartilhados entre módulos
  modules/         # Módulos verticais autônomos
    agent/          # AgentService, Runtimes, GeminiPlanner
    agentsession/   # AgentSessionService
    orchestrator/   # OrchestratorService
    prompt/         # PromptService
    review/         # ReviewService
    run/            # RunService
    task/           # TaskService
    taskgraph/      # TaskGraphService
    trigger/        # TriggerService
    workunit/       # WorkUnitService
```

## Mapeamento de Serviços

| Serviço Antigo | Localização Nova |
|----------------|------------------|
| `internal/services/task_service.go` | `internal/modules/task/service.go` |
| `internal/services/run_service.go` | `internal/modules/run/service.go` |
| `internal/services/workunit_service.go` | `internal/modules/workunit/service.go` |
| `internal/services/agentsession_service.go` | `internal/modules/agentsession/service.go` |
| `internal/services/event_service.go` | `internal/core/event/service.go` |
| `internal/services/task_graph_service.go` | `internal/modules/taskgraph/service.go` |
| `internal/services/prompt_service.go` | `internal/modules/prompt/service.go` |
| `internal/services/agent_service.go` | `internal/modules/agent/service.go` |
| `internal/services/review_service.go` | `internal/modules/review/service.go` |
| `internal/services/trigger_service.go` | `internal/modules/trigger/service.go` |
| `internal/services/orchestrator_service.go` | `internal/modules/orchestrator/service.go` |

## Regra de Ouro

**Módulos verticais NUNCA importam outros módulos diretamente.**

Comunicação cross-module ocorre exclusivamente via:
1. `internal/core/transition/` - Helpers cross-domain (TransitionInput, OperationResult, AppendTransition, AppendServiceEvent)
2. `internal/bootstrap/services.go` - Adapters que conectam módulos sem dependências diretas

## Benefícios da Migração

### 1. Isolamento para LLMs
- Cada módulo vertical pode ser compreendido isoladamente por agentes de IA
- Redução de contexto desnecessário ao trabalhar em uma entidade específica
- Menor carga cognitiva para LLMs operarem no código

### 2. Escalabilidade
- Módulos podem crescer independentemente sem afetar outros
- Novos módulos podem ser adicionados sem modificar código existente
- Boundaries claras facilitam testes e manutenção

### 3. Separação de Responsabilidades
- Cada módulo contém toda a lógica relacionada à sua entidade
- Models, Service, Repository, Validation e Queries co-localizados
- Fronteiras claras entre domínios

### 4. Testabilidade
- Cada módulo pode ser testado isoladamente
- Mocks e stubs são mais fáceis de criar
- Testes de integração podem focar em interações via core/transition

## Como Funciona a Comunicação Cross-Module

### Exemplo: OrchestratorService precisa chamar TaskService

**Antigo (import direto):**
```go
import "internal/services/task_service"

func (s *OrchestratorService) RunTask(...) {
    task, err := s.taskService.GetTask(...)
}
```

**Atual (via adapter):**
```go
// internal/modules/orchestrator/service.go
type OrchestratorService struct {
    taskAdapter TaskAdapter  // Interface definida localmente
}

// internal/bootstrap/services.go
func wireOrchestratorService(taskService *task.Service) *orchestrator.Service {
    taskAdapter := &TaskAdapterImpl{taskService: taskService}
    return orchestrator.NewService(taskAdapter, ...)
}
```

## Padrão de Módulo Vertical

Cada módulo vertical contém:
- `README.md` - Visão geral e contrato
- `CONTRACTS.md` - Invariantes e regras de negócio
- `doc.go` - Documentação do pacote
- `models.go` - Tipos e entidades
- `service.go` - Serviço de domínio
- `repository.go` - Acesso a dados
- `queries.go` - Queries SQL
- `validation.go` - Validadores

## Referências

- ADR 0022: LLM-Optimized Module Architecture
- ADR 0024: Deprecation of ADR 0017 - Domain Services Layer
- docs/architecture/repo-structure.md
- docs/architecture/module_index.md
- docs/development/CODING_STANDARDS.md
