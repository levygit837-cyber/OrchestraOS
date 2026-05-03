# ADR 0001: Repositório Como Fonte de Verdade

## Contexto

O projeto será operado por humano e agentes de IA. Slack será usado para comunicação e automações, mas mensagens soltas não são suficientes para preservar contexto, decisões e histórico técnico.

## Decisão

O repositório será a fonte de verdade para código, documentação, canvas, políticas de autonomia e decisões arquiteturais. Slack será usado como cockpit operacional. GitHub será usado para issues, pull requests e histórico de mudanças. Codex atuará sobre o repositório.

## Consequências

- Agentes terão contexto direto e versionado.
- Decisões importantes precisam ser registradas em arquivos.
- Slack não deve virar arquivo morto de decisões finais.
- O projeto ganha base para auditoria e autonomia progressiva.

## Alternativas Consideradas

- Usar apenas Slack: rápido, mas fraco para versionamento e contexto técnico.
- Usar apenas documentos visuais: bom para brainstorm, mas menos confiável para agentes.
- Usar apenas GitHub Issues: bom para execução, mas insuficiente para visão e operação cotidiana.

