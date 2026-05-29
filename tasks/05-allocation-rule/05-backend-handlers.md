# 05-allocation-rule/05 — Handlers: remover valor mensal fixo e ajustar KPIs

## Objetivo
Adaptar os handlers HTTP ao novo modelo: sem `monthly_amount` projetado e com os
novos status no resumo do backtest.

## Pré-requisitos
- 05-allocation-rule/02

## Passos
1. `internal/httpapi/backtest_handler.go`:
   - `loadGoalPlan`: parar de calcular `perMonth`/`MonthlyAmount`; montar cada
     `backtest.Allocation` só com `AccountID` e `Percentage`.
   - `buildResult`:
     - `vaultBal += mv.Amount` para **todos** os movimentos (skip tem `Amount = 0`),
       o que já soma corretamente os `COMPLETED`.
     - `completed` conta apenas `StatusCompleted`; `failed = total - completed`.
     - `worst_account` = maior taxa de **não-`COMPLETED`** por conta.
     - `goal_reached = vaultBal >= target` (inalterado).
2. `internal/httpapi/goals_handler.go` (`GetGoal`):
   - Remover o cálculo `perMonth` e o campo `MonthlyAmount` de `GoalAllocationDTO`
     (não é mais projetável). Ajustar o exemplo Swagger do DTO.
3. Regenerar a doc Swagger:
   ```sh
   mise run swagger
   ```

## Critério de aceite
- [ ] `go build ./...` e `go test ./...` ok.
- [ ] `GET /goals/{id}` não retorna mais `monthly_amount` nas alocações.
- [ ] `POST /goals/{id}/backtest` retorna movimentos com os novos status e
      `vault_balance` coerente (= soma dos COMPLETED).

## Referências PRD
- RF-04, RF-05, RF-06
