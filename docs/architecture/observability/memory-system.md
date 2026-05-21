# Sistema de Memoria Recursiva

## Objetivo

Definir como o OrchestraOS pode ter memoria de longa duracao para agentes sem transformar conversa solta, transcript bruto ou embedding em fonte de verdade.

A ideia e boa para o produto, desde que seja tratada como uma camada derivada, auditavel e controlada pelo Orchestrator. A memoria deve recuperar contexto util de runs anteriores, decisoes, padroes, diffs, validacoes, ferramentas usadas e erros aprendidos, mas sempre com referencia para a evidencia canonica.

## Decisao De Produto

A decisao arquitetural de adotar **Recursive Memory** esta documentada em [ADR 0009: Observabilidade, Event Store e Memoria Recursiva](/docs/adr/0009-observability-and-memory.md).

Este documento detalha a implementacao tecnica, modelo de dados, pipeline de deduplicacao e operacao do sistema de memoria.

A memoria nao substitui repositorio, canvas, ADRs, issues, PRs, Event Store, Agent Task Ledger, Agent Checkpoints, artefatos ou diffs. E um indice operacional derivado dessas fontes.

## Comparacao Com Tracing

A fronteira entre tracing (Event Store) e memoria recursiva esta definida em [ADR 0009](/docs/adr/0009-observability-and-memory.md).

Regra curta:

- tracing e para o Orchestrator entender e controlar a execucao;
- memoria e para agentes receberem contexto reutilizavel;
- memoria deriva do tracing, mas tracing nao deriva da memoria;
- tracing registra "o que aconteceu";
- memoria registra "o que vale lembrar".

## Principios

Principios gerais estao em [ADR 0009](/docs/adr/0009-observability-and-memory.md). Aqui estao as restricoes operacionais especificas:

- Memorias stale devem ser expiradas, supersedidas ou rebaixadas no ranking.
- Segredos, tokens, dados sensiveis e outputs perigosos nao devem ser indexados.
- A gravacao de memoria deve acontecer em segundo plano sempre que possivel.
- O agente nao deve decidir sozinho qual memoria permanente criar.

## Fontes Canonicas

Memorias podem ser derivadas de:

| Fonte | Uso |
| --- | --- |
| `docs/canvas/project-canvas.md` | Visao, principios, riscos e fronteiras do projeto. |
| `docs/adr/` | Decisoes arquiteturais e alternativas rejeitadas. |
| Event Store | Historico operacional normalizado de runs. |
| Agent Task Ledger | Estado vivo da work unit durante execucao. |
| Agent Checkpoints | Snapshots estruturados de progresso. |
| Tool executions | Ferramentas usadas, resultado, erro e impacto. |
| Artifacts | Diffs, logs, validacoes, patches e evidencias. |
| PRs e issues | Historico externo revisavel quando GitHub estiver integrado. |
| PromptSnapshot | Contexto e instrucoes efetivamente entregues ao agente. |

## Tipos De Memoria

| Tipo | Conteudo | Exemplos |
| --- | --- | --- |
| `project_context` | Informacoes persistentes sobre visao, principios e operacao do projeto. | Fonte de verdade, autonomia aprovada, prioridade do MVP. |
| `architectural_decision` | Decisoes e consequencias derivadas de ADRs. | Orchestrator como control plane; Event Store como historico canonico. |
| `domain_knowledge` | Conhecimento sobre dominio funcional ou tecnico. | Como work units, checkpoints e prompt snapshots se relacionam. |
| `code_pattern` | Padroes recorrentes de implementacao no repositorio. | Estilo de validacao, naming, estrutura de pacotes. |
| `implementation_fact` | Fatos de implementacao observados em diffs e artefatos. | Arquivo alterado, refatoracao concluida, migracao aplicada. |
| `agent_operational` | Aprendizados sobre execucao de agentes. | Um toolset foi insuficiente; uma validacao detectou loop. |
| `tool_integration` | Comportamento de ferramentas e conectores. | GitHub exige aprovacao para PR; Slack e conector opcional. |
| `failure_lesson` | Falhas, riscos e formas de evitar repeticao. | Teste instavel, timeout, permissao negada, rollback necessario. |
| `user_intent` | Preferencias ou intencoes estaveis explicitamente registradas. | Produto local-first; preferencia por mudancas pequenas. |
| `policy_permission` | Limites de autonomia e permissao. | Autonomia Nivel 2 no MVP; rede exige aprovacao. |

