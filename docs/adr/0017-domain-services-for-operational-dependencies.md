# ADR 0017: Servicos de Dominio para Dependencias Operacionais

## Contexto

O OrchestraOS ja possui contratos, persistencia inicial e Event Store para entidades M0 como `Task`, `Run`, `WorkUnit`, `AgentSession` e `Event`.

A ADR 0014 definiu que a CLI deve continuar fina e que regras de negocio, validacao de contratos, transicoes de estado e consistencia de auditoria devem migrar para servicos internos quando crescerem. A ADR 0016 definiu que transicoes de status devem passar por servicos internos de comando, gravando evento e atualizando projecao na mesma transacao.

Com a evolucao do MVP, operacoes como criar task, decompor work units, iniciar run, concluir execucao, registrar evento, aplicar retry, lidar com timeout e atualizar estados relacionados nao devem ficar espalhadas entre CLI, repositorios, runtime de agente e chamadas diretas ao banco. Essa dispersao aumentaria o risco de:

- transicoes invalidas;
- eventos persistidos sem read model correspondente;
- read models atualizados sem evento canonico;
- retries duplicarem efeitos;
- timeouts deixarem estado parcialmente escrito;
- validacoes divergirem entre interfaces;
- comandos concluirem com sucesso mesmo quando auditoria ou persistencia falhar.

## Decisao

O OrchestraOS adotara uma camada de servicos internos de dominio para coordenar comandos operacionais sobre `Task`, `Run`, `WorkUnit`, `AgentSession` e eventos.

No escopo inicial, os servicos aprovados sao:

- `TaskService`
- `RunService`
- `WorkUnitService`
- `AgentSessionService`
- `EventService`

Esses servicos devem ser a fronteira obrigatoria para comandos que alteram estado de dominio. CLI, TUI, GitHub, runtimes de agente e conectores futuros devem chamar servicos, nao repositorios diretamente, quando a operacao tiver regra de negocio, transicao de estado ou auditoria obrigatoria.

Repositorios continuam responsaveis apenas por acesso a dados. O Event Store continua sendo o historico operacional canonico. O `EventService` nao substitui o Event Store; ele coordena validacao, append, idempotencia, consulta operacional e integracao transacional com outros servicos quando necessario.

### Responsabilidades iniciais

`TaskService` deve coordenar criacao, triagem, planejamento, agendamento, pausa, conclusao, falha e cancelamento de tasks, incluindo validacao de entrada, transicoes permitidas e eventos correspondentes.

`WorkUnitService` deve coordenar criacao e atualizacao de work units, validacao de dependencias, bloqueios, atribuicao, conclusao, falha e cancelamento, preservando o DAG aciclico e os criterios de aceite.

`RunService` deve coordenar inicio, retomada, retry, timeout, conclusao, falha e cancelamento de runs, incluindo tentativas, limites operacionais, evidencias de validacao e estados relacionados de work unit ou task quando aplicavel.

`AgentSessionService` deve coordenar criacao, conexao, heartbeat, checkpoint, desconexao, retomada, encerramento, timeout e falha de sessoes de agente, incluindo compatibilidade com a run associada, sandbox ativo, ultimo evento visto e estado recuperavel da sessao.

`EventService` deve coordenar comandos de append, validacao de envelope e payload, idempotencia por `event_id`, consulta operacional e replay orientado a servicos, mantendo compatibilidade com reducers puros e com o Event Store existente.

### Persistencia

Todos os servicos devem persistir os efeitos relevantes das suas operacoes.

Operacoes que mudam estado devem registrar evento canonico e atualizar read model/projecao relacional quando houver projecao aplicavel. A ausencia de persistencia de evento deve impedir a atualizacao da projecao correspondente.

Atualizacoes diretas de status por CLI, runtime ou repositorio nao devem ser usadas para fluxos operacionais. Repositorios podem expor primitivas de leitura e escrita, mas nao devem decidir transicoes, retry, timeout, conclusao ou compensacao.

### Validacao de entrada

Servicos devem validar entradas nas bordas do sistema antes de executar efeitos.

A validacao minima inclui:

- campos obrigatorios e formatos de identificadores;
- valores aceitos para status, prioridade, risco e tipo de evento;
- existencia das entidades referenciadas;
- compatibilidade entre `task_id`, `work_unit_id`, `run_id` e `agent_session_id`;
- compatibilidade entre sessao de agente, run, agente, sandbox e conexao ativa;
- dependencias aciclicas entre work units;
- criterios de aceite e evidencias exigidas para conclusao;
- payloads de eventos conforme schemas versionados;
- politicas de autonomia e risco quando a operacao puder executar ou desbloquear acao de agente.

Falhas de validacao devem retornar erro tipado e nao devem disparar retries automaticos.

### Transicoes de estado

Transicoes de estado de `Task`, `WorkUnit`, `Run` e `AgentSession` devem ser decididas pelos servicos, usando state machines explicitas e testaveis.

Cada transicao deve:

