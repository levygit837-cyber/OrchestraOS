# CHECKLIST — ORCH-F05-R01-A02: Review Service + Validation Gate

**Agente:** Agente 2 (Review Service)  
**Ferramenta:** Kimi Code Extension  
**Início:** _pendente_  
**Status:** in_progress

---

## Checklist de Execução

- [ ] 1. Ler documentação obrigatória (README, AGENTS.md, prompt/CONTRACTS, taskgraph/CONTRACTS, roadmap Fase 5/9)
- [ ] 2. Analisar código existente: prompt/catalog/, domain/types.go, contracts/schemas/
- [ ] 3. Adicionar domain types em `internal/domain/types.go` (Review, ReviewStatus, ValidationGate)
- [ ] 4. Criar migration `migrations/013_reviews.sql` (tabela reviews, constraints, índices)
- [ ] 5. Criar `internal/modules/review/repository.go` + `queries.go` (CRUD, ListByTask, ListPending)
- [ ] 6. Criar `internal/modules/review/service.go` (Create, Start, SubmitVerdict, ListPending)
- [ ] 7. Criar `internal/modules/review/validation.go` + `events.go` + `models.go`
- [ ] 8. Criar `contracts/schemas/review.schema.json`
- [ ] 9. Adicionar perfil `reviewer` ao catálogo de prompts (fragmentos .md + manifest.json)
- [ ] 10. Atualizar toolset para incluir perfil `reviewer` (tools safe/guarded apenas)
- [ ] 11. Criar testes unitários `internal/modules/review/service_test.go` (Create, SubmitVerdict, imutabilidade)
- [ ] 12. Rodar `go test ./...` — verificar regressão em todos os pacotes
- [ ] 13. Code review auto-crítico (veredicto imutável, eventos, schema sync)
- [ ] 14. Correções pós-review
- [ ] 15. Validar build: `go build ./...` sem erros + entrega final ao usuário

## Notas de Progresso
<!-- Adicione notas curtas a cada iteração significativa -->
