# Git Message Bus (GMB) — Comunicação entre Agentes

## O que é

O GMB é uma **branch dedicada** (`orchestrator-comms`) que funciona como message bus entre agentes executores. Usa git como infraestrutura de comunicação — persistente, auditável, versionada.

## Por que Git

| Vantagem | Explicação |
|----------|------------|
| Persistente | Mensagens não se perdem se um agente crashar |
| Auditável | `git log` mostra toda a conversa entre agentes |
| Versionada | Conflitos de escrita são detectados e resolvidos pelo git |
| Sem infra nova | Usa o que já existe (git) |

## Estrutura

```
orchestrator-comms/  (branch)
├── inbox/
│   ├── A01/           # Mensagens destinadas ao Agente 1
│   ├── A02/           # Mensagens destinadas ao Agente 2
│   └── A03/           # Mensagens destinadas ao Agente 3
├── threads/           # Conversas por tópico
│   └── thread-001-userprofile.md
└── signals/           # Sinalizações simples (JSON)
    └── A02-blocked-A01.json
```

## Convenções

### Inbox
- Cada agente tem sua própria pasta: `inbox/{AGENTE_ID}/`
- Mensagens são arquivos markdown: `YYYY-MM-DD-{NNN}-{tipo}.md`
- O agente destino lê e **deleta** a mensagem após processar (ou move para `inbox/{AGENTE_ID}/processed/`)

### Threads
- Tópicos de discussão persistentes
- Nome: `thread-{NNN}-{topico-resumido}.md`
- Todos os agentes envolvidos leem e escrevem no mesmo thread
- Formato padronizado (veja exemplo abaixo)

### Signals
- Arquivos JSON de 1 bit (booleano/enum)
- Nome: `{ORIGEM}-{acao}-{DESTINO}.json`
- Exemplo: `A02-done-with-api.json` = "A02 terminou a API"
- O agente destino lê e **deleta** após processar

## Como Usar

### Agente que precisa de algo (origem)

```bash
# 1. Fazer checkout da branch comms
git fetch origin orchestrator-comms
git worktree add ../comms-temp origin/orchestrator-comms

# 2. Criar thread
cd ../comms-temp
cat > threads/thread-042-userprofile.md << 'EOF'
# Thread 042: Exportar UserProfile

**De:** A01 (Kimi-CLI)
**Para:** A02 (WindSurf)
**Tipo:** request
**Prioridade:** blocking
**Data:** 2026-05-12T14:00:00Z

## Request

Preciso que o módulo `users` exporte:

```go
type UserProfile struct {
    ID   string
    Name string
}
```

## Contexto

Estou implementando `internal/modules/agent/service.go` e o método `FindOrCreate` 
precisa vincular um agente a um perfil de usuário.

## Deadline

Não é urgente, mas bloqueia o item 5 do meu checklist.
EOF

# 3. Notificar destino
mkdir -p inbox/A02
cat > inbox/A02/2026-05-12-042-userprofile.md << 'EOF'
**Thread:** thread-042-userprofile.md
**De:** A01
**Tipo:** request
**Prioridade:** blocking
EOF

# 4. Commit e push
git add .
git commit -m "comms: A01→A02 thread-042 request UserProfile"
git push origin orchestrator-comms

# 5. Limpar worktree temporário
cd ..
git worktree remove comms-temp
```

### Agente que recebe notificação (destino)

```bash
# No início de cada Ralph Loop:
git fetch origin orchestrator-comms

# Verificar inbox
if git show origin/orchestrator-comms:inbox/A02/ 2>/dev/null | grep -q .; then
    echo "Você tem mensagens pendentes"
    # Ler mensagens, processar, responder no thread
fi
```

### Respondendo no thread

```bash
# Editar o thread existente
cat >> threads/thread-042-userprofile.md << 'EOF'

---

## Resposta de A02

**Data:** 2026-05-12T14:10:00Z
**Status:** done

Implementado em `internal/modules/users/profile.go`:

```go
package users

type UserProfile struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}
```

Exportado via `users.UserProfile`.
EOF

# Notificar origem
mkdir -p inbox/A01
cat > inbox/A01/2026-05-12-042-resposta.md << 'EOF'
**Thread:** thread-042-userprofile.md
**De:** A02
**Tipo:** response
**Status:** done
EOF
```

## Polling vs. Notificação

**O GMB usa polling via git.** Cada agente, no início do Ralph Loop, faz `git fetch` da branch `orchestrator-comms` e verifica seu inbox.

**Por que não WebSocket?**
- WebSocket adiciona infraestrutura (servidor, porta, manutenção)
- Agentes operam em ciclos de minutos, não milissegundos
- Polling via git é suficiente e mais simples
- Se no futuro a latência de minutos for problema, aí se discute WebSocket

## Regras

1. **Uma mensagem por arquivo** — não edite mensagens já enviadas
2. **Deletar após processar** — inbox não é arquivo morto
3. **Thread é append-only** — adicione respostas no final, nunca delete
4. **Commits pequenos** — um commit por interação
5. **Não use o GMB para chat** — apenas requests/responses estruturados
6. **Prioridade blocking = pare e espere** — prioridade normal = continue com o que for possível

## Exemplos

Veja os arquivos `EXEMPLO-` neste diretório.
