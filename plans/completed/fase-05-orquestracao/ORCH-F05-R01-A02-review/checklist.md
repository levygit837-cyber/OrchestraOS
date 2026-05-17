# CHECKLIST — ORCH-F05-R01-A02: Review Service + Validation Gate

**Agente:** Agente 2 (Review Service)  
**Ferramenta:** Kimi Code Extension  
**Início:** 2026-05-12  
**Status:** completed

---

## Checklist de Execução

- [x] 1. Ler documentação obrigatória (README, AGENTS.md, prompt/CONTRACTS, taskgraph/CONTRACTS, roadmap Fase 5/9)
- [x] 2. Analisar código existente no escopo
- [x] 3. Adicionar domain types em `internal/domain/types.go` (Review, ReviewStatus, ValidationGate)
- [x] 4. Criar migration `migrations/013_reviews.sql` (tabela reviews, constraints, índices)
- [x] 5. Criar `internal/modules/review/repository.go` + `queries.go` (CRUD, ListByTask, ListPending)
- [x] 6. Criar `internal/modules/review/service.go` (Create, Start, SubmitVerdict, ListPending)
- [x] 7. Criar `internal/modules/review/validation.go` + `events.go` + `models.go`
- [x] 8. Criar `contracts/schemas/review.schema.json`
- [x] 9. Adicionar perfil `reviewer` ao catálogo de prompts (fragmentos .md + manifest.json)
- [x] 10. Atualizar toolset para incluir perfil `reviewer` (tools safe/guarded apenas)
- [x] 11. Criar testes de integração `tests/integration/review_service_test.go`
- [x] 12. Rodar `go test ./...` — verificar regressão em todos os pacotes
- [x] 13. Code review auto-crítico (veredicto imutável, eventos, schema sync)
- [x] 14. Correções pós-review
- [x] 15. Validar build: `go build ./...` sem erros + entrega final ao usuário

## Notas de Progresso

- Código do ReviewService entregue e funcional. Service, repository, models, validation, events, queries, contract, doc, README, CONTRACTS.md implementados.
- Migrations 013_reviews.sql e 014_reviews_unique_gate.sql criadas.
- Schema JSON review.schema.json criado em contracts/schemas/domain/.
- Perfil `reviewer` já existia no catálogo de prompts (manifest.json + fragments).
- Testes de integração em `tests/integration/review_service_test.go` cobrem Create, Get, Start, SubmitVerdict.
- Todos os testes do projeto passam (`go test ./...` ok).
- Checklist atualizado para refletir conclusão real. R01 pronta para merge.
