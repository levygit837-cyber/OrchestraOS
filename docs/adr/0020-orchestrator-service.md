# ADR 0020: Orchestrator Service e Loop de Orquestracao

## Contexto

O OrchestraOS possui todos os servicos de dominio necessarios para executar o fluxo de orquestracao: `TaskService`, `TaskGraphService`, `PromptService`, `RunService`, `AgentSessionService`, `EventService`, `WorkUnitService`. Possui tambem dois runtimes implementados (`FakeRuntime`, `GeminiRuntime`) e um Prompt Composer com fragmentos versionados e selecao de toolset por perfil.

No entanto, nenhum componente conecta esses servicos em um fluxo automatizado. O fluxo completo hoje requer intervencao manual via CLI em cada passo:

1. `task create`
2. `task graph create`
3. `run start`

As ADRs 0002, 0006 e 0017 definem que o Orchestrator deve ser o control plane central, mediando decomposicao, spawn de agentes, monitoramento de eventos, aprovacao de ferramentas e replanejamento. Mas nenhuma ADR define como o Orchestrator deve ser implementado.

A analise de gaps (`docs/analysis/orchestrator-agent-gap-analysis.md`) identifica esse como o gap mais critico do sistema: sem um loop de orquestracao, o OrchestraOS e uma plataforma de execucao manual, nao um sistema de orquestracao de agentes.

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

O `OrchestratorService` deve ser implementado em `internal/services/orchestrator_service.go`, seguindo o padrao dos demais servicos de dominio (recebe `*sql.DB`, usa transacoes, emite eventos).

## Consequencias

- O sistema passa a ter um fluxo automatizado de ponta a ponta, acessivel por uma unica chamada.
- A CLI pode expor `task run --id <task_id>` como comando que delega ao `OrchestratorService`.
- Testes de integracao E2E podem validar o fluxo completo chamando `OrchestratorService.RunTask()`.
- O `OrchestratorService` concentra a logica de coordenacao, evitando que CLI, runtime ou testes reimplementem o fluxo.
- O primeiro corte sera sequencial e sem LLM. Paralelismo e decisoes inteligentes entram depois.
- A implementacao deve reusar exclusivamente servicos de dominio existentes, nao repositorios diretos.

## Alternativas consideradas

- **Orchestrator como agente LLM**: flexivel e inteligente, mas aumenta custo, latencia e superficie de falha antes de validar o fluxo basico.
- **Orchestrator como script CLI**: rapido de implementar, mas fraco para testes, composicao e auditoria.
- **Manter orquestracao manual**: funciona para prototipo, mas impede validacao E2E e adocao real do sistema.
- **Workflow engine externa (Temporal, etc.)**: robusta para durabilidade e retry, mas adiciona dependencia pesada antes de validar os fluxos centrais do MVP.
- **Orchestrator distribuido com filas**: escalavel, mas prematuro para 1-5 agentes locais.
