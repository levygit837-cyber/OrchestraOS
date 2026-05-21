# Engenharia Reversa WindSurf - Relatório Final

> Data: 2026-05-17
> Analista: Kimi Code (Agente Autônomo)
> Ferramentas Usadas: strace, tcpdump, strings, objdump, grep, python, node.js, curl, ss, ps, readelf

## TL;DR

O **WindSurf Language Server** usa **Connect Protocol** (gRPC sobre HTTP/1.1) com **141 métodos** (11 streaming). O streaming funciona via `StreamCascadeReactiveUpdates`, mas **requer o cliente Connect oficial** (`@connectrpc/connect`). Requisições manuais são rejeitadas com "empty reply" ou "unsupported protocol version 0".

## O Que Funciona

### 1. Detecção Automática
✅ Porta do Language Server (varia, atual: 34567)
✅ CSRF Token (do ambiente do processo)
✅ API Key (do backend WindAgent)
✅ PIDs dos processos

### 2. Endpoints Unary Confirmados
✅ `Heartbeat` - mantém conexão viva
✅ `StartCascade` - cria nova sessão
✅ `GetCascadeTrajectory` - obtém estado atual
✅ `SendUserCascadeMessage` - envia mensagem
✅ `InitializeCascadePanelState` - configura painel

### 3. Streaming Descoberto (mas não ativado externamente)
🔍 `StreamCascadeReactiveUpdates` - Server Streaming
🔍 `StreamCascadeSummariesReactiveUpdates` - Server Streaming
🔍 `StreamCascadePanelReactiveUpdates` - Server Streaming
🔍 `StreamUserTrajectoryReactiveUpdates` - Server Streaming
🔍 `GetChatMessage` - Server Streaming (!! chat direto !!)
🔍 `RawGetChatMessage` - Server Streaming
🔍 `HandleStreamingCommand` - Server Streaming
🔍 `HandleStreamingTab` - Server Streaming
🔍 `HandleStreamingTerminalCommand` - Server Streaming
🔍 `GetDeepWiki` - Server Streaming
🔍 `StreamTerminalShellCommand` - Client Streaming

### 4. Comunicação Interna Mapeada
```
WindSurf IDE
  └── Extension Host (VS Code)
       └── Extension Server (porta ~40-43k)
            └── Language Server (porta ~34k, Go binary)
                 ├── API Server (https://server.self-serve.windsurf.com)
                 ├── Inference Server (https://inference.codeium.com)
                 └── Google Cloud (34.49.14.144:443, etc.)
```

### 5. Integrações do WindSurf
✅ **Kimi Code**: Processo `kimi --wire --no-thinking`, ACP/JSON-RPC 2.0 over stdio
✅ **Devin**: WebSocket `wss://app.devin.ai/api/acp/live`, ACP over WebSocket
✅ **MCP Servers**: Pencil, Slack, GitHub, Context7 via stdio/JSON-RPC

## O Que Não Funciona (Ainda)

### Streaming Externo ❌
- Requisições manuais com `curl` falham
- Servidor fecha conexão ("empty reply", "socket hang up")
- Erro "unsupported protocol version 0" no servidor
- Causa raiz: requer cliente Connect oficial com `useBinaryFormat: true`

### Formato Protobuf ❌
- Tentativas de serialização manual falham
- O Connect Protocol tem um envelope específico que não conseguimos replicar
- O `@bufbuild/protobuf` v2 tem API diferente da v1

### Autenticação Completa ❌
- CSRF token funciona para unary básico
- Mas pode faltar validação de API key real para streaming
- O servidor pode rejeitar conexões de fontes não-autorizadas

## Schemas Críticos Extraídos

### StreamReactiveUpdatesRequest
```protobuf
message StreamReactiveUpdatesRequest {
  int32 protocol_version = 1;  // = 1
  string id = 2;               // cascadeId
}
```

### StreamReactiveUpdatesResponse
```protobuf
message StreamReactiveUpdatesResponse {
  int64 version = 1;
  DiffMessage diff = 2;
  bytes full_state = 3;        // estado serializado (provavelmente JSON)
}
```

