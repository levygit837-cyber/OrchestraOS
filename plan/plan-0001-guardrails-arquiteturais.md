# Plan: Guardrails Arquiteturais — Mitigação de Gargalos da Vertical Slice

## Contexto

A arquitetura modular foi concluída com sucesso: 8 módulos em `internal/modules/`, 6 pacotes em `internal/core/`, testes passando, e documentação (`doc.go`, `README.md`, `CONTRACTS.md`) em todos os pacotes. Porém, como diagnosticado na análise crítica, **documentação não é enforcement**. Hoje não existem:

- Testes de arquitetura que quebram quando um módulo importa outro indevidamente
- CI que valide builds, lint e regras arquiteturais
- Linter de imports (`depguard`, `forbidigo`)
- Verificação automática de que `CONTRACTS.md` está sincronizado com o código
- Template padronizado para novos módulos
- Guia de estilo codificado (além do `AGENTS.md` textual)

Sem esses guardrails, LLMs operando no código vão inevitavelmente:
1. Criar imports cíclicos ou violar fronteiras de módulo
2. Deixar `CONTRACTS.md` desatualizado após mudar regras no código
3. Introduzir padrões inconsistentes entre módulos
4. Commitar código que compila mas quebra invariantes arquiteturais

Este plano entrega **guardrails técnicos enforceable** que transformam a documentação em lei.

---

## Diagnóstico dos Gargalos Identificados

### Gargalo 1: Import de `modules/event` dentro de `modules/agent` (violação real)
`internal/modules/agent/runtime.go` usa `*event.Envelope` na interface `Runtime`. Como `event.Envelope` é apenas um alias para `domain.EventEnvelope`, o correto é usar `*domain.EventEnvelope` diretamente, tornando `agent` um módulo folha. **Impacto:** quebra a regra de módulos folha e cria uma dependência desnecessária.

### Gargalo 2: Ausência de testes de arquitetura
Não existe nenhum teste Go que verifique se `modules/run` importa `modules/task/service.go` em vez de apenas `modules/task` (repositório). O compilador permite qualquer import válido. **Impacto:** drift arquitetural silencioso.

### Gargalo 3: Múltiplas fontes de verdade para state machine
As transições válidas estão codificadas em `core/statemachine/statemachine.go` **e** copiadas manualmente nos 4 `CONTRACTS.md` dos módulos com state machine. **Impacto:** quando o código muda, os contratos envelhecem.

### Gargalo 4: Sem CI / sem lint
Não existe `.github/workflows/`, `.golangci.yml`, nem `Makefile`. **Impacto:** uma LLM pode commitar código que não passa em `go vet`, tem imports não utilizados, ou viola regras de estilo.

### Gargalo 5: Sem template de novo módulo
Criar um novo módulo exige copiar manualmente a estrutura de outro. **Impacto:** inconsistência entre módulos, esquecimento de `README.md`/`CONTRACTS.md`/`queries.go`.

---

## Estratégia de Mitigação

A abordagem é **incremental e verificável**: cada fase entrega valor imediato e pode ser revertida independentemente.

### Abordagem Recomendada: "Guardrails como Código"

Em vez de depender de documentação textual, transformar todas as regras arquiteturais em:
1. **Testes Go** que falham (`t.Errorf`) quando violadas
2. **Linter** que falha (`exit 1`) quando violadas
3. **Scripts** que falham quando `CONTRACTS.md` diverge do código
4. **CI** que impede merge quando qualquer guardrail falha

### Alternativa Descartada: "Apenas Documentação + Revisão Humana"
Revisar manualmente cada PR gerado por LLM não escala. O objetivo é autonomia de nível 3-4 (IA abre PRs e executa tarefas). Revisão humana de regras mecânicas (imports, presença de arquivos) é desperdício.

---

## Fases de Implementação

### Fase 1: Correções Arquiteturais Imediatas
**Objetivo:** Eliminar violações de import existentes antes de codificar as regras.

1. **Remover dependência `agent → event`**
   - Substituir `*event.Envelope` por `*domain.EventEnvelope` em `agent/runtime.go`, `agent/fake_runtime.go`, `agent/gemini_runtime.go`
   - Remover import de `internal/core/event` de todos os arquivos em `agent/`
   - Ajustar `tests/integration/fake_runtime_test.go` e `tests/integration/interaction_test.go` se necessário
   - **Validação:** `go test ./...` passa; teste de arquitetura (da Fase 2) confirma que `agent` não importa nenhum `modules/*`

2. **Verificar e documentar imports entre módulos legítimos**
   - `task` importa `run` e `workunit` (repositórios apenas — legítimo)
   - `run` importa `workunit` (para `EventTypeForStatus` e validações — legítimo)
   - `agentsession` importa `run` (repositório apenas para pause — legítimo)
   - `taskgraph` importa `task` e `workunit` (repositórios — legítimo)
   - `prompt` importa 4 módulos (repositórios para composição — legítimo)
   - **Documentar essas dependências permitidas** no teste de arquitetura da Fase 2 como "allowlist"

---

