# Setup Opcional do Slack MCP para OrchestraOS

Slack não é dependência do MVP atual. Este setup documenta uma integração opcional de chat que pode ser reativada futuramente se GitHub/CLI não forem suficientes para captura, notificações ou rotinas.

Workspace: **kanbanespaco.slack.com**
Servidor MCP: **coreyepstein/advanced-slack-mcp**
Status: ✅ **CONFIGURADO E FUNCIONANDO**

---

## Resumo da configuração

| Campo | Valor |
|-------|-------|
| Workspace | `kanbanespaco` |
| Team ID | `T0B254XCX40` |
| Bot | `@orchestraos` |
| User | `@levygamer72` |
| Bot Token | `xoxb-...` ✅ |
| User Token | `xoxp-...` ✅ |

---

## Arquivos configurados

- `~/.kimi/mcp.json` — Configuração do servidor MCP no Kimi Code CLI
- `~/.codex/config.toml` — Configuração do servidor MCP no Codex (OpenAI)
- `OrchestraOS/.codex/config.toml` — Configuração do servidor MCP por projeto no Codex
- `~/.codeium/windsurf/mcp_config.json` — Configuração do servidor MCP no Windsurf
- `~/tools/slack-mcp/workspaces.json` — Configuração do workspace

---

## Ferramentas disponíveis (12)

| Tool | Descrição |
|------|-----------|
| `list_channels` | Lista canais do workspace |
| `read_channel` | Lê mensagens recentes de um canal |
| `read_thread` | Lê uma thread completa |
| `get_channel_info` | Detalhes do canal (tópico, propósito, membros) |
| `send_message` | Envia mensagem para canal |
| `reply_thread` | Responde em uma thread |
| `send_dm` | Envia DM para usuário |
| `list_users` | Lista membros do workspace |
| `find_user` | Busca usuário por nome, email ou ID |
| `add_reaction` | Adiciona reação emoji |
| `search_messages` | 🔍 Busca mensagens (requer user token — OK!) |
| `upload_file` | Faz upload de arquivo |

---

## Como usar no Kimi

O servidor MCP está ativo automaticamente. Você pode pedir coisas como:

> *"Liste os canais do workspace kanbanespaco"*

> *"Leia as últimas 10 mensagens do canal #geral"*

> *"Busque mensagens sobre 'deploy' nos últimos 7 dias"*

> *"Envie uma mensagem no canal #dev dizendo que o build foi concluído"*

> *"Quem são os membros do workspace?"*

---

## Observações importantes

### Acesso a canais
- Com o **User Token** (`xoxp-`), o Kimi acessa automaticamente todos os canais que você já é membro. **Não precisa convidar o bot.**
- Para canais que você **não** participa, o bot `@orchestraos` precisa ser convidado:
  ```
  /invite @orchestraos
  ```

### Mensagens aparecem como você
- Como configuramos o **User Token**, quando o Kimi envia mensagens, elas aparecem como se **você** tivesse enviado (foto e nome @levygamer72).
- Se quiser que apareça como o bot, remova o User Token do `~/.kimi/mcp.json`.

### Segurança
- Os tokens **não estão no git**. O `mcp.json` está fora do projeto (em `~/.kimi/`) e o `workspaces.json` não armazena tokens (usa variáveis de ambiente).
- **Nunca commite tokens** em arquivos versionados.

---

## Testar conexão manualmente

```bash
cd /home/levybonito/tools/slack-mcp
npm run auth:check
```

Ou via Kimi:
```bash
kimi mcp test slack
```

---

## 8. Configurar no Codex (OpenAI)

✅ **Já configurado!**

### Configuração global
Arquivo: `~/.codex/config.toml`

```toml
[mcp_servers.slack]
command = "npx"
args = ["tsx", "/home/levybonito/tools/slack-mcp/src/server.ts"]
env = { SLACK_TOKEN_KANBANESPACO_BOT = "xoxb-...", SLACK_TOKEN_KANBANESPACO_USER = "xoxp-..." }
startup_timeout_sec = 30
```

### Configuração por projeto
Arquivo: `OrchestraOS/.codex/config.toml`

O projeto `OrchestraOS` também foi adicionado como **trusted** no Codex.

Verifique com:
```bash
codex mcp list
codex mcp get slack
```

---

## Configurar no Windsurf

✅ **Já configurado!**

Arquivo: `~/.codeium/windsurf/mcp_config.json`

```json
{
  "mcpServers": {
    "slack": {
      "name": "slack",
      "transport": "stdio",
      "command": "npx",
      "args": [
        "tsx",
        "/home/levybonito/tools/slack-mcp/src/server.ts"
      ],
      "env": {
        "SLACK_TOKEN_KANBANESPACO_BOT": "xoxb-...",
        "SLACK_TOKEN_KANBANESPACO_USER": "xoxp-..."
      }
    }
  }
}
```

Reinicie o Windsurf ou recarregue a janela (`Ctrl+Shift+P` → `Developer: Reload Window`) para aplicar as mudanças.

---

## Troubleshooting

| Problema | Solução |
|----------|---------|
| `channel_not_found` | Você não é membro do canal. Entre no canal ou convide `@orchestraos` |
| `not_in_channel` | Bot não está no canal privado. Use `/invite @orchestraos` |
| `missing_scope` | Falta permissão no Slack App. Vá em api.slack.com/apps → OAuth & Permissions e adicione o scope |
| `invalid_auth` | Token expirado. Regenere em api.slack.com/apps → OAuth & Permissions → Reinstall |
