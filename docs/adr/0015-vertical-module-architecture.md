# 0022. Arquitetura de Módulos Verticais — Decisão, Padronização e Políticas

**Status:** Superseded by ADR-0019  
**Data original:** 2026-05-11  
**Última atualização:** 2026-05-21

> ⚠️ **AVISO:** Esta ADR foi substituída pela [ADR-0019](0019-simplified-modular-architecture.md). A arquitetura Vertical Slice foi julgada excessivamente complexa e não verificável. Mantemos este documento para preservar o histórico de decisões.


**Data:** 2026-05-11

## 1. Contexto

O OrchestraOS tem como premissa ser um sistema mantido, evoluído e operado primariamente por Agentes de IA (LLMs). Para que um LLM seja eficaz, ele precisa receber o **contexto correto** na sua janela de leitura, com o mínimo de ruído (informações irrelevantes) possível.

Atualmente, o projeto utiliza uma **Arquitetura em Camadas (Layered Architecture)**, separando o código por preocupação técnica (Technical Cohesion):
- `internal/domain/` (Modelos e tipos para todo o sistema)
- `internal/services/` (Regras de negócio para todo o sistema)
- `internal/repository/` (Acesso a dados para todo o sistema)

O problema estrutural para LLMs é que, ao executar uma tarefa relacionada a uma única entidade (ex: alterar o status de uma `Task`), o agente precisa ler arquivos de 3 a 4 diretórios distintos. Como esses arquivos contêm definições de múltiplas outras entidades (ex: `domain/types.go` contém `Task`, `WorkUnit`, `AgentSession`, etc.), a janela de contexto do LLM é preenchida com informações que ele não precisa.

Isso causa três problemas centrais:
1. **Desperdício de Tokens:** Gastamos tokens enviando código irrelevante para o LLM ler.
2. **Alta Carga Cognitiva e Alucinação:** Quanto mais código desnecessário o LLM lê, maior a chance dele se confundir ou gerar código que afeta outras entidades sem querer.
3. **Quebra de Contratos:** Como as lógicas estão fisicamente distantes, o LLM frequentemente altera um modelo no `domain` mas esquece de atualizar a interface correspondente no `repository`.

## 2. Decisão

Adotaremos uma arquitetura baseada em **Vertical Slices** (ou Módulos Verticais), otimizada para agentes de IA. A partir de agora, o projeto será estruturado por **Domínio de Negócio (Feature/Module Cohesion)**, e não mais por camada técnica.

1. **Estrutura Base:** 
   O código das entidades será migrado para `internal/modules/<nome_da_entidade>/` (ou `internal/features/`).
   Dentro dessa pasta, todas as camadas daquela funcionalidade residirão juntas (tipos, serviços, repositórios).
   
   Exemplo:
   ```text
   internal/modules/task/
     models.go       (Tipos de Task)
     service.go      (Regras de negócio de Task)
     repository.go   (Persistência de Task)
     events.go       (Eventos gerados pelo domínio Task)
   ```

2. **Isolamento Estrito (Regra de Ouro):**
   Um módulo em `internal/modules/*` **NÃO PODE** importar arquivos de outro módulo em `internal/modules/*`. Eles devem ser complementamente autônomos.
   Se precisarem se comunicar, devem fazê-lo de forma assíncrona (via `core/eventstore`) ou ser orquestrados por uma camada de aplicação superior (`cmd/` ou `internal/modules/orchestrator/`), que os interliga usando dependências puras (ex: IDs genéricos em vez de structs concretas).

## 3. Consequências

- **Aumento de Eficiência do LLM:** Para dar manutenção em um módulo, o Agente de IA só precisará listar e ler o conteúdo de *uma única pasta*. O contexto será 100% focado e extremamente econômico em tokens.
- **Maior Escalabilidade do Projeto:** Em vez de arquivos monstruosos e pastas de serviços com dezenas de arquivos interligados, o crescimento do projeto será linear. Nova funcionalidade = Nova pasta no `modules/`. Nenhuma outra parte do código será impactada por colisões de arquivos.
- **Facilidade de Extração:** Se no futuro decidirmos que o módulo de `Task` deve ser um microsserviço independente (separado do `Orchestrator`), a extração é trivial: basta mover a pasta `task` para outro repositório.
- **Refatoração Inicial:** Haverá um esforço inicial de engenharia (via LLMs) para migrar a estrutura atual das pastas `domain`, `services` e `repository` para os módulos verticais, resolvendo possíveis dependências circulares.

