# 0029. Workspace Manager Nativo para Paralelização de Agentes

**Status:** Proposed  
**Data:** 2026-05-18

---

## 1. Contexto

O OrchestraOS foi projetado para ser operado por múltiplos Agentes de IA trabalhando em paralelo. A premissa central do sistema é transformar intenção humana em execução distribuída, com orquestração, auditoria e isolamento.

No entanto, o versionamento de código — fundamental para rastreabilidade e integração — não foi projetado para este cenário. O Git, ferramenta padrão do projeto, assume **um usuário humano, um contexto de trabalho, uma branch ativa por vez**.

### 1.1 Falhas das Soluções Existentes

Testamos e analisamos três abordagens antes desta decisão:

| Abordagem | Resultado | Razão da Falha |
|-----------|-----------|----------------|
| **Git worktrees** | Conflitos de refs, arquivos "sumindo" | Worktrees compartilham o mesmo `.git/` e namespace de branches. Se dois agentes operam na mesma branch, competem pelo mesmo HEAD |
| **Git clones independentes** | Coordenação impossível | Cada clone é um universo isolado. Não há awareness entre eles. Merge vira caixa de surpresa no final |
| **Branches separadas + worktrees** | Parcial, mas frágil | Ainda compartilham `.git/`. Operações destrutivas (`reset --hard`, `clean -fd`) afetam todos os worktrees |

### 1.2 Problemas Raiz

O problema não é falta de diretórios separados — é falta de **coordenação ativa de estado** entre processos paralelos sobre um repositório compartilhado. Especificamente:

1. **Isolamento Insuficiente**: Git não tem o conceito de "workspace de processo". Refs são globais por repositório.
2. **Rastreabilidade Operacional**: Git rastreia commits, não *intenção*. Não sabemos por que um arquivo foi modificado ou se duas modificações são semanticamente conflitantes.
3. **Rollback Granular**: Git trabalha em commits (estado global). Não é possível reverter apenas um arquivo afetado por conflito sem afetar o restante do workspace.
4. **Retomada de Contexto**: Um agente interrompido precisa re-ler o projeto do zero ou depender de conversa solta como memória.

### 1.3 Necessidades do OrchestraOS

Com base na evolução do projeto e nas necessidades de MAS (Massive Agents System), documentadas em `docs/architecture/massive-agents-system.md`, identificamos quatro requisitos inegociáveis:

| Requisito | Justificativa |
|-----------|---------------|
| Histórico temporário por agente | Rollback, descarte, retomada de tarefa |
| Interrupção e retomada | Agentes podem ser pausados por humanos ou pelo orquestrador |
| Merge orquestrado com explicação | Conflitos devem ser detectados previamente e explicados, não surpresas |
| Snapshots por arquivo | Rollback deve afetar apenas arquivos conflitantes, não workspaces inteiros |
| Grafo de operações | Toda modificação é um evento; o sistema mantém dependências entre eventos |

Nenhuma ferramenta existente (Git, Mercurial, Perforce, Plastic SCM) atende todos estes requisitos simultaneamente sem customização pesada. Perforce e Plastic SCM resolvem isolamento, mas não fornecem grafo de operações ou retomada com contexto de agente.

---

## 2. Decisão

### 2.1 Decisão Principal

Construiremos o **Workspace Manager (WSM)** como componente nativo do OrchestraOS. O WSM é uma camada de orquestração que opera **sobre** o Git, não o substituindo.

> **Regra absoluta:** Agentes de IA nunca executam comandos `git` diretamente. Toda interação com versionamento passa pelo WSM.

### 2.2 Componentes do WSM

O WSM é composto por quatro subsistemas:

#### 2.2.1 Virtual Workspace Service

Cria e gerencia ambientes de trabalho isolados por agente.

- Cada workspace tem **branch exclusiva** (`refs/wsm/{agent-id}/{workspace-id}`)
- Isolamento de filesystem via `git-clone` (padrão), `overlayfs` (otimização Linux) ou `bind-copy`
- O workspace é a única interface que o agente vê — ele não sabe que Git existe

