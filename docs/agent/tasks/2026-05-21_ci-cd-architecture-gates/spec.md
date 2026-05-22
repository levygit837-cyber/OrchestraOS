---
tipo: spec
task-id: 2026-05-21_ci-cd-architecture-gates
domain: transversal
status: em-andamento
---

# Spec: CI/CD Architecture Gates

## Jobs para ci.yml

### module-boundaries
```yaml
module-boundaries:
  name: Module Boundaries
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version-file: go.mod
        cache: true
    - run: |
        echo "FAILURE REASON: Cross-module imports detected."
        echo "Modules must NOT import other modules (ADR-0030)."
        go test ./tests/architecture/... -run TestModuleBoundaries -v -count=1
```

### repository-purity
```yaml
repository-purity:
  name: Repository Purity
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version-file: go.mod
        cache: true
    - run: |
        echo "FAILURE REASON: Business logic detected in repository.go."
        echo "repository.go must contain only CRUD."
        go test ./tests/architecture/... -run TestRepositoryPurity -v -count=1
```

### domain-integrity
```yaml
domain-integrity:
  name: Domain Integrity
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version-file: go.mod
        cache: true
    - run: |
        echo "FAILURE REASON: Shared entity types not in internal/domain/."
        echo "All shared entity types must live in internal/domain/ (ADR-0030)."
        go test ./tests/architecture/... -run TestDomainImportIntegrity -v -count=1
```

### ignored-errors
```yaml
ignored-errors:
  name: Ignored Errors
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version-file: go.mod
        cache: true
    - run: |
        echo "FAILURE REASON: Code anomalies detected."
        echo "Check for _ = ..., panic(), fmt.Println, inline SQL."
        go test ./tests/architecture/... -run TestCodeAnomalies -v -count=1
```

### bootstrap-di-check
```yaml
bootstrap-di-check:
  name: Bootstrap DI Check
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - run: |
        echo "FAILURE REASON: cmd/ instantiates repositories/services directly."
        echo "cmd/ must use bootstrap/ for DI wiring."
        if grep -rn "NewRepository\|NewService" cmd/ --include="*.go"; then
          exit 1
        fi
```

## Jobs para pr-gate.yml

Mesmos jobs com prefixo `PR Gate /`.

## Critérios de Aceitação
- [ ] `ci.yml` contém os 5 novos jobs
- [ ] `pr-gate.yml` contém os 5 novos jobs
- [ ] Cada job tem mensagem de falha descritiva
- [ ] YAML válido
