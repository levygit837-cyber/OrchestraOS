# ADR 0007: Sistema de Composição de Prompts

## Contexto

O Orchestrator precisa gerar prompts precisos para agentes diferentes, tarefas diferentes, niveis de risco diferentes e modos operacionais diferentes. Um unico prompt monolitico tende a ficar rigido, conflituoso e dificil de testar.

O projeto tambem precisa registrar qual prompt foi usado em cada run para permitir auditoria, comparacao de resultados e melhoria continua.

## Decisão

O OrchestraOS tera um **Prompt Composition System** baseado em fragmentos versionados, com suporte futuro a especializacao dinamica controlada por run.

O Orchestrator montara dois artefatos principais por `WorkUnit`:

- `SystemPrompt`: define persona operacional, limites, politicas, ferramentas, checkpoints e contratos de saida.
- `TaskPrompt`: define objetivo especifico, contexto da task, arquivos ou dominios sob responsabilidade, criterios de aceite, validacoes esperadas e evidencias exigidas.

Fragmentos de prompt devem ser pequenos, versionados e classificados por funcao:

- politica global e autonomia;
- instrucoes do repositorio;
- persona ou papel do agente;
- modo operacional;
- dominio tecnico;
- regras de ferramentas;
- contrato de comunicacao;
- formato de saida;
- validacao;
- registro de progresso e todo persistente.

O Orchestrator deve criar ou referenciar um `PromptSnapshot` para cada run, registrando fragmentos usados, versoes, ordem de montagem, variaveis aplicadas e hash do resultado. A partir do corte M3, snapshots identicos sao deduplicados por `composition_hash`; a rastreabilidade por run continua no evento `prompt.snapshot_created`, que informa se o snapshot foi reutilizado e o `count_used` atual.

## Especialização Dinâmica Controlada

Além dos fragmentos estáticos, o Orchestrator poderá criar `DynamicPromptFragment` temporário quando uma `WorkUnit` fugir dos perfis e fragmentos pré-definidos.

Essa flexibilidade fica aprovada como direção futura, mas não como requisito obrigatório do primeiro corte do MVP.

Regras:

- Personas principais continuam pré-definidas.
- Fragmentos dinâmicos só podem especializar contexto, estratégia, heurísticas e instruções de domínio da run.
- Fragmentos dinâmicos não podem sobrescrever política global, autonomia, segurança, ownership, permissões, critérios de aceite ou contrato de saída.
- Todo fragmento dinâmico deve ser salvo no `PromptSnapshot`, com motivo, autor, hash e validade limitada à run.
- Quando um fragmento dinâmico se repetir em várias runs, ele deve virar candidato a fragmento estático versionado.

## Toolset Por Run

O Orchestrator também poderá selecionar um toolset mínimo para cada `AgentSession`.

Objetivo:

- reduzir contexto;
- reduzir superfície de erro;
- limitar permissões ao escopo real da work unit;
- tornar tool use mais auditável.

Se o agente precisar de uma ferramenta fora do toolset, ele deve emitir um pedido estruturado. O Orchestrator pode negar, aprovar uma expansão limitada, criar nova work unit especializada ou reiniciar a sessão com novo prompt/toolset.

## Reconfiguração De AgentSession

Quando for necessário alterar ferramentas, persona operacional ou fragmentos dinâmicos durante uma run, o sistema deve fazer uma reconfiguração auditável.

Essa reconfiguração deve:

- preservar o `AgentTaskLedger`;
- criar novo `PromptSnapshot`;
- criar novo `ToolsetSnapshot`;
- registrar motivo e decisão;
- encerrar ou reiniciar a sessão anterior de forma controlada;
- manter rastreabilidade entre sessão anterior e sessão nova.

O agente não pode modificar seu próprio toolset diretamente. Ele pode apenas solicitar mudança ao Orchestrator.

## Consequências

- Prompts ficam reutilizaveis, auditaveis e testaveis.
- O Orchestrator ganha flexibilidade para montar agentes especializados sem duplicar instrucoes.
- Conflitos entre fragmentos precisam ser tratados por metadados como prioridade, incompatibilidades e requisitos.
- Prompts passam a ser parte do dominio do produto, nao apenas texto solto.
- Toolsets por run reduzem contexto e permissões desnecessárias.
- Especialização dinâmica aumenta flexibilidade, mas exige snapshots, validação e trilha de auditoria.
- Reconfiguração de sessão aumenta complexidade e deve entrar depois do fluxo estático estar funcionando.

## Alternativas consideradas

- **Prompt unico por agente**: simples no inicio, mas rigido e dificil de manter.
- **Prompt livre gerado totalmente pelo modelo**: flexivel, mas pouco auditavel e dificil de reproduzir.
- **Templates estaticos por tipo de task**: melhor que prompt unico, mas ainda limita combinacoes e especializacoes.
- **Ferramentas completas para todos os agentes**: simples de operar, mas aumenta contexto, permissões e risco.
- **Agente autoaltera suas ferramentas**: flexível, mas viola controle, auditoria e política de autonomia.
