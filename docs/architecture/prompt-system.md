# Sistema de Prompts

## Objetivo

Permitir que o Orchestrator monte prompts precisos, auditaveis e flexiveis para agentes especializados sem criar um prompt monolitico dificil de manter.

## Artefatos

### SystemPrompt

Define como o agente deve operar.

Conteudo tipico:

- papel do agente;
- autonomia permitida;
- limites de seguranca;
- ferramentas disponiveis;
- regras de checkpoint;
- formato de eventos;
- comportamento em erro;
- contrato de saida.

### TaskPrompt

Define o trabalho concreto.

Conteudo tipico:

- objetivo;
- contexto;
- memorias recuperadas quando disponiveis;
- arquivos ou modulos sob responsabilidade;
- nao escopo;
- criterios de aceite;
- validacoes esperadas;
- evidencias exigidas;
- riscos conhecidos;
- todo ledger inicial.

### PromptFragment

Bloco reutilizavel.

Campos recomendados:

```json
{
  "id": "fragment.agent.code_worker",
  "version": "1.0.0",
  "kind": "persona",
  "title": "Code Worker",
  "priority": 300,
  "applies_when": {
    "work_unit_type": "code_change"
  },
  "requires": [
    "fragment.policy.autonomy_level_2"
  ],
  "conflicts_with": [
    "fragment.agent.reviewer_only"
  ],
  "body": "..."
}
```

### DynamicPromptFragment

Bloco temporario criado pelo Orchestrator para uma run especifica quando os fragmentos existentes nao cobrem bem a necessidade da work unit.

Campos recomendados:

```json
{
  "id": "dynamic_fragment.run_123.specialized_migration_reviewer",
  "kind": "dynamic_domain",
  "title": "Especializacao temporaria para revisar migracao",
  "created_for_run_id": "run_123",
  "created_for_work_unit_id": "wu_456",
  "reason": "A task exige revisar migracao especifica que nao possui fragmento estatico.",
  "expires_after_run": true,
  "priority": 650,
  "body": "..."
}
```

Regras:

- deve ser usado apenas na run para a qual foi criado;
- deve entrar no `PromptSnapshot`;
- nao pode contradizer politicas globais;
- nao pode alterar nivel de autonomia;
- nao pode ampliar ownership ou permissoes;
- deve virar candidato a fragmento estatico se for reutilizado com frequencia.

### RetrievedMemoryBundle

Bloco de contexto recuperado pelo Sistema de Memoria Recursiva para uma `AgentSession` ou checkpoint.

Campos recomendados:

```json
{
  "id": "memory_bundle_123",
  "run_id": "run_123",
  "agent_session_id": "session_789",
  "query_reason": "Contexto inicial para work unit de prompting.",
  "memory_refs": ["mem_123", "mem_456"],
  "evidence_refs": [
    "docs/adr/0007-prompt-composition-system.md",
    "docs/architecture/memory-system.md"
  ],
  "token_budget": 1200,
  "created_at": "2026-05-03T12:00:00Z"
}
```

Regras:

- deve ser pequeno e deduplicado;
- deve citar ids de memoria e evidencias;
- nao pode sobrescrever politica, ADR, autonomia, ownership ou criterio de aceite;
- deve entrar no `PromptSnapshot` quando usado no prompt inicial;
- quando injetado durante checkpoint, deve ser registrado como evento auditavel;
- nao deve repetir memorias ja entregues na mesma run sem motivo registrado.

### ToolsetSnapshot

Registro das ferramentas disponiveis para uma `AgentSession`.

Campos recomendados:

```json
{
  "id": "toolset_snapshot_123",
  "run_id": "run_123",
  "agent_session_id": "session_789",
  "tools": [
    {
      "name": "filesystem.read",
      "scope": "owned_worktree",
      "risk": "safe"
    },
    {
      "name": "tests.run",
      "scope": "local_no_network",
      "risk": "guarded"
    }
  ],
  "created_reason": "Work unit de implementacao local sem rede.",
  "created_at": "2026-05-03T12:00:00Z"
}
```

### AgentSessionReconfiguration

Registro de mudanca controlada de prompt ou toolset durante uma run.

Campos recomendados:

```json
{
  "id": "reconfig_123",
  "run_id": "run_123",
  "previous_agent_session_id": "session_old",
  "next_agent_session_id": "session_new",
  "reason": "Agente precisa rodar validacao que nao estava no toolset inicial.",
  "changes": {
    "added_tools": ["tests.run"],
    "added_fragments": ["dynamic_fragment.run_123.test_strategy"]
  },
  "ledger_preserved": true,
  "created_at": "2026-05-03T12:10:00Z"
}
```

## Tipos de Fragmento

