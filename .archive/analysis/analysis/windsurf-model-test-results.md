# Resultados dos Testes com Modelos WindSurf

## Data
2026-05-17

## Configuração Atualizada
- **API Key**: ✅ Válida (sk-ws-01-...)
- **CSRF Token**: ✅ Extraído automaticamente do processo
- **Porta**: 33945 (Connect Protocol)
- **Plano**: Free (TEAMS_TIER_DEVIN_FREE)

## Modelos Testados

### ✅ SWE-1.6 Slow (`swe-1-6-slow`)
- **Status**: FUNCIONANDO
- **Créditos**: 1.0x (custo padrão)
- **Resultado**: Processou mensagem e retornou resposta completa
- **Steps gerados**: 8
  1. RETRIEVE_MEMORY
  2. USER_INPUT
  3. PLANNER_RESPONSE ("I'll check the key documentation files...")
  4. VIEW_FILE
  5. VIEW_FILE
  6. VIEW_FILE
  7. CHECKPOINT
  8. PLANNER_RESPONSE (resposta completa sobre OrchestraOS)

### ❌ Kimi K2.6 (`kimi-k2-6`)
- **Status**: DISABLED no plano Free
- **Não disponível** para uso

### ❌ SWE-1.6 Fast (`swe-1-6-fast`)
- **Status**: DISABLED no plano Free

## Modelos Disponíveis no Plano Free

### Modelos ENABLED:
1. `swe-1-6-slow` - SWE-1.6 Slow ✅
2. `MODEL_CLAUDE_4_SONNET_BYOK` - Claude Sonnet 4 (Bring Your Own Key)
3. `MODEL_CLAUDE_4_SONNET_THINKING_BYOK` - Claude Sonnet 4 Thinking (BYOK)
4. `MODEL_CLAUDE_4_OPUS_BYOK` - Claude Opus 4 (BYOK)
5. `MODEL_CLAUDE_4_OPUS_THINKING_BYOK` - Claude Opus 4 Thinking (BYOK)

### Aliases disponíveis:
- `MODEL_ALIAS_CASCADE_BASE` (0.5x créditos)
- `MODEL_ALIAS_VISTA` (1.0x créditos)
- `MODEL_ALIAS_SHAMU` (0.25x créditos)

## Conclusão

✅ **Streaming funcionando perfeitamente!**
✅ **Sessão persistente (cascadeId reutilizável)**
✅ **Modelo SWE-1.6 Slow processou com sucesso**
✅ **API key válida e funcionando**

⚠️ **Limitações do plano Free:**
- Kimi K2.6 não disponível
- SWE-1.6 Fast não disponível
- Apenas SWE-1.6 Slow disponível
- 2500 créditos mensais de prompt
- 500 créditos mensais de flow

## Próximo Passo

Integrar o streaming no WindAgent Backend usando:
1. Modelo `swe-1-6-slow` como padrão
2. Sessões persistentes
3. Streaming real (sem polling)
