# PLANO — {NOME_DA_TAREFA}

**ID:** {PLAN_ID}  
**Tipo:** cenario-based  
**Agente:** {AGENT_ID}  
**Criado em:** {ISO_TIMESTAMP}  
**Atualizado em:** {ISO_TIMESTAMP}  
**Status:** rascunho | ativo | concluído

---

## 1. Contexto e Problema

### 1.1 Qual feature ou fluxo estamos implementando
{Descrição clara da funcionalidade do ponto de vista do usuário.}

### 1.2 Por que isso importa
{Impacto esperado: experiência do usuário, conversão, retenção, etc.}

### 1.3 Restrições
- **Técnicas:** ...
- **Negócio:** ...
- **Tempo:** ...

---

## 2. Snapshot do Estado Atual

{Como o sistema se comporta ANTES das mudanças.}

### 2.1 Fluxos Existentes Relacionados
- {Fluxo 1}
- {Fluxo 2}

### 2.2 Contratos e Interfaces Existentes
- ...

### 2.3 Testes Existentes Relacionados
- ...

---

## 3. Análise de Alternativas

### 3.1 Alternativa A: {Nome}
**Descrição:** ...
**Fluxos afetados:** ...
**Riscos:** ...
**Tempo estimado:** Xh
**Prós:** ...
**Contras:** ...

### 3.2 Alternativa B: {Nome}
**Descrição:** ...
**Fluxos afetados:** ...
**Riscos:** ...
**Tempo estimado:** Yh
**Prós:** ...
**Contras:** ...

### 3.3 Alternativa C: {Nome}
**Descrição:** ...
**Fluxos afetados:** ...
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

## 5. Cenários

### 5.1 Cenário 1: {Título — Caminho Feliz}

**DADO** {pré-condição}
**QUANDO** {ação do usuário ou evento}
**ENTÃO** {resultado esperado}
**E** {resultado adicional}

#### Implementação Técnica
- {O que precisa ser construído para este cenário funcionar}
- {Quais endpoints, funções ou componentes são necessários}

#### Critérios de Aceite
- [ ] Critério verificável 1
- [ ] Critério verificável 2

#### Dados de Teste
- {Entrada exemplo}
- {Saída esperada}

---

### 5.2 Cenário 2: {Título — Variação}

**DADO** {pré-condição diferente}
**QUANDO** {ação diferente}
**ENTÃO** {resultado diferente}
**E** {resultado adicional}

#### Implementação Técnica
- ...

#### Critérios de Aceite
- [ ] ...

#### Dados de Teste
- ...

---

### 5.3 Cenário 3: {Título — Caso de Erro/Borda}

**DADO** {pré-condição}
**QUANDO** {ação que causa erro ou condição de borda}
**ENTÃO** {comportamento de erro esperado}
**E** {feedback para o usuário}

#### Implementação Técnica
- {Como o erro será tratado}
- {Qual status code, mensagem ou estado será retornado}

#### Critérios de Aceite
- [ ] ...

#### Dados de Teste
- ...

---

### 5.4 Cenário 4: {Título — Outro Caso Relevante}
...

---

## 6. Mapeamento de Cenários para Implementação

| Cenário | Endpoint/Função | Módulo | Status |
|---|---|---|---|
| 1 | `POST /api/...` | `internal/modules/...` | pendente |
| 2 | `GET /api/...` | `internal/modules/...` | pendente |
| 3 | ... | ... | ... |

---

## 7. Fronteiras Técnicas

| Tipo | Caminhos |
|---|---|
| **TOUCH** | `path/...` |
| **EVITAR** | `path/...` |
| **DEPENDE_DE** | `interface/contract` |

---

## 8. Estratégia de Testes

### 8.1 Abordagem
- **Testes de unidade:** {para cada cenário}
- **Testes de integração:** {fluxo end-to-end}
- **Testes de UI/E2E:** {se aplicável}

### 8.2 Cobertura Mínima
- [ ] Cenário feliz
- [ ] Cada variação de entrada
- [ ] Cada caso de erro documentado
- [ ] {Específico do domínio}

### 8.3 Cenários de Teste Automatizado
1. {Cenário 1}
2. {Cenário 2}
3. {Cenário 3}

### 8.4 Validação
"Esses testes cobrem todos os cenários do usuário?" → {Justificativa}

---

## 9. Estratégia de Debug

{Omitir seção inteira se não aplicável.}

### 9.1 Ambiente
- **Escopo:** frontend / backend / full-stack
- **Ferramentas disponíveis:** ...

### 9.2 Dev Server Isolado
- Como iniciar:
- Comando:

### 9.3 O que Será Testado no Debug
- {Simular cada cenário no ambiente de desenvolvimento}

### 9.4 Métricas e Critérios de Aceite Observáveis
- {Tempo de resposta, estados da UI, mensagens de erro}

---

## 10. Riscos e Mitigações

| Risco | Probabilidade | Impacto | Mitigação | Responsável |
|---|---|---|---|---|
| ... | Alta/Média/Baixa | Alto/Médio/Baixo | ... | agente/usuário |

---

## 11. Workspace Setup Notes

### 11.1 Pré-requisitos Técnicos
- Branch/worktree:
- Variáveis de ambiente:
- Dependências a instalar:

### 11.2 Comandos de Setup
```bash
# exemplo
```

---

## 12. Critérios de Aceite Globais

- [ ] Todos os cenários documentados funcionam conforme especificado
- [ ] Casos de erro retornam feedback adequado ao usuário
- [ ] Performance é aceitável (definir métrica)

---

## 13. Estimativa

- **Tempo total estimado:** Xh
- **Complexidade:** Baixa / Média / Alta
- **Incertezas:** ...

---

## 14. Próximos Passos

1. ...
2. ...
3. ...

---

## 15. Notas e Decisões Adicionais

{Espaço livre para seções extras: performance, segurança, analytics, etc.}
