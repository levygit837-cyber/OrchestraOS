# Análise de Gaps: Fluxo Ideal de Orquestração vs. Estado Atual

> Data: 2026-05-09
> Escopo: Avaliar o estado atual do OrchestraOS contra o fluxo ideal de orquestração de agentes, identificar funcionalidades ausentes e documentar necessidades de implementação.
> Autonomia aprovada: Nível 2 (IA implementa com revisão humana).

---

## 1. Resumo Executivo

O OrchestraOS possui uma base sólida de infraestrutura: Event Store com idempotência, State Machine, Prompt Composition com fragmentos versionados, Task Graph acíclico, e um runtime de inferência real (Gemini). No entanto, **a camada de orquestração inteligente ainda não existe**. O sistema hoje é uma plataforma de execução de agentes onde humanos (ou scripts) orquestram manualmente o fluxo. O "Agente Orquestrador" que recebe mensagens em linguagem natural, decompõe tasks, seleciona perfis, cria agentes e monitora execuções **ainda não foi implementado**.

Este documento mapeia o estado atual, o fluxo ideal, os gaps funcionais, e propõe a estratégia para viabilizar testes de inferência real através dos serviços.

---

## 2. Fluxo Ideal (Visão Alvo)

```text
┌──────────────┐     ┌─────────────────┐     ┌─────────────┐     ┌─────────────────┐
│   Usuário    │────▶│  Orchestrator   │────▶│    Task     │────▶│   Task Graph    │
│ (linguagem   │     │  (Agente LLM)   │     │   Created   │     │  (Decompose)    │
│  natural)    │     │                 │     │             │     │                 │
└──────────────┘     └─────────────────┘     └─────────────┘     └─────────────────┘
                                                                         │
                              ┌────────────────────────────────────────┘
                              ▼
                    ┌─────────────────┐
                    │   WorkUnits     │
                    │ (com perfil de  │
                    │  agente atrib.) │
                    └────────┬────────┘
                             │
         ┌───────────────────┼───────────────────┐
         ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  Agent Executor │ │  Agent Executor │ │  Agent Executor │
│   (code_worker) │ │ (docs_writer)   │ │   (reviewer)    │
│                 │ │                 │ │                 │
│ SystemPrompt    │ │ SystemPrompt    │ │ SystemPrompt    │
│ TaskPrompt      │ │ TaskPrompt      │ │ TaskPrompt      │
│ Toolset         │ │ Toolset         │ │ Toolset         │
│ Run + Session   │ │ Run + Session   │ │ Run + Session   │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             ▼
                    ┌─────────────────┐
                    │  Orchestrator   │
                    │ (Monitoramento, │
                    │  Aprovações,    │
                    │  Replanejamento)│
                    └─────────────────┘
```

### Etapas do Fluxo Ideal

1. **Entrada em Linguagem Natural**: Usuário envia mensagem descrevendo uma necessidade (ex: "Crie uma API de autenticação com JWT e refresh tokens").
2. **Criação de Task pelo Orchestrator**: O Orchestrator interpreta a mensagem, extrai título, descrição, critérios de aceite, prioridade e nível de risco. Cria a Task via `TaskService`.
3. **Decomposição Inteligente**: O Orchestrator analisa a complexidade da task. Se necessário, decompõe em WorkUnits usando LLM (não apenas heurística local). Cada WorkUnit recebe:
   - Objetivo claro
   - Critérios de aceite específicos
   - **Perfil de agente adequado** (code_worker, docs_writer, reviewer, debugger)
   - Dependências explícitas
   - Paths de propriedade e leitura
4. **Criação de Agentes Dinâmicos**: Para cada WorkUnit, o Orchestrator (ou sistema) cria um `Agent` com o perfil correto, monta `SystemPrompt` e `TaskPrompt` via `PromptService`, e seleciona o `Toolset` apropriado.
5. **Spawn de Runs e Sessions**: O Orchestrator agenda a execução respeitando dependências do Task Graph. Para cada WorkUnit pronta:
   - Cria `Run` via `RunService`
   - Cria `AgentSession` via `AgentSessionService`
   - Inicia o runtime (Fake, Gemini, ou futuro CodexCLI)