- validar estado atual e estado desejado;
- validar pre-condicoes de dominio;
- persistir evento correspondente;
- atualizar a projecao dentro da mesma transacao quando houver mais de uma escrita relacionada;
- preservar timestamps e evidencias anteriores;
- retornar erro quando o historico ou o estado atual nao permitir a transicao.

Conclusoes para estados finais como `completed`, `failed` ou `cancelled` devem registrar motivo, evidencia ou justificativa suficiente para auditoria.

### Retrys, timeouts e erros

Servicos devem usar erros tipados para diferenciar, no minimo:

- entrada invalida;
- transicao invalida;
- conflito de concorrencia ou idempotencia;
- violacao de politica;
- falha de persistencia;
- falha de runtime ou conector externo;
- timeout;
- falha interna inesperada.

Retries automaticos so devem ser usados para operacoes idempotentes e falhas transitorias, como indisponibilidade temporaria de banco, runtime ou conector externo. Erros de validacao, politica e transicao invalida nao devem ser retentados automaticamente.

Operacoes com retry devem ter:

- limite de tentativas;
- timeout por tentativa;
- timeout total da operacao;
- backoff progressivo;
- registro de tentativas relevantes por evento ou metadado auditavel;
- chave de idempotencia quando houver risco de duplicar efeito.

Timeouts devem cancelar o trabalho em andamento por contexto de execucao quando possivel. Se um timeout ocorrer depois de efeito parcial externo, o servico deve registrar estado recuperavel e exigir retry idempotente, compensacao ou decisao humana conforme o risco.

### Atomicidade

Quando uma operacao envolver mais de uma escrita relacionada, as escritas internas devem ocorrer de forma atomica.

Exemplos:

- criar task e registrar `task.created`;
- criar work units e registrar eventos da decomposicao;
- iniciar run e atualizar status da work unit relacionada;
- conectar sessao de agente e registrar evento de conexao;
- registrar checkpoint e atualizar `last_checkpoint_at` da sessao;
- registrar heartbeat e atualizar `last_heartbeat_at` da sessao;
- desconectar sessao por timeout e marcar run como recuperavel quando aplicavel;
- concluir run, registrar validacao e atualizar work unit ou task;
- registrar evento e atualizar projecao relacional;
- cancelar task e cancelar work units/runs pendentes relacionadas.

O padrao preferido e uma transacao de banco envolvendo Event Store e projecoes. Quando a operacao envolver efeito externo que nao participa da transacao, o servico deve usar evento de intencao, outbox, confirmacao posterior ou acao compensatoria. O sistema nao deve reportar sucesso operacional quando auditoria ou persistencia obrigatoria falhar.

## Consequencias

- A CLI e interfaces futuras ficam mais finas e consistentes.
- Regras de negocio passam a ter uma fronteira testavel, em vez de ficarem espalhadas por comandos e repositorios.
- O Event Store permanece canonico, mas a atualizacao de read models fica coordenada com transicoes de dominio.
- Retries e timeouts passam a ser politicas explicitas por servico, reduzindo duplicidade e estados parciais.
- Sessoes de agente passam a ter ciclo de vida proprio, sem esconder regras de conexao, heartbeat e retomada dentro de runs ou runtimes.
- Fluxos com multiplas escritas ganham atomicidade e melhor diagnostico de falha.
- Testes de unidade e integracao devem cobrir validacao, transicoes permitidas e proibidas, idempotencia, retries, timeouts e rollback transacional.
- A implementacao inicial fica um pouco mais estruturada, mas evita que a CLI ou os repositorios se tornem o local implicito de orquestracao.

Riscos e cuidados:

- Evitar criar um servico generico de CRUD que apenas repasse chamadas para repositorios.
- Evitar duplicar regras entre reducers, servicos e banco; reducers reconstroem estado, servicos decidem comandos.
- Manter `EventService` pequeno e alinhado ao Event Store existente.
- Nao introduzir dependencias externas de workflow, retry ou state machine sem necessidade concreta.
- Manter a fronteira entre `RunService` e `AgentSessionService` clara: runs representam tentativas de execucao; sessoes representam a conexao operacional viva ou recuperavel de um agente nessa run.

## Alternativas consideradas

- **Manter logica na CLI e nos repositorios**: seria mais rapido no curto prazo, mas manteria risco de divergencia entre eventos, estados e read models.
- **Usar apenas constraints e transacoes SQL**: bloquearia algumas inconsistencias, mas deixaria regras de dominio menos testaveis e espalhadas em migrations.
- **Criar um servico generico para todas as entidades**: reduziria arquivos, mas enfraqueceria a linguagem do dominio e misturaria regras distintas de task, work unit, run e evento.
- **Adotar workflow engine agora**: poderia resolver retries e durabilidade, mas adicionaria complexidade antes de validar os fluxos centrais do MVP.
- **Usar somente Event Store sem projecoes transacionais**: simplificaria escrita canonica, mas prejudicaria consultas operacionais e deixaria a CLI/TUI dependentes de replay para todo estado consultavel.
