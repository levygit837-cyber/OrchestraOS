# ADR 0010: Operação GitHub-First e Chat Opcional

## Contexto

As decisões iniciais tratavam Slack como cockpit operacional complementar. Na prática, o fluxo mais natural do MVP e do operador atual e trabalhar com agentes, repositório, GitHub, branches, worktrees, issues, pull requests, diffs e evidências versionadas.

Slack adiciona uma camada conversacional útil no futuro, mas hoje não e hábito operacional do usuário e não deve virar dependência do caminho crítico. O projeto precisa reduzir superfícies externas até validar o núcleo de orquestração.

## Decisão

O MVP será **GitHub-first**.

O caminho operacional principal será:

- repositório como fonte de verdade;
- CLI local para comandos do Orchestrator;
- Git worktree e branch por task;
- GitHub Issues para backlog e escopo quando necessário;
- GitHub Pull Requests para revisão, evidências e merge;
- agentes executando em sandboxes controlados;
- Event Store como histórico operacional interno.

Slack, Discord, e-mail ou qualquer outra interface conversacional serão tratados como **conectores opcionais futuros**, não como dependências do MVP.

## Consequências

- O MVP fica mais simples, auditável e alinhado ao fluxo de engenharia.
- A integração com GitHub ganha prioridade sobre qualquer integração de chat.
- Worktrees e PRs passam a ser o principal mecanismo de isolamento, revisão e integração.
- Documentos de Slack podem permanecer como referência opcional, mas não devem aparecer como requisito de arquitetura, entrega ou aceite do MVP.
- Aprovações críticas devem acontecer pela CLI, GitHub ou mecanismo interno persistido antes de existirem conectores de chat.

## Alternativas consideradas

- **Slack como cockpit inicial**: bom para notificações e conversa, mas cria dependência operacional que o usuário não usa com frequência.
- **CLI-only sem GitHub**: simples localmente, mas perde revisão, histórico colaborativo e PRs.
- **Desktop/Web desde o início**: melhora experiência visual, mas aumenta custo antes de validar o núcleo.
- **GitHub-first**: aproveita primitives já fortes para engenharia: issues, branches, PRs, checks, reviews e histórico.
