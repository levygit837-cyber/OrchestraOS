# ADR 0023: Hybrid Intelligent Orchestrator Architecture

**Data:** 2026-05-11

**Status:** Decidido

---

## 1. Contexto

O OrchestraOS possui uma infraestrutura solida de execucao de agentes: Event Store, State Machine, Task Graph, Prompt Composer, Runtimes (Fake e Gemini), e servicos de dominio bem definidos. Porem, a camada de **inteligencia de orquestracao** ainda nao existe.

O sistema hoje opera em dois extremos:

1. **Go deterministico** (`OrchestratorService`, `TaskGraphService`, etc.): executa regras rigidas, transiciona estados, valida dependencias, mas nao entende semantica, contexto ou intencao. Ele reage a eventos, mas nao antecipa problemas.
2. **Agentes executores LLM** (`code_worker`, `docs_writer`, etc.): executam trabalho tecnico isolado, mas nao tem visao global do sistema, nao coordenam uns com os outros, e nao tomam decisoes estrategicas de orquestracao.

A analise de gaps (`docs/analysis/architecture/orchestrator-agent-gap-analysis.md`) identificou que sem um "Agente Orquestrador", o sistema e uma plataforma de execucao manual, nao um sistema de orquestracao autonomo.

Foi discutida a possibilidade de tornar o proprio `OrchestratorService` um agente LLM. Isso traria inteligencia, mas a superficie de falha seria enorme: um LLM tomando decisoes de estado, transacoes e seguranca diretamente aumenta latencia, custo e risco operacional.

A arquitetura hibrida proposta separa **decisoes estrategicas** (que usam LLM) de **decisoes taticas** (que usam codigo Go deterministico). Isso reduz custo, latencia e superficie de falha.

## 2. Decisao

O OrchestraOS adotara uma **arquitetura de Orquestracao Hibrida** com dois sistemas distintos e cooperativos:

### 2.1. Sistema de Orquestracao Inteligente (LLM)

Um **Agente Orquestrador Inteligente** (`Intelligent Orchestrator Agent`) que atua como intermediador estrategico.

**Caracteristicas:**
- E um **agente como outro qualquer**, com sua propria `AgentSession`, `Run` e `PromptSnapshot`.
- Nao executa codigo, nao edita arquivos, nao acessa filesystem.
- Nao tem acesso direto ao banco de dados, repositorios ou servicos internos.
- Consome informacao atraves de uma **Observation API** controlada pelo `OrchestratorService`.
- Emite comandos estruturados que o `OrchestratorService` valida antes de executar.

**Responsabilidades estrategicas:**
- Receber e interpretar mensagens em linguagem natural para criar tasks.
- Decompor tasks em work units com semantica e contexto (usando `GeminiPlanner` ou similar).
- Selecionar perfis dinamicos de agente para cada work unit (`code_worker`, `docs_writer`, `reviewer`, `debugger`).
- Diagnosticar stalls, loops e anomalias em execucoes.
- Decidir sobre replanejamento apos falha de work unit.
- Aprovar ou negar ferramentas de risco medio e alto (com base em contexto, nao apenas regras rigidas).
- Sugerir intervencoes em agentes (dicas, pausas, reinicios).
- Escalonar para aprovacao humana quando necessario.

### 2.2. Sistema de Orquestracao Deterministico (Go)

O `OrchestratorService` existente continua como **control plane central e gatekeeper**.

**Caracteristicas:**
- Servico Go deterministico em `internal/modules/orchestrator/service.go`.
- Unica entidade com acesso direto aos servicos de dominio e ao Event Store.
- Responsavel por validar, persistir e executar TODAS as operacoes.
- Reage a eventos estruturados do sistema e a comandos sugeridos pelo Agente Inteligente.

