# Análise Completa: Streaming do WindSurf Language Server

> Data: 2026-05-17
> Status: Contrato mapeado, streaming não ativado externamente

## Resumo Executivo

O WindSurf Language Server usa **Connect Protocol** (gRPC sobre HTTP/1.1) para comunicação interna. Foram mapeados **141 métodos**, incluindo **11 endpoints de streaming**. O streaming real funciona via `StreamCascadeReactiveUpdates` (Server Streaming), mas requer:
1. Cliente Connect Protocol com `useBinaryFormat: true`
2. Protobuf binário corretamente serializado
3. Autenticação válida (CSRF token + API key)
4. Sessão Cascade ativa (cascadeId válido)

## Arquitetura de Comunicação

```
WindSurf IDE (Electron)
  └── Extension Server (porta ~40-43k, VS Code Extension Host)
       └── Language Server (porta ~34k, processo Go)
            ├── Servidor API (https://server.self-serve.windsurf.com)
            ├── Servidor Inference (https://inference.codeium.com)
            └── Outros serviços (Google Cloud IPs)
```

## Endpoints de Streaming Mapeados

| Método | Tipo | Request | Response |
|--------|------|---------|----------|
| StreamCascadeReactiveUpdates | Server Streaming | StreamReactiveUpdatesRequest | StreamReactiveUpdatesResponse |
| StreamCascadeSummariesReactiveUpdates | Server Streaming | StreamReactiveUpdatesRequest | StreamReactiveUpdatesResponse |
| StreamCascadePanelReactiveUpdates | Server Streaming | StreamReactiveUpdatesRequest | StreamReactiveUpdatesResponse |
| StreamUserTrajectoryReactiveUpdates | Server Streaming | StreamReactiveUpdatesRequest | StreamReactiveUpdatesResponse |
| GetChatMessage | Server Streaming | GetChatMessageRequest | GetChatMessageResponse |
| RawGetChatMessage | Server Streaming | RawGetChatMessageRequest | RawGetChatMessageResponse |
| HandleStreamingCommand | Server Streaming | HandleStreamingCommandRequest | HandleStreamingCommandResponse |
| HandleStreamingTab | Server Streaming | HandleStreamingTabRequest | HandleStreamingTabResponse |
| HandleStreamingTerminalCommand | Server Streaming | HandleStreamingTerminalCommandRequest | HandleStreamingTerminalCommandResponse |
| GetDeepWiki | Server Streaming | GetDeepWikiRequest | GetDeepWikiResponse |
| StreamTerminalShellCommand | Client Streaming | TerminalShellCommandStreamChunk | StreamTerminalShellCommandResponse |

## Schemas Protobuf Críticos

### StreamReactiveUpdatesRequest
```protobuf
message StreamReactiveUpdatesRequest {
  int32 protocol_version = 1;  // deve ser 1
  string id = 2;               // cascadeId
}
```

### StreamReactiveUpdatesResponse
```protobuf
message StreamReactiveUpdatesResponse {
  int64 version = 1;
  DiffMessage diff = 2;
  bytes full_state = 3;        // estado serializado
}
```

### GetChatMessageRequest
```protobuf
message GetChatMessageRequest {
  Metadata metadata = 1;
  repeated ChatMessage chat_messages = 3;
  Document active_document = 5;
  repeated string open_document_uris = 12;
  repeated string workspace_uris = 13;
  string active_selection = 11;
  ContextInclusionType context_inclusion_type = 8;
  Model chat_model = 9;
  string chat_model_name = 14;
  string system_prompt_override = 10;
  EnterpriseExternalModelConfig enterprise_chat_model_config = 15;
}
```

### SendUserCascadeMessageRequest
```protobuf
message SendUserCascadeMessageRequest {
  string cascade_id = 1;
  repeated TextOrScopeItem items = 2;
  Metadata metadata = 3;
  ExperimentConfig experiment_config = 4;
  CascadeConfig cascade_config = 5;
  repeated ImageData images = 6;
  repeated string recipe_ids = 7;
  bool blocking = 8;           // ← IMPORTANTE!
  repeated CortexTrajectoryStep additional_steps = 9;
}
```

## Configuração do Cliente Connect

A extensão WindSurf configura o transporte assim:
```javascript
createConnectTransport({
  baseUrl: `http://${languageServerAddress}`,
  useBinaryFormat: true,        // protobuf binário
  httpVersion: "1.1",
  interceptors: [csrfInterceptor]
});
```

## Problemas Encontrados

### 1. "unsupported protocol version 0"
- O servidor interpreta requisições mal formadas como protocolo versão 0
- Acontece quando o envelope Connect não está correto
- **Solução necessária**: usar cliente Connect oficial (`@connectrpc/connect`)

### 2. "Empty reply from server" / "socket hang up"
- O servidor fecha a conexão TCP antes de enviar resposta HTTP
- Pode ser devido a:
  - Autenticação inválida
  - Formato do protobuf incorreto
  - Falta de campos obrigatórios

### 3. "run state not found"
- Acontece quando tentamos enviar mensagem para um cascadeId inexistente
- Precisa criar sessão com `StartCascade` primeiro

## Integrações Internas do WindSurf

### Kimi Code (extensão)
- Roda como processo independente: `kimi --wire --no-thinking`
- Protocolo ACP (Agent Communication Protocol) sobre JSON-RPC 2.0
- Logs: `~/.config/Windsurf/logs/*/Kimi Code.log`
- Eventos de streaming: `ContentPart`, `ToolCall`, `ToolCallPart`, `StatusUpdate`, `TurnEnd`

### Devin (extensão)
- Conecta via WebSocket: `wss://app.devin.ai/api/acp/live`
- Também usa ACP sobre WebSocket

### MCP Servers
- Pencil (design), Slack, GitHub, Context7
- Comunicação via stdio (JSON-RPC 2.0)

## Próximos Passos Recomendados

### Opção A: Cliente Connect Oficial (Recomendada)
1. Instalar `@connectrpc/connect` e `@connectrpc/connect-node`
2. Definir os schemas protobuf manualmente (ou extrair do .proto)
3. Implementar cliente streaming usando `createPromiseClient`
4. Testar com sessão real do WindSurf

### Opção B: Proxy de Tráfego
1. Interceptar o tráfego entre Extension Server e Language Server
2. Decodificar o protobuf binário
3. Replicar o formato exato

### Opção C: Reutilizar Extensão
1. Criar uma extensão WindSurf customizada
2. Usar a API interna da extensão WindSurf para fazer chamadas
3. Exportar os dados via WebSocket/SSE

## Decisões Pendentes

1. **Formato do Bus de Mensagens**: SSE, WebSocket, ou gRPC?
2. **Persistência**: SQLite, Redis, ou arquivo?
3. **Autenticação entre agentes**: tokens, JWT, ou outro?
4. **Integração com Kimi**: ACP bridge ou stdio?
