# CHECKLIST — ORCH-F28-R01-A06: Cleanup Final

**Agente:** A06 (Cleanup)  
**Início:** 2026-05-17  
**Status:** in_progress

---

## Checklist de Execução

- [ ] 1. Verificar que A01-A05 entregaram (grep por imports de coordination)
- [ ] 2. Se houver imports restantes, reportar ao usuário e NÃO prosseguir
- [ ] 3. Remover diretório `internal/core/coordination/`
- [ ] 4. Verificar novamente: `grep -r "internal/core/coordination" --include="*.go" .` → ZERO resultados
- [ ] 5. Atualizar `tests/architecture/module_boundaries_test.go`
- [ ] 6. Atualizar `tests/architecture/transition_imports_test.go`
- [ ] 7. Atualizar `AGENTS.md` com regra de nomes proibidos
- [ ] 8. Rodar `go build ./...` — deve passar
- [ ] 9. Rodar `go test ./...` — deve passar
- [ ] 10. Rodar `./scripts/verify-contracts.sh` — deve passar
- [ ] 11. Rodar `./scripts/lint.sh` — deve passar
- [ ] 12. Code review auto-critico
- [ ] 13. Commit via `./scripts/safe-commit.sh`
- [ ] 14. Atualizar este checklist como completo
- [ ] 15. Entrega final ao usuário

## Notas de Progresso
<!-- O agente adiciona notas curtas aqui -->
