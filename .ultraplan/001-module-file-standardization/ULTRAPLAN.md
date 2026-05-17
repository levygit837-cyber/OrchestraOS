# ULTRAPLAN: Padronização de Arquivos por Módulos

**ID:** ULTRA-001  
**Tarefa:** Padronizar arquivos obrigatórios em todos os módulos de `internal/modules/*`  
**Tipo:** Por Domínio (cada módulo é uma fronteira independente)  
**Status:** Aprovado  
**Criado em:** 2026-05-17  
**Autor:** Kimi Code CLI (via skill ultraplan)

---

## 1. Resumo Executivo

O projeto OrchestraOS adota uma arquitetura de **Módulos Verticais** (ADR-0022), onde cada entidade de domínio reside em um módulo autônomo em `internal/modules/<entidade>/`. A ADR-0022 define **10 arquivos obrigatórios** por módulo para garantir consistência, facilitar a manutenção por agentes de IA e permitir validações automatizadas.

Atualmente, **9 de 10 módulos** estão em conformidade. O módulo **`prompt/`** está incompleto (faltam `events.go` e `validation.go`) e possui desvios arquiteturais adicionais (importações diretas de outros módulos, arquivos legacy fora do padrão).

Este plano detalha a correção do módulo `prompt` e a criação de um mecanismo de auditoria para prevenir regressões futuras.

---

## 2. Contexto

### 2.1 Por que existe este padrão?

A arquitetura de Módulos Verticais foi adotada para otimizar o projeto para agentes de IA (LLMs). Quando um agente precisa modificar uma funcionalidade, ele deve conseguir ler **apenas uma pasta** e ter todo o contexto necessário. A padronização de arquivos garante que:

- O agente saiba exatamente onde encontrar cada tipo de informação
- Ferramentas automatizadas possam verificar a saúde estrutural do projeto
- Novos módulos sejam criados com estrutura previsível

### 2.2 O que a ADR-0022 exige?

**Arquivos Obrigatórios (MUST):**

| Arquivo | Responsabilidade |
|---------|------------------|
| `doc.go` | Package documentation e context briefing para LLMs |
| `contract.go` | ModuleContract + regras hierárquicas (global, tipo, específica) |
| `README.md` | Propósito, file map, allowed dependencies |
| `CONTRACTS.md` | Invariantes, state machine, boundary rules |
| `models.go` | Tipos próprios (structs, enums, constants) |
| `queries.go` | SQL constants (mesmo que placeholder vazio) |
| `repository.go` | CRUD puro, zero business logic |
| `service.go` | Lógica principal, transações, eventos |
| `events.go` | Mapeamento de status/ações para tipos de evento |
| `validation.go` | Validações sintáticas de input |

**Exceção:** `orchestrator/` é módulo de coordenação — `repository.go` pode ser placeholder, mas deve existir.

### 2.3 Estado Atual Detalhado

| Módulo | doc.go | contract.go | README.md | CONTRACTS.md | models.go | events.go | queries.go | repository.go | service.go | validation.go | Tipo |
|--------|--------|-------------|-----------|--------------|-----------|-----------|------------|---------------|------------|---------------|------|
| `agent` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Domínio |
| `agentsession` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Domínio |
| `orchestrator` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Coordenação |
| **`prompt`** | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | Domínio |
| `review` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Domínio |
| `run` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Domínio |
| `task` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Domínio |
| `taskgraph` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Domínio |
| `trigger` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Domínio |
| `workunit` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | Domínio |

### 2.4 Problemas no Módulo `prompt`

1. **`events.go` ausente**: O módulo emite eventos (`prompt.snapshot_created`, `toolset.snapshot_created`) mas não possui o arquivo padronizado.
2. **`validation.go` ausente**: Validações estão espalhadas em `service.go` e `composer.go`, violando a regra de centralização.
3. **Importações diretas em `service.go`**: Importa `run`, `task`, `workunit`, `agentsession` para usar structs diretamente, violando ADR-0022 Seção 4.2 (Pilar 2).
4. **Arquivos legacy fora do padrão**: `types.go` e `repository_snapshot.go` existem mas são reconhecidos como legacy no próprio README.

---

## 3. Decisão

Adotar a **Alternativa B + C**: Refatorar o módulo `prompt` para alinhamento total com ADR-0022, e criar um script de auditoria estrutural para CI.

### 3.1 Justificativa

