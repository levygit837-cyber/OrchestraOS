# Plano de Implementação

Este roadmap organiza o fluxo ideal em fatias tecnicas pequenas e verificaveis.

## Fluxo Alvo

```text
UserMessage
-> Orchestrator
-> Break Tasks into DAG
-> Create Prompts/SystemPrompts
-> Setup Sandbox
-> Spawn Agents
-> Run Task
-> Logs and Events
-> Orchestrator Live View Loop
-> Send Messages to Agents if needed
-> Approve or Deny Tools
-> Agent Loops and Checkpoints
-> Task Complete
-> Approval or Deny Merge
-> Review Diffs
```

## Estratégia

Implementar primeiro um "walking skeleton": uma versao simples que atravessa o fluxo inteiro com agente fake, banco local e CLI. Depois trocar partes por implementacoes reais.

## Milestones

### M0: Contratos e Esqueleto

Objetivo:

- definir estrutura inicial do repo;
- criar tipos principais;
- criar schemas;
- preparar migrations iniciais;
- criar CLI minima.

Aceite:

- `Task`, `Run`, `Event` e `WorkUnit` existem no dominio;
- CLI consegue criar uma task local;
- evento `task.created` e persistido;
- teste de schema falha com payload invalido.

### M1: Event Store e State Machine

Objetivo:

- persistir eventos;
- reconstruir estado de task/run;
- validar transicoes.

Aceite:

- eventos sao idempotentes;
- consulta por `task_id` e `run_id` funciona;
- transicoes invalidas sao bloqueadas;
- replay reconstrói estado esperado.

### M2: Task Graph

Objetivo:

- transformar uma task em DAG de work units;
- validar ciclos e conflitos.

Aceite:

- grafo aciclico valido e persistido;
- ciclo e rejeitado;
- conflito de `owned_paths` bloqueia paralelismo;
- nova versao do grafo pode ser criada em replanejamento.

### M3: Prompt Composer

Objetivo:

- registrar fragmentos;
- montar SystemPrompt e TaskPrompt;
- gerar PromptSnapshot.
- selecionar toolset minimo por AgentSession.

Aceite:

- fragmentos obrigatorios entram no prompt;
- conflito entre fragmentos e detectado;
- snapshot tem hash e versoes;
- task prompt inclui objetivo, criterios, validacao e ledger inicial.
- ToolsetSnapshot e criado para a sessao.

### M4: Agent Runtime Fake

Objetivo:

- simular agente de forma deterministica;
- exercitar WebSocket, heartbeat, checkpoint e conclusao.
- persistir AgentCheckpoint com resumo minimo e evidencias.

Aceite:

- agente fake conecta;
- envia `agent.started`;
- envia heartbeat;
- atualiza ledger;
- emite `agent.checkpoint_reached` com goal atual, arquivos tocados e evidencias;
- solicita ferramenta simulada;
- conclui com artifact fake.

### M5: WebSocket e Live View

Objetivo:

- manter canal vivo agente-orchestrator;
- enviar comandos pendentes;
- suportar reconexao.

Aceite:

- CLI consegue assistir eventos de uma run;
- `message.interrupt` chega no proximo checkpoint;
- reconexao com `last_seen_event_id` reenvia pendencias;
- timeout marca sessao como desconectada.

### M6: Sandbox Manager

Objetivo:

- criar branch e worktree por work unit;
- iniciar container com limites;
- preservar artefatos.

Aceite:

- worktree fica fora do repo principal;
- container nao e privilegiado;
- Docker socket nao e montado;
- diff e coletado ao final;
- limpeza nao apaga evidencias.

### M7: Policy Engine e Tools

Objetivo:

- classificar ferramentas;
- autoaprovar acoes seguras;
- exigir aprovacao para acoes sensiveis.

Aceite:

- leitura e validacao local sao permitidas;
- rede, segredo, push e PR exigem aprovacao;
- acao proibida e negada;
- todas as decisoes viram eventos.

### M7.5: Especialização Dinâmica Controlada

