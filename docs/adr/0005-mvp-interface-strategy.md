# ADR 0005: Interface Inicial do MVP

## Contexto

O OrchestraOS precisa de uma forma inicial de operar o Orchestrator, criar tasks, acompanhar runs, aprovar ferramentas e inspecionar evidencias. As opcoes consideradas sao scripts soltos, CLI, Desktop e Web.

O projeto ainda esta em fase de validacao do nucleo: task graph, prompt composer, sandbox, agentes, eventos, politicas e auditoria. Uma interface rica antes desse nucleo estar validado pode aumentar custo sem reduzir o maior risco tecnico.

GitHub sera o complemento operacional principal da CLI. Conectores de chat podem existir no futuro, mas nao fazem parte do caminho critico do MVP.

## Decisão

O MVP tera uma progressao em tres camadas:

1. **Scripts de bootstrap** para desenvolvimento local e validacao dos componentes internos.
2. **CLI fina** como primeira interface oficial do MVP para criar tasks, iniciar runs, acompanhar eventos, enviar mensagens, aprovar ou negar ferramentas e revisar evidencias.
3. **GitHub** como superficie externa principal para issues, pull requests, revisoes, checks e evidencias.

Desktop e Web nao entram como requisito do primeiro MVP.

Desktop permanece como candidato forte para uma experiencia local rica no futuro, especialmente para uso solo, visualizacao de traces e intervencao em agentes. Web permanece como candidato quando houver servidor remoto, multiplos usuarios, acesso de equipe ou necessidade de painel operacional compartilhado.

## Consequências

- O nucleo do Orchestrator pode ser validado sem depender de UI complexa.
- Scripts continuam permitidos, mas nao devem virar interface permanente ou contrato publico.
- A CLI cria um contrato operacional testavel e automatizavel.
- GitHub complementa a CLI com revisao, historico e integracao via pull requests.
- Chat fica opcional e nao substitui comandos auditaveis, PRs ou estado persistido.
- A decisao Desktop vs Web sera revisitada depois que o fluxo local fim a fim estiver funcionando.

---

## 2. Evolução para TUI como Interface Local Principal

### 2.1 Contexto adicional

Depois da primeira CLI mínima, ficou claro que operações como acompanhar runs, inspecionar eventos, aprovar ferramentas, ver sessões de agente e revisar evidências exigem uma experiência viva e navegável.

Uma CLI tradicional é adequada para automação, testes e comandos pontuais, mas fica limitada para observabilidade operacional contínua. O projeto é local-first no MVP, com Go, Postgres, Event Store e GitHub-first como restrições já aprovadas.

O framework escolhido para a implementação da TUI é o Bubble Tea.

### 2.2 Decisão

O MVP deve evoluir da CLI como interface humana principal para uma TUI local.

A TUI passa a ser a superfície primária para uso humano local em:

- dashboard de tasks, work units e runs;
- live view de eventos e sessões de agente;
- aprovações e negações de ferramentas;
- inspeção de checkpoints, evidências e falhas;
- comandos guiados para criar tasks, iniciar runs e acompanhar validações.

A CLI atual não deve ser removida imediatamente. Ela deve permanecer como camada headless para automação, testes, scripts e operação por CI/local shell. A TUI deve reutilizar serviços internos e contratos compartilhados, não chamar comandos shell da CLI como integração principal.

O framework definido para implementação é:

- **Bubble Tea + Bubbles + Lip Gloss**, da Charmbracelet.

Motivos da escolha:

- é Go-native e combina com a stack aprovada;
- usa arquitetura baseada em estado e mensagens, adequada para eventos, runs e sessões;
- facilita testes de update/model sem terminal real;
- tem componentes reutilizáveis para listas, tabelas, forms, spinners, viewport e keybindings;
- permite evoluir telas ricas sem introduzir TypeScript, Electron ou Web antes da hora.

### 2.3 Consequências (TUI)

- A interface humana fica mais adequada para operação contínua de agentes.
- O núcleo do produto continua local-first e Go-first.
- A CLI deixa de ser o centro da experiência humana, mas segue importante como contrato automatizável.
- O projeto precisará separar melhor serviços de aplicação da camada `cmd`, para que CLI e TUI compartilhem a mesma lógica.
- A TUI deve respeitar a autonomia M0 aprovada: sugestão e execução com revisão humana, sem operar acima de Nível 2.

O desenvolvimento começará com um protótipo para validar:

- renderizar lista de tasks/runs/eventos com dados reais do Postgres;
- navegar sem bloquear leitura de eventos;
- aceitar entradas de comando guiadas;
- testar transições de tela e estado sem terminal real;
- avaliar ergonomia de tabelas, filtros e live log.

### 2.4 Alternativas consideradas (TUI)

- **Bubble Tea / Bubbles / Lip Gloss**: escolhido por oferecer melhor alinhamento com Go, facilidade de testes de update/model e arquitetura orientada a eventos.
- **tview / tcell**: entrega CRUD e layouts rapidamente, com componentes prontos. Descartada por ser menos elegante para fluxos complexos por mensagens.
- **termui**: simples para dashboards, mas menos adequada para uma aplicação operacional com forms, navegação e estados ricos.
- **TUI em Rust com ratatui**: excelente qualidade técnica, mas cria uma segunda stack e aumenta custo do MVP.
- **Web ou Desktop agora**: continuam adiados; oferecem mais riqueza visual, mas aumentam escopo antes de estabilizar o control plane.
- **Apenas CLI tradicional**: preserva simplicidade, mas não resolve bem live view, aprovações e acompanhamento contínuo.

---

## Apêndice A: Histórico de Evolução

| Data | Evento | ADR Original |
| --- | --- | --- |
| 2026-05-10 | Interface inicial definida (scripts → CLI → GitHub) | ADR 0005 |
| 2026-05-10 | TUI elevada a interface local principal | ADR 0015 |
| 2026-05-17 | Ambos consolidados neste documento único | — |

## Apêndice B: Alternativas Consideradas (Interface Inicial)

- **Apenas scripts**: rápido para protótipo, mas fraco como contrato operacional e difícil de padronizar.
- **Desktop desde o início**: atraente para experiência local, mas adiciona custo antes de validar o control plane.
- **Web desde o início**: bom para colaboração futura, mas prematuro para um MVP local-first.
- **Chat como interface principal**: prático para conversa, mas limitado para auditoria, operação detalhada, replay e segurança.
