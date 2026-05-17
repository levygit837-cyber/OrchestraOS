# Visão: Agent Communication Hub (ACH)

> **Status**: Ideia Futura | **Prioridade**: Baixa | **Complexidade**: Alta
> **Data de Documentação**: 2026-05-17
> **Motivo**: Pesquisa técnica concluída. Alta complexidade de implementação. Retorno imediato baixo para o estágio atual do OrchestraOS.

---

## Resumo Executivo

Esta é a documentação de uma **ideia futura** para o OrchestraOS: uma plataforma de comunicação entre agentes AI heterogêneos, inspirada no ACP (Agent Communication Protocol) da Anthropic, porém mais nichada e integrada à arquitetura do OrchestraOS.

> ⚠️ **DECISÃO ARQUITETURAL**: NÃO implementar no momento. O OrchestraOS ainda está em fase de consolidação de sua camada de orquestração básica. A complexidade de integrar múltiplos protocolos de agentes (Windsurf Connect, Kimi ACP, Devin WebSocket, etc.) não justifica o investimento de tokens/dinheiro no estágio atual.

---

## Contexto

### O Que É
Uma camada de comunicação inter-agentes que permite:
- Agentes de diferentes plataformas (Windsurf, Kimi, Devin, etc.) trocarem mensagens
- Um agente delegar subtarefas para outro agente especializado
- Compartilhamento de contexto e estado entre agentes
- Rastreamento de auditoria de decisões multi-agente

### Por Que Foi Pesquisado
Durante o desenvolvimento do OrchestraOS, identificou-se que o sistema poderia se beneficiar de uma abordagem multi-agente onde:
1. O Orchestrator central coordena tarefas
2. Agentes especializados (Windsurf para código, Kimi para análise, Devin para automação) executam subtarefas
3. Cada agente opera em seu sandbox/isolamento nativo
4. Resultados são consolidados pelo Orchestrator

### Por Que Foi Arquivado Como Ideia Futura
- **Custo de implementação alto**: Requer engenharia reversa de protocolos proprietários
- **Mudanças frequentes**: Protocolos de agentes mudam constantemente (ex: Windsurf mudou de porta 3x)
- **Retorno não justifica investimento agora**: O OrchestraOS precisa primeiro consolidar sua camada de runtime único
- **Tokens gastos**: Já foram consumidos ~50K+ tokens em pesquisa reversa sem entregar valor de produto

---

## Aprendizados Técnicos (Pesquisa Reversa)

### 1. WindSurf Language Server

#### Protocolo
- **Nome**: Connect Protocol (gRPC sobre HTTP/1.1)
- **Porta**: 33945 (`--server_port`)
- **Porta LSP (não usar)**: 34567 (`--lsp_port`)
- **Formato**: Binary protobuf + envelope de 5 bytes obrigatório

#### Envelope Connect Protocol
```
[flags: 1 byte = 0x00] [length: 4 bytes big-endian] [protobuf payload]
```

#### Autenticação
- **CSRF Token**: Extraído de `WINDSURF_CSRF_TOKEN` no `/proc/PID/environ`
- **API Key**: Configurada no arquivo `.env` do projeto (ex: `sk-ws-01-...`)
- **Installation ID**: UUID fixo por instalação

#### Ciclo de Vida de Sessão
```
StartCascade → InitializeCascadePanelState → SendUserCascadeMessage → StreamCascadeReactiveUpdates → GetCascadeTrajectory
```

#### Modelos Disponíveis (Plano Free)
| Modelo | Status | Multiplicador |
|--------|--------|---------------|
| `swe-1-6-slow` | ✅ ENABLED | 1.0x |
| `kimi-k2-6` | ❌ DISABLED | - |
| `swe-1-6-fast` | ❌ DISABLED | - |
| `MODEL_ALIAS_CASCADE_BASE` | ✅ | 0.5x |
| `MODEL_ALIAS_VISTA` | ✅ | 1.0x |
| `MODEL_ALIAS_SHAMU` | ✅ | 0.25x |

#### Endpoints Principais (141 métodos mapeados)
- `StartCascade` - Cria sessão (unary, JSON/proto)
- `SendUserCascadeMessage` - Envia mensagem (unary, JSON/proto)
- `StreamCascadeReactiveUpdates` - Streaming de estado (requer envelope)
- `GetCascadeTrajectory` - Histórico completo (unary, JSON/proto)
- `Heartbeat` - Verificação de conexão

### 2. Kimi Code

#### Protocolo
- **Nome**: ACP (Agent Communication Protocol) / JSON-RPC 2.0
- **Transporte**: stdio (stdin/stdout)
- **Binário**: `kimi --wire --no-thinking`
- **Porta**: Não usa porta de rede (comunicação local via pipe)

#### Características
- Totalmente isolado do sistema de arquivos
- Comunicação apenas via JSON-RPC sobre stdio
- Requer adaptador dedicado para converter entre JSON-RPC e o bus de mensagens

### 3. Devin

#### Protocolo
- **Transporte**: WebSocket
- **URL**: `wss://app.devin.ai/api/acp/live`
- **Autenticação**: Token via WebSocket handshake

---

## Arquitetura Ideal (Futura)

