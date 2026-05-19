# Análise Crítica dos Prompts — Auto-avaliação

> Data: 2026-05-17
> Prompts avaliados: Command Center (68 linhas) + Task Canvas (128 linhas)

---

## Command Center — 68 linhas

### ✅ Pontos Fortes
- Contexto anti-chat/anti-IDE/anti-kanban é claro
- As 4 seções principais estão bem definidas
- Paleta e tipografia específicas
- Estados vazios pensados

### ❌ Pontos Fracos Críticos

| # | Problema | Risco no AIDesigner | Gravidade |
|---|---|---|---|
| 1 | **Sem dados de exemplo reais** | Vai inventar "Task 1", "Agent A", layout genérico de SaaS | 🔴 Alta |
| 2 | **Hierarquia visual ambígua** | Fila de Ações deveria DOMINAR (40% altura), mas não especifiquei %. Pode virar faixa fina irrelevante | 🔴 Alta |
| 3 | **"Sidebar direita OU painel inferior"** | Ambiguidade pode gerar layout esquisito meio-termo | 🟡 Média |
| 4 | **Sem especificação de interações** | Clicar [Aprovar] deveria ser inline, não navegar para outra tela. Disse "tudo leva ao Task Canvas" — incorreto | 🔴 Alta |
| 5 | **Sem referência de proporção** | Linear tem sidebar 260px, Vercel tem cards grandes. Não disse qual modelo de densidade | 🟡 Média |
| 6 | **Progresso "X de Y" é vago** | Pode virar texto simples ao invés de barra visual rica com mini-DAG | 🟡 Média |

### Veredito
**INSUFICIENTE.** O AIDesigner vai entregar um dashboard genérico de SaaS com os elementos solicitados, mas sem a densidade e hierarquia corretas de um sistema de orquestração.

---

## Task Canvas — 128 linhas

### ✅ Pontos Fortes
- DAG bem especificado com tipos de arestas
- 5 abas no painel de detalhes
- Níveis de zoom definidos
- Interações descritas

### ❌ Pontos Fracos Críticos

| # | Problema | Risco no AIDesigner | Gravidade |
|---|---|---|---|
| 1 | **Muito complexo para uma geração** | 128 linhas = pode simplificar demais ou ignorar metade (esquecer artefatos, colapsar em diagrama simples) | 🔴 Alta |
| 2 | **Canvas "infinito" em HTML estático** | AIDesigner gera HTML, não engine de canvas. Pode entregar diagrama estático feio ou div scrollável básico | 🔴 Alta |
| 3 | **Sem exemplo de DAG concreto** | "wu-001, wu-002" é vazio. Precisa de dados reais: "Schema SQL → Repository → Service → E2E" | 🔴 Alta |
| 4 | **Posicionamento dos nós** | Não especifiquei layout topológico (esquerda→direita? topo→baixo?). Pode fazer diagrama circular caótico | 🔴 Alta |
| 5 | **Painel vazio** | Não disse o que acontece quando nenhuma WU está selecionada | 🟡 Média |
| 6 | **Mini-map pode ser complexo demais** | Para HTML estático, pode sair bugado ou ser omitido | 🟢 Baixa |
| 7 | **5 abas é muito para primeira geração** | AIDesigner pode fazer abas genéricas sem conteúdo real em cada uma | 🟡 Média |

### Veredito
**DEMAIS + INCOMPLETO.** 128 linhas é excessivo para o AIDesigner absorver de uma vez. Ele vai pegar a "vibe" mas perder a precisão. O resultado pode ser um diagrama bonito que não parece um sistema operacional vivo.

---

## 🎯 Problema Raiz

Os prompts descrevem **O QUÊ** mas não **COMO** em termos de:
1. **Dados reais** — o AIDesigner precisa de conteúdo concreto para criar layouts densos
2. **Hierarquia espacial** — % de tela, tamanhos relativos, prioridade visual
3. **Interações específicas** — o que acontece em cada clique, hover, scroll
4. **Modelo mental** — como o usuário navega, qual o fluxo de atenção

---

## ✅ Solução Proposta

### Antes de gerar no AIDesigner:
1. **Criar dados de exemplo concretos** — uma task realista do OrchestraOS
2. **Criar wireframe/protótipo HTML simples** — validar layout espacialmente
3. **Melhorar prompts** — usar protótipo como referência + dados reais

### Na geração no AIDesigner:
1. **Command Center**: Incluir screenshot do protótipo como referência visual
2. **Task Canvas**: Simplificar prompt para focar apenas em DAG + painel de detalhes básico. Abas específicas em refinamentos futuros.

---

## 📊 Resumo

| Aspecto | Command Center | Task Canvas |
|---|---|---|
| Tamanho do prompt | Insuficiente (68 linhas) | Excessivo (128 linhas) |
| Dados de exemplo | ❌ Ausentes | ❌ Ausentes |
| Hierarquia espacial | ❌ Ambígua | ❌ Ausente |
| Interações | ❌ Genéricas | ⚠️ Parciais |
| Risco de resultado genérico | 🔴 Alto | 🔴 Alto |
| Precisa de protótipo? | ✅ Sim | ✅ Sim |
