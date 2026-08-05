<div align="center">

# OrchestraOS

**Orquestrador fino em Go que transforma critérios de aceite em uma DAG de unidades de trabalho.**

[![CI](https://github.com/levygit837-cyber/OrchestraOS/actions/workflows/ci.yml/badge.svg)](https://github.com/levygit837-cyber/OrchestraOS/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.24-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Architecture](https://img.shields.io/badge/architecture-6%20gates-6b7280)](#gates-arquiteturais)

</div>

## Navegação rápida

[Pivot](#o-pivot) · [Fluxo](#fluxo-executável) · [Arquitetura](#arquitetura) · [Execução](#rodar) · [Limites](#limites-atuais)

## O pivot

O OrchestraOS começou com uma arquitetura maior do que a capacidade executável justificava. O
projeto foi reduzido a um pipeline legível, com regras que o CI consegue verificar. Essa decisão
removeu dezenas de milhares de linhas de arquitetura prematura e consolidou uma trilha menor para
planejar e executar uma tarefa.

O projeto está em evolução e não é uma plataforma autônoma completa. Seu valor atual está na
simplificação, nas fronteiras explícitas e na supervisão de mudanças produzidas por agentes de
desenvolvimento.

## Fluxo executável

```text
Task
  → planner heurístico
  → DAG de WorkUnits
  → ordenação topológica
  → runtime fake, Gemini ou DeepSeek
  → RunResult
```

O CLI aceita um título e ao menos dois critérios. Marcadores `[after: 1,2]` expressam dependências
entre critérios antes da construção da DAG.

```bash
go build -o orchestraos ./cmd/orchestraos

./orchestraos run "Adicionar login" \
  "Criar formulário" \
  "Adicionar serviço de autenticação" \
  "[after: 1,2] Validar integração"
```

O provider padrão é `fake`. Providers reais exigem `GEMINI_API_KEY` ou `DEEPSEEK_API_KEY`:

```bash
./orchestraos run --provider gemini --model gemini-2.5-flash \
  "Analisar tarefa" "Propor solução" "[after: 1] Revisar resultado"
```

## Arquitetura

```text
cmd/orchestraos/       # CLI e composição
internal/
├── domain/            # tipos puros
├── planner/           # critérios → DAG por heurística local
├── executor/          # execução sequencial em ordem topológica
├── runtime/           # contrato e runtime fake
├── provider/          # adapters Gemini e DeepSeek
├── store/             # contrato e persistência in-memory
├── decomposer/        # decomposição LLM ainda fora do caminho principal
├── daggen/            # construção e validação de grafos
├── assignment/        # atribuição experimental de agentes
└── event/             # estruturas de eventos ainda não compostas

tests/architecture/    # invariantes verificadas por AST
docs/adr/              # decisões e pivots arquiteturais
```

As dependências fluem para `domain`; SQL permanece confinado a `store`; o orquestrador apenas compõe
planner, executor, runtime e persistência.

## Gates arquiteturais

Seis testes impedem regressões estruturais:

| Gate | Regra verificada |
|---|---|
| Direção de dependências | imports entre packages seguem a lista permitida |
| Pureza do domínio | `domain/` não importa packages internos |
| Tamanho de package | nenhum package ultrapassa o orçamento definido |
| Tamanho de função | corpos de função permanecem dentro do limite |
| Confinamento de SQL | SQL existe somente em `store/` |
| Estado global | não há globais mutáveis |

## Rodar

Pré-requisito: Go 1.24.

```bash
go build ./...
go test ./... -race -count=1
go vet ./...
```

Com `golangci-lint` instalado:

```bash
make check
```

No snapshot auditado, havia 61 funções de teste e 3.013 linhas Go de produção. O workflow do commit
`586850b` passou build, race tests, vet, lint e os seis gates em 22 de julho de 2026. A verificação
local não foi repetida nesta máquina porque o toolchain Go não está instalado.

## Limites atuais

- o runtime padrão é fake e sempre devolve sucesso;
- Gemini e DeepSeek retornam texto, mas o resultado não é validado contra os critérios de aceite;
- a DAG é executada de forma sequencial;
- a única implementação de store usada pelo CLI é in-memory;
- a decomposição semântica por LLM, streaming, assignment e eventos possuem módulos/testes, mas não
  estão conectados ao caminho principal;
- não há agentes isolados editando workspace, policy engine, sandbox ou painel web;
- PostgreSQL permanece planejado, sem implementação de store ligada ao CLI.

## Documentação

- [Project canvas](docs/canvas/project-canvas.md)
- [ADR do Thin Orchestrator](docs/adr/0020-thin-orchestrator-pipeline.md)
- [ADR da geração de DAG](docs/adr/0021-agent-based-dag-generation.md)
- [Padrões de código](docs/development/CODING_STANDARDS.md)
- [Instruções para agentes](AGENTS.md)

## Desenvolvimento

O pivot e as implementações foram conduzidos com agentes de desenvolvimento, incluindo PRs do Devin.
O trabalho humano incluiu escolha do alvo arquitetural, redução de escopo, revisão, integração e uso
dos gates para aceitar ou rejeitar mudanças.

## Licença

O repositório não declara atualmente uma licença de reutilização.
