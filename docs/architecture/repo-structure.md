# Estrutura do Repositorio

Este documento define a estrutura atual do repositório do OrchestraOS, refletindo a arquitetura de módulos verticais (Vertical Slice Architecture) conforme ADR 0022.

## Decisao

O repositório adota uma arquitetura de **Módulos Verticais** para otimizar o sistema para operação por agentes de IA (LLMs), reduzindo contexto desnecessário e aumentando escalabilidade.

```text
cmd/orchestraos/
internal/
  bootstrap/
  core/
    apperrors/
    db/
    event/
    eventstore/
    transition/
    serialization/
    statemachine/
    transition/
    validation/
  domain/
  migrations/
  modules/
    agent/
    agentsession/
    orchestrator/
    prompt/
    review/
    run/
    task/
    taskgraph/
    trigger/
    workunit/
contracts/
  schemas/
migrations/
tests/
  architecture/
  contracts/
  integration/
docs/
  adr/
  analysis/
  architecture/
  canvas/
  contracts/
  development/
  implementation/
  management/
  slack/
  templates/
```

## Responsabilidades

### cmd/orchestraos/
Entrada da CLI e comandos locais do MVP.

### internal/bootstrap/
Injeção de dependências (DI) e wiring de serviços. Contém factories para criar serviços de domínio com suas dependências configuradas. Também contém adapters que conectam módulos verticais sem dependências diretas entre eles.

### internal/core/
Componentes compartilhados usados por todos os módulos verticais. Não contém lógica de domínio específica.

- `apperrors/`: Tipagem de erros padronizados com código e operação
- `db/`: Helpers de transação (BeginTx, CommitTx, RollbackTx, EnsureRowsAffected)
- `event/`: EventService wrapper do Event Store
- `eventstore/`: Store de eventos com validação schema, append e replay
- `transition/`: Helpers cross-domain (TransitionInput, OperationResult, AppendTransition, AppendServiceEvent)
- `serialization/`: Marshalling genérico de payloads
- `statemachine/`: Regras de transição de estado e replay
- `transition/`: Payload builders para transições
- `validation/`: Validadores genéricos (UUID, texto, priority, risk, runtime)

### internal/domain/
Tipos compartilhados entre módulos que não pertencem a um único módulo vertical. Contém tipos como checkpoint, event payloads, e tipos de domínio compartilhados.

### internal/migrations/
Migrations do banco de dados usando goose.

### internal/modules/
**Módulos Verticais autônomos** conforme ADR 0022. Cada módulo representa uma entidade de domínio e contém toda a lógica relacionada a essa entidade.

- `agent/`: Agent entities, Runtimes (Fake, Gemini, Codex), GeminiPlanner
- `agentsession/`: Ciclo de vida de sessões de agente, heartbeat, checkpoint, timeout
- `orchestrator/`: OrchestratorService - coordena fluxo end-to-end de tasks
- `prompt/`: Prompt Composer, catálogo de fragmentos versionados, snapshots
- `review/`: Gates de revisão e validação
- `run/`: Tentativas de execução, retry, timeout, projeção
- `task/`: Tasks, triagem, planejamento
- `taskgraph/`: Task Graph, decomposição (local + LLM), validação de DAG
- `trigger/`: Detecção de anomalias (stalls, loops)
- `workunit/`: Work units, dependências, ownership, paths

**Regra de Ouro (ADR 0022):** Módulos verticais NUNCA importam outros módulos diretamente. Comunicação cross-module ocorre via `internal/modules/orchestrator/` (camada de orquestração canônica) ou interfaces DI com adapters em `internal/bootstrap/services.go`.

### contracts/
Contratos JSON versionados como artefatos independentes.

### contracts/schemas/
JSON Schemas executáveis, separados por domínio e protocolo.

### migrations/
Migrations SQL incrementais do banco de dados.

### tests/
Validações de contrato e testes de integração sem depender de serviços externos.

- `architecture/`: Testes de contrato de módulos e fronteiras
- `contracts/`: Testes de schemas JSON
- `integration/`: Testes E2E de fluxos completos

### docs/
Fonte de verdade para arquitetura, canvas, ADRs, contratos narrativos e operação.

## Regras

- **Isolamento de Módulos:** Módulos verticais não se importam diretamente. Comunicação via orchestrator/ ou DI.
- **O dominio (internal/domain/) não deve depender de banco, WebSocket, GitHub, Docker ou CLI.**
- JSON Schemas são contratos de borda; tipos Go são o modelo interno.
- Schemas devem rejeitar campos desconhecidos por padrão.
- Novas dependências só devem entrar quando a validação com biblioteca padrão não for suficiente.
- Mudanças arquiteturais relevantes continuam exigindo ADR.
- Mudanças de contrato devem atualizar schemas JSON correspondentes.
- Cada módulo vertical deve ter: README.md, CONTRACTS.md, doc.go, models.go, service.go, repository.go, queries.go, validation.go

## Escopo Atual

O código atual implementa os seguintes módulos verticais:

- **agent/**: AgentService, FakeRuntime, GeminiRuntime, GeminiPlanner
- **agentsession/**: AgentSessionService com ciclo de vida completo
- **orchestrator/**: OrchestratorService coordena fluxo end-to-end
- **prompt/**: PromptService com catálogo de fragmentos versionados
- **review/**: ReviewService para gates de validação
- **run/**: RunService com retry, timeout e validação
- **task/**: TaskService com triagem e planejamento
- **taskgraph/**: TaskGraphService com decomposição heurística e LLM
- **trigger/**: TriggerService para detecção de anomalias
- **workunit/**: WorkUnitService com validação de dependências e owned_paths

## Referências

- ADR 0022: LLM-Optimized Module Architecture
- ADR 0024: Deprecation of ADR 0017
- docs/architecture/module_index.md
- docs/development/CODING_STANDARDS.md
