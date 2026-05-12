# 🎯 Tarefa: Implementar Review Service, Validation Gate e Perfil Reviewer no Catálogo de Prompts

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.  
> **Branch esperada:** `agent-a02/review`  
> **Worktree esperada:** `../orchestraos-a02`  
> Se não estiver isolado, execute: `cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A02 review`

## Contexto do Projeto
- **Nome:** OrchestraOS
- **Linguagem:** Go 1.24
- **Banco:** PostgreSQL (lib/pq)
- **Migrations:** Goose v3 (pressly/goose/v3)
- **Arquitetura:** Event Sourcing parcial + State Machine + Módulos Verticais
- **Padrões:** Cada módulo tem service.go + repository.go + queries.go + models.go + events.go + contract.go
- **Prompt System:** Catálogo embarcado em `internal/modules/prompt/catalog/` com manifest.json + arquivos .md por categoria.

## Documentação Obrigatória
ANTES de escrever código, leia:
1. `/home/levybonito/Documentos/OrchestraOS/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/AGENTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/implementation/roadmap.md` (seção Fase 5 e Fase 9)
4. `/home/levybonito/Documentos/OrchestraOS/internal/modules/prompt/README.md`
5. `/home/levybonito/Documentos/OrchestraOS/internal/modules/prompt/CONTRACTS.md`
6. `/home/levybonito/Documentos/OrchestraOS/internal/modules/taskgraph/CONTRACTS.md` (para entender TaskGraph)
7. `/home/levybonito/Documentos/OrchestraOS/contracts/schemas/planner-output.schema.json`

## O que Já Existe
- Sistema de prompts com catálogo em `internal/modules/prompt/catalog/`.
- Perfis existentes no planner: `code_worker`, `docs_writer`, `debugger`, `default`. **Não existe `reviewer`**.
- `TaskGraph` e `WorkUnit` já suportam `assigned_agent_profile`.
- Event Store, State Machine e todos os serviços de domínio base estão prontos.
- Não existe módulo `review/`.

## O que Você Deve Implementar

### Parte A: Módulo Review Completo
Crie `internal/modules/review/` como um módulo de domínio completo:

1. **Migration `migrations/013_reviews.sql`**
   - Tabela `reviews` com: id (UUID PK), run_id (UUID, nullable), work_unit_id (UUID, nullable), task_id (UUID, nullable), agent_session_id (UUID, nullable), reviewer_agent_id (UUID, nullable), gate_type (`hard`, `soft`, `policy`), status (`pending`, `in_progress`, `approved`, `changes_requested`, `needs_discussion`), verdict_reason (TEXT), evidence_refs (TEXT[]), criteria_checked (JSONB), created_at, updated_at, completed_at.
   - Índices em run_id, work_unit_id, task_id, status.

2. **`internal/modules/review/service.go`**
   - `ReviewService` struct com `*sql.DB`
   - `Create(ctx, input) -> (*transition.OperationResult[*domain.Review], error)` — cria review para uma work unit
   - `Start(ctx, reviewID, input) -> ...` — inicia review session
   - `SubmitVerdict(ctx, reviewID, verdict, reason, evidenceRefs) -> ...` — emite veredicto
   - `GetByID(ctx, id) -> (*domain.Review, error)`
   - `ListByTask(ctx, taskID) -> ([]*domain.Review, error)`
   - `ListPending(ctx) -> ([]*domain.Review, error)`
   - Eventos: `review.created`, `review.started`, `review.verdict_submitted`

3. **`internal/modules/review/repository.go`**
   - CRUD puro: Create, GetByID, UpdateStatus, ListByTask, ListPending

4. **`internal/modules/review/queries.go`**
   - SQL constants

5. **`internal/modules/review/models.go`**
   - Aliases para domain types ou structs locais se necessário

6. **`internal/modules/review/events.go`**
   - Mapeamento event types

7. **`internal/modules/review/contract.go`, `doc.go`, `README.md`, `CONTRACTS.md`**
   - Siga o padrão dos outros módulos
   - CONTRACTS.md deve documentar invariants: um review por gate, veredicto imutável após emitido

8. **`internal/modules/review/validation.go`**
   - Validação de verdict válido, gate_type válido

### Parte B: Adicionar Types ao Domain
Em `internal/domain/types.go`, adicione:
- `ReviewStatus` type + constantes: `pending`, `in_progress`, `approved`, `changes_requested`, `needs_discussion`
- `ReviewDecision` type + constantes (pode ser alias de ReviewStatus se apropriado)
- `ValidationGate` type + constantes: `hard`, `soft`, `policy`
- `Review` struct com todos os campos da tabela
- `ReviewCriteriaChecked` struct para o JSONB de critérios

### Parte C: Adicionar Schema JSON
Em `contracts/schemas/review.schema.json`:
- Schema JSON Schema 2020-12 para a entidade Review
- Siga o padrão dos schemas existentes (ex: `task.schema.json`)

