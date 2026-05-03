# Modelo de Gestão do Projeto

## Fonte de Verdade

Use o repositório como fonte de verdade. Slack é ótimo para conversa, captura e notificações, mas decisões finais, arquitetura, código, políticas e documentação devem estar versionados.

## Trio Recomendado

- Slack: entrada de ideias, comandos, discussões rápidas, avisos e rotinas.
- GitHub: issues, branches, pull requests, histórico e revisão.
- Codex: análise, implementação, manutenção de documentação e validação técnica.

## Fluxo de Trabalho

1. Capturar ideia em Slack ou em `docs/canvas/project-canvas.md`.
2. Transformar a ideia em item pequeno: problema, objetivo, escopo e critério de aceite.
3. Registrar como issue ou tarefa.
4. Codex implementa em branch curta.
5. Validar com testes, lint ou revisão manual.
6. Atualizar docs quando necessário.
7. Registrar decisões importantes como ADR.

## Ritual Leve

- Diário: capturar ideias e decidir o próximo item pequeno.
- Semanal: revisar backlog, riscos e decisões pendentes.
- Mensal: revisar visão, métricas e nível de autonomia permitido aos agentes.

## Estrutura de Backlog

Categorias sugeridas:

- `idea`: ideia bruta ainda sem escopo.
- `discovery`: precisa de investigação.
- `feature`: entrega funcional.
- `automation`: rotina automatizada.
- `architecture`: decisão estrutural.
- `ops`: operação, monitoramento ou incidente.
- `docs`: documentação e clareza.

Prioridades sugeridas:

- `P0`: bloqueia o projeto.
- `P1`: importante para o MVP.
- `P2`: melhora relevante.
- `P3`: pode esperar.

## Definição de Pronto

Um item está pronto para execução quando tem:

- Objetivo claro.
- Escopo pequeno.
- Critérios de aceite.
- Risco conhecido.
- Validação esperada.

## Definição de Concluído

Um item só deve ser considerado concluído quando:

- A mudança foi implementada.
- A validação relevante foi executada ou o motivo da falta foi registrado.
- Documentos afetados foram atualizados.
- Riscos restantes foram descritos.

## Roadmap Inicial

1. Fundação: documentação, regras de agentes, canvas e Slack.
2. MVP operacional: capturar ideia, transformar em tarefa, executar com Codex e reportar status.
3. Memória e auditoria: registrar decisões, logs e histórico de ações.
4. Orquestrador: agentes com papéis, permissões e políticas.
5. Autonomia progressiva: automações por nível de risco.

