# Proposta: Sistema de Comunicação e Orquestração entre Agentes

> Data: 2026-05-17
> Contexto: Avaliação de arquitetura para comunicação cross-agent entre WindSurf Cascade e Kimi Code CLI
> Baseado em: Projeto WindAgent (engenharia reversa do Windsurf) + Pesquisa de protocolos ACP/A2A/MCP

---

## 1. ENTENDIMENTO DO QUE JÁ EXISTE

### 1.1 WindSurf (Projeto WindAgent)

O projeto WindAgent em `/home/levybonito/wind` já mapeou profundamente o WindSurf:

**Arquitetura do WindSurf:**
```
Usuário/IDE → Language Server (local, 127.0.0.1:PORTA) → Cascade Planner → API Codeium
```

**Endpoints descobertos no Language Server:**
| Endpoint | Status | Função |
|----------|--------|--------|
| `StartCascade` | ✅ | Inicia sessão agêntica |
| `SendUserCascadeMessage` | ✅ | Envia mensagem |
| `GetCascadeTrajectory` | ✅ | Obtém histórico/estado da sessão |
| `InitializeCascadePanelState` | ✅ | Inicializa workspace |
| `Heartbeat` | ✅ | Health check |
| `GetUserStatus` | ✅ | Modelos disponíveis |
| `GetCompletions` | 🟡 | Autocomplete (não chat) |
| Chat direto | ❌ | **NÃO EXISTE** |

**Protocolo:** Connect Protocol (HTTP/1.1 + JSON) com polling adaptativo. **NÃO é SSE nativo.**

**CSRF Token:** Extraído do ambiente do processo `language_server` via `/proc/PID/environ`.

**Autenticação:**
- Token `sk-ws-*` é específico do Codeium (não funciona na API Moonshot direta)
- O Language Server gera tokens JWT internos dinamicamente (não extraídos com sucesso)
- O servidor remoto `server.codeium.com` requer auth interna

**Tool Calls mapeadas:** `read_file`, `write_to_file`, `edit`, `multi_edit`, `grep_search`, `find_by_name`, `list_dir`, `code_search`, `run_command`, `ask_user_question`, `skill`

**Estrutura do Trajectory (para leitura em tempo real):**
```json
{
  "status": "CASCADE_RUN_STATUS_RUNNING",
  "trajectory": {
    "steps": [
      { "type": "CORTEX_STEP_TYPE_PLANNER_RESPONSE", "plannerResponse": { ... } },
      { "type": "CORTEX_STEP_TYPE_LIST_DIRECTORY", "status": "CORTEX_STEP_STATUS_DONE", ... },
      { "type": "CORTEX_STEP_TYPE_VIEW_FILE", ... }
    ]
  }
}
```

### 1.2 Kimi Code CLI

**Protocolo ACP (Agent Client Protocol):**
- Comunicação JSON-RPC 2.0 sobre stdio
- Comando: `kimi acp`
- Suporta streaming via SSE no modo API
- Tem sessões persistentes
- Suporta MCP servers nativamente (`kimi mcp`)
- Arquitetura aberta (diferente do WindSurf fechado)

---

## 2. ANÁLISE: TIPOS DE COMUNICAÇÃO POSSÍVEIS

### 2.1 Opção A: Leitura Direta dos Apps (Engenharia Reversa)

**WindSurf:**
- ✅ Possível: polling no `GetCascadeTrajectory` para ver estado em tempo real
- ✅ Possível: enviar mensagens via `SendUserCascadeMessage`
- ❌ Limitação: cada mensagem cria um NOVO cascade (não é possível continuar uma sessão existente do IDE facilmente)
- ❌ Limitação: não há SSE nativo - é polling
- ❌ Limitação: CSRF token muda quando WindSurf reinicia
- ❌ Limitação: sistema fechado, pode quebrar com updates

**Kimi Code:**
- ✅ Possível: modo `kimi acp` expõe JSON-RPC sobre stdio
- ✅ Possível: sessões persistentes via `session_id`
- ✅ Streaming nativo suportado
- ✅ Arquitetura aberta, estável

### 2.2 Opção B: Canal de Comunicação Universal (Message Bus)

**Conceito:** Um serviço intermediário onde todos os agentes publicam e consomem mensagens.

```
┌─────────────┐     ┌─────────────────────────┐     ┌─────────────┐
│  WindSurf   │────▶│   Agent Message Bus     │◀────│  Kimi Code  │
│   (IDE)     │◀────│  (Canal Universal)      │────▶│   (CLI)     │
└─────────────┘     └─────────────────────────┘     └─────────────┘
                           │
                    ┌──────┴──────┐
                    │  Dashboard  │
                    │  Web/SSE    │
                    └─────────────┘
```

