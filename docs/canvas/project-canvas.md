# Canvas do Projeto

Canvas principal do projeto. Mantido em texto para que agentes e humanos consigam ler, revisar e versionar o contexto.

## Visão

Construir um Sistema de Orquestração de Agentes capaz de transformar intenção humana em planejamento, execução e validação contínua.

## Premissa Central

O sistema começa como pipeline local (Task → DAG → Execution), com autonomia aumentando progressivamente por políticas, testes e observabilidade.

## Usuário Inicial

- Fundador/operador solo.
- Pessoa que precisa transformar ideias em execução organizada via agentes de IA.

## Problemas

- Ideias ficam na mente e se perdem antes de virar execução.
- Agentes sem contexto suficiente produzem resultados inconsistentes.
- Autonomia sem auditoria cria risco operacional.

## Proposta de Valor

Pipeline onde tarefas são decompostas em DAG de work units, executadas por agentes isolados, com estado persistido e auditável.

## Arquitetura Atual

Pipeline architecture (Thin Orchestrator):

```
Task → planner.Plan() → []WorkUnit (DAG)
     → executor.Execute() → topological sort → runtime.Execute() per WU
     → store persists all state transitions
```

Packages: `domain/`, `planner/`, `executor/`, `runtime/`, `store/`, `event/`, `apperrors/`

3 regras: domain puro, dependências para baixo, SQL confinado a store/.

## Componentes Futuros

- Runtime real (Gemini, Codex) substituindo fake runtime.
- PostgreSQL store substituindo in-memory.
- CLI expandida com comandos de gestão de tasks.
- Painel web para visualização de DAGs e execuções.
- Policy engine para controle de autonomia.
- Sandbox com workspace isolation per work unit.

## Princípios

- Fonte de verdade versionada.
- Autonomia progressiva.
- Tudo importante deve deixar rastro.
- Decisões antes de automações irreversíveis.
- Agentes executam dentro de limites explícitos.

## Riscos

- Automatizar caos em vez de organizar o sistema.
- Dar autonomia antes de existir teste, log e rollback.
- Criar complexidade técnica antes de validar o fluxo operacional.

## Métricas Iniciais

- Tasks criadas e decompostas em DAG.
- Work units executadas com sucesso.
- Tempo entre criação da task e conclusão.
- Cobertura dos testes de arquitetura.
