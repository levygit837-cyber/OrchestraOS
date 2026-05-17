# PLANO — {NOME_DA_TAREFA}

**ID:** {PLAN_ID}  
**Tipo:** faseado  
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

### 2.1 Módulos Afetados
- `path/to/module` — {resumo do estado atual}

### 2.2 Contratos e Interfaces Existentes
- ...

### 2.3 Testes Existentes Relacionados
- ...

---

## 3. Análise de Alternativas

### 3.1 Alternativa A: {Nome}
**Descrição:** ...
**Arquivos tocados:** ...
**Riscos:** ...
**Tempo estimado:** Xh
**Prós:** ...
**Contras:** ...

### 3.2 Alternativa B: {Nome}
**Descrição:** ...
**Arquivos tocados:** ...
**Riscos:** ...
**Tempo estimado:** Yh
**Prós:** ...
**Contras:** ...

### 3.3 Alternativa C: {Nome}
**Descrição:** ...
**Arquivos tocados:** ...
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

## 5. Plano de Implementação

### 5.1 Fases

#### Fase 1: {Nome}
**Objetivo:** ...
**Entregável:** ...
**Dependências:** ...

#### Fase 2: {Nome}
**Objetivo:** ...
**Entregável:** ...
**Dependências:** ...

#### Fase 3: {Nome}
**Objetivo:** ...
**Entregável:** ...
**Dependências:** ...

### 5.2 Fronteiras

| Tipo | Caminhos |
|---|---|
| **TOCAR** | `path/...` |
| **EVITAR** | `path/...` |
| **DEPENDE_DE** | `interface/contract` |

### 5.3 Dependências Externas
- ...

### 5.4 Ordem de Integração
{Se houver dependências entre fases, qual a ordem sugerida?}

---

## 6. Estratégia de Testes

### 6.1 Abordagem
- **Testes reais vs stubs/mocks:** {quando usar cada um}
- **Testes de integração:** {como serão feitos}

### 6.2 Cobertura Mínima
- [ ] Happy path
- [ ] Casos de erro principais
- [ ] Casos de borda
- [ ] {Específico do domínio}

### 6.3 Cenários de Teste
1. {Cenário 1}
2. {Cenário 2}
3. {Cenário 3}

### 6.4 Validação
"Esses testes exercitam o problema real?" → {Justificativa}

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

## 8. Riscos e Mitigações

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

## 10. Critérios de Aceite

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
