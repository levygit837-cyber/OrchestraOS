# 🔬 Engenharia Reversa WindSurf: Descobertas sobre Streaming

> Data: 2026-05-17
> Analista: Kimi Code (investigação ativa)
> Status: **DESCobertAS REVOLUCIONÁRIAS**

---

## 🎯 RESUMO EXECUTIVO

Após investigação ativa no sistema WindSurf em execução, descobrimos que **o WindSurf TEM streaming nativo**, mas o projeto WindAgent anterior não o encontrou porque:

1. O streaming usa **Connect Protocol** (Buf) com envelope proprietário
2. Existem **DOIS serviços** rodando (porta 34665 e 43279)
3. Múltiplos endpoints de streaming existem no binary mas não foram testados
4. O Kimi Code roda **nativamente dentro do WindSurf** via ACP
5. O Devin ACP também roda dentro do WindSurf

---

## 🔍 ARQUITETURA DO WINDSURF (Descoberta Atual)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         WINDSURF PROCESS ARCHITECTURE                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────┐     ┌─────────────────────┐     ┌───────────────┐ │
│  │   WindSurf IDE      │◀───▶│ Extension Server    │◀───▶│ Language      │ │
│  │   (Electron/VSCode) │     │ Porta 40819         │     │ Server        │ │
│  │   PID: 4360         │     │ (VSCode Extension   │     │ PID: 6219     │ │
│  │                     │     │  Host Bridge)       │     │               │ │
│  └─────────────────────┘     └─────────────────────┘     │  ┌─────────┐  │ │
│                                                          │  │Porta    │  │ │
│  ┌─────────────────────┐                                 │  │34665    │  │ │
│  │  Devin ACP          │                                 │  │(LSP     │  │ │
│  │  /usr/share/...     │                                 │  │Principal│  │ │
│  │  devin acp          │                                 │  └─────────┘  │ │
│  └─────────────────────┘                                 │  ┌─────────┐  │ │
│                                                          │  │Porta    │  │ │
│  ┌─────────────────────┐                                 │  │43279    │  │ │
│  │  Kimi Code          │                                 │  │(Cascade │  │ │
│  │  (dentro do IDE)    │                                 │  │Service) │  │ │
│  │  kimi-sdk:protocol  │                                 │  └─────────┘  │ │
│  └─────────────────────┘                                 └───────────────┘ │
│                                                                             │
│  Conexões externas:                                                         │
│  • 35.223.238.178:443 → server.codeium.com (inferência)                   │
│  • 34.49.14.144:443   → server.self-serve.windsurf.com                    │
│  • 104.18.20.246:443  → API Moonshot (Kimi Code standalone)               │
│  • wss://app.devin.ai → Devin Cloud WebSocket                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 📡 DOIS SERVIÇOS NO LANGUAGE SERVER

| Porta | FD | Função | Heartbeat | Cascade |
|-------|-----|--------|-----------|---------|
| **34665** | fd=3 | **LSP Principal** | ✅ `{"lastExtensionHeartbeat":"..."}` | ✅ Com API Key |
| **43279** | fd=13 | **Serviço Secundário** | ❌ Vazio | ❌ Vazio |

**Conclusão:** O serviço Cascade está na porta **34665** (não na 43279 como o projeto WindAgent assumia). A porta 43279 pode ser um serviço interno ou pode ter mudado de propósito em versões mais recentes.

---

## 🌊 ENDPOINTS DE STREAMING DESCOBERTOS

### Confirmados no Binary (strings)

| Endpoint | Tipo | Status |
|----------|------|--------|
| `StreamCascadeReactiveUpdates` | Server Stream | ⚠️ Requer Connect Protocol envelope |
| `StreamCascadeSummariesReactiveUpdates` | Server Stream | Não testado |
| `CreateTrajectoryShareStream` | Bidirectional Stream | Não testado |
| `HandleStreamingCommand` / `HandleStreamingCommandStream` | Stream | Não testado |
| `HandleStreamingTab` / `HandleStreamingTabV2` | Server Stream | Não testado |
| `StreamTerminalShellCommand` | Server Stream | Não testado |
| `GetStreamingCompletions` | Server Stream | Não testado |
| `GetDevstralStream` | Server Stream | Não testado |

### Teste Real: `StreamCascadeReactiveUpdates`

**Request:**
```bash
POST /exa.language_server_pb.LanguageServerService/StreamCascadeReactiveUpdates
Content-Type: application/connect+json
Connect-Protocol-Version: 1
x-codeium-csrf-token: <token>

# Body requer envelope Connect Protocol de 5 bytes:
# Byte 0: flags (compression + end-of-stream)
# Bytes 1-4: tamanho da mensagem (big-endian uint32)
# Depois: JSON payload
```

**Resultados dos testes:**

| Flags Enviado | Resposta do Servidor |
|---------------|---------------------|
| `0x00` | `unsupported protocol version 0 (only 1 is supported)` |
| `0x01` | `sent compressed message without compression support` |
| `0x02` | `unmarshal end stream message: json: cannot unmarshal string...` |
| `0x80` | `invalid envelope flags 128` |

**Interpretação:** O servidor usa um formato de envelope proprietário que NÃO é o Connect Protocol padrão da Buf. Ele requer:
- Protocol version = 1 (codificado de forma diferente)
- Sem compressão
- Formato de envelope específico do WindSurf

---

## ✅ ENDPOINTS NÃO-STREAMING QUE FUNCIONAM

### `GetCascadeTrajectorySteps` (NOVO!)
```bash
POST /exa.language_server_pb.LanguageServerService/GetCascadeTrajectorySteps
```

**Retorna:** Apenas os `steps` do trajectory (mais leve que `GetCascadeTrajectory` completo)

**Exemplo de resposta:**
```json
{
  "steps": [
    {
      "type": "CORTEX_STEP_TYPE_RETRIEVE_MEMORY",
      "status": "CORTEX_STEP_STATUS_DONE",
      "metadata": { "createdAt": "...", "executionId": "..." }
    },
    {
      "type": "CORTEX_STEP_TYPE_USER_INPUT",
      "status": "CORTEX_STEP_STATUS_DONE",
      "userInput": { "userResponse": "Say hello", "items": [...] }
    },
    ...
  ]
}
```

**Vantagem sobre `GetCascadeTrajectory`:**
- Retorna só os steps (sem metadata extra)
- Possivelmente mais rápido
- Ideal para polling otimizado

### `GetCascadeTrajectory` (já conhecido)
```bash
POST /exa.language_server_pb.LanguageServerService/GetCascadeTrajectory
```

Retorna o objeto completo com `status`, `trajectory.steps`, etc.

---

## 🧩 KIMI CODE DENTRO DO WINDSURF

### Descoberta Extraordinária

O WindSurf tem uma **extensão Kimi Code** que roda nativamente dentro do IDE!

**Logs encontrados:**
```
/home/levybonito/.config/Windsurf/logs/.../output_logging_.../3-Kimi Code.log
```

**Como funciona:**
1. A extensão spawna: `kimi --work-dir <projeto> --wire --no-thinking`
2. Comunicação via **JSON-RPC over stdio** (mesmo protocolo ACP)
3. O Kimi SDK (`kimi-sdk:protocol`) gerencia a comunicação
4. Suporta eventos: `StatusUpdate`, `tool_use`, `thinking`, etc.

**Inicialização:**
```json
{
  "jsonrpc": "2.0",
  "id": "1_...",
  "method": "initialize",
  "params": {
    "protocol_version": "1.7",
    "client": { "name": "kimi-agent-sdk/0.1.8", "version": "0.1.8" },
    "capabilities": { "supports_question": true, "supports_plan_mode": true }
  }
}
```

**Resposta do servidor:**
```json
{
  "protocol_version": "1.10",
  "server": { "name": "Kimi Code CLI", "version": "1.43.0" },
  "slash_commands": [
    { "name": "init", "description": "Analyze codebase and generate AGENTS.md" },
    { "name": "compact", "description": "Compact context" },
    { "name": "clear", "description": "Clear context", "aliases": ["reset"] },
    { "name": "yolo", "description": "Toggle YOLO mode" },
    { "name": "afk", "description": "Toggle afk mode" },
    { "name": "plan", "description": "Toggle plan mode" }
  ]
}
```

**Eventos recebidos do Kimi:**
```json
{
  "jsonrpc": "2.0",
  "method": "event",
  "params": {
    "type": "StatusUpdate",
    "payload": {
      "context_usage": null,
      "token_usage": null,
      "mcp_status": {
        "loading": true,
        "connected": 0,
        "total": 2,
        "tools": 0,
        "servers": [
          { "name": "slack", "status": "connecting", "tools": [] },
          { "name": "aidesigner", "status": "connecting", "tools": [] }
        ]
      }
    }
  }
}
```

**Implicação:** Podemos nos comunicar com o Kimi Code dentro do WindSurf via ACP! E o Kimi Code standalone também roda (PID 7158, 8176).

---

## 🤖 DEVIN ACP DENTRO DO WINDSURF

### Processos encontrados:
```
PID 6307: /usr/share/windsurf/resources/app/extensions/windsurf/devin/bin/devin acp
PID 6308: /usr/share/windsurf/resources/app/extensions/windsurf/devin/bin/devin acp --agent-type summarizer
```

### Conexão Devin Cloud via WebSocket:
```
wss://app.devin.ai/api/acp/live
```

### Logs:
```
/home/levybonito/.config/Windsurf/logs/.../Windsurf ACP devin-cloud.log
→ "Connecting to remote ACP: wss://app.devin.ai/api/acp/live"
→ "WebSocket connected"
```

### Registro de agentes ACP no WindSurf:
```
Registering agent "devin-cli"
Registering agent "devin-cloud"
Registering agent "summary-agent"
```

---

## 📊 MCP SERVERS NO LANGUAGE SERVER

O Language Server conecta a múltiplos MCP servers via stdio JSON-RPC:

| MCP Server | Status |
|------------|--------|
| `pencil` | ✅ Conectado (design tool) |
| `slack` | ✅ Conectado (36 tools) |
| `aidesigner` | ✅ Conectado |
| `github` | ✅ Conectado |
| `context7` | ✅ Conectado |

**Config:** `~/.codeium/windsurf/mcp_config.json`

---

## 🔐 AUTENTICAÇÃO NO EXTENSION SERVER (Porta 40819)

O Extension Server na porta 40819 requer um **CSRF token diferente** do Language Server. Ele retorna:
```
HTTP/1.1 403 Forbidden
Invalid CSRF token
```

**Possíveis fontes do token:**
- `VSCODE_IPC_HOOK` socket (`/run/user/1000/vscode-ccc00a58-1.11-main.sock`)
- LocalStorage do Electron
- Comunicação interna via pipe/pipe nomeado

**Endpoints do Extension Server:**
```
exa.extension_server_pb.ExtensionServerService/LanguageServerStarted
exa.extension_server_pb.ExtensionServerService/SubscribeNativeValues
exa.extension_server_pb.ExtensionServerService/WatchForLints
exa.extension_server_pb.ExtensionServerService/OpenDiffZones
exa.extension_server_pb.ExtensionServerService/GetLSPCompletionItems
exa.extension_server_pb.ExtensionServerService/UpdateCascadeTrajectorySummaries
```

---

## 🎯 PADRÕES DE COMUNICAÇÃO EM APLICAÇÕES SIMILARES

| Aplicação | Protocolo | Streaming | Como o IDE recebe updates |
|-----------|-----------|-----------|---------------------------|
| **WindSurf** | Connect Protocol (Buf) | ✅ Server streaming | Conexão persistente HTTP/2 ou envelopes Connect |
| **Kimi Code** | ACP (JSON-RPC stdio) | ✅ Via events | Stdio pipes com eventos JSON-RPC |
| **Claude Code** | Claude Stream JSON | ✅ Server streaming | SSE ou stream JSON |
| **Cursor** | JSON Event Stream | ✅ Server streaming | Eventos JSON delimitados |
| **VS Code Copilot** | LSP + Stream | ✅ Server streaming | LSP notifications + stream |
| **Zed** | ACP (JSON-RPC) | ✅ Bidirectional | Stdio + ACP events |

**Conclusão:** Todos os IDEs modernos usam **conexões persistentes** (não polling) para receber atualizações em tempo real. O WindSurf não é exceção - ele usa Connect Protocol streaming.

---

## 🚀 ALTERNATIVAS AO POLLING (Ranqueadas)

### 1. 🥇 **Connect Protocol Streaming** (Ideal mas complexo)
- Usar o endpoint `StreamCascadeReactiveUpdates` 
- Requer engenharia reversa do envelope proprietário
- **Vantagem:** Streaming nativo, tempo real
- **Desvantagem:** Formato não documentado, pode quebrar

### 2. 🥈 **Polling Otimizado com `GetCascadeTrajectorySteps`** (Recomendado)
- Usar o novo endpoint que retorna só os steps
- Intervalo adaptativo (500ms durante atividade, 2s em idle)
- **Vantagem:** Funciona hoje, estável, baixo overhead
- **Desvantagem:** Ainda é polling

### 3. 🥉 **Observação via Kimi ACP** (Alternativa criativa)
- Como o Kimi roda dentro do WindSurf, podemos:
  - Criar um MCP server que o Kimi usa
  - O MCP server publica no nosso Message Bus
  - O Kimi "vê" o que o WindSurf está fazendo
- **Vantagem:** Usa canal oficial (MCP)
- **Desvantagem:** Indireto, requer que o Kimi esteja ativo

### 4. **File System Watcher** (Hack)
- O WindSurf salva trajectory em arquivos temporários
- Watch por mudanças nos arquivos
- **Vantagem:** Não requer API
- **Desvantagem:** Muito indireto, não confiável

---

## 🛠️ RECOMENDAÇÃO FINAL

Para o **serviço separado** (não OrchestraOS):

```
┌─────────────────────────────────────────────────────────────────────┐
│                    AGENT COMMUNICATION HUB                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐ │
│  │ WindSurf Adapter│    │  Kimi Adapter   │    │ Future Adapters │ │
│  │                 │    │                 │    │                 │ │
│  │ Polling otimizado│   │ ACP JSON-RPC    │    │ ...             │ │
│  │ GetCascadeTrajectorySteps              │    │                 │ │
│  │ Intervalo: 500ms│    │ Stdio pipe      │    │                 │ │
│  └────────┬────────┘    └────────┬────────┘    └─────────────────┘ │
│           │                      │                                   │
│           └──────────────────────┼───────────────────────────────────┘
│                                  ▼                                   │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              Message Bus (Event-Driven)                      │   │
│  │  • Canais: chat, tasks, progress, tools, thinking            │   │
│  │  • SSE streaming para observadores                           │   │
│  │  • Persistência SQLite/Redis                                 │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**O WindSurf Adapter deve usar:**
1. `GetCascadeTrajectorySteps` em vez de `GetCascadeTrajectory` (mais leve)
2. Intervalo adaptativo (rápido durante atividade, lento em idle)
3. Deduplicação de steps (só publica mudanças)
4. Reconexão automática quando portas/CSRF mudam

**Investigação futura (baixa prioridade):**
- Decodificar o envelope do Connect Protocol streaming do WindSurf
- Isso permitiria eliminar o polling completamente
- Requer análise do binary ou captura de tráfego com SSLKEYLOGFILE

---

## 📋 PRÓXIMOS PASSOS

1. **Implementar o Hub com polling otimizado** (2-3 dias)
   - Usar `GetCascadeTrajectorySteps`
   - SSE para clientes
   - Reconexão automática

2. **Investigar o envelope Connect Protocol** (futuro)
   - Usar SSLKEYLOGFILE + Wireshark para capturar tráfego
   - Ou analisar o binary com ghidra/IDA

3. **Conectar Kimi ACP** (1-2 dias)
   - Spawn `kimi acp` process
   - Bridge para o Message Bus
   - Suportar eventos de status

---

*Descobertas feitas em investigação ativa no sistema WindSurf em execução.*
