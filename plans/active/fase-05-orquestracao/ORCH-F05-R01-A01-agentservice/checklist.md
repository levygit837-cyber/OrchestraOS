# CHECKLIST — ORCH-F05-R01-A01: AgentService + Validação AgentID

**Agente:** Agente 1 (AgentService)  
**Ferramenta:** Windsurf  
**Início:** _pendente_  
**Status:** in_progress

---

## Checklist de Execução

- [ ] 1. Ler documentação obrigatória (README, AGENTS.md, agent/CONTRACTS, agentsession/CONTRACTS, ADR 0021)
- [ ] 2. Analisar código existente: agent/ (runtime, models, contract), agentsession/service.go, domain/types.go
- [ ] 3. Criar migration `migrations/012_agents.sql` (tabela, constraints, índices)
- [ ] 4. Criar `internal/modules/agent/repository.go` + `queries.go` (CRUD, FindByProfileAndRuntime)
- [ ] 5. Criar `internal/modules/agent/service.go` (Create, GetByID, FindOrCreate)
- [ ] 6. Criar `internal/modules/agent/validation.go` (profile, runtime_type, name)
- [ ] 7. Criar `internal/modules/agent/events.go` + atualizar `models.go`
- [ ] 8. Atualizar `internal/modules/agentsession/service.go` — validar AgentID via AgentReader interface
- [ ] 9. Criar testes unitários `internal/modules/agent/service_test.go` (Create, GetByID, FindOrCreate, validações)
- [ ] 10. Atualizar testes existentes de agentsession que usam AgentID arbitrário
- [ ] 11. Rodar `go test ./...` — verificar regressão em todos os pacotes
- [ ] 12. Code review auto-crítico (lógica, transações, SQL injection, erros silenciados)
- [ ] 13. Correções pós-review
- [ ] 14. Validar build: `go build ./...` sem erros
- [ ] 15. Atualizar este checklist como completo + entrega final ao usuário

## Notas de Progresso
<!-- Adicione notas curtas a cada iteração significativa -->
