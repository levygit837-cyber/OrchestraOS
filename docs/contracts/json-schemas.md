# JSON Schemas

Este documento define os contratos iniciais de eventos e comandos. Os schemas executaveis ficam em `contracts/schemas/`; este arquivo continua como indice narrativo e regra de evolucao dos contratos.

## Convenções

- Schemas usam JSON Schema draft 2020-12.
- Todo evento ou comando deve ter `id`, `type`, `task_id`, `created_at` e `payload`.
- `run_id` e obrigatorio para eventos ligados a execucao (`run.*`, `agent.*`, `tool.*`) e opcional para eventos que existem antes de uma run, como `task.created` e `work_unit.created`.
- Campos desconhecidos devem ser rejeitados nos limites externos e aceitos com cuidado nos limites internos apenas quando houver versionamento.
- `event_id` deve ser idempotente.
- `sequence` deve ser monotonicamente crescente por `run_id` quando aplicavel.

## Escopo M0 De Schemas Executaveis

| Schema | Tipo Go | Finalidade |
| --- | --- | --- |
| `contracts/schemas/domain/task.schema.json` | `internal/domain.Task` | Contrato da entidade `Task`. |
| `contracts/schemas/domain/run.schema.json` | `internal/domain.Run` | Contrato da entidade `Run`. |
| `contracts/schemas/domain/work-unit.schema.json` | `internal/domain.WorkUnit` | Contrato da entidade `WorkUnit`. |
| `contracts/schemas/domain/agent.schema.json` | `internal/domain.Agent` | Contrato minimo da entidade `Agent`. |
| `contracts/schemas/domain/agent-session.schema.json` | `internal/domain.AgentSession` | Contrato da entidade `AgentSession`. |
| `contracts/schemas/protocol/event-envelope.schema.json` | `internal/domain.EventEnvelope` | Envelope comum de eventos e comandos; representa `Event` no M0. |

Nao criar no M0:

- `orchestrator.schema.json`: o Orchestrator e componente/control plane, nao entidade de dominio necessaria para persistencia inicial.
- `communication-protocol.schema.json`: o contrato relevante ja e o envelope versionado e os payloads de eventos/comandos.
- `session.schema.json`: `AgentSession` cobre o caso operacional inicial; sessao generica fica para CLI, GitHub ou conectores quando precisarem de estado vivo proprio.

Essa decisao de escopo esta registrada na [ADR 0013](../adr/0013-m0-domain-contract-scope.md). O pacote Go `contracts` embute `contracts/schemas/` para que testes e futuras bordas do sistema usem os mesmos artefatos versionados.

## Envelope Base

```json
{
  "$id": "https://orchestraos.local/schemas/protocol/event-envelope.schema.json",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "type": "object",
  "additionalProperties": false,
  "required": [
    "id",
    "type",
    "version",
    "task_id",
    "sequence",
    "priority",
    "requires_ack",
    "created_at",
    "payload"
  ],
  "allOf": [
    {
      "if": {
        "properties": {
          "type": { "pattern": "^(run|agent|tool)\\." }
        },
        "required": ["type"]
      },
      "then": {
        "required": ["run_id"]
      }
    }
  ],
  "properties": {
    "id": { "type": "string", "minLength": 1 },
    "type": { "type": "string", "minLength": 1 },
    "version": { "type": "string", "pattern": "^v[0-9]+$" },
    "task_id": { "type": "string", "minLength": 1 },
    "run_id": { "type": "string", "minLength": 1 },
    "work_unit_id": { "type": "string" },
    "agent_id": { "type": "string" },
    "trace_id": { "type": "string" },
    "span_id": { "type": "string" },
    "parent_span_id": { "type": "string" },
    "sequence": { "type": "integer", "minimum": 0 },
    "priority": {
      "type": "string",
      "enum": ["interrupt", "checkpoint", "notification", "background"]
    },
    "requires_ack": { "type": "boolean" },
    "created_at": { "type": "string", "format": "date-time" },
    "payload": { "type": "object" }
  }
}
```

## Tipos Principais

