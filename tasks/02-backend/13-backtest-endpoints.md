# 02-backend/13 — Endpoints do backtest (executar + consultar)

## Objetivo
Expor o backtest via HTTP: executar e persistir movimentos, depois consultar com KPIs (RF-05, RF-06).

## Pré-requisitos
- 02-backend/10
- 02-backend/11

## Passos
1. `POST /goals/{goalID}/backtest`:
   - Carrega `goal`, `vault`, `allocations` do DB.
   - Monta `backtest.Plan` (incluindo `MonthlyAmount` por alocação).
   - Cria um `BalanceFn` que delega pra `balance.Repo.Reconstruct`.
   - Chama `backtest.Run`.
   - Numa transação: **DELETE** prévios `goal_vault_movements` daquele `vault_id` (re-executar é idempotente), depois INSERT em massa dos novos.
   - Retorna o resultado completo (vide GET abaixo).
2. `GET /goals/{goalID}/backtest`:
   - Lê `goal_vault_movements` do DB pelo `vault_id`.
   - Calcula em código (uma única passada):
     - `total_months = duration_months * len(allocations)` (linhas no plano)
     - `completed_count`, `failed_count`
     - `vault_balance` = SUM(`amount`) onde `status='COMPLETED'`
     - `goal_reached` = `vault_balance >= target_amount`
     - `worst_account` = `account_id` com maior taxa de falha (RF-06)
     - `monthly_series` = pra cada `reference_month`, `vault_balance_eom` acumulado
   - Resposta JSON:
     ```json
     {
       "summary": {
         "completed_months_pct": 0.83,
         "vault_balance": "8400.00",
         "target_amount": "10000.00",
         "goal_reached": false,
         "worst_account_id": "..."
       },
       "movements": [ { "reference_month": "...", "account_id": "...", "status": "...", "amount": "..." } ],
       "vault_evolution": [ { "month": "2024-06-01", "balance": "1000.00" }, ... ]
     }
     ```
3. Se não houver movimentos persistidos, `GET` retorna `404` com hint `"run POST /goals/<id>/backtest first"`.

## Critério de aceite
- [ ] `POST /goals/<id>/backtest` → 200, e `goal_vault_movements` no DB tem `duration_months * len(allocations)` linhas.
- [ ] Rodar `POST` duas vezes seguidas mantém o mesmo número de linhas (idempotente).
- [ ] `GET /goals/<id>/backtest` retorna `summary`, `movements`, `vault_evolution`.
- [ ] `SUM(amount) WHERE status='COMPLETED'` no DB == `summary.vault_balance`.

## Referências PRD
- RF-05, RF-06, §10 (coerência: soma de cumpridos = saldo do baú)