## Dominios De Memoria

Dominios iniciais recomendados:

- `project`;
- `architecture`;
- `orchestration`;
- `prompting`;
- `tracing`;
- `sandbox`;
- `policy`;
- `github`;
- `cli`;
- `docs`;
- `codebase`;
- `testing`;
- `integrations`.

## Componentes

Os componentes abaixo sao servicos logicos. No primeiro corte, eles podem viver no mesmo processo do Orchestrator ou em jobs internos. A separacao serve para deixar claro quem decide, quem persiste, quem busca e quem injeta contexto.

| Componente | Responsabilidade | Entrada | Saida |
| --- | --- | --- | --- |
| Memory Intake | Consumir fontes canonicas sem bloquear a run. | Eventos, checkpoints, ledger updates, artefatos, ADRs e decisoes. | Itens brutos para avaliacao. |
| Memory Candidate Extractor | Transformar informacao bruta em candidatos pequenos. | Itens do intake. | `MemoryCandidate`. |
| Memory Classifier | Classificar relevancia, tipo, dominio, risco e confianca. | `MemoryCandidate`. | Candidato enriquecido ou rejeitado. |
| Semantic Dedupe | Detectar memoria semanticamente equivalente antes de persistir. | Candidato enriquecido e memorias existentes. | `create`, `merge`, `supersede`, `reject` ou `review_required`. |
| Memory Store | Persistir metadata, evidencias, clusters e referencias de embedding. | Candidato aprovado ou merge. | `MemoryRecord` ou atualizacao de cluster. |
| Memory Retriever | Buscar memorias relevantes para uma run ou checkpoint. | Consulta contextual do Orchestrator. | `RetrievedMemoryBundle` candidato. |
| Prompt Composer | Inserir memoria recuperada no prompt ou checkpoint. | Bundle aprovado pelo Orchestrator. | Prompt ou comando com bloco de memoria. |
| Memory Audit Log | Registrar todas as decisoes de memoria. | Eventos dos componentes. | Trilha auditavel de criacao, merge, rejeicao e injecao. |

### Memory Intake

Recebe eventos, checkpoints, ledger updates, artefatos e decisoes humanas. Deve ser assincrono e tolerante a falha. Se o servico de memoria cair, a run do agente nao deve falhar.

### Memory Candidate Extractor

Transforma entradas canonicas em candidatos de memoria. A primeira versao deve usar fontes estruturadas antes de analisar texto livre:

- checkpoint concluido;
- diff relevante;
- validacao concluida;
- ferramenta aprovada, negada ou com erro;
- decisao humana;
- ADR nova ou alterada;
- mudanca em documento arquitetural.

### Memory Classifier

Classifica cada candidato por:

- tipo de memoria;
- dominio;
- escopo;
- impacto;
- confianca;
- risco;
- validade temporal;
- evidencias relacionadas.

### Semantic Dedupe

Decide se um candidato representa uma memoria nova ou se e apenas outra forma de expressar uma memoria existente.

Esse componente nao compara apenas texto. Ele compara a ideia canonica:

- tipo de memoria;
- dominio;
- escopo;
- sujeito;
- relacao;
- objeto;
- qualificadores;
- evidencias;
- validade temporal;
- impacto pratico para agentes.

Exemplo:

```text
Candidato A:
"O Prompt Composer precisa registrar o hash do prompt por run."

Candidato B:
"Cada AgentSession deve ter PromptSnapshot com fragmentos, versoes e hash."
```

