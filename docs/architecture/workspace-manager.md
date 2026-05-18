# Workspace Manager (WSM) — Arquitetura

> Este documento descreve a arquitetura do **Workspace Manager**, componente responsável por permitir que múltiplos Agentes de IA trabalhem paralelamente na mesma codebase com isolamento, rastreabilidade e controle de merge.
>
> Para a decisão arquitetural que aprovou esta solução, consulte [ADR-0029: Workspace Manager Nativo para Paralelização de Agentes](/docs/adr/0029-workspace-manager-for-agents.md).

---

## 1. Visão Geral

O Workspace Manager (WSM) é uma camada de orquestração de filesystem e versionamento construída **nativamente sobre o OrchestraOS**. Ele não substitui o Git — usa o Git como backend de persistência — mas cria uma abstração que torna o Git operacional para sistemas multi-agente.

Cada agente recebe um **Virtual Workspace** isolado, com snapshots temporários durante a execução e consolidação ao final. O WSM mantém um **grafo de operações** sobre todas as modificações, permitindo rollback parcial, resolução de conflitos e merge coordenado.

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                         OrchestraOS Control Plane                           │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                     Workspace Manager (WSM)                         │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │   │
│  │  │   Virtual    │  │   Snapshot   │  │      Merge Orchestrator  │  │   │
│  │  │  Workspace   │  │   Engine     │  │      (Conflict Graph)    │  │   │
│  │  │   Service    │  │              │  │                          │  │   │
│  │  └──────────────┘  └──────────────┘  └──────────────────────────┘  │   │
│  │  ┌──────────────────────────────────────────────────────────────┐  │   │
│  │  │              Operation Event Graph (CRDT/OT)                 │  │   │
│  │  └──────────────────────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│                        ┌───────────┴───────────┐                            │
│                        ▼                       ▼                            │
│  ┌──────────────────────────┐      ┌──────────────────────────┐            │
│  │  Git Backend (Backend)   │      │   Semantic Lock Service  │            │
│  │  (refs, commits, trees)  │      │   (domínios reservados)  │            │
│  └──────────────────────────┘      └──────────────────────────┘            │
└─────────────────────────────────────────────────────────────────────────────┘
         │                    │                    │
    ┌────┴────┐          ┌────┴────┐          ┌────┴────┐
    ▼         ▼          ▼         ▼          ▼         ▼
