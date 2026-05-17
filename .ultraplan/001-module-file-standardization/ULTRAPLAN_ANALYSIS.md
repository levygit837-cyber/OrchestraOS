# ULTRAPLAN_ANALYSIS: Padronização de Arquivos por Módulos

## Fase 0: Premissa Validada

> Entendi que você quer padronizar a estrutura de arquivos obrigatórios em todos os módulos de `internal/modules/*`, garantindo que cada módulo contenha a base mínima definida pela ADR-0022, e criar um plano para corrigir as inconsistências atuais. Correto?

**Premissa assumida:** Sim. O objetivo é garantir que todos os módulos sigam o padrão de 10 arquivos obrigatórios e que quaisquer desvios sejam identificados e corrigidos.

---

## Fase 1: Snapshot do Estado Atual

### Estrutura de Módulos

Existem **10 módulos** em `internal/modules/`:

| Módulo | Tipo | Status |
|--------|------|--------|
| `agent` | Domínio | Completo (10/10) |
| `agentsession` | Domínio | Completo (10/10) |
| `orchestrator` | Coordenação | Completo (10/10) |
| `prompt` | Domínio | **Incompleto (8/10)** |
| `review` | Domínio | Completo (10/10) |
| `run` | Domínio | Completo (10/10) |
| `task` | Domínio | Completo (10/10) |
| `taskgraph` | Domínio | Completo (10/10) |
| `trigger` | Domínio | Completo (10/10) |
| `workunit` | Domínio | Completo (10/10) |

### Arquivos Obrigatórios por Módulo (ADR-0022, Seção 3.2)

| Arquivo | Responsabilidade | prompt Status |
|---------|------------------|---------------|
| `doc.go` | Package documentation | ✅ |
| `contract.go` | ModuleContract + regras hierárquicas | ✅ |
| `README.md` | Propósito, file map, allowed dependencies | ✅ |
| `CONTRACTS.md` | Invariantes, state machine, boundary rules | ✅ |
| `models.go` | Tipos próprios (structs, enums, constants) | ✅ |
| `queries.go` | SQL constants | ✅ |
| `repository.go` | CRUD puro, zero business logic | ✅ |
| `service.go` | Lógica principal, transações, eventos | ✅ |
| `events.go` | `EventTypeForStatus(status Status) string` | ❌ **FALTA** |
| `validation.go` | Validações sintáticas de input | ❌ **FALTA** |

### Problemas Específicos no Módulo `prompt`

1. **Falta `events.go`**: O módulo prompt emite eventos (visto em `service.go` — `"prompt.snapshot_created"` e `"toolset.snapshot_created"`), mas não tem uma função padronizada `EventTypeForStatus`. O módulo prompt não tem um `Status` enum tradicional como os outros módulos (não gerencia lifecycle states), mas a ADR-0022 ainda exige o arquivo.

2. **Falta `validation.go`**: O módulo prompt faz validação inline em `service.go` e `composer.go`, mas não tem um arquivo dedicado. Isso viola a regra "Input validation MUST use core/validation at module boundaries."

3. **Importações problemáticas em `service.go`**: O arquivo importa diretamente `internal/modules/run`, `internal/modules/task`, `internal/modules/workunit`, `internal/modules/agentsession` — isso viola a ADR-0022 Seção 4.2 Pilar 2. O certo seria usar interfaces DI com tipos importados apenas como retorno.

4. **Arquivo `types.go` existe** (fora do padrão): O módulo prompt tem `types.go` e `repository_snapshot.go` que fogem do padrão. O README.md já reconhece isso como "legacy — will merge into models.go".

### Arquivos Opcionais (`fetch.go`) — Estado

| Módulo | fetch.go | Justificativa |
|--------|----------|---------------|
| agent | ❌ | Não expõe RequireByID |
| agentsession | ✅ | Expõe RequireByID |
| orchestrator | ❌ | Não precisa (coordenação) |
| prompt | ❌ | Não expõe RequireByID |
| review | ❌ | Não expõe RequireByID |
| run | ✅ | Expõe RequireByID |
| task | ✅ | Expõe RequireByID |
| taskgraph | ❌ | Não expõe RequireByID |
| trigger | ✅ | Expõe RequireByID |
| workunit | ✅ | Expõe RequireByID |

Isso está correto — `fetch.go` é opcional e só deve existir quando o módulo exporta `RequireByID` para DI de outros módulos.

---

## Fase 2: Reflexão — Que Problema Real Estamos Resolvendo?

### Problema Principal

A **inconsistência estrutural** gera:
1. **Falha de ferramentas automatizadas**: Scripts que assumem os 10 arquivos obrigatórios quebram no `prompt`.
2. **Carga cognitiva para LLMs**: Quando um agente de IA lê o módulo `prompt`, ele espera encontrar `events.go` e `validation.go` baseado no template e na ADR-0022. A ausência gera confusão.
3. **Degradação gradual**: Se um módulo foge do padrão, outros tendem a seguir o exemplo, erodindo a arquitetura.

### Problema Secundário

O módulo `prompt` tem **importações diretas de outros módulos** em `service.go`, violando a ADR-0022 Seção 4.2. Isso é um problema arquitetural mais grave do que arquivos faltando.

