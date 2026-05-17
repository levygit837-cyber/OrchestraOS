# 0027. Directory Semantic Renaming — Nomes que Refletem Responsabilidade

**Data:** 2026-05-17
**Status:** Proposed
**Relacionada:** ADR-0022, ADR-0025, ADR-0026

---

## 1. Contexto

Após a adoção do ADR-0022 (Vertical Slice Architecture), do ADR-0025 (Module Standardization) e do ADR-0026 (Module Import Policy), a estrutura de `internal/` está arquiteturalmente coerente. Porém, **alguns diretórios ainda carregam nomes que não refletem sua responsabilidade real**, gerando:

- **Ambiguidade para novos desenvolvedores** — `internal/core/event/` não define eventos, mas faz *append* deles com validação.
- **Conflitos de importação** — `event` é tão genérico que quase todos os arquivos precisam de alias (`eventmod`), indicando que o nome não é descritivo o suficiente.
- **Paralisia de análise em agentes de IA** — um agente que vê `orchestrator/` espera encontrar um *orquestrador de agentes* (decide qual tarefa executar, aloca recursos), mas encontra um *workflow engine* que executa uma única task sequencialmente.
- **Referências internas de renomeação não executadas** — o próprio `doc.go` de `orchestrator/` já documenta desde 2026-05-17 que o módulo "will be renamed to runner/ or taskflow/", mas nenhuma ADR formaliza isso.

A migração de entity structs de `internal/domain/` para `internal/modules/*` (ADR-0026, Pilar 1) é o trabalho prioritário, mas **os nomes dos diretórios devem estar corretos antes ou durante essa migração**, pois um agente renomeando `domain.WorkUnit` para `workunit.WorkUnit` precisa saber que `workunit/` é o nome final, não um placeholder.

---

## 2. Decisão

Renomearemos os diretórios cujos nomes são semanticamente incorretos ou ambíguos. Diretórios com nomes já claros permanecem inalterados.

### 2.1 Diretórios a Renomear

| Diretório Atual                  | Novo Nome                  | Por que renomear?                                                                                                                                                                                            |
|----------------------------------|----------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `internal/modules/orchestrator/` | `internal/modules/taskflow/` | Não é um "orchestrator" no sentido de DDD/Orchestration. É um *workflow engine* que executa uma task do início ao fim. O nome "orchestrator" será reservado para um futuro módulo de alto nível (Agent Orchestrator / Director). |
| `internal/core/event/`           | `internal/core/eventappend/` | O pacote não *define* eventos nem seus tipos. Ele implementa o **serviço de append idempotente** com validação de schema e deduplicação. O nome `event` é tão genérico que obriga alias em todo import.         |

### 2.2 Diretórios que Permanecem com Movimentação Interna

| Diretório                   | Ação                                                | Por que?                                                                                                                                                                                                         |
|-----------------------------|-----------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `internal/core/transition/` | Manter nome; mover `eventops.go` para `eventappend/` | O nome `transition` é semanticamente correto para *state-machine transition helpers*. O arquivo `eventops.go`, porém, contém `AppendServiceEvent` e `AppendTransition` — essas são **operações de event store**, não de transição de estado. Devem viver junto ao serviço de append. |

### 2.3 Diretórios que NÃO Serão Renomeados

