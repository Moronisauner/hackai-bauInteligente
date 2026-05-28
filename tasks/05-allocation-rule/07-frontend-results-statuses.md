# 05-allocation-rule/07 — Frontend: 4 status na tabela de resultados

## Objetivo
Distinguir cheio / parcial / sem evolução / sem saldo na tabela mês a mês do backtest.

## Pré-requisitos
- 05-allocation-rule/05
- 05-allocation-rule/06

## Passos
1. `src/pages/BacktestResultsPage.tsx`, componente `MonthlyTable`:
   - Substituir o booleano `ok = status === 'COMPLETED'` por um mapa de
     status → { ícone, cor, label }:
     - `COMPLETED` → ✅ verde
     - `PARTIAL` → ⚠️ âmbar (mostrar valor reservado)
     - `SKIPPED_NO_GROWTH` → ➖ cinza ("sem evolução")
     - `FAILED_INSUFFICIENT_BALANCE` → ❌ vermelho
   - Manter o valor formatado em BRL ao lado.
2. Se houver legenda/KPIs que assumam binário, ajustar o texto.

## Critério de aceite
- [ ] Rodar um backtest que produza os quatro status e ver cada um renderizado de
      forma distinta (cor + ícone).
- [ ] `npm run build` sem erros.

## Referências PRD
- RF-06
