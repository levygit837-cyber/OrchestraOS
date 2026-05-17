# PLANO — {NOME_DA_TAREFA}

**ID:** {PLAN_ID}  
**Tipo:** por-dominio  
**Agente:** {AGENT_ID}  
**Criado em:** {ISO_TIMESTAMP}  
**Atualizado em:** {ISO_TIMESTAMP}  
**Status:** rascunho | ativo | concluído

---

## 1. Contexto e Problema

### 1.1 O que estamos resolvendo
{Descrição clara do problema. Por que isso precisa ser feito?}

### 1.2 Por que isso importa
{Impacto esperado: usuário final, performance, manutenibilidade, negócio.}

### 1.3 Restrições
- **Técnicas:** ...
- **Negócio:** ...
- **Tempo:** ...

---

## 2. Snapshot do Estado Atual

{Como o código está ANTES das mudanças.}

### 2.1 Domínios Existentes
- `path/to/domain-a` — {resumo}
- `path/to/domain-b` — {resumo}

### 2.2 Contratos e Interfaces entre Domínios
- ...

### 2.3 Testes Existentes Relacionados
- ...

---

## 3. Análise de Alternativas

### 3.1 Alternativa A: {Nome}
**Descrição:** ...
**Domínios tocados:** ...
**Riscos:** ...
**Tempo estimado:** Xh
**Prós:** ...
**Contras:** ...

### 3.2 Alternativa B: {Nome}
**Descrição:** ...
**Domínios tocados:** ...
**Riscos:** ...
**Tempo estimado:** Yh
**Prós:** ...
**Contras:** ...

### 3.3 Alternativa C: {Nome}
**Descrição:** ...
**Domínios tocados:** ...
**Riscos:** ...
**Tempo estimado:** Zh
**Prós:** ...
**Contras:** ...

---

## 4. Decisão

### 4.1 Alternativa Escolhida: {A / B / C}

### 4.2 Justificativa
{Por que esta foi a melhor opção?}

### 4.3 Prós
- ...

### 4.4 Contras e Mitigações
- ...

### 4.5 Alternativas Descartadas (e por quê)
- ...

---

## 5. Domínios Afetados

### 5.1 Domínio 1: {Nome}

**TOUCH:** `path/to/module/`
**AVOID:** `path/to/other/`
**DEPENDE_DE:** `interface/contract`

#### O que fazer
- {Item 1}
- {Item 2}
- {Item 3}

#### Critérios de Aceite
- [ ] Critério 1
- [ ] Critério 2
- [ ] Critério 3

#### Estratégia de Testes
- **Abordagem:** ...
- **Cobertura:** ...
- **Cenários:** ...

#### Riscos Específicos
| Risco | Probabilidade | Impacto | Mitigação |
|---|---|---|---|
| ... | Alta/Média/Baixa | Alto/Médio/Baixo | ... |

---

### 5.2 Domínio 2: {Nome}

**TOUCH:** `path/to/module/`
**AVOID:** `path/to/other/`
**DEPENDE_DE:** `interface/contract`

#### O que fazer
- {Item 1}
- {Item 2}
- {Item 3}

#### Critérios de Aceite
- [ ] Critério 1
- [ ] Critério 2
- [ ] Critério 3

#### Estratégia de Testes
- **Abordagem:** ...
- **Cobertura:** ...
- **Cenários:** ...

#### Riscos Específicos
| Risco | Probabilidade | Impacto | Mitigação |
|---|---|---|---|
| ... | Alta/Média/Baixa | Alto/Médio/Baixo | ... |

---

### 5.3 Domínio 3: {Nome}

**TOUCH:** `path/to/module/`
**AVOID:** `path/to/other/`
**DEPENDE_DE:** `interface/contract`

#### O que fazer
- {Item 1}
- {Item 2}
- {Item 3}

#### Critérios de Aceite
- [ ] Critério 1
- [ ] Critério 2
- [ ] Critério 3

#### Estratégia de Testes
- **Abordagem:** ...
- **Cobertura:** ...
- **Cenários:** ...

#### Riscos Específicos
| Risco | Probabilidade | Impacto | Mitigação |
|---|---|---|---|
| ... | Alta/Média/Baixa | Alto/Médio/Baixo | ... |

---

## 6. Integração entre Domínios

### 6.1 Contratos Compartilhados
{Quais interfaces, eventos ou schemas são compartilhados?}

### 6.2 Dependências Cruzadas
| Domínio | Depende de | Bloqueia |
|---|---|---|
| A | B | C |
| B | — | A |

### 6.3 Ordem de Implementação
{Se houver dependências, qual a ordem sugerida?}

---

## 7. Estratégia de Debug

{Omitir seção inteira se não aplicável.}

### 7.1 Ambiente
- **Escopo:** frontend / backend / full-stack
- **Ferramentas disponíveis:** ...

### 7.2 Dev Server Isolado
- Como iniciar:
- Comando:

### 7.3 O que Será Testado no Debug
- ...

### 7.4 Métricas e Critérios de Aceite Observáveis
- ...

---

## 8. Riscos Globais e Mitigações

| Risco | Probabilidade | Impacto | Mitigação | Responsável |
|---|---|---|---|---|
| ... | Alta/Média/Baixa | Alto/Médio/Baixo | ... | agente/usuário |

---

## 9. Workspace Setup Notes

### 9.1 Pré-requisitos Técnicos
- Branch/worktree:
- Variáveis de ambiente:
- Dependências a instalar:

### 9.2 Comandos de Setup
```bash
# exemplo
```

---

## 10. Critérios de Aceite Globais

- [ ] Critério 1 (verificável, testável)
- [ ] Critério 2 (verificável, testável)
- [ ] Critério 3 (verificável, testável)

---

## 11. Estimativa

- **Tempo total estimado:** Xh
- **Complexidade:** Baixa / Média / Alta
- **Incertezas:** ...

---

## 12. Próximos Passos

1. ...
2. ...
3. ...

---

## 13. Notas e Decisões Adicionais

{Espaço livre para seções extras: performance, segurança, migrations, etc.}
