# Instruções Para Agentes

Este arquivo deve ser lido por qualquer agente antes de editar o projeto.

## Prioridades

1. Preservar a intenção do usuário.
2. Manter o repositório como fonte de verdade.
3. Fazer mudanças pequenas, verificáveis e reversíveis.
4. Atualizar documentação quando a mudança alterar comportamento, arquitetura ou processo.
5. Nunca tratar conversa solta, chat, comentários avulsos ou memória do agente como fonte definitiva.

## Fluxo Obrigatório

1. Entender o item de trabalho.
2. Consultar o canvas em `docs/canvas/project-canvas.md`.
3. Consultar decisões arquiteturais em `docs/adr/`.
4. Seguir o `docs/agent/PLAYBOOK.md` para gerar artefatos necessários (BRIEFING, SPEC, PLAN quando aplicável).
5. Implementar a menor mudança suficiente.
6. Rodar validações relevantes.
7. Registrar o que mudou e qualquer risco restante.

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

## Commits e Branches

**NUNCA commit ou push diretamente na branch `main`.**

Sempre use o script controlado:
```bash
./scripts/safe-commit.sh "mensagem do commit"
```

Este script automaticamente:
- Cria uma feature branch se você estiver na `main`
- Roda todas as validações (`go vet`, architecture tests, contracts)
- Só commita se tudo passar

Depois do commit, push a feature branch e abra um Pull Request. Aguarde o CI passar antes de mergear.

Para instalar os hooks localmente (proteção adicional):
```bash
cp scripts/pre-commit.sh .git/hooks/pre-commit
cp scripts/pre-push.sh .git/hooks/pre-push
chmod +x .git/hooks/pre-commit .git/hooks/pre-push
```

## Novo Módulo

Antes de criar um novo módulo, execute `./scripts/new-module.sh <nome>` para gerar a estrutura padronizada.
Após implementar, execute `./scripts/verify-contracts.sh` e `./scripts/lint.sh` antes de commitar.

## Padrões de Código

Consulte `docs/development/CODING_STANDARDS.md` para regras detalhadas de estilo, naming, error handling e testes.

## Decisões

Decisões relevantes devem virar ADR em `docs/adr/`. Use o formato:

- Contexto
- Decisão
- Consequências
- Alternativas consideradas

## Tipos de Plano

O projeto suporta 4 tipos de plano. O tipo é metadado que orienta o executor:

| Tipo | Quando Usar |
|---|---|
| **Faseado** | Fluxo sequencial com dependências temporais |
| **Por Domínio** | Módulos independentes evoluem separadamente |
| **Árvore de Decisões** | Problema técnico aberto com múltiplos caminhos |
| **Cenário-Based** | Feature user-facing com múltiplos fluxos de usuário |

Consulte `docs/development/plan-types.md` para detalhes e templates.

Todo plano deve declarar seu tipo no front matter ou cabeçalho.
