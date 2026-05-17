# WindSurf Language Server - Contrato Completo

> Data: 2026-05-17
> Fonte: Engenharia reversa do extension.js e language_server binary

## Endpoints de Streaming Descobertos

### 1. StreamCascadeReactiveUpdates
- **Tipo**: Server Streaming
- **Request**: `StreamReactiveUpdatesRequest`
  - `protocol_version` (int32) = 1
  - `id` (string) = cascadeId
- **Response**: `StreamReactiveUpdatesResponse`
  - `version` (int64)
  - `diff` (DiffMessage) - delta de mudanças
  - `full_state` (bytes) - estado completo serializado

### 2. StreamCascadeSummariesReactiveUpdates
- **Tipo**: Server Streaming
- **Mesmo request/response que acima**

### 3. StreamCascadePanelReactiveUpdates
- **Tipo**: Server Streaming
- **Mesmo request/response**

### 4. StreamUserTrajectoryReactiveUpdates
- **Tipo**: Server Streaming
- **Request/Response**: Mesmo formato

### 5. GetChatMessage
- **Tipo**: Server Streaming
- **Request**: `GetChatMessageRequest`
  - `metadata` (Metadata)
  - `chat_messages` (repeated ChatMessage)
  - `active_document` (Document)
  - `open_document_uris` (repeated string)
  - `workspace_uris` (repeated string)
  - `chat_model` (enum Model)
  - `chat_model_name` (string)
  - `system_prompt_override` (string)
- **Response**: `GetChatMessageResponse` (streaming)

### 6. RawGetChatMessage
- **Tipo**: Server Streaming
- Similar ao GetChatMessage mas raw

### 7. HandleStreamingCommand
- **Tipo**: Server Streaming

### 8. HandleStreamingTab
- **Tipo**: Server Streaming

### 9. HandleStreamingTerminalCommand
- **Tipo**: Server Streaming

### 10. StreamTerminalShellCommand
- **Tipo**: Client Streaming

## Configuração do Cliente Connect

A extensão WindSurf usa:
```javascript
createConnectTransport({
  baseUrl: `http://${languageServerAddress}`,
  useBinaryFormat: true,        // ← CRÍTICO: usa protobuf binário
  httpVersion: "1.1",           // ← HTTP/1.1 para local
  interceptors: [csrfInterceptor]
});
```

Para servidor remoto (API):
```javascript
createConnectTransport({
  baseUrl: apiServerUrl,
  useBinaryFormat: true,
  httpVersion: "2",             // ← HTTP/2 para remoto
  interceptors: [apiKeyInterceptor]
});
```

## Formato do Protocolo

### Unary (ex: Heartbeat, StartCascade)
- **Request**: POST com Content-Type `application/json` OU `application/proto`
- **Body**: JSON ou protobuf binário
- **Response**: JSON ou protobuf binário

### Server Streaming (ex: StreamCascadeReactiveUpdates)
- **Request**: POST com Content-Type `application/connect+proto`
- **Body**: protobuf binário (SEM envelope no request)
- **Response**: Stream de envelopes binários
  - Cada envelope: 1 byte flags + 4 bytes length (big-endian) + N bytes protobuf
  - Flags: bit 0 = compressed

## Servidores Remotos
- `https://server.self-serve.windsurf.com` (API server)
- `https://inference.codeium.com` (Inference API)
- IP: `34.49.14.144:443` (Google Cloud)
- IP: `35.223.238.178:443` (outro servidor)

## Autenticação
- CSRF Token: extraído de `WINDSURF_CSRF_TOKEN` no ambiente do processo
- API Key: `sk-ws-01-...` (do backend WindAgent)
- O language_server valida o CSRF token localmente

## Comunicação Interna
- **Extension Server** (porta ~40-43k) ↔ **Language Server** (porta ~34k)
- A extensão WindSurf (VS Code) se comunica com Extension Server via stdio/LSP
- O Extension Server se comunica com Language Server via Connect Protocol HTTP
- O Language Server se comunica com servidores remotos via HTTPS

## Observações
- O binary do language_server foi stripado (sem símbolos de debug)
- O Connect Protocol streaming exige `useBinaryFormat: true`
- O envelope de streaming pode ser customizado (diferente do padrão Buf Connect)
- Requisições com Content-Type `application/connect+proto` falham com "empty reply" quando o corpo está incorreto

## Próximos Passos
1. Implementar cliente Connect usando `@connectrpc/connect-node`
2. Gerar stubs protobuf a partir dos schemas extraídos
3. Testar streaming com protobuf binário corretamente serializado
4. Decodificar o `full_state` bytes para extrair o estado da trajetória
