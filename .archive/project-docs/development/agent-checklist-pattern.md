# Agent Checklist Pattern (ACP)

**Status:** Proposta consolidada  
**Contexto:** Orquestração paralela de agentes em ferramentas distintas  
**Alternativa a:** Dependência de memória/histórico compartilhado entre agentes

---

## 1. Problema

Agentes executores operam em ferramentas separadas (Windsurf, Kimi Code Extension, Kimi-CLI, etc.) sem memória compartilhada. O Orquestrador não tem visibilidade do progresso de um agente sem:

- Acessar o histórico da ferramenta (inviável)
- Esperar o agente "terminar" e reportar (perde granularidade)
- Perguntar ao usuário (fricção humana desnecessária)

Isso cria uma **caixa preta** durante a execução paralela.

---

## 2. Análise de Alternativas

| Abordagem | Prós | Contras | Viabilidade |
|---|---|---|---|
| **SetTodoList nativo** | Integrado, rápido, sem I/O | Não persiste entre sessões; invisível para outros agentes; sem trilha de auditoria | ❌ Insuficiente para orquestração cross-tool |
| **CHECKLIST.md persistente** | Visível para todos; trilha de auditoria; sobrevive a crashes | I/O adicional; risco de desatualização; overhead se mal formatado | ✅ Viável COM regras rígidas |
| **Event Store (banco)** | Canônico, auditável, queryável | Overkill para progresso de agente; requer infraestrutura | ⚠️ Prematuro para MVP |
| **Chat/Thread como memória** | Zero infraestrutura | Volátil; depende de ferramenta específica; não versionável | ❌ Não atende requisito de persistência |

### Veredito

**A abordagem híbrida é a vencedora:**

- **SetTodoList** → uso interno do Orquestrador para planejamento e controle operacional
- **CHECKLIST.md** → persistência de progresso para agentes executores, com formato mínimo e atualizações explícitas

---

## 3. Estratégia Consolidada: Agent Checklist Pattern (ACP)

### 3.1 Princípios

1. **Mínimo viável:** O CHECKLIST.md contém APENAS checkboxes e IDs. Não é documentação, é controle.
2. **Auto-contido:** Cada tarefa tem seu próprio CHECKLIST. Não há checklist global.
3. **Atualização explícita:** O agente executor SEMPRE atualiza o checklist após completar um item significativo.
4. **Leitura obrigatória:** O agente executor RELÊ o checklist no início de cada iteração significativa.
5. **Imutabilidade de itens completos:** Itens marcados não são desmarcados. Erros são corrigidos, não revertidos.

### 3.2 Ralph Loop — Padrão de Execução

```
while (checkbox não marcada no CHECKLIST.md):
    1. LER CHECKLIST.md → identificar próximo item pendente
    2. EXECUTAR o item (código, teste, refactor)
    3. VALIDAR o item (testes passam? comportamento correto?)
    4. ATUALIZAR CHECKLIST.md → marcar checkbox como concluído
    5. COMMIT (opcional, mas recomendado)
```

O nome "Ralph Loop" é uma metáfora operacional: o agente "rala" (trabalha) em ciclos curtos, sempre verificando o que falta antes de continuar.

### 3.3 Formato do CHECKLIST.md

```markdown
# CHECKLIST — {TarefaID}: {Nome da Tarefa}

**Agente:** {identificador do agente}  
**Início:** {ISO timestamp}  
**Status:** {in_progress | completed | blocked}

---

## Checklist de Execução

- [x] 1. Ler documentação obrigatória (README, CONTRACTS, ADRs)
- [x] 2. Analisar código existente no escopo
- [ ] 3. Implementar {item específico}
- [ ] 4. Implementar {item específico}
- [ ] 5. Criar testes
- [ ] 6. Rodar testes existentes (regressão)
- [ ] 7. Code review auto-crítico
- [ ] 8. Correções pós-review
- [ ] 9. Validação final (`go test ./...`, `go build ./...`)
- [ ] 10. Entrega final ao usuário

## Notas de Progresso
<!-- O agente pode adicionar notas curtas aqui, uma por linha -->
- 2026-05-11T10:00:00Z — Lido código existente, identificado gap em X
- 2026-05-11T10:30:00Z — Implementado service.go, testes passando
```

