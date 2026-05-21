# 0028. Padrões de Arquitetura Core e Nomenclatura — Eliminação do Deus Pacote

**Status:** Superseded by ADR 0022 (conteúdo absorvido em 2026-05-18)  
**Data:** 2026-05-17

> **Nota:** Este ADR foi absorvido pelo ADR 0022. O conteúdo abaixo é mantido para referência histórica.
> As regras de nomenclatura, template de `doc.go` e nomes proibidos de arquivos foram incorporados à Seção 5 do ADR 0022.

---

## 1. Contexto

Após a consolidação da Arquitetura de Módulos Verticais (ADR-0022), o projeto possui uma estrutura de módulos (`internal/modules/*`) altamente padronizada e previsível. Cada módulo segue um template rígido de arquivos (`models.go`, `service.go`, `repository.go`, `queries.go`, etc.), o que torna o código extremamente navegável por agentes de IA.

No entanto, a camada de infraestrutura compartilhada (`internal/core/*`) e a camada de coordenação cross-module (`internal/core/coordination/`) acumularam inconsistências estruturais e nomes genéricos que quebram a previsibilidade do sistema:

1. **`core/coordination/`** evoluiu para um **"Deus Pacote"** — acumula 7 responsabilidades distintas (runtime relay, prompt orchestration, cascade cancellation, run-workunit sync, session timeout, projection updates, cross-module validation), importando 8 dos 10 módulos do sistema.
2. **Falta de padrão unificado** para `core/*` — enquanto `modules/*` tem template rígido, `core/*` cresceu organicamente sem convenção.

## 2. Decisão

### 2.1 Eliminação do Padrão "Deus Pacote"

**Regra absoluta:** Nenhum pacote em `internal/core/*` ou `internal/modules/*` pode acumular mais de **3 responsabilidades de domínio distintas**.

**Corolário:** `core/coordination/` será **esvaziado de lógica de negócio** e posteriormente removido. Sua lógica será distribuída para os módulos que possuem o processo.

#### 2.1.1 Eliminação de Pacotes "Deus"

**Regra absoluta:** Nenhum pacote em `internal/core/*` ou `internal/modules/*` pode acumular mais de **3 responsabilidades de domínio distintas**.

**Regra de ouro para cross-module:** Se um fluxo envolve módulos A e B, a orquestração deve residir no módulo que **inicia o processo** ou que **possui o aggregate raiz** do fluxo. Nunca em um pacote utilitário genérico.

### 2.2 Padrão de Nomenclatura de Arquivos

#### 2.2.1 Nomes Obrigatórios (MUST)

Todo pacote em `internal/core/*` e `internal/modules/*` deve usar nomes descritivos que comuniquem **o quê** o arquivo faz, não **que tipo** de código ele contém.

**Para `internal/modules/*`:**

| Nome do Arquivo | Conteúdo Esperado | Obrigatório? |
|-----------------|-------------------|--------------|
| `doc.go` | Package documentation com responsabilidade, dependências e regras | ✅ MUST |
| `contract.go` | ModuleContract + regras hierárquicas (GLOBAL, TYPE, SPECIFIC) | ✅ MUST |
| `README.md` | Propósito, file map, allowed dependencies | ✅ MUST |
| `CONTRACTS.md` | Invariantes, state machine, boundary rules, file decomposition | ✅ MUST |
| `models.go` | Tipos próprios (structs, enums, constants) do domínio | ✅ MUST |
| `queries.go` | SQL constants (INSERT, UPDATE, DELETE, SELECT) | ✅ MUST |
| `repository.go` | CRUD puro, zero business logic | ✅ MUST |
| `service.go` | Lógica principal, transações, eventos | ✅ MUST |
| `events.go` | `EventTypeForStatus`, mapeamento evento ↔ status | ✅ MUST |
| `validation.go` | Validações sintáticas de input (UUID, texto, enums) | ✅ MUST |
| `fetch.go` | `RequireByID` e helpers de consulta para DI | 🟡 MAY |
| `service_<verbo>.go` | Decomposição de `service.go` quando > 300 linhas | 🟡 MAY |
| `types.go` | Tipos auxiliares que poluem `models.go` | 🟡 MAY |

**Para `internal/core/*`:**

