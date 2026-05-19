# Tipos de Plano — Quando Usar Cada Um

Este documento define os tipos de plano suportados no OrchestraOS. O tipo de um plano é metadado — ele orienta o agente executor sobre como interpretar a estrutura, mas não limita o conteúdo.

## Tipos Disponíveis

| Tipo | Melhor Quando | Evitar Quando | Analogia |
|---|---|---|---|
| **Faseado** | Fluxo sequencial com dependências temporais claras | Sistema altamente modular com fronteiras independentes | Pipeline CI/CD |
| **Por Domínio** | Módulos independentes que podem evoluir separadamente | Fluxo único onde passo B só funciona depois de A | Microserviços |
| **Árvore de Decisões** | Problema técnico aberto com múltiplos caminhos possíveis | A decisão já está tomada; só falta executar | RFC técnico |
| **Cenário-Based** | Feature user-facing com múltiplos fluxos de usuário | Refatoração puramente técnica sem mudança de comportamento | User Story Mapping |

---

## Faseado (Padrão)

### Quando Performa Bem
- Nova feature que exige: migration → model → repository → service → handler
- Refatoração onde a ordem importa (ex: renomear campo no banco antes de atualizar código)
- Onboarding de novo módulo com passos obrigatórios

### Quando Performa Mal
- Múltiplos módulos independentes são afetados — o plano fica confuso misturando domínios
- A decisão técnica ainda não foi tomada — o plano fica cheio de "se X, então Y"

### Estrutura
```
Fase 1: Setup
Fase 2: Core Implementation
Fase 3: Integração
Fase 4: Testes e Validação
```

---

## Por Domínio

### Quando Performa Bem
- Sistema modular onde `billing/`, `auth/`, `notifications/` podem evoluir separadamente
- Cada módulo tem seu próprio contrato e não bloqueia os outros
- Equipe trabalha em paralelo por domínio

### Quando Performa Mal
- Há dependência circular entre domínios (ex: `billing` depende de `auth`, que depende de `billing`)
- O trabalho é essencialmente um fluxo sequencial que atravessa todos os domínios

### Estrutura
```
## Domínio: Auth
- TOUCH: internal/modules/auth/
- ENTREGA: ...

## Domínio: Billing
- TOUCH: internal/modules/billing/
- ENTREGA: ...

## Domínio: Notifications
- TOUCH: internal/modules/notifications/
- ENTREGA: ...

## Integração
- Como os domínios se comunicam
```

---

## Árvore de Decisões

### Quando Performa Bem
- Escolha arquitetural não trivial (ex: "usar event sourcing ou state machine?")
- Tecnologia nova sendo avaliada (ex: "GraphQL vs REST?")
- Trade-off de performance vs manutenibilidade

### Quando Performa Mal
- A decisão já foi tomada em ADR anterior — reabrir é perda de tempo
- O problema é puramente implementação, não decisão

### Estrutura
```
## Decisão 1: Arquitetura de Comunicação

### Opção A: Síncrono (RPC)
- Prós: ...
- Contras: ...

### Opção B: Assíncrono (Eventos)
- Prós: ...
- Contras: ...

### Escolha: B
→ Impacta Decisão 2: Qual message broker?

## Decisão 2: Message Broker

### Opção A: RabbitMQ
### Opção B: Kafka
### Opção C: PostgreSQL LISTEN/NOTIFY

### Escolha: C
```

---

## Cenário-Based

### Quando Performa Bem
- Feature com múltiplos fluxos de usuário (ex: checkout, onboarding, recuperação de senha)
- O foco é comportamento observável, não estrutura interna
- Stakeholders não-técnicos precisam validar o escopo

### Quando Performa Mal
- A mudança é puramente técnica (ex: atualizar versão de lib, otimizar query)
- Não há variação de comportamento do usuário

### Estrutura
```
## Cenário 1: Usuário Novo Faz Login
DADO que o usuário não existe
QUANDO ele tenta fazer login
ENTÃO ele é redirecionado para cadastro
E recebe email de boas-vindas

## Cenário 2: Usuário Existente Faz Login
DADO que o usuário existe e está ativo
QUANDO ele faz login com credenciais válidas
ENTÃO ele recebe token JWT
E é redirecionado para dashboard

## Cenário 3: Usuário Bloqueado
DADO que o usuário está bloqueado
QUANDO ele tenta fazer login
ENTÃO ele vê mensagem de conta suspensa
E recebe email com instruções
```

---

## Escolhendo o Tipo

Use este fluxo:

```
A decisão técnica principal já está tomada?
├── Não → Árvore de Decisões
│
└── Sim → A mudança afeta comportamento do usuário?
    ├── Sim → Cenário-Based
    │
    └── Não → Os módulos são independentes?
        ├── Sim → Por Domínio
        │
        └── Não → Faseado
```

---

## Metadado no Plano

Todo plano deve declarar seu tipo no front matter:

```yaml
---
tipo: faseado | por-dominio | arvore-decisoes | cenario-based
---
```

Isso permite que agentes executores saibam como priorizar a leitura do plano antes de começar.
