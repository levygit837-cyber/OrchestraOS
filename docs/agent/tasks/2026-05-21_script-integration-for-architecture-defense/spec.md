---
tipo: spec
task-id: 2026-05-21_script-integration-for-architecture-defense
domain: transversal
status: em-andamento
---

# Spec: Script Integration for Architecture Defense

## Alterações por Script

### verify-module-structure.sh

**Mudança principal:** Reduzir arquivos obrigatórios de 10 para 5.

**Antes:**
```bash
MANDATORY_FILES=(
    doc.go
    contract.go
    README.md
    CONTRACTS.md
    models.go
    events.go
    queries.go
    repository.go
    service.go
    validation.go
)
```

**Depois:**
```bash
MANDATORY_FILES=(
    doc.go
    README.md
    models.go
    repository.go
    service.go
)
```

**Notas:**
- `contract.go` e `CONTRACTS.md` são mantidos se existirem, mas não são mais obrigatórios.
- `events.go`, `queries.go`, `validation.go` são opcionais.

### pre-commit.sh

**Ordem de execução:**
1. `go vet ./...`
2. `go test ./tests/architecture/... -count=1`
3. `./scripts/go/verify-module-structure.sh`
4. `./scripts/go/verify-contracts.sh`

### lint.sh

**Ordem de execução:**
1. `go vet ./...`
2. `go test ./tests/architecture/... -count=1`
3. `./scripts/go/verify-module-structure.sh`
4. `./scripts/go/verify-contracts.sh`
5. `golangci-lint run ./...` (se disponível)

### verify-contracts.sh

**Atualização de comentário:**
```bash
# verify-contracts.sh
# Runs the full architecture test suite including:
# - module boundaries (zero cross-module imports)
# - repository purity (CRUD only)
# - domain import integrity (entity types in domain/)
# - code anomalies (ignored errors, panic, fmt.Println)
```

### AGENTS.md

**Adições:**
- Seção "Validações Obrigatórias" instruindo a rodar `./scripts/go/lint.sh` e `./scripts/go/verify-contracts.sh`
- Nota: "NUNCA confie apenas em `go test ./...` — sempre rode os scripts de validação arquitetural."

## Critérios de Aceitação
- [ ] `verify-module-structure.sh` exige apenas 5 arquivos
- [ ] `pre-commit.sh` chama `verify-module-structure.sh`
- [ ] `lint.sh` chama `verify-module-structure.sh`
- [ ] `AGENTS.md` atualizado
- [ ] Scripts testados manualmente
