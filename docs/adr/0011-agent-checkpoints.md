# ADR 0011: Agent Checkpoints

## Contexto

Agentes paralelos precisam executar work units sem carregar contexto indefinidamente. O projeto quer preservar continuidade e auditoria, mas tambem reduzir deriva de contexto, mistura de objetivos e dependencia de memoria conversacional.

Foi considerada uma estrategia mais ampla com repouso por aprovacao de ferramentas, sub-sessoes e cache frio de arquivos. Para o momento, a decisao e documentar apenas o mecanismo de `AgentCheckpoint`.

## Decisão

O OrchestraOS tera `AgentCheckpoint` como fronteira persistente de progresso dentro de uma `AgentSession`.

Um checkpoint representa um ponto seguro em que o agente registra:

- goal atual;
- goals concluidos;
- todos pendentes;
- arquivos lidos;
- arquivos modificados;
- evidencias geradas;
- decisoes locais;
- bloqueios;
- riscos;
- resumo minimo necessario para continuar;
- proximo goal sugerido.

O agente deve emitir checkpoints em momentos naturais:

- ao concluir um goal curto;
- antes de mudar de foco;
- antes de validar;
- depois de produzir diff relevante;
- antes de encerrar uma work unit;
- quando o Orchestrator solicitar explicitamente.

O Orchestrator usa checkpoints para:

- reconstruir progresso da run;
- detectar deriva de objetivo;
- decidir se a work unit terminou;
- preparar continuacao futura com contexto menor;
- revisar evidencias sem depender do transcript completo.

## Fora Desta Decisão

Esta ADR nao aprova neste momento:

- espera temporizada por aprovacao de ferramenta;
- hibernacao automatica por tool request;
- sub-sessoes;
- cache frio de arquivos;
- troca automatica de contexto durante espera.

Esses temas podem ser reavaliados depois que checkpoints, ledger e Event Store estiverem funcionando.

## Consequências

- O sistema ganha pontos de recuperacao e auditoria sem introduzir scheduler complexo.
- Agentes podem trabalhar em ciclos curtos com menos risco de perder objetivo.
- O Event Store precisa persistir checkpoints como eventos consultaveis.
- O Agent Task Ledger continua sendo a memoria operacional viva; checkpoint e snapshot de progresso em um momento especifico.
- Futuras continuacoes de sessao poderao usar checkpoints, mas isso nao entra no primeiro escopo.

## Alternativas consideradas

- **Sem checkpoints estruturados**: simples, mas deixa o historico dependente de logs e transcript.
- **Sub-sessoes desde o inicio**: melhora isolamento, mas aumenta complexidade antes de validar o loop basico.
- **Checkpoint apenas textual**: facil de gerar, mas fraco para consulta, validacao e retomada.
- **Checkpoint em todo todo pequeno**: excesso de eventos e overhead operacional.