6. **Execução Isolada**: Cada agente executor trabalha em sua WorkUnit isolada, emitindo eventos (heartbeat, checkpoint, tool_request, completed).
7. **Monitoramento Contínuo**: O Orchestrator observa o stream de eventos. Ele pode:
   - Aprovar/negar ferramentas (tool.approved / tool.denied)
   - Interromper uma execução (message.interrupt)
   - Solicitar replanejamento se uma WorkUnit falhar
   - Reagir a checkpoints para tomar decisões
8. **Validação e Completude**: Ao final, o Orchestrator valida evidências, revisa diffs, e decide se a Task está completa ou se requer retry/replanejamento.

---

## 3. Estado Atual do Sistema

### 3.1 O que já existe e funciona

| Componente | Estado | Detalhes |
|---|---|---|
| **Event Store** | ✅ Completo | Append-only, idempotente, sequenciado, validação JSON Schema, replay de estado |
| **State Machine** | ✅ Completo | Transições validadas para Task, WorkUnit, Run, AgentSession. Replay estrito. |
| **TaskService** | ✅ Completo | CRUD + transições de status. Não decompõe automaticamente. |
| **WorkUnitService** | ✅ Completo | CRUD + transições + validação de dependências e owned_paths. |
| **RunService** | ✅ Completo | CRUD + start/validate/complete/fail/cancel/timeout + retry com backoff. |
| **AgentSessionService** | ✅ Completo | Ciclo de vida completo + heartbeat + checkpoint + recoverable state. |
| **EventService** | ✅ Completo | Wrapper do Event Store. |
| **PromptService** | ✅ Completo | Prepara prompts e snapshots. Recebe perfil da WorkUnit. |
| **Prompt Composer** | ✅ Completo | Monta SystemPrompt + TaskPrompt a partir de fragmentos embedados. |
| **Toolset Selector** | ✅ Completo | Seleção hardcoded por perfil (code_worker, reviewer, debugger, docs_writer, fake). |
| **TaskGraphService** | ✅ Completo | Decomposição heurística local (sem LLM). 2-5 work units por task. |
| **GeminiRuntime** | ✅ Completo | Inferência real multi-turn com function calling. |
| **FakeRuntime** | ✅ Completo | Simulação determinística para testes. |
| **Commander** | ✅ Completo | Transições atômicas de estado com eventos. |
| **JSON Schema / Contratos** | ✅ Completo | Schemas de domínio e protocolo validados. |
| **CLI** | ✅ Completo | Comandos para task, workunit, run, agentsession, event, migrate. |

### 3.2 Como o fluxo funciona hoje (manual)

```text
Usuário roda comandos manualmente:

1. orchestraos task create --title "..." --description "..."
   -> Task criada no estado "created"

2. orchestraos task graph create --task-id <id>
   -> TaskGraphService.Decompose() com heurística local
   -> WorkUnits criadas com AssignedAgentProfile: "default" (→ code_worker)

3. orchestraos run start --workunit-id <id> --runtime fake|gemini
   -> RunService.Create() + RunService.Start()
   -> AgentSessionService.Create() + Connect()
   -> PromptService.PrepareRunPrompt()
   -> runtime.Start()
   -> CLI consome eventos em loop
   -> RunService.Validate() + Complete() manualmente (ou pelo runtime)
```

**Observação crítica:** Todo o fluxo é manual via CLI. Não existe um agente que automatize essas transições.

---

## 4. Mapeamento de Gaps

### Gap 1: Agente Orquestrador (CRÍTICO)

**Status:** ❌ NÃO EXISTE

**Descrição:** Não há nenhuma entidade runtime (nem código Go, nem agente LLM) que desempenhe o papel de "Orchestrator Inteligente". O sistema hoje tem apenas um fragmento de prompt (`prompt.communication.agent_orchestrator`) que instrui agentes executores a "comunicar-se com o orquestrador via eventos estruturados". O "orquestrador" aqui é o backend Go em si, não um agente de IA.

**Impacto:** Sem o Orchestrator Agent, não é possível:
- Receber mensagens em linguagem natural
- Tomar decisões de decomposição baseadas em semântica
- Selecionar perfis de agente dinamicamente
- Monitorar e intervir em execuções automaticamente
- Replanejar tasks sem intervenção humana

**O que precisa ser construído:**
1. Entidade `Orchestrator` (ou uso do GeminiRuntime com perfil especial)
2. Serviço `OrchestratorService` que encapsula a lógica de orquestração
3. Loop de orquestração que observe o Event Store e tome ações
4. Perfil de agente `orchestrator` no catálogo de prompts
5. Toolset do orquestrador (criar task, decompor, criar agente, aprovar tool, etc.)

