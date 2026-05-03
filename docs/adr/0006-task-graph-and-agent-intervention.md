# ADR 0006: Decomposição de Tasks e Intervenção em Agentes

## Contexto

O fluxo desejado comeca em uma mensagem humana para o Orchestrator. O Orchestrator deve entender a intencao, quebrar o trabalho em partes menores, montar prompts, preparar sandboxes, iniciar agentes, observar traces e intervir quando necessario.

Foi discutida a ideia de um "DAG cyclic". Tecnicamente, um DAG e um grafo direcionado aciclico. O projeto precisa de dependencias aciclicas entre unidades de trabalho, mas os agentes terao loops internos de execucao, checkpoints, validacao e correcao de rota.

Tambem ha necessidade de conversar com o Orchestrator e, quando apropriado, interferir no andamento de um agente especifico sem quebrar auditoria e isolamento.

## Decisão

O Orchestrator usara um **Task Graph aciclico** para decompor trabalho em unidades menores.

- O grafo de tasks deve ser um DAG.
- Ciclos de tentativa, validacao e correcao acontecem dentro de cada `Run` ou `AgentLoop`, nao como dependencias ciclicas entre nodes.
- Cada node do grafo representa uma `WorkUnit` executavel por agente ou humano.
- Edges representam dependencias, artefatos exigidos, ordem de execucao ou restricoes de conflito.
- O Orchestrator deve atribuir ownership de arquivos, modulos ou dominios para reduzir conflito entre agentes paralelos.

Intervencoes humanas devem passar pelo Orchestrator:

- Mensagem ao Orchestrator cria evento auditavel.
- Mensagem para agente especifico vira comando ou notificacao mediada pelo Orchestrator.
- Aprovacoes, negacoes, pausas e cancelamentos viram eventos persistidos.
- Agentes nao devem abrir comunicacao direta entre si sem mediacao do Orchestrator.

## Consequências

- O sistema diferencia planejamento aciclico de execucao iterativa.
- Fica mais simples detectar bloqueios, paralelizar nodes independentes e evitar conflitos de arquivos.
- Intervencoes diretas continuam possiveis, mas com trilha de auditoria.
- O Orchestrator precisa manter uma live view dos traces e saber quando pausar, avisar ou corrigir uma run.

## Alternativas consideradas

- **Grafo ciclico de tasks**: expressivo, mas dificulta agendamento, conclusao, rollback e explicabilidade.
- **Lista linear de subtasks**: simples, mas limita paralelismo e dependencia entre artefatos.
- **Agentes decidindo livremente a decomposicao**: flexivel, mas reduz controle, previsibilidade e isolamento.
- **Intervencao direta no processo do agente**: rapida, mas fraca para auditoria e recuperacao.
