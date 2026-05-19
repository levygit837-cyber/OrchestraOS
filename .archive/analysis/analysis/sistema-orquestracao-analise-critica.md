# Análise Crítica — Sistema de Orquestração de Agentes

**Autor:** Orquestrador (auto-análise)  
**Data:** 2026-05-11  
**Status:** Validação interna  
**Regra:** Identificar falhas reais, não vender ilusões.

---

## 1. A Decomposição é Escalável?

### Veredicto: **Até 5 agentes, sim. Além disso, não.**

#### O que funciona
- Fronteiras de diretório (`TOCAR`/`EVITAR`) evitam conflitos de merge em ~90% dos casos
- Numeração serializada (`ORCH-F05-R01-A01`) permite rastrear quem fez o quê
- A arquitetura do OrchestraOS (zero imports cross-module) facilita a separação

#### O que quebra
- **Decomposição é O(n²) para o orquestrador.** Com 3 agentes, verificar dependências é trivial. Com 8 agentes, o orquestrador precisa mapear 28 potenciais interseções. Erros de dependência oculta se tornam inevitáveis.
- **Nem todo trabalho é paralelizável.** Refatorações cross-cutting (ex: mudar a interface de `EventEnvelope`) não podem ser decompostas. O sistema força uma ilusão de paralelismo onde só existe serialismo.
- **O gargalo vira o orquestrador humano.** Mesmo que eu (IA) faça a decomposição, o usuário ainda precisa validar, distribuir, e integrar. Isso não escala.

#### Limite real
Este sistema funciona para **equipes de 2-5 agentes** trabalhando em módulos bem isolados. Além disso, vira burocracia.

---

## 2. Templates de Plans são Eficazes?

### Veredicto: **Úteis como acelerador, perigosos como muleta.**

#### O que funciona
- Reduz tempo de planejamento em tarefas repetitivas ("criar módulo Go")
- Padroniza entregas — agentes sabem o que esperar
- `plans/templates/modulo-go-completo.md` é um bom ponto de partida

#### O que quebra
- **Templates engessam o pensamento.** Um agente seguindo um template pode criar arquivos que o projeto não precisa (ex: `validation.go` vazio só porque o template manda).
- **Cada tarefa é única.** Bugfixes, refatorações, otimizações de performance, investigações — nenhuma se encaixa em template. Se 30% das tarefas são atípicas, o template vira ruído.
- **"Template-driven development"** — o agente segue a estrutura sem entender o porquê, criando código ceremonial.

#### Recomendação
Mantenha **2-3 templates** no máximo. Um para "módulo novo", um para "refatoração", um para "bugfix". Nada mais. E deixe claro que o agente pode ignorar o template se justificar.

---

## 3. CHECKLISTS.md são Úteis?

### Veredicto: **O conceito é valioso. O arquivo é frágil.**

#### O que funciona
- Traz visibilidade cross-tool sem acessar histórico
- Serve como trilha de auditoria modesta
- Força o agente a pensar em etapas, não em "fazer tudo de uma vez"

#### O que quebra brutalmente
- **Não há enforcement.** O agente pode simplesmente não atualizar o checklist. Não há mecanismo que force isso. É uma convenção gentil, não uma regra.
- **Checklists desatualizam em minutos.** Um agente focado em código esquece de marcar checkbox. Ao final da tarefa, o checklist está 50% desatualizado — tornando-se mentira documentada.
- **Overhead de I/O.** A cada ciclo do Ralph Loop, o agente lê e escreve um arquivo. Em uma sessão de 2h, isso são ~20 operações de I/O que poderiam ser substituídas por... nada. O agente já sabe o que fez.

#### O verdadeiro valor
Não está no **arquivo**, mas no **comportamento** (Ralph Loop). O ciclo "ler → executar → validar → continuar" é útil independente de haver um `.md`. O checklist é apenas um espelho desse comportamento.

#### Recomendação
Mantenha os checklists, mas **não os trate como fonte de verdade**. Trate-os como diário de bordo — útil para inspeção humana, inútil para controle automatizado. O controle real vem de **testes e build**.

---

## 4. Como Garantir que Agentes Seguem as Skills?