---

### Gap 2: Decomposição Inteligente de Tasks (CRÍTICO)

**Status:** ⚠️ PARCIAL (heurística local apenas)

**Descrição:** O `TaskGraphService` usa apenas `local_heuristic_v1`, que:
- Conta palavras dos `acceptance_criteria` para calcular peso
- Usa greedy algorithm para agrupar em 2-5 work units
- Respeita apenas dependências explícitas `[after: N]`
- Sempre atribui `AssignedAgentProfile: "default"` (→ code_worker)
- Não entende semântica da task

**Impacto:**
- Não consegue decompor tasks que não têm acceptance criteria pré-formatados
- Não seleciona o perfil correto de agente para cada work unit
- Não estima complexidade ou risco da decomposição
- Não gera objetivos e critérios de aceite customizados por work unit (usa o texto original)

**O que precisa ser construído:**
1. **Planner LLM-based**: Um novo `PlannerStrategy` (ex: `llm_gemini_v1`) que:
   - Recebe título, descrição e contexto da task
   - Usa LLM para gerar work units com objetivos claros
   - Seleciona o perfil de agente adequado para cada work unit
   - Define dependências semânticas (não apenas sintáticas)
   - Gera `ValidationPlan` customizado por work unit
2. **Fallback para heurística**: Se o LLM falhar ou a task for simples, usar a heurística local
3. **Validação do plano LLM**: Verificar ciclos, limites de work units, balanceamento, e schemas

---

### Gap 3: Atribuição Dinâmica de Perfis de Agente (CRÍTICO)

**Status:** ❌ NÃO EXISTE

**Descrição:** Todas as work units criadas pela decomposição usam `AssignedAgentProfile: "default"`. Não há lógica que analise o tipo de trabalho e atribua `docs_writer`, `reviewer`, ou `debugger`.

**Impacto:**
- Documentação técnica é escrita por `code_worker` em vez de `docs_writer`
- Revisões de código não usam o perfil `reviewer` especializado
- Tarefas de debug não usam o perfil `debugger`

**O que precisa ser construído:**
1. No planner LLM: instruir o modelo a selecionar o perfil adequado para cada work unit
2. No `TaskGraphService`: aceitar perfis vindos do planner e validá-los contra `prompting.SelectToolset`
3. Toolset estendido para orquestrador: `orchestrator.assign_profile`

---

### Gap 4: Criação e Spawn Automático de Agentes (CRÍTICO)

**Status:** ❌ NÃO EXISTE

**Descrição:** O `Agent` em `domain/types.go` é uma entidade estática de configuração. Não existe um serviço que "cria agentes dinamicamente" para cada work unit. Hoje, o `AgentSessionService.Create` recebe um `AgentID` qualquer (gerado na CLI como `"agent-"+uuid`), mas não existe `AgentService` que gerencie agentes disponíveis.

**Impacto:**
- Não há pool de agentes
- Não há registro de quais agentes existem, seus perfis e capacidades
- Não há lógica de match entre work unit e agente disponível

**O que precisa ser construído:**
1. **`AgentService`**: CRUD de Agentes com perfil, capabilities, runtime type, status
2. **`AgentPool` / `AgentRegistry`**: Lista de agentes disponíveis e ocupados
3. **Match logic**: Dado uma WorkUnit, selecionar ou criar um Agent compatível
4. **Spawn automático**: Quando uma WorkUnit está pronta para execução, criar automaticamente:
   - `Agent` (se necessário)
   - `Run`
   - `AgentSession`
   - Iniciar o runtime apropriado

---

### Gap 5: Loop de Orquestração e Monitoramento (CRÍTICO)

**Status:** ❌ NÃO EXISTE

**Descrição:** Não existe um processo background ou loop que:
- Observe work units pendentes e as inicie automaticamente
- Monitore eventos de agentes em execução
- Tome decisões de aprovação/negativa de ferramentas
- Reaja a falhas com retry ou replanejamento

**Impacto:**
- O usuário precisa manualmente iniciar cada run
- Não há "live view" automatizado do Orchestrator
- Falhas em work units não disparam ações corretivas automaticamente

**O que precisa ser construído:**
1. **`OrchestratorLoop`**: Goroutine ou processo que:
   - Poll ou listen eventos do Event Store
   - Identifique work units no estado `scheduled` com dependências satisfeitas
   - Crie runs e sessions automaticamente
   - Consuma eventos de runtime e tome decisões
