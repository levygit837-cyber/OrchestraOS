# ADR 0002: Orchestrator Como Control Plane

## Contexto

O OrchestraOS precisa executar multiplos agentes em paralelo. Cada agente tera contexto isolado, sandbox proprio e worktree separado por task.

Agentes podem precisar trocar informacoes durante a execucao, mas o projeto tambem exige fonte de verdade versionada, trilha de auditoria, politicas de autonomia, aprovacao de ferramentas e possibilidade de rollback.

Uma malha livre de agentes conversando diretamente reduziria controle, dificultaria auditoria e aumentaria risco de vazamento de contexto entre sandboxes.

## Decisão

O OrchestraOS usara um Orchestrator central como **control plane**.

O Orchestrator sera responsavel por:

- Criar e administrar tasks.
- Agendar execucoes.
- Criar sandboxes e worktrees.
- Iniciar e encerrar agentes.
- Receber eventos estruturados.
- Controlar aprovacao de ferramentas.
- Mediar comunicacao entre agentes.
- Persistir auditoria.
- Reportar status ao GitHub e à CLI; conectores de chat podem ser adicionados depois.

Agentes serao tratados como workers isolados. Eles podem solicitar informacoes de outros agentes, mas a comunicacao deve passar pelo Orchestrator.

O canal vivo entre agente e Orchestrator sera WebSocket. Eventos e comandos relevantes deverao ser persistidos no Event Store.

## Consequências

- O sistema ganha controle, auditabilidade e aplicacao consistente de politicas.
- Agentes permanecem isolados por task e contexto.
- Comunicacao entre agentes fica mais segura e rastreavel.
- O Orchestrator vira componente critico e precisa ser simples, testavel e observavel.
- A arquitetura reduz flexibilidade espontanea entre agentes em troca de seguranca operacional.

## Alternativas Consideradas

- **Agentes autonomos peer-to-peer**: mais flexivel, mas fraco para auditoria, controle de permissoes e isolamento de contexto.
- **Framework multiagente como nucleo principal**: acelera prototipo, mas amarra o dominio do produto a decisoes de um framework externo.
- **Workflow engine desde o inicio**: util para execucoes duraveis, mas adiciona complexidade antes de validar o MVP.
- **Chat como coordenador central**: pratico para conversa, mas inadequado como fonte de verdade e trilha tecnica.