- O custo de correção agora é baixo (apenas 1 módulo desalinhado).
- A importação direta de outros módulos é uma dívida técnica que piorará com o tempo.
- A auditoria em CI previne que novos módulos ou modificações futuras erodam o padrão.
- `events.go` não é "forçado": o módulo prompt *efetivamente emite* eventos. O arquivo conterá a função de mapeamento desses eventos.

---

## 4. Alternativas Consideradas

### Alternativa A: Criar arquivos stub mínimos
- Criar `events.go` e `validation.go` vazios ou mínimos no prompt.
- **Eliminada**: Não resolve o problema de importações diretas nem consolida arquivos legacy. Seria uma correção superficial.

### Alternativa B: Refatorar prompt completamente + Auditoria CI ⭐ Escolhida
- Corrigir todos os desvios estruturais e arquiteturais do `prompt`, mais auditoria.
- **Prós**: Resolve problema raiz, alinha com ADR-0022, previne regressões.
- **Contras**: Maior esforço inicial (2-3h), risco de regressão no prompt service.

### Alternativa C: Script de auditoria apenas
- Criar verificação sem corrigir o prompt.
- **Eliminada**: Não resolve o estado atual; apenas previne futuros problemas.

### Alternativa D: Exemptar módulos sem state machine
- Argumentar que `prompt` não precisa de `events.go` porque não tem lifecycle states.
- **Eliminada**: Quebra a uniformidade do padrão. O propósito da padronização é justamente a previsibilidade. Criar exceções gera precedente perigoso.

---

## 5. Estratégia de Implementação

### Fase 1: Preparação e Validação
1. Executar testes existentes: `go test ./internal/modules/prompt/...` (baseline)
2. Executar architecture tests: `go test ./tests/architecture/...` (baseline)

### Fase 2: Criar Arquivos Obrigatórios Faltantes

#### 5.2.1 `internal/modules/prompt/events.go`
```go
package prompt

// EventTypeForAction retorna o tipo de evento para uma ação do prompt.
// O módulo prompt não possui state machine tradicional, mas emite eventos
// de snapshot lifecycle.
func EventTypeForAction(action string) string {
    switch action {
    case "snapshot_created":
        return "prompt.snapshot_created"
    case "toolset_snapshot_created":
        return "toolset.snapshot_created"
    default:
        return "prompt." + action
    }
}
```

#### 5.2.2 `internal/modules/prompt/validation.go`
- Extrair validações de `service.go` e `composer.go`
- Implementar funções como `ValidateFragment()`, `ValidateSnapshotInput()`, `ValidateToolsetInput()`
- Usar `core/validation` como base

### Fase 3: Refatorar `service.go`

#### 5.3.1 Eliminar imports diretos de outros módulos
Extrair interfaces DI no topo do arquivo:

```go
// Interfaces DI — tipos importados apenas como retorno
type RunReader interface {
    GetByID(id string) (*run.Run, error)
}

type TaskReader interface {
    GetByID(id string) (*task.Task, error)
}

type WorkUnitReader interface {
    GetByID(id string) (*workunit.WorkUnit, error)
}

type AgentSessionReader interface {
    GetByID(id string) (*agentsession.AgentSession, error)
}
```

Modificar `PrepareAndPersistInput` para usar interfaces:
```go
type PrepareAndPersistInput struct {
    Run                    RunReader  // interface, não struct
    WorkUnit               WorkUnitReader
    Task                   TaskReader
    Session                AgentSessionReader
    // ... campos restantes
}
```

**Nota**: O wiring real das implementações permanece em `internal/bootstrap/services.go`.

### Fase 4: Consolidar Arquivos Legacy

| Ação | Origem | Destino |
|------|--------|---------|
| Mover tipos auxiliares | `types.go` | `models.go` |
| Mover CRUD de snapshots | `repository_snapshot.go` | `repository.go` |
| Remover arquivos vazios | `types.go`, `repository_snapshot.go` | — |

### Fase 5: Atualizar Documentação

#### 5.5.1 `README.md`
- Atualizar File Map para refletir a estrutura final
- Remover menções a "legacy"
- Listar `service_*.go` se houver decomposição

#### 5.5.2 `CONTRACTS.md`
- Adicionar seção "File Decomposition" se `service.go` for decomposto
- Documentar interfaces DI
- Atualizar invariantes

### Fase 6: Criar Script de Auditoria

