# ADR 0013: Escopo M0 De Schemas E Tipos De Dominio

## Contexto

A milestone M0 exige criar a estrutura inicial do repositorio, tipos principais e schemas para o nucleo do OrchestraOS.

Durante o planejamento, surgiu a possibilidade de criar schemas e tipos para `Task`, `Run`, `Event`, `WorkUnit`, `Agent`, `Orchestrator`, `CommunicationProtocol` e `Session`. Parte desses nomes representa entidades operacionais do dominio, enquanto outros representam componentes arquiteturais ou abstracoes que ainda nao tem necessidade concreta de persistencia e validacao propria no primeiro esqueleto.

Criar contratos demais no M0 aumentaria o risco de modelagem prematura, dificultaria mudancas pequenas e poderia transformar detalhes de implementacao ainda incertos em contratos publicos cedo demais.

## Decisão

No M0, os schemas executaveis e tipos principais devem cobrir apenas:

- `Task`
- `Run`
- `Event`
- `WorkUnit`
- `Agent`
- `AgentSession`

`Orchestrator` permanece como componente arquitetural e deve aparecer inicialmente como pacote ou servico da implementacao, nao como entidade de dominio com schema proprio.

`CommunicationProtocol` permanece documentado por `docs/architecture/communication-protocol.md` e pelo envelope versionado de eventos e comandos. O contrato executavel relevante no M0 e o `EventEnvelope` com payloads versionados.

`Session` generica fica adiada. O primeiro caso operacional de sessao e `AgentSession`; sessoes para CLI, GitHub, humano ou conectores devem ser modeladas depois, quando precisarem de estado vivo proprio.

## Consequências

- O primeiro esqueleto fica menor, mais verificavel e reversivel.
- O Event Store, o DAG de work units e a execucao por agente continuam cobertos pelos contratos necessarios.
- Agentes futuros tem uma regra clara para nao criar schemas de `Orchestrator`, `CommunicationProtocol` ou `Session` generica sem nova necessidade concreta.
- Quando CLI, GitHub ou conectores precisarem de sessao propria, uma nova decisao ou atualizacao de ADR deve definir esse contrato.

## Alternativas consideradas

- **Criar schemas para todos os nomes agora**: daria aparencia de completude, mas aumentaria acoplamento e abstracao prematura.
- **Manter apenas documentacao conceitual, sem schemas executaveis no M0**: reduziria trabalho inicial, mas deixaria a validacao de contratos fraca para Event Store e bordas do sistema.
- **Criar apenas `Task`, `Run`, `Event` e `WorkUnit`**: seria ainda menor, mas deixaria sem contrato o primeiro vinculo operacional entre run e agente.