**Responsabilidades taticas:**
- Transicionar estados via state machines existentes.
- Validar dependencias entre work units antes de iniciar execucao.
- Validar conflitos de `owned_paths` via `WorkUnitService`.
- Aplicar timeout e retry conforme `RunService.Timeout()` e `RunService.Retry()`.
- Detectar heartbeat ausente e marcar sessao como desconectada.
- Auto-aprovar ferramentas seguras (`risk: safe`).
- Escalar ferramentas de risco para o Agente Inteligente ou humano.
- Gerenciar o ciclo de vida do WebSocket (reconexao, replay, ack).
- Coordenar comunicacao cross-module (modulos NUNCA conversam diretamente).

### 2.3. Regra de Ouro: Módulos Nunca Conversam Diretamente

**Nenhum modulo do sistema pode comunicar-se diretamente com outro modulo.**

Toda interacao cross-module ocorre exclusivamente atraves do `OrchestratorService`, que:
1. Recebe a solicitacao (de agente, humano, ou sistema).
2. Valida a operacao (permissoes, estados, regras de negocio).
3. Orquestra a chamada aos servicos de dominio necessarios.
4. Persiste o resultado no Event Store.
5. Retorna a resposta controlada.

Isso inclui:
- Agente Executor -> Agente Executor: comunicacao mediada pelo Orchestrator.
- Agente Inteligente -> Servicos de Domínio: **proibido**. Deve passar pelo OrchestratorService.
- Servico de Task -> Servico de Prompt: **proibido**. Deve passar pelo OrchestratorService.
- Qualquer modulo -> Qualquer modulo: **proibido sem intermediacao do Orchestrator**.

### 2.4. Comunicacao entre os dois sistemas

A comunicacao entre o Agente Inteligente e o OrchestratorService segue o mesmo protocolo de eventos usado por todos os agentes:

```text
Agente Inteligente (LLM)
  |
  |-- Eventos estruturados (via WebSocket ou Event Store)
  |   Ex: orchestrator.suggest_replan, orchestrator.suggest_profile,
  |       orchestrator.approve_tool, orchestrator.intervene_run
  |
  v
OrchestratorService (Go)
  |
  |-- Valida comando (estado permitido? perfil existe? tool liberada?)
  |-- Persiste evento de decisao
  |-- Executa via servico de dominio
  |
  v
Servicos de Dominio (Task, Run, Prompt, Policy, etc.)
```

O OrchestratorService pode **rejeitar** uma sugestao do Agente Inteligente se ela violar politicas, estados invalidos ou regras de seguranca. A rejeicao vira evento auditavel.

### 2.5. Observation API

O Agente Inteligente nao le o Event Store brut. O OrchestratorService fornece resumos estruturados:

```go
type Observation struct {
    TaskSummary       TaskStatusSummary
    ActiveRuns        []RunObservation
    PendingApprovals  []ToolRequestSummary
    Anomalies         []AnomalyReport        // detectados pelo Go
    RecentCheckpoints []CheckpointSummary
    MemoryHints       []MemorySummary
    BudgetStatus      BudgetReport
    SystemAlerts      []SystemAlert
}
```

Isso reduz custo de tokens, evita sobrecarga cognitiva do LLM e mantem o controle do Go sobre o que e exposto.

### 2.6. Toolset do Agente Orquestrador

O Agente Inteligente recebe um toolset exclusivo de decisao:

```go
type OrchestratorToolset struct {
    // Analise
    AnalyzeTaskGraph     func(taskID string) TaskGraphAnalysis
    CompareCheckpoints   func(runID string, checkpointIDs []string) DeltaReport
    DetectAnomaly        func(taskID string) AnomalyReport
    QueryEventStore      func(filter EventFilter) EventSummary

    // Decisao (sugestoes que o Go validara)
    SuggestReplan        func(taskID string, reason string) ReplanSuggestion
    SuggestProfile       func(workUnitID string) ProfileSuggestion
    SuggestIntervention  func(runID string, interventionType string) InterventionSuggestion

    // Acoes mediadas
    RequestToolApproval  func(toolRequestID string, decision string, rationale string) ToolDecision
    SendAgentMessage     func(runID string, message string, priority string) MessageResult
    InjectMemory         func(runID string, memoryIDs []string) InjectionResult
    UpdatePolicy         func(scope string, policyChange PolicyDelta) PolicyResult

    // Escalonamento
    EscalateToHuman      func(taskID string, reason string, urgency string) EscalationResult
}
```

