# ADR 0018: Planner Local Heuristico Para Task Graph

## Contexto

O projeto precisa transformar `Task` em um grafo aciclico persistente de `WorkUnit`s. A arquitetura ja decidiu que o grafo deve ser um DAG e que loops de tentativa pertencem a `Run` ou `AgentLoop`, nao as dependencias entre nodes.

Tambem existe risco em introduzir um planner por LLM cedo demais: ele aumentaria variabilidade, exigiria prompt versionado, fallback, politica propria e validacao mais ampla antes do nucleo de persistencia estar estabilizado.

## Decisao

O primeiro corte executavel usara um planner local deterministico baseado em `Task.acceptance_criteria`.

- O planner exige pelo menos dois criterios de aceite.
- O planner gera de 2 a 5 `WorkUnit`s.
- O balanceamento usa peso textual dos criterios.
- O maior peso de uma work unit nao pode ultrapassar 1.5x o menor peso.
- Dependencias sao independentes por padrao.
- Dependencias explicitas usam marcador inicial `[after: 1,2]` nos criterios de aceite.
- Ciclos e referencias a criterios inexistentes rejeitam o graph antes da persistencia.
- `task.graph_created` e o evento canonico do planejamento e seu payload inclui `task_id` consistente com o envelope.
- `task_graphs` e a projecao relacional versionada.
- Apenas um graph por task fica `active`; replanejamento supersede o ativo anterior e cria nova versao.
- `work_units.task_id` referencia a Task real e `work_units.task_graph_id` referencia a versao do graph.
- Graphs planejados nao aceitam criacao manual posterior de `WorkUnit`; alteracoes exigem replanejamento.

## Consequencias

- O sistema passa a ter decomposicao automatica testavel e reproduzivel.
- A validacao de DAG e persistencia ficam independentes de LLM.
- A qualidade semantica da decomposicao depende da qualidade dos criterios de aceite.
- Tasks sem criterios suficientes sao rejeitadas em vez de gerar work units fracas.
- Um planner por LLM pode ser adicionado depois como nova `planner_strategy`, produzindo o mesmo contrato validado.

## Alternativas consideradas

- **LLM Planner agora**: melhor potencial semantico, mas maior custo operacional e superficie de falha.
- **Receber plano externo e apenas validar**: menor escopo, mas nao entrega decomposicao automatica.
- **Lista linear de work units**: simples, mas perde paralelismo e nao representa dependencias reais.
