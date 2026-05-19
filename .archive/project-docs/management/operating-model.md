# Modelo de Gestão do Projeto

## Fonte de Verdade

Use o repositório como fonte de verdade. GitHub deve concentrar backlog, branches, pull requests, revisão e histórico operacional externo. Chats podem ajudar no futuro, mas não são fonte definitiva.

## Trio Recomendado

- CLI: controle inicial do MVP, comandos auditáveis e operação local.
- GitHub: issues, branches, pull requests, histórico e revisão.
- Agentes: execução técnica em sandboxes, worktrees e branches isoladas.
- Chat opcional futuro: captura rápida, avisos e notificações quando houver necessidade real.
- Codex: análise, implementação, manutenção de documentação e validação técnica.

## Fluxo de Trabalho

1. Capturar ideia na CLI, GitHub Issue ou em `docs/canvas/project-canvas.md`.
2. Transformar a ideia em item pequeno: problema, objetivo, escopo e critério de aceite.
3. Registrar como issue ou task.
4. Orchestrator quebra em work units quando houver paralelismo real.
5. Codex implementa em branch curta e sandbox.
6. Validar com testes, lint ou revisão manual.
7. Atualizar docs quando necessário.
8. Registrar decisões importantes como ADR.

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

1. Fundação: documentação, regras de agentes, canvas e GitHub-first.
2. MVP operacional: CLI, task/run/event store, task graph, prompt composer e agente fake.
3. Execução real: sandbox/worktree, Codex/CLI, policy engine e tool approvals.
4. Memória e auditoria: tracing normalizado, ledger, checkpoints, logs, histórico de ações e memória recursiva derivada de evidências.
5. Orquestrador avançado: agentes com papéis, permissões, live view e intervenções.
6. Autonomia progressiva: automações por nível de risco.
