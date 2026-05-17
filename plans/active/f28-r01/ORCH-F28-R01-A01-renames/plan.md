# ORCH-F28-R01-A01: Fase 1 — Renomeações Mecânicas

## Contexto do Projeto

**Projeto:** OrchestraOS (Go monolith modular)  
**Arquitetura:** Vertical Slice (ADR-0022) + Hybrid Intelligent Orchestrator (ADR-0023)  
**Meta:** Eliminar nomes genéricos de arquivos em `internal/core/*` conforme ADR-0028  
**Risco:** **BAIXO** — zero lógica alterada, apenas renomear arquivos e ajustar imports.

---

## Documentação Obrigatória (LEIA ANTES DE COMEÇAR)

1. `docs/adr/0028-core-architecture-and-naming-standards.md` — Seção 4 "Plano de Migração", Fase 1
2. `AGENTS.md` — Regras de commits (`./scripts/safe-commit.sh`)
3. Este arquivo (`plan.md`)

---

## O que Já Existe vs O que Deve Ser Feito

### Arquivos a renomear

| Arquivo Atual | Novo Arquivo | Motivo |
|---|---|---|
| `internal/core/db/txkit.go` | `internal/core/db/transactions.go` | `kit.go` é nome proibido |
| `internal/core/serialization/serialization.go` | `internal/core/serialization/marshal.go` | Nome repetido do pacote não comunica função |
| `internal/core/validation/validation.go` | `internal/core/validation/validators.go` | Nome repetido do pacote; `validators` comunica conteúdo |
| `internal/core/transition/helpers.go` | `internal/core/transition/payload.go` + `internal/core/transition/audit.go` | `helpers` é nome proibido |
| `internal/core/transition/eventops.go` | `internal/core/transition/append.go` | `ops` é abreviação vaga |

**Regra:** Conteúdo dos arquivos NÃO muda. Apenas renomear e ajustar imports em quem os consome.

### Divisão de `transition/helpers.go`

Analise o conteúdo de `transition/helpers.go` e divida em:
- **`payload.go`**: Funções relacionadas a construção de payload (`TransitionPayload`, `TransitionContext`, builders de payload)
- **`audit.go`**: Funções relacionadas a auditoria/regras de estados finais

Se não houver clara separação entre payload e audit, coloque tudo em `payload.go` e deixe `audit.go` vazio (será removido no cleanup).

---

## Fronteiras de Isolamento

**TOCAR:**
- `internal/core/db/txkit.go` → `internal/core/db/transactions.go`
- `internal/core/serialization/serialization.go` → `internal/core/serialization/marshal.go`
- `internal/core/validation/validation.go` → `internal/core/validation/validators.go`
- `internal/core/transition/helpers.go` → `transition/payload.go` + `transition/audit.go`
- `internal/core/transition/eventops.go` → `internal/core/transition/append.go`
- Todos os arquivos que importam os símbolos acima (ajustar import paths se necessário)

**EVITAR:**
- `internal/modules/*` — não toque em nenhum módulo
- `internal/core/coordination/` — não toque
- `internal/bootstrap/` — não toque exceto se houver import direto dos arquivos renomeados
- Qualquer lógica de negócio — esta tarefa é 100% mecânica

---

## Interfaces Contratuais

Nenhuma. Esta tarefa é puramente mecânica.

---

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

Você deve executar esta tarefa em ciclos curtos usando o arquivo de checklist persistente.

**Caminho do checklist:** `plans/active/f28-r01/ORCH-F28-R01-A01-renames/checklist.md`

**A cada iteração:**
1. **LER** o checklist para identificar o próximo item pendente
2. **EXECUTAR** o item (renomear arquivo, ajustar imports)
3. **VALIDAR** o item (`go build ./...` deve passar)
4. **ATUALIZAR** o checklist marcando o item como concluído
5. **CONTINUAR** para o próximo item

**Regras do Ralph Loop:**
- Nunca pule um item sem marcá-lo no checklist
- Se encontrar bloqueio, adicione uma nota na seção "Notas de Progresso"
- Ao final de cada ciclo significativo, faça um commit pequeno via `./scripts/safe-commit.sh`
- O checklist é sua fonte de verdade de progresso

---

## Regras de Implementação

1. **NUNCA modifique lógica.** Apenas renomear arquivos e ajustar imports/referências.
2. Use `git mv` quando possível para preservar histórico de git.
3. Se houver `package` declaration ou comentários que referenciam o nome antigo do arquivo, atualize-os.
4. Após cada renomeação, rode `go build ./...` para garantir que não quebrou nada.
5. Se `helpers.go` não tiver separação clara entre payload e audit, prefira: coloque tudo em `payload.go` e remova `helpers.go`. Não crie `audit.go` vazio.

---

## Regras Rígidas de Testes

1. Rode `go test ./...` antes e depois de cada renomeação.
2. Rode `./scripts/verify-contracts.sh` ao final.
3. Rode `./scripts/lint.sh` ao final.
4. Todos os testes devem passar. Se algum teste falhar, é porque você quebrou um import — corrija.

---

## Code Review Auto-Crítico Obrigatório

Antes de entregar, responda estas perguntas:

- [ ] Algum arquivo `helpers.go`, `txkit.go`, `eventops.go`, ou `serialization.go` ainda existe?
- [ ] Todos os imports em outros pacotes apontam para os novos nomes?
- [ ] `go build ./...` passa sem erros?
- [ ] `go test ./...` passa sem falhas?
- [ ] Nenhuma lógica foi alterada (apenas renomeações)?

---

## Critérios de Aceite Verificáveis

1. `internal/core/db/txkit.go` NÃO existe mais; `internal/core/db/transactions.go` existe
2. `internal/core/serialization/serialization.go` NÃO existe mais; `internal/core/serialization/marshal.go` existe
3. `internal/core/validation/validation.go` NÃO existe mais; `internal/core/validation/validators.go` existe
4. `internal/core/transition/helpers.go` NÃO existe mais; `internal/core/transition/payload.go` existe (e opcionalmente `audit.go`)
5. `internal/core/transition/eventops.go` NÃO existe mais; `internal/core/transition/append.go` existe
6. `go build ./...` passa
7. `go test ./...` passa
8. `./scripts/verify-contracts.sh` passa
9. `./scripts/lint.sh` passa

---

## Entrega Final

1. Commit com `./scripts/safe-commit.sh "refactor(core): rename generic files per ADR-0028 Fase 1"`
2. Push da feature branch
3. Reportar ao usuário: "Fase 1 completa. Todos os arquivos renomeados. Build verde."