### Fase 2: Testes de Arquitetura (`tests/architecture/`)
**Objetivo:** Criar testes Go que quebram quando fronteiras são violadas.

1. **Criar `tests/architecture/module_boundaries_test.go`**
   - Usar `golang.org/x/tools/go/packages` para analisar AST de cada módulo
   - Regra: `modules/agent` não pode importar nenhum `modules/*`
   - Regra: `modules/event` não pode importar nenhum `modules/*`
   - Regra: nenhum módulo pode importar `modules/*/service.go` diretamente (exceto `bootstrap`)
   - Regra: `modules/*` só pode importar `modules/*` se estiver em allowlist documentada
   - **Validação:** rodar `go test ./tests/architecture/...` — deve passar com o estado atual

2. **Criar `tests/architecture/module_files_test.go`**
   - Regra: todo diretório em `internal/modules/*` deve conter `README.md` e `CONTRACTS.md`
   - Regra: todo diretório em `internal/modules/*` deve conter `queries.go`
   - Regra: todo diretório em `internal/modules/*` deve conter `doc.go`
   - **Validação:** deve passar com o estado atual (já criamos todos)

3. **Criar `tests/architecture/queries_purity_test.go`**
   - Regra: `queries.go` em qualquer módulo só pode conter constantes/variáveis `string` (SQL)
   - Não pode conter funções, imports de pacotes não-stdlib, ou lógica
   - **Validação:** deve passar com o estado atual

---

### Fase 3: Linter de Imports (`golangci-lint`)
**Objetivo:** Quebrar o build local e em CI quando imports proibidos são introduzidos.

1. **Instalar `golangci-lint`**
   - Script `scripts/install-tools.sh` ou instrução em `AGENTS.md`
   - Versão pinada (ex: v1.59) para reprodutibilidade

2. **Criar `.golangci.yml`**
   - `depguard`: bloquear imports de `internal/services` e `internal/repository` (pacotes removidos)
   - `forbidigo`: proibir `fmt.Println`, `panic` em código de produção
   - `errcheck`: forçar tratamento de erros
   - `goimports`: garantir formatação de imports
   - `gofmt`: formatação padrão
   - `govet`: análise estática básica
   - `staticcheck`: análise avançada
   - Excluir `vendor/`, `migrations/` de algumas regras se necessário

3. **Criar `scripts/lint.sh`**
   - Wrapper que roda `golangci-lint run ./...`
   - Retorna `exit 1` se houver violações
   - **Validação:** rodar no estado atual — deve passar (ou corrigir pequenas violações encontradas)

---

### Fase 4: Script de Verificação de CONTRACTS
**Objetivo:** Detectar drift entre `CONTRACTS.md` e o código fonte de verdade.

1. **Criar `scripts/verify-contracts.sh`** (ou `tests/architecture/contracts_sync_test.go`)
   - Extrair state machine de `core/statemachine/statemachine.go` via regex/AST
   - Para cada módulo com state machine (`task`, `workunit`, `run`, `agentsession`):
     - Verificar se o `CONTRACTS.md` contém todas as transições listadas no código
     - Verificar se não contém transições que não existem no código
   - Verificar se todos os módulos em `internal/modules/` possuem `README.md` e `CONTRACTS.md`
   - Verificar se a seção "Allowed Dependencies" em cada `README.md` reflete os imports reais do pacote
   - **Output:** diff claro mostrando o que está desatualizado
   - **Retorno:** `exit 0` se sincronizado, `exit 1` se drift detectado
   - **Validação:** rodar no estado atual — deve passar (ou corrigir divergências)

2. **Decisão de design: State Machine como fonte única de verdade**
   - Opção A: Gerar `CONTRACTS.md` automaticamente a partir de `statemachine.go` (complexo, requer template engine)
   - Opção B: Manter duplicação mas verificar via script/teste (mais simples, aceita pequeno overhead)
   - **Recomendação:** Opção B por agora. A duplicação é aceitável se houver teste que detecta drift. Opção A pode ser implementada futuramente se o projeto crescer além de 15 módulos.

---

### Fase 5: Template de Novo Módulo
**Objetivo:** Padronizar a criação de módulos para evitar inconsistência.

1. **Criar `docs/templates/module/`**
   - `doc.go` — template com placeholders `{{MODULE_NAME}}`, `{{RESPONSIBILITY}}`
   - `README.md` — template com seções padrão
   - `CONTRACTS.md` — template com invariants, state machine, boundary rules genéricos
   - `models.go` — template vazio com `package` e aliases comuns
   - `queries.go` — template vazio
   - `repository.go` — template com estrutura base (`Repository struct`, `NewRepository`)
   - `service.go` — template com `Service struct`, `NewService`, `Create` stub
   - `.gitkeep` para `catalog/` se o módulo for de prompt

