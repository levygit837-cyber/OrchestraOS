# ADR 0024: Deprecation of ADR 0017 - Domain Services Layer

**Data:** 2026-05-13

**Status:** Decidido

---

## Contexto

A ADR 0017 (Domain Services for Operational Dependencies) foi aprovada em 2026 para estabelecer uma camada de serviços de domínio em `internal/services/` como fronteira obrigatória para comandos que alteram estado operacional.

No entanto, a ADR 0022 (LLM-Optimized Module Architecture) foi aprovada posteriormente, estabelecendo uma arquitetura de **Módulos Verticais (Vertical Slices)** onde cada entidade de domínio tem seu próprio módulo autônomo em `internal/modules/<entity>/`.

O códigobase foi migrado para a arquitetura de módulos verticais, eliminando a camada `internal/services/`. Os serviços de domínio descritos na ADR 0017 foram redistribuídos:

- `TaskService` → `internal/modules/task/service.go`
- `RunService` → `internal/modules/run/service.go`
- `WorkUnitService` → `internal/modules/workunit/service.go`
- `AgentSessionService` → `internal/modules/agentsession/service.go`
- `EventService` → `internal/core/event/service.go`
- `TaskGraphService` → `internal/modules/taskgraph/service.go`
- `PromptService` → `internal/modules/prompt/service.go`
- `AgentService` → `internal/modules/agent/service.go`
- `ReviewService` → `internal/modules/review/service.go`
- `TriggerService` → `internal/modules/trigger/service.go`
- `OrchestratorService` → `internal/modules/orchestrator/service.go`

## Decisão

A ADR 0017 é marcada como **DEPRECATED**. Os princípios fundamentais da ADR 0017 continuam válidos, mas a implementação mudou:

### Princípios Mantidos (Validos)

- Serviços de domínio são a fronteira obrigatória para comandos que alteram estado
- CLI, TUI, GitHub, runtimes e conectores devem chamar serviços, não repositórios diretamente
- Transições de estado devem passar por serviços com validação e eventos
- Event Store permanece como fonte de verdade operacional
- Retries, timeouts e validação de entrada são responsabilidades dos serviços
- Atomicidade de múltiplas escritas relacionadas deve ser preservada

### Implementação Atual (ADR 0022)

- **Módulos Verticais:** Cada entidade de domínio tem seu próprio módulo autônomo
- **Comunicação Cross-Module:** Módulos NUNCA importam outros módulos diretamente
- **Core/Orchestration:** Helpers cross-domain residem em `internal/core/coordination/`
- **Bootstrap:** Wiring de dependências via `internal/bootstrap/services.go` com adapters
- **Isolamento para LLMs:** Cada módulo pode ser compreendido isoladamente por agentes de IA

## Consequências

- A documentação que referenciava `internal/services/` deve ser atualizada para `internal/modules/*/service.go`
- Referências à ADR 0017 devem ser substituídas por referências à ADR 0022
- O princípio de "serviços como fronteira de comando" permanece válido, mas a estrutura organizacional mudou
- Novos serviços devem seguir o padrão de módulos verticais, não criar uma camada `internal/services/`

## Referências

- ADR 0017: Domain Services for Operational Dependencies (DEPRECATED)
- ADR 0022: LLM-Optimized Module Architecture
- ADR 0020: Orchestrator Service
- ADR 0021: Agent Service
- ADR 0023: Hybrid Intelligent Orchestrator
- docs/architecture/migration-vertical-slices.md