**Vantagens:**
- Desacopla os agentes (não depende de APIs internas)
- Todos os agentes "falam a mesma língua"
- Permite auditoria, logging, persistência
- Funciona mesmo se um agente reiniciar
- Permite orquestração (um agente delega para outro)

**Desvantagens:**
- Requer um serviço rodando (backend)
- Overhead de arquitetura
- Latência adicional

---

## 3. RECOMENDAÇÃO: ARQUITETURA HÍBRIDA

### 3.1 A Alternativa Ideal

> **Canal de Comunicação Universal (Message Bus) + Adapters por Agente**

**Por quê:**
1. O WindSurf é um sistema **fechado** - engenharia reversa é frágil
2. O Kimi Code é **aberto** via ACP - fácil de integrar
3. Você quer comunicação **assíncrona** - ideal para message bus
4. Você quer orquestração entre branches - precisa de persistência
5. Você quer ver em **tempo real** - SSE do message bus é perfeito

### 3.2 Arquitetura Proposta

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         AGENT COMMUNICATION HUB                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐             │
│  │   Adapter    │    │   Adapter    │    │   Adapter    │             │
│  │  WindSurf    │    │  Kimi Code   │    │   Future...  │             │
│  │  (polling)   │    │   (ACP)      │    │              │             │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘             │
│         │                   │                   │                       │
│         └───────────────────┼───────────────────┘                       │
│                             ▼                                           │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │              Message Bus (Event-Driven)                      │     │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │     │
│  │  │  chat   │ │  tasks  │ │ progress│ │  tools  │           │     │
│  │  │ channel │ │ channel │ │ channel │ │ channel │           │     │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘           │     │
│  └──────────────────────┬──────────────────────────────────────┘     │
│                         │                                               │
│                         ▼                                               │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │                    SSE Streaming Layer                        │     │
│  │       (tempo real para observadores/dashboard)                │     │
│  └──────────────────────┬──────────────────────────────────────┘     │
│                         │                                               │
│                         ▼                                               │
│  ┌──────────────────────────────────────────────────────────────┐     │
│  │              Persistence Layer (SQLite/Redis)                 │     │
│  │  - mensagens  - estados  - tool calls  - sessões              │     │
│  └──────────────────────────────────────────────────────────────┘     │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. FORMATO DA IMPLEMENTAÇÃO

### 4.1 SKILL vs MCP vs Backend Service

| Critério | SKILL | MCP Server | Backend Service |
|----------|-------|-----------|-----------------|
| **Escopo** | Dentro do agente | Ferramenta externa | Serviço independente |
| **Reutilização** | Por projeto | Qualquer MCP client | Qualquer agente |
| **Complexidade** | Baixa | Média | Alta |
| **Persistência** | Nenhuma | Nenhuma (stateless) | Total |
| **Streaming** | Limitado | Via SSE | Completo |
| **Orquestração** | Não | Não | Sim |

**Recomendação: Backend Service + MCP Bridge**

> O Hub deve ser um **Backend Service** (pode ser o próprio OrchestraOS evoluído) que expõe um **MCP Server** para que agentes possam enviar/receber mensagens como tool calls.

**Motivo:**
- Você precisa de **persistência** (agentes em branches diferentes)
- Você precisa de **orquestração** (delegação de tarefas)
- Você precisa de **streaming em tempo real** (observar agentes)
- MCP sozinho é stateless - não resolve o problema de comunicação assíncrona
- SKILL é limitada ao contexto de um único agente

### 4.2 Proposta de Implementação: `agent-comm-hub`

**Componentes:**

#### 4.2.1 Message Bus (Core)
```
POST /api/v1/messages          # Enviar mensagem
GET  /api/v1/messages          # Listar mensagens (com polling ou SSE)
GET  /api/v1/messages/stream   # SSE - stream de mensagens em tempo real
GET  /api/v1/agents            # Listar agentes online
GET  /api/v1/agents/:id/status # Status de um agente
POST /api/v1/tasks             # Criar tarefa orquestrada
GET  /api/v1/tasks/:id         # Ver progresso da tarefa
```

