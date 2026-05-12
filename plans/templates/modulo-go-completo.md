# Template: Módulo Go Completo (OrchestraOS)

Use este template como base para criar planos de implementação de novos módulos no OrchestraOS.

## Contexto

- **Projeto:** OrchestraOS
- **Linguagem:** Go 1.24
- **Banco:** PostgreSQL (lib/pq)
- **Migrations:** Goose v3
- **Arquitetura:** Event Sourcing + State Machine + Módulos Verticais

## Estrutura do Módulo

Todo módulo deve seguir o padrão:

```
internal/modules/{nome}/
├── contract.go          # Regras críticas (invariants, boundaries)
├── CONTRACTS.md         # Documentação completa de contratos
├── doc.go               # Propósito e dependências
├── events.go            # Mapeamento de eventos
├── models.go            # Aliases de tipos / structs locais
├── queries.go           # SQL constants
├── README.md            # Propósito, file map, dependencies
├── repository.go        # CRUD puro
├── service.go           # Regras de negócio
└── validation.go        # Validações de input
```

## Checklist de Implementação

- [ ] Criar migration SQL (`migrations/{NNN}_{nome}.sql`)
- [ ] Adicionar domain types em `internal/domain/types.go`
- [ ] Criar schema JSON em `contracts/schemas/{nome}.schema.json`
- [ ] Criar `contract.go` com regras críticas
- [ ] Criar `CONTRACTS.md` com invariants, state machine, boundaries
- [ ] Criar `doc.go`
- [ ] Criar `events.go` com mapeamento de event types
- [ ] Criar `models.go`
- [ ] Criar `queries.go` com SQL constants
- [ ] Criar `README.md`
- [ ] Criar `repository.go` (CRUD puro)
- [ ] Criar `service.go` (regras de negócio + emissão de eventos)
- [ ] Criar `validation.go`
- [ ] Criar testes unitários `*_test.go`
- [ ] Atualizar state machine se houver novos status (`internal/core/statemachine/`)
- [ ] Rodar `go test ./...` (regressão)
- [ ] Rodar `go build ./...`
- [ ] Code review auto-crítico

## Regras de Implementação

1. Siga rigorosamente o padrão dos módulos existentes (task, run, workunit, agentsession)
2. Use `uuid.New().String()` para novos IDs
3. Use `time.Now().UTC()` para timestamps
4. Valide entradas nas bordas usando `internal/core/validation/`
5. Trate erros com `apperrors.Wrap()` ou `apperrors.New()`
6. Nunca silencie erros
7. Não adicione dependências externas sem justificativa
8. Não quebre testes existentes

## Testes

- Testes REAIS, não mockados de forma que escondam bugs
- Determinísticos, flexíveis, eficientes
- Cobertura mínima: caminho feliz + erros + validações de input
- Para integração: use FakeRuntime/stubs, não APIs externas
