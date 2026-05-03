# Instruções Para Agentes

Este arquivo deve ser lido por qualquer agente antes de editar o projeto.

## Prioridades

1. Preservar a intenção do usuário.
2. Manter o repositório como fonte de verdade.
3. Fazer mudanças pequenas, verificáveis e reversíveis.
4. Atualizar documentação quando a mudança alterar comportamento, arquitetura ou processo.
5. Nunca tratar Slack, conversa solta ou memória do agente como fonte definitiva.

## Fluxo Obrigatório

1. Entender o item de trabalho.
2. Consultar o canvas em `docs/canvas/project-canvas.md`.
3. Consultar decisões arquiteturais em `docs/adr/`.
4. Implementar a menor mudança suficiente.
5. Rodar validações relevantes.
6. Registrar o que mudou e qualquer risco restante.

## Boas Práticas de Código

- Preferir código simples, explícito e testável.
- Usar tipagem estática quando a linguagem permitir.
- Separar regras de negócio, integrações externas e interface de usuário.
- Validar entradas nas bordas do sistema.
- Tratar erros com contexto suficiente para diagnóstico.
- Evitar dependências novas sem justificativa.
- Criar testes para lógica de orquestração, permissões, automações e integrações.
- Nunca colocar segredos em arquivos versionados.
- Documentar contratos entre agentes, ferramentas e serviços usando schemas ou OpenAPI quando aplicável.

## Política de Autonomia

O projeto deve ganhar autonomia por níveis:

- Nível 0: humano decide e executa.
- Nível 1: IA sugere planos e documentos.
- Nível 2: IA implementa com revisão humana.
- Nível 3: IA abre PRs com testes e evidências.
- Nível 4: IA executa tarefas operacionais de baixo risco dentro de políticas definidas.
- Nível 5: IA opera domínios aprovados com monitoramento, trilha de auditoria e rollback.

Nenhum agente deve assumir autonomia maior que a aprovada explicitamente nos documentos do projeto.

## Decisões

Decisões relevantes devem virar ADR em `docs/adr/`. Use o formato:

- Contexto
- Decisão
- Consequências
- Alternativas consideradas

