# Prompt: Task Canvas do OrchestraOS

## Contexto
O OrchestraOS é um sistema de orquestração de agentes. Uma Task é decomposta em um grafo direcionado acíclico (DAG) de Work Units. Cada Work Unit tem dependências em outras (blocks, requires_artifact, conflicts_with). Agentes executam Work Units em sandboxes isoladas. Esta view mostra UMA task em detalhe completo.

## Propósito da Página
Esta é a tela de DRILL-DOWN para uma task específica. Ela responde: "Como esta task está estruturada? Quem depende de quem? O que cada agente está fazendo? Que artefatos foram produzidos?"

## Estrutura do Layout

### Barra de Header
- Seta de voltar ← para o Command Center
- Título da task + ID (mono)
- Badge de estado da task (running / paused / completed / failed)
- Breadcrumb: OrchestraOS / task-099
- Botões de ação: [Pausar Task] [Cancelar] [Replanejar]

### Área Principal do Canvas (centro, 70% da largura)
Um CANVAS INFINITO mostrando o Task Graph como um grafo direcionado acíclico:

#### Nós de Work Unit
Retângulos (~220px × 90px) contendo:
- Topo: ID da WU em mono (wu-001) + badge de estado (ponto pequeno + label)
- Meio: Título da work unit (1-2 linhas máximo)
- Base: Avatar do agente (círculo 24px) + nome do agente OU "Não atribuído"

Cor da borda do nó = cor do estado:
- Cinza (#5C5C66): idle / pending
- Azul (#007AFF): running
- Verde (#00CC66): completed
- Vermelho (#FF3B30): failed
- Roxo (#9047FF): blocked (aguardando dependência)
- Âmbar (#FF9500): paused / aguardando aprovação

#### Arestas de Dependência
Linhas curvas bezier conectando os nós:
- Linha sólida: `blocks` (WU A deve completar antes de B começar)
- Linha tracejada: `requires_artifact` (B precisa de artefato de A)
- Linha pontilhada: `informs` (A notifica B, sem bloqueio rígido)
- Linha vermelha: `conflicts_with` (potencial conflito, precisa de revisão)

As arestas têm pontas de seta mostrando direção. Hover na aresta mostra tooltip com tipo de dependência.

#### Nós de Artefato
Nós menores (~140px × 50px) conectados às WUs que os produziram:
- Ícone representando tipo de artefato (container, schema, diff, test_report)
- Nome do artefato
- Conectado via aresta "produced_by" (linha fina)

#### Avatares de Agente
Avatares circulares (28px) posicionados SOBRE ou PRÓXIMO às WUs em execução:
- Anel pulsante quando o agente está executando
- Conectado à WU por linha tracejada sutil
- Hover mostra nome do agente + status da sessão

#### Caminho Crítico
O caminho de dependência mais longo é destacado sutilmente (arestas levemente mais grossas, nós com brilho sutil) para que o usuário veja o que bloqueia a conclusão geral.

### Painel de Detalhes (sidebar direita, 30% da largura)
Clicar em uma Work Unit abre este painel com abas:

**Aba: Visão Geral**
- Descrição do objetivo
- Critérios de aceite (checklist)
- Caminhos owned (arquivos/módulos)
- Agente atribuído + capabilities
- Duração estimada vs real

**Aba: Checkpoints**
Timeline vertical dos checkpoints:
- Número do checkpoint + timestamp
- Texto de resumo
- Arquivos lidos / modificados
- Decisões tomadas
- Riscos identificados
- Sugestão de próximo objetivo

**Aba: Eventos**
Stream de eventos filtrado apenas para esta WU:
- agent.checkpoint_reached
- tool.requested / tool.approved / tool.executed
- orchestrator.intervention
- Todos com timestamps e payloads

**Aba: Artefatos**
Lista de artefatos produzidos:
- Ícone de tipo + nome
- Criado em
- Botão [Preview] para artefatos visualizáveis
- Botão [Download] para arquivos

**Aba: Terminal**
Visão ao vivo do terminal da sandbox do agente (se running):
- Saída scrollável
- Prompt de comando na base
- Auto-scroll para a saída mais nova

### Painel Inferior (colapsável)
**Input de Intervenção**
Quando uma WU tem uma requisição de ferramenta pendente:
- Mostra detalhes da requisição: nome da ferramenta, parâmetros de input, avaliação de risco
- Contexto: por que o agente precisa disso, o que será afetado
- Botões [Aprovar] [Rejeitar] [Modificar] [Adicionar Contexto]
- Área de texto para feedback humano

## Níveis de Zoom
- **Zoom 100%**: WUs individuais totalmente legíveis, bom para 4-8 WUs
- **Zoom 75%**: Visão compacta, bom para 8-15 WUs
- **Zoom 50%**: Visão de pássaro, mostra estrutura inteira da task, WUs viram quadrados coloridos
- **Ajustar à Tela**: Auto-zoom para mostrar todas as WUs

## Direção Visual
- Canvas escuro (fundo #1A1A1F)
- Nós flutuam no canvas com sombra sutil (sem sombras pesadas)
- Pontos de grid (muito sutis, #2A2A36) como referência espacial — NÃO é um grid técnico tipo blueprint
- Pan: click-drag no canvas vazio
- Zoom: roda do mouse ou pinch
- Mini-map no canto inferior direito mostrando retângulo do viewport
- Tipografia: Inter para labels, Inter Mono para IDs
- Sem elementos sci-fi, sem neon, sem overlays tipo HUD

## Restrições Chave
- O DAG é o ELEMENTO PRINCIPAL — todo o resto é secundário
- Dependências DEVEM ser visíveis — este é o core value proposition
- WUs bloqueadas devem mostrar CLARAMENTE POR QUE estão bloqueadas (qual WU upstream)
- WUs em execução devem parecer VIVAS (agente pulsando, arestas animadas)
- WUs falhadas devem chamar atenção sem ser alarmantes
- Nós de artefato não devem poluir — colapsar quando zoomed out
