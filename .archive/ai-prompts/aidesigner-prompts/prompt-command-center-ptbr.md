# Prompt: Command Center do OrchestraOS

## Contexto
O OrchestraOS é um sistema de orquestração de agentes. Múltiplos agentes de IA executam unidades de trabalho (Work Units) em um grafo direcionado acíclico (DAG). Humanos supervisionam, aprovam requisições de ferramentas e intervêm quando necessário. Isso NÃO é um chat, NÃO é uma IDE, NÃO é um quadro kanban.

## Propósito da Página
Esta é a TELA PRINCIPAL — a primeira tela que o usuário vê ao abrir o OrchestraOS. Ela deve responder: "O que precisa da minha atenção agora? O que está rodando? Qual a saúde do sistema?"

## Estrutura do Layout

### Barra de Navegação Superior (48px de altura)
- Esquerda: Logo do OrchestraOS (minimalista, baseado em texto) + badge de ambiente (ex: "local" / "prod")
- Centro: Busca global / gatilho do command palette
- Direita: Indicador de status do sistema (healthy/degraded) + avatar do usuário

### Área de Conteúdo Principal (abaixo da navegação)

#### Seção 1: Fila de Ações Pendentes (scroll horizontal, prioridade máxima)
Uma fileira de cards compactos mostrando itens que precisam de intervenção humana:
- Cards de requisição de ferramenta: ícone, nome da ferramenta (file_write, shell.exec), arquivo/comando alvo, nível de risco (safe/guarded/destructive), nome do agente, dois botões [Aprovar] [Rejeitar]
- Cards de revisão: thumbnail do PR/diff, agente que criou, botão [Revisar]
- Cards de falha: work unit falhada, snippet do erro, botão [Investigar]

Cada card: 280px de largura, fundo escuro de superfície, borda esquerda colorida pelo risco (verde/amarelo/vermelho).

#### Seção 2: Tarefas Ativas (grid de 2 colunas no desktop)
Cards mostrando as tasks em execução no momento:
- Título da task + ID (mono)
- Progresso: X de Y work units completadas
- Mini barra de progresso horizontal
- Avatar do agente atual + nome
- Badge de estado (running / paused / review)
- Próximo checkpoint ou aprovação pendente

Card: 400px de largura, borda sutil, hover revela botão "Abrir Canvas".

#### Seção 3: Saúde do Sistema (uma linha, 4 métricas)
- Agentes Ativos: 3/5 (com ponto pulsante se < 100%)
- Taxa de Sucesso: 94% (mini gráfico sparkline)
- Eventos/min: 12
- Aprovações Pendentes: 2 (clicável, rola para Fila de Ações)

#### Seção 4: Stream de Eventos Recentes (sidebar direita ou painel inferior)
Lista vertical dos últimos eventos:
- Timestamp (mono, 11px)
- Badge de tipo de evento (pequeno, colorido)
- Entidade (task-099, wu-002, codex-builder)
- Descrição
- Auto-scroll, mais novo no topo

### Estados Vazios
- Nenhuma task ativa: "Nenhuma task em execução. Crie uma nova task para começar." + botão prominente "+ Nova Task"
- Nenhuma ação pendente: "Tudo certo. Sistema rodando normalmente." com check verde sutil

## Direção Visual
- Dark mode (fundo #1A1A1F)
- Alta densidade de informação — mostre muitos dados sem bagunçar
- Cards com border-radius de 6px, bordas de 1px sutis
- Acentos de cor funcionais apenas para estados (azul=running, verde=success, vermelho=failure, âmbar=warning, roxo=blocked)
- Tipografia: Inter para UI, Inter Mono para IDs e timestamps
- Sem gradientes, sem glows, sem elementos decorativos
- Profissional, calmo, autoritário — tipo Linear + Vercel + GitHub Actions

## Restrições Chave
- A FILA DE AÇÕES é o elemento principal — intervenções devem ser impossíveis de não notar
- Tasks ativas mostram PROGRESSO, não apenas status
- Saúde do sistema deve ser visível num relance — números, não gráficos complexos
- Tudo clicável leva ao Task Canvas
