# ADR 0020: Serviços de Orquestração — OrchestratorService e AgentService

**Status:** Consolidated (absorve: ADR 0021)  
**Data original:** 2026-05-12  
**Última atualização:** 2026-05-17

## Contexto

O OrchestraOS possui todos os servicos de dominio necessarios para executar o fluxo de orquestracao: `TaskService`, `TaskGraphService`, `PromptService`, `RunService`, `AgentSessionService`, `EventService`, `WorkUnitService`. Possui tambem dois runtimes implementados (`FakeRuntime`, `GeminiRuntime`) e um Prompt Composer com fragmentos versionados e selecao de toolset por perfil.

No entanto, nenhum componente conecta esses servicos em um fluxo automatizado. O fluxo completo hoje requer intervencao manual via CLI em cada passo:

1. `task create`
2. `task graph create`
3. `run start`

As ADRs 0002, 0006 e 0017 definem que o Orchestrator deve ser o control plane central, mediando decomposicao, spawn de agentes, monitoramento de eventos, aprovacao de ferramentas e replanejamento. Mas nenhuma ADR define como o Orchestrator deve ser implementado.

A analise de gaps (`docs/analysis/architecture/orchestrator-agent-gap-analysis.md`) identifica esse como o gap mais critico do sistema: sem um loop de orquestracao, o OrchestraOS e uma plataforma de execucao manual, nao um sistema de orquestracao de agentes.

A arquitetura hibrida proposta pela analise recomenda separar decisoes estrategicas (que usam LLM) de decisoes taticas (que usam codigo Go deterministico). Isso reduz custo, latencia e superficie de falha.

## Decisao

O OrchestraOS implementara um `OrchestratorService` como servico interno de dominio com um loop de orquestracao.

### Responsabilidades do OrchestratorService

O `OrchestratorService` coordena o fluxo operacional completo de uma task, desde a decomposicao ate a conclusao. Ele nao e um agente LLM; e um servico Go que orquestra chamadas aos servicos de dominio existentes.

Responsabilidades iniciais:

- Receber uma task e executar o fluxo completo de decomposicao, preparacao e execucao.
- Chamar `TaskGraphService.Decompose()` com a planner strategy configurada.
- Para cada work unit pronta (dependencias satisfeitas, paths disponiveis): criar `Run`, criar `AgentSession`, preparar prompt via `PromptService`, iniciar runtime.
- Consumir eventos do runtime via relay (conforme ADR 0019) e rotear para servicos de dominio.
- Detectar conclusao, falha ou timeout de work units e tomar acoes correspondentes.
- Registrar todas as decisoes como eventos auditaveis.

### Metodo principal: RunTask

```
OrchestratorService.RunTask(ctx, taskID, options) -> OrchestratorResult
```

Este metodo deve:

1. Obter a task via `TaskService`.
2. Decompor via `TaskGraphService.Decompose()`.
3. Resolver a ordem de execucao respeitando dependencias do DAG.
4. Para cada work unit executavel:
   a. Criar `Run` via `RunService.Create()`.
   b. Criar `AgentSession` via `AgentSessionService.Create()`.
   c. Preparar prompt via `PromptService.PrepareRunPrompt()`.
   d. Instanciar runtime conforme `options.RuntimeType` (fake, gemini).
   e. Iniciar runtime e relay de eventos.
5. Aguardar conclusao de cada work unit ou falha.
6. Quando todas as work units estiverem completas, marcar a task como completa.

### Decisoes taticas (Go deterministico)

O `OrchestratorService` deve tomar decisoes deterministicas para:

- Transicionar estados via state machines existentes.
- Validar dependencias entre work units antes de iniciar execucao.
- Validar conflitos de `owned_paths` via `WorkUnitService`.
- Aplicar timeout e retry conforme `RunService.Timeout()` e `RunService.Retry()`.
- Detectar heartbeat ausente e marcar sessao como desconectada.

### Decisoes estrategicas (LLM futuro)

O `OrchestratorService` nao usara LLM no primeiro corte. Decisoes que futuramente poderao usar LLM incluem:

- Replanejamento apos falha de work unit.
- Diagnostico de stalls anomalos.
- Ajuste dinamico de perfis de agente.
- Interpretacao de mensagem em linguagem natural para criacao de task.

Essas capacidades devem entrar depois do fluxo deterministico estar validado.

### Paralelismo

No primeiro corte, o `OrchestratorService` executara work units sequencialmente na ordem topologica do DAG. Work units sem dependencias entre si poderao ser executadas em paralelo em corte posterior.

O limite inicial de paralelismo e 1 agente por vez. Paralelismo real (2-5 agentes conforme canvas) entra depois de sandbox e policy engine estarem funcionando.

### Onde implementar