#### 2.2.2 Snapshot Engine

Gerencia dois tipos de snapshot:

| Tipo | Quando Criado | Persistência | Granularidade |
|------|---------------|--------------|---------------|
| **Transient Snapshot** | Durante execução (auto/manual) | Temporário — descartado após finalização | Por arquivo |
| **Persistent Snapshot** | Ao finalizar tarefa com sucesso | Commit Git permanente | Commit completo |

- Transient Snapshots permitem rollback por arquivo, retomada de contexto e inspeção de histórico.
- Persistent Snapshots tornam-se commits reais no repositório, integráveis via merge.

#### 2.2.3 Operation Event Graph

Um **DAG (Directed Acyclic Graph)** onde:

- **Nós** são eventos de modificação de arquivo (`OperationEvent`)
- **Arestas** são dependências de estado entre eventos
- **Componentes conectados** revelam colisões potenciais

O grafo opera em níveis de granularidade configuráveis:

| Nível | Unidade | Requisito |
|-------|---------|-----------|
| `file` | Arquivo inteiro | MVP — imediato |
| `line-range` | Range de linhas | Fase 2 — requer parser |
| `symbol` | Função, struct, interface | Fase 3 — requer Code Intelligence |
| `architecture` | Módulo/pacote | Sempre ativo para Semantic Locking |

#### 2.2.4 Merge Orchestrator

Serviço que coordena a integração de snapshots no branch principal.

**Fases:**
1. **Análise** — consulta Event Graph, identifica colisões
2. **Tentativa Automática** — merge sem conflito é aplicado automaticamente
3. **Bloqueio com Explicação** — conflitos geram relatório estruturado com contexto
4. **Resolução** — humana, orquestrada (nova task) ou automatizada (se policy permitir)
5. **Integração** — commit final no branch principal

**Semantic Lock Service** complementa o merge: reserva domínios/arquivos para agentes antes do início, reduzindo a incidência de conflitos.

### 2.3 Estratégia de Isolamento de Filesystem

| Estratégia | Mecanismo | Quando Usar |
|------------|-----------|-------------|
| `git-clone` | Clone completo do repositório | Padrão universal — funciona em qualquer SO |
| `overlayfs` | Overlay filesystem do Linux | Otimização automática quando detectado Linux |
| `bind-copy` | Cópia seletiva + diretório de trabalho | Fallback para cenários específicos |

A estratégia é selecionada automaticamente pelo WSM com base no ambiente. O agente não escolhe e não precisa saber qual está em uso.

### 2.4 Namespace Git

O WSM usa refs privadas para não poluir o namespace de branches do usuário:

```
refs/wsm/{agent-id}/
  {workspace-id}/        ← branch do workspace
  snapshots/
    transient-{seq}      ← refs para snapshots temporários
```

Transient Snapshots são armazenados como objetos Git (tree + blob) em refs especiais, mas não fazem parte do histórico principal. São garbage coletados após descarte do workspace.

### 2.5 Interface para Agentes

O agente interage com o WSM através de uma API que esconde completamente o Git:

```go
// FileAPI — operações de filesystem
type FileAPI interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, content []byte) error
    DeleteFile(path string) error
    ListDirectory(path string) ([]FileInfo, error)
}

// WorkspaceAPI — gestão do workspace
type WorkspaceAPI interface {
    GetStatus() WorkspaceStatus
    GetDiff() DiffSummary
    CreateCheckpoint() (SnapshotID, error)
    RequestMerge() (MergeResult, error)
}

// ContextAPI — contexto para retomada
type ContextAPI interface {
    GetTask() Task
    GetHistory() []OperationEvent
    GetConflicts() []ConflictReport
}
```

---

## 3. Consequências

### 3.1 Positivas

