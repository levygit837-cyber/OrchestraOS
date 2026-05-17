# CHECKLIST — ORCH-F28-R01-A01: Renomeações Mecânicas

**Agente:** A01 (Renames)  
**Início:** 2026-05-17  
**Status:** in_progress

---

## Checklist de Execução

- [ ] 1. Ler documentação obrigatória (ADR-0028 seção 4 Fase 1, AGENTS.md)
- [ ] 2. Renomear `internal/core/db/txkit.go` → `internal/core/db/transactions.go`
- [ ] 3. Renomear `internal/core/serialization/serialization.go` → `internal/core/serialization/marshal.go`
- [ ] 4. Renomear `internal/core/validation/validation.go` → `internal/core/validation/validators.go`
- [ ] 5. Analisar `internal/core/transition/helpers.go` e dividir em `payload.go` + `audit.go`
- [ ] 6. Renomear `internal/core/transition/eventops.go` → `internal/core/transition/append.go`
- [ ] 7. Ajustar todos os imports/referências em outros pacotes
- [ ] 8. Rodar `go build ./...` — deve passar
- [ ] 9. Rodar `go test ./...` — deve passar
- [ ] 10. Rodar `./scripts/verify-contracts.sh` — deve passar
- [ ] 11. Rodar `./scripts/lint.sh` — deve passar
- [ ] 12. Code review auto-critico (responder perguntas do plan.md)
- [ ] 13. Commit via `./scripts/safe-commit.sh`
- [ ] 14. Atualizar este checklist como completo
- [ ] 15. Entrega final ao usuário

## Notas de Progresso
<!-- O agente adiciona notas curtas aqui -->
