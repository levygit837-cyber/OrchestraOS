# ADR 0007: Ciclo Operacional do Agente — Prompts, Ledger e Checkpoints

**Status:** Consolidated (absorve: ADR 0008, ADR 0011)  
**Data original:** 2026-05-10  
**Última atualização:** 2026-05-17

---

## Contexto

O Orchestrator precisa gerar prompts precisos para agentes diferentes, tarefas diferentes, níveis de risco diferentes e modos operacionais diferentes. Um único prompt monolítico tende a ficar rígido, conflituoso e difícil de testar.

Agentes também podem perder o fio da task durante execuções longas, loops de validação ou troca de contexto. O projeto precisa de uma forma persistente de manter objetivo, critérios de aceite, pendências, bloqueios e evidências sem transformar memória temporária em decisão definitiva.

Agentes paralelos precisam executar work units sem carregar contexto indefinidamente. O projeto quer preservar continuidade e auditoria, mas também reduzir deriva de contexto, mistura de objetivos e dependência de memória conversacional.

---

## 1. Composição de Prompts

### 1.1 Decisão

O OrchestraOS terá um **Prompt Composition System** baseado em fragmentos versionados, com suporte futuro a especialização dinâmica controlada por run.

O Orchestrator montará dois artefatos principais por `WorkUnit`:

- `SystemPrompt`: define persona operacional, limites, políticas, ferramentas, checkpoints e contratos de saída.
- `TaskPrompt`: define objetivo específico, contexto da task, arquivos ou domínios sob responsabilidade, critérios de aceite, validações esperadas e evidências exigidas.

Fragmentos de prompt devem ser pequenos, versionados e classificados por função:

- política global e autonomia;
- instruções do repositório;
- persona ou papel do agente;
- modo operacional;
- domínio técnico;
- regras de ferramentas;
- contrato de comunicação;
- formato de saída;
- validação;
- registro de progresso e todo persistente.

O Orchestrator deve criar ou referenciar um `PromptSnapshot` para cada run, registrando fragmentos usados, versões, ordem de montagem, variáveis aplicadas e hash do resultado. A partir do corte M3, snapshots idênticos são deduplicados por `composition_hash`; a rastreabilidade por run continua no evento `prompt.snapshot_created`, que informa se o snapshot foi reutilizado e o `count_used` atual.

### 1.2 Especialização Dinâmica Controlada

Além dos fragmentos estáticos, o Orchestrator poderá criar `DynamicPromptFragment` temporário quando uma `WorkUnit` fugir dos perfis e fragmentos pré-definidos.

Essa flexibilidade fica aprovada como direção futura, mas não como requisito obrigatório do primeiro corte do MVP.

Regras:

- Personas principais continuam pré-definidas.
- Fragmentos dinâmicos só podem especializar contexto, estratégia, heurísticas e instruções de domínio da run.
- Fragmentos dinâmicos não podem sobrescrever política global, autonomia, segurança, ownership, permissões, critérios de aceite ou contrato de saída.
- Todo fragmento dinâmico deve ser salvo no `PromptSnapshot`, com motivo, autor, hash e validade limitada à run.
- Quando um fragmento dinâmico se repetir em várias runs, ele deve virar candidato a fragmento estático versionado.

### 1.3 Toolset Por Run

O Orchestrator também poderá selecionar um toolset mínimo para cada `AgentSession`.

Objetivo:

- reduzir contexto;
- reduzir superfície de erro;
- limitar permissões ao escopo real da work unit;
- tornar tool use mais auditável.

Se o agente precisar de uma ferramenta fora do toolset, ele deve emitir um pedido estruturado. O Orchestrator pode negar, aprovar uma expansão limitada, criar nova work unit especializada ou reiniciar a sessão com novo prompt/toolset.

### 1.4 Reconfiguração De AgentSession

Quando for necessário alterar ferramentas, persona operacional ou fragmentos dinâmicos durante uma run, o sistema deve fazer uma reconfiguração auditável.

Essa reconfiguração deve:

- preservar o `AgentTaskLedger`;
- criar novo `PromptSnapshot`;
- criar novo `ToolsetSnapshot`;
- registrar motivo e decisão;
- encerrar ou reiniciar a sessão anterior de forma controlada;
- manter rastreabilidade entre sessão anterior e sessão nova.

O agente não pode modificar seu próprio toolset diretamente. Ele pode apenas solicitar mudança ao Orchestrator.

### 1.5 Consequências (Prompts)

- Prompts ficam reutilizáveis, auditáveis e testáveis.
- O Orchestrator ganha flexibilidade para montar agentes especializados sem duplicar instruções.
- Conflitos entre fragmentos precisam ser tratados por metadados como prioridade, incompatibilidades e requisitos.
- Prompts passam a ser parte do domínio do produto, não apenas texto solto.
- Toolsets por run reduzem contexto e permissões desnecessárias.
- Especialização dinâmica aumenta flexibilidade, mas exige snapshots, validação e trilha de auditoria.
- Reconfiguração de sessão aumenta complexidade e deve entrar depois do fluxo estático estar funcionando.

### 1.6 Alternativas consideradas (Prompts)

- **Prompt único por agente**: simples no início, mas rígido e difícil de manter.
- **Prompt livre gerado totalmente pelo modelo**: flexível, mas pouco auditável e difícil de reproduzir.
- **Templates estáticos por tipo de task**: melhor que prompt único, mas ainda limita combinações e especializações.
- **Ferramentas completas para todos os agentes**: simples de operar, mas aumenta contexto, permissões e risco.
- **Agente autoaltera suas ferramentas**: flexível, mas viola controle, auditoria e política de autonomia.

---

## 2. Ledger de Progresso

### 2.1 Contexto adicional

Foi considerada a ideia de um sistema de todos persistente para os prompts das tasks.

### 2.2 Decisão

O MVP incluirá um **Agent Task Ledger** persistente por `WorkUnit`.

O ledger deve conter:

- objetivo da work unit;
- critérios de aceite;
- escopo autorizado;
- lista de todos;
- itens concluídos com evidência;
- bloqueios;
- riscos identificados;
- próximo checkpoint esperado;
- resumo curto do estado atual.

O agente deve consultar e atualizar o ledger em checkpoints. O Orchestrator deve usar o ledger para detectar progresso parado, divergência de objetivo, loops e pendências antes de considerar uma task concluída.

O ledger não substitui ADR, issue, PR, Event Store ou documentação versionada. Ele é memória operacional da run.

Checkpoints são snapshots persistidos de progresso em pontos seguros. O ledger representa o estado operacional vivo; o checkpoint registra um momento específico desse estado com evidências, arquivos tocados e próximo goal sugerido.

### 2.3 Consequências (Ledger)

- Agentes ganham continuidade durante execuções longas.
- O Orchestrator tem uma visão objetiva do progresso além do texto livre do chat.
- A conclusão da task pode exigir todos resolvidos, justificativa para pendências e evidências registradas.
- Checkpoints permitem reconstruir progresso sem depender do transcript completo.
- O sistema precisa evitar que o ledger vire um backlog paralelo sem revisão.

### 2.4 Alternativas consideradas (Ledger)

- **Sem todos persistentes**: reduz escopo, mas aumenta risco de deriva e esquecimento.
- **Todos apenas no prompt**: simples, mas se perde em contexto longo e não gera auditoria estruturada.
- **Backlog completo dentro do agente**: poderoso, mas mistura planejamento de produto com execução operacional.

---

## 3. Checkpoints

### 3.1 Contexto adicional

Foi considerada uma estratégia mais ampla com repouso por aprovação de ferramentas, sub-sessões e cache frio de arquivos. Para o momento, a decisão é documentar apenas o mecanismo de `AgentCheckpoint`.

### 3.2 Decisão

O OrchestraOS terá `AgentCheckpoint` como fronteira persistente de progresso dentro de uma `AgentSession`.

Um checkpoint representa um ponto seguro em que o agente registra:

- goal atual;
- goals concluídos;
- todos pendentes;
- arquivos lidos;
- arquivos modificados;
- evidências geradas;
- decisões locais;
- bloqueios;
- riscos;
- resumo mínimo necessário para continuar;
- próximo goal sugerido.

O agente deve emitir checkpoints em momentos naturais:

- ao concluir um goal curto;
- antes de mudar de foco;
- antes de validar;
- depois de produzir diff relevante;
- antes de encerrar uma work unit;
- quando o Orchestrator solicitar explicitamente.

`AgentSessionService` é a fronteira canônica para persistir checkpoints. Runtimes e interfaces devem encaminhar sinais ou eventos de checkpoint para o service, em vez de atualizar `last_checkpoint_at` diretamente via repositório.

A política inicial de checkpoint automático considera pontos seguros:

- checkpoint emitido pelo runtime;
- goal concluído;
- mudança de foco;
- início de validação;
- diff relevante produzido;
- pedido de ferramenta;
- uso ou conclusão bem-sucedida de ferramenta sem pedido prévio de aprovação;
- preparação para conclusão;
- timeout com estado recuperável.

O comando manual de checkpoint da CLI permanece apenas como mecanismo de debug/teste e não deve ser o caminho operacional principal.

O Orchestrator usa checkpoints para:

- reconstruir progresso da run;
- detectar deriva de objetivo;
- decidir se a work unit terminou;
- preparar continuação futura com contexto menor;
- revisar evidências sem depender do transcript completo.

### 3.3 Fora Desta Decisão

Esta ADR não aprova neste momento:

- espera temporizada por aprovação de ferramenta;
- hibernação automática por tool request;
- sub-sessões;
- cache frio de arquivos;
- troca automática de contexto durante espera.

Esses temas podem ser reavaliados depois que checkpoints, ledger e Event Store estiverem funcionando.

### 3.4 Consequências (Checkpoints)

- O sistema ganha pontos de recuperação e auditoria sem introduzir scheduler complexo.
- Agentes podem trabalhar em ciclos curtos com menos risco de perder objetivo.
- O Event Store precisa persistir checkpoints como eventos consultáveis.
- Cada checkpoint atualiza `last_checkpoint_at`, `last_seen_event_id` e `recoverable_state` da `AgentSession` na mesma transação do evento.
- Checkpoints de uma sessão podem ser listados em ordem de `sequence` para reconstruir o progresso e recuperar o último estado continuável.
- O Agent Task Ledger continua sendo a memória operacional viva; checkpoint é snapshot de progresso em um momento específico.
- Futuras continuações de sessão poderão usar checkpoints, mas isso não entra no primeiro escopo.

### 3.5 Alternativas consideradas (Checkpoints)

- **Sem checkpoints estruturados**: simples, mas deixa o histórico dependente de logs e transcript.
- **Sub-sessões desde o início**: melhora isolamento, mas aumenta complexidade antes de validar o loop básico.
- **Checkpoint apenas textual**: fácil de gerar, mas fraco para consulta, validação e retomada.
- **Checkpoint em todo todo pequeno**: excesso de eventos e overhead operacional.

---

## Apêndice A: Histórico de Evolução

| Data | Evento | ADR Original |
| --- | --- | --- |
| 2026-05-10 | Prompt Composition System definido | ADR 0007 |
| 2026-05-10 | Agent Task Ledger definido | ADR 0008 |
| 2026-05-10 | Agent Checkpoints definidos | ADR 0011 |
| 2026-05-17 | Ambos consolidados neste documento único | — |

## Apêndice B: Distinção Ledger vs Checkpoint

| Aspecto | Ledger | Checkpoint |
|---------|--------|------------|
| Natureza | Memória operacional viva | Snapshot persistido em momento específico |
| Atualização | Contínua (em checkpoints) | Em pontos seguros definidos |
| Conteúdo | Estado atual, todos, bloqueios, riscos | Evidências, arquivos tocados, próximo goal |
| Propósito | Orientar agente durante execução | Recuperar progresso e auditar |
