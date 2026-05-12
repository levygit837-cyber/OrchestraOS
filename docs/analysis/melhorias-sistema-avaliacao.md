# Avaliação de Melhorias Propostas ao Sistema de Orquestração

**Data:** 2026-05-11  
**Regra:** Identificar o que é viável agora, o que é armadilha, e o que é futuro.

---

## 1. Feedback Loop — Validação de Plano pelo Agente

### Proposta
O agente, ao receber um plano, analisa-o e pode:
- Solicitar mais contexto
- Fazer perguntas ao usuário
- Questionar a decomposição

### Avaliação: ✅ **VIÁVEL — Implementar imediatamente**

**Por que funciona:**
- Custa zero infraestrutura. É apenas uma fase adicional na skill `execute`.
- Evita o problema do "plano errado sendo executado cegamente".
- Um agente que encontra ambiguidade no plano pode parar antes de gastar tokens/turnos codando algo errado.

**Implementação:**
Adicionar na skill `execute`, Fase 1.4:

```markdown
### 1.4 Validação do Plano

Após ler o plano e a documentação obrigatória, avalie:

- [ ] Entendi EXATAMENTE o que devo implementar?
- [ ] As fronteiras (TOCAR/EVITAR) são claras?
- [ ] Tenho acesso a todas as informações necessárias?
- [ ] Não há ambiguidades que precisem de esclarecimento?

Se algum item acima for "não", você DEVE:
1. Listar suas dúvidas de forma estruturada
2. Solicitar esclarecimento ao usuário/orquestrador
3. NÃO comece a implementar até as dúvidas serem resolvidas
```

**Custo:** Negligenciável. **Valor:** Alto.

---

## 2. Verificação de Merge

### Proposta
Agentes terem informações de branch, saber quando mergear, revisar PRs uns dos outros.

### Avaliação: ⚠️ **ARMADILHA — Não delegue merge a agentes**

**O problema:**

| Aspecto | Realidade |
|---------|-----------|
| Agentes em tools diferentes | Não compartilham filesystem nem git state |
| Worktrees isoladas | Cada agente vê apenas sua própria branch |
| Merge requer resolução de conflitos | Isso exige entendimento de TODOS os contextos |
| PR review | Um agente não pode revisar código de outro sem ler o plano completo dele |

**A ilusão:** A ideia de "último agente mergeia tudo" parece automação, mas é delegar a operação mais arriscada (merge) para a entidade com menos contexto (um agente que só viu seu próprio plano).

**O que acontece na prática:**
- Agente 3 tenta mergear branch do Agente 1
- Conflito em `domain/types.go`
- Agente 3 não sabe qual type do Agente 1 é mais importante
- Agente 3 chuta uma resolução
- Build quebra
- O usuário precisa desfazer tudo

**Solução pragmática:**

Merge é responsabilidade do **Orquestrador** (humano ou IA central), nunca do agente executor.

```
FLUXO CORRETO:
1. Agente 1 entrega → push branch A01
2. Agente 2 entrega → push branch A02
3. Agente 3 entrega → push branch A03
4. ORQUESTRADOR faz merge sequencial:
   - Checkout main
   - Merge A01 → testa → OK
   - Merge A02 → testa → OK
   - Merge A03 → testa → OK
5. Se conflito: Orquestrador decide (tem contexto de todos os planos)
```

**Regra de ouro:** Agentes executores **empurram branches**. Quem **puxa para main** é o orquestrador.

---

## 3. Rollback

### Proposta
"Se falhar, documente o estado e peça ajuda." Branches isoladas, parar output e aguardar.

### Avaliação: ✅ **VIÁVEL — Implementar imediatamente**

**Por que funciona:**
- Branches isoladas já são rollback nativo (git)
- "Parar e pedir ajuda" é comportamental, não infraestrutural
- Custa zero implementação

**Implementação na skill `execute`:**

```markdown
### Rollback e Falha

Se em qualquer momento você:
- Não conseguir fazer os testes passarem após 3 tentativas
- Encontrar um bloqueio técnico que não consegue resolver
- Perceber que o plano está incorreto (requisito impossível, dependência inexistente)
- Quebrar tests existentes e não entender por que

Você DEVE:
1. PARAR imediatamente (não continue codando no escuro)
2. Documentar o estado atual no checklist (Notas de Progresso)
3. Reportar ao usuário:
   - O que você tentou
   - Onde parou
   - Qual erro encontrou
   - O que precisa para continuar
4. NÃO faça merge. NÃO entregue código quebrado.
```

