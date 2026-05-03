# ADR 0001: Repositório Como Fonte de Verdade

## Contexto

O projeto será operado por humano e agentes de IA. Conversas soltas podem apoiar comunicação no futuro, mas não são suficientes para preservar contexto, decisões e histórico técnico.

## Decisão

O repositório será a fonte de verdade para código, documentação, canvas, políticas de autonomia e decisões arquiteturais. GitHub será usado para issues, pull requests, reviews, checks e histórico de mudanças. Codex atuará sobre o repositório.

## Consequências

- Agentes terão contexto direto e versionado.
- Decisões importantes precisam ser registradas em arquivos.
- Chat não deve virar arquivo morto de decisões finais.
- O projeto ganha base para auditoria e autonomia progressiva.

## Alternativas Consideradas

- Usar apenas chat: rápido, mas fraco para versionamento e contexto técnico.
- Usar apenas documentos visuais: bom para brainstorm, mas menos confiável para agentes.
- Usar apenas GitHub Issues: bom para execução, mas insuficiente para visão e operação cotidiana.
