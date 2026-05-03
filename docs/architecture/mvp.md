# MVP Local-First

## Escopo

O MVP deve validar o fluxo minimo de orquestracao usando servidor local, CLI fina, Codex/CLI em sandbox, worktrees e GitHub.

O desenho deve ser compativel com execucao futura em servidor, mas a primeira entrega pode rodar localmente.

## Premissas

- Deployment inicial: local.
- Deployment alvo: servidor.
- Runtime inicial de agente: Codex/CLI.
- Paralelismo inicial: 2 a 5 agentes.
- Interface inicial: scripts de bootstrap e CLI fina.
- Superficie externa principal: GitHub.
- Conectores de chat: fora do caminho critico do MVP.
- Fonte de verdade: repositorio.
- Autonomia inicial: Nivel 2.

## Fluxo Principal

1. Pedido entra pela CLI ou GitHub.
2. Orchestrator cria task e registra evento inicial.
3. Orchestrator normaliza intencao, risco, escopo e criterios de aceite.
4. Task Graph Planner quebra o trabalho em DAG de work units.
5. Prompt Composer monta SystemPrompt e TaskPrompt por work unit.
6. Policy Engine classifica risco e permissoes.
7. Scheduler agenda execucao respeitando dependencias, ownership e limite de paralelismo.
8. Sandbox Manager cria branch, worktree e container.
9. Agent Runtime inicia Codex/CLI com contexto controlado.
10. Agente executa loops com checkpoints, ledger persistente, eventos e tool requests.
11. Orchestrator acompanha live view, aprova, nega, pausa ou envia mensagens mediadas.
12. Agente produz diff, validacoes e resumo.
13. Orchestrator coleta evidencias, exige review, decide merge e publica status na CLI/GitHub quando aprovado.

## Entregaveis do MVP

- Criar task manualmente ou via comando simples.
- Decompor task em Task Graph aciclico.
- Montar PromptSnapshot por work unit.
- Manter Agent Task Ledger por work unit.
- Criar worktree e branch por task.
- Iniciar ate 2 agentes em paralelo no primeiro corte, evoluindo para 5.
- Receber heartbeat e eventos estruturados por WebSocket.
- Persistir tasks, runs e eventos em Postgres.
- Aplicar matriz simples de aprovacao de ferramentas.
- Coletar diff e logs ao fim da execucao.
- Publicar status resumido na CLI e em GitHub Issue/PR quando aplicavel.
- Criar ou atualizar issue/PR no GitHub quando aprovado.

## Fora do MVP

- Painel web completo.
- Aplicativo desktop.
- Kubernetes.
- Sandboxing com Firecracker.
- Protocolo A2A completo.
- Marketplace de agentes.
- Sistema de memoria recursiva e memoria vetorial compartilhada.
- Autonomia nivel 4 ou 5.
- Execucao distribuida em multiplos servidores.

## Criterios de Aceite

O MVP estara validado quando:

- Uma task puder ser criada pela CLI e decomposta em work units.
- Uma work unit puder ser executada em worktree isolado por um agente Codex/CLI.
- O Orchestrator conseguir montar prompts e registrar PromptSnapshot.
- O Agent Task Ledger for atualizado em checkpoints.
- O Orchestrator conseguir pausar ou negar uma ferramenta solicitada.
- Eventos principais ficarem persistidos e consultaveis.
- O resultado tiver diff, validacao e resumo.
- CLI/GitHub receberem status final com evidencias.
- GitHub receber branch, issue ou PR conforme politica da task.

## Riscos Restantes

- Docker reduz risco, mas nao e sandbox forte contra codigo malicioso.
- Codex/CLI pode ter comportamento dificil de padronizar sem wrapper proprio.
- WebSocket sozinho nao garante durabilidade; eventos precisam ser persistidos.
- Chat nao pode virar fonte de verdade das decisoes.
- Paralelismo entre agentes pode gerar conflitos se duas tasks editarem o mesmo modulo.
