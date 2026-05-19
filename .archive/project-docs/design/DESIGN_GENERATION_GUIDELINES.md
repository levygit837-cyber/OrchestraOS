# OrchestraOS — Design Generation Guidelines

> **Purpose:** Prevenir convergência estética indesejada quando gerando múltiplos protótipos visuais com ferramentas de IA.
>
> **Status:** Adopted
>
> **Applies to:** AIDesigner, qualquer geração de protótipos visuais do OrchestraOS

---

## O Problema

Quando geramos múltiplos protótipos com prompts similares (mesmo com diretrizes estéticas diferentes), modelos de design tendem a convergir para:

1. **A mesma paleta dark** — fundo #050505-#0A0A0A, texto cinza claro
2. **As mesmas fontes** — Space Grotesk, JetBrains Mono, Inter
3. **O mesmo layout** — Canvas central + HUD flutuantes + inspector lateral
4. **Os mesmos elementos visuais** — glows âmbar, bordas sutis, glassmorphism
5. **A mesma sensação** — "dark tech terminal/dashboard"

Isso invalida o propósito de explorar **direções estéticas distintas**.

---

## Princípios para Geração Divergente

### 1. Varie a Paleta Fundamental, Não Apenas o Accent

❌ **Errado:**
> "Dark background (#070707)... with amber accent"  
> "Dark background (#0A0A0A)... with green accent"  
> "Dark background (#050505)... with blue accent"

✅ **Certo:**
> "Pure white (#FFFFFF) background with navy blue (#1E3A5F) primary and coral (#FF6B6B) accent"

> "Warm cream (#F5F0E8) background with forest green (#2D5A3D) primary and burnt orange (#CC5500) accent"

> "Slate gray (#334155) background with electric purple (#A855F7) primary and neon cyan (#06B6D4) accent"

**Regra:** Se 2 prompts compartilham a mesma cor de fundo dentro de 10% de similaridade, eles não são direções diferentes — são variações da mesma direção.

---

### 2. Varie a Tipografia como Elemento Dominante

❌ **Errado:** Space Grotesk + JetBrains Mono em todos os protótipos

✅ **Certo:**
- Protótipo A: Serif editorial (Playfair Display + Source Code Pro)
- Protótipo B: Geometric sans (Futura/Outfit + IBM Plex Mono)
- Protótipo C: Humanist sans (Frutiger/Satoshi + Fira Code)
- Protótipo D: Display/Brand (Clash Display + Space Mono)

**Regra:** A escolha tipográfica deve ser um dos primeiros 3 elementos mencionados no prompt.

---

### 3. Varie a Metáfora Visual

Cada protótipo deve evocar uma **metáfora espacial diferente**:

| Protótipo | Metáfora | Sensação |
|-----------|----------|----------|
| A | Centro de comando militar | Precisão, urgência, alerta |
| B | Revista editorial de luxo | Clareza, respiração, hierarquia |
| C | Laboratório médico | Higiene, dados, confiança |
| D | Studio de música eletrônica | Ritmo, energia, camadas |
| E | Jardim botânico digital | Crescimento, organicidade, vida |

**Regra:** O prompt deve começar com a metáfora, não com a funcionalidade.

---

### 4. Varie o Layout Fundamental

❌ **Errado:** Todos com "Canvas central + HUD overlays + inspector + input bar"

✅ **Certo:**
- **Canvas-first:** Grafo como elemento dominante (80% da tela)
- **Sidebar-first:** Navegação vertical dominante, canvas secundário
- **Card-first:** Dashboard de cards expansíveis, sem canvas
- **Timeline-first:** Linha do tempo vertical como elemento principal
- **Grid-first:** Grid rigoroso tipo spreadsheet, dados em células

**Regra:** Mencione explicitamente a distribuição de espaço em porcentagens.

---

### 5. Varie a Densidade de Informação

| Direção | Densidade | Exemplo |
|---------|-----------|---------|
| Airy | 30% de preenchimento | Apple Health, Notion |
| Balanced | 60% de preenchimento | Linear, Vercel |
| Dense | 90% de preenchimento | Bloomberg Terminal, Sentry |

**Regra:** Especifique a densidade explicitamente no prompt.

---

## Checklist Pré-Geração

Antes de gastar um crédito de design, verifique:

- [ ] **Paleta:** A cor de fundo é diferente dos protótipos já gerados (>20% de diferença de luminosidade)?
- [ ] **Tipografia:** As fontes são diferentes dos protótipos anteriores?
- [ ] **Metáfora:** A sensação/espacialidade é distintamente diferente?
- [ ] **Layout:** A organização espacial dos elementos é diferente?
- [ ] **Densidade:** A quantidade de informação por pixel é diferente?
- [ ] **Formas:** Os border-radius, cantos, e geometria são diferentes?
- [ ] **Efeitos:** O tratamento de sombras, glows, blur é diferente?

**Se 4+ itens não forem diferentes, reescreva o prompt.**

---

## Prompt Template Divergente

Use esta estrutura para garantir divergência:

```
[METÁFORA] A [metáfora espacial] interface for an AI agent orchestration system.

[PALETA] [Cor de fundo] background with [cor primária] primary and [cor de destaque] accent. 
[Restrições de paleta específicas - ex: "nunca use âmbar ou preto puro"]

[TIPOGRAFIA] [Fonte principal] for headings and UI, [Fonte mono] for data. 
[Tratamento tipográfico específico - ex: "all caps labels", "extreme weight contrast"]

[LAYOUT] [X]% of screen for [elemento principal], [Y]% for [elemento secundário]. 
[Organização espacial específica - ex: "left sidebar 240px", "bottom drawer", "floating center modal"]

[DENSIDADE] [Airy/Balanced/Dense] information display.
[Quanto espaço em branco ou quantos dados por polegada]

[FORMAS] [Border-radius específico], [estilo de cantos], [tratamento de bordas].
[Ex: "16px rounded everywhere", "sharp 0px corners", "cut chamfer corners"]

[EFEITOS] [Shadows/glows/blur específicos ou ausência deles].
[Ex: "no glows, flat shadows only", "heavy glassmorphism", "neon outlines"]

[ELEMENTOS VISUAIS DISTINTIVOS] 2-3 elementos que tornam este design único.
[Ex: "horizontal timeline instead of graph", "circular radial layout", "card stack metaphor"]

[CONTEXTO DO PRODUTO] OrchestraOS is an AI agent orchestration system...
[Fluxo de elementos: UserMessage → Orchestrator → Task → Work Units → Runs → Sessions]
```

---

## Anti-Patterns a Evitar

1. **"Dark background with [color] accent"** — Isso gera 95% dos designs dark tech.
2. **Space Grotesk como default** — É a fonte mais comum em designs de IA.
3. **Glassmorphism como solução universal** — Funciona para tudo, portanto não diferencia nada.
4. **Canvas + HUD + Inspector + Input Bar** — Essa estrutura força convergência de layout.
5. **Glows âmbar para estados ativos** — É o padrão mais óbvio; experimente pulse, bounce, fill, ou outline.
6. **Prompts que começam com funcionalidade** — Comece com sensação/estética, termine com funcionalidade.

---

## Referências de Direções Estéticas Distintas

Para uso em prompts futuros, aqui estão direções que garantem divergência:

| Nome | Fundo | Primária | Accent | Fonte | Densidade | Layout |
|------|-------|----------|--------|-------|-----------|--------|
| Obsidian Glass | #070707 | #1C1C1C | #F59E0B | Space Grotesk | Balanced | Canvas + HUD |
| Cyber-Op | #050505 | #0a0a0c | #FBBF24 | Space Grotesk | Dense | Canvas + HUD |
| Swiss Precision | #0A0A0A | #141414 | #F59E0B | Space Grotesk | Dense | Canvas + Sidebar |
| Editorial Light | #FFFFFF | #F8F7F4 | #1E3A5F | Playfair + JetBrains | Airy | Sidebar + Cards |
| Medical Clean | #F0F4F8 | #FFFFFF | #0EA5E9 | Inter + Roboto Mono | Balanced | Grid + Timeline |
| Nature Organic | #F5F0E8 | #E8E0D0 | #2D5A3D | Satoshi + Source Code | Airy | Card Stack |
| Neon Nightclub | #0F0F23 | #1A1A2E | #A855F7 | Clash Display + Fira | Dense | Radial + HUD |
| Brutalist Raw | #E5E5E5 | #FFFFFF | #FF0000 | Helvetica + Courier | Dense | Grid + Overlap |

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-05-17 | Criado este guideline | Após 3 protótipos gerados convergirem para estética dark tech similar |