#### 5.6.1 `scripts/verify-module-structure.sh`
```bash
#!/usr/bin/env bash
set -euo pipefail

# Verifica se todos os módulos em internal/modules/* possuem
# os 10 arquivos obrigatórios da ADR-0022.

MANDATORY=(doc.go contract.go README.md CONTRACTS.md models.go events.go queries.go repository.go service.go validation.go)
MODULES_DIR="internal/modules"
EXIT_CODE=0

for module in "$MODULES_DIR"/*/; do
    module_name=$(basename "$module")
    for file in "${MANDATORY[@]}"; do
        if [ ! -f "$module/$file" ]; then
            echo "ERROR: $module_name missing mandatory file: $file"
            EXIT_CODE=1
        fi
    done
done

exit $EXIT_CODE
```

#### 5.6.2 Integrar no CI
Adicionar passo no `.github/workflows/ci.yml`:
```yaml
- name: Verify Module Structure
  run: ./scripts/verify-module-structure.sh
```

---

## 6. Estratégia de Testes

### 6.1 Testes de Compilação
```bash
go build ./internal/modules/prompt
go vet ./internal/modules/prompt/...
```

### 6.2 Testes Unitários Existentes
```bash
go test ./internal/modules/prompt/...
```
**Critério de aceite**: Todos os testes existentes devem continuar passando.

### 6.3 Testes de Arquitetura
```bash
go test ./tests/architecture/...
```
**Critério de aceite**: Verificar que `prompt/` não importa mais outros módulos diretamente (exceto para interfaces DI).

### 6.4 Testes do Script de Auditoria
1. **Cenário positivo**: Rodar em módulo completo → exit code 0
2. **Cenário negativo**: Simular arquivo faltando → exit code 1
3. **CI**: O workflow deve falhar se o script retornar erro

---

## 7. Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação |
|-------|--------------|---------|-----------|
| Regressão no `PromptService` | Média | Alto | Executar testes existentes antes e depois; fazer mudanças incrementais |
| Interfaces DI quebram bootstrap | Média | Alto | Atualizar `internal/bootstrap/services.go` junto; compilar o projeto inteiro |
| `events.go` sem utilidade clara | Baixa | Baixa | Documentar que prompt emite eventos de snapshot; função é mapeamento, não state machine |
| Resistência a consolidar arquivos legacy | Baixa | Média | Fazer backup via git; se algo quebrar, reverter a consolidação e manter arquivos separados temporariamente |

---

## 8. Checklist de Execução

- [ ] Executar `go test ./internal/modules/prompt/...` (baseline)
- [ ] Executar `go test ./tests/architecture/...` (baseline)
- [ ] Criar `internal/modules/prompt/events.go`
- [ ] Criar `internal/modules/prompt/validation.go`
- [ ] Refatorar `service.go` — extrair interfaces DI
- [ ] Atualizar `internal/bootstrap/services.go` (wiring das interfaces)
- [ ] Consolidar `types.go` → `models.go`
- [ ] Consolidar `repository_snapshot.go` → `repository.go`
- [ ] Remover arquivos consolidados
- [ ] Atualizar `prompt/README.md`
- [ ] Atualizar `prompt/CONTRACTS.md`
- [ ] Criar `scripts/verify-module-structure.sh`
- [ ] Tornar script executável (`chmod +x`)
- [ ] Integrar script no `.github/workflows/ci.yml`
- [ ] Executar `go test ./internal/modules/prompt/...` (validação)
- [ ] Executar `go test ./tests/architecture/...` (validação)
- [ ] Executar `./scripts/verify-module-structure.sh` (validação)
- [ ] Executar `./scripts/safe-commit.sh`

---

## 9. Próximos Passos

1. **Aprovar este ULTRAPLAN** ✅
2. **Executar implementação** usando a skill `execute` ou `track-implementation`
3. **Abrir PR** com as mudanças
4. **Aguardar CI** passar antes de mergear

---

## Apêndice A: Referências

- **ADR-0022**: `docs/adr/0022-vertical-module-architecture.md` — Arquitetura de Módulos Verticais, padronização, política de importação
- **Template de Módulo**: `docs/templates/module/` — 10 arquivos template para novos módulos
- **Script de Novo Módulo**: `scripts/new-module.sh` — Cria módulos seguindo ADR-0022
- **Bootstrap**: `internal/bootstrap/services.go` — Wiring de dependências

## Apêndice B: Histórico de Mudanças

| Data | Autor | Mudança |
|------|-------|---------|
| 2026-05-17 | Kimi Code CLI | Criação inicial do ULTRAPLAN |