### GetChatMessageRequest (Streaming de Chat!)
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
  bool blocking = 8;           // ← bloqueia até completar?
  repeated CortexTrajectoryStep additional_steps = 9;
}
```

## Descobertas Importantes

### 1. O Language Server é um Proxy
O binary `language_server_linux_x64` atua como proxy entre o IDE local e os servidores remotos da Codeium/WindSurf. Ele:
- Recebe requisições Connect Protocol localmente
- Encaminha para `server.self-serve.windsurf.com` e `inference.codeium.com`
- Mantém estado local (sessões, trajetórias, cache)

### 2. A Extensão WindSurf é o Cliente Oficial
A extensão VS Code (`extension.js`) contém:
- Cliente Connect completo (`@connectrpc/connect`)
- Schemas protobuf de 1736 mensagens
- Lógica de streaming com `for await...of`
- Interceptors para CSRF e API key

### 3. Kimi Code é Independente
- Roda como processo separado (`kimi` binary)
- Comunica com WindSurf via ACP (JSON-RPC 2.0)
- Tem seu próprio protocolo de eventos (`ContentPart`, `ToolCall`, etc.)
- NÃO usa o Language Server do WindSurf para streaming

### 4. Devin Também é Independente
- Conecta diretamente aos servidores Devin via WebSocket
- Usa ACP sobre WebSocket
- Não passa pelo Language Server

## Caminhos Para Implementar Streaming

### Caminho A: Cliente Connect Oficial (Recomendado)
**Complexidade**: Alta
**Confiabilidade**: Alta

1. Instalar `@connectrpc/connect` e `@connectrpc/connect-node`
2. Definir schemas protobuf manualmente (usando API v2)
3. Implementar cliente com `createPromiseClient`
4. Testar streaming com sessão real

**Problema**: API do @bufbuild/protobuf v2 é diferente da v1. Precisa aprender a nova API.

### Caminho B: Reutilizar Código da Extensão
**Complexidade**: Média
**Confiabilidade**: Muito Alta

1. Criar uma extensão WindSurf customizada
2. Importar o cliente Connect da extensão oficial
3. Fazer chamadas de streaming internamente
4. Exportar dados via WebSocket/SSE para o Bus

**Problema**: Requer conhecimento de desenvolvimento de extensões VS Code.

### Caminho C: Proxy de Tráfego
**Complexidade**: Média
**Confiabilidade**: Média

1. Interceptar todo o tráfego entre Extension Server e Language Server
2. Decodificar protobuf binário
3. Replicar o formato exato nas requisições

**Problema**: Requer MITM constante e pode quebrar com atualizações.

### Caminho D: Integrar com Kimi Code Diretamente
**Complexidade**: Baixa
**Confiabilidade**: Alta

1. Em vez de usar o Language Server do WindSurf, integrar com o Kimi Code
2. O Kimi já tem streaming via ACP/JSON-RPC
3. Criar um bridge que converte eventos Kimi para o Bus

**Problema**: Limitado ao Kimi Code, não acessa o Cascade do WindSurf.

## Recomendação Final

Para o **Agent Communication Hub**, recomendo uma **abordagem híbrida**:

1. **WindSurf Adapter**: Usar Caminho B (extensão customizada) ou Caminho C (proxy) para acessar o Cascade
2. **Kimi Adapter**: Usar Caminho D (integração direta via ACP stdio)
3. **Bus Central**: SSE + SQLite para persistência

Isso permite:
- Streaming real do WindSurf Cascade (sem pooling)
- Streaming real do Kimi Code (já funciona)
- Persistência de conversas
- Comunicação assíncrona entre agentes

## Próximo Passo Imediato

Implementar o **Kimi Adapter** primeiro (mais fácil), pois:
- O protocolo ACP é documentado nos logs
- É JSON-RPC 2.0 (fácil de implementar)
- Já temos os eventos de streaming mapeados
- Não depende do Connect Protocol

Depois, retornar ao **WindSurf Adapter** com mais tempo para:
- Aprender a API do @bufbuild/protobuf v2
- Implementar o cliente Connect correto
- Testar o streaming real
