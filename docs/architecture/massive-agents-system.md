# Proposta Futura: Massive Agents System

## Status

Esta e uma proposta futura, nao uma decisao arquitetural.

Ela nao altera o MVP atual, nao substitui as ADRs existentes e nao aprova aumento de autonomia. O limite inicial continua sendo 2 a 5 agentes paralelos, com autonomia Nivel 2 e revisao humana.

## Ideia

Massive Agents System, ou MAS, e uma proposta para permitir execucao controlada de muitos agentes em paralelo sobre um projeto grande.

O objetivo e reduzir custo, tokens, conflito e risco operacional por meio de escopos extremamente pequenos, isolados e auditaveis. Em vez de cada agente receber o repositorio inteiro ou um contexto amplo, o Orchestrator entregaria apenas a unidade de codigo, dependencias e validacoes necessarias para aquela work unit.

## Principio Central

MAS nao deve ser tratado como "um agente por linhas". Linhas sao metadados uteis, mas a unidade operacional precisa ser semantica e verificavel.

Unidades candidatas:

- funcao;
- metodo;
- struct, classe ou interface;
- arquivo pequeno;
- pacote ou modulo;
- contrato ou schema;
- teste relacionado;
- integracao isolavel;
- migracao ou camada de persistencia.

Ranges de linhas podem ser usados para delimitar leitura e escrita, mas o isolamento deve considerar imports, dependencias, chamadas, invariantes, testes e contratos publicos.

## Code Intelligence Necessaria

Antes de spawnar muitos agentes, o Orchestrator precisaria mapear o codigo com uma camada de inteligencia estatica.

Essa camada deve produzir, no minimo:

- arquivos do repositorio;
- simbolos e ranges de linhas;
- imports e dependencias;
- call graph quando aplicavel;
- ownership de arquivos, modulos e dominios;
- testes relacionados a cada unidade;
- read set permitido;
- write set permitido;
- conflitos entre work units;
- comandos de validacao por bloco;
- risco estimado da unidade;
- custo maximo de tokens, passos e tempo por agente.

Tecnologias candidatas podem incluir LSP, tree-sitter, ferramentas nativas da linguagem, analisadores estaticos e grafo de dependencias. Nenhuma tecnologia fica decidida por esta proposta.

## WorkUnitScope

Cada agente MAS deveria receber um escopo explicito e pequeno.

Exemplo conceitual:

```text
WorkUnitScope
- objective
- authorized_read_paths
- authorized_write_paths
- authorized_symbols
- authorized_line_ranges
- dependency_context
- blocked_paths
- acceptance_criteria
- validation_commands
- max_tokens
- max_steps
- timeout
- stop_policy
```

O agente deve pedir intervencao do Orchestrator quando precisar sair desse escopo. A expansao de escopo deve gerar evento, justificativa e aprovacao conforme politica.

## Fluxo Proposto

```text
Repository Snapshot
-> Code Intelligence Map
-> Dependency Graph
-> Semantic Work Unit Partitioning
-> Conflict Analysis
-> Budget Estimation
-> MAS Plan
-> Agent Spawn With Isolated WorkUnitScope
-> Minimal Agent Loop
-> Local Validation
-> Checkpoint and Ledger Update
-> Orchestrator Aggregation
-> Human Review and Merge Gate
```

## Fases De Adocao

### Fase 1: MAS Read-Only

Muitos agentes analisam blocos isolados e retornam riscos, assumptions, inconsistencias e oportunidades de melhoria.

Nao ha edicao de codigo.

### Fase 2: MAS Assisted Write

Poucos agentes podem editar work units com write set exclusivo, validacao isolada e revisao humana obrigatoria.

Essa fase ainda respeita autonomia Nivel 2.

### Fase 3: MAS Verified Parallel

Mais agentes rodam em paralelo apenas quando o sistema possui testes, checkpoints, ledger, Event Store, Policy Engine, Sandbox Manager e merge gate maduros.

### Fase 4: MAS High Fanout

Execucao com dezenas de agentes so deve existir em dominios com isolamento comprovado, validacao automatica forte, rollback claro e limites de custo.

Essa fase exige nova decisao arquitetural antes de ser implementada.

## Guardrails

- O Orchestrator continua sendo o control plane.
- Agentes nao se comunicam diretamente.
- Escopo pequeno nao implica independencia real.
- Nenhum agente deve editar fora do write set autorizado.
- Work units conflitantes nao devem rodar em paralelo.
- Conclusao exige evidencia de validacao ou justificativa.
- Resultados de agentes sao evidencias operacionais, nao fonte de verdade definitiva.
- Conversa solta ou assumptions de agentes nao viram decisao sem registro em documento apropriado.
- Limites de custo devem existir antes de aumentar fanout.
- Alto paralelismo nao deve substituir revisao humana enquanto a autonomia aprovada for Nivel 2.

## Riscos

- Explosao de custo por excesso de agentes ou loops.
- Fragmentacao artificial que remove contexto essencial.
- Falsos negativos quando uma unidade pequena depende de invariantes globais.
- Conflitos de merge por write sets mal calculados.
- Validacao local insuficiente para comportamento sistemico.
- Complexidade prematura antes de validar o MVP.

## Perguntas Abertas

- Qual ferramenta sera usada para mapear simbolos, imports e call graph por linguagem?
- Como estimar custo antes do spawn de agentes?
- Qual granularidade minima gera ganho real sem perder contexto?
- Como detectar dependencia semantica que nao aparece no grafo estatico?
- Como agregar findings de dezenas de agentes sem criar ruido operacional?
- Quais dominios do produto podem permitir MAS primeiro?

## Fora De Escopo Por Enquanto

- Implementar MAS no MVP.
- Aprovar 10, 20 ou 100 agentes em paralelo.
- Remover revisao humana.
- Aumentar autonomia para Nivel 3, 4 ou 5.
- Criar uma malha peer-to-peer entre agentes.
- Trocar o Orchestrator central por um framework multiagente externo.