---

## Fase 3: Alternativas

### Alternativa A: Criar arquivos stub mínimos no prompt

- Criar `events.go` com `EventTypeForStatus` usando um status enum mínimo (ou adaptado para prompt)
- Criar `validation.go` com validações básicas extraídas de `service.go`/`composer.go`
- **Pros**: Rápido, resolve a inconsistência imediata, baixo risco
- **Cons**: Não resolve o problema de importações diretas; `events.go` pode parecer forçado já que prompt não tem lifecycle states

### Alternativa B: Refatorar prompt + criar arquivos obrigatórios

- Criar `events.go` e `validation.go`
- Extrair interfaces DI de `service.go` para eliminar imports diretos de outros módulos
- Mover lógica de `types.go` para `models.go` (consolidar)
- Mover lógica de `repository_snapshot.go` para `repository.go`
- **Pros**: Resolve problema raiz, alinha totalmente com ADR-0022
- **Cons**: Mais complexo, maior risco de regressão, toca em muitos arquivos

### Alternativa C: Criar um script de auditoria + corrigir automaticamente

- Criar script `scripts/verify-module-structure.sh` que verifica os 10 arquivos obrigatórios
- Rodar em CI para prevenir regressões futuras
- Corrigir prompt manualmente
- **Pros**: Previne problemas futuros, automatiza verificação
- **Cons**: Não resolve o problema atual por si só

### Alternativa D: Revisar a ADR-0022 para exemptar módulos sem state machine

- Argumentar que `prompt` não precisa de `events.go` porque não tem status lifecycle
- **Pros**: Evita arquivo forçado
- **Cons**: Quebra a uniformidade que é o propósito da padronização; cria precedente perigoso

---

## Fase 4: Análise Comparativa

| Critério | Alt A (Stub) | Alt B (Refatorar) | Alt C (Auditoria) | Alt D (Exempt) |
|----------|-------------|-------------------|-------------------|----------------|
| Complexidade | Baixa | Alta | Baixa | Muito Baixa |
| Tempo Estimado | 30min | 2-3h | 1h | 15min |
| Riscos Principais | events.go sem propósito claro | Regressão em service.go | Nenhum | Erosão do padrão |
| Testabilidade | Boa | Requer testes extras | N/A | Boa |
| Alinhamento com Projeto | Moderado | Forte | Forte | Fraco |
| Resolve problema raiz? | Não | Sim | Parcialmente | Não |

**Decisão**: Escolher **Alternativa B + C** (refatorar prompt + criar auditoria).

**Justificativa**:
- O módulo `prompt` é o único desalinhado. Corrigir agora é barato.
- A importação direta de outros módulos em `service.go` é uma violação arquitetural que vai piorar.
- A auditoria em CI previne regressão — todo novo módulo será verificado.
- `events.go` não é "forçado": prompt *emite* eventos (`prompt.snapshot_created`, `toolset.snapshot_created`). O arquivo pode conter uma função `EventTypeForStatus` adaptada ou uma função `EventTypeForAction` que mapeie ações para tipos de evento.

---

## Fase 5: Estratégia de Testes

### Testes para a refatoração do módulo `prompt`

1. **Testes de compilação**: `go build ./internal/modules/prompt` — deve passar
2. **Testes existentes**: `go test ./internal/modules/prompt/...` — todos devem continuar passando
3. **Testes de arquitetura**: `go test ./tests/architecture/...` — verificar se a regra de importação entre módulos é respeitada
4. **Validação manual**: Verificar se `service.go` não importa mais módulos diretamente (exceto para interfaces DI)

### Testes para o script de auditoria

1. **Cenário: módulo completo** → deve passar
2. **Cenário: módulo faltando arquivo obrigatório** → deve falhar com código de erro não-zero
3. **Cenário: módulo com `service.go` > 300 linhas sem `service_*.go`** → deve emitir warning
4. **Integração CI**: O script deve ser chamado no `.github/workflows/ci.yml`

---

## Fase 6: Estratégia de Debug

Não aplicável. Esta é uma tarefa puramente de backend/estática. O debug será feito via:
- `go vet ./internal/modules/prompt`
- `go test ./internal/modules/prompt/...`
- `go test ./tests/architecture/...`

---

## Fase 7: Resumo para Validação com Usuário

### Resumo Executivo

O módulo `prompt` é o único entre 10 módulos que não segue o padrão de 10 arquivos obrigatórios da ADR-0022. Além dos arquivos faltando (`events.go`, `validation.go`), o módulo tem importações diretas de outros módulos que violam a política de isolamento. O plano propõe:

1. Criar `events.go` e `validation.go` no `prompt`
2. Refatorar `service.go` para usar interfaces DI (eliminando imports diretos)
3. Consolidar `types.go` em `models.go` e `repository_snapshot.go` em `repository.go`
4. Criar script de auditoria `scripts/verify-module-structure.sh` para CI
5. Atualizar `README.md` e `CONTRACTS.md` do prompt

**Estimativa de tempo**: 2-3 horas
**Riscos**: Regressão no serviço de prompt; mitigado por testes existentes
**Próximos passos**: Implementar as mudanças e validar com `go test`
