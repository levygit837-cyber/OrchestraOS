# Resumo do Plano: Padronização de Arquivos por Módulos

## Contexto
O projeto OrchestraOS possui 10 módulos em `internal/modules/`. A ADR-0022 define 10 arquivos obrigatórios por módulo. Atualmente, **9 de 10 módulos estão completos**; apenas `prompt/` está incompleto.

## Estado Atual
| Módulo | Arquivos Obrigatórios | Status |
|--------|----------------------|--------|
| agent, agentsession, orchestrator, review, run, task, taskgraph, trigger, workunit | 10/10 | ✅ Completo |
| prompt | 8/10 | ❌ Faltam `events.go` e `validation.go` |

## Problemas Adicionais no `prompt`
- `service.go` importa diretamente `run`, `task`, `workunit`, `agentsession` (viola ADR-0022)
- Arquivos `types.go` e `repository_snapshot.go` fogem do padrão (reconhecidos como legacy)

## Alternativa Escolhida
**Refatorar prompt + Auditoria em CI**

### Ações
1. Criar `events.go` — mapear ações do prompt para tipos de evento (snapshot_created, etc.)
2. Criar `validation.go` — extrair validações de `service.go`/`composer.go`
3. Refatorar `service.go` — substituir imports diretos por interfaces DI
4. Consolidar `types.go` → `models.go` e `repository_snapshot.go` → `repository.go`
5. Atualizar `README.md` e `CONTRACTS.md`
6. Criar `scripts/verify-module-structure.sh` + integrar no CI

## Estimativa
- **Tempo**: 2-3 horas
- **Risco**: Regressão no prompt service → mitigado por testes existentes

## Próximos Passos
1. Aprovar este plano
2. Executar implementação via skill `execute` ou `track-implementation`
3. Validar com `go test ./...`