| Tipo | Função |
| --- | --- |
| `global_policy` | Regras que sempre prevalecem. |
| `repo_context` | Instrucoes do repositorio e AGENTS.md. |
| `autonomy_policy` | Nivel de autonomia e limites. |
| `persona` | Papel do agente. |
| `dynamic_persona` | Especializacao temporaria baseada em persona principal. |
| `operating_mode` | Como executar: implementador, revisor, pesquisador, depurador. |
| `domain` | Contexto tecnico ou de produto. |
| `dynamic_domain` | Contexto temporario criado para uma run especifica. |
| `retrieved_memory` | Contexto recuperado de memoria recursiva com evidencias. |
| `tool_policy` | Ferramentas permitidas e pedidos de aprovacao. |
| `dynamic_tool_guidance` | Orientacao temporaria sobre uso de ferramentas aprovadas na run. |
| `communication` | Eventos, mensagens e checkpoints. |
| `validation` | Testes, evidencias e criterios de pronto. |
| `output_contract` | Formato final esperado. |
| `ledger` | Regras para atualizar todos e progresso. |

## Ordem de Precedência

1. Politicas globais e seguranca.
2. Politica de autonomia.
3. Instrucoes versionadas do repositorio.
4. Regras do Orchestrator para a run.
5. Persona e modo operacional.
6. Contexto de dominio.
7. Memoria recuperada com evidencias.
8. Task prompt.
9. Formato de saida.

Fragmentos de menor prioridade nao podem contradizer fragmentos de maior prioridade.

## Montagem

O Orchestrator deve:

1. classificar a task;
2. selecionar perfil de agente;
3. selecionar fragmentos obrigatorios;
4. selecionar fragmentos condicionais;
5. resolver conflitos;
6. renderizar variaveis;
7. selecionar toolset minimo;
8. recuperar memoria relevante quando a capacidade estiver habilitada;
9. criar `PromptSnapshot`;
10. criar `ToolsetSnapshot`;
11. persistir hashes e referencias;
12. iniciar agente com o prompt e toolset montados.

## PromptSnapshot

Cada run deve registrar:

- ids e versoes dos fragmentos;
- fragmentos dinamicos usados;
- memorias recuperadas usadas no prompt inicial;
- valores de variaveis usados;
- system prompt renderizado ou referencia segura;
- task prompt renderizado ou referencia segura;
- hash dos prompts;
- data de criacao;
- run associada.

## Especializacao Dinamica

O Orchestrator pode criar especializacoes temporarias para uma run quando:

- nenhuma persona estatica cobre bem o contexto;
- a task exige conhecimento de dominio muito especifico;
- a work unit precisa de heuristicas temporarias de investigacao;
- a decomposicao criou um papel pequeno demais para justificar um agente fixo;
- a combinacao de fragmentos existentes ficaria grande ou ambigua.

Nao deve criar fragmento dinamico quando:

- a task cabe em persona principal existente;
- o objetivo ainda esta mal definido;
- a especializacao tentaria contornar politica;
- a especializacao ampliaria permissao sem aprovacao;
- a mudanca deveria ser uma ADR, issue ou documento versionado.

## Toolset Minimo Por Run

Cada `AgentSession` deve iniciar com o menor conjunto de ferramentas necessario.

Exemplos:

| Perfil | Toolset inicial |
| --- | --- |
| `docs_writer` | leitura do worktree, edicao em caminhos de docs, diff local. |
| `code_worker` | leitura, edicao no ownership, testes locais, diff local. |
| `reviewer` | leitura, diff, testes locais se aprovados, comentarios estruturados. |
| `debugger` | leitura, testes locais, logs, comandos diagnosticos com timeout. |

O agente pode solicitar ferramenta ausente por evento estruturado. O pedido deve incluir:

- ferramenta desejada;
- motivo;
- escopo;
- risco percebido;
- alternativa tentada;
- impacto de negar.

## Reconfiguracao De Sessao

Se o Orchestrator aprovar mudanca de toolset ou prompt durante uma run, ele deve preferir uma reconfiguracao explicita em vez de alterar o agente silenciosamente.

Fluxo recomendado:

1. agente solicita ferramenta ou especializacao ausente;
2. Orchestrator avalia politica, risco e escopo;
3. Orchestrator decide negar, aprovar expansao limitada, criar nova work unit ou reconfigurar sessao;
4. se reconfigurar, cria novo `PromptSnapshot` e `ToolsetSnapshot`;
5. preserva `AgentTaskLedger`;
6. reinicia ou substitui a `AgentSession`;
7. registra `agent.session_reconfigured`.

Essa operacao nao deve apagar historico da sessao anterior.

## Todo Ledger no Prompt

O TaskPrompt deve incluir um bloco inicial de ledger:

```text
Objective:
- ...

Acceptance Criteria:
- ...

Todo Ledger:
- [ ] Confirmar escopo e arquivos sob ownership.
- [ ] Implementar menor mudanca suficiente.
- [ ] Rodar validacao definida.
- [ ] Registrar evidencia e riscos restantes.
```

O agente nao deve tratar o ledger como backlog livre. Ele deve atualizar apenas o progresso operacional da work unit.

## Qualidade de Prompt

Um prompt bom deve:

- dizer exatamente o objetivo;
- declarar limites;
- informar quais ferramentas exigem aprovacao;
- informar quais ferramentas estao disponiveis naquela sessao;
- definir como pedir ajuda;
- definir validacao;
- definir saida;
- reduzir ambiguidade sem sufocar julgamento tecnico.

Um prompt ruim normalmente:

- mistura varias personas;
- contradiz politicas;
- nao define criterios de aceite;
- pede autonomia maior que a permitida;
- omite como registrar evidencias.