| Tipo | Direção | Finalidade |
| --- | --- | --- |
| `user.message_received` | Intake -> Orchestrator | Registrar entrada humana. |
| `task.created` | Orchestrator -> Store | Criar task. |
| `task.graph_created` | Orchestrator -> Store | Registrar DAG planejado. |
| `run.started` | Orchestrator -> Store | Iniciar run. |
| `agent.started` | Agent -> Orchestrator | Confirmar agente vivo. |
| `agent.heartbeat` | Agent -> Orchestrator | Sinal de vida. |
| `agent.checkpoint_reached` | Agent -> Orchestrator | Ponto seguro para comandos pendentes. |
| `agent.message` | Bidirecional | Mensagem auditavel. |
| `prompt.dynamic_fragment_created` | Orchestrator -> Store | Registrar fragmento temporario de prompt. |
| `toolset.snapshot_created` | Orchestrator -> Store | Registrar ferramentas disponiveis para a sessao. |
| `toolset.change_requested` | Agent -> Orchestrator | Solicitar ferramenta ausente ou expansao de toolset. |
| `agent.session_reconfigured` | Orchestrator -> Store | Registrar reinicio ou troca controlada de prompt/toolset. |
| `tool.requested` | Agent -> Orchestrator | Solicitar ferramenta. |
| `tool.approved` | Orchestrator -> Agent | Aprovar ferramenta. |
| `tool.denied` | Orchestrator -> Agent | Negar ferramenta. |
| `tool.completed` | Agent -> Orchestrator | Resultado de ferramenta. |
| `artifact.created` | Agent -> Orchestrator | Registrar evidencia. |
| `validation.completed` | Agent -> Orchestrator | Registrar validacao. |
| `task.completed` | Agent/Orchestrator -> Store | Finalizar task ou work unit. |
| `task.failed` | Agent/Orchestrator -> Store | Registrar falha. |
| `run.pause` | Orchestrator -> Agent | Pausar run. |
| `run.resume` | Orchestrator -> Agent | Retomar run. |
| `run.cancel` | Orchestrator -> Agent | Cancelar run. |
| `message.notify` | Orchestrator -> Agent | Notificacao assíncrona. |
| `message.interrupt` | Orchestrator -> Agent | Intervencao prioritária. |
| `policy.updated` | Orchestrator -> Agent | Atualizar politica aplicavel. |
| `memory.candidate_created` | Memory -> Store | Registrar candidato de memoria derivado de fonte canonica. |
| `memory.candidate_rejected` | Memory -> Store | Registrar descarte de candidato por ruido, duplicidade, risco ou baixa relevancia. |
| `memory.record_created` | Memory -> Store | Persistir memoria aprovada com evidencias. |
| `memory.record_superseded` | Memory -> Store | Registrar que uma memoria substituiu outra. |
| `memory.retrieval_requested` | Orchestrator -> Memory | Solicitar memorias relevantes para uma run ou checkpoint. |
| `memory.bundle_created` | Memory -> Store | Registrar conjunto de memorias selecionadas. |
| `memory.bundle_injected` | Orchestrator -> Agent | Registrar contexto de memoria entregue ao agente. |
| `memory.feedback_recorded` | Agent/Orchestrator -> Memory | Registrar uso, rejeicao ou obsolescencia de memoria. |
| `memory.record_expired` | Memory -> Store | Registrar expiracao de memoria temporaria ou stale. |

## Payloads

### `user.message_received`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["source", "author_id", "text"],
  "properties": {
    "source": { "type": "string", "enum": ["cli", "github", "desktop", "web", "chat_connector"] },
    "author_id": { "type": "string" },
    "conversation_id": { "type": "string" },
    "external_ref": { "type": "string" },
    "text": { "type": "string", "minLength": 1 },
    "attachments": { "type": "array", "items": { "type": "object" } }
  }
}
```

### `task.created`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["title", "description", "priority", "risk_level", "acceptance_criteria"],
  "properties": {
    "title": { "type": "string" },
    "description": { "type": "string" },
    "priority": { "type": "string", "enum": ["P0", "P1", "P2", "P3"] },
    "risk_level": { "type": "string", "enum": ["low", "medium", "high", "critical"] },
    "acceptance_criteria": { "type": "array", "items": { "type": "string" } },
    "non_goals": { "type": "array", "items": { "type": "string" } }
  }
}
```

