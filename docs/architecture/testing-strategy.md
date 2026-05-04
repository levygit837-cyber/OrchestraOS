# Estratégia de Testes

## Princípio

Testes devem aumentar confianca real. O projeto deve evitar testes que apenas confirmam mocks, espelham a implementacao ou passariam mesmo com comportamento incorreto.

Um teste util deve falhar se:

- a regra de negocio for quebrada;
- a politica permitir algo proibido;
- o evento ficar invalido;
- o agente perder estado;
- o rollback nao preservar evidencia;
- o Orchestrator nao conseguir reconstruir a run.

## Pirâmide Recomendada

| Camada | Objetivo |
| --- | --- |
| Unitarios | Regras puras: state machine, policy, validacao de schema, prompt selection. |
| Contrato | JSON Schema, envelopes, tool inputs/outputs, comandos e eventos. |
| Integracao | Postgres, WebSocket, Event Store, outbox, sandbox manager. |
| E2E local | Fluxo completo com repositorio temporario e agente fake ou Codex controlado. |
| Falha/recuperacao | Timeout, reconexao, tool denied, agente travado, conflito de arquivos. |

## Domínios e Testes

### Task State Machine

Cobrir:

- transicoes validas;
- transicoes proibidas;
- cancelamento;
- falha;
- retomada;
- conclusao com evidencias.

Exemplo de risco:

- `completed` sem validacao ou justificativa nao deve ser aceito.

### Task Graph

Cobrir:

- deteccao de ciclos;
- dependencias ausentes;
- paralelismo permitido;
- conflito de ownership;
- replanejamento versionado.

Exemplo de risco:

- duas work units com escrita no mesmo arquivo nao devem rodar em paralelo sem lock ou aprovacao.

### Policy Engine

Cobrir com testes tabulares:

- ferramenta;
- escopo;
- autonomia;
- risco;
- destino;
- decisao esperada.

Exemplo de risco:

- `git push` nao pode ser autoaprovado no Nivel 2.

### Prompt Composer

Cobrir:

- selecao de fragmentos obrigatorios;
- resolucao de conflito;
- precedencia de politica;
- snapshot e hash;
- ausencia de fragmento obrigatorio;
- prompt final contendo criterios de aceite.

Evitar:

- testar apenas se uma string fixa foi concatenada. O teste deve verificar invariantes.

### Event Store

Cobrir:

- validacao de envelope antes da persistencia;
- preenchimento automatico de `id`, `sequence` e `created_at` antes da validacao;
- `run_id` opcional para eventos de task/work unit e obrigatorio para eventos de runtime;
- idempotencia por `event_id`;
- ordenacao por `sequence`;
- consulta por `task_id` e `run_id`;
- replay de estado;
- replay estrito que falha em historicos com transicoes invalidas;
- persistencia de comandos pendentes.

### WebSocket

Cobrir:

- handshake;
- heartbeat;
- checkpoint;
- reconexao com `last_seen_event_id`;
- entrega de comandos pendentes;
- desconexao inesperada.

### Agent Checkpoint

Cobrir:

- checkpoint emitido ao concluir goal curto;
- checkpoint com goal atual, arquivos lidos, arquivos modificados e evidencias;
- checkpoint persistido no Event Store;
- ultimo checkpoint reconstrui resumo minimo da sessao;
- checkpoint nao marca work unit como concluida sem criterios de aceite;
- excesso de checkpoints pequenos nao deve ser necessario para progresso normal.

### Sandbox Manager

Cobrir com repositorio temporario real:

- criacao de branch;
- criacao de worktree fora do repo principal;
- limites de escrita;
- limpeza ao final;
- retencao de artefatos;
- bloqueio de caminhos proibidos.

### Agent Runtime

Cobrir:

- agente fake deterministico para E2E rapido;
- agente fake propagando erro de persistencia, heartbeat, checkpoint e conclusao;
- agente real controlado para smoke test;
- max steps;
- timeout;
- loop detection;
- tool request;
- task ledger update.

### Recursive Memory

Cobrir quando a capacidade entrar no escopo:

- memoria criada apenas com evidencia canonica;
- candidato duplicado rejeitado;
- memoria com segredo rejeitada ou sanitizada;
- retrieval respeitando projeto, repositorio, dominio e paths;
- `RetrievedMemoryBundle` dentro do orcamento de tokens;
- mesma memoria nao injetada repetidamente na mesma run;
- memoria stale supersedida ou expirada;
- falha do servico de memoria sem interromper AgentSession.

### GitHub e Conectores Externos

Cobrir:

- adaptadores com contrato;
- idempotencia de notificacao;
- retry via outbox;
- falha externa sem perder estado interno.

Nao depender de GitHub real ou conectores de chat reais em testes unitarios.

## Testes Que Devem Ser Evitados

- Mock que retorna exatamente o que a implementacao espera sem validar contrato.
- Snapshot gigante de prompt sem invariantes.
- Teste que so verifica que uma funcao foi chamada.
- E2E que ignora exit code real.
- Teste que passa mesmo quando nenhuma validacao foi executada.

## Evidência de Teste

Toda run concluida deve registrar:

- comandos executados;
- exit codes;
- resumo de saida;
- arquivos de relatorio;
- justificativa quando teste relevante nao foi executado.

## Critério Para Aceitar Uma Mudança

Uma mudanca pode ser aceita quando:

- os testes relevantes foram executados;
- falhas conhecidas foram registradas;
- riscos restantes estao claros;
- evidencias sao rastreaveis por `task_id` e `run_id`.
