# ADR 0009: Observabilidade — Event Store, Tracing e Memória Recursiva

**Status:** Consolidated (absorve: ADR 0012)  
**Data original:** 2026-05-10  
**Última atualização:** 2026-05-17

## Contexto

O Orchestrator deve analisar o tracing dos agentes, entender progresso, detectar falhas, aprovar ferramentas e reconstruir o historico de uma run. Logs livres sao uteis para diagnostico, mas insuficientes como estrutura principal de auditoria.

O sistema tambem precisa preservar mensagens, comandos, tool calls, checkpoints, artefatos, validacoes e decisoes humanas em uma forma consultavel.

## Decisão

O Event Store sera a fonte canonica do historico operacional.

O tracing sera normalizado em entidades e eventos correlacionados por:

- `task_id`;
- `run_id`;
- `agent_id`;
- `work_unit_id`;
- `trace_id`;
- `span_id`;
- `parent_span_id`;
- `sequence`;
- `created_at`.

Eventos estruturados representam mudancas de estado, pedidos de ferramenta, comandos, checkpoints, mensagens e artefatos. Logs textuais e saidas volumosas devem ser armazenados como `LogEntry` ou `Artifact`, referenciados pelos eventos.

O desenho deve manter compatibilidade conceitual com OpenTelemetry, sem exigir OpenTelemetry completo no MVP.

## 1. Event Store e Tracing

O Event Store é a fonte canônica do histórico operacional.

O tracing é normalizado em entidades e eventos correlacionados por `task_id`, `run_id`, `agent_id`, `work_unit_id`, `trace_id`, `span_id`, `parent_span_id`, `sequence` e `created_at`.

Eventos estruturados representam mudanças de estado, pedidos de ferramenta, comandos, checkpoints, mensagens e artefatos. Logs textuais e saídas volumosas são armazenados como `LogEntry` ou `Artifact`, referenciados pelos eventos.

O desenho mantém compatibilidade conceitual com OpenTelemetry, sem exigir OpenTelemetry completo no MVP.

### 1.1 Consequências do Event Store

- O Orchestrator consegue reconstruir e analisar runs sem depender de conversas soltas.
- A live view pode ser derivada do Event Store e dos eventos em tempo real.
- Replay, auditoria e diagnóstico ficam possíveis desde o início.
- O MVP precisa tratar idempotência, ordenação e correlação de eventos como parte do núcleo.

### 1.2 Alternativas consideradas (Event Store)

- **Salvar apenas transcript do agente**: simples, mas fraco para consulta, replay e políticas.
- **Usar OpenTelemetry completo desde o início**: robusto, mas aumenta escopo do MVP.
- **Guardar logs por arquivo sem estrutura**: útil como evidência bruta, mas insuficiente para controle operacional.

---

## 2. Memória Recursiva (Derivada)

### 2.1 Contexto adicional

Agentes longos e paralelos precisam recuperar contexto útil de sessões anteriores, runs em andamento, diffs, validações, ferramentas usadas, checkpoints e decisões. Sem uma camada de memória, cada nova interação depende demais do prompt inicial, do transcript recente ou de leitura manual dos documentos.

O projeto já decidiu que: o repositório é fonte de verdade; o Event Store é o histórico operacional canônico; o Agent Task Ledger é memória operacional viva da work unit; checkpoints são snapshots estruturados de progresso; prompts são montados por fragmentos versionados e registrados em `PromptSnapshot`.

A memória recursiva respeita essas decisões. Ela não pode virar fonte definitiva paralela nem permitir que informação sem evidência substitua ADR, canvas, issue, PR, ledger ou evento canônico.

### 2.2 Decisão

O OrchestraOS adotará **Recursive Memory** como capacidade arquitetural futura.

Recursive Memory será uma camada derivada do Event Store, documentos versionados, checkpoints, ledgers, artefatos, tool executions, validações e decisões humanas. Será usada para recuperar contexto relevante e injetá-lo em agentes por meio do Orchestrator e do Prompt Composer.

Princípios:

- memória não é fonte canônica;
- toda memória precisa de evidência rastreável;
- memórias devem ser classificadas por tipo, domínio, escopo, impacto, confiança e validade;
- deduplicação semântica é obrigatória antes de persistir ou injetar memória;
- memórias com textos diferentes, mas mesma proposição canônica, devem ser mescladas no mesmo cluster;
- o mesmo bundle não deve ser repetido inutilmente na mesma run;
- recuperação deve respeitar permissão, autonomia, ownership de caminhos e isolamento entre tasks;
- memória recuperada não pode sobrescrever políticas, ADRs, critérios de aceite ou instruções do repositório;
- o serviço deve ser assíncrono e não pode bloquear a execução do agente em caso de falha;
- embeddings e análise semântica devem ser adicionados depois que Event Store, checkpoints, ledger e PromptSnapshot estiverem funcionando.

O sistema deve iniciar com recuperação estruturada simples, baseada em ADRs, canvas e checkpoints. Busca vetorial, embeddings em background e consolidação por chunks entram depois.