┌───────┐ ┌───────┐  ┌───────┐ ┌───────┐  ┌───────┐ ┌───────┐
│Agent-1│ │Agent-2│  │Agent-3│ │Agent-4│  │Agent-5│ │Agent-N│
│  WS   │ │  WS   │  │  WS   │ │  WS   │  │  WS   │ │  WS   │
└───────┘ └───────┘  └───────┘ └───────┘  └───────┘ └───────┘
```

---

## 2. Problema

O Git foi projetado para **um usuário humano** em um único contexto de trabalho. Quando múltiplos Agentes de IA operam simultaneamente na mesma codebase, encontramos três categorias de problema:

| Categoria | Problema | Impacto |
|-----------|----------|---------|
| **Isolamento** | Worktrees compartilham o mesmo `.git/` e namespace de refs | Commits de um agente corrompem o estado de outro na mesma branch |
| **Coordenação** | Clones independentes não têm awareness de mudanças paralelas | Merge vira caixa de surpresa; conflitos só aparecem no final |
| **Rastreabilidade** | Git não rastreia *intenção* de modificação | Não sabemos se dois agentes tocaram no mesmo arquivo por acidente ou design |

Soluções parciais (worktrees, clones, branches separadas) falham porque tratam o sintoma (necessidade de diretórios separados) e não a causa (necessidade de **coordenação ativa de estado entre processos paralelos**).

---

## 3. Princípios

O WSM é regido por 7 princípios inegociáveis:

1. **Isolamento Total por Agente** — O filesystem de um agente nunca é visível ou mutável por outro agente.
2. **Snapshot por Estado de Arquivo** — Snapshots capturam o estado individual de arquivos, não o workspace inteiro. Rollback afeta apenas arquivos conflitantes.
3. **Toda Modificação é um Evento** — Nenhuma escrita acontece sem ser registrada no grafo de operações.
4. **Git é Backend, não Interface** — Agentes não executam `git` diretamente. O WSM traduz operações do agente em commits Git.
5. **Merge é Orquestrado** — Conflitos não são surpresas. O sistema detecta colisões antes de permitir que o merge prossiga.
6. **Retomada sem Re-leitura** — Um agente interrompido retoma do último snapshot com contexto completo (diffs, estado, tarefa).
7. **Descarte Atômico** — Um workspace pode ser completamente descartado sem deixar vestígios no repositório principal.

---

## 4. Arquitetura de Camadas

```text
┌────────────────────────────────────────────┐
│  Layer 4: Agent Runtime Interface          │
│  (API que o agente consome — sem git)      │
├────────────────────────────────────────────┤
│  Layer 3: Workspace Lifecycle Service      │
│  (init → snapshot → commit → merge → end)  │
├────────────────────────────────────────────┤
│  Layer 2: Virtual Filesystem Engine        │
│  (overlay/COW, diff engine, patch apply)   │
├────────────────────────────────────────────┤
│  Layer 1: Git Abstraction & Event Store    │
│  (refs, commits, trees, operation events)  │
└────────────────────────────────────────────┘
```

### Layer 1 — Git Abstraction & Event Store

Responsável por toda interação com o repositório Git e pelo grafo de operações.

- **GitAdapter**: traduz comandos do WSM para operações Git (checkout, commit, branch, merge).
- **OperationEventStore**: persiste cada modificação de arquivo como um evento estruturado.
- **RefManager**: gerencia refs privadas do WSM (`refs/wsm/{agent-id}/{snapshot-id}`).

### Layer 2 — Virtual Filesystem Engine

Responsável por criar e gerenciar o diretório de trabalho de cada agente.

- **WorkspaceProvider**: cria o diretório físico ou virtual do agente.
- **DiffEngine**: calcula diffs entre snapshots, entre workspace e base, entre branches.
- **PatchApplier**: aplica patches com estratégia de 3-way merge.
- **SnapshotStore**: armazena snapshots temporários e reais.

### Layer 3 — Workspace Lifecycle Service

Orquestra o ciclo de vida completo de um workspace.

- **LifecycleService**: `InitWorkspace`, `Checkpoint`, `Suspend`, `Resume`, `Finalize`, `Discard`.
- **SnapshotService**: cria snapshots temporários (a cada diff) e snapshots reais (ao finalizar).
- **MergeService**: coordena tentativas de merge, detecta conflitos, orquestra resolução.

### Layer 4 — Agent Runtime Interface

A única interface que o agente vê.

- **FileAPI**: `ReadFile`, `WriteFile`, `DeleteFile`, `ListDirectory`.
- **WorkspaceAPI**: `GetStatus`, `GetDiff`, `CreateCheckpoint`, `RequestMerge`.
- **ContextAPI**: `GetTask`, `GetHistory`, `GetConflicts`.

---

## 5. Componentes Principais

### 5.1 Virtual Workspace

Um Virtual Workspace é uma instância de trabalho isolada atribuída a um agente para uma tarefa.

```go
type VirtualWorkspace struct {
    ID              string        // UUID do workspace
    AgentID         string        // Agente dono
    TaskID          string        // Tarefa associada
    BaseCommit      string        // Commit base (ponto de partida)
    BranchName      string        // Branch exclusiva do agente
    RootPath        string        // Caminho físico no filesystem
    Status          WorkspaceStatus // init, active, suspended, merging, finalized, discarded
    CreatedAt       time.Time
    LastCheckpoint  *Snapshot
}
```

**Regra absoluta:** Nenhum workspace pode compartilhar `BranchName` com outro workspace ativo.

#### Estratégias de Isolamento de Filesystem

O WSM suporta três estratégias de isolamento, selecionáveis por configuração:

| Estratégia | Mecanismo | Prós | Contras |
|------------|-----------|------|---------|
| `git-clone` | Clone completo independente | Total isolamento, funciona em qualquer SO | Overhead de disco, tempo de clone |
| `overlayfs` | Overlay filesystem (Linux) | Leve, instantâneo, descarte rápido | Apenas Linux, requer kernel com overlay |
| `bind-copy` | Cópia seletiva + diretório de trabalho | Portável, controle total | Overhead de cópia inicial |

A estratégia padrão para MVP é `git-clone`. `overlayfs` é ativada automaticamente quando detectado Linux com suporte.

### 5.2 Snapshot Engine

O Snapshot Engine gerencia dois tipos de snapshot:

#### Snapshot Temporário (Transient Snapshot)

- Criado automaticamente a cada modificação significativa (após N segundos de inatividade ou comando explícito do agente).
- Armazena o estado completo de cada arquivo modificado.
- Usado para rollback, retomada e inspeção de histórico.
- **Descartado** quando o workspace é finalizado ou descartado.
- NÃO é commit Git — é armazenado como objetos Git (tree + blob) em refs privadas do WSM (`refs/wsm/{agent-id}/snapshots/transient-{seq}`), garbage coletados após descarte.

```go
type TransientSnapshot struct {
    ID            string
    WorkspaceID   string
    Sequence      int           // Ordem dentro do workspace
    FileStates    []FileState   // Estado individual de cada arquivo
    CreatedAt     time.Time
    Trigger       string        // auto, manual, pre-merge
}