**Schema de Mensagem:**
```json
{
  "id": "msg-uuid",
  "from": "agent-id|windsurf-session-x|kimi-session-y",
  "to": "agent-id|broadcast|orchestrator",
  "channel": "chat|task|progress|tool_call|thinking",
  "timestamp": "2026-05-17T00:00:00Z",
  "payload": {
    "type": "text|tool_call|tool_result|thinking|status_update|request_help",
    "content": "...",
    "metadata": {
      "branch": "feature/x",
      "workspace": "/path/to/project",
      "toolCalls": [...],
      "progress": 0.75
    }
  }
}
```

#### 4.2.2 WindSurf Adapter
```typescript
class WindSurfAdapter {
  // Entrada: conecta ao Language Server local
  // Saída: publica no Message Bus
  
  async connect() {
    // Detecta porta e CSRF do processo language_server
    // Inicia polling do GetCascadeTrajectory
  }
  
  async publishActivity(cascadeId) {
    // Lê trajectory a cada 500ms
    // Extrai: text, thinking, toolCalls, status
    // Publica no Message Bus como mensagens
  }
  
  async sendMessage(cascadeId, message) {
    // Envia via SendUserCascadeMessage
  }
}
```

#### 4.2.3 Kimi Code Adapter
```typescript
class KimiCodeAdapter {
  // Entrada: conecta via ACP (kimi acp)
  // Saída: publica no Message Bus
  
  async connect() {
    // Spawns `kimi acp` process
    // Comunica via JSON-RPC over stdio
  }
  
  async publishActivity(sessionId) {
    // Lê eventos ACP
    // Publica no Message Bus
  }
}
```

#### 4.2.4 MCP Bridge (para integração nativa)
```
# O Hub expõe um MCP Server com tools:

- `send_message`      → Envia msg para outro agente
- `read_messages`     → Lê mensagens não lidas
- `get_agent_status`  → Ver status de um agente
- `delegate_task`     → Delega tarefa a outro agente
- `subscribe_channel` → Inscreve em canal (retorna SSE URL)
```

---

## 5. COMO OS AGENTES VERÃO PROGRESSO UNS DOS OUTROS

### 5.1 Modelo de Canais

```
# Canal de Chat (conversa livre)
/agent-chat
  - Agente A: "Oi B, você já terminou o refactor do auth?"
  - Agente B: "Terminei! Branch: feature/auth-refactor. Quer que eu faça o merge?"
  - Agente A: "Não, deixa que eu reviso primeiro."

# Canal de Tarefas (orquestração)
/tasks
  - Task #1: "Refatorar auth" → Agente B (status: done)
  - Task #2: "Revisar auth"   → Agente A (status: in_progress)

# Canal de Progresso (streaming em tempo real)
/progress/:agent-id
  - SSE stream com:
    - tool_call: { tool: "read_file", params: {...} }
    - tool_result: { output: "..." }
    - thinking: "Analisando a estrutura..."
    - text: "Encontrei o problema..."

# Canal de Ferramentas (compartilhamento de execução)
/tools
  - Agente B publica: "Executei `npm test` → resultado: 5 falhas"
  - Agente A vê e decide: "Vou corrigir essas falhas na minha branch"
```

### 5.2 Dashboard de Observação

```
┌──────────────────────────────────────────────────────────────┐
│ 👁️ Agent Monitor                                        [live]│
├──────────────────────────────────────────────────────────────┤
│                                                              │
│ 🟢 Agente A (WindSurf)              🟡 Agente B (Kimi)      │
│    Branch: main                        Branch: feature/x     │
│    Status: idle                        Status: working...    │
│    Última: "Revisando PR #42"          Última: "Refatorando  │
│                                          módulo de auth"     │
│                                                              │
│ ┌─ Progresso ao vivo ──────────────────────────────────────┐│
│ │ Agente B:                                                  ││
│ │   🔧 read_file → src/auth/service.ts                       ││
│ │   💭 "Vou extrair a lógica de validação..."               ││
│ │   ✏️  edit → src/auth/service.ts (3 substituições)        ││
│ │   🔧 run_command → npm test (executando...)               ││
│ └──────────────────────────────────────────────────────────┘│
│                                                              │
│ 💬 Chat entre Agentes:                                       │
│ A: Você já viu o erro no teste #3?                          │
│ B: Sim! É porque mudamos a interface. Vou corrigir.         │
│ A: Obrigado, me avisa quando terminar.                      │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## 6. DECISÕES DE ARQUITETURA

### 6.1 Streaming: Polling vs SSE vs WebSocket

| Tecnologia | WindSurf Nativo | Kimi Nativo | Nossa Solução |
|------------|-----------------|-------------|---------------|
| SSE | ❌ Não suporta | ✅ Suporta | ✅ Use para o Hub |
| WebSocket | ❌ Não | ❌ Não | ✅ Opção alternativa |
| Polling | ✅ (obrigatório) | ❌ Não precisa | ✅ Para adaptar WindSurf |
| Long Polling | ❌ Não | ✅ Via ACP | ✅ Fallback |

**Decisão:** O Hub usa **SSE** para clientes. O WindSurf Adapter faz **polling** interno no Language Server e publica no Hub via SSE.

### 6.2 Formato: SKILL ou MCP?

**NÃO é SKILL** porque:
- SKILL é específica do contexto de um agente
- Não permite comunicação cross-agent persistente
- Não tem streaming para observadores externos

**É um Backend Service com MCP Bridge** porque:
- Backend = persistência, orquestração, streaming
- MCP = integração nativa com Kimi Code (e futuros agentes)
- WindSurf Adapter = bridge para o sistema fechado

### 6.3 Onde Rodar

```
Opção 1: Processo local (mesma máquina)
  - Vantagem: Baixa latência, fácil de conectar ao WindSurf LS
  - Desvantagem: Só funciona localmente

