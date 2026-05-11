# ADR 0021: Agent Service e Registro de Agentes

## Contexto

O OrchestraOS define `Agent` como entidade do dominio (`domain/types.go`), com campos como `ID`, `Name`, `RuntimeType`, `SystemProfile`, `Status`, e `CreatedAt`. A migration de `agents` existe e a entidade e persistida.

No entanto, nao existe um servico de dominio para agentes. A ADR 0017 lista os servicos aprovados como `TaskService`, `RunService`, `WorkUnitService`, `AgentSessionService` e `EventService`. `AgentService` nao aparece nessa lista.

Na pratica atual:

- A CLI `run start` gera um `AgentID` como `"agent-" + uuid.New().String()` inline.
- O `AgentSessionService.Create()` aceita qualquer `AgentID` sem validar existencia do agente.
- Nao existe registro de quais agentes existem, seus perfis, runtime type, capabilities ou status.
- Nao existe logica de match entre `WorkUnit.AssignedAgentProfile` e agente disponivel.

Com a introducao do `OrchestratorService` (ADR 0020), que precisa criar agentes automaticamente para cada work unit, e necessario ter um servico que registre, consulte e gerencie agentes.

## Decisao

O OrchestraOS adicionara `AgentService` a lista de servicos de dominio aprovados.

### Responsabilidades do AgentService

`AgentService` deve coordenar criacao, consulta, atualizacao de status e desativacao de agentes.

No escopo inicial:

- `Create(ctx, input) -> Agent`: cria um agente com nome, `RuntimeType`, `SystemProfile` e persiste. Emite evento `agent.created`.
- `GetByID(ctx, id) -> Agent`: consulta um agente por ID.
- `FindOrCreate(ctx, profile, runtimeType) -> Agent`: busca agente disponivel com o perfil solicitado ou cria um novo. Esta e a interface principal para o `OrchestratorService`.

### Validacao

O `AgentSessionService.Create()` deve passar a validar que o `AgentID` referencia um agente existente no banco. Isso fecha o gap de integridade referencial.

### Perfis de agente

Os perfis validos ja estao definidos no planner (`ValidatePlannerProfile`):

- `code_worker`
- `docs_writer`
- `reviewer`
- `debugger`
- `default`

O `AgentService` deve usar a mesma lista de perfis validos.

### RuntimeType

O `AgentService` deve aceitar os runtime types ja definidos em `agent/runtime.go`:

- `fake`
- `gemini`
- `codex_cli`
- `external`

### Pool e reutilizacao

No primeiro corte, cada work unit cria um agente novo via `FindOrCreate`. Reutilizacao de agentes ociosos entre work units e uma direcao futura que nao entra neste escopo.

### Onde implementar

O `AgentService` deve ser implementado em `internal/services/agent_service.go`, seguindo o padrao dos demais servicos (recebe `*sql.DB`, usa transacoes, emite eventos).

## Consequencias

- O `OrchestratorService` (ADR 0020) pode criar agentes automaticamente para cada work unit.
- A `AgentSession` passa a referenciar agentes reais e consultaveis.
- O sistema ganha registro auditavel de quais agentes existem e qual seu perfil.
- A CLI `run start` deve migrar de gerar `AgentID` inline para usar `AgentService.FindOrCreate()`.
- A validacao de `AgentID` em `AgentSessionService.Create()` adiciona uma restricao que pode quebrar testes existentes que usam IDs arbitrarios. Testes devem ser atualizados para criar agentes via servico.

## Alternativas consideradas

- **Manter agentes como IDs soltos sem registro**: simples, mas impede match por perfil, consulta de historico e validacao de integridade.
- **Criar AgentPool com capacidades avancadas**: util para escala, mas prematuro para o primeiro corte.
- **Embutir logica de agentes no OrchestratorService**: reduziria um servico, mas violaria separacao de responsabilidades e dificultaria reutilizacao.
- **Registrar agentes apenas no Event Store sem projecao**: manteria canoncidade, mas dificultaria consulta rapida por perfil e status.