Objetivo:

- permitir `DynamicPromptFragment` temporario;
- permitir solicitacao de ferramenta ausente;
- reconfigurar AgentSession com novo PromptSnapshot e ToolsetSnapshot quando aprovado.

Aceite:

- fragmento dinamico nao pode sobrescrever politica global;
- expansao de toolset exige evento e decisao do Orchestrator;
- ledger e preservado entre sessoes;
- historico da sessao anterior continua consultavel.

Observacao: esta milestone e futura. Nao bloqueia o primeiro MVP estatico.

### M8: Codex/CLI Runtime

Objetivo:

- substituir agente fake por Codex/CLI em sandbox;
- controlar prompt, eventos e tool requests.

Aceite:

- agente Codex executa work unit simples;
- eventos principais sao emitidos;
- max steps e timeout funcionam;
- resultado inclui diff, validacao e resumo.

### M9: Review e Merge Gate

Objetivo:

- revisar diff;
- aprovar ou negar integracao;
- registrar decisao.

Aceite:

- CLI mostra diff e evidencias;
- merge exige aprovacao;
- negacao preserva branch e artefatos;
- conclusao publica resumo na CLI/GitHub quando configurado.

### M10: GitHub Complementar

Objetivo:

- receber pedidos por issue quando aplicavel;
- publicar status em issue/PR;
- criar PR quando aprovado.

Aceite:

- GitHub Issue cria ou referencia task;
- status final e enviado para CLI/GitHub;
- GitHub PR so e aberto com aprovacao;
- falha de conector usa outbox e retry.

### M11: Memoria Recursiva

Objetivo:

- criar memoria operacional derivada de fontes canonicas;
- recuperar contexto util para agentes sem depender de transcript completo;
- deduplicar e auditar memorias criadas e injetadas.

Aceite:

- `MemoryRecord` sempre referencia evidencia canonica;
- ingestao inicial usa ADRs, canvas e checkpoints;
- memoria repetida nao gera duplicata;
- `RetrievedMemoryBundle` respeita escopo, dominio, paths e autonomia;
- bundles injetados sao registrados para evitar repeticao na mesma run;
- falha do servico de memoria nao interrompe AgentSession.

Observacao: esta milestone e futura. Ela depende de Event Store, Agent Task Ledger, Agent Checkpoints, PromptSnapshot e Artifact Manager.

## Ordem Recomendada

1. M0
2. M1
3. M2
4. M3
5. M4
6. M5
7. M6
8. M7
9. M8
10. M9
11. M10
12. M11

## Backlog Inicial Sugerido

| Prioridade | Item |
| --- | --- |
| P0 | Criar dominio `Task`, `Run`, `Event`, `WorkUnit`. |
| P0 | Criar Event Store local em Postgres. |
| P0 | Criar CLI minima para task/run/event. |
| P0 | Criar validador de JSON Schema. |
| P1 | Criar planner DAG simples. |
| P1 | Criar Prompt Composer com fragmentos estaticos. |
| P1 | Criar ToolsetSnapshot minimo por AgentSession. |
| P1 | Criar agente fake via WebSocket. |
| P1 | Criar AgentCheckpoint persistente por AgentSession. |
| P1 | Criar Policy Engine minimo. |
| P1 | Criar Sandbox Manager com worktree. |
| P2 | Integrar Codex/CLI em container. |
| P2 | Adicionar GitHub Issue/PR gate. |
| P3 | Adicionar DynamicPromptFragment e reconfiguracao controlada de AgentSession. |
| P3 | Adicionar memoria recursiva derivada de ADRs, canvas e checkpoints. |
| P3 | Adicionar conector de chat opcional. |

## Fora Do Primeiro Corte

- Desktop app.
- Web dashboard.
- NATS.
- Temporal.
- gVisor ou Firecracker.
- memoria vetorial compartilhada antes de memoria estruturada, Event Store, checkpoints e deduplicacao.
- marketplace de agentes.
- autonomia nivel 4 ou 5.