- **Paralelização real**: agentes trabalham simultaneamente sem corromper o estado uns dos outros
- **Rollback granular**: apenas arquivos conflitantes são revertidos; trabalho não-afetado é preservado
- **Retomada com contexto**: agente interrompido retoma do último snapshot com histórico completo, sem re-leitura
- **Merge previsível**: conflitos são detectados antes de serem commitados no branch principal
- **Audit trail completo**: Event Graph registra quem fez o quê, quando e por dependência de quê
- **Git preservado**: não abandonamos o ecossistema Git — branches, PRs, CI/CD continuam funcionando
- **Evolução para MAS**: a arquitetura suporta de 2 agentes (MVP) a dezenas (futuro) sem mudança estrutural

### 3.2 Negativas

- **Complexidade nova**: o WSM é um componente substancial — não uma biblioteca de 200 linhas
- **Overhead de storage**: Transient Snapshots ocupam espaço em disco (mitigável com compressão e GC)
- **Latência de I/O**: cada operação de escrita gera evento + snapshot potencial — mais I/O que editar diretamente
- **Curva de aprendizado**: desenvolvedores humanos precisam entender o WSM além do Git
- **Dependência do WSM**: se o WSM falha, todos os agentes ficam sem capacidade de versionamento
- **Garbage collection**: refs e objetos Git temporários precisam de rotina de limpeza para não incharem o repositório

### 3.3 Riscos

| Risco | Mitigação |
|-------|-----------|
| Snapshot explosion (muitos transient snapshots) | Limite de 50 snapshots por workspace + compressão + GC |
| WSM vira gargalo | WSM é stateless por workspace; pode escalar horizontalmente |
| Merge Orchestrator incorreto | Merge sempre bloqueia em dúvida; nunca faz merge silencioso |
| Locks semânticos órfãos | TTL automático em todos os locks (padrão: 4h) |
| Perda de transient snapshots | Snapshots replicados em cache local + event store |

---

## 4. Plano de Implementação

### 4.1 Fase 1: Fundação (MVP — WSM v0.1)

**Objetivo:** Agentes conseguem trabalhar em workspaces isolados com snapshot e merge manual.

| Entrega | Descrição |
|---------|-----------|
| `VirtualWorkspace` | Struct, criação, ciclo de vida básico (INIT → ACTIVE → FINALIZED → DISCARDED) |
| `git-clone` strategy | Clone completo por workspace, branch exclusiva |
| `SnapshotEngine` | Transient Snapshots manuais (`CreateCheckpoint`) e auto (a cada N operações) |
| `FileAPI` | ReadFile, WriteFile, DeleteFile com geração de `OperationEvent` |
| `MergeService` básico | Cria Persistent Snapshot (commit Git) + abre PR no GitHub |

**Validação:** 2 agentes editam arquivos diferentes em workspaces separados. Ambos são mergeados via PR sem conflito.

### 4.2 Fase 2: Event Graph e Detecção de Conflito

**Objetivo:** O sistema detecta conflitos antes do merge.

| Entrega | Descrição |
|---------|-----------|
| `OperationEvent` + DAG | Toda escrita gera evento; grafo em memória por workspace |
| `ConflictDetector` | Compara Event Graphs de dois workspaces, identifica colisões de arquivo |
| `MergeOrchestrator` | Estados: PENDING → ANALYZING → CLEAN_MERGE / CONFLICT_DETECTED |
| Relatório de conflito | JSON estruturado com arquivos, agentes, linhas e sugestão de resolução |

**Validação:** Agente A edita `service.go`. Agente B edita `service.go` em paralelo. O segundo merge é bloqueado com relatório explicativo.

### 4.3 Fase 3: Retomada e Rollback

**Objetivo:** Agentes podem ser suspensos, retomados e fazer rollback por arquivo.

| Entrega | Descrição |
|---------|-----------|
| `Suspend` / `Resume` | Persiste estado do workspace + Event Graph; restaura na retomada |
| `RollbackFile` | Reverte um único arquivo para snapshot anterior |
| `GetHistory` | Agente consulta todas as operações que fez no workspace |
| `overlayfs` strategy | Otimização automática para Linux |