type FileState struct {
    Path        string
    ContentHash string        // SHA-256 do conteúdo
    Content     []byte        // Conteúdo completo (compressão opcional)
    Operation   OpType        // create, modify, delete
    PreviousID  string        // ID do snapshot anterior deste arquivo
}
```

#### Snapshot Real (Persistent Snapshot)

- Criado apenas ao **finalizar** a tarefa com sucesso.
- Convertido em **commit Git** na branch do agente.
- Torna-se parte do histórico permanente do repositório.
- Pode ser referenciado por outros agentes e pelo Merge Orchestrator.

```go
type PersistentSnapshot struct {
    ID            string
    WorkspaceID   string
    CommitHash    string        // SHA do commit Git
    BranchName    string
    ParentCommit  string        // Commit base
    DiffSummary   DiffSummary   // Metadados do que mudou
    CreatedAt     time.Time
}
```

#### Rollback por Arquivo

Quando um conflito é detectado, o WSM pode realizar **rollback seletivo**:

```
Conflito detectado: service.go modificado por Agent-A e Agent-B

Ação: Rollback de service.go no workspace de Agent-B
       → service.go volta ao estado do snapshot N-2
       → Demais arquivos de Agent-B permanecem intactos
       → Agent-B recebe notificação com contexto do conflito
```

Isso é possível porque snapshots armazenam **estado por arquivo**, não estado global do workspace.

### 5.3 Operation Event Graph

Toda modificação de arquivo gera um evento estruturado que alimenta um grafo de operações.

```go
type OperationEvent struct {
    ID          string
    WorkspaceID string
    AgentID     string
    TaskID      string
    Timestamp   time.Time
    Type        OpType        // create, modify, delete, rename
    Target      FileTarget    // arquivo ou diretório afetado
    ContentHash string        // hash do conteúdo pós-modificação
    DiffHash    string        // hash do diff aplicado
    Dependencies []string     // IDs de eventos que este evento depende
    Metadata    OpMetadata    // linhas afetadas, símbolos tocados, etc.
}
```

#### Estrutura do Grafo

O grafo é um **DAG (Directed Acyclic Graph)** onde:

- **Nós** são eventos de operação (`OperationEvent`).
- **Arestas** são dependências (evento B depende do estado produzido por evento A).
- **Componentes conectados** representam modificações no mesmo arquivo ou domínio semântico.

```
Evento-1: Agent-A cria auth.go
    │
    ▼
Evento-2: Agent-A modifica auth.go (linhas 10-30)
    │
    ├──► Evento-3: Agent-B modifica auth.go (linhas 25-40)  ← CONFLITO!
    │         (depende do estado do Evento-1, colide com Evento-2)
    │
    ▼
Evento-4: Agent-A cria user_service.go
    │
    ▼
