# ADR 0014: Persistencia M0, CLI Minima e Testes de Integracao

## Contexto

O M0 precisava transformar os contratos iniciais de dominio em uma primeira operacao local verificavel. As entidades aprovadas na ADR 0013 (`Task`, `Run`, `Event`, `WorkUnit`, `Agent` e `AgentSession`) precisavam sair do nivel documental e ganhar persistencia, migrations e uma interface minima de uso.

A ADR 0005 definiu uma CLI fina como primeira interface oficial do MVP, antes de Desktop ou Web. A ADR 0009 definiu o Event Store como historico operacional canonico. A ADR 0003 definiu Go e Postgres como stack inicial.

O trabalho implementado introduziu:

- comandos em `cmd/orchestraos` para operar tasks, work units, runs, eventos, sessoes de agente e migrations;
- migrations SQL com goose para `tasks`, `work_units`, `runs`, `agent_sessions` e `events`;
- repositorios Postgres para entidades do dominio M0;
- Event Store com validacao de envelope por JSON Schema;
- runtime fake para simular interacoes iniciais de agente;
- testes de integracao cobrindo envelope, replay, consultas, ciclo task -> work unit -> run e runtime fake.

## Decisao

O M0 adotara a combinacao abaixo como primeira base executavel local:

- `cmd/orchestraos` como camada operacional fina, sem regra de negocio complexa.
- `internal/repository` como camada de acesso Postgres para entidades M0.
- `internal/eventstore` como ponto unico de append, consulta e replay de eventos operacionais.
- Migrations SQL versionadas em `migrations/`, executadas por goose.
- `internal/agent.FakeRuntime` como runtime deterministico o suficiente para validar o fluxo antes do Codex/CLI real.
- Testes de integracao usando `TEST_DB_DSN`, com skip quando o banco nao estiver disponivel.

A CLI deve continuar fina: comandos podem orquestrar chamadas, mas regras de negocio, validacao de contratos, transicoes de estado e consistencia de auditoria devem migrar para servicos internos quando crescerem.

## Consequencias

- O projeto passa a ter uma primeira superficie executavel para validar persistencia, eventos e interacoes.
- O Event Store se torna o caminho obrigatorio para registrar eventos operacionais relevantes.
- As migrations passam a ser parte do contrato de evolucao do M0 e devem ser revisadas junto com os tipos e schemas.
- O runtime fake permite testar fluxo sem depender ainda de sandbox, worktree ou runtime Codex real.
- A CLI atual e aceitavel como bootstrap, mas nao deve acumular logica de orquestracao.

Riscos conhecidos que precisam ser tratados antes de aprovar a execucao ampla:

- alinhar o preenchimento automatico de `id`, `sequence` e `created_at` com a validacao do Event Store;
- ajustar a obrigatoriedade de `run_id` no envelope para eventos que ainda nao pertencem a uma run;
- garantir que transicoes de estado nao apaguem timestamps anteriores;
- evitar que comandos ignorem erro ao persistir eventos ou atualizar estado;
- isolar testes de integracao para nao dependerem de dados residuais do mesmo banco.

Atualizacao de implementacao:

- `EventStore.Append` deve preencher `id`, `sequence`, `created_at`, prioridade padrao e payload vazio antes de validar o envelope.
- O schema do `EventEnvelope` exige `run_id` apenas para eventos `run.*`, `agent.*` e `tool.*`.
- Erros internos podem ser tipados com `internal/apperrors` para diferenciar entrada invalida, validacao, persistencia, runtime e falhas internas.
- Updates de status devem preservar timestamps ja definidos e alterar apenas os campos logicamente aplicaveis a transicao.
- Runtime fake e comandos CLI nao devem transformar erro de persistencia/evento em sucesso operacional.
- O pacote `internal/domain` continua sendo o local adequado para os tipos M0; nao ha necessidade atual de mover para outro diretorio apenas por organizacao.

## Alternativas consideradas

- **Manter apenas tipos e schemas sem banco**: reduziria escopo, mas atrasaria a validacao real do Event Store.
- **Criar primeiro um Orchestrator completo**: aumentaria a chance de modelagem prematura antes de validar persistencia basica.
- **Usar apenas scripts de SQL e comandos manuais**: seria rapido, mas fraco como contrato operacional.
- **Persistir eventos sem validacao de schema**: facilitaria o bootstrap, mas enfraqueceria auditoria e replay desde o inicio.
