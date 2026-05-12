# CHECKLIST — ORCH-F05-R01-A03: Triggers Configuráveis (Camada 1)

**Agente:** Agente 3 (Triggers)  
**Ferramenta:** Kimi-CLI  
**Início:** _pendente_  
**Status:** in_progress

---

## Checklist de Execução

- [x] 1. Ler documentação obrigatória (README, AGENTS.md, run/CONTRACTS, agentsession/CONTRACTS, eventstore)
- [x] 2. Analisar código existente: eventstore/, run/service.go, agentsession/service.go, domain/types.go
- [x] 3. Adicionar domain types em `internal/domain/types.go` (Trigger, TriggerType, TriggerStatus, AnomalyType, ResolutionAction)
- [x] 4. Criar migration `migrations/014_triggers.sql` (tabela triggers, constraints, índices)
- [x] 5. Criar `internal/modules/trigger/repository.go` + `queries.go` (CRUD, ListActive, ListByRun)
- [x] 6. Criar `internal/modules/trigger/service.go` (Create, EvaluateRun, EvaluateSession, Resolve, Dismiss)
- [x] 7. Criar `internal/modules/trigger/detectors.go` (Stall, Loop, Drift, PathViolation, Token, Steps, Time)
- [x] 8. Criar `internal/modules/trigger/thresholds.go` + `validation.go` + `events.go` + `models.go`
- [x] 9. Criar `contracts/schemas/trigger.schema.json`
- [x] 10. Criar testes unitários para cada detector (casos positivos e negativos)
- [x] 11. Criar testes de integração `tests/integration/trigger_test.go` (EvaluateRun, ListActive)
- [x] 12. Rodar `go test ./...` — verificar regressão em todos os pacotes
- [x] 13. Code review auto-crítico (determinismo, falsos positivos, memory leaks)
- [x] 14. Correções pós-review
- [x] 15. Validar build: `go build ./...` sem erros + entrega final ao usuário

## Notas de Progresso
<!-- Adicione notas curtas a cada iteração significativa -->
