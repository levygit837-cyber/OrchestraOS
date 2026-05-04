# ADR 0015: TUI Como Interface Local Principal

## Contexto

A ADR 0005 definiu uma progressao inicial baseada em scripts de bootstrap, CLI fina e GitHub. Depois da primeira CLI minima, ficou claro que operacoes como acompanhar runs, inspecionar eventos, aprovar ferramentas, ver sessoes de agente e revisar evidencias exigem uma experiencia viva e navegavel.

Uma CLI tradicional e adequada para automacao, testes e comandos pontuais, mas fica limitada para observabilidade operacional continua. O projeto e local-first no MVP, com Go, Postgres, Event Store e GitHub-first como restricoes ja aprovadas.

Ainda nao ha decisao final sobre o framework TUI.

## Decisao

O MVP deve evoluir da CLI como interface humana principal para uma TUI local.

A TUI passa a ser a superficie primaria para uso humano local em:

- dashboard de tasks, work units e runs;
- live view de eventos e sessoes de agente;
- aprovacoes e negacoes de ferramentas;
- inspecao de checkpoints, evidencias e falhas;
- comandos guiados para criar tasks, iniciar runs e acompanhar validacoes.

A CLI atual nao deve ser removida imediatamente. Ela deve permanecer como camada headless para automacao, testes, scripts e operacao por CI/local shell. A TUI deve reutilizar servicos internos e contratos compartilhados, nao chamar comandos shell da CLI como integracao principal.

Framework ainda sera decidido por spike tecnico curto. A recomendacao inicial e:

- **Bubble Tea + Bubbles + Lip Gloss**, da Charmbracelet, como primeira opcao.

Motivos:

- e Go-native e combina com a stack aprovada;
- usa arquitetura baseada em estado e mensagens, adequada para eventos, runs e sessoes;
- facilita testes de update/model sem terminal real;
- tem componentes reutilizaveis para listas, tabelas, forms, spinners, viewport e keybindings;
- permite evoluir telas ricas sem introduzir TypeScript, Electron ou Web antes da hora.

## Consequencias

- A interface humana fica mais adequada para operacao continua de agentes.
- O nucleo do produto continua local-first e Go-first.
- A CLI deixa de ser o centro da experiencia humana, mas segue importante como contrato automatizavel.
- O projeto precisara separar melhor servicos de aplicacao da camada `cmd`, para que CLI e TUI compartilhem a mesma logica.
- A TUI deve respeitar a autonomia M0 aprovada: sugestao e execucao com revisao humana, sem operar acima de Nivel 2.

Antes da implementacao da TUI, o projeto deve criar um plano pequeno de spike com criterios:

- renderizar lista de tasks/runs/eventos com dados reais do Postgres;
- navegar sem bloquear leitura de eventos;
- aceitar entradas de comando guiadas;
- testar transicoes de tela e estado sem terminal real;
- avaliar ergonomia de tabelas, filtros e live log;
- medir o custo de manter CLI e TUI sobre os mesmos servicos internos.

## Alternativas consideradas

- **Bubble Tea / Bubbles / Lip Gloss**: melhor alinhamento com Go, testes e fluxos por eventos. E a opcao recomendada para spike.
- **tview / tcell**: entrega CRUD e layouts rapidamente, com componentes prontos. Pode ser melhor se o foco for velocidade de painel administrativo, mas tende a ser menos elegante para fluxos complexos por mensagens.
- **termui**: simples para dashboards, mas menos adequada para uma aplicacao operacional com forms, navegacao e estados ricos.
- **TUI em Rust com ratatui**: excelente qualidade tecnica, mas cria uma segunda stack e aumenta custo do MVP.
- **Web ou Desktop agora**: continuam adiados; oferecem mais riqueza visual, mas aumentam escopo antes de estabilizar o control plane.
- **Apenas CLI tradicional**: preserva simplicidade, mas nao resolve bem live view, aprovacoes e acompanhamento continuo.
