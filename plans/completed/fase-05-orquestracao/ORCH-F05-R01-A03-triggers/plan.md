# 🎯 Tarefa: Implementar Sistema de Triggers Configuráveis com Detecção de Anomalias

> **⚠️ OBRIGAÇÃO DE ISOLAMENTO:** Antes de começar, confirme que está isolado.  
> **Branch esperada:** `agent-a03/triggers`  
> **Worktree esperada:** `../orchestraos-a03`  
> Se não estiver isolado, execute: `cd /home/levybonito/Documentos/OrchestraOS && ./scripts/bootstrap-agent-worktree.sh A03 triggers`

## Contexto do Projeto
- **Nome:** OrchestraOS
- **Linguagem:** Go 1.24
- **Banco:** PostgreSQL (lib/pq)
- **Migrations:** Goose v3 (pressly/goose/v3)
- **Arquitetura:** Event Sourcing parcial + State Machine + Módulos Verticais
- **Padrões:** Cada módulo tem service.go + repository.go + queries.go + models.go + events.go + contract.go
- **Event Store:** Append-only com sequência, idempotência por event_id, replay de estado.

## Documentação Obrigatória
ANTES de escrever código, leia:
1. `/home/levybonito/Documentos/OrchestraOS/README.md`
2. `/home/levybonito/Documentos/OrchestraOS/AGENTS.md`
3. `/home/levybonito/Documentos/OrchestraOS/docs/implementation/roadmap.md` (seção Fase 5, item 6)
4. `/home/levybonito/Documentos/OrchestraOS/internal/core/eventstore/README.md` (se existir)
5. `/home/levybonito/Documentos/OrchestraOS/internal/modules/run/CONTRACTS.md`
6. `/home/levybonito/Documentos/OrchestraOS/internal/modules/agentsession/CONTRACTS.md`

## O que Já Existe
- Event Store completo em `internal/core/eventstore/`
- `RunService` com retry policy e timeout
- `AgentSessionService` com heartbeat e checkpoint
- `WorkUnitService` com validação de paths e dependências
- Não existe módulo `trigger/`.
- Não existe detecção automática de stall, loop, drift ou violação de paths.

## O que Você Deve Implementar

### Parte A: Módulo Trigger Completo
Crie `internal/modules/trigger/` como um módulo de domínio completo:

1. **Migration `migrations/014_triggers.sql`**
   - Tabela `triggers` com: id (UUID PK), run_id (UUID FK nullable), task_id (UUID FK nullable), agent_session_id (UUID FK nullable), trigger_type (`threshold`, `anomaly`, `heartbeat_timeout`, `policy`), status (`active`, `triggered`, `resolved`, `dismissed`), anomaly_type (`stall`, `loop`, `drift`, `path_violation`, `token_exceeded`, `steps_exceeded`, `time_exceeded`), threshold_value (JSONB), current_value (JSONB), triggered_at, resolved_at, resolution_action (`pause`, `cancel`, `notify`, `escalate`), created_at.
   - Índices em run_id, task_id, status, anomaly_type, trigger_type.

2. **`internal/modules/trigger/service.go`**
   - `TriggerService` struct com `*sql.DB`
   - `Create(ctx, input) -> (*transition.OperationResult[*domain.Trigger], error)`
   - `EvaluateRun(ctx, runID) -> ([]*domain.Trigger, error)` — avalia uma run em busca de anomalias
   - `EvaluateSession(ctx, sessionID) -> ([]*domain.Trigger, error)` — avalia heartbeat e checkpoints
   - `EvaluateWorkUnit(ctx, workUnitID) -> ([]*domain.Trigger, error)` — avalia path violations
   - `Resolve(ctx, triggerID, action, reason) -> ...`
   - `Dismiss(ctx, triggerID, reason) -> ...`
   - `ListActive(ctx) -> ([]*domain.Trigger, error)`
   - `ListByRun(ctx, runID) -> ([]*domain.Trigger, error)`
   - Eventos: `trigger.created`, `trigger.triggered`, `trigger.resolved`, `trigger.dismissed`

3. **`internal/modules/trigger/detectors.go`**
   - Implemente detectores determinísticos (sem LLM, sem randomização):
     - **StallDetector:** detecta quando uma run/sessão não emite eventos por X tempo (threshold configurável)
     - **LoopDetector:** detecta quando o mesmo tipo de evento se repete N vezes em sequência (ex: 5 heartbeats sem progresso)
     - **DriftDetector:** detecta quando o agente executa steps/tools fora do escopo definido em `owned_paths` / `read_paths`
     - **PathViolationDetector:** detecta quando um agente tenta modificar paths não pertencentes a ele
     - **TokenThresholdDetector:** detecta quando tokens consumidos excedem threshold (campo current_value como proxy)
     - **StepsThresholdDetector:** detecta quando steps excedem max_steps configurado
     - **TimeThresholdDetector:** detecta quando tempo total de execução excede timeout configurado
   - Cada detector recebe o estado atual (via Event Store replay ou parâmetros) e retorna `[]*domain.Trigger` ou erro.
   - Os detectores são puros (sem side effects, sem I/O).

4. **`internal/modules/trigger/thresholds.go`**
   - `ThresholdConfig` struct com campos para cada tipo de threshold
   - `DefaultThresholds() -> ThresholdConfig` — retorna thresholds padrão conservadores
   - `ValidateThresholds(config) error` — valida que thresholds são positivos e razoáveis

