# OrchestraOS

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)](https://postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white)](https://docker.com)
[![CI](https://img.shields.io/badge/CI-Passing-2ea44f?logo=githubactions&logoColor=white)](.github/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> **Sistema operacional de projeto** que orquestra múltiplos agentes de IA para decompor objetivos em tarefas executáveis, executá-las em sandbox isolada e manter rastreabilidade completa via event sourcing.

---

## O Problema

Coordenar agentes de IA em projetos reais é caótico: tarefas se perdem, contexto se dissolve, execução fica opaca. OrchestraOS resolve isso com um **orquestrador central**, **grafos de tarefas acíclicos (DAG)** e **ledger de eventos persistente** — tudo via CLI e banco de dados relacional.

---

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Core | Go 1.24 |
| CLI | Cobra |
| Banco de Dados | PostgreSQL 16 (Event Store + Estado Transacional) |
| Migrations | Goose |
| Comunicação | WebSocket (gorilla/websocket) |
| Planejamento | Google Gemini API |
| Validação | JSON Schema |
| Containerização | Docker Compose |
| CI/CD | GitHub Actions |

---

## Funcionalidades Principais

- **Decomposição Inteligente de Tarefas** — Converte objetivos em grafos acíclicos (DAG) de work units via planejador heurístico + Gemini
- **Máquina de Estados Operacional** — Transições validadas para Task, Run, AgentSession e Review com rollback automático
- **Event Sourcing** — Ledger imutável de eventos versionados com replay de estado
- **Orquestração Multi-Agente** — Coordenação centralizada com sandboxes isoladas (git worktree + Docker)
- **Sistema de Prompts Versionados** — Composição de system prompts e task prompts com hash de snapshot
- **Gates de Validação** — Reviews obrigatórios por tipo de gate (code, security, design, etc.)
- **CLI Completa** — Gerenciamento de tasks, runs, work units, agent sessions e eventos via terminal
- **Testes de Arquitetura** — Verificação automatizada de fronteiras entre módulos, pureza de domínio e anti-padrões

---

## Arquitetura

```
┌─────────────────────────────────────────────┐
│  CLI (Cobra)  │  GitHub API  │  WebSocket   │
├─────────────────────────────────────────────┤
│  Orchestrator (coordenação central)         │
├─────────────────────────────────────────────┤
│  Módulos Verticais (clean architecture)     │
│  Task · Run · WorkUnit · Agent · AgentSession │
│  TaskGraph · Prompt · Review · Trigger      │
├─────────────────────────────────────────────┤
│  Core (event store · state machine · transition) │
├─────────────────────────────────────────────┤
│  PostgreSQL  ·  JSON Schemas  ·  Docker     │
└─────────────────────────────────────────────┘
```

- **Vertical Slice Architecture** — cada módulo encapsula seus próprios models, repository, service e eventos
- **Dependency Inversion** — adapters bridge entre módulos sem violar fronteiras de importação
- **Golden Rules** — módulos nunca importam uns aos outros; apenas o orchestrator/bootstrap/cmd podem fazê-lo
- **ADR-driven** — 18 Architecture Decision Records documentam cada decisão estrutural

---

## Como Rodar

### Pré-requisitos
- Go 1.24+
- Docker & Docker Compose
- PostgreSQL client (opcional)

### 1. Clone e infraestrutura
```bash
git clone https://github.com/levygit837-cyber/OrchestraOS.git
cd OrchestraOS
docker compose up -d   # sobe Postgres na porta 55432
```

### 2. Migrations
```bash
go run ./cmd/orchestraos migrate up
```

### 3. CLI
```bash
go run ./cmd/orchestraos --help
go run ./cmd/orchestraos task create --title "Implementar login" --description "..."
go run ./cmd/orchestraos task list
```

### 4. Testes
```bash
go test ./... -race          # todos os testes
go test ./tests/architecture/... -v   # testes de arquitetura
go test ./tests/contracts/... -v      # validação de schemas
```

---

## Estatísticas do Projeto

| Métrica | Valor |
|---------|-------|
| Linhas de Go | ~24.000 |
| Arquivos .go | 142 |
| Testes | 38 suites |
| Commits | 129 |
| ADRs | 18 |
| Módulos | 10 |
| Jobs de CI | 9 |

---

## Status

**MVP Foundation — Funcional e em evolução ativa**

- ✅ Core de event sourcing e state machine
- ✅ 10 módulos de domínio implementados
- ✅ CLI operacional
- ✅ CI/CD completo com lint, testes de arquitetura e verificação de contratos
- ✅ Docker Compose para desenvolvimento local
- 🔄 Em desenvolvimento: runtime real de agentes, painel web, NATS JetStream

---

## Aprendizados Técnicos

Este projeto foi construído do zero com foco em **maturidade de código**:

- **Arquitetura limpa na prática** — separação real entre domain, core e modules, não apenas em pastas
- **Testes de arquitetura** — go test validando que `internal/modules/*` nunca importam uns aos outros
- **Event Sourcing sem framework** — implementação própria de event store, replay e state machine
- **CI/CD como qualidade** — pipeline que quebra em anti-padrões (panic, fmt.Println, inline SQL, arquivos `utils.go`)
- **Documentação viva** — ADRs mantidos sincronizados com código via testes automatizados
- **Contratos primários** — JSON Schemas como fonte de verdade para eventos e payloads

---

## Documentação

- [docs/architecture/](docs/architecture/) — visão geral, modelo de domínio, protocolos
- [docs/adr/](docs/adr/) — decisões arquiteturais registradas
- [AGENTS.md](AGENTS.md) — regras para agentes de IA trabalhando no repositório

---

## Licença

[MIT](LICENSE)