**Validação:** Agente é suspenso no meio da tarefa. Retoma 2h depois com contexto intacto. Faz rollback de um arquivo conflitante sem perder o restante.

### 4.4 Fase 4: Merge Orquestrado e Semantic Locks

**Objetivo:** Merge pode ser resolvido automaticamente ou delegado.

| Entrega | Descrição |
|---------|-----------|
| `SemanticLockService` | Reserva de domínio/arquivo antes do início da task |
| `AutoResolution` | Tentativa de merge 3-way para conflitos simples (policy-gated) |
| `ResolutionTask` | Cria nova task de resolução de conflito e assigna a agente/humano |
| `Integrated` state | Merge final commitado no `main` com trilha de auditoria |

**Validação:** Dois workspaces conflitantes são detectados. Um é mergeado automaticamente. O outro gera task de resolução.

### 4.5 Fase 5: Granularidade Fina e MAS

**Objetivo:** Suporte a dezenas de agentes com mínimo de conflitos.

| Entrega | Descrição |
|---------|-----------|
| Code Intelligence integration | Mapeamento de símbolos via LSP/Tree-sitter |
| `symbol` granularity | Event Graph rastreia funções/structs, não apenas arquivos |
| `line-range` granularity | Merge por range de linhas |
| Garbage collection | Rotina de limpeza de transient snapshots e refs órfãs |

---

## 5. Alternativas Consideradas

### 5.1 Perforce (Helix Core)

**Prós:** workspaces isolados nativos, locking forte, usado em indústria de games.  
**Contras:** Licenciamento por usuário, curva de aprendizado, ecossistema fechado, não resolve grafo de operações ou retomada de contexto.  
**Decisão:** Rejeitado. Custo e vendor lock-in incompatíveis com o espírito open-source do OrchestraOS.

### 5.2 Plastic SCM / Unity Version Control

**Prós:** branches isoladas por workspace, merge visual excelente.  
**Contras:** Pagamento por usuário, integração com GitHub limitada, não resolve Event Graph.  
**Decisão:** Rejeitado. Mesmos problemas de vendor lock-in.

### 5.3 Git + Hooks de Prevenção

**Ideia:** usar git hooks para bloquear pushes conflitantes.  
**Contras:** Hooks rodam no momento do push — conflito já aconteceu. Não resolve isolamento durante execução.  
**Decisão:** Rejeitado. Trata sintoma, não causa.

### 5.4 Cada Agente em VM/Firecracker com Repo Completo

**Ideia:** isolar agentes em micro-VMs, cada uma com seu próprio repositório Git independente.  
**Prós:** Isolamento total de filesystem, processo e rede.  
**Contras:** Overhead de VM, sincronização entre repos é manual, não resolve coordenação de merge.  
**Decisão:** Parcialmente aceito. VMs são estratégia de sandbox (ADR-0004), mas não substituem o WSM. O WSM orquestra os workspaces, independente de estarem em VM, container ou host.

### 5.5 Usar Git Como Está (Status Quo)

**Ideia:** continuar com worktrees e branches, apenas documentar melhor.  
**Contras:** Problemas documentados na seção 1.1 são inerentes ao modelo do Git. Não há documentação que resolva competição por refs compartilhadas.  
**Decisão:** Rejeitado. O status quo impede o paralelismo que é premissa do OrchestraOS.

---

## 6. Exemplos

### 6.1 Criação de Workspace

```go
wsm := NewWorkspaceManager(gitAdapter, eventStore)

ws, err := wsm.InitWorkspace(InitRequest{
    TaskID:     "task-uuid",
    AgentID:    "agent-codex-1",
    BaseCommit: "abc123", // main HEAD
    Strategy:   "git-clone",
})

// ws.BranchName == "refs/wsm/agent-codex-1/{workspace-uuid}"
// ws.RootPath == "~/.local/share/orchestraos/wsm/{workspace-uuid}"
```