**Custo:** Negligenciável. **Valor:** Alto (evita código quebrado na main).

---

## 4. Comunicação Real-Time (GMB — Git Message Bus)

### Proposta
Branch dedicada `orchestrator-comms` com inbox, threads, signals. WebSocket como notificador.

### Avaliação: 🟡 **VIÁVEL COM CORTES SEVEREOS — Simplifique ou morra**

#### A ideia brilhante
Usar git como message bus é **elegante**:
- Persistente (não perde mensagens)
- Auditável (`git log` mostra toda a conversa)
- Versionado (conflitos são detectáveis)
- Sem infraestrutura externa (usa o que já existe)

#### A armadilha: WebSocket
A proposta de WebSocket server local (`curl localhost:8765/notify`) é **overkill**:

| Problema | Por que é ruim |
|----------|----------------|
| Precisa implementar e manter um WS server | Código novo para manter |
| Agente precisa saber usar `curl` corretamente | Mais instrução, mais erro |
| Latência falsa | WS diz "olhe o GMB", mas o agente só olha no próximo Ralph Loop (minutos depois) |
| Complexidade de deploy | O WS server precisa estar rodando antes dos agentes |

**O WebSocket está resolvendo um problema que não existe.** Os agentes não precisam de notificação em tempo real — eles operam em ciclos de minutos, não milissegundos.

#### A versão que funciona: GMB Puro (sem WebSocket)

```
orchestrator-comms/  (branch dedicada)
├── inbox/
│   ├── A01/           # Mensagens para Agente 1
│   ├── A02/
│   └── A03/
├── threads/           # Conversas por tópico
│   └── thread-042-userprofile.md
└── signals/           # Sinalizações simples
    └── A02-blocked-A01.json
```

**Como funciona (polling via git):**

1. No início de cada Ralph Loop, o agente faz:
   ```bash
   git fetch origin orchestrator-comms
   git checkout origin/orchestrator-comms -- inbox/A01/
   ```

2. Lê mensagens no inbox, processa, responde.

3. No final do loop, faz commit e push:
   ```bash
   git add .
   git commit -m "comms: A01→A02 request UserProfile"
   git push origin orchestrator-comms
   ```

4. Outro agente vê no próximo loop.

**Por que polling é suficiente:**
- Ralph Loop já é cíclico (a cada 5-20 minutos)
- Não há urgência real-time entre agentes
- Git já resolve concorrência (se dois agentes commitam, um ganha, outro faz rebase)
- Zero código novo para manter

#### Recomendação para GMB

Implementar **sem WebSocket**. Apenas a branch `orchestrator-comms` com convenções de diretório. Polling via `git fetch` no início de cada Ralph Loop.

Se no futuro a latência de minutos for problema, aí sim se discute WebSocket. Não antes.

---

## 5. Resumo das Recomendações

| Melhoria | Status | Ação | Custo |
|----------|--------|------|-------|
| Feedback Loop | ✅ Viable | Adicionar à skill `execute` | Zero |
| Rollback | ✅ Viable | Adicionar à skill `execute` | Zero |
| Verificação de Merge | ⚠️ Não delegar | Merge é do Orquestrador, não do agente | N/A |
| GMB (Git Message Bus) | 🟡 Viable com cortes | Implementar sem WebSocket, apenas branch + polling | Baixo |
| WebSocket notifier | ❌ Overkill | Deixar para futuro se latência for problema | Alto |

---

## 6. O Que Implementar Agora

### Já (zero custo):
1. Atualizar skill `execute` com Feedback Loop
2. Atualizar skill `execute` com Rollback
3. Documentar que merge é do Orquestrador (na skill `orchestrate`)

### Próximo sprint (baixo custo):
4. Criar branch `orchestrator-comms`
5. Documentar convenções de inbox/thread/signal
6. Adicionar na skill `execute`: "no início de cada loop, faça `git fetch origin orchestrator-comms`"

### Nunca (sem necessidade real):
7. WebSocket server local