### `task.graph_created`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["graph_id", "graph_version", "nodes", "edges"],
  "properties": {
    "graph_id": { "type": "string" },
    "graph_version": { "type": "integer", "minimum": 1 },
    "nodes": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["id", "title", "objective", "acceptance_criteria", "owned_paths"],
        "properties": {
          "id": { "type": "string" },
          "title": { "type": "string" },
          "objective": { "type": "string" },
          "agent_profile": { "type": "string" },
          "owned_paths": { "type": "array", "items": { "type": "string" } },
          "read_paths": { "type": "array", "items": { "type": "string" } },
          "acceptance_criteria": { "type": "array", "items": { "type": "string" } },
          "validation_plan": { "type": "array", "items": { "type": "string" } }
        }
      }
    },
    "edges": {
      "type": "array",
      "items": {
        "type": "object",
        "additionalProperties": false,
        "required": ["from", "to", "type"],
        "properties": {
          "from": { "type": "string" },
          "to": { "type": "string" },
          "type": { "type": "string", "enum": ["blocks", "requires_artifact", "requires_review", "conflicts_with"] },
          "reason": { "type": "string" }
        }
      }
    }
  }
}
```

### `run.started`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["attempt", "triggered_by"],
  "properties": {
    "attempt": { "type": "integer", "minimum": 1 },
    "triggered_by": { "type": "string" },
    "work_unit_id": { "type": "string" },
    "agent_id": { "type": "string" }
  }
}
```

### `agent.started`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["agent_session_id", "runtime_type", "capabilities"],
  "properties": {
    "agent_session_id": { "type": "string" },
    "runtime_type": { "type": "string", "enum": ["codex_cli", "fake", "external"] },
    "capabilities": { "type": "array", "items": { "type": "string" } },
    "toolset_snapshot_id": { "type": "string" }
  }
}
```

### `agent.heartbeat`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["agent_session_id", "status"],
  "properties": {
    "agent_session_id": { "type": "string" },
    "status": { "type": "string", "enum": ["starting", "running", "waiting_approval", "paused", "stopping"] },
    "last_checkpoint_id": { "type": "string" },
    "ledger_summary": { "type": "string" }
  }
}
```

### `agent.checkpoint_reached`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["checkpoint_id", "current_goal", "ledger", "minimal_summary"],
  "properties": {
    "checkpoint_id": { "type": "string" },
    "current_goal": { "type": "string" },
    "completed_goals": { "type": "array", "items": { "type": "string" } },
    "ledger": {
      "type": "object",
      "required": ["todos", "completed", "blockers", "current_summary"],
      "properties": {
        "todos": { "type": "array", "items": { "type": "string" } },
        "completed": { "type": "array", "items": { "type": "string" } },
        "blockers": { "type": "array", "items": { "type": "string" } },
        "current_summary": { "type": "string" }
      }
    },
    "files_read": { "type": "array", "items": { "type": "string" } },
    "files_modified": { "type": "array", "items": { "type": "string" } },
    "evidence_refs": { "type": "array", "items": { "type": "string" } },
    "decisions": { "type": "array", "items": { "type": "string" } },
    "risks": { "type": "array", "items": { "type": "string" } },
    "minimal_summary": { "type": "string" },
    "next_goal_suggestion": { "type": "string" },
    "pending_questions": { "type": "array", "items": { "type": "string" } }
  }
}
```

### `agent.message`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["from", "to", "text"],
  "properties": {
    "from": { "type": "string" },
    "to": { "type": "string" },
    "text": { "type": "string" },
    "thread_id": { "type": "string" },
    "intent": { "type": "string", "enum": ["status", "question", "instruction", "warning", "summary"] }
  }
}
```

### `tool.requested`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["tool", "intent", "risk", "scope", "input"],
  "properties": {
    "tool": { "type": "string" },
    "intent": { "type": "string" },
    "risk": { "type": "string", "enum": ["safe", "guarded", "approval_required", "destructive", "forbidden"] },
    "scope": { "type": "string" },
    "input": { "type": "object" },
    "expected_output": { "type": "string" }
  }
}
```

### `toolset.change_requested`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["requested_tool", "reason", "scope", "risk", "fallback_attempted"],
  "properties": {
    "requested_tool": { "type": "string" },
    "reason": { "type": "string" },
    "scope": { "type": "string" },
    "risk": { "type": "string", "enum": ["safe", "guarded", "approval_required", "destructive", "forbidden"] },
    "fallback_attempted": { "type": "string" },
    "impact_if_denied": { "type": "string" }
  }
}
```

### `prompt.dynamic_fragment_created`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["fragment_id", "kind", "reason", "body_hash", "expires_after_run"],
  "properties": {
    "fragment_id": { "type": "string" },
    "kind": { "type": "string" },
    "reason": { "type": "string" },
    "body_hash": { "type": "string" },
    "expires_after_run": { "type": "boolean" },
    "created_for_work_unit_id": { "type": "string" }
  }
}
```

### `toolset.snapshot_created`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["toolset_snapshot_id", "agent_session_id", "tools", "created_reason"],
  "properties": {
    "toolset_snapshot_id": { "type": "string" },
    "agent_session_id": { "type": "string" },
    "tools": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["name", "scope", "risk"],
        "properties": {
          "name": { "type": "string" },
          "scope": { "type": "string" },
          "risk": { "type": "string" }
        }
      }
    },
    "created_reason": { "type": "string" }
  }
}
```

