# Plano de Transicao CLI -> TUI

## Objetivo

Evoluir a interface humana local do OrchestraOS de uma CLI fina para uma TUI operacional, mantendo comandos headless para automacao, testes e scripts.

## Premissas

- A stack principal continua Go + Postgres.
- O Event Store continua sendo a fonte canonica de historico operacional.
- A autonomia aprovada para o MVP continua limitada ao Nivel 2.
- A TUI deve reutilizar servicos internos compartilhados com a CLI.
- O framework TUI ainda nao esta decidido, mas Bubble Tea e a primeira opcao para spike.

## Recomendacao de Stack

Opcao recomendada para spike:

- `github.com/charmbracelet/bubbletea` para runtime/modelo de aplicacao TUI.
- `github.com/charmbracelet/bubbles` para componentes.
- `github.com/charmbracelet/lipgloss` para estilos.
- `github.com/charmbracelet/glamour` apenas se houver necessidade de renderizar Markdown no terminal.

Opcao alternativa:

- `github.com/rivo/tview` + `github.com/gdamore/tcell` caso o foco seja montar rapidamente telas CRUD e tabelas administrativas.

Decisao sugerida:

- escolher Bubble Tea se a prioridade for fluxo por eventos, testes de estado, live view e UX mais refinada;
- escolher tview se a prioridade for velocidade de entrega de formularios/tabelas tradicionais.

## Fase 0: Corrigir Base Antes do Spike

1. Manter `internal/migrations` compilando e apontando para as migrations SQL versionadas do repositorio.
2. Manter `EventStore.Append` preenchendo campos gerados antes da validacao.
3. Manter o schema do `EventEnvelope` com `run_id` condicional para eventos de runtime.
4. Garantir que updates de status preservem timestamps anteriores.
5. Fazer comandos falharem quando append de evento ou update de estado falhar.
6. Isolar testes de integracao com dados proprios por teste, evitando dependencia de estado residual.

Resultado esperado:

- `go list ./...` passa.
- testes de repositorio e Event Store passam contra banco limpo.
- comandos minimos de task, work unit, run e event funcionam sem inconsistencias conhecidas.

## Fase 1: Separar Servicos Compartilhados

Criar uma camada pequena de servicos de aplicacao para reutilizacao por CLI e TUI:

- `TaskService`
- `WorkUnitService`
- `RunService`
- `EventService`
- `AgentSessionService`

Esses servicos devem concentrar:

- validacao de entrada;
- transicoes de estado;
- append de eventos;
- erros com contexto;
- operacoes atomicas quando houver mais de uma escrita relacionada.

Resultado esperado:

- comandos Cobra ficam finos;
- TUI consegue chamar os mesmos servicos sem shelling out;
- testes de regra de negocio ficam fora da camada de terminal.

## Fase 2: Spike de Framework TUI

Implementar dois prototipos pequenos, se necessario:

1. Bubble Tea:
   - lista de tasks;
   - painel de eventos;
   - detalhe de run;
   - refresh manual;
   - teste de update/model.

2. tview:
   - mesma navegacao basica;
   - tabela de eventos;
   - formulario simples de criar task.

Criterios de avaliacao:

- facilidade de testar estado;
- ergonomia para live view de eventos;
- suporte a formularios e tabelas;
- custo de composicao de telas;
- simplicidade para manter keybindings e filtros;
- compatibilidade com operacao local-first.

Resultado esperado:

- decisao final de framework registrada em nova ADR ou atualizacao da ADR 0015.

## Fase 3: MVP da TUI

Primeiras telas:

- Home com resumo de tasks, runs e sessoes.
- Tasks: listar, criar, abrir detalhe.
- Work Units: listar por task e criar work unit.
- Runs: iniciar run fake, listar runs e abrir detalhe.
- Events: listar, filtrar por task/run/work unit e replay.
- Agent Sessions: status, heartbeat/checkpoint, detalhe.

Primeiros fluxos:

- criar task;
- criar work unit;
- iniciar run fake;
- acompanhar eventos;
- revisar resultado;
- abrir evidencias/checkpoints quando existirem.

## Fase 4: Validacao

Validacoes minimas:

- testes unitarios dos servicos;
- testes de update/model da TUI;
- testes de integracao Postgres para repositories e Event Store;
- execucao manual documentada dos fluxos principais;
- registro de riscos restantes ao final de cada entrega.

## Fora do Escopo Inicial

- painel web;
- desktop app;
- runtime Codex/CLI completo;
- aprovacoes de ferramentas com politica sofisticada;
- multiusuario;
- conectores de chat.