### Veredicto: **Não é possível garantir. É impossível.**

#### A realidade
- Skills são sugestões de sistema. O agente pode ignorá-las.
- O contexto da conversa compete com a skill. Um prompt do usuário dizendo "só faz isso rápido" sobrescreve a skill.
- Não há penalidade por não seguir. O agente não "perde pontos" se pular o code-review.
- Agentes diferentes (Windsurf, Kimi-CLI, etc.) podem interpretar a mesma skill de formas distintas.

#### O que realmente funciona como controle
| Mecanismo | Eficácia | Por quê |
|-----------|----------|---------|
| Skills | Baixa | Sugestão, não lei |
| Prompts detalhados | Média | O agente pode ignorar partes |
| Checklists | Baixa | O agente pode não atualizar |
| **Testes automatizados** | **Alta** | Falha = erro objetivo, não negociável |
| **Build quebrado** | **Alta** | Não compila = não entra |
| **Revisão humana** | **Alta** | Último gate real |

#### Recomendação
Invista em **gates objetivos**, não em convenções subjetivas:
1. `go test ./...` deve passar — gate real
2. `go build ./...` deve passar — gate real
3. Architecture tests (`tests/architecture/`) — gate real
4. Skills e checklists são **diretrizes**, não **gates**

---

## 5. Paralelismo Realmente Funciona?

### Veredicto: **Sim, mas com ressalvas enormes.**

#### O que é paralelismo real aqui
Não é que 3 agentes editem o mesmo arquivo simultaneamente. É que:
- Agente 1 edita `internal/modules/agent/` na branch `agent-1`
- Agente 2 edita `internal/modules/review/` na branch `agent-2`
- Agente 3 edita `internal/modules/trigger/` na branch `agent-3`

Eles trabalham em **worktrees isoladas**, depois fazem merge.

#### O que funciona perfeitamente
- Módulos completamente independentes (nosso caso atual: agent, review, trigger)
- Diretórios disjuntos = zero conflitos de merge
- Cada agente tem seu próprio plano, checklist, e contexto

#### O que quebra
- **Dependências lógicas.** Se o Agente 2 precisa usar uma interface que o Agente 1 ainda não criou, não há paralelismo. O Agente 2 fica bloqueado ou tem que criar stubs.
- **Merge serial.** Os 3 agentes "trabalham em paralelo", mas o merge é serial. Se o Agente 1 quebrou tests, os merges 2 e 3 ficam pendentes.
- **Oração pelo domínio.** Se dois agentes adicionam types em `internal/domain/types.go`, há conflito. Resolver é fácil, mas requer intervenção humana — que é serial.
- **Refatorações.** Nenhuma refatoração cross-cutting pode ser paralelizada com este sistema.

#### Veredicto honesto
O paralelismo de **arquivo** é real e funciona bem. O paralelismo de **lógica** só funciona se o orquestrador definiu interfaces perfeitas antes. No nosso caso atual (Fase 5), os 3 módulos são independentes — então **sim, o paralelismo vai funcionar**. Mas não generalize isso para qualquer tarefa.

---

## 6. Composição: Orquestrador Inteligente + Agentes Menores

### Veredicto: **Eficaz para tarefas complexas. Overkill para tarefas simples.**

#### O que funciona
- **Separação de concerns.** Eu (orquestrador) faço o trabalho cognitivo pesado: entender o projeto, mapear dependências, definir fronteiras. O agente executor foca em codificar.
- **Especialização.** Um agente que só pensa em "como implementar AgentService" faz um trabalho melhor do que um agente que precisa pensar em "como implementar AgentService + Review + Trigger + como integrar tudo".
- **Redução de contexto.** Prompts menores = menos alucinação. Cada agente recebe apenas o contexto que precisa.

