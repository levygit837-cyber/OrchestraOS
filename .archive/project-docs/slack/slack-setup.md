# Configuração Opcional do Slack

Slack não é dependência do MVP. Este documento fica como referência para uma integração futura de chat, caso o projeto precise de captura rápida, notificações ou rotinas por conversa.

O caminho operacional principal do MVP é GitHub-first: CLI local, branches, worktrees, issues, pull requests, reviews e checks.

## Workspace

Sugestão inicial:

- Nome do workspace: usar o nome escolhido do projeto.
- Plano inicial: gratuito enquanto o projeto estiver em descoberta; revisar plano pago quando precisar de histórico maior, workflows avançados, canvas autônomos ou integrações mais fortes.
- Idioma: português, se o projeto for pessoal; inglês, se houver intenção de equipe internacional.

## Canais Opcionais

Estrutura opcional alinhada ao workspace atual:

- `#00-inbox`: entrada geral, capturas rápidas e mensagens ainda sem triagem.
- `#01-diario-dev`: registro diário, decisões pequenas, ideias e contexto de execução.
- `#02-backlog`: features, bugs, melhorias e priorização.
- `#03-decisoes`: decisões importantes, links para ADRs e mudanças de direção.
- `#04-arquitetura`: arquitetura, stack, protocolos, fluxos e tradeoffs técnicos.
- `#05-bugs-debug`: falhas, logs, hipóteses, incidentes e rollback.
- `#05-agentes`: comportamento de agentes, runtimes, tool use e políticas específicas.
- `#06-prompts-contexto`: prompt fragments, SystemPrompts, TaskPrompts e contexto de agentes.
- `#07-research`: pesquisa técnica, referências e benchmarks.
- `#09-release-changelog`: releases, changelog e histórico de entregas.
- `#10-ideias-livre`: ideias brutas antes de virar backlog.

Enquanto Slack não for parte do fluxo diário, não é necessário manter canais ativos.

Observação: se o conector usado pelo agente não tiver ferramenta para criar canais, a criação deve ser feita manualmente no Slack. Depois disso, convide o bot ou garanta que o usuário conectado participe do canal.

## Convenções de Mensagem

Use prefixos para facilitar busca e automação:

- `[IDEIA]`: pensamento bruto.
- `[TASK]`: solicitação que deve virar task no Orchestrator.
- `[PEDIDO-CODEX]`: solicitação técnica que deve virar ação no repositório enquanto o Orchestrator ainda não existir.
- `[DECISAO]`: decisão tomada.
- `[BLOQUEIO]`: algo impedindo avanço.
- `[RISCO]`: risco técnico, operacional ou estratégico.
- `[STATUS]`: atualização curta.
- `[PROMPT]`: proposta ou alteração de prompt fragment.
- `[INCIDENTE]`: falha relevante que exige análise.

## Threads

- Uma thread por assunto.
- Não misturar decisão, execução e debate longo no canal principal.
- Quando uma thread gerar decisão, registrar no repositório ou em ADR.

## Canvas no Slack

Use Slack Canvas apenas se o conector de chat virar parte do processo. Mantenha uma cópia estruturada no repositório em Markdown, porque agentes e Codex conseguem versionar e revisar texto com mais segurança.

Canvas sugeridos no Slack:

- Projeto: visão, canais, links importantes.
- Backlog: tarefas em triagem.
- Decisões: índice de ADRs.
- Rotinas: checklists semanais e mensais.

## Workflows Iniciais

Alguns workflows, conectores e recursos avançados podem depender do plano do Slack. Só ative quando GitHub/CLI deixarem de ser suficientes.

1. Captura de ideia:
   - Gatilho: atalho ou formulário no Slack.
   - Saída: mensagem em `#10-ideias-livre` ou `#00-inbox` com campos de problema, ideia e urgência.

2. Pedido para Task:
   - Gatilho: formulário em `#02-backlog`.
   - Saída: mensagem padronizada com objetivo, arquivos afetados, critério de aceite e validação esperada.

3. Decisão tomada:
   - Gatilho: formulário em `#03-decisoes`.
   - Saída: lembrete para criar ou atualizar ADR no repositório.

4. Revisão semanal:
   - Gatilho: agenda semanal.
   - Saída: checklist em `#00-inbox` com backlog, riscos, decisões e próximos passos.

## Integrações

Instalar somente quando houver necessidade real:

- GitHub: PRs, issues, commits e checks.
- Google Drive ou equivalente: arquivos externos que não devem ficar no repo.
- Calendário: revisões semanais e checkpoints.
- Futuro app próprio: comandos de orquestração, permissões e logs.

## Notificações

- Silenciar canais que não exigem ação imediata.
- Ativar notificação forte apenas para `#00-inbox`, `#02-backlog`, `#03-decisoes` e menções diretas.
- Usar lembretes para revisão semanal e decisões pendentes.

## Segurança

- Não publicar chaves de API, tokens ou senhas.
- Usar variáveis de ambiente e gerenciador de segredos.
- Separar canais de incidentes e decisões.
- Registrar ações autônomas de agentes com data, motivo e resultado.
