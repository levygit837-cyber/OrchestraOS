# 🎯 Tarefa: Implementar AgentService com Persistência e Validar AgentID em AgentSessionService

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.  
> **Branch esperada:** `agent-a01/agentservice`  
> **Worktree esperada:** `../orchestraos-a01`  
> Se não estiver isolado, execute: `cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A01 agentservice`

## Contexto do Projeto
- **Nome:** OrchestraOS
- **Linguagem:** Go 1.24
- **Banco:** PostgreSQL (lib/pq)
- **Migrations:** Goose v3 (pressly/goose/v3)
- **Arquitetura:** Event Sourcing parcial + State Machine + Módulos Verticais
- **Padrões:** Cada módulo tem service.go + repository.go + queries.go + models.go + events.go + contract.go
- **Regra de Ouro:** Zero imports cross-module. Apenas internal/core/orchestration/ pode importar múltiplos módulos.

## Documentação Obrigatória
ANTES de escrever código, leia:
1. `/home/levybonito/Documentos/OrchestraOS/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/AGENTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/internal/modules/agent/README.md`
4. `/home/levybonito/Documentos/OrchestraOS/internal/modules/agent/CONTRACTS.md`
5. `/home/levybonito/Documentos/OrchestraOS/internal/modules/agentsession/README.md`
6. `/home/levybonito/Documentos/OrchestraOS/internal/modules/agentsession/CONTRACTS.md`
7. `/home/levybonito/Documentos/OrchestraOS/docs/adr/0021-agent-service.md`

## O que Já Existe
- `internal/modules/agent/` tem: runtime.go (interface Runtime), fake_runtime.go, gemini_runtime.go, models.go (com `Agent` struct local), contract.go, doc.go. **NÃO tem service, repository, queries**.
- `internal/modules/agentsession/` tem service.go completo com Create, Connect, Disconnect, Resume, Stop, Timeout, Fail, Heartbeat, Checkpoint.
- `AgentSessionService.Create()` aceita `AgentID` como string qualquer, sem validar existência no banco.
- `internal/domain/types.go` já define `type Agent struct` com: ID, Name, Profile, Capabilities, AllowedTools, DefaultPromptFragments, RuntimeType.
- Não existe tabela `agents` no banco. Nenhuma migration a criou.
- O campo `agent_id` em `agent_sessions` é VARCHAR(255) sem FK.

## O que Você Deve Implementar

### Parte A: Módulo Agent Completo
Crie o módulo `internal/modules/agent/` como um módulo de domínio completo, seguindo o padrão dos módulos existentes (task, run, workunit, agentsession):

1. **Migration `migrations/012_agents.sql`**
   - Crie tabela `agents` com: id (UUID PK), name, profile, capabilities (TEXT[]), allowed_tools (TEXT[]), default_prompt_fragments (TEXT[]), runtime_type, status, created_at, updated_at.
   - Adicione índices apropriados.
   - Adicione CHECK constraint para runtime_type válido: `fake`, `gemini`, `codex_cli`, `external`.
   - Adicione CHECK constraint para profile válido: `code_worker`, `docs_writer`, `reviewer`, `debugger`, `default`.

2. **`internal/modules/agent/service.go`**
   - `AgentService` struct com `*sql.DB`
   - `Create(ctx, input) -> (*transition.OperationResult[*domain.Agent], error)`
   - `GetByID(ctx, id) -> (*domain.Agent, error)`
   - `FindOrCreate(ctx, profile, runtimeType) -> (*domain.Agent, error)` — busca agente ativo com perfil, ou cria novo
   - Emite eventos `agent.created` via Event Store
   - Use o padrão de transação dos outros serviços (BeginTx, CommitTx, RollbackTx)

3. **`internal/modules/agent/repository.go`**
   - CRUD puro: Create, GetByID, FindByProfileAndRuntime, List
   - Siga o padrão de outros repositories (ex: `internal/modules/agentsession/repository.go`)

4. **`internal/modules/agent/queries.go`**
   - SQL constants, seguindo o padrão dos outros módulos

5. **`internal/modules/agent/events.go`**
   - Mapeamento de event types para status, se necessário

6. **`internal/modules/agent/validation.go`**
   - Validação de inputs: profile válido, runtime_type válido, name não vazio

7. **`internal/modules/agent/models.go`**
   - Adapte para usar `domain.Agent` em vez do `Agent` local, se apropriado. Ou mantenha alias. Siga o padrão dos outros módulos.

8. **`internal/modules/agent/contract.go`** e **`CONTRACTS.md`**
   - Atualize para refletir que o módulo agora também gerencia persistência de agentes
   - Mantenha as regras críticas existentes

9. **`internal/modules/agent/README.md`**
   - Atualize para documentar o service, repository e uso do FindOrCreate

### Parte B: Validar AgentID em AgentSessionService
1. Em `internal/modules/agentsession/service.go`, na função `Create()`, após validar o input, verifique se o `AgentID` referencia um agente existente na tabela `agents`.
   - Você precisará de um `AgentReader` interface injetado no `AgentSessionService`, OU fazer a validação via query direta no repository. 
   - **Preferência:** siga o padrão de DI do projeto. Veja como `WorkUnitService` usa `TaskReader` interface para evitar import cíclico. Crie uma interface `AgentReader` com `GetByID(ctx, id) (*domain.Agent, error)` e injete no `AgentSessionService`.
   - Se o agente não existir, retorne `apperrors.New(apperrors.CodeNotFound, ...)`.

