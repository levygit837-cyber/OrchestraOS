# Configuração Recomendada do Slack

Slack deve ser usado como cockpit operacional, não como fonte de verdade definitiva.

## Workspace

Sugestão inicial:

- Nome do workspace: usar o nome escolhido do projeto.
- Plano inicial: gratuito enquanto o projeto estiver em descoberta; revisar plano pago quando precisar de histórico maior, workflows avançados, canvas autônomos ou integrações mais fortes.
- Idioma: português, se o projeto for pessoal; inglês, se houver intenção de equipe internacional.

## Canais Iniciais

- `#00-hq`: centro do projeto, anúncios e status geral.
- `#01-ideias`: captura rápida de ideias ainda brutas.
- `#02-planejamento`: priorização, escopo e decisões de produto.
- `#03-engenharia`: implementação, arquitetura e dúvidas técnicas.
- `#04-codex`: pedidos para o Codex, logs resumidos e evidências de execução.
- `#05-automacoes`: workflows, bots e eventos automáticos.
- `#06-decisoes`: links para ADRs, decisões finalizadas e mudanças de direção.
- `#07-incidentes`: problemas, falhas, rollback e análise posterior.

Para começo solo, mantenha poucos canais ativos. Se ficar demais, use apenas `#00-hq`, `#01-ideias`, `#04-codex` e `#06-decisoes`.

## Convenções de Mensagem

Use prefixos para facilitar busca e automação:

- `[IDEIA]`: pensamento bruto.
- `[PEDIDO-CODEX]`: solicitação que deve virar ação no repositório.
- `[DECISAO]`: decisão tomada.
- `[BLOQUEIO]`: algo impedindo avanço.
- `[RISCO]`: risco técnico, operacional ou estratégico.
- `[STATUS]`: atualização curta.

## Threads

- Uma thread por assunto.
- Não misturar decisão, execução e debate longo no canal principal.
- Quando uma thread gerar decisão, registrar no repositório ou em ADR.

## Canvas no Slack

Use Slack Canvas para notas humanas de projeto, reuniões e checklists. Porém, mantenha uma cópia estruturada no repositório em Markdown, porque agentes e Codex conseguem versionar e revisar texto com mais segurança.

Canvas sugeridos no Slack:

- Projeto: visão, canais, links importantes.
- Backlog: tarefas em triagem.
- Decisões: índice de ADRs.
- Rotinas: checklists semanais e mensais.

## Workflows Iniciais

Alguns workflows, conectores e recursos avançados podem depender do plano do Slack. Comece simples e só pague quando o fluxo estiver validado.

1. Captura de ideia:
   - Gatilho: atalho ou formulário no Slack.
   - Saída: mensagem em `#01-ideias` com campos de problema, ideia e urgência.

2. Pedido para Codex:
   - Gatilho: formulário em `#04-codex`.
   - Saída: mensagem padronizada com objetivo, arquivos afetados, critério de aceite e validação esperada.

3. Decisão tomada:
   - Gatilho: formulário em `#06-decisoes`.
   - Saída: lembrete para criar ou atualizar ADR no repositório.

4. Revisão semanal:
   - Gatilho: agenda semanal.
   - Saída: checklist em `#00-hq` com backlog, riscos, decisões e próximos passos.

## Integrações

Instalar quando houver necessidade real:

- GitHub: PRs, issues, commits e checks.
- Google Drive ou equivalente: arquivos externos que não devem ficar no repo.
- Calendário: revisões semanais e checkpoints.
- Futuro app próprio: comandos de orquestração, permissões e logs.

## Notificações

- Silenciar canais que não exigem ação imediata.
- Ativar notificação forte apenas para `#00-hq`, `#04-codex` e menções diretas.
- Usar lembretes para revisão semanal e decisões pendentes.

## Segurança

- Não publicar chaves de API, tokens ou senhas.
- Usar variáveis de ambiente e gerenciador de segredos.
- Separar canais de incidentes e decisões.
- Registrar ações autônomas de agentes com data, motivo e resultado.