Evento-5: Agent-C modifica login.html (sem relação com auth.go)
```

#### Detecção de Conflito pelo Grafo

Conflitos são detectados **antes do merge**, analisando o grafo:

1. Dois eventos modificam o mesmo arquivo → **potencial conflito de conteúdo**
2. Dois eventos modificam símbolos relacionados (ex: função e seu teste) → **potencial conflito semântico**
3. Um evento deleta um arquivo que outro modifica → **conflito de exclusão**
4. Eventos em arquivos independentes → **sem conflito, merge paralelo permitido**

#### Granularidade do Grafo

O grafo pode operar em diferentes níveis de granularidade, configurável por projeto:

| Nível | Unidade de Rastreamento | Uso |
|-------|------------------------|-----|
| `file` | Arquivo inteiro | MVP, mais simples |
| `symbol` | Função, struct, interface | Requer Code Intelligence (LSP/Tree-sitter) |
| `line-range` | Range de linhas | Balanceio entre simplicidade e precisão |
| `architecture` | Módulo/pacote | Alto nível, para Semantic Locking |

O padrão para o MVP é `file`. Níveis mais finos são ativados quando o módulo de Code Intelligence estiver disponível.

### 5.4 Merge Orchestrator

O Merge Orchestrator é a camada que decide **quando** e **como** os snapshots de agentes são integrados ao branch principal.

#### Estados do Merge

```
PENDING → ANALYZING → (CLEAN_MERGE | CONFLICT_DETECTED)
                          │                    │
                          ▼                    ▼
                    AUTO_MERGED          AWAITING_RESOLUTION
                          │                    │
                          ▼                    ▼
                    VERIFIED             RESOLVED
                          │                    │
                          └──────────┬─────────┘
                                     ▼
                               INTEGRATED
```

#### Fases do Merge

1. **Análise** — O Merge Orchestrator consulta o Event Graph e identifica todas as colisões.
2. **Tentativa Automática** — Para conflitos simples (arquivos diferentes, mudanças ortogonais), o merge é aplicado automaticamente via `git merge-tree` ou diff patch.
3. **Bloqueio com Explicação** — Se conflitos existem, o merge é bloqueado. O sistema gera um relatório estruturado:
   ```json
   {
     "merge_id": "merge-uuid",
     "status": "CONFLICT_DETECTED",
     "conflicts": [
       {
         "file": "service.go",
         "type": "content",
         "agents": ["agent-a", "agent-b"],
         "agent_a_changes": "linhas 10-30: adicionou validação",
         "agent_b_changes": "linhas 25-40: refatorou função",
         "suggested_resolution": "refatorar função com validação incluída"
       }
     ],
     "rollback_available": ["agent-b:service.go@snapshot-3"]
   }
   ```
4. **Resolução** — Pode ser:
   - **Humana**: um desenvolvedor resolve o conflito manualmente.
   - **Orquestrada**: o Orchestrator cria uma nova tarefa de resolução e assigna a um agente.
   - **Automática**: se a política permitir, um agente resolvedor tenta auto-merge com estratégia de 3-way.
5. **Integração** — Após resolução, o resultado é commitado no branch principal.

#### Semantic Lock Service

Antes de um agente iniciar, o Semantic Lock Service reserva domínios:

```go
type SemanticLock struct {
    AgentID       string
    TaskID        string
    WorkspaceID   string
    LockedPaths   []string      // arquivos ou diretórios reservados
    LockType      LockType      // exclusive, shared-read
    ExpiresAt     time.Time     // TTL para evitar locks órfãos
}
```

- **Exclusive**: apenas o agente dono pode escrever. Outros agentes que precisarem do mesmo arquivo são enfileirados ou redirecionados.
- **Shared-Read**: múltiplos agentes podem ler, mas nenhum pode escrever.

**Importante:** locks são uma otimização para reduzir conflitos, não uma garantia. Agentes são não-determinísticos e podem expandir escopo. O Event Graph e o Merge Orchestrator são a última linha de defesa.

---

## 6. Ciclo de Vida do Workspace

```
┌─────────┐    ┌──────────┐    ┌──────────┐    ┌───────────┐    ┌──────────┐
│  INIT   │───►│  ACTIVE  │───►│ CHECKPOINT│───►│  MERGING  │───►│ FINALIZED│
└────┬────┘    └────┬─────┘    └────┬─────┘    └─────┬─────┘    └────┬─────┘
     │              │               │                 │                │
     │              │               │                 │                │
     ▼              ▼               ▼                 ▼                ▼
┌─────────┐    ┌──────────┐    ┌──────────┐    ┌───────────┐    ┌──────────┐
│DISCARDED│    │SUSPENDED │    │ RESUMED  │    │CONFLICT   │    │INTEGRATED│
└─────────┘    └──────────┘    └──────────┘    │RESOLUTION │    └──────────┘
                                                └───────────┘