2. **Criar `scripts/new-module.sh`**
   - Uso: `./scripts/new-module.sh <nome-do-modulo>`
   - Copia templates de `docs/templates/module/` para `internal/modules/<nome>/`
   - Substitui placeholders básicos (`{{MODULE_NAME}}`, `{{PACKAGE}}`)
   - Cria `go test ./internal/modules/<nome>` stub
   - Atualiza `internal/bootstrap/services.go` com comentário indicando onde adicionar o novo serviço
   - **Validação:** rodar `./scripts/new-module.sh testmod` e verificar estrutura; rodar `go test ./internal/modules/testmod` (deve passar com stub)

---

### Fase 6: CI Mínimo (GitHub Actions)
**Objetivo:** Impedir que código quebrando guardrails seja mergeado.

1. **Criar `.github/workflows/ci.yml`**
   - Trigger: `push` em `main`, `pull_request`
   - Jobs:
     - `build`: `go build ./...`
     - `test`: `go test ./... -race -count=1`
     - `vet`: `go vet ./...`
     - `lint`: `golangci-lint run ./...` (usando cache de actions)
     - `architecture`: `go test ./tests/architecture/... -v`
     - `contracts`: `./scripts/verify-contracts.sh`
   - Usar `actions/setup-go@v5` com versão do `go.mod`
   - Usar `golangci/golangci-lint-action@v6` para lint

2. **Criar `.github/workflows/pr-checklist.yml`** (opcional, mas recomendado)
   - Comentário automático em PRs lembrando o autor de ler `CONTRACTS.md`
   - Não é bloqueante, é educativo

3. **Validação:** push em branch de teste e verificar se todos os jobs passam

---

### Fase 7: Guia de Estilo Codificado
**Objetivo:** Substituir regras textuais por regras enforceable.

1. **Criar `docs/development/CODING_STANDARDS.md`**
   - Regras já presentes em `AGENTS.md` + regras novas descobertas durante o refactor
   - Seção específica para "Padrões de Módulo" (obrigatório: README.md, CONTRACTS.md, queries.go, doc.go)
   - Seção "Anti-padrões" com exemplos de código (before/after)

2. **Atualizar `AGENTS.md`**
   - Adicionar referência ao `CODING_STANDARDS.md`
   - Adicionar instrução: "Antes de criar um novo módulo, execute `./scripts/new-module.sh`"
   - Adicionar instrução: "Após modificar `core/statemachine/statemachine.go`, execute `./scripts/verify-contracts.sh`"

3. **Regras adicionais no `.golangci.yml`**
   - `gocritic`: detecta padrões não idiomáticos
   - ` ineffassign`: detecta atribuições inúteis
   - `misspell`: detecta erros de digitação em comentários
   - `nolintlint`: evita supressões abusivas de lint

---

## Critérios de Aceite

- [ ] `modules/agent` não importa nenhum `modules/*`
- [ ] `go test ./tests/architecture/...` passa e falha se um import proibido é introduzido
- [ ] `golangci-lint run ./...` passa no estado atual e bloqueia imports proibidos futuros
- [ ] `./scripts/verify-contracts.sh` passa e detecta drift entre `statemachine.go` e `CONTRACTS.md`
- [ ] `./scripts/new-module.sh foo` cria um módulo funcional e testável em `internal/modules/foo/`
- [ ] CI executa build, test, lint, architecture tests e contract verification em todo push/PR
- [ ] Todos os módulos existentes continuam passando em `go test ./...`

---

## Riscos e Mitigações

| Risco | Mitigação |
|---|---|
| Testes de arquitetura ficam lentos (analisar AST de todo o projeto) | Cachear resultados de `packages.Load` entre testes; limitar análise a `internal/modules/*` e `internal/core/*` |
| `golangci-lint` tem falso-positivo em código gerado ou migrations | Configurar `skip-dirs` em `.golangci.yml` para `migrations/`, `vendor/` |
| Script `verify-contracts.sh` quebra com mudanças de formatação em `CONTRACTS.md` | Usar normalização (trim, lowercase) antes de comparar; não exigir match exato de markdown |
| LLM ignora CI e commita direto na main | Proteger branch `main` com ruleset que exige PR + CI passando |
| Template de módulo fica desatualizado | Incluir teste de arquitetura que valida estrutura de módulos recém-criados |

---

## Estimativa de Esforço

| Fase | Tempo Estimado | Complexidade |
|---|---|---|
| Fase 1: Correções Imediatas | 15 min | Baixa |
| Fase 2: Testes de Arquitetura | 45 min | Média |
| Fase 3: Linter | 20 min | Baixa |
| Fase 4: Verificação de CONTRACTS | 30 min | Média |
| Fase 5: Template de Módulo | 20 min | Baixa |
| Fase 6: CI | 20 min | Baixa |
| Fase 7: Guia de Estilo | 15 min | Baixa |
| **Total** | **~2h 45min** | **Média** |

---

## Notas de Execução

- Cada fase é independente; pode ser revertida sem afetar as outras.
- A ordem recomendada é: Fase 1 → Fase 2 → Fase 3 → Fase 4 → Fase 5 → Fase 7 → Fase 6.
- Fase 6 (CI) por último porque depende dos artefatos das fases anteriores.
- Após cada fase: rodar `go test ./...` e garantir que não houve regressão.
