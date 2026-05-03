# ADR 0009: Normalização de Histórico e Tracing

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

## Fronteira Com Memoria Recursiva

Tracing e Memoria Recursiva sao parecidos nas fontes, mas possuem papeis diferentes.

O Event Store definido nesta ADR e o historico operacional canonico. Ele preserva o que aconteceu em uma run, em ordem, com correlacao, idempotencia, eventos, logs, artefatos, tool calls, checkpoints e decisoes humanas. Seu consumidor principal e o Orchestrator, incluindo live view, replay, auditoria, diagnostico, politicas e recuperacao.

A Memoria Recursiva, definida na ADR 0012, e uma projecao derivada desse historico e dos documentos versionados. Ela nao preserva tudo. Ela seleciona, resume, classifica, deduplica e recupera apenas informacoes que valem ser reutilizadas como contexto por agentes.

Regra de fronteira:

- Event Store responde: **o que aconteceu?**
- Tracing responde: **como uma run evoluiu e como reconstruir seu estado?**
- Memoria Recursiva responde: **o que deve ser lembrado e reaproveitado como contexto?**

Memoria Recursiva pode ler eventos de tracing, mas nao substitui tracing. Se houver conflito, o Orchestrator deve preferir Event Store, ADRs, artefatos e documentos versionados.

## Consequências

- O Orchestrator consegue reconstruir e analisar runs sem depender de conversas soltas.
- A live view pode ser derivada do Event Store e dos eventos em tempo real.
- Replay, auditoria e diagnostico ficam possiveis desde o inicio.
- O MVP precisa tratar idempotencia, ordenacao e correlacao de eventos como parte do nucleo.

## Alternativas consideradas

- **Salvar apenas transcript do agente**: simples, mas fraco para consulta, replay e politicas.
- **Usar OpenTelemetry completo desde o inicio**: robusto, mas aumenta escopo do MVP.
- **Guardar logs por arquivo sem estrutura**: util como evidencia bruta, mas insuficiente para controle operacional.