5. **`internal/modules/trigger/repository.go`**
   - CRUD puro: Create, GetByID, UpdateStatus, ListActive, ListByRun

6. **`internal/modules/trigger/queries.go`**
   - SQL constants

7. **`internal/modules/trigger/models.go`**
   - Aliases para domain types

8. **`internal/modules/trigger/events.go`**
   - Mapeamento event types

9. **`internal/modules/trigger/contract.go`, `doc.go`, `README.md`, `CONTRACTS.md`**
   - Siga o padrão dos outros módulos

10. **`internal/modules/trigger/validation.go`**
    - Validação de trigger_type, anomaly_type, threshold_value

### Parte B: Adicionar Types ao Domain
Em `internal/domain/types.go`, adicione:
- `TriggerType` type + constantes: `threshold`, `anomaly`, `heartbeat_timeout`, `policy`
- `TriggerStatus` type + constantes: `active`, `triggered`, `resolved`, `dismissed`
- `AnomalyType` type + constantes: `stall`, `loop`, `drift`, `path_violation`, `token_exceeded`, `steps_exceeded`, `time_exceeded`
- `ResolutionAction` type + constantes: `pause`, `cancel`, `notify`, `escalate`
- `Trigger` struct com todos os campos
- `ThresholdConfig` struct (ou deixe no módulo trigger se preferir)

### Parte C: Adicionar Schema JSON
Em `contracts/schemas/trigger.schema.json`:
- Schema JSON Schema 2020-12 para a entidade Trigger
- Siga o padrão dos schemas existentes

## Fronteiras de Isolamento

### ✅ Você PODE e DEVE tocar:
- `migrations/014_triggers.sql`
- `internal/modules/trigger/*` (diretório completo novo)
- `internal/domain/types.go` (adicionar Trigger, TriggerType, TriggerStatus, AnomalyType, ResolutionAction)
- `contracts/schemas/trigger.schema.json`
- `tests/integration/*` (testes de trigger)
- `internal/modules/trigger/*_test.go` (novos testes)

### 🚫 Você NÃO DEVE tocar:
- `internal/modules/agent/` — pertence a outro agente
- `internal/modules/agentsession/` — pertence a outro agente
- `internal/modules/review/` — pertence a outro agente
- `internal/modules/task/`, `run/`, `workunit/`, `taskgraph/`, `prompt/` (exceto para ler como referência)
- `cmd/` — pertence a rodada futura
- `internal/services/orchestrator_service.go` — não existe ainda
- `internal/core/eventstore/` — pode ler mas não altere a lógica core

## Ralph Loop — Execução Iterativa (OBRIGATÓRIO)

Você deve executar esta tarefa em ciclos curtos usando o arquivo de checklist persistente.

**Caminho do checklist:** `plans/ORCH-F05-R01-A03-triggers-checklist.md`

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
1. Detectores devem ser DETERMINÍSTICOS. Mesmos eventos → mesmos triggers, sempre.
2. Detectores não devem ter side effects. Eles apenas ANALISAM e RETORNAM triggers.
3. O `TriggerService` é quem persiste triggers e emite eventos.
4. Thresholds devem ter valores padrão conservadores (ex: stall > 60s, loop > 5 repetições).
5. Use `json.RawMessage` ou `map[string]interface{}` para campos JSONB (threshold_value, current_value).
6. Siga o padrão de transação dos outros serviços.

## Testes — Regras Rígidas
- Teste cada detector com casos positivos (anomalia detectada) e negativos (normalidade)
- Teste StallDetector com eventos espaçados vs eventos ausentes
- Teste LoopDetector com repetição de eventos idênticos
- Teste PathViolationDetector com paths dentro e fora do escopo
- Teste Thresholds com valores dentro e fora do limite
- Teste que triggers são persistidos e eventos emitidos
- Teste `ListActive` retorna apenas triggers não resolvidos
- Testes determinísticos, flexíveis, eficientes
- Não dependa de tempo real — use timestamps fixos nos testes

## Code Review Auto-Crítico (OBRIGATÓRIO)
- [ ] Detectores são realmente determinísticos? (Sem time.Now() sem mock?)
- [ ] LoopDetector não gera falso positivo em eventos legítimos repetidos?
- [ ] PathViolationDetector lida corretamente com paths absolutos vs relativos?
- [ ] Thresholds padrão não são agressivos demais (não vão disparar em execução normal)?
- [ ] JSONB é validado antes de persistir?
- [ ] Eventos de trigger não causam loop infinito (trigger cria evento que cria trigger)?
- [ ] Memory leaks em caches ou buffers de detector?

## Critérios de Aceite
- [ ] Migration 014 cria tabela triggers com constraints
- [ ] `TriggerService.Create()` persiste trigger e emite evento
- [ ] Cada detector (stall, loop, drift, path_violation, token, steps, time) detecta corretamente
- [ ] `EvaluateRun`, `EvaluateSession`, `EvaluateWorkUnit` retornam triggers quando há anomalias
- [ ] Domain types adicionados em `types.go`
- [ ] Schema `trigger.schema.json` válido
- [ ] Testes de todos os detectores passam
- [ ] `go test ./...` passa
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
```
