# Tasks Transversais

Este diretorio armazena artefatos de tarefas que **nao pertencem a um unico dominio** ou que sao **pequenas demais** para justificar um contexto de domain.

## Quando usar

- A task toca multiplos modulos sem um domain claro (ex: refatoracao global)
- A task e de infraestrutura/core (ex: atualizar `internal/core/eventstore/`)
- A task e simples (1 arquivo, comportamento localizado)

## Quando NAO usar

- A task pertence a um domain claro → use `docs/agent/domains/<domain>/`

## Estrutura

```
tasks/
└── YYYY-MM-DD_<slug>/
    ├── briefing.md
    ├── spec.md        (se aplicavel)
    ├── plan.md        (se aplicavel)
    └── review.md
```

Consulte `docs/agent/ARTIFACT_ORGANIZATION.md` para a convencao completa de naming.