2. Atualize `NewAgentSessionService` para aceitar o `AgentReader`.

3. Atualize `internal/modules/agentsession/contract.go` se necessário.

### Parte C: Atualizar Testes
1. Atualize testes existentes em `tests/integration/` e `internal/modules/agentsession/` que criam `AgentSession` com `AgentID` arbitrário. Eles devem primeiro criar um agente via `AgentService.Create()` ou usar um ID de agente existente.
2. Crie testes unitários para `AgentService` em `internal/modules/agent/service_test.go`:
   - Teste Create com caminho feliz
   - Teste GetByID
   - Teste FindOrCreate (cria novo, depois encontra existente)
   - Teste validação de profile inválido
   - Teste validação de runtime_type inválido
3. Crie testes para a validação de AgentID em `internal/modules/agentsession/service_test.go` (ou adapte existentes).

## Fronteiras de Isolamento

### ✅ Você PODE e DEVE tocar:
- `migrations/012_agents.sql`
- `internal/modules/agent/*` (todos os arquivos deste diretório)
- `internal/modules/agentsession/service.go` (apenas para validar AgentID)
- `internal/modules/agentsession/contract.go`, `README.md`, `CONTRACTS.md` (se precisar documentar a mudança)
- `tests/integration/*` (atualizar testes que quebrarem)
- `internal/modules/agent/*_test.go` (novos testes)

### 🚫 Você NÃO DEVE tocar:
- `internal/modules/review/` — pertence a outro agente
- `internal/modules/trigger/` — pertence a outro agente
- `internal/modules/prompt/catalog/` — pertence a outro agente
- `internal/modules/task/`, `run/`, `workunit/`, `taskgraph/`, `prompt/` (exceto para ler como referência)
- `cmd/` — pertence a rodada futura
- `internal/services/orchestrator_service.go` — não existe ainda, rodada futura
- `internal/core/orchestration/` — não altere sem necessidade crítica

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

Você deve executar esta tarefa em ciclos curtos usando o arquivo de checklist persistente.

**Caminho do checklist:** `plans/ORCH-F05-R01-A01-agentservice-checklist.md`

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
1. Siga rigorosamente o padrão de código dos módulos existentes (task, run, workunit, agentsession).
2. Use `uuid.New().String()` para novos IDs, como os outros serviços.
3. Use `time.Now().UTC()` para timestamps.
4. Valide entradas nas bordas usando `internal/core/validation/`.
5. Trate erros com `apperrors.Wrap()` ou `apperrors.New()`.
6. Nunca silencie erros. Sempre retorne ou logue.
7. Não adicione novas dependências externas (go get) sem justificativa explícita.
8. Não quebre testes existentes. Rode `go test ./...` antes e depois.
9. State machine: se adicionar novos status para Agent, atualize `internal/core/statemachine/`.

## Testes — Regras Rígidas
- Crie testes REAIS usando o padrão de teste do projeto.
- Use `sql.NullString`, `sql.NullTime` quando apropriado, seguindo o padrão existente.
- Para testes de integração com banco: use o helper de teste existente no projeto (veja `tests/integration/services_test.go` como referência).
- Testes determinísticos: nada de randomização sem seed fixo.
- Testes flexíveis: use factories/helpers, não hardcode UUIDs.
- Cobertura mínima: Create, GetByID, FindOrCreate, validações de input, validação de AgentID em AgentSession.

## Code Review Auto-Crítico (OBRIGATÓRIO)
Após implementar e testar, faça releitura pragmática buscando:
- [ ] Erros de lógica em FindOrCreate (race condition entre find e create)
- [ ] AgentID vazio ou inválido sendo aceito em algum caminho
- [ ] SQL injection em queries (use parâmetros $1, $2)
- [ ] Transações não comitadas ou não rollbackadas
- [ ] Eventos não emitidos ou emitidos com payload errado
- [ ] State machine desatualizada para novos status de Agent
- [ ] Testes que passam mas não testam comportamento real
- [ ] Imports cíclicos (verifique com `go build ./...`)
- Corrija tudo antes de entregar.

## Critérios de Aceite
- [ ] Migration `012_agents.sql` cria tabela com constraints corretas
- [ ] `AgentService.Create()` persiste agente e emite evento `agent.created`
- [ ] `AgentService.GetByID()` retorna agente existente
- [ ] `AgentService.FindOrCreate()` reutiliza agente existente ou cria novo
- [ ] `AgentSessionService.Create()` rejeita AgentID inexistente com erro NotFound
- [ ] Testes existentes de agentsession continuam passando (adaptados)
- [ ] Novos testes de AgentService passam
- [ ] `go test ./...` passa sem erros
- [ ] `go build ./...` compila sem erros
- [ ] Code review auto-crítico realizado
- [ ] Checklist de execução completamente marcado

## Entrega Final
Ao concluir, responda ao usuário com:
1. **Resumo Executivo**
2. **Arquivos Criados/Modificados**
3. **Status dos Critérios de Aceite**
4. **Decisões Tomadas**
5. **Riscos ou Débitos**
6. **Instruções Git** — solicite: merge, PR, commit, push, ou revisão