### 6.2 Operação de Escrita e Snapshot

```go
// Agente edita arquivo
fileAPI := ws.FileAPI()
fileAPI.WriteFile("service.go", newContent)

// WSM automaticamente:
// 1. Calcula diff
// 2. Cria OperationEvent no grafo
// 3. Atualiza Transient Snapshot se necessário
```

### 6.3 Bloqueio de Merge com Explicação

```json
{
  "merge_id": "merge-uuid",
  "status": "CONFLICT_DETECTED",
  "workspace_a": "ws-agent-1",
  "workspace_b": "ws-agent-2",
  "base_commit": "abc123",
  "conflicts": [
    {
      "file": "internal/modules/auth/service.go",
      "type": "content",
      "agents": ["agent-1", "agent-2"],
      "agent_1_changes": "linhas 45-60: adicionou método ValidateToken()",
      "agent_2_changes": "linhas 50-70: refatorou autenticação para usar OAuth",
      "suggested_resolution": "combinar ValidateToken() com fluxo OAuth",
      "rollback_available": [
        {
          "target": "agent-2",
          "file": "internal/modules/auth/service.go",
          "snapshot_id": "snap-uuid-7",
          "description": "Estado antes da refatoração OAuth"
        }
      ]
    }
  ],
  "resolution_options": [
    "manual_human_review",
    "create_resolution_task",
    "rollback_and_retry"
  ]
}
```

### 6.4 Retomada de Workspace Suspenso

```go
ws, err := wsm.ResumeWorkspace("workspace-uuid")

// Agente recebe:
// - Estado atual do filesystem
// - Lista de OperationEvents que ele mesmo gerou
// - Diffs pendentes
// - Task original
// Nenhuma re-leitura do projeto é necessária
```

---

## 7. Regras para Agentes de IA (Resumo)

Quando um agente de IA operar sob o WSM:

1. **NUNCA** execute `git add`, `git commit`, `git push`, `git merge` ou qualquer comando git diretamente.
2. **SEMPRE** use a `FileAPI` do WSM para ler, escrever ou deletar arquivos.
3. **SEMPRE** crie um checkpoint (`CreateCheckpoint`) antes de mudar de foco ou solicitar ferramenta de alto risco.
4. **NUNCA** assuma que outro agente não está editando o mesmo arquivo — o WSM detectará, mas a prevenção é melhor que a correção.
5. **SEMPRE** consulte `GetConflicts()` antes de iniciar modificação em arquivo que pode ser compartilhado.
6. **NUNCA** edite fora do `RootPath` do workspace atribuído.
7. **SEMPRE** finalize o workspace via `RequestMerge()` — nunca copie arquivos manualmente para fora.

---

## Apêndice A: Checklist de Validação

Antes de qualquer PR que toque no WSM:

- [ ] Workspace cria e descarta sem deixar refs órfãs
- [ ] Dois workspaces simultâneos não compartilham estado
- [ ] Transient Snapshot pode ser restaurado integralmente
- [ ] Rollback por arquivo preserva arquivos não-afetados
- [ ] Merge com conflito gera relatório estruturado
- [ ] Workspace suspenso retoma com contexto intacto
- [ ] `go test ./...` passa
- [ ] `./scripts/verify-contracts.sh` passa
- [ ] `./scripts/lint.sh` passa
- [ ] Architecture tests passam (boundaries, module deps)

---

## Apêndice B: Relação com ADRs Existentes

| ADR | Relação |
|-----|---------|
| ADR-0004 | WSM substitui worktree como mecanismo de sandbox; containers continuam válidos para isolamento de processo |
| ADR-0022 | WSM será um módulo vertical em `internal/modules/workspace/` |
| ADR-0023 | Orchestrator Inteligente usará WSM para spawnar e gerenciar workspaces |
| ADR-0028 | WSM segue padrões de nomenclatura — nenhum `helpers.go` ou `utils.go` |
