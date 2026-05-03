# Canvas do Projeto

Este é o canvas principal do projeto. Ele deve ser mantido em texto para que Codex, agentes futuros e humanos consigam ler, revisar e versionar o contexto sem depender de uma ferramenta visual externa.

## Visão

Construir um Sistema de Orquestração de Agentes capaz de transformar intenção humana em planejamento, execução, validação, documentação e operação contínua.

## Premissa Central

O sistema começa como "minha mente + eu no projeto", com IA auxiliando organização e execução. Com o tempo, a autonomia aumenta por políticas, permissões, observabilidade e limites claros.

## Usuário Inicial

- Fundador/operador solo.
- Pessoa com muitas ideias e necessidade de transformar pensamento em execução organizada.
- Quer que IA gerencie trabalho, memória operacional, decisões e automações.

## Problemas

- Ideias ficam na mente e se perdem antes de virar execução.
- Ferramentas desconectadas dificultam continuidade.
- Agentes podem agir sem contexto suficiente se não houver fonte de verdade.
- Autonomia sem trilha de auditoria cria risco operacional.

## Proposta de Valor

Um sistema operacional de projeto onde agentes entendem contexto, propõem próximos passos, executam tarefas, registram decisões e operam rotinas com supervisão calibrada.

## Componentes Iniciais

- Repositório: fonte de verdade para código, documentação, canvas e decisões.
- Codex: execução técnica, análise de código e manutenção dos arquivos do projeto.
- GitHub: issues, branches, pull requests, revisão, checks e histórico.
- CLI local: primeira interface oficial para operar o MVP.
- Canvas textual: contexto estruturado para humanos e IA.

## Arquitetura Inicial Decidida

- O sistema terá um Orchestrator central como control plane.
- Agentes serão workers isolados por task, inicialmente usando Codex/CLI.
- Cada task terá branch, worktree e sandbox próprios.
- A comunicação agente-orquestrador será feita por WebSocket, com eventos persistidos.
- A comunicação entre agentes será mediada pelo Orchestrator para preservar auditoria, política e isolamento de contexto.
- O MVP começa local-first, mas com desenho compatível com servidor.
- A operação inicial será GitHub-first, usando issues, branches, worktrees, pull requests, reviews e checks.
- A interface inicial do MVP será scripts de bootstrap + CLI fina.
- O paralelismo inicial esperado será de 2 a 5 agentes.
- A autonomia aprovada para o MVP será Nível 2.
- O Orchestrator quebrará tasks em Task Graph acíclico e montará prompts por fragmentos versionados.
- Agentes manterão ledger persistente de progresso por work unit.
- Agentes registrarão checkpoints estruturados em pontos seguros para auditoria e controle de contexto.
- Memória recursiva será uma camada derivada do Event Store, checkpoints, ledger, artefatos e documentos versionados, usada para recuperar contexto sem virar fonte de verdade paralela.

## Componentes Futuros

- Implementação completa do Orchestrator de agentes.
- Sistema de Memória Recursiva com deduplicação, evidências, busca estruturada e embeddings em segundo plano.
- Painel web para projetos, tarefas, agentes e automações.
- Aplicativo desktop para live view local, traces e intervenção em agentes.
- Conectores opcionais de chat, incluindo Slack, quando houver necessidade real.
- Políticas de autonomia por área e nível de risco.
- Conectores com calendário, e-mail, documentos, banco de dados e ferramentas externas.

## Princípios

- Fonte de verdade versionada.
- Autonomia progressiva.
- Tudo importante deve deixar rastro.
- Decisões antes de automações irreversíveis.
- Agentes executam dentro de limites explícitos.

## Riscos

- Automatizar caos em vez de organizar o sistema.
- Dar autonomia antes de existir teste, log e rollback.
- Usar chat ou conversa solta como memória definitiva.
- Misturar ideias, decisões e execução no mesmo canal.
- Criar complexidade técnica antes de validar o fluxo operacional.

## Métricas Iniciais

- Ideias capturadas por semana.
- Ideias convertidas em tarefas pequenas.
- Tarefas concluídas com evidência.
- Decisões registradas em ADR.
- Tempo entre ideia e primeira entrega funcional.
- Incidentes ou ações revertidas por agente.

## Próxima Fronteira

Construir o primeiro MVP: um fluxo local em que uma mensagem entra pela CLI ou GitHub, vira task, e decomposta em work units, recebe prompts montados pelo Orchestrator, cria worktree/sandbox, executa agente Codex/CLI, registra traces e ledger, e retorna com status, diff, validações, evidências e decisão de merge via GitHub/CLI.