2. **Decision Engine**: Regras para:
   - Auto-aprovar ferramentas seguras (`risk: safe`)
   - Escalar ferramentas de risco para aprovação humana/orquestrador
   - Detectar stalls (agente sem heartbeat, run travada)
   - Decidir retry vs. replanejamento vs. falha
3. **Interrupt system**: Enviar `message.interrupt` para agentes quando necessário

---

### Gap 6: Sandbox Manager (MÉDIO)

**Status:** ❌ NÃO EXISTE (planejado para M6 no roadmap)

**Descrição:** Não há gerenciamento de sandbox. Runs não criam branches, worktrees, ou containers isolados. O `GeminiRuntime` opera sem sandbox de filesystem.

**Impacto:**
- Agentes não têm isolamento real
- Não há diff coletado automaticamente
- Não há garantia de que um agente não modifica arquivos fora dos seus `OwnedPaths`

**O que precisa ser construído:**
1. `SandboxService` que cria branch + worktree por work unit
2. Isolamento de filesystem (mesmo que via diretórios)
3. Coleta de diff ao final da run

---

### Gap 7: Policy Engine para Ferramentas (MÉDIO)

**Status:** ❌ NÃO EXISTE (planejado para M7 no roadmap)

**Descrição:** Não há engine que classifique requisições de ferramentas e decida aprovação automática ou necessidade de intervenção.

**Impacto:**
- Todas as `tool.approved` são automáticas (o CLI aprova implicitamente)
- Não há proteção contra ações destrutivas não autorizadas

**O que precisa ser construído:**
1. `PolicyEngine` que avalie cada `tool.requested` contra políticas
2. Auto-aprovação para tools `risk: safe`
3. Queue de aprovação para tools `risk: sensitive`
4. Registro de decisões como eventos

---

### Gap 8: Testes de Inferência Real através dos Serviços (CRÍTICO)

**Status:** ❌ NÃO EXISTE

**Descrição:**
- `gemini_inference_test.go` testa o `GeminiRuntime` isoladamente (sem serviços)
- Nenhum teste usa `GeminiRuntime` através de `RunService.Start`
- Nenhum teste valida o fluxo completo: `Task` → `TaskGraph` → `WorkUnit` → `Prompt` → `Run` → `AgentSession` → `GeminiRuntime` → eventos → `Complete`

**Impacto:**
- Não temos certeza se os prompts gerados pelo `PromptService` funcionam bem com o `GeminiRuntime`
- Não sabemos se o function calling do Gemini funciona corretamente com o `ToolsetSnapshot`
- Não há validação de que o sistema inteiro funciona com inferência real

**O que precisa ser construído:**
1. Testes de integração E2E com `GeminiRuntime`
2. Testes que validem `PromptService.PrepareRunPrompt` → `GeminiRuntime.Start`
3. Testes de tool calling real através dos serviços
4. Testes que validem a compatibilidade do schema de ferramentas com a API Gemini

---

### Gap 9: Memória e Contexto do Orchestrator (MÉDIO)

**Status:** ❌ NÃO EXISTE (planejado para M11 no roadmap)

**Descrição:** O Orchestrator não tem memória de execuções passadas, decisões anteriores, ou contexto acumulado do projeto.

**Impacto:**
- Cada task é tratada isoladamente
- O Orchestrator não aprende com padrões de decomposição anteriores
- Não há contexto de ADRs, canvas, ou código existente injetado no prompt do orquestrador

**O que precisa ser construído:**
1. `MemoryService` (futuro)
2. Ingestão de ADRs, canvas, e checkpoints
3. Injeção de contexto relevante no prompt do Orchestrator

---

## 5. Análise de Eficiência do Fluxo Ideal

### O fluxo ideal é eficiente para um sistema de orquestração de agentes?

**Resposta: Sim, com ressalvas de implementação.**

#### Pontos Fortes do Fluxo Ideal

1. **Separação clara de responsabilidades**: Orchestrator decide, agentes executam. Isso evita que um único agente LLM acumule muita responsabilidade.
2. **WorkUnits isoladas**: Cada work unit tem seu próprio agente, prompt, toolset e sandbox. Isso permite paralelismo seguro.
3. **Event sourcing**: Toda decisão e ação deixa rastro. O Orchestrator pode reconstruir o estado a qualquer momento.
4. **Prompt composition**: Fragmentos versionados permitem evolução controlada dos prompts por perfil.
5. **Checkpointing**: Agentes registram progresso, permitindo recuperação e observabilidade.