---

## 2. Evolução — Deprecação da Camada `internal/services/`

### 2.1 Contexto adicional

A ADR 0017 (Domain Services for Operational Dependencies) foi aprovada em 2026 para estabelecer uma camada de serviços de domínio em `internal/services/` como fronteira obrigatória para comandos que alteram estado operacional.

No entanto, a arquitetura de Módulos Verticais foi aprovada posteriormente, estabelecendo que cada entidade de domínio tem seu próprio módulo autônomo em `internal/modules/<entity>/`.

### 2.2 Decisão

A ADR 0017 é marcada como **DEPRECATED**. Os princípios fundamentais continuam válidos, mas a implementação mudou:

**Princípios Mantidos (Válidos):**

- Serviços de domínio são a fronteira obrigatória para comandos que alteram estado
- CLI, TUI, GitHub, runtimes e conectores devem chamar serviços, não repositórios diretamente
- Transições de estado devem passar por serviços com validação e eventos
- Event Store permanece como fonte de verdade operacional
- Retries, timeouts e validação de entrada são responsabilidades dos serviços
- Atomicidade de múltiplas escritas relacionadas deve ser preservada

**Implementação Atual (Vertical Slices):**

- Cada entidade de domínio tem seu próprio módulo autônomo
- Comunicação Cross-Module: Módulos NUNCA importam outros módulos diretamente
- Bootstrap: Wiring de dependências via `internal/bootstrap/services.go` com adapters
- Isolamento para LLMs: Cada módulo pode ser compreendido isoladamente por agentes de IA

**Mapeamento de migração:**

| Serviço (antigo) | Módulo (novo) |
|------------------|---------------|
| `TaskService` | `internal/modules/task/service.go` |
| `RunService` | `internal/modules/run/service.go` |
| `WorkUnitService` | `internal/modules/workunit/service.go` |
| `AgentSessionService` | `internal/modules/agentsession/service.go` |
| `EventService` | `internal/core/event/service.go` |
| `TaskGraphService` | `internal/modules/taskgraph/service.go` |
| `PromptService` | `internal/modules/prompt/service.go` |
| `AgentService` | `internal/modules/agent/service.go` |
| `ReviewService` | `internal/modules/review/service.go` |
| `TriggerService` | `internal/modules/trigger/service.go` |
| `OrchestratorService` | `internal/modules/orchestrator/service.go` |

---

## 3. Padronização de Estrutura de Módulos

### 3.1 Contexto adicional

Após a adoção da arquitetura de Módulos Verticais, a migração revelou inconsistências estruturais profundas: regras de `contract.go` inventadas por módulo; arquivos obrigatórios faltando; decomposições `service_*.go` não documentadas; regras de importação de `internal/core/*` ad hoc; e falta de distinção entre regra global, regra de tipo de módulo e regra específica.

### 3.2 Decisão

Adotaremos um **padrão unificado e hierárquico** para todos os módulos em `internal/modules/*`.

**Hierarquia de Regras:**

```
REGRAS GLOBAIS      → Válidas para TODO módulo em internal/modules/*
REGRAS DE TIPO      → Válidas para um tipo de módulo (domínio, orquestração, infra)
REGRAS ESPECÍFICAS  → Válidas apenas para este módulo
```

**Arquivos Obrigatórios (MUST):**

| Arquivo | Responsabilidade |
|---------|------------------|
| `doc.go` | Package documentation, context briefing |
| `contract.go` | ModuleContract + regras hierárquicas |
| `README.md` | Propósito, file map, allowed dependencies |
| `CONTRACTS.md` | Invariantes, state machine, boundary rules |
| `models.go` | Tipos próprios (structs, enums, constants) |
| `queries.go` | SQL constants (mesmo que placeholder) |
| `repository.go` | CRUD puro, zero business logic |
| `service.go` | Lógica principal, transações, eventos |
| `events.go` | `EventTypeForStatus(status Status) string` |
| `validation.go` | Validações sintáticas de input |

