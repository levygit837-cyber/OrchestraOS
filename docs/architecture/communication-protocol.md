# Protocolo de Comunicacao

## Objetivo

Definir como agentes e Orchestrator trocam eventos, comandos, notificacoes, checkpoints e pedidos de aprovacao.

O WebSocket e o canal vivo. O Event Store e a fonte de verdade.

Os schemas detalhados ficam em `docs/contracts/json-schemas.md`.

## Envelope Padrao

Toda mensagem relevante deve seguir um envelope estavel.

```json
{
  "id": "evt_01HX0000000000000000000000",
  "type": "tool.requested",
  "task_id": "task_123",
  "run_id": "run_456",
  "agent_id": "agent_789",
  "work_unit_id": "wu_001",
  "trace_id": "trace_123",
  "span_id": "span_456",
  "sequence": 42,
  "priority": "checkpoint",
  "requires_ack": true,
  "created_at": "2026-05-03T12:00:00Z",
  "payload": {}
}
```

Campos obrigatorios:

- `id`: identificador unico e ordenavel.
- `type`: tipo do evento ou comando.
- `task_id`: task relacionada.
- `run_id`: execucao relacionada.
- `agent_id`: agente relacionado, quando aplicavel.
- `work_unit_id`: unidade de trabalho relacionada, quando aplicavel.
- `trace_id`, `span_id` e `sequence`: correlacao e ordenacao para tracing, quando aplicavel.
- `priority`: prioridade operacional.
- `requires_ack`: indica se a outra ponta precisa confirmar recebimento.
- `created_at`: data de criacao.
- `payload`: dados especificos do tipo.

## Prioridades

| Prioridade | Semantica |
| --- | --- |
| `interrupt` | O agente deve parar assim que for seguro e processar a mensagem. |
| `checkpoint` | O agente deve processar ao chegar no proximo checkpoint. |
| `notification` | Informacao relevante, nao bloqueante. |
| `background` | Contexto passivo ou informacao de baixa prioridade. |

## Eventos Enviados Pelo Agente

Eventos minimos:

- `agent.started`
- `agent.heartbeat`
- `agent.checkpoint_reached`
- `agent.message`
- `agent.ledger_updated`
- `agent.loop_detected`
- `task.graph_created`
- `prompt.snapshot_created`
- `tool.requested`
- `tool.completed`
- `tool.failed`
- `artifact.created`
- `validation.started`
- `validation.completed`
- `task.completed`
- `task.failed`

## Comandos Enviados Pelo Orchestrator

Comandos minimos:

- `task.start`
- `task.pause`
- `task.resume`
- `task.cancel`
- `message.notify`
- `message.interrupt`
- `tool.approved`
- `tool.denied`
- `policy.updated`
- `ledger.update_requested`

## Pedido de Ferramenta

Exemplo de `tool.requested`:

```json
{
  "id": "evt_01HXTOOLREQUEST",
  "type": "tool.requested",
  "task_id": "task_123",
  "run_id": "run_456",
  "agent_id": "agent_789",
  "priority": "interrupt",
  "requires_ack": true,
  "created_at": "2026-05-03T12:05:00Z",
  "payload": {
    "tool": "shell.exec",
    "reason": "Rodar testes unitarios antes de concluir a task.",
    "risk": "low",
    "command": "go test ./...",
    "scope": "task_worktree"
  }
}
```

## Reconexao e Replay

Agentes devem reconectar informando o ultimo evento processado.

```json
{
  "type": "agent.reconnected",
  "task_id": "task_123",
  "run_id": "run_456",
  "agent_id": "agent_789",
  "last_seen_event_id": "evt_01HXLAST"
}
```

O Orchestrator deve reenviar comandos pendentes e eventos necessarios para restaurar o estado da sessao.

## Contratos e Validacao

Todos os tipos de eventos e comandos devem ter JSON Schema. Entradas externas vindas de GitHub, conectores opcionais, agentes e ferramentas devem ser validadas nas bordas do sistema.

O protocolo interno pode evoluir, mas deve manter versionamento explicito quando houver quebra de compatibilidade.

Eventos ja persistidos nao devem ser reescritos. Quando houver mudanca quebrante, criar nova versao do schema e manter compatibilidade de leitura.