#### Riscos e Ineficiências Potenciais

1. **Latência do Orchestrator LLM**: Se o Orchestrator usar LLM para *cada* decisão (monitorar eventos, aprovar tools, replanejar), o overhead de latência e custo será alto. **Recomendação**: Use regras determinísticas para decisões simples (auto-aprovação de tools seguras, transições de estado óbvias) e reserve LLM para decisões complexas (decomposição, replanejamento, diagnóstico de falha).

2. **Complexidade do loop de orquestração**: Um loop que observa tudo e decide tudo pode se tornar um gargalo. **Recomendação**: Arquitetura híbrida — o backend Go (não-LLM) gerencia transições de estado, filas, e timeouts; o Orchestrator LLM intervém apenas em pontos de decisão estratégica.

3. **Custo de múltiplos agentes LLM**: Se cada work unit spawnar um runtime Gemini separado, o custo de API pode ser alto. **Recomendação**: Para o MVP, aceitar o custo. Para escala futura, considerar batching ou reutilização de sessões.

4. **Supervisão humana (Nível 2)**: Como a autonomia aprovada é Nível 2, o fluxo ideal deve incluir gates de aprovação humana antes de ações irreversíveis (merge, push, destruição de dados). O fluxo atual do roadmap já prevê isso (M9: Review e Merge Gate).

#### Recomendação Arquitetural

O fluxo ideal é arquitetonicamente correto, mas a implementação deve ser **híbrida**:

