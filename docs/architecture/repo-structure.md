# Estrutura Inicial do Repositorio

Este documento define a estrutura inicial planejada para transformar a arquitetura documentada em codigo pequeno, testavel e evolutivo.

## Decisao

O repositorio deve iniciar como um modulo Go para o nucleo do Orchestrator, mantendo contratos JSON versionados como artefatos independentes.

```text
cmd/orchestraos/
internal/domain/
contracts/
contracts/schemas/
tests/
docs/
```

## Responsabilidades

| Caminho | Responsabilidade |
| --- | --- |
| `cmd/orchestraos/` | Entrada futura da CLI fina e comandos locais do MVP. |
| `internal/domain/` | Tipos centrais do dominio: Task, Run, Event, WorkUnit, Agent e AgentSession. |
| `contracts/` | Pacote Go leve para localizar e embutir contratos versionados. |
| `contracts/schemas/` | JSON Schemas executaveis, separados por dominio e protocolo. |
| `tests/` | Validacoes de contrato sem depender de servicos externos. |
| `docs/` | Fonte de verdade para arquitetura, canvas, ADRs, contratos narrativos e operacao. |

## Regras

- O dominio nao deve depender de banco, WebSocket, GitHub, Docker ou CLI.
- JSON Schemas sao contratos de borda; tipos Go sao o modelo interno inicial.
- Schemas devem rejeitar campos desconhecidos por padrao.
- Novas dependencias so devem entrar quando a validacao com biblioteca padrao nao for suficiente.
- Mudancas arquiteturais relevantes continuam exigindo ADR.
- Mudancas de contrato devem atualizar `docs/contracts/json-schemas.md` e os arquivos `.schema.json`.

## Escopo Do Primeiro Esqueleto

O primeiro esqueleto de codigo deve criar apenas os tipos e contratos executaveis necessarios para M0:

- `Task`
- `Run`
- `Event`
- `WorkUnit`
- `Agent`
- `AgentSession`

`Orchestrator` deve aparecer como pacote/servico quando a implementacao do control plane comecar. `CommunicationProtocol` deve continuar representado pelo envelope de eventos/comandos e pela documentacao do protocolo. `Session` generica fica adiada ate existir estado vivo para CLI, GitHub, humano ou conectores.

Essa fronteira esta registrada na [ADR 0013](../adr/0013-m0-domain-contract-scope.md).

## Fora do Escopo Inicial

- Implementacao real de Postgres, WebSocket, Docker ou GitHub.
- Migrations de banco.
- CLI completa.
- Runtime real do Codex/CLI.