**Exceção:** `orchestrator/` é módulo de coordenação, não de domínio. Não possui tabela própria, então `repository.go` pode ser placeholder, mas o arquivo **deve existir**.

**Arquivos Opcionais (MAY):**

| Arquivo | Quando usar |
|---------|-------------|
| `fetch.go` | Exportar `RequireByID` como DI para outros módulos |
| `service_<sub>.go` | `service.go` > 300 linhas |
| `types.go` | Tipos auxiliares que não cabem em `models.go` |
| `*_test.go` | Testes unitários |
| `<domain>.go` | Lógica específica (planners, detectores, composers) |

**Regra de `service_<sub>.go`:**

1. Só é permitido se `service.go` tiver **> 300 linhas**.
2. O nome deve seguir `service_<verbo>.go` (ex: `service_create.go`, `service_retry.go`).
3. Deve ser listado no `README.md` na seção "File Map".
4. Deve ter uma justificativa de 1 linha no `CONTRACTS.md`.

### 3.3 Regras de Importação Padronizadas

**Regras Globais (todos os módulos):**

1. NEVER import `internal/modules/*` for services, repositories, or business logic.
   - ALLOWED: import types (structs, enums) from another module **only** for DI interface return types.
2. NEVER import `internal/domain` for entity structs or entity enums.
   - ALLOWED: import `EventEnvelope`, `EventPriority`, checkpoint types, and generic event payloads only.
3. NEVER write SQL strings outside `queries.go`.
4. NEVER call `panic()` — always return `apperrors.Error`.
5. NEVER put business logic inside `repository.go`.
6. NEVER call another module's Service methods — use DI interfaces or `core/transition` helpers.
7. NEVER ignore errors (`_ = someCall()`) without a documented reason.
8. ALWAYS emit a domain event inside the same transaction for every mutation.

**Regras de Tipo de Módulo:**

- **Módulos de Domínio** (`task`, `run`, `workunit`, `taskgraph`, `agentsession`, `agent`, `prompt`, `trigger`, `review`):
  - Status transitions MUST call `core/statemachine.CanTransition` before mutating.
  - Terminal statuses (`completed`, `failed`, `cancelled`, `stopped`) are immutable.
  - Every mutation MUST emit exactly one domain event.
  - Input validation MUST use `core/validation` at module boundaries.

- **Módulo de Coordenação** (`orchestrator`):
  - `OrchestratorService` is the ONLY component that may coordinate cross-module operations.
  - NEVER import repositories directly — use only domain services injected via Dependencies.
  - NEVER bypass `core/statemachine` rules of individual modules.
  - Work units are executed sequentially in the first cut (parallelism is future work).

**Matriz de Importação de `internal/core/*`:**