### Parte D: Adicionar Perfil Reviewer ao Catálogo de Prompts
Em `internal/modules/prompt/catalog/`:
1. Crie arquivo `manifest.json` se não existir, ou adapte o existente.
2. Adicione fragmentos de prompt para o perfil `reviewer`:
   - `persona.reviewer.md` — persona do reviewer
   - `tool_policy.reviewer.md` — políticas de ferramentas do reviewer (leitura segura, nada destrutivo)
   - `output_contract.reviewer.md` — formato de saída esperado (veredicto estruturado)
3. Atualize `internal/modules/prompt/toolset.go` (ou o arquivo equivalente) para incluir o perfil `reviewer` no `SelectToolset`. Se não encontrar, procure onde os toolsets por perfil são definidos.
4. O perfil `reviewer` deve ter tools classificadas como `safe` ou `guarded`, nenhuma `destructive` ou `approval_required`.

## Fronteiras de Isolamento

### ✅ Você PODE e DEVE tocar:
- `migrations/013_reviews.sql`
- `internal/modules/review/*` (diretório completo novo)
- `internal/domain/types.go` (adicionar Review, ReviewStatus, ValidationGate types)
- `contracts/schemas/review.schema.json`
- `internal/modules/prompt/catalog/*` (adicionar perfil reviewer)
- `internal/modules/prompt/toolset.go` (ou equivalente)
- `tests/integration/*` (testes de review)
- `internal/modules/review/*_test.go` (novos testes)

### 🚫 Você NÃO DEVE tocar:
- `internal/modules/agent/` — pertence a outro agente
- `internal/modules/agentsession/` — pertence a outro agente
- `internal/modules/trigger/` — pertence a outro agente
- `internal/modules/task/`, `run/`, `workunit/`, `taskgraph/` (exceto para ler como referência)
- `cmd/` — pertence a rodada futura
- `internal/core/statemachine/` — só altere se adicionar novos aggregates (mas consulte primeiro)
- `internal/services/orchestrator_service.go` — não existe ainda

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

Você deve executar esta tarefa em ciclos curtos usando o arquivo de checklist persistente.

**Caminho do checklist:** `plans/ORCH-F05-R01-A02-review-checklist.md`

**A cada iteração:**
1. **LER** o checklist para identificar o próximo item pendente
2. **EXECUTAR** o item (código, teste, refactor)
3. **VALIDAR** o item (testes passam? comportamento correto?)
4. **ATUALIZAR** o checklist marcando o item como concluído
5. **CONTINUAR** para o próximo item

**Regras do Ralph Loop:**
- Nunca pule um item sem marcá-lo no checklist
- Se encontrar bloqueio, adicione uma nota na seção "Notas de Progresso"
- Se precisar adicionar itens ao checklist, faça-o (são raras exceções)
- Ao final de cada ciclo significativo, faça um commit pequeno
- O checklist é sua fonte de verdade de progresso

## Regras de Implementação
1. Siga rigorosamente o padrão dos módulos existentes.
2. State Machine: se Review tiver transições de status válidas, adicione em `internal/core/statemachine/`.
3. Use transações (BeginTx/CommitTx/RollbackTx) para todas as operações de escrita.
4. Eventos devem ser idempotentes (use event_id deduplication do Event Store).
5. Veredicto (`approved`, `changes_requested`, `needs_discussion`) deve ser imutável após emitido.

## Testes — Regras Rígidas
- Teste Create review com caminho feliz
- Teste SubmitVerdict com cada tipo de veredicto
- Teste que veredicto não pode ser alterado após emitido
- Teste ListPending retorna apenas reviews pendentes
- Teste validação de gate_type inválido
- Teste validação de verdict inválido
- Testes de integração com banco real (seguir padrão do projeto)
- Determinísticos, flexíveis, eficientes

## Code Review Auto-Crítico (OBRIGATÓRIO)
- [ ] Review pode ser criado sem work_unit_id? A lógica permite casos inválidos?
- [ ] Veredicto pode ser sobrescrito? Deve ser imutável.
- [ ] Eventos são emitidos corretamente para cada transição?
- [ ] JSONB `criteria_checked` é validado antes de persistir?
- [ ] Perfil reviewer no catálogo não permite tools destrutivas?
- [ ] Schema JSON está sync com struct de domínio?

## Critérios de Aceite
- [ ] Migration 013 cria tabela reviews com constraints
- [ ] `ReviewService.Create()` persiste review e emite evento
- [ ] `ReviewService.SubmitVerdict()` emite veredicto imutável
- [ ] `GetByID` e `ListPending` funcionam corretamente
- [ ] Domain types adicionados em `types.go`
- [ ] Schema `review.schema.json` válido e testado
- [ ] Perfil `reviewer` existe no catálogo de prompts
- [ ] Toolset para `reviewer` classifica tools corretamente
- [ ] Testes passam (`go test ./...`)
- [ ] `go build ./...` compila
- [ ] Code review auto-crítico realizado
- [ ] Checklist de execução completamente marcado

## Entrega Final
Ao concluir, responda ao usuário com:
1. **Resumo Executivo**
2. **Arquivos Criados/Modificados**
3. **Status dos Critérios de Aceite**
4. **Decisões Tomadas**
5. **Riscos ou Débitos**
6. **Instruções Git**
