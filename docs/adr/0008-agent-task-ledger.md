# ADR 0008: Ledger Persistente de Progresso dos Agentes

## Contexto

Agentes podem perder o fio da task durante execucoes longas, loops de validacao ou troca de contexto. O projeto precisa de uma forma persistente de manter objetivo, criterios de aceite, pendencias, bloqueios e evidencias sem transformar memoria temporaria em decisao definitiva.

Foi considerada a ideia de um sistema de todos persistente para os prompts das tasks.

## Decisão

O MVP incluira um **Agent Task Ledger** persistente por `WorkUnit`.

O ledger deve conter:

- objetivo da work unit;
- criterios de aceite;
- escopo autorizado;
- lista de todos;
- itens concluidos com evidencia;
- bloqueios;
- riscos identificados;
- proximo checkpoint esperado;
- resumo curto do estado atual.

O agente deve consultar e atualizar o ledger em checkpoints. O Orchestrator deve usar o ledger para detectar progresso parado, divergencia de objetivo, loops e pendencias antes de considerar uma task concluida.

O ledger nao substitui ADR, issue, PR, Event Store ou documentacao versionada. Ele e memoria operacional da run.

Checkpoints sao snapshots persistidos de progresso em pontos seguros. O ledger representa o estado operacional vivo; o checkpoint registra um momento especifico desse estado com evidencias, arquivos tocados e proximo goal sugerido.

## Consequências

- Agentes ganham continuidade durante execucoes longas.
- O Orchestrator tem uma visao objetiva do progresso alem do texto livre do chat.
- A conclusao da task pode exigir todos resolvidos, justificativa para pendencias e evidencias registradas.
- Checkpoints permitem reconstruir progresso sem depender do transcript completo.
- O sistema precisa evitar que o ledger vire um backlog paralelo sem revisao.

## Alternativas consideradas

- **Sem todos persistentes**: reduz escopo, mas aumenta risco de deriva e esquecimento.
- **Todos apenas no prompt**: simples, mas se perde em contexto longo e nao gera auditoria estruturada.
- **Backlog completo dentro do agente**: poderoso, mas mistura planejamento de produto com execucao operacional.