**Regras de formato:**
- Máximo 15 itens no checklist principal
- Cada item é uma unidade de trabalho de 5-20 minutos
- Seção "Notas de Progresso" é opcional, mas recomendada
- Não use HTML, tabelas complexas, ou formatação pesada
- Apenas markdown puro

### 3.4 Numeração e Serialização

Cada plano de orquestração recebe um ID serializado:

```
{ORQUESTRADOR}-{FASE}-{RODADA}-{AGENTE}

Exemplos:
  ORCH-F05-R01-A01 → Orquestrador, Fase 5, Rodada 1, Agente 1
  ORCH-F05-R01-A02 → Orquestrador, Fase 5, Rodada 1, Agente 2
  ORCH-F06-R01-A01 → Orquestrador, Fase 6, Rodada 1, Agente 1
```

Arquivos gerados:
```
plans/
├── ORCH-F05-R01-A01-agentservice.md        # Plano + Prompt
├── ORCH-F05-R01-A01-agentservice-checklist.md  # Checklist de execução
├── ORCH-F05-R01-A02-review.md
├── ORCH-F05-R01-A02-review-checklist.md
├── ORCH-F05-R01-A03-triggers.md
└── ORCH-F05-R01-A03-triggers-checklist.md
```

### 3.5 Ciclo de Vida do Checklist

| Fase | Responsável | Ação |
|------|-------------|------|
| Criação | Orquestrador | Gera CHECKLIST-{id}.md com itens não marcados |
| Execução | Agente Executor | Lê, executa, atualiza, repete (Ralph Loop) |
| Inspeção | Orquestrador / Usuário | Lê CHECKLIST-{id}.md para verificar progresso sem acessar ferramenta do agente |
| Arquivamento | Orquestrador | Quando tarefa concluída, renomeia para `-completed.md` |

---

## 4. Mitigação de Riscos

| Risco | Mitigação |
|---|---|
| Agente esquece de atualizar checklist | Prompt instrucional OBRIGA atualização após cada item. Ralph Loop reforça isso. |
| Checklist fica desatualizado | Formato mínimo (apenas checkboxes) reduz fricção de atualização. |
| Overhead de I/O | Checklist é arquivo pequeno (< 2KB). Leitura/escrita é trivial. |
| Conflito de escrita | Cada agente escreve APENAS no seu próprio checklist. Nunca compartilhado. |
| Poluição do filesystem | Checklists ficam em `plans/`, fora do código. Podem ser `.gitignore` se desejado. |

---

## 5. Quando NÃO Usar

Não use CHECKLIST.md quando:
- A tarefa é trivial (< 10 minutos de trabalho)
- O agente executa na MESMA ferramenta que o orquestrador (aí SetTodoList basta)
- Não há paralelismo (apenas 1 agente trabalhando)
- O projeto não tem estrutura de diretórios estável

---

## 6. Integração com Skill Orquestrate

A skill `orchestrate` deve, na **Fase 3 (Criação dos Prompts)**, incluir automaticamente:

1. Geração do CHECKLIST-{id}.md junto com o prompt
2. Instrução no prompt do agente para usar o Ralph Loop
3. Referência ao caminho do checklist no prompt

Exemplo de instrução adicionada ao prompt do agente:

```markdown
## Ralph Loop — Execução Iterativa

Você deve executar esta tarefa em ciclos curtos usando o arquivo de checklist:

**Caminho do checklist:** `plans/ORCH-F05-R01-A01-agentservice-checklist.md`

**A cada iteração:**
1. Leia o checklist para identificar o próximo item pendente
2. Execute o item
3. Valide (testes passam? código correto?)
4. Atualize o checklist marcando o item como concluído
5. Continue para o próximo item

**Regra de ouro:** Nunca pule um item sem marcá-lo. Se encontrar bloqueio, adicione uma nota na seção "Notas de Progresso".
```

---

## 7. Exemplo Prático

Veja `plans/ORCH-F05-R01-A01-agentservice-checklist.md` para um checklist real aplicado ao plano de orquestração da Fase 5.

---

## Decisões

- **Formato:** Markdown puro, máximo 15 itens, sem tabelas complexas.
- **Localização:** `plans/CHECKLIST-{id}.md` (paralelo ao plano).
- **Persistência:** Versionado no git (trilha de auditoria) ou `.gitignore` (se considerado ruído).
- **Atualização:** Responsabilidade do agente executor, obrigatoriedade imposta pelo prompt.
- **Ralph Loop:** Padrão de execução, não ferramenta. O loop é comportamental.