Opção 2: Container Docker
  - Vantagem: Portável, pode rodar em qualquer lugar
  - Desvantagem: Precisa de network access para WindSurf LS

Opção 3: Integrado no OrchestraOS
  - Vantagem: Reutiliza infraestrutura existente
  - Desvantagem: Acopla com o sistema

Recomendação: Opção 1 (local daemon) como MVP, evoluindo para Opção 2.
```

---

## 7. PRÓXIMOS PASSOS PRÁTICOS

### MVP 1: Prova de Conceito (1-2 dias)

1. **Criar um serviço Node.js/Go mínimo:**
   - Endpoint SSE: `/stream`
   - Endpoint POST: `/message`
   - Memória in-memory (sem persistência ainda)

2. **Criar WindSurf Adapter:**
   - Reutilizar código do `windsurf-client.ts`
   - Fazer polling do `GetCascadeTrajectory`
   - Publicar eventos no SSE

3. **Criar cliente simples:**
   - Script que conecta ao SSE e imprime mensagens
   - Script que envia mensagens via POST

### MVP 2: Integração com Kimi Code (2-3 dias)

1. **Criar Kimi ACP Adapter:**
   - Spawn `kimi acp` como subprocess
   - Parse JSON-RPC over stdio
   - Bridge para o Message Bus

2. **Adicionar persistência:**
   - SQLite para mensagens
   - Redis para estado em tempo real

### MVP 3: Orquestração (3-5 dias)

1. **Task delegation:**
   - Um agente pode criar uma tarefa para outro
   - Notificações quando tarefa muda de status

2. **Dashboard web:**
   - Interface para ver agentes em tempo real
   - Histórico de conversas
   - Visualização de tool calls

---

## 8. RISCOS E MITIGAÇÕES

| Risco | Probabilidade | Impacto | Mitigação |
|-------|--------------|---------|-----------|
| WindSurf muda API interna | Alta | Alto | Adapter desacoplado; testes automatizados |
| CSRF token expira | Média | Médio | Redetecção automática (já existe no projeto) |
| Polling consome CPU | Média | Baixo | Intervalo adaptativo (500ms-2s) |
| Dois agentes editam mesmos arquivos | Média | Alto | Sistema de locks no Message Bus |
| Kimi ACP muda protocolo | Baixa | Médio | ACP é especificação aberta |

---

## 9. CONCLUSÃO

**A alternativa ideal é um Canal de Comunicação Universal (Message Bus) implementado como Backend Service com MCP Bridge.**

**Por que não apenas engenharia reversa direta:**
- O WindSurf é intencionalmente fechado
- O polling direto é frágil e quebra com updates
- Não resolve o problema de múltiplos agentes em máquinas diferentes

**Por que Message Bus:**
- Desacopla os agentes
- Permite comunicação assíncrona (essencial para branches diferentes)
- Permite observação em tempo real via SSE
- É extensível para novos agentes futuros
- Pode evoluir para orquestração completa

**Implementação sugerida:**
1. Backend Go/Node.js com SSE
2. WindSurf Adapter (reutilizar código existente do projeto)
3. Kimi ACP Adapter
4. MCP Server para integração nativa
5. Dashboard web para observação

---

*Análise baseada no projeto WindAgent existente, documentação do Kimi CLI ACP, e pesquisa de protocolos MCP/A2A/ACP.*
