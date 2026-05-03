# ADR 0012: Sistema De Memoria Recursiva

## Contexto

Agentes longos e paralelos precisam recuperar contexto util de sessoes anteriores, runs em andamento, diffs, validacoes, ferramentas usadas, checkpoints e decisoes. Sem uma camada de memoria, cada nova interacao depende demais do prompt inicial, do transcript recente ou de leitura manual dos documentos.

Foi proposta uma memoria recursiva em que informacoes relevantes geradas pelos agentes sao salvas, categorizadas, indexadas por embeddings e devolvidas como contexto durante execucoes futuras ou durante a propria run.

O projeto ja decidiu que:

- o repositório e fonte de verdade;
- o Event Store e o historico operacional canonico;
- o Agent Task Ledger e memoria operacional viva da work unit;
- checkpoints sao snapshots estruturados de progresso;
- prompts sao montados por fragmentos versionados e registrados em `PromptSnapshot`.

A memoria recursiva precisa respeitar essas decisoes. Ela nao pode virar fonte definitiva paralela nem permitir que informacao sem evidencia substitua ADR, canvas, issue, PR, ledger ou evento canonico.

ADR 0009 e esta ADR tratam de informacoes parecidas, mas em camadas diferentes. ADR 0009 define o historico operacional canonico para o Orchestrator. Esta ADR define uma projecao derivada, semantica e deduplicada para recuperar contexto util para agentes.

## Decisão

O OrchestraOS adotara **Recursive Memory** como capacidade arquitetural futura.

Recursive Memory sera uma camada derivada do Event Store, documentos versionados, checkpoints, ledgers, artefatos, tool executions, validacoes e decisoes humanas. Ela sera usada para recuperar contexto relevante e injeta-lo em agentes por meio do Orchestrator e do Prompt Composer.

A memoria recursiva tera estes principios:

- memoria nao e fonte canonica;
- toda memoria precisa de evidencia rastreavel;
- memorias devem ser classificadas por tipo, dominio, escopo, impacto, confianca e validade;
- deduplicacao semantica e obrigatoria antes de persistir ou injetar memoria;
- memorias com textos diferentes, mas mesma proposicao canonica, devem ser mescladas no mesmo cluster;
- o mesmo bundle nao deve ser repetido inutilmente na mesma run;
- recuperacao deve respeitar permissao, autonomia, ownership de caminhos e isolamento entre tasks;
- memoria recuperada nao pode sobrescrever politicas, ADRs, criterios de aceite ou instrucoes do repositorio;
- o servico deve ser assincrono e nao pode bloquear a execucao do agente em caso de falha;
- embeddings e analise semantica devem ser adicionados depois que Event Store, checkpoints, ledger e PromptSnapshot estiverem funcionando.

O sistema deve iniciar com recuperacao estruturada simples, baseada em ADRs, canvas e checkpoints. Busca vetorial, embeddings em background e consolidacao por chunks entram depois.

## Relação Com ADR 0009

| Camada | Fonte de verdade? | Consumidor principal | Forma | Finalidade |
| --- | --- | --- | --- | --- |
| ADR 0009: Event Store e tracing | Sim, para historico operacional. | Orchestrator. | Eventos ordenados, logs, spans, artefatos e checkpoints. | Auditoria, replay, diagnostico, live view, politica e recuperacao de run. |
| ADR 0012: Recursive Memory | Nao, e derivada. | Agentes via Orchestrator e Prompt Composer. | Memorias pequenas, classificadas, deduplicadas, com evidencias. | Recuperar contexto reutilizavel e reduzir dependencia de transcript completo. |

Regra pratica:

- tracing guarda eventos mesmo que nunca virem memoria;
- memoria so guarda o que passou por classificacao, deduplicacao e evidencia;
- tracing e granular e historico;
- memoria e seletiva e semantica;
- tracing e otimizado para o Orchestrator reconstruir e controlar;
- memoria e otimizada para o agente receber contexto util.

## Funcionamento Aprovado

O fluxo alvo sera:

1. Eventos relevantes entram no Event Store.
2. Memory Intake consome eventos, checkpoints, ledgers, diffs, validacoes e decisoes.
3. Candidate Extractor cria candidatos pequenos de memoria.
4. Classifier filtra ruido, segredos e informacoes de baixo valor.
5. Semantic Dedupe compara hash, origem, escopo, proposicao canonica, cluster e similaridade.
6. O candidato e rejeitado, mesclado, usado para superseder memoria antiga ou aprovado como memoria nova.
7. Memory Store persiste `MemoryRecord` com evidencia canonica.
8. Retriever consulta memorias no inicio da AgentSession e em checkpoints relevantes.
9. Prompt Composer cria `RetrievedMemoryBundle`.
10. Bundle e injetado no agente com ids e evidencias.
11. Memory Audit Log registra criacao, recuperacao, injecao, rejeicao, merge, supersessao e feedback.

## Escopo Inicial

Nao entra no primeiro MVP.

Pre-requisitos:

- Event Store persistente;
- Agent Task Ledger;
- Agent Checkpoints;
- PromptSnapshot;
- Artifact Manager;
- politica minima de permissao e escopo.

Primeiro corte futuro:

- `MemoryRecord` relacional em Postgres;
- `semantic_key` e `cluster_id` para deduplicacao semantica;
- ingestao apenas de ADRs, canvas e checkpoints;
- recuperacao por filtros estruturados;
- registro de bundles injetados;
- sem dependencia obrigatoria de vector database.

Segundo corte futuro:

- embeddings em lote;
- busca semantica hibrida;
- consolidacao por chunks, como 20k e 40k tokens;
- ranking por recencia, confianca, fonte e correspondencia de paths;
- supersessao e expiracao automatica.

## Consequências

- Agentes ganham continuidade sem depender de transcript completo.
- O Orchestrator consegue retornar contexto util com menor custo de tokens.
- Runs futuras podem aproveitar aprendizados de diffs, validacoes e falhas anteriores.
- A memoria melhora planejamento, revisao e implementacao repetida em dominios recorrentes.
- O sistema ganha mais um componente operacional, exigindo testes, auditoria e politicas de acesso.
- Memoria stale, duplicada ou sem evidencia pode degradar agentes; por isso dedupe, escopo e evidencia sao obrigatorios.
- A latencia deve ser controlada com background jobs, batching e timeout de retrieval.
- O custo de embeddings deve ser tratado como otimizacao posterior, nao fundacao do MVP.

## Alternativas consideradas

- **Sem memoria recursiva**: simples e alinhado ao MVP, mas limita continuidade e reaproveitamento de conhecimento entre runs.
- **Salvar apenas transcript completo**: facil, mas caro, ruidoso e fraco para busca, auditoria e controle de contexto.
- **Vector store como fonte de verdade**: melhora busca semantica, mas viola a decisao de repositorio/Event Store como fontes canonicas.
- **Embedding em tempo real para toda mensagem e tool call**: contexto fresco, mas aumenta custo, latencia e ruido.
- **Analise apenas por chunks grandes**: reduz custo, mas perde contexto operacional importante em checkpoints e eventos criticos.
- **Memoria livre controlada pelo agente**: flexivel, mas arriscada para deduplicacao, seguranca, isolamento e auditoria.
