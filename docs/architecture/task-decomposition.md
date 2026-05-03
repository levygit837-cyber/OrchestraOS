# Decomposição de Tasks

## Objetivo

Definir como o Orchestrator transforma uma mensagem humana em um plano executavel por agentes paralelos.

## Fluxo

```text
UserMessage
-> Orchestrator Intake
-> Intent Normalization
-> Task Creation
-> Risk and Scope Assessment
-> Task Graph Planning
-> WorkUnit Ownership
-> Prompt Assembly
-> Sandbox Setup
-> Agent Spawn
-> Agent Loop
-> Live View and Intervention
-> Validation
-> Review Diffs
-> Merge Decision
```

## Task Graph

O grafo de decomposicao deve ser aciclico.

Nodes:

- representam `WorkUnit`;
- tem objetivo pequeno;
- tem ownership explicito;
- tem criterios de aceite;
- tem validacao esperada;
- podem ser executados por agente ou humano.

Edges:

- representam dependencia;
- podem exigir artefato produzido por outro node;
- podem bloquear paralelismo;
- podem registrar conflito de ownership.

## Planejamento Pelo Orchestrator

O Orchestrator deve produzir:

- titulo e resumo da task;
- objetivo principal;
- nao escopo;
- criterios de aceite;
- riscos conhecidos;
- task graph;
- lista de work units;
- ownership de arquivos ou modulos;
- estrategia de validacao;
- politica de ferramentas;
- prompts iniciais;
- plano de rollback esperado.

## Paralelismo

Uma work unit so deve rodar em paralelo quando:

- nao depende de node ainda pendente;
- nao escreve nos mesmos caminhos de outra work unit ativa;
- nao exige segredo ou ferramenta bloqueada sem aprovacao;
- tem criterio de aceite proprio;
- tem validacao isolavel.

## Intervenção Humana

Humanos podem:

- conversar com o Orchestrator;
- enviar mensagem mediada para um agente;
- aprovar ou negar ferramenta;
- pausar, retomar ou cancelar run;
- pedir replanejamento;
- rejeitar merge.

Toda intervencao deve virar evento.

## Saída Esperada do Planejamento

Exemplo conceitual:

```json
{
  "task": {
    "title": "Implementar Event Store minimo",
    "risk_level": "medium",
    "acceptance_criteria": [
      "Persistir eventos com task_id e run_id",
      "Consultar eventos por run",
      "Validar schema do envelope"
    ]
  },
  "graph": {
    "nodes": [
      {
        "id": "wu_001",
        "objective": "Criar schema SQL inicial",
        "owned_paths": ["migrations/"]
      },
      {
        "id": "wu_002",
        "objective": "Implementar repository Go",
        "owned_paths": ["internal/events/"]
      }
    ],
    "edges": [
      {
        "from": "wu_001",
        "to": "wu_002",
        "type": "blocks"
      }
    ]
  }
}
```

## Critérios de Qualidade

- Uma work unit deve caber em uma branch curta.
- Uma work unit deve poder falhar sem invalidar o plano inteiro.
- O plano deve declarar incertezas.
- O plano deve evitar que dois agentes editem o mesmo arquivo sem coordenacao explicita.
- O Orchestrator deve preferir menos agentes bem coordenados a paralelismo artificial.