```

### 6.1 INIT

1. Orchestrator solicita workspace para Task T ao WSM.
2. WSM verifica Semantic Lock Service — o domínio da tarefa está livre?
3. WSM cria Virtual Workspace com branch exclusiva (`wsm/{agent-id}/{task-id}`).
4. WSM inicializa o diretório de trabalho (clone/overlay/cópia).
5. WSM registra o workspace no Event Store.
6. Agente recebe contexto completo: task, base commit, workspace path, regras.

### 6.2 ACTIVE

1. Agente executa a tarefa usando a FileAPI do WSM.
2. Cada `WriteFile` gera um `OperationEvent` no grafo.
3. Snapshot Engine cria Transient Snapshots automaticamente (debounce de 30s ou a cada 10 operações).
4. Agente pode solicitar checkpoint manual via `CreateCheckpoint`.

### 6.3 CHECKPOINT

1. Snapshot Engine persiste o estado atual de todos os arquivos modificados.
2. WSM atualiza o ledger do AgentSession com metadados do checkpoint.
3. Se o agente for suspenso, ele retoma deste ponto.

### 6.4 SUSPEND / RESUME

**Suspend** (manual ou pelo Orchestrator):
1. Criar Transient Snapshot final.
2. Persistir estado do grafo.
3. Liberar recursos do filesystem (manter snapshot em cache).
4. Marcar workspace como `suspended`.

**Resume**:
1. Restaurar workspace a partir do último snapshot.
2. Recarregar Event Graph.
3. Entregar ao agente: estado atual, diffs pendentes, histórico de operações.
4. Agente continua sem re-ler o projeto do zero.

### 6.5 MERGING

1. Agente sinaliza conclusão da tarefa.
2. WSM cria Persistent Snapshot (commit Git na branch do agente).
3. Merge Orchestrator inicia análise.
4. Se clean merge → auto-merge → estado INTEGRATED.
5. Se conflito → estado CONFLICT_DETECTED → relatório gerado → resolução orquestrada.

### 6.6 FINALIZED / DISCARDED

**Finalized**:
- Merge integrado com sucesso.
- Workspace mantido por política de retenção (ex: 24h), depois limpo.

**Discarded**:
- Agente ou humano descarta o workspace.
- Todos os Transient Snapshots são removidos.
- Se Persistent Snapshot existir, a branch é deletada (`git branch -D`).
- Nenhum vestígio permanece no repositório principal.

---

## 7. Integração com Git

O WSM não substitui o Git. Ele cria uma camada operacional sobre o Git.

### 7.1 Estratégia de Refs

O WSM usa refs privadas para não poluir o namespace de branches do usuário:

```
refs/
  heads/
    main
    feature-xyz
  wsm/                          ← namespace privado do WSM
    {agent-id}/
      {workspace-id}/           ← branch do workspace
        HEAD
      snapshots/
        transient-{seq}         ← refs para snapshots temporários
```

### 7.2 Commits e Árvore

- **Transient Snapshots**: armazenados como objetos Git (tree + blob) em refs especiais, mas não fazem parte do histórico principal. São garbage coletados após descarte do workspace.
- **Persistent Snapshots**: commits normais em branches `refs/wsm/{agent-id}/{workspace-id}`. Podem ser mergeados via PR ou fast-forward.

### 7.3 Sincronização com Upstream

Cada workspace é inicializado a partir de um commit base (tipicamente `main`). Durante a vida do workspace:

1. O WSM pode opcionalmente fazer `git fetch` para detectar mudanças no upstream.
2. Se o base divergiu significativamente, o WSM pode rebasar o workspace automaticamente ou solicitar intervenção.
3. O Merge Orchestrator sempre integra contra o `main` mais recente no momento do merge.

---

## 8. Regras e Invariantes

| # | Invariante | Garantia |
|---|------------|----------|
| 1 | Dois workspaces ativos nunca compartilham a mesma branch | Unicidade de `BranchName` verificada no `INIT` |
| 2 | Um arquivo modificado por um agente só pode ser mergeado se seu Event Graph não conflitar com outro agente | Merge Orchestrator valida antes de integrar |
| 3 | Rollback sempre pode ser feito por arquivo | Snapshot Engine armazena estado por arquivo |
| 4 | Nenhum agente executa `git` diretamente | Agent Runtime Interface esconde completamente o Git |
| 5 | Snapshots temporários são descartáveis | Garbage collection automática após `FINALIZED` ou `DISCARDED` |
| 6 | Locks semânticos têm TTL | Previne locks órfãos quando agentes morrem |
| 7 | Todo evento tem timestamp e agente de origem | Audit trail completo no Event Graph |
| 8 | Merge nunca perde dados por overwrite silencioso | Conflitos sempre bloqueiam e exigem resolução explícita |

---

## 9. Fluxos Detalhados

### 9.1 Inicialização de Workspace

```text
Orchestrator ──► WSM.InitWorkspace(TaskID, AgentID, BaseCommit)
                      │
                      ├──► SemanticLockService.Reserve(DomainOf(Task))
                      │         │
                      │         └──► LOCKED / QUEUED
                      │
                      ├──► GitAdapter.CreateBranch(refs/wsm/{agent}/{ws})
                      │
                      ├──► WorkspaceProvider.Create(strategy=git-clone|overlayfs|bind-copy)
                      │
                      └──► EventStore.Append(WorkspaceCreatedEvent)
                                │
                                └──► Retorna VirtualWorkspace para Orchestrator