```text
┌─────────────────────────────────────────────────────────────┐
│                    ORQUESTRAÇÃO HÍBRIDA                     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Decisões ESTRATÉGICAS (LLM):                               │
│  - Decompor task em work units                              │
│  - Selecionar perfil de agente                              │
│  - Replanejar após falha                                    │
│  - Diagnosticar stalls anômalos                             │
│                                                             │
│  Decisões TÁTICAS (Código Go determinístico):               │
│  - Transicionar estados (state machine)                     │
│  - Validar dependências                                     │
│  - Auto-aprovar tools seguras                               │
│  - Escalonar work units pendentes para execução             │
│  - Timeout e retry com backoff                              │
│  - Detectar conflitos de owned_paths                        │
│                                                             │
│  Execução (Runtime LLM):                                    │
│  - Agente executor processa work unit                       │
│  - Function calling para tools                              │
│  - Checkpoints e heartbeats                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

Isso preserva a eficiência, reduz custo/latência, e mantém o sistema robusto.

---

## 6. Necessidades de Implementação para Viabilizar Testes com Gemini

Para criarmos testes que validem o fluxo ideal usando inferência real (Gemini), precisamos primeiro construir (ou stubbar) os componentes ausentes. Abaixo, a lista de necessidades em ordem de prioridade:

### Prioridade P0 (Bloqueia testes E2E)

| # | Necessidade | Onde implementar | Complexidade |
|---|-------------|------------------|--------------|
| 1 | **Planner LLM-based** (ou stub que simule decomposição inteligente) | `internal/modules/taskgraph/service.go` (nova strategy) ou serviço separado | Média |
| 2 | **Atribuição de perfil dinâmico** nas work units | `internal/modules/taskgraph/service.go` | Baixa |
| 3 | **AgentService** (registro de agentes) | `internal/modules/agent/service.go` + migration | Baixa |
| 4 | **OrchestratorService** (loop que cria runs/sessions automaticamente) | `internal/modules/orchestrator/service.go` | Alta |
| 5 | **Teste E2E: Task → TaskGraph → WorkUnits → Prompt → GeminiRuntime → Complete** | `tests/integration/gemini_orchestration_test.go` | Alta |

### Prioridade P1 (Melhora significativamente os testes)

| # | Necessidade | Onde implementar | Complexidade |
|---|-------------|------------------|--------------|
| 6 | **Policy Engine mínimo** para auto-aprovação de tools seguras | `internal/modules/policy/service.go` (futuro) | Média |
| 7 | **Sandbox Manager mínimo** (worktree por work unit) | `internal/sandbox/` (novo package) | Média |
| 8 | **Prompt do perfil "orchestrator"** no catálogo | `internal/prompting/catalog/fragments/` | Baixa |
| 9 | **Toolset do orquestrador** | `internal/prompting/toolset.go` | Baixa |

### Prioridade P2 (Futuro próximo)

| # | Necessidade | Onde implementar | Complexidade |
|---|-------------|------------------|--------------|
| 10 | **MemoryService** para contexto do projeto | `internal/memory/` (novo package) | Alta |
| 11 | **WebSocket / Live View** do Orchestrator | `internal/websocket/` ou similar | Alta |
| 12 | **GitHub Connector** | `internal/connectors/github.go` | Média |

---

## 7. Estratégia Sugerida para Testes

Dado que queremos testar o fluxo ideal usando o `GeminiRuntime`, proponho a seguinte estratégia em fases:

### Fase 1: Teste E2E Manual (Sem Orchestrator Automático)

Criar um teste de integração que execute o fluxo completo, mas com passos explícitos no teste (simulando o que o Orchestrator faria):

```go
func TestGeminiEndToEnd_OrchestratorSimulation(t *testing.T) {
    // 1. Criar Task via TaskService
    // 2. Decompor via TaskGraphService (usando heurística local por enquanto)
    // 3. Para cada WorkUnit:
    //    a. Criar Agent com perfil apropriado (ou usar "default")
    //    b. Criar Run via RunService
    //    c. Criar AgentSession via AgentSessionService
    //    d. Preparar Prompt via PromptService
    //    e. Iniciar GeminiRuntime com o prompt preparado
    //    f. Consumir eventos e rotear para serviços
    //    g. Aguardar agent.completed
    //    h. Validar e Completar Run
    // 4. Validar estado final da Task e Event Store
}
```

Este teste valida que o **sistema como um todo funciona com inferência real**, mesmo sem o Orchestrator automático.

### Fase 2: Introduzir Decomposição LLM

Substituir ou complementar o planner heurístico por um planner que use Gemini:

```go
// Novo input para TaskGraphService
type DecomposeTaskGraphInput struct {
    TaskID         string
    PlannerStrategy string // "local_heuristic_v1" | "llm_gemini_v1"
    // ...
}
```

Criar teste que valide a decomposição LLM com tasks reais.

### Fase 3: Orchestrator Loop Automático

Implementar o `OrchestratorService` e testar o loop automático:

```go
func TestOrchestratorLoop_SingleTask(t *testing.T) {
    // 1. Enviar mensagem para Orchestrator
    // 2. Orchestrator cria task
    // 3. Orchestrator decompõe
    // 4. Orchestrator spawna agentes e inicia runs
    // 5. Orchestrator monitora e completa
    // 6. Validar estado final
}
```

---

## 8. Conclusão

O OrchestraOS possui uma **infraestrutura excepcionalmente sólida** para orquestração de agentes. O Event Store, State Machine, Prompt Composer, e Runtime Gemini são componentes maduros e bem testados isoladamente.

O gap central é a **camada de inteligência de orquestração**: não existe um Agente Orquestrador que receba linguagem natural, decompõe tasks com LLM, selecione perfis, crie agentes dinamicamente, e monitore execuções. O sistema hoje é uma plataforma de execução manual (via CLI) de agentes previamente configurados.

**Para viabilizar os testes solicitados com Gemini**, a menor mudança suficiente seria:
1. Criar um teste E2E que simule o Orchestrator manualmente (executando o fluxo passo a passo no teste)
2. Introduzir atribuição de perfil nas work units (mesmo que heurística simples)
3. Registrar agentes explicitamente no banco (`AgentService` mínimo)
4. Executar o `GeminiRuntime` através de `RunService.Start` + `AgentSessionService` + `PromptService`

Isso nos daria **imediatamente** a capacidade de validar se o sistema inteiro funciona com inferência real, sem depender da construção completa do Orchestrator automático.

A construção do Orchestrator automático é um projeto maior (múltiplas milestones do roadmap) e deve seguir o plano de integração formal do projeto.

---

## 9. Referências

- `docs/canvas/project-canvas.md` — Visão e premissas do projeto
- `docs/implementation/roadmap.md` — Milestones planejadas
- `internal/modules/taskgraph/service.go` — Decomposição atual (heurística)
- `internal/modules/prompt/service.go` — Montagem de prompts
- `internal/agent/gemini_runtime.go` — Runtime de inferência
- `tests/integration/services_test.go` — Testes de integração existentes
- `tests/integration/fake_runtime_test.go` — Testes com FakeRuntime
