# 03-frontend/07 — Página de resultados do backtest

## Objetivo
Tela `/goals/:goalID` que dispara (se necessário) e exibe o backtest: KPIs, tabela mês-a-mês e gráfico de evolução do baú (jornada item 6).

## Pré-requisitos
- 03-frontend/06
- 02-backend/13

## Passos
1. `npm install recharts`.
2. Em `src/pages/BacktestResultsPage.tsx`:
   - No mount, tentar `getBacktest(goalID)`. Se 404, chamar `runBacktest(goalID)` automaticamente.
   - Botão "Re-executar backtest" sempre disponível (chama `runBacktest`).
3. Layout em 3 blocos:
   - **KPIs** (cards):
     - "Meses cumpridos": `completed_months_pct` em %
     - "Saldo do baú": `vault_balance` (BRL) — destaque grande
     - "Meta": `target_amount` (BRL)
     - "Meta atingida?": badge verde/vermelho conforme `goal_reached`
     - "Conta com mais falhas": `worst_account_id` (enriquecer com `brand_name + number` via `getGoal`)
   - **Evolução do baú** (LineChart do recharts): eixo X = `month`, Y = `balance`. Linha alvo horizontal em `target_amount`.
   - **Tabela mês-a-mês**: pivot por mês (linhas) × conta (colunas), célula = ✅ ou ❌ + valor. Rolagem horizontal se muitas contas.
4. Botão "Voltar pra contas" → `/users/<userID>` (precisa pegar `user_id` via `getGoal`).

## Critério de aceite
- [ ] Abrir `/goals/<id>` mostra os 3 blocos.
- [ ] KPI "Meta atingida?" combina com `vault_balance >= target_amount`.
- [ ] Gráfico mostra linha crescente e linha alvo horizontal.
- [ ] Tabela mostra ✅/❌ por mês × conta.
- [ ] "Re-executar backtest" atualiza os dados sem reload.

## Referências PRD
- Jornada §4 item 6, RF-05, RF-06