### 2.3 Relação com Event Store

| Camada | Fonte de verdade? | Consumidor principal | Forma | Finalidade |
| --- | --- | --- | --- | --- |
| Event Store e tracing | Sim, para histórico operacional. | Orchestrator. | Eventos ordenados, logs, spans, artefatos e checkpoints. | Auditoria, replay, diagnóstico, live view, política e recuperação de run. |
| Recursive Memory | Não, é derivada. | Agentes via Orchestrator e Prompt Composer. | Memórias pequenas, classificadas, deduplicadas, com evidências. | Recuperar contexto reutilizável e reduzir dependência de transcript completo. |

Regra prática:

- tracing guarda eventos mesmo que nunca virem memória;
- memória só guarda o que passou por classificação, deduplicação e evidência;
- tracing é granular e histórico;
- memória é seletiva e semântica;
- tracing é otimizado para o Orchestrator reconstruir e controlar;
- memória é otimizada para o agente receber contexto útil.

### 2.4 Funcionamento aprovado

Fluxo alvo:

1. Eventos relevantes entram no Event Store.
2. Memory Intake consome eventos, checkpoints, ledgers, diffs, validações e decisões.
3. Candidate Extractor cria candidatos pequenos de memória.
4. Classifier filtra ruído, segredos e informações de baixo valor.
5. Semantic Dedupe compara hash, origem, escopo, proposição canônica, cluster e similaridade.
6. O candidato é rejeitado, mesclado, usado para superseder memória antiga ou aprovado como memória nova.
7. Memory Store persiste `MemoryRecord` com evidência canônica.
8. Retriever consulta memórias no início da AgentSession e em checkpoints relevantes.
9. Prompt Composer cria `RetrievedMemoryBundle`.
10. Bundle é injetado no agente com ids e evidências.
11. Memory Audit Log registra criação, recuperação, injeção, rejeição, merge, supersessão e feedback.

### 2.5 Escopo inicial

Não entra no primeiro MVP.

Pré-requisitos: Event Store persistente; Agent Task Ledger; Agent Checkpoints; PromptSnapshot; Artifact Manager; política mínima de permissão e escopo.

Primeiro corte futuro: `MemoryRecord` relacional em Postgres; `semantic_key` e `cluster_id` para deduplicação semântica; ingestão apenas de ADRs, canvas e checkpoints; recuperação por filtros estruturados; registro de bundles injetados; sem dependência obrigatória de vector database.

Segundo corte futuro: embeddings em lote; busca semântica híbrida; consolidação por chunks (20k e 40k tokens); ranking por recência, confiança, fonte e correspondência de paths; supersessão e expiração automática.

### 2.6 Consequências (Memória Recursiva)

- Agentes ganham continuidade sem depender de transcript completo.
- O Orchestrator consegue retornar contexto útil com menor custo de tokens.
- Runs futuras podem aproveitar aprendizados de diffs, validações e falhas anteriores.
- A memória melhora planejamento, revisão e implementação repetida em domínios recorrentes.
- O sistema ganha mais um componente operacional, exigindo testes, auditoria e políticas de acesso.
- Memória stale, duplicada ou sem evidência pode degradar agentes; por isso dedupe, escopo e evidência são obrigatórios.
- A latência deve ser controlada com background jobs, batching e timeout de retrieval.
- O custo de embeddings deve ser tratado como otimização posterior, não fundação do MVP.

### 2.7 Alternativas consideradas (Memória Recursiva)

- **Sem memória recursiva**: simples e alinhado ao MVP, mas limita continuidade e reaproveitamento de conhecimento entre runs.
- **Salvar apenas transcript completo**: fácil, mas caro, ruidoso e fraco para busca, auditoria e controle de contexto.
- **Vector store como fonte de verdade**: melhora busca semântica, mas viola a decisão de repositório/Event Store como fontes canônicas.
- **Embedding em tempo real para toda mensagem e tool call**: contexto fresco, mas aumenta custo, latência e ruído.
- **Análise apenas por chunks grandes**: reduz custo, mas perde contexto operacional importante em checkpoints e eventos críticos.
- **Memória livre controlada pelo agente**: flexível, mas arriscada para deduplicação, segurança, isolamento e auditoria.

---

## Apêndice A: Histórico de Evolução

| Data | Evento | ADR Original |
| --- | --- | --- |
| 2026-05-10 | Event Store e tracing definidos como fonte canônica | ADR 0009 |
| 2026-05-10 | Recursive Memory definida como camada derivada | ADR 0012 |
| 2026-05-17 | Ambas consolidadas neste documento único | — |

## Apêndice B: Regra de Fronteira Consolidada

- Event Store responde: **o que aconteceu?**
- Tracing responde: **como uma run evoluiu e como reconstruir seu estado?**
- Memória Recursiva responde: **o que deve ser lembrado e reaproveitado como contexto?**

Memória Recursiva pode ler eventos de tracing, mas não substitui tracing. Se houver conflito, o Orchestrator deve preferir Event Store, ADRs, artefatos e documentos versionados.