Os textos sao diferentes, mas a logica e a mesma: o sistema precisa preservar a composicao do prompt por run para auditoria. O resultado correto e associar as duas evidencias ao mesmo cluster de memoria, nao criar duas memorias independentes.

### Memory Store

Armazena metadata relacional e referencia de embedding.

Postgres deve ser suficiente para o primeiro desenho. Um indice vetorial pode ser adicionado quando houver dor real de recuperacao semantica. O storage deve permitir filtros estruturados mesmo quando busca vetorial existir.

### Memory Retriever

Recebe uma consulta contextual do Orchestrator e retorna um conjunto pequeno de memorias relevantes. A consulta deve considerar:

- objetivo da task;
- work unit;
- perfil do agente;
- caminhos sob ownership;
- dominio tecnico;
- checkpoints recentes;
- toolset disponivel;
- memorias ja injetadas na run.

### Prompt Composer

Injeta memorias recuperadas como um bloco controlado no prompt ou como comando no proximo checkpoint. O bloco deve citar ids de memoria e evidencias.

### Memory Audit Log

Registra:

- candidato criado;
- candidato rejeitado;
- memoria criada;
- memoria supersedida;
- recuperacao solicitada;
- bundle injetado;
- motivo de nao injecao;
- feedback de uso.

## Modelo Conceitual

```json
{
  "id": "mem_123",
  "type": "implementation_fact",
  "domain": "prompting",
  "scope": {
    "project_id": "project_orchestraos",
    "repo_id": "repo_orchestraos",
    "task_id": "task_456",
    "work_unit_id": "wu_789",
    "paths": ["docs/architecture/interface/prompt-system.md"]
  },
  "title": "Prompt Composer registra PromptSnapshot por run",
  "canonical_claim": {
    "subject": "Prompt Composer",
    "relation": "must_record",
    "object": "PromptSnapshot por run",
    "qualifiers": ["fragmentos", "versoes", "hashes", "toolset"]
  },
  "semantic_key": "prompting:prompt_composer:must_record:prompt_snapshot_per_run",
  "cluster_id": "mem_cluster_prompt_snapshot_audit",
  "content": "O Prompt Composer deve registrar fragmentos, variaveis, hashes e toolset usado em cada AgentSession.",
  "evidence_refs": [
    {
      "kind": "adr",
      "ref": "docs/adr/0007-prompt-composition-system.md"
    }
  ],
  "confidence": 0.92,
  "impact": "medium",
  "status": "active",
  "content_hash": "sha256:...",
  "embedding_ref": "embedding_123",
  "supersedes": [],
  "created_at": "2026-05-03T12:00:00Z"
}
```

## Ciclo De Vida

1. Um evento canonico acontece no Event Store.
2. Memory Intake recebe ou consome o evento pela outbox.
3. Candidate Extractor cria candidatos pequenos e citaveis.
4. Classifier remove ruido, segredos, dados sensiveis e conteudo sem impacto.
5. Semantic Dedupe cria uma representacao canonica e compara com memorias existentes.
6. O candidato e rejeitado, mesclado, usado para superseder memoria antiga ou aprovado como memoria nova.
7. Candidato aprovado vira `MemoryRecord`.
8. Embedding e indice sao atualizados em segundo plano.
9. Ao iniciar uma AgentSession, o Orchestrator consulta memorias relevantes.
10. Durante checkpoints, o Orchestrator pode consultar novamente se houve mudanca de goal.
11. Prompt Composer cria `RetrievedMemoryBundle`.
12. Bundle injetado e registrado no Memory Audit Log e no PromptSnapshot quando fizer parte do prompt inicial.
13. Feedback de uso, contradicao ou obsolescencia atualiza ranking e status.

## Recuperacao

A recuperacao deve combinar filtros estruturados e busca semantica.

Filtros recomendados:

- `project_id`;
- `repo_id`;
- `domain`;
- `type`;
- `paths`;
- `agent_profile`;
- `risk_level`;
- `status = active`;
- memorias ainda nao injetadas na run.

Sinais de ranking:

- similaridade semantica com o objetivo;
- correspondencia de caminhos;
- autoridade da fonte;
- recencia;
- confianca;
- impacto;
- uso anterior com sucesso;
- ausencia de contradicao com documentos atuais.

Limites iniciais:

- recuperar no maximo 3 a 10 memorias por bundle;
- manter o bloco de memoria dentro de um orcamento de tokens por perfil de agente;
- nao repetir a mesma memoria na mesma run;
- preferir memoria citada em ADR ou documento versionado quando houver empate;
- marcar memorias contraditorias como suspeitas em vez de injeta-las.

## Deduplicacao E Obsolescencia

O sistema deve evitar salvar e injetar memorias repetidas.

Duplicacao deve ser tratada em camadas. Texto igual e apenas o caso mais simples. O problema principal e memoria com texto diferente e mesma logica.

Mecanismos recomendados:

- `content_hash` para conteudo normalizado;
- `source_fingerprint` para origem canonica;
- `semantic_key` derivada da proposicao canonica;
- `cluster_id` para agrupar memorias equivalentes ou quase equivalentes;
- similaridade semantica acima de limite configuravel;
- chave composta por tipo, dominio, escopo e titulo normalizado;
- relacao `supersedes` quando uma memoria nova substitui outra;
- expiracao por TTL para fatos operacionais temporarios;
- invalidacao quando arquivo, ADR ou politica fonte muda;
- log de bundles ja injetados por run.

Memorias devem ser imutaveis em conteudo. Correcoes devem criar nova versao e superseder a anterior.

### Identidade Semantica

Antes de persistir, cada candidato deve ser convertido para uma proposicao canonica:

```json
{
  "type": "architectural_decision",
  "domain": "prompting",
  "scope": {
    "project_id": "project_orchestraos",
    "repo_id": "repo_orchestraos",
    "paths": ["docs/architecture/interface/prompt-system.md"]
  },
  "canonical_claim": {
    "subject": "Prompt Composer",
    "relation": "must_record",
    "object": "PromptSnapshot por run",
    "qualifiers": ["fragmentos", "versoes", "hashes"]
  }
}
```

A chave semantica e criada a partir dessa estrutura, nao do texto final da memoria.

```text
semantic_key = type + domain + normalized_scope + subject + relation + object + qualifiers
```

Se duas memorias tiverem a mesma chave semantica, o sistema deve preferir merge de evidencias em vez de criar nova memoria.

### Decisoes De Dedupe

| Caso | Decisao |
| --- | --- |
| Mesmo `source_fingerprint` | Rejeitar duplicata. |
| Mesmo `semantic_key` | Mesclar evidencias no mesmo cluster. |
| Alta similaridade, mesmo tipo, dominio e escopo | Mesclar ou pedir revisao automatizada. |
| Mesma ideia, mas fonte nova mais autoritativa | Superseder ou aumentar confianca da memoria existente. |
| Mesmo assunto, mas qualificadores diferentes | Manter no mesmo cluster, mas como registros separados. |
| Contradicao entre memoria nova e existente | Marcar conflito e nao injetar ate resolver. |
| Escopo diferente | Criar memoria separada ou subcluster por escopo. |

Exemplo: uma memoria dizendo "rede exige aprovacao" para o MVP e outra dizendo "rede exige aprovacao para agentes Nivel 2" podem estar no mesmo cluster, mas a segunda e mais especifica. O sistema nao deve apagar a primeira automaticamente; deve preferir a memoria mais especifica no retrieval quando o contexto for Nivel 2.

### Pipeline De Dedupe

