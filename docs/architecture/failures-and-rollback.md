# Falhas e Rollback

## Objetivo

Definir como o OrchestraOS lida com falhas sem perder auditoria, evidencias ou controle operacional.

## Tipos de Resposta

| Tipo | Uso |
| --- | --- |
| `abort` | Parar a run antes de causar efeito externo. |
| `retry` | Reexecutar etapa idempotente. |
| `resume` | Continuar apos reconexao ou aprovacao. |
| `replan` | Criar nova versao do task graph. |
| `revert` | Desfazer mudanca versionada por git. |
| `compensate` | Aplicar acao corretiva quando nao ha rollback perfeito. |
| `quarantine` | Isolar artefato, sandbox, segredo ou dependencia suspeita. |

## Princípios

- Preferir falhar fechado: bloquear acao duvidosa.
- Nunca apagar evidencias necessarias para diagnostico.
- Rollback deve preservar historico, nao esconder que a falha ocorreu.
- Mudancas externas exigem plano de reversao ou compensacao.
- Runs devem ser idempotentes quando possivel.

## Falhas Comuns

### Agente Entra em Loop

Deteccao:

- `max_steps` excedido;
- tempo maximo excedido;
- muitos checkpoints sem progresso no ledger;
- repeticao de tool requests equivalentes;
- ausencia de novo artefato ou diff.

Resposta:

1. pausar run;
2. registrar evento `agent.loop_detected`;
3. resumir estado atual;
4. pedir decisao humana ou replanejar;
5. manter worktree e logs para analise.

### WebSocket Cai

Resposta:

1. marcar sessao como `disconnected`;
2. manter run em estado recuperavel;
3. agente reconecta com `last_seen_event_id`;
4. Orchestrator reenvia comandos pendentes;
5. se timeout expirar, pausar ou falhar run.

### Tool Request Negado

Resposta:

1. registrar negacao e motivo;
2. enviar comando ao agente pedindo alternativa;
3. se nao houver alternativa, marcar work unit como bloqueada;
4. preservar justificativa no resumo final.

### Conflito Entre Worktrees

Resposta:

1. detectar conflito antes do merge;
2. bloquear integracao automatica;
3. pedir rebase/merge assistido em worktree separado;
4. exigir review humana;
5. nunca sobrescrever mudanca aprovada sem decisao explicita.

### Teste Falha

Resposta:

1. registrar comando, exit code e resumo;
2. permitir uma tentativa de correcao dentro do limite da run;
3. se persistir, concluir como `failed` ou `needs_human_review`;
4. nao marcar task como concluida sem justificativa.

### Sandbox Falha

Resposta:

1. parar container se ainda estiver ativo;
2. coletar logs;
3. preservar worktree se houver diff;
4. limpar recursos temporarios seguros;
5. marcar run como falha de infraestrutura se nao houve erro do agente.

### Segredo Exposto

Resposta:

1. pausar novas runs;
2. revogar ou rotacionar segredo;
3. redigir logs e artefatos quando possivel;
4. registrar incidente;
5. revisar politica que permitiu acesso.

### GitHub ou Conector Externo Indisponível

Resposta:

1. persistir evento e outbox;
2. retry com backoff;
3. nao perder estado interno;
4. mostrar pendencia na CLI;
5. enviar notificacao quando conector voltar.

## Rollback Por Git

No MVP, o principal rollback de codigo e nao integrar a branch da task.

Fluxo:

1. agente trabalha em branch e worktree isolados;
2. diff e validacoes sao revisados;
3. se rejeitado, branch permanece como evidencia ou e descartada conforme politica;
4. se aprovado e mergeado, reversao posterior deve usar commit de revert ou PR de rollback.

Nunca usar rollback destrutivo que apague trilha de auditoria.

## Rollback De Dados

Para banco de dados:

- usar migrations versionadas;
- preferir mudancas expand-contract;
- fazer backup antes de migration destrutiva;
- testar rollback ou migration compensatoria;
- registrar versao aplicada no Event Store.

No MVP local, migrations destrutivas exigem aprovacao.

## Rollback De Planejamento

Se o plano estiver errado:

1. cancelar ou pausar work units afetadas;
2. criar nova versao do `TaskGraph`;
3. preservar versao antiga;
4. explicar motivo do replanejamento;
5. reemitir prompts para novas work units.

## Checklist De Falha Crítica

- Pausar novas runs.
- Preservar logs e artefatos.
- Identificar escopo afetado.
- Bloquear ferramentas relacionadas.
- Revogar segredos se necessario.
- Criar resumo de incidente.
- Definir acao corretiva.
- Retomar apenas com politica atualizada.