```

### 9.2 Operação de Escrita do Agente

```text
Agente ──► FileAPI.WriteFile(path, content)
              │
              ├──► DiffEngine.CalcDiff(path, oldContent, newContent)
              │
              ├──► SnapshotEngine.SaveFileState(path, newContent)
              │
              ├──► EventStore.Append(OperationEvent{
              │         Type: modify,
              │         Target: path,
              │         DiffHash: hash(diff),
              │         ...
              │      })
              │
              └──► Atualiza grafo de operações
```

### 9.3 Tentativa de Merge

```text
Agente/Orch ──► MergeService.RequestMerge(WorkspaceID)
                     │
                     ├──► GitAdapter.FetchLatestMain()
                     │
                     ├──► MergeOrchestrator.Analyze(EventGraph, WorkspaceID)
                     │         │
                     │         ├──► Sem conflitos ──► AutoMerge ──► INTEGRATED
                     │         │
                     │         └──► Com conflitos ──► GenerateReport
                     │                                    │
                     │                                    ├──► Enfileira para resolução humana
                     │                                    ├──► Ou cria Task de resolução para agente
                     │                                    └──► Ou tenta AutoResolution se permitido
                     │
                     └──► Se INTEGRATED → SnapshotEngine.PromoteToPersistent()
```

---

## 10. Segurança e Isolamento

- **Filesystem**: cada workspace roda em um diretório separado. Sem acesso cruzado.
- **Processo**: workspaces podem rodar em containers (Docker/gVisor) para isolamento de processo.
- **Rede**: conforme ADR-0004, rede bloqueada por padrão.
- **Segredos**: nunca persistidos no workspace. Injetados via variáveis de ambiente temporárias.
- **Git**: agentes não têm acesso ao `.git/` diretamente — interagem apenas via WSM API.

---

## 11. Evolução Futura

| Fase | Capacidade | Dependência |
|------|-----------|-------------|
| MVP | `git-clone`, snapshot file-level, merge manual | Módulos task, run, agentsession existentes |
| Fase 2 | `overlayfs` como otimização | Linux-only, detectado automaticamente |
| Fase 3 | Granularidade `symbol` no Event Graph | Code Intelligence (LSP/Tree-sitter) |
| Fase 4 | Auto-resolution de conflitos | Policy Engine + Modelo de LLM dedicado |
| Fase 5 | CRDTs para edição colaborativa em tempo real | Pesquisa ativa, não prioridade |

---

## 12. Glossário

| Termo | Definição |
|-------|-----------|
| **Workspace** | Diretório de trabalho isolado de um agente |
| **Snapshot Temporário** | Estado capturado durante execução, descartável |
| **Snapshot Real** | Commit Git consolidado ao final da tarefa |
| **Event Graph** | DAG de todas as operações de modificação |
| **Semantic Lock** | Reserva de domínio/arquivo para um agente |
| **Merge Orchestrator** | Serviço que coordena integração de workspaces |
| **Base Commit** | Commit do branch principal no momento da criação do workspace |