```
┌─────────────────────────────────────────────────────────────┐
│                    Agent Communication Hub                   │
│                         (ACH)                                │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   WindSurf  │  │    Kimi     │  │       Devin         │  │
│  │   Adapter   │  │   Adapter   │  │      Adapter        │  │
│  │             │  │             │  │                     │  │
│  │ Connect     │  │  JSON-RPC   │  │    WebSocket        │  │
│  │ Protocol    │  │   over stdio│  │    Client           │  │
│  │ Port 33945  │  │             │  │                     │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘  │
│         │                │                     │             │
│         └────────────────┼─────────────────────┘             │
│                          ▼                                   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              Message Bus (SQLite + SSE)               │   │
│  │                                                      │   │
│  │  • Persistência de mensagens                         │   │
│  │  • Streaming via Server-Sent Events                  │   │
│  │  • Roteamento entre agentes                          │   │
│  │  • Auditoria e trilha de decisões                    │   │
│  └──────────────────────────────────────────────────────┘   │
│                          │                                   │
│                          ▼                                   │
│  ┌──────────────────────────────────────────────────────┐   │
│  │              OrchestratorService                      │   │
│  │         (Coordenação Central do OrchestraOS)          │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### Componentes do ACH (Futuro)

#### 1. Message Bus
- **Tecnologia**: SQLite (local) + SSE (streaming)
- **Função**: Persistir e rotear mensagens entre agentes
- **Schema**: Tabela `messages` com campos: id, from_agent, to_agent, type, payload, timestamp, session_id

#### 2. WindSurf Adapter
- Conecta ao Language Server na porta 33945
- Traduz chamadas do ACH para Connect Protocol
- Mantém sessões persistentes (cascadeId)
- Faz streaming de atualizações via `StreamCascadeReactiveUpdates`

#### 3. Kimi Adapter
- Spawna processo `kimi --wire --no-thinking`
- Traduz JSON-RPC 2.0 para formato do ACH
- Gerencia ciclo de vida do processo Kimi

#### 4. Devin Adapter
- Conecta via WebSocket a `wss://app.devin.ai/api/acp/live`
- Traduz mensagens WebSocket para formato do ACH
- Gerencia reconexões automáticas

#### 5. Session Manager
- Rastreia sessões ativas de cada agente
- Permite retomar conversas anteriores
- Isola contexto entre diferentes sessões

---

## Quando Implementar (Critérios)

### Pré-condições
1. ✅ OrchestraOS tem runtime estável de execução de tasks
2. ✅ Sistema de eventos e state machine funcionando end-to-end
3. ✅ CLI permite criar tasks, decompor em work units, e executar
4. ✅ Integração com GitHub funciona (issues, PRs, reviews)
5. ❌ **NÃO ATINGIDO**: Orquestração multi-agente é necessária para casos de uso reais

### Gatilhos para Priorização
- Quando uma task requer expertise em múltiplos domínios (código + design + infra)
- Quando o usuário precisa de diferentes "personalidades" de AI (analista vs implementador)
- Quando o custo de um único agente é maior que a soma de agentes especializados
- Quando houver demanda real de delegação entre agentes

### Estimativa de Esforço
- **Pesquisa/POC**: ~2-3 dias (já feita parcialmente)
- **Implementação MVP**: ~2-3 semanas
- **Produção estável**: ~1-2 meses
- **Manutenção contínua**: Alto (protocolos mudam frequentemente)

---

## Documentação Técnica Detalhada

Para referência futura, os detalhes técnicos completos estão em:
- `/docs/analysis/windsurf-streaming-success.md` - Protocolo Connect decodificado
- `/docs/analysis/windsurf-error-analysis.md` - Análise de erros e permissões
- `/docs/analysis/windsurf-model-test-results.md` - Modelos testados e resultados
- `/tmp/windsurf-real-interaction.js` - Script completo de demonstração

---

## Lições Aprendidas

1. **Não reinvente protocolos**: O Connect Protocol do Windsurf é complexo e proprietário. Sempre que possível, use APIs públicas/documentadas.

2. **Autenticação é frágil**: CSRF tokens mudam a cada restart do Language Server. API keys podem expirar. Sempre implemente fallback e redetecção.

3. **Planos limitam opções**: O plano Free do Windsurf não permite acesso a modelos premium (Kimi K2.6, SWE-1.6 Fast). Considere isso no design.

4. **Streaming é melhor que polling**: Uma vez decodificado o envelope, o streaming é confiável e eficiente. Não use polling para atualizações.

5. **Sessões são persistentes**: O cascadeId pode ser reutilizado para múltiplas mensagens. Não crie uma sessão por mensagem.

---

## Próximos Passos (Se/Voltar a Implementar)

1. **Extrair schema protobuf completo** do binary do Language Server
2. **Implementar Message Bus** com SQLite + SSE
3. **Criar WindSurf Adapter** com streaming real
4. **Adicionar Kimi Adapter** via JSON-RPC stdio
5. **Integrar com OrchestratorService** do OrchestraOS

---

## Nota Final

> "A melhor arquitetura é aquela que você não precisa construir agora."

Esta ideia foi arquivada não por falta de mérito técnico, mas por **timing estratégico**. O OrchestraOS precisa primeiro ser excelente em orquestrar UM agente antes de orquestrar MÚLTIPLOS agentes.

**Reavalie esta decisão quando:**
- O OrchestraOS estiver em produção estável
- Houver demanda real de multi-agente
- Os protocolos de agentes estiverem mais maduros/standardizados
- O time tiver budget dedicado para integração

---

*Documento criado por decisão explícita após consumo significativo de tokens em pesquisa sem entrega de valor de produto imediato.*
