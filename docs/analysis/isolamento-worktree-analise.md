# Análise: Isolamento Automático de Worktree/Branch para Agentes

## Problema

Agentes executores podem estar editando a branch `main` diretamente. Sem isolamento:
- Conflitos imediatos entre agentes
- Impossibilidade de rollback individual
- Merge se torna caos

## Alternativas Analisadas

### A. Agente cria worktree/branch sozinho
**Como:** Skill execute instrui agente a rodar `git worktree add` e `git checkout -b`
**Prós:** Autônomo, não depende de orquestrador
**Contras:** 
- Agente pode errar (path errado, branch existente, permissões)
- Em IDEs (Windsurf), o agente não controla o workspace — o usuário precisa abrir o diretório correto
- Adiciona complexidade ao agente que deveria focar em código

### B. Orquestrador cria worktree/branch ANTES via script
**Como:** Script `bootstrap-agent.sh` cria tudo antes do agente começar
**Prós:** Determinístico, confiável, o agente só verifica
**Contras:** Requer ação do orquestrador/usuário antes de cada agente

### C. Planos incluem comando de setup específico
**Como:** Cada plano tem "Passo 0: Execute este comando para isolar"
**Prós:** Explícito, automático, adaptado à tarefa específica
**Contras:** Ainda depende do agente executar corretamente

### D. Verificação obrigatória + parada se não isolado (RECOMENDADO)
**Como:** Skill execute verifica isolamento. Se não estiver, PARA e instrui o usuário a isolar.
**Prós:** Seguro, simples, funciona em qualquer ferramenta
**Contras:** Requer intervenção humana para criar worktree em IDEs

## Veredito

**A melhor abordagem é híbrida:**

1. **Skill execute** → Verificação obrigatória + PARADA se não isolado
2. **Planos individuais** → Incluem comando de setup específico (para CLI)
3. **Script bootstrap** → Opcional, para automatizar criação
4. **Para IDEs** → O usuário é instruído a abrir o workspace correto ANTES de chamar o agente

## Limitação por Ferramenta

| Ferramenta | Pode criar worktree sozinho? | Estratégia |
|------------|------------------------------|------------|
| Kimi-CLI | ✅ Sim | Agente executa `git worktree add` |
| Windsurf | ❌ Não | Usuário abre workspace correto antes |
| Cursor | ❌ Não | Usuário abre workspace correto antes |
| Codex | ⚠️ Parcial | Depende do ambiente |

**Por que IDEs não podem:** O agente dentro do Windsurf/Cursor opera no workspace aberto no editor. Ele não pode "mudar de diretório" sem que o usuário mude o workspace da IDE.
