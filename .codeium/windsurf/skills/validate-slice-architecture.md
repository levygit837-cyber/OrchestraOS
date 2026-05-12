---
description: Validar arquitetura slice (Vertical Slices) conforme ADR 0022
---

# Skill: Validar Arquitetura Slice

## Descrição

Valida a arquitetura slice (Vertical Slices) do projeto conforme ADR 0022, verificando que módulos em `internal/modules/*` não importam outros módulos, garantindo isolamento estrito.

## Regras Validadas

1. **Isolamento entre módulos**: Módulos em `internal/modules/*` NÃO podem importar outros módulos em `internal/modules/*` (ADR 0022: strict isolation)
2. **Auto-import proibido**: Módulos não podem importar a si mesmos
3. **Exceção para orchestration**: Apenas `core/orchestration` pode ter cross-module imports para execução

## Como Usar

```bash
go run scripts/validate-slice-architecture.go
```

## Comportamento

- Analisa TODOS os módulos em `internal/modules/*` obrigatoriamente
- Verifica somente as primeiras linhas (imports) para eficiência e performance
- Relata todas as violações encontradas com arquivo, linha, import e motivo
- Exit code 0 se passar, 1 se houver violações

## Saída

**Sucesso:**
```
Found 7 modules to validate
✅ All slice architecture rules passed!
```

**Falha:**
```
Found 7 modules to validate
❌ Found 2 architecture violations:

1. internal/modules/agent/service.go:15
   Import: github.com/levygit837-cyber/OrchestraOS/internal/modules/run
   Module: agent -> Target: run
   Reason: Modules cannot import other modules (ADR 0022: strict isolation)
```

## Implementação

- Script em Go que percorre recursivamente `internal/modules/*`
- Usa regex para extrair imports dos blocos `import (...)`
- Eficiente: lê apenas as linhas de import, não o arquivo inteiro
- Detecta violações baseado no caminho do import

## Quando Executar

- Antes de commits que adicionam novos módulos
- Em CI/CD para garantir que a arquitetura seja preservada
- Após refactors que podem ter introduzido imports cruzados
