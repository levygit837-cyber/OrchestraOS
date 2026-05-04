# ADR 0016: State Machine Event-Sourced

## Contexto

O M0 ja possuia persistencia de `Task`, `WorkUnit`, `Run`, `AgentSession` e eventos, mas as transicoes de status ainda podiam ser feitas diretamente pelos repositorios ou CLI. Isso deixava risco de estado relacional divergir do Event Store, dificultava replay e permitia transicoes invalidas.

O projeto tambem definiu que o Event Store e o historico operacional canonico, enquanto tabelas relacionais devem servir como projecoes consultaveis.

## Decisao

O OrchestraOS adotara uma State Machine event-sourced para `Task`, `WorkUnit`, `Run` e `AgentSession`.

- Eventos continuam sendo a fonte canonica do historico operacional.
- Tabelas relacionais continuam existindo como read models/projecoes para consulta rapida.
- Transicoes de status devem passar por servicos internos de comando.
- Cada comando valida a transicao, grava o evento e atualiza a projecao dentro da mesma transacao.
- Reducers puros em Go reconstroem estado por replay de eventos.
- O replay usado por CLI e por `ReplayState`/`ReplayRunState` e estrito: inconsistencias historicas ou transicoes invalidas retornam erro, em vez de serem normalizadas silenciosamente.
- Transicoes para `completed` exigem evidencia de validacao ou justificativa.
- O Event Store controla `sequence`; callers nao devem definir ordenacao operacional.
- `event_id` e idempotente: reprocessar o mesmo evento nao cria duplicata.
- A retomada neste corte sera feita por replay de eventos e consulta do ultimo `agent.checkpoint_reached`.

Eventos operacionais essenciais passam a ter validacao minima de payload:

- `agent.ledger_updated`
- `agent.checkpoint_reached`
- `artifact.created`
- `validation.completed`
- `prompt.snapshot_created`
- `toolset.snapshot_created`

Nao serao criadas tabelas de snapshots agregados neste corte. Checkpoints permanecem eventos estruturados e consultaveis.

## Consequencias

- O sistema bloqueia transicoes invalidas antes de alterar read models.
- Replay passa a reconstruir estado esperado, nao apenas listar eventos, e falha quando o historico invalido nao pode ser projetado com seguranca.
- Falha ao persistir evento impede a atualizacao da projecao na mesma transacao.
- O CLI e runtimes devem usar servicos de comando para transicoes, evitando updates diretos de status.
- Sequencias podem ter lacunas quando eventos invalidos ou duplicados consomem `nextval`, o que e aceitavel porque a garantia necessaria e ordenacao monotonicamente crescente.
- Persistencia operacional essencial fica auditavel sem criar snapshots agregados prematuros.

## Alternativas consideradas

- **Manter updates diretos nos repositorios**: simples, mas permitia divergencia entre status e eventos.
- **State Machine somente no banco**: bloquearia parte dos erros, mas reduziria testabilidade das regras de negocio.
- **Snapshots agregados desde agora**: facilitaria retomada rapida, mas aumentaria escopo antes de validar replay e checkpoints.
- **Dependencia externa de FSM**: reduziria codigo proprio, mas adicionaria acoplamento desnecessario para o escopo atual.
