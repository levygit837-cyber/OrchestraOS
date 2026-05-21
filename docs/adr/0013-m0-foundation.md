# ADR 0013: Fundação Técnica M0 — Schemas, Persistência e Testes

**Status:** Consolidated (absorve: ADR 0014)  
**Data original:** 2026-05-10  
**Última atualização:** 2026-05-17

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

`CommunicationProtocol` permanece documentado por `docs/architecture/protocols/communication-protocol.md` e pelo envelope versionado de eventos e comandos. O contrato executavel relevante no M0 e o `EventEnvelope` com payloads versionados.

`Session` generica fica adiada. O primeiro caso operacional de sessao e `AgentSession`; sessoes para CLI, GitHub, humano ou conectores devem ser modeladas depois, quando precisarem de estado vivo proprio.

## Consequências

- O primeiro esqueleto fica menor, mais verificavel e reversivel.
- O Event Store, o DAG de work units e a execucao por agente continuam cobertos pelos contratos necessarios.
- Agentes futuros tem uma regra clara para nao criar schemas de `Orchestrator`, `CommunicationProtocol` ou `Session` generica sem nova necessidade concreta.
- Quando CLI, GitHub ou conectores precisarem de sessao propria, uma nova decisao ou atualizacao de ADR deve definir esse contrato.

---

## 2. Persistência, CLI Mínima e Testes de Integração

### 2.1 Contexto adicional

O M0 precisava transformar os contratos iniciais de domínio em uma primeira operação local verificável. As entidades aprovadas na seção 1 (`Task`, `Run`, `Event`, `WorkUnit`, `Agent` e `AgentSession`) precisavam sair do nível documental e ganhar persistência, migrations e uma interface mínima de uso.

O trabalho implementado introduziu:

- comandos em `cmd/orchestraos` para operar tasks, work units, runs, eventos, sessões de agente e migrations;
- migrations SQL com goose para `tasks`, `work_units`, `runs`, `agent_sessions` e `events`;
- repositórios Postgres para entidades do domínio M0;
- Event Store com validação de envelope por JSON Schema;
- runtime fake para simular interações iniciais de agente;
- testes de integração cobrindo envelope, replay, consultas, ciclo task → work unit → run e runtime fake.

### 2.2 Decisão

O M0 adotará a combinação abaixo como primeira base executável local:

- `cmd/orchestraos` como camada operacional fina, sem regra de negócio complexa.
- `internal/repository` como camada de acesso Postgres para entidades M0.
- `internal/eventstore` como ponto único de append, consulta e replay de eventos operacionais.
- Migrations SQL versionadas em `migrations/`, executadas por goose.
- `internal/agent.FakeRuntime` como runtime determinístico o suficiente para validar o fluxo antes do Codex/CLI real.
- Testes de integração usando `TEST_DB_DSN`, com skip quando o banco não estiver disponível.

A CLI deve continuar fina: comandos podem orquestrar chamadas, mas regras de negócio, validação de contratos, transições de estado e consistência de auditoria devem migrar para serviços internos quando crescerem.

### 2.3 Consequências (Persistência e CLI)

- O projeto passa a ter uma primeira superfície executável para validar persistência, eventos e interações.
- O Event Store se torna o caminho obrigatório para registrar eventos operacionais relevantes.
- As migrations passam a ser parte do contrato de evolução do M0 e devem ser revisadas junto com os tipos e schemas.
- O runtime fake permite testar fluxo sem depender ainda de sandbox, worktree ou runtime Codex real.
- A CLI atual é aceitável como bootstrap, mas não deve acumular lógica de orquestração.

Riscos conhecidos que precisam ser tratados antes de aprovar a execução ampla:

- alinhar o preenchimento automático de `id`, `sequence` e `created_at` com a validação do Event Store;
- ajustar a obrigatoriedade de `run_id` no envelope para eventos que ainda não pertencem a uma run;
- garantir que transições de estado não apaguem timestamps anteriores;
- evitar que comandos ignorem erro ao persistir eventos ou atualizar estado;
- isolar testes de integração para não dependerem de dados residuais do mesmo banco.

Atualização de implementação:

- `EventStore.Append` deve preencher `id`, `sequence`, `created_at`, prioridade padrão e payload vazio antes de validar o envelope.
- O schema do `EventEnvelope` exige `run_id` apenas para eventos `run.*`, `agent.*` e `tool.*`.
- Erros internos podem ser tipados com `internal/apperrors` para diferenciar entrada inválida, validação, persistência, runtime e falhas internas.
- Updates de status devem preservar timestamps já definidos e alterar apenas os campos logicamente aplicáveis à transição.
- Runtime fake e comandos CLI não devem transformar erro de persistência/evento em sucesso operacional.
- O pacote `internal/domain` continua sendo o local adequado para os tipos M0; não há necessidade atual de mover para outro diretório apenas por organização.

### 2.4 Alternativas consideradas (Persistência e CLI)

- **Manter apenas tipos e schemas sem banco**: reduziria escopo, mas atrasaria a validação real do Event Store.
- **Criar primeiro um Orchestrator completo**: aumentaria a chance de modelagem prematura antes de validar persistência básica.
- **Usar apenas scripts de SQL e comandos manuais**: seria rápido, mas fraco como contrato operacional.
- **Persistir eventos sem validação de schema**: facilitaria o bootstrap, mas enfraqueceria auditoria e replay desde o início.

---

## Apêndice A: Histórico de Evolução

| Data | Evento | ADR Original |
|------|--------|--------------|
| 2026-05-10 | Escopo de schemas e tipos M0 definido | ADR 0013 |
| 2026-05-10 | Persistência, CLI e testes de integração definidos | ADR 0014 |
| 2026-05-17 | Ambos consolidados neste documento único | — |

## Apêndice B: Alternativas Consideradas (Escopo de Schemas)

- **Criar schemas para todos os nomes agora**: daria aparência de completude, mas aumentaria acoplamento e abstração prematura.
- **Manter apenas documentação conceitual, sem schemas executáveis no M0**: reduziria trabalho inicial, mas deixaria a validação de contratos fraca para Event Store e bordas do sistema.
- **Criar apenas `Task`, `Run`, `Event` e `WorkUnit`**: seria ainda menor, mas deixaria sem contrato o primeiro vínculo operacional entre run e agente.
