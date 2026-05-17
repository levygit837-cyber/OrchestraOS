# ADR 0017: Servicos de Dominio para Dependencias Operacionais

**Status:** ⚠️ **DEPRECATED**

## Decisao Substituta

Este ADR foi **substituido** pela ADR 0022 (LLM-Optimized Module Architecture) e formalmente deprecado pela ADR 0024.

Os principios de "servicos como fronteira de comando" continuam validos, mas a implementacao migrou de `internal/services/` para módulos verticais em `internal/modules/*/service.go`.

Para a definicao atual da arquitetura de modulos, consulte:

- [ADR 0022: LLM-Optimized Module Architecture](/docs/adr/0022-llm-optimized-module-architecture.md)
- [ADR 0024: Deprecation of ADR 0017](/docs/adr/0024-deprecation-of-adr-0017.md)
- [ADR 0025: Module Standardization](/docs/adr/0025-module-standardization.md)