| Nome do Arquivo | Conteúdo Esperado | Obrigatório? |
|-----------------|-------------------|--------------|
| `doc.go` | Package documentation | ✅ MUST |
| `errors.go` | Tipos de erro, códigos, construtores | ✅ MUST (em `apperrors/`) |
| `types.go` | Tipos puros compartilhados (structs, interfaces) | 🟡 MAY |
| `validators.go` | Funções de validação de primitivos | 🟡 MAY |
| `repository.go` | CRUD de infraestrutura (ex: eventstore) | 🟡 MAY |
| `queries.go` | SQL de infraestrutura | 🟡 MAY |
| `store.go` | Camada de storage com regras de negócio leves | 🟡 MAY |
| `marshal.go` | Serialização de payloads | 🟡 MAY |
| `statemachine.go` | Regras de transição e tabela de estados | 🟡 MAY |
| `replay.go` | Projeção de estado a partir de eventos | 🟡 MAY |
| `transitions.go` | Builders de payload e contexto de transição | 🟡 MAY |
| `append.go` | Operações de append no event store | 🟡 MAY |
| `audit.go` | Regras de auditoria para estados finais | 🟡 MAY |
| `transactions.go` | Helpers de lifecycle de transação SQL | 🟡 MAY |
| `conn.go` | Config e abertura de conexão DB | 🟡 MAY |

#### 2.2.2 Nomes Proibidos (MUST NOT)

Os seguintes nomes de arquivo são **estritamente proibidos** em todo o projeto:

| Nome Proibido | Motivo |
|---------------|--------|
| `helpers.go` | Não comunica conteúdo. É um "lixo" semântico. |
| `utils.go` | Mesmo problema. "Util" significa tudo e nada. |
| `common.go` | O que é "comum"? Comum para quem? |
| `base.go` | Sugere herança. Go não tem herança. |
| `misc.go` | "Miscelânea" é admissão de falta de coesão. |
| `kit.go` / `txkit.go` | Jargão de framework. Não descreve função. |
| `ops.go` / `eventops.go` | Abreviação vaga. "Operations" de quê? |
| `stuff.go` / `things.go` | (Óbvio, mas já vimos em projetos reais) |

**Exceção única:** Um arquivo pode se chamar `<package>.go` (ex: `validation.go` em `package validation`) **apenas se** for o único arquivo do pacote. Pacotes com 2+ arquivos devem usar nomes descritivos.

### 2.3 Padrão de Estrutura de Pacotes `core/`

Cada pacote em `internal/core/*` deve seguir o princípio da **Responsabilidade Única a Nível de Pacote**:

```
internal/core/
  apperrors/         → Erros tipados. Leaf package. Não importa ninguém.
  db/                → Abstração de DB, transações, advisory locks.
  validation/        → Validação de primitivos (UUID, texto, enums).
  serialization/     → Marshaling de payloads para eventos.
  statemachine/      → Regras de transição de estado (tabela + CanTransition).
  transition/        → Builders de payload/contexto para transições.
  eventstore/        → Persistência de eventos com schema validation.
  event/             → Serviço de append idempotente (wrapper do eventstore).
```

### 2.4 Regra de SQL e Encapsulamento

**Regra absoluta:** Todo SQL que opera sobre uma tabela `X` deve residir em `internal/modules/X/queries.go` ou `internal/modules/X/repository.go`.

**Proibição:** `internal/core/*` nunca contém SQL de tabelas de domínio (`runs`, `tasks`, `work_units`, `agent_sessions`, etc.). `core/` só pode conter SQL de tabelas de infraestrutura (`events`, se houver uma tabela genérica de eventos gerenciada pelo eventstore).

### 2.5 Padrão de Documentação para LLMs

Cada `doc.go` deve seguir um template padronizado para máxima eficiência de contexto:

```go
// Package <nome> <verbo> <objeto>.
//
// # Responsibility
// <uma frase clara do que este pacote faz>
//
// # Key Types
//   - <Type>: <descrição de uma linha>
//
// # Dependencies
//   - core/<X>: <para que serve o import>
//
// # Related Packages
//   - modules/<Y>: <relação com este pacote>
//
// # Rules (para LLMs)
//   - <regra específica que agentes devem seguir>
package <nome>
```

## 3. Consequências

### 3.1 Positivas

- **Zero ambiguidade para LLMs:** Nomes de arquivos comunicam conteúdo. Um agente nunca precisa abrir um `helpers.go` para descobrir o que está dentro.
- **Encapsulamento restaurado:** SQL volta para os módulos donos das tabelas. Nenhum pacote orfão.
- **Coesão por responsabilidade:** Cada pacote faz uma coisa. Se um pacote precisa de 7 arquivos com nomes diferentes, isso é um sinal de que talvez sejam 2 pacotes.
- **Arquitetura Vertical Slice real:** Cross-module não é um pacote separado — é uma capacidade dos módulos que expõem interfaces de coordenação.

### 3.2 Negativas

- **Refatoração de `coordination/`:** ~700 linhas distribuídas entre 7 módulos. Requer toque em múltiplos testes.
- **Renomeação de arquivos:** ~10 arquivos em `core/*` precisam ser renomeados. Impacto mecânico, mas zero lógica alterada.
- **Mudança de mentalidade:** Desenvolvedores (humanos e IA) precisam parar de "jogar código no coordination" e pensar "qual módulo possui este processo?"

