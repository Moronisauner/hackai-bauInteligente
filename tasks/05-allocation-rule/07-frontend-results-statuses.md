# 05-allocation-rule/07 — Frontend: status na tabela de resultados

## Objetivo
Distinguir cheio / sem evolução na tabela mês a mês do backtest.

## Pré-requisitos
- 05-allocation-rule/05
- 05-allocation-rule/06

## Passos
1. `src/pages/BacktestResultsPage.tsx`, componente `MonthlyTable`:
   - Substituir o booleano `ok = status === 'COMPLETED'` por um mapa de
     status → { ícone, cor, label }:
     - `COMPLETED` → ✅ verde
     - `SKIPPED_NO_GROWTH` → ➖ cinza ("sem evolução")
   - Manter o valor formatado em BRL ao lado.
2. Se houver legenda/KPIs que assumam binário, ajustar o texto.

## Critério de aceite
- [ ] Rodar um backtest que produza os dois status e ver cada um renderizado de
      forma distinta (cor + ícone).
- [ ] `npm run build` sem erros.

## Referências PRD
- RF-06