1. Normalizar texto: caixa, espacos, nomes conhecidos, paths e aliases.
2. Extrair proposicao canonica: sujeito, relacao, objeto, qualificadores e escopo.
3. Calcular `content_hash` e `semantic_key`.
4. Buscar candidatos existentes por tipo, dominio, escopo e entidades citadas.
5. Comparar embeddings apenas dentro desse conjunto reduzido.
6. Aplicar decisao: rejeitar, mesclar, superseder, criar novo registro ou mandar para revisao.
7. Registrar a decisao no Memory Audit Log.

Essa abordagem reduz custo e evita comparar cada memoria nova com todo o banco.

## Injeção No Agente

O bloco de memoria deve ser explicito e limitado:

```text
Retrieved Memory:
- [mem_123] O Prompt Composer registra PromptSnapshot por run.
  Evidence: docs/adr/0007-prompt-composition-system.md
- [mem_456] O MVP nao inclui memoria vetorial compartilhada no primeiro corte.
  Evidence: docs/implementation/roadmap.md
```

Regras:

- memoria injetada e contexto auxiliar, nao instrucao de maior prioridade;
- o agente deve preferir o repositório e ADRs se houver conflito;
- o agente pode citar memoria usada em checkpoint ou conclusao;
- o agente pode sinalizar memoria stale;
- o Orchestrator deve registrar o bundle entregue.

## Custo E Latencia

A implementacao deve evitar analise cara a cada mensagem ou tool call.

Estrategia recomendada:

- caminho quente: eventos estruturados, filtros baratos e retrieval com timeout curto;
- caminho frio: analise mais profunda em background;
- embeddings em lote;
- consolidacao quando a sessao atingir marcos configuraveis, por exemplo 20k e 40k tokens;
- extracao imediata apenas para eventos de alto valor, como checkpoint, diff relevante, validacao, erro recorrente e decisao humana;
- cache por task, dominio e caminhos;
- fallback silencioso: se memoria nao responder, o agente continua com canvas, ADRs, ledger e prompt normal.

Essa estrategia reduz custo sem perder continuidade.

## Politica De Seguranca

Antes de indexar, o sistema deve:

- remover segredos;
- evitar armazenar tokens, chaves, senhas e credenciais;
- respeitar escopo de repositorio e worktree;
- bloquear vazamento entre tasks sem permissao;
- aplicar a politica de autonomia vigente;
- registrar quem ou qual evento originou a memoria;
- permitir auditoria e exclusao de memoria sensivel.

## Eventos Recomendados

Tipos futuros:

- `memory.candidate_created`;
- `memory.candidate_rejected`;
- `memory.record_created`;
- `memory.record_superseded`;
- `memory.retrieval_requested`;
- `memory.bundle_created`;
- `memory.bundle_injected`;
- `memory.feedback_recorded`;
- `memory.record_expired`.

Esses eventos devem seguir o envelope padrao em `docs/contracts/json-schemas.md` quando a implementacao existir.

## Plano De Implementacao Recomendado

Recursive Memory nao deve bloquear o MVP inicial. A ordem segura e:

1. Event Store funcionando.
2. Agent Task Ledger funcionando.
3. Agent Checkpoints funcionando.
4. PromptSnapshot funcionando.
5. Artifact Manager registrando diffs, logs e validacoes.
6. Primeiro Memory Intake baseado apenas em checkpoints e ADRs.
7. Memory Retriever com filtros estruturados.
8. Embeddings e busca semantica.
9. Consolidacao por chunks e thresholds.
10. Injeção dinamica durante checkpoints.

## Criterios De Aceite Futuros

- Memorias criadas sempre possuem evidencia canonica.
- Memorias repetidas nao sao criadas em duplicidade.
- A mesma memoria nao e injetada duas vezes na mesma run sem motivo.
- Recuperacao respeita escopo, dominio e caminhos.
- PromptSnapshot registra memorias usadas no prompt inicial.
- Checkpoint registra quando o agente usou ou rejeitou memoria.
- Falha do servico de memoria nao interrompe execucao do agente.
- Segredos nao aparecem em memoria indexada.
- Memoria stale pode ser supersedida ou expirada.