| Diretório                    | Por que permanece?                                                                                                                                                                              |
|------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `internal/core/`             | O nome é vazio semanticamente ("core" não diz o que contém), mas a renomeação teria impacto em ~40 arquivos. O custo/benefício não justifica. Os subpacotes já têm nomes claros (exceto `event/`, tratado acima). |
| `internal/core/coordination/`  | ADR-0026 reafirmou o papel de cross-module coordination. O nome agora é semanticamente correto.                                                                                                |
| `internal/core/db/`            | `db` é genérico, mas aceitável para "database infrastructure helpers". Não há ambiguidade com outros pacotes.                                                                                    |
| `internal/core/apperrors/`    | Nome claro e específico.                                                                                                                                                                        |
| `internal/core/eventstore/`   | Nome claro; diferenciado de `event/` (que será `eventappend/`).                                                                                                                                  |
| `internal/core/statemachine/` | Nome claro e descritivo.                                                                                                                                                                        |
| `internal/core/serialization/` | Nome claro.                                                                                                                                                                                     |
| `internal/core/validation/`  | Nome claro.                                                                                                                                                                                     |
| `internal/modules/agent/`    | Nome da entidade (`Agent`). Correto.                                                                                                                                                            |
| `internal/modules/agentsession/` | Nome composto é aceitável em Go. A alternativa `session/` é muito genérica.                                                                                                                      |
| `internal/modules/prompt/`   | Nome do domínio (`Prompt`). Correto.                                                                                                                                                            |
| `internal/modules/review/`   | Nome do domínio (`Review`). Correto.                                                                                                                                                            |
| `internal/modules/run/`      | Nome da entidade (`Run`). Correto.                                                                                                                                                              |
| `internal/modules/task/`     | Nome da entidade (`Task`). Correto.                                                                                                                                                             |
| `internal/modules/taskgraph/` | Nome da entidade (`TaskGraph`). Composto mas aceitável em Go.                                                                                                                                  |
| `internal/modules/trigger/`  | Nome do domínio (`Trigger`). Correto.                                                                                                                                                           |
| `internal/modules/workunit/` | Nome da entidade (`WorkUnit`). Composto mas aceitável em Go.                                                                                                                                    |
| `internal/bootstrap/`        | Nome semanticamente correto: é o pacote de wiring/DI.                                                                                                                                          |
| `internal/domain/`           | Será reduzido a tipos genuinamente compartilhados (ADR-0026). O nome `domain` como "infraestrutura de eventos" é um resíduo semântico da arquitetura em camadas, mas sua renomeação é discutida na seção 6 (Alternativas). |

---

## 3. Análise das Alternativas por Diretório

### 3.1 `orchestrator` → `taskflow`

#### Alternativa A: `runner/`

- **Prós:** Curto, descritivo de "executa coisas".
- **Contras:** Colide semanticamente com o módulo `run/` (executa uma WorkUnit). Um desenvolvedor não distingue `runner/` de `run/` sem ler o doc.
- **Veredicto:** Rejeitada. Ambiguidade com `run/`.

#### Alternativa B: `workflow/`

- **Prós:** Termo consagrado em DDD/ES para "motor de execução de processos".
- **Contras:** Muito amplo. O projeto já tem `taskgraph/` (define o workflow) e `run/` (executa uma unidade). `workflow/` sugere que contém a definição do workflow, não a execução.
- **Veredicto:** Rejeitada. Amplitude excessiva.

#### Alternativa C: `taskexecutor/`

- **Prós:** Descritivo exato.
- **Contras:** Muito longo para um nome de pacote Go (13 caracteres). Quebra a fluidez de leitura em imports.
- **Veredicto:** Rejeitada. Verbosidade.

#### Alternativa D (Vencedora): `taskflow/`

- **Prós:** Composto de `task` (domínio conhecido) + `flow` (execução sequencial). Descreve exatamente "o fluxo de execução de uma task". Não colide com nenhum módulo existente. Curto o suficiente (8 caracteres).
- **Contras:** Nenhum significativo.
- **Veredicto:** Aceita.

### 3.2 `event` → `eventappend`

#### Alternativa A: `eventlog/`

- **Prós:** Sugere persistência e leitura de eventos.
- **Contras:** O pacote não é um log. Não oferece leitura/querieda. É um serviço de *escrita* com validação.
- **Veredicto:** Rejeitada. Semântica invertida.

#### Alternativa B: `eventjournal/`

- **Prós:** Termo técnico correto em Event Sourcing (append-only journal).
- **Contras:** Jargão de nicho. Um desenvolvedor Go comum não associa "journal" a "append de eventos".
- **Veredicto:** Rejeitada. Jargão excessivo.

#### Alternativa C (Vencedora): `eventappend/`

- **Prós:** Composto de `event` + `append`. Descreve o serviço exato: append de eventos. Diferencia claramente de `eventstore/` (persistência). Elimina a necessidade de alias (`eventappend` é único no projeto).
- **Contras:** Composto, mas aceitável em Go (ex: `strings.Builder`, `httptest.Server`).
- **Veredicto:** Aceita.

---

## 4. Consequências

### Positivas

