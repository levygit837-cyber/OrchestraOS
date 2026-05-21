# Tasks Transversais

Este diretorio armazena artefatos de **qualquer tarefa que nao pertence a um unico dominio**.

O tamanho ou complexidade da task nao importa — uma task transversal pode ser tao simples quanto um ajuste de infraestrutura ou tao complexa quanto uma refatoracao cross-module com plano por dominio.

## Quando usar

- A task toca `internal/core/*` (infraestrutura compartilhada)
- A task toca 2+ modulos de `internal/modules/`
- A task e de infraestrutura pura (CI, migrations, docker)
- A task nao tem um domain claro como "dono"

## Quando NAO usar

- A task toca APENAS um modulo de `internal/modules/` e NAO toca `internal/core/` → use `docs/agent/domains/<domain>/`

## Estrutura

```
tasks/
└── YYYY-MM-DD_<slug>/
    ├── briefing.md
    ├── spec.md        (se aplicavel)
    ├── plan.md        (se aplicavel, qualquer tipo)
    └── review.md
```

Para tasks cross-module complexas, o `plan.md` pode usar o tipo **Por Dominio** (`docs/development/plan-types.md`) com secoes por modulo afetado.

Consulte `docs/agent/ARTIFACT_ORGANIZATION.md` para a convencao completa de naming.