## 4. Plano de Migração

### Fase 1: Renomeações Mecânicas (PR isolado, baixo risco)

| Ação | Arquivos |
|------|----------|
| `db/txkit.go` → `db/transactions.go` | 1 |
| `serialization/serialization.go` → `serialization/marshal.go` | 1 |
| `validation/validation.go` → `validation/validators.go` | 1 |
| `transition/helpers.go` → `transition/payload.go` + `transition/audit.go` | 2 |
| `transition/eventops.go` → `transition/append.go` | 1 |

**Validação:** `go build ./...`, `go test ./...`, `./scripts/go/verify-contracts.sh`

### Fase 2: Esvaziamento de `coordination/` (PR por módulo, médio risco)

| PR | Conteúdo | Módulos Afetados |
|----|----------|------------------|
| #1 | Mover `UpdateRunProjection` → `run/repository.go` | `run`, `coordination` |
| #2 | Mover `CancelTaskDependents` → `task/service_cascade.go` | `task`, `coordination` |
| #3 | Mover `TransitionRunWithWorkUnit` → `run/service_workunit.go` | `run`, `workunit`, `coordination` |
| #4 | Mover `AgentSessionTimeout` → `agentsession/service_timeout.go` | `agentsession`, `coordination` |
| #5 | Mover `PromptOrchestrator` → `prompt/service_orchestrate.go` | `prompt`, `coordination` |
| #6 | Mover `RuntimeEventRelay` → `run/service_relay.go` | `run`, `coordination` |
| #7 | Remover `coordination/` vazio + atualizar docs | Todos |

**Regra:** Cada PR deve manter builds verdes e não pode mergear se quebrar testes de arquitetura.

### Fase 3: Atualização de Documentação

- Atualizar ADR-0022 para remover referências a `core/coordination` como camada de orquestração.
- Atualizar `module_index.md` para refletir nova estrutura.
- Atualizar `AGENTS.md` com regra de "nomes proibidos de arquivos".

## 5. Exemplos

### 5.1 Antes (Problema)

```
core/
  coordination/
    helpers.go              ← O que está aqui? Ninguém sabe.
    agentsession_orchestrator.go  ← Não é um orchestrator.
    eventops.go             ← "Ops" de quê?
    queries.go              ← SQL de runs, WUs, sessions misturado
```

### 5.2 Depois (Ideal)

```
core/
  apperrors/
    doc.go
    errors.go
  db/
    doc.go
    conn.go
    dbtx.go
    transactions.go         ← era txkit.go
  validation/
    doc.go
    validators.go           ← era validation.go
  serialization/
    doc.go
    marshal.go              ← era serialization.go
  statemachine/
    doc.go
    statemachine.go
    replay.go
  transition/
    doc.go
    types.go
    payload.go              ← era helpers.go (parte)
    audit.go                ← era helpers.go (parte)
    append.go               ← era eventops.go
  eventstore/
    doc.go
    store.go
    repository.go
    queries.go
    validator.go
  eventappend/
    doc.go
    models.go
    service.go
```

```
modules/
  task/
    ...
    service_cascade.go      ← veio de coordination/cascade.go
  run/
    ...
    service_workunit.go     ← veio de coordination/run_workunit_sync.go
    service_relay.go        ← veio de coordination/runtime_relay.go
  agentsession/
    ...
    service_timeout.go      ← veio de coordination/agentsession_orchestrator.go
  prompt/
    ...
    service_orchestrate.go  ← veio de coordination/prompt_orchestrator.go
```

## 6. Regras para Agentes de IA (Resumo)

Quando um agente de IA modificar este projeto:

1. **NUNCA crie um arquivo chamado `helpers.go`, `utils.go`, `common.go`, `base.go`, `misc.go`, `kit.go`, ou `ops.go`.**
2. **SEMPRE** pergunte: "Qual módulo possui este processo?" antes de colocar código em `core/`.
3. **SEMPRE** coloque SQL no `queries.go` do módulo dono da tabela.
4. **SEMPRE** use `service_<verbo>.go` para decomposição de serviços grandes.
5. **NUNCA** adicione lógica cross-module a `core/` — adicione ao módulo dono do aggregate raiz.

---

## Apêndice A: Checklist de Validação

Antes de mergear qualquer PR que toque em `internal/core/*` ou `internal/modules/*`:

- [ ] Nenhum arquivo novo tem nome proibido (`helpers`, `utils`, `common`, etc.)
- [ ] Todo pacote tem `doc.go` com template completo
- [ ] SQL está no módulo dono da tabela
- [ ] `go build ./...` passa
- [ ] `go test ./...` passa
- [ ] `./scripts/go/verify-contracts.sh` passa
- [ ] `./scripts/go/lint.sh` passa
- [ ] Architecture tests passam (especialmente `module_boundaries_test.go`)