- **Eliminação de aliases de import obrigatórios:** `eventappend/` nunca precisará de alias; `taskflow/` é único.
- **Zero ambiguidade para LLMs:** Um agente lendo `taskflow/doc.go` sabe imediatamente que não é um "orquestrador de agentes".
- **Reserva semântica de `orchestrator/`:** O nome fica livre para um futuro módulo `director/` ou `orchestrator/` que de fato orquestre múltiplas tasks.
- **Coesão em `core/`:** `eventappend/` concentra toda a lógica de append (incluindo o que hoje está em `transition/eventops.go`).

### Negativas

- **Impacto em imports:** `orchestrator/` é importado por `cmd/orchestraos/`, `bootstrap/`, `tests/`, e referenciado em múltiplos ADRs. A renomeação exige busca-e-substituição em ~15 arquivos.
- **Impacto em scripts e ferramentas:** Qualquer script que referencie `internal/modules/orchestrator/` quebrará.
- **Matriz de allowed imports em `module_boundaries_test.go`:** Deverá ser atualizada.
- **Contratos de módulo:** `contract.go`, `README.md` e `CONTRACTS.md` de `orchestrator/` precisam ser regenerados com o novo nome.

---

## 5. Checklist de Implementação

- [ ] Criar `internal/modules/taskflow/` com todo o conteúdo de `orchestrator/`.
- [ ] Atualizar `doc.go` de `taskflow/` para remover a referência de "future rename".
- [ ] Atualizar `ModuleContract.Name` em `contract.go` para `"taskflow"`.
- [ ] Atualizar todos os imports de `.../orchestrator` para `.../taskflow` em:
  - `cmd/orchestraos/`
  - `internal/bootstrap/`
  - `internal/core/coordination/`
  - `tests/integration/`
  - `tests/architecture/`
- [ ] Renomear `internal/core/event/` para `internal/core/eventappend/`.
- [ ] Atualizar todos os imports de `.../core/event` para `.../core/eventappend`.
- [ ] Mover `internal/core/transition/eventops.go` para `internal/core/eventappend/ops.go`.
- [ ] Atualizar `module_boundaries_test.go` para refletir os novos caminhos.
- [ ] Executar `go test ./...`, `go build ./...`, `./scripts/lint.sh`.
- [ ] Commit via `./scripts/safe-commit.sh`.

---

## 6. Alternativas Consideradas (Macro)

### Alternativa A: Renomear `internal/core/` para `internal/infra/` ou `internal/platform/`

- **Como funciona:** Substituir `core` por um nome mais descritivo.
- **Problema:** Impacto em ~40 arquivos. O nome `core` é vazio, mas não é incorreto — é convencional em Go para "shared internal packages". O retorno não justifica o custo.
- **Descartada:** Custo/benefício desfavorável.

### Alternativa B: Renomear `internal/domain/` para `internal/shared/` ou `internal/events/`

- **Como funciona:** `domain/` deixará de conter entity structs e virará um pacote de infraestrutura de eventos/checkpoints.
- **Problema:** Alto impacto em todos os módulos que ainda importam `domain` durante a transição. Deveria ser feito *depois* da migração A02-A09 estar 100% completa, quando `domain/` tiver apenas `EventEnvelope`, `EventPriority` e tipos de checkpoint.
- **Descartada para esta ADR:** Será tratada em ADR futura após a conclusão da migração de tipos.

### Alternativa C: Não renomear nada; apenas documentar

- **Como funciona:** Manter nomes atuais e adicionar comentários explicativos.
- **Problema:** O próprio `doc.go` de `orchestrator/` já admite que o nome está errado e precisa ser renomeado. Documentação não resolve ambiguidade de import.
- **Descartada:** A ambiguidade é real e documentada internamente.

---

## 7. Notas de Implementação

- A renomeação de `orchestrator/` → `taskflow/` pode ser feita via `git mv` para preservar histórico.
- A renomeação de `event/` → `eventappend/` pode ser feita via `git mv` para preservar histórico.
- Os arquivos `transition/eventops.go` devem ser movidos com `git mv` para manter blame.
- Recomenda-se executar esta ADR **após** a conclusão da migração A02-A09 (domain types → modules) para evitar renomear arquivos que ainda contêm referências a `domain.Run`, `domain.Task`, etc. Ou, alternativamente, executar **antes** para que os novos tipos locais sejam criados já no diretório corretamente nomeado.
