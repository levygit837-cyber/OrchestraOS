# 0022. LLM-Optimized Module Architecture (Vertical Slices)

**Data:** 2026-05-11

## 1. Contexto

O OrchestraOS tem como premissa ser um sistema mantido, evoluído e operado primariamente por Agentes de IA (LLMs). Para que um LLM seja eficaz, ele precisa receber o **contexto correto** na sua janela de leitura, com o mínimo de ruído (informações irrelevantes) possível.

Atualmente, o projeto utiliza uma **Arquitetura em Camadas (Layered Architecture)**, separando o código por preocupação técnica (Technical Cohesion):
- `internal/domain/` (Modelos e tipos para todo o sistema)
- `internal/services/` (Regras de negócio para todo o sistema)
- `internal/repository/` (Acesso a dados para todo o sistema)

O problema estrutural para LLMs é que, ao executar uma tarefa relacionada a uma única entidade (ex: alterar o status de uma `Task`), o agente precisa ler arquivos de 3 a 4 diretórios distintos. Como esses arquivos contêm definições de múltiplas outras entidades (ex: `domain/types.go` contém `Task`, `WorkUnit`, `AgentSession`, etc.), a janela de contexto do LLM é preenchida com informações que ele não precisa.

Isso causa três problemas centrais:
1. **Desperdício de Tokens:** Gastamos tokens enviando código irrelevante para o LLM ler.
2. **Alta Carga Cognitiva e Alucinação:** Quanto mais código desnecessário o LLM lê, maior a chance dele se confundir ou gerar código que afeta outras entidades sem querer.
3. **Quebra de Contratos:** Como as lógicas estão fisicamente distantes, o LLM frequentemente altera um modelo no `domain` mas esquece de atualizar a interface correspondente no `repository`.

## 2. Decisão

Adotaremos uma arquitetura baseada em **Vertical Slices** (ou Módulos Verticais), otimizada para agentes de IA. A partir de agora, o projeto será estruturado por **Domínio de Negócio (Feature/Module Cohesion)**, e não mais por camada técnica.

1. **Estrutura Base:** 
   O código das entidades será migrado para `internal/modules/<nome_da_entidade>/` (ou `internal/features/`).
   Dentro dessa pasta, todas as camadas daquela funcionalidade residirão juntas (tipos, serviços, repositórios).
   
   Exemplo:
   ```text
   internal/modules/task/
     models.go       (Tipos de Task)
     service.go      (Regras de negócio de Task)
     repository.go   (Persistência de Task)
     events.go       (Eventos gerados pelo domínio Task)
   ```

2. **Isolamento Estrito (Regra de Ouro):**
   Um módulo em `internal/modules/*` **NÃO PODE** importar arquivos de outro módulo em `internal/modules/*`. Eles devem ser complementamente autônomos.
   Se precisarem se comunicar, devem fazê-lo de forma assíncrona (via `core/eventstore`) ou ser orquestrados por uma camada de aplicação superior (`cmd/` ou `internal/orchestration/`), que os interliga usando dependências puras (ex: IDs genéricos em vez de structs concretas).

## 3. Consequências

- **Aumento de Eficiência do LLM:** Para dar manutenção em um módulo, o Agente de IA só precisará listar e ler o conteúdo de *uma única pasta*. O contexto será 100% focado e extremamente econômico em tokens.
- **Maior Escalabilidade do Projeto:** Em vez de arquivos monstruosos e pastas de serviços com dezenas de arquivos interligados, o crescimento do projeto será linear. Nova funcionalidade = Nova pasta no `modules/`. Nenhuma outra parte do código será impactada por colisões de arquivos.
- **Facilidade de Extração:** Se no futuro decidirmos que o módulo de `Task` deve ser um microsserviço independente (separado do `Orchestrator`), a extração é trivial: basta mover a pasta `task` para outro repositório.
- **Refatoração Inicial:** Haverá um esforço inicial de engenharia (via LLMs) para migrar a estrutura atual das pastas `domain`, `services` e `repository` para os módulos verticais, resolvendo possíveis dependências circulares.

## 4. Alternativas Consideradas

A decisão por *Vertical Slices* (Módulos Verticais) foi tomada após comparar como as diferentes arquiteturas se comportam quando **o desenvolvedor principal é um LLM**:

### Alternativa A: Clean Architecture / Layered Architecture (Nossa estrutura atual)
- **Como funciona:** Separa o código por camadas técnicas (Ports & Adapters, Services, Repositories).
- **Escalabilidade:** Escala moderadamente bem para humanos, pois padroniza onde encontrar cada tipo técnico de arquivo. Porém, arquivos compartilhados (como um `types.go` global) tornam-se gargalos.
- **Para LLMs:** É a **pior opção**. Um LLM lê código baseando-se em requisições de *funcionalidades*. Para adicionar o recurso de "arquivamento de tarefas", o LLM tem que editar 4 lugares distintos, lendo dezenas de arquivos de contexto alheio. O risco de quebrar o estado global e a ineficiência de prompt a descartam.

### Alternativa B: Microsserviços
- **Como funciona:** Cada funcionalidade é um projeto isolado rodando em seu próprio processo/contêiner.
- **Escalabilidade:** Escalação máxima de infraestrutura e equipe.
- **Para LLMs:** É excelente em termos de isolamento de contexto (um LLM só vê o código do microsserviço). No entanto, o LLM precisaria entender e versionar contratos de rede, chamadas gRPC/REST, Dockerfiles complexos e tratamento de falhas distribuídas. Para a fase MVP do OrchestraOS, é um *over-engineering* pesado que travaria a velocidade do Agente.

### A Vencedora: Vertical Slice Architecture (Modular Monolith)
- **Como funciona:** Mantém o projeto como um monólito (fácil de compilar, testar e debugar localmente), mas divide o código internamente como se fossem microsserviços lógicos (Módulos).
- **Escalabilidade:** É a arquitetura mais **sustentável** e escalável a longo prazo para o nosso momento. Cada nova funcionalidade apenas adiciona uma "fatia" (pasta) nova ao bolo, sem inchar as camadas existentes.
- **Para LLMs:** Junta "o melhor dos dois mundos". Fornece o isolamento cirúrgico de contexto (o Agente só lê a pasta específica) sem a sobrecarga operacional de gerenciar rede e infraestrutura de microsserviços. É, disparada, a arquitetura ideal para codebases totalmente operados por IA.
