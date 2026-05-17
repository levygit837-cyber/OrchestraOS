# CHECKLIST — ORCH-F28-R01-A01: Renomeações Mecânicas

**Agente:** A01 (Renames)  
**Início:** 2026-05-17  
**Status:** in_progress

---

## Checklist de Execução

- [x] 1. Ler documentação obrigatória (ADR-0028 seção 4 Fase 1, AGENTS.md)
- [x] 2. Renomear `internal/core/db/txkit.go` → `internal/core/db/transactions.go`
- [x] 3. Renomear `internal/core/serialization/serialization.go` → `internal/core/serialization/marshal.go`
- [x] 4. Renomear `internal/core/validation/validation.go` → `internal/core/validation/validators.go`
- [x] 5. Analisar `internal/core/transition/helpers.go` e dividir em `payload.go` + `audit.go`
- [x] 6. Renomear `internal/core/transition/eventops.go` → `internal/core/transition/append.go`
- [x] 7. Ajustar todos os imports/referências em outros pacotes (nenhuma alteração necessária — imports são por pacote, não por arquivo)
- [x] 8. Rodar `go build ./...` — passou
- [x] 9. Rodar `go test ./...` — passou
- [x] 10. Rodar `./scripts/verify-contracts.sh` — passou
- [x] 11. Rodar `./scripts/lint.sh` — passou
- [x] 12. Code review auto-critico (responder perguntas do plan.md)
  - [x] Nenhum arquivo `helpers.go`, `txkit.go`, `eventops.go`, ou `serialization.go` ainda existe
  - [x] Todos os imports em outros pacotes apontam para os novos nomes (imports são por pacote, não por arquivo; nenhum quebrou)
  - [x] `go build ./...` passa sem erros
  - [x] `go test ./...` passa sem falhas
  - [x] Nenhuma lógica foi alterada (apenas renomeações mecânicas)
- [ ] 13. Commit via `./scripts/safe-commit.sh`
- [ ] 14. Atualizar este checklist como completo
- [ ] 15. Entrega final ao usuário

## Notas de Progresso
<!-- O agente adiciona notas curtas aqui -->