| Pacote | O que é | Pode ser importado por | Regra |
|--------|---------|----------------------|-------|
| `core/apperrors` | Erros tipados | **Todos** | Sempre permitido |
| `core/db` | Transações, helpers SQL | **Todos** | Sempre permitido |
| `core/validation` | Validações sintáticas | **Todos** | Sempre permitido |
| `core/event` | Event envelope genérico | **Todos** | Sempre permitido |
| `core/serialization` | Marshal de payloads | Módulos que emitem eventos | Se precisar de payload |
| `core/statemachine` | Regras de transição | Módulos com state machine | Se tiver status transitions |
| `core/transition` | Helpers de transição | Módulos com state machine | Se tiver status transitions |
| `core/eventstore` | Leitura de eventos | Módulos que leem event store | Se precisar de deduplicação |
(Nenhum pacote core/* restringido — todos os módulos usam core/* conforme necessidade.)

### 3.4 Template Unificado de `contract.go`

Todo `contract.go` deve seguir a estrutura de comentários com 3 blocos hierárquicos (GLOBAL RULES, MODULE-TYPE RULES, MODULE-SPECIFIC RULES) e a seção ALLOWED/FORBIDDEN core/* IMPORTS.

Todo `README.md` deve conter uma seção "File Map" que lista **todos** os arquivos Go do módulo.

Todo `CONTRACTS.md` deve conter uma seção "File Decomposition" se houver `service_*.go`.

---

## 4. Política de Importação Hierárquica

### 4.1 Contexto adicional

A migração para Módulos Verticais criou módulos autônomos, mas agentes de IA executando a migração encontraram contradições: a regra de "nunca importar outro módulo" conflitava com a necessidade de usar tipos de entidades em interfaces DI. O código de `run/service.go` importava `taskmod` para `TaskReader`, e ainda usava `domain.WorkUnit` porque não sabia se podia importar `workunit`.

### 4.2 Decisão

Adotaremos uma política de importação **hierárquica e explícita**, com três pilares:

**Pilar 1: `internal/domain` é um pacote de infraestrutura, não de entidades**

Após a migração, `internal/domain` **não deve mais conter entity structs** (`Task`, `WorkUnit`, `Run`, etc.). Ele deve conter **apenas tipos genuinamente compartilhados** entre múltiplos módulos.

**O que FICA em `domain` (permanente):**

| Tipo | Por que fica? |
|------|---------------|
| `EventEnvelope` | Todos os módulos emitem e consomem eventos |
| `EventPriority` | Constantes usadas por todos os módulos |
| `CheckpointTrigger`, `CheckpointInput`, `HeartbeatInput`, etc. | Usados por `agentsession`, `coordination`, e runtimes |
| Payloads de eventos genéricos | "Value objects" de eventos, não entidades |

**O que SAI de `domain` (migrar para módulos):**

| Tipo | Destino |
|------|---------|
| `Task`, `TaskStatus`, `Priority`, `RiskLevel` | `modules/task` |
| `WorkUnit`, `WorkUnitStatus` | `modules/workunit` |
| `Run`, `RunStatus`, `RunResult` | `modules/run` |
| `Agent`, `AgentRuntimeType`, `AgentStatus` | `modules/agent` |
| `AgentSession`, `AgentSessionStatus` | `modules/agentsession` |
| `TaskGraph`, `TaskGraphStatus` | `modules/taskgraph` |
| `PromptFragment`, `PromptSnapshot`, `ToolsetSnapshot` | `modules/prompt` |
| `Trigger`, `TriggerStatus`, `TriggerType` | `modules/trigger` |
| `Review`, `ReviewStatus`, `ValidationGate` | `modules/review` |

**Regra absoluta:** Um módulo em `internal/modules/*` NUNCA importa `internal/domain` para obter uma **entity struct** ou **enum de entity**.

**Pilar 2: Imports entre módulos são permitidos para Interfaces DI**

Um módulo `A` pode importar um módulo `B` **se e somente se**:

1. O import é usado **exclusivamente** em uma **interface de Injeção de Dependência (DI)**.
2. O tipo importado é usado **apenas como tipo de retorno** da interface.
3. `A` **nunca** chama `b.Service`, `b.Repository`, ou qualquer função/lógica de negócio de `B`.
4. A implementação da interface é **injetada em `internal/bootstrap/`**.

**Exemplo válido:**

```go
// internal/modules/run/service.go
import taskmod "github.com/levygit837-cyber/OrchestraOS/internal/modules/task"

// TaskReader é uma interface DI. run NUNCA chama task.Service diretamente.
type TaskReader interface {
    GetByID(id string) (*taskmod.Task, error)
}
```

**Pilar 3: `orchestrator` é a exceção arquitetural para coordenação cross-module**

> Nota: `internal/core/coordination/` foi removido (ADR 0028, concluído 2026-05-17). Sua lógica foi distribuída para os módulos de domínio e para `core/transition`.

- `internal/modules/orchestrator`: Pode importar **qualquer módulo** e `internal/domain`. É o único módulo de orquestração de alto nível.
- `internal/bootstrap`: Pode importar **tudo**. É a camada de wiring/DI.
- `cmd/orchestraos`: Pode importar **tudo**. É a camada de entrada.

---

## 5. Nomenclatura e Padrões de Arquivos

### 5.1 Contexto adicional

Após a adoção da arquitetura de módulos verticais, identificamos que nomes de arquivos genéricos (`helpers.go`, `utils.go`, etc.) dificultam a navegação por LLMs. Agentes de IA dependem de nomes descritivos para inferir conteúdo sem abrir arquivos.

### 5.2 Decisão

Adotaremos padrões de nomenclatura rigorosos para `internal/core/*` e `internal/modules/*`.

**Nomes Proibidos (MUST NOT):**

| Nome Proibido | Motivo |
|---------------|--------|
| `helpers.go` | Não comunica conteúdo. É um "lixo" semântico. |
| `utils.go` | Mesmo problema. "Util" significa tudo e nada. |
| `common.go` | O que é "comum"? Comum para quem? |
| `base.go` | Sugere herança. Go não tem herança. |
| `misc.go` | "Miscelânea" é admissão de falta de coesão. |
| `kit.go` / `txkit.go` | Jargão de framework. Não descreve função. |
| `ops.go` / `eventops.go` | Abreviação vaga. "Operations" de quê? |

**Exceção única:** Um arquivo pode se chamar `<package>.go` (ex: `validation.go` em `package validation`) **apenas se** for o único arquivo do pacote. Pacotes com 2+ arquivos devem usar nomes descritivos.

**Template Unificado de `doc.go`:**

Cada `doc.go` deve seguir a estrutura:
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

### 5.3 Consequências

- **Zero ambiguidade para LLMs:** Nomes de arquivos comunicam conteúdo.
- **Coesão por responsabilidade:** Cada pacote faz uma coisa. Se um pacote precisa de 7 arquivos com nomes diferentes, isso é um sinal de que talvez sejam 2 pacotes.

---

## Apêndice A: Histórico de Evolução

| Data | Evento | ADR Original |
| --- | --- | --- |
| 2026-05-11 | Decisão de migrar para Vertical Slices | ADR 0022 |
| 2026-05-13 | Deprecação formal da ADR 0017; mapeamento de serviços → módulos | ADR 0024 |
| 2026-05-16 | Padronização unificada de estrutura, template contract.go e regras hierárquicas | ADR 0025 |
| 2026-05-17 | Política de importação refinada: 3 pilares (domain=infra, DI permitido, exceções) | ADR 0026 |
| 2026-05-17 | Padrões de nomenclatura e nomes proibidos de arquivos | ADR 0028 |
| 2026-05-17 | Todos consolidados neste documento único | — |

## Apêndice B: Alternativas Consideradas (Arquitetura)

### Alternativa A: Clean Architecture / Layered Architecture (Nossa estrutura atual)
- **Como funciona:** Separa o código por camadas técnicas (Ports & Adapters, Services, Repositories).
- **Escalabilidade:** Escala moderadamente bem para humanos, pois padroniza onde encontrar cada tipo técnico de arquivo. Porém, arquivos compartilhados (como um `types.go` global) tornam-se gargalos.
- **Para LLMs:** É a **pior opção**. Um LLM lê código baseando-se em requisições de *funcionalidades*. Para adicionar o recurso de "arquivamento de tarefas", o LLM tem que editar 4 lugares distintos, lendo dezenas de arquivos de contexto alheio. O risco de quebrar o estado global e a ineficiência de prompt a descartam.

### Alternativa B: Microsserviços
- **Como funciona:** Cada funcionalidade é um projeto isolado rodando em seu próprio processo/contêiner.
- **Escalabilidade:** Escalação máxima de infraestrutura e equipe.
- **Para LLMs:** É excelente em termos de isolamento de contexto (um LLM só vê o código do microsserviço). No entanto, o LLM precisaria entender e versionar contratos de rede, chamadas gRPC/REST, Dockerfiles complexos e tratamento de falhas distribuídas. Para a fase MVP do OrchestraOS, é um *over-engineering* pesado que travaria a velocidade do Agente.

### A Vencedora: Vertical Slice Architecture (Modular Monolith)
- **Como funciona:** Mantém o projeto como um monólito (fácil de compilar, testar e debugar localmente), mas divide o código internamente como se fossem microsserviços lógicos (Módulos).
- **Escalabilidade:** É a arquitetura mais **sustentável** e escalável a longo prazo para o nosso momento. Cada nova funcionalidade apenas adiciona uma "fatia" (pasta) nova ao bolo, sem inchar as camadas existentes.
- **Para LLMs:** Junta "o melhor dos dois mundos". Fornece o isolamento cirúrgico de contexto (o Agente só lê a pasta específica) sem a sobrecarga operacional de gerenciar rede e infraestrutura de microsserviços. É, disparada, a arquitetura ideal para codebases totalmente operados por IA.