### `agent.session_reconfigured`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["previous_agent_session_id", "next_agent_session_id", "reason", "ledger_preserved"],
  "properties": {
    "previous_agent_session_id": { "type": "string" },
    "next_agent_session_id": { "type": "string" },
    "reason": { "type": "string" },
    "ledger_preserved": { "type": "boolean" },
    "prompt_snapshot_id": { "type": "string" },
    "toolset_snapshot_id": { "type": "string" },
    "added_tools": { "type": "array", "items": { "type": "string" } },
    "removed_tools": { "type": "array", "items": { "type": "string" } },
    "added_fragments": { "type": "array", "items": { "type": "string" } }
  }
}
```

### `tool.approved` e `tool.denied`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["tool_request_id", "decision", "reason", "decider"],
  "properties": {
    "tool_request_id": { "type": "string" },
    "decision": { "type": "string", "enum": ["approved", "denied"] },
    "reason": { "type": "string" },
    "decider": { "type": "string" },
    "expires_at": { "type": "string", "format": "date-time" },
    "constraints": { "type": "object" }
  }
}
```

### `tool.completed`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["tool_request_id", "status"],
  "properties": {
    "tool_request_id": { "type": "string" },
    "status": { "type": "string", "enum": ["completed", "failed", "cancelled"] },
    "exit_code": { "type": "integer" },
    "output_ref": { "type": "string" },
    "summary": { "type": "string" },
    "error": { "type": "string" }
  }
}
```

### `artifact.created`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["artifact_id", "kind", "uri", "summary"],
  "properties": {
    "artifact_id": { "type": "string" },
    "kind": { "type": "string", "enum": ["diff", "patch", "test_report", "log_bundle", "prompt_snapshot", "task_summary", "review_note"] },
    "uri": { "type": "string" },
    "checksum": { "type": "string" },
    "summary": { "type": "string" }
  }
}
```

### `validation.completed`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["status", "commands"],
  "properties": {
    "status": { "type": "string", "enum": ["passed", "failed", "skipped"] },
    "commands": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["command", "exit_code"],
        "properties": {
          "command": { "type": "string" },
          "exit_code": { "type": "integer" },
          "summary": { "type": "string" }
        }
      }
    },
    "skip_reason": { "type": "string" }
  }
}
```

### `task.completed`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["summary", "evidence_refs", "remaining_risks"],
  "properties": {
    "summary": { "type": "string" },
    "evidence_refs": { "type": "array", "items": { "type": "string" } },
    "remaining_risks": { "type": "array", "items": { "type": "string" } },
    "diff_ref": { "type": "string" },
    "validation_ref": { "type": "string" }
  }
}
```

### `task.failed`

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["reason", "recoverable"],
  "properties": {
    "reason": { "type": "string" },
    "recoverable": { "type": "boolean" },
    "failure_type": {
      "type": "string",
      "enum": ["agent_error", "policy_denied", "validation_failed", "infra_error", "timeout", "cancelled"]
    },
    "evidence_refs": { "type": "array", "items": { "type": "string" } }
  }
}
```

## Comandos

Comandos usam o mesmo envelope base, com `type` de comando.

Payload minimo para comandos de run:

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["reason"],
  "properties": {
    "reason": { "type": "string" },
    "requested_by": { "type": "string" }
  }
}
```

Payload para `message.interrupt`:

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["text", "target"],
  "properties": {
    "target": { "type": "string", "enum": ["orchestrator", "agent"] },
    "text": { "type": "string" },
    "requires_response": { "type": "boolean" }
  }
}
```

Payload para `policy.updated`:

```json
{
  "type": "object",
  "additionalProperties": false,
  "required": ["policy_id", "version", "summary"],
  "properties": {
    "policy_id": { "type": "string" },
    "version": { "type": "string" },
    "summary": { "type": "string" },
    "effective_immediately": { "type": "boolean" }
  }
}
```

## Regras De Evolução

- Mudanca compativel incrementa apenas schema interno.
- Mudanca quebrante exige novo `version`.
- Eventos ja persistidos nunca devem ser reescritos.
- Se payload bruto precisar ser preservado, guardar como artifact ou campo de diagnostico separado.