**Nenhuma dessas ferramentas executa codigo ou altera estado diretamente.** Elas emitem eventos/sugestoes que o OrchestratorService processa.

## 3. Consequencias

### Positivas

- **Segregacao de responsabilidades clara**: LLM decide estrategicamente; Go executa e protege taticamente.
- **Seguranca operacional**: O Go pode rejeitar decisoes perigosas ou invalidas do LLM.
- **Custo controlado**: O LLM do orquestrador so e ativado quando ha decisoes estrategicas pendentes, nao a cada evento.
- **Auditabilidade total**: Toda decisao do LLM passa pelo Event Store como evento estruturado.
- **Testabilidade**: O fluxo deterministico pode ser testado sem depender do LLM.
- **Isolamento de modulos preservado**: Cross-module continua obrigatoriamente via OrchestratorService.
- **Autonomia progressiva**: Podemos aumentar gradualmente o que o Agente Inteligente decide, sem mudar a arquitetura.

### Negativas / Riscos

- **Latencia adicional**: Decisoes estrategicas exigem round-trip com LLM.
- **Complexidade de integracao**: Dois sistemas precisam se coordenar sem acoplamento direto.
- **Custo de LLM do orquestrador**: Cada decisao estrategica consome tokens.
- **Dependencia da qualidade do Observation API**: Se o resumo for ruim, o LLM decidira mal.
- **Risco de loop**: Agente Inteligente sugere, Go aplica, novo estado gera nova sugestao — precisa de limites.

## 4. Alternativas Consideradas

### Alternativa A: Orchestrator como agente LLM puro
- **Descricao**: O proprio `OrchestratorService` seria um agente LLM que toma todas as decisoes.
- **Problema**: Superficie de falha enorme. LLM nao deve transicionar estados, gerenciar transacoes ou aplicar seguranca diretamente. Custo e latencia seriam proibitivos para operacoes taticas.
- **Veredicto**: Rejeitado.

### Alternativa B: Orchestrator como script CLI
- **Descricao**: Orquestracao via scripts de CLI sequenciais.
- **Problema**: Fragil, dificil de testar, sem auditoria estruturada, sem capacidade de decisao inteligente.
- **Veredicto**: Rejeitado.

### Alternativa C: Workflow engine externa (Temporal, etc.)
- **Descricao**: Usar uma engine de workflow para orquestrar agentes.
- **Problema**: Adiciona dependencia pesada, perde flexibilidade para decisoes inteligentes, aumenta complexidade operacional.
- **Veredicto**: Rejeitado para o MVP.

### Alternativa D: Dois sistemas completamente separados
- **Descricao**: LLM e Go como servicos independentes se comunicando via fila.
- **Problema**: Adiciona complexidade de rede, deployment e consistencia. Para 1-5 agentes locais, e over-engineering.
- **Veredicto**: Rejeitado. Os dois sistemas coexistirao no mesmo processo Go, com o LLM como runtime dentro do `OrchestratorService`.

### Alternativa E: Arquitetura hibrida (Escolhida)
- **Descricao**: Go deterministico como control plane e gatekeeper; LLM como agente cliente sugestor.
- **Beneficio**: Junta robustez do Go com inteligencia do LLM, mantendo seguranca, custo controlado e arquitetura evolutiva.
- **Veredicto**: Aprovado.

## 5. Referencias

- `docs/analysis/architecture/orchestrator-agent-gap-analysis.md`
- `docs/adr/0002-orchestrator-control-plane.md`
- `docs/adr/0014-orchestration-services.md`
- `docs/architecture/orchestration.md`
- `docs/architecture/protocols/communication-protocol.md`
- `docs/architecture/agents/intelligent-orchestrator-agent.md`
- `docs/architecture/observability/orchestrator-observation-api.md`
- `docs/architecture/protocols/orchestrator-intervention-protocol.md`
- `docs/architecture/agents/multi-agent-coordination.md`