O `OrchestratorService` deve ser implementado em `internal/modules/orchestrator/service.go`, seguindo o padrao dos demais servicos de dominio (recebe `*sql.DB`, usa transacoes, emite eventos).

**Nota de implementação atual:** Conforme ADR 0022 (Vertical Slice Architecture), o OrchestratorService foi implementado em `internal/modules/orchestrator/` com adapters via `internal/bootstrap/services.go` para conectar outros módulos verticais sem dependências diretas entre eles.

## Consequencias

- O sistema passa a ter um fluxo automatizado de ponta a ponta, acessivel por uma unica chamada.
- A CLI pode expor `task run --id <task_id>` como comando que delega ao `OrchestratorService`.
- Testes de integracao E2E podem validar o fluxo completo chamando `OrchestratorService.RunTask()`.
- O `OrchestratorService` concentra a logica de coordenacao, evitando que CLI, runtime ou testes reimplementem o fluxo.
- O primeiro corte sera sequencial e sem LLM. Paralelismo e decisoes inteligentes entram depois.
- A implementacao deve reusar exclusivamente servicos de dominio existentes, nao repositorios diretos.

---

## 2. AgentService e Registro de Agentes

### 2.1 Contexto adicional

O OrchestraOS define `Agent` como entidade do domínio, mas não existia um serviço de domínio para agentes. Na prática, a CLI `run start` gerava um `AgentID` inline sem registro, e `AgentSessionService.Create()` aceitava qualquer `AgentID` sem validar existência.

Com a introdução do `OrchestratorService`, que precisa criar agentes automaticamente para cada work unit, é necessário ter um serviço que registre, consulte e gerencie agentes.

### 2.2 Decisão

O OrchestraOS adicionará `AgentService` à lista de serviços de domínio aprovados.

Responsabilidades iniciais:

- `Create(ctx, input) -> Agent`: cria um agente com nome, `RuntimeType`, `SystemProfile` e persiste. Emite evento `agent.created`.
- `GetByID(ctx, id) -> Agent`: consulta um agente por ID.
- `FindOrCreate(ctx, profile, runtimeType) -> Agent`: busca agente disponível com o perfil solicitado ou cria um novo. Interface principal para o `OrchestratorService`.

Validação: `AgentSessionService.Create()` deve validar que o `AgentID` referencia um agente existente.

Perfis de agente válidos (definidos no planner):

- `code_worker`
- `docs_writer`
- `reviewer`
- `debugger`
- `default`

Runtime types aceitos:

- `fake`
- `gemini`
- `codex_cli`
- `external`

No primeiro corte, cada work unit cria um agente novo via `FindOrCreate`. Reutilização de agentes ociosos é direção futura.

O `AgentService` deve ser implementado em `internal/modules/agent/service.go`.

### 2.3 Consequências (AgentService)

- O `OrchestratorService` pode criar agentes automaticamente para cada work unit.
- A `AgentSession` passa a referenciar agentes reais e consultáveis.
- O sistema ganha registro auditável de quais agentes existem e qual seu perfil.
- A CLI `run start` deve migrar de gerar `AgentID` inline para usar `AgentService.FindOrCreate()`.
- A validação de `AgentID` em `AgentSessionService.Create()` pode quebrar testes existentes. Testes devem ser atualizados para criar agentes via serviço.

### 2.4 Alternativas consideradas (AgentService)

- **Manter agentes como IDs soltos sem registro**: simples, mas impede match por perfil, consulta de histórico e validação de integridade.
- **Criar AgentPool com capacidades avançadas**: útil para escala, mas prematuro para o primeiro corte.
- **Embutir lógica de agentes no OrchestratorService**: reduziria um serviço, mas violaria separação de responsabilidades.
- **Registrar agentes apenas no Event Store sem projeção**: manteria canonicalidade, mas dificultaria consulta rápida por perfil e status.

---

## Apêndice A: Histórico de Evolução

| Data | Evento | ADR Original |
| --- | --- | --- |
| 2026-05-12 | OrchestratorService definido como loop de orquestração | ADR 0020 |
| 2026-05-12 | AgentService definido para registro e match de agentes | ADR 0021 |
| 2026-05-17 | Ambos consolidados neste documento único | — |

## Apêndice B: Alternativas Consideradas (Orquestração)

- **Orchestrator como agente LLM**: flexível e inteligente, mas aumenta custo, latência e superfície de falha antes de validar o fluxo básico.
- **Orchestrator como script CLI**: rápido de implementar, mas fraco para testes, composição e auditoria.
- **Manter orquestração manual**: funciona para protótipo, mas impede validação E2E e adoção real do sistema.
- **Workflow engine externa (Temporal, etc.)**: robusta para durabilidade e retry, mas adiciona dependência pesada antes de validar os fluxos centrais do MVP.
- **Orchestrator distribuído com filas**: escalável, mas prematuro para 1-5 agentes locais.
