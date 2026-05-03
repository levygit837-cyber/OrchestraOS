# Estratégia de Interface

## Decisão Prática

O MVP começa com scripts de bootstrap e evolui rapidamente para uma CLI fina. GitHub e a superficie externa principal. Chat entra apenas como conector opcional futuro. Desktop e Web ficam para depois da validação do nucleo.

## Camadas

### Scripts de Bootstrap

Uso:

- subir banco local;
- rodar migrations;
- iniciar Orchestrator;
- iniciar agente fake;
- limpar worktrees temporarios;
- executar cenarios de teste.

Scripts sao internos ao desenvolvimento. Eles podem mudar sem compatibilidade.

### CLI do MVP

Uso:

- criar task a partir de texto;
- iniciar run;
- listar tasks e runs;
- acompanhar eventos;
- enviar mensagem para Orchestrator;
- enviar mensagem mediada para agente;
- aprovar ou negar tool request;
- pausar, retomar ou cancelar run;
- coletar diff e evidencias.

Comandos conceituais:

```text
orchestra task create --message "..."
orchestra task plan <task_id>
orchestra run start <work_unit_id>
orchestra run watch <run_id>
orchestra run message <run_id> --to orchestrator --text "..."
orchestra agent message <agent_session_id> --text "..."
orchestra tool approve <tool_request_id>
orchestra tool deny <tool_request_id> --reason "..."
orchestra review diff <run_id>
```

### GitHub

Uso:

- registrar issues quando houver backlog externo;
- criar branches e pull requests;
- revisar diffs e evidencias;
- rodar checks;
- manter historico de integracao;
- controlar merge.

GitHub nao substitui o Event Store, mas e o principal registro externo de revisao e integracao.

### Chat Opcional Futuro

Slack, Discord ou e-mail podem ser adicionados depois para:

- captura rapida;
- avisos;
- notificacoes;
- resumos;
- pedidos simples de status.

Chat nao e dependencia do MVP e nao deve ser a unica forma de aprovar acoes sensiveis.

### Desktop Futuro

Bom candidato quando o produto precisar de:

- live view visual de agents e traces;
- intervencao em threads de agentes;
- revisao rica de DAG, diff e artefatos;
- experiencia local para operador solo.

### Web Futuro

Bom candidato quando o produto precisar de:

- servidor remoto;
- equipe;
- permissao por usuario;
- painel compartilhado;
- auditoria e historico acessiveis de varios dispositivos.

## Critério Para Mudar de Camada

Desktop, Web ou conectores de chat so devem virar prioridade quando pelo menos estes fluxos funcionarem por CLI/GitHub:

- criar task;
- gerar task graph;
- iniciar agente;
- receber eventos;
- aprovar ou negar ferramenta;
- concluir run com evidencias;
- revisar diff.