#### O que quebra
- **Overhead de comunicação.** O orquestrador gasta tokens/turnos para decompor. Para uma tarefa de 30 minutos, a decomposição leva 10 minutos — overhead de 33%.
- **Latência entre orquestrador e executor.** Eu entrego o plano, o usuário copia para outra ferramenta, o agente lê, executa, entrega. Esse handoff tem fricção.
- **O "agente menor" é o mesmo modelo.** Eu e o agente executor somos o mesmo modelo (Kimi/GPT/Claude). A divisão "orquestrador inteligente + agente burro" é artificial. O agente executor é tão capaz quanto eu. A única diferença é o contexto que recebe.
- **Erros do orquestrador se propagam.** Se eu errei na decomposição (ex: disse que dois módulos são independentes quando não são), o erro é sistêmico. Os 3 agentes vão trabalhar em paralelo sobre premissas falsas.

#### Quando usar esta composição
| Cenário | Usar? | Por quê |
|---------|-------|---------|
| Tarefa > 2h de trabalho | ✅ Sim | A decomposição vale o overhead |
| Módulos claramente independentes | ✅ Sim | Paralelismo real é possível |
| Tarefa < 30 min | ❌ Não | Overhead de orquestração > ganho |
| Refatoração cross-cutting | ❌ Não | Não é paralelizável |
| Bugfix urgente | ❌ Não | Mão única é mais rápida |
| Integração E2E | ⚠️ Depende | Só se os componentes já existem |

---

## 7. Falhas Arquiteturais do Sistema

### Falha 1: Não há feedback loop
O sistema é unidirecional: Orquestrador → Plano → Agente → Entrega. Não há mecanismo para o agente dizer "essa decomposição está errada" ou "preciso de mais contexto". O agente é tratado como executor passivo.

**Mitigação:** Adicionar uma fase de "validação do plano pelo agente" antes da execução. O agente lê o plano e pode solicitar esclarecimentos.

### Falha 2: Integração é manual
Quando os 3 agentes terminam, o usuário precisa fazer merge manualmente. Não há automação para detectar conflitos, rodar testes integrados, ou fazer rollback.

**Mitigação:** Script de integração (`scripts/merge-plans.sh`) que verifica conflitos, roda tests, e reporta status.

### Falha 3: Sem rollback
Se um agente entregar código quebrado, não há mecanismo de rollback além do git. O plano não contempla "o que fazer se o Agente 2 falhar".

**Mitigação:** Adicionar critério de aceite obrigatório: "se falhar, documente o estado e peça ajuda". E manter branches isoladas.

### Falha 4: Skills são voláteis
As skills estão em `~/.claude/skills/`, `~/.windsurf/skills/`, etc. Se o usuário muda de máquina ou reinstala uma ferramenta, as skills somem.

**Mitigação:** Versionar as skills no repositório do projeto (ex: `.skills/orchestrate.md` e `.skills/execute.md`) e symlink para os diretórios das ferramentas.

---

## 8. Score Final

| Aspecto | Nota (0-10) | Justificativa |
|---------|-------------|---------------|
| Escalabilidade | 6 | Funciona até 5 agentes. Depois vira burocracia. |
| Eficácia dos templates | 5 | Úteis como ponto de partida, perigosos como padrão rígido. |
| Utilidade dos checklists | 5 | Conceito valioso, arquivo frágil. O comportamento (Ralph Loop) é mais útil que o artefato. |
| Garantia de compliance | 3 | Impossível garantir. Skills são convenções, não leis. Gates objetivos (testes/build) são o controle real. |
| Viabilidade do paralelismo | 7 | Funciona para módulos independentes. Quebra para dependências lógicas e refatorações. |
| Composição Orquestrador+Executor | 7 | Eficaz para tarefas complexas. Overkill para tarefas simples. |
| **Média** | **5.5** | **Sistema viável, mas com falhas reais que precisam de mitigação.** |

---

## 9. Recomendações para o Futuro

1. **Foque em gates objetivos, não em convenções.** `go test ./...` é um gate real. Checklist é uma convenção.
2. **Mantenha a decomposição para tarefas > 2h.** Para tarefas pequenas, mão única é mais eficiente.
3. **Versione as skills no repositório.** Não dependa de diretórios locais.
4. **Automatize a integração.** Script que verifica conflitos e roda tests após os agentes entregarem.
5. **Aceite que checklists serão imperfeitos.** Use-os como diário de bordo, não como controle.
6. **Limite a 5 agentes por rodada.** Além disso, divida em waves.
