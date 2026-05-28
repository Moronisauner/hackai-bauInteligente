# 02-backend/11 — Engine de backtest (puro, testável)

## Objetivo
Função pura que, dado um plano e um provedor de saldo, produz a sequência de movimentos do backtest (sem tocar DB).

## Pré-requisitos
- 02-backend/05 (assinatura de saldo)

## Passos
1. Em `internal/backtest/engine.go`:
   ```go
   type Plan struct {
       StartDate      time.Time
       DurationMonths int
       WithdrawalDay  int
       TargetAmount   decimal.Decimal
       Allocations    []Allocation
   }
   type Allocation struct {
       AccountID    string
       Percentage   int
       MonthlyAmount decimal.Decimal // já calculado por RF-04
   }
   type Movement struct {
       ReferenceMonth time.Time      // primeiro dia do mês de competência
       MovementDate   time.Time      // data efetiva (= ReferenceMonth com dia = WithdrawalDay)
       AccountID      string
       Amount         decimal.Decimal
       Status         string         // "COMPLETED" | "FAILED_INSUFFICIENT_BALANCE"
   }
   type BalanceFn func(ctx context.Context, accountID string, atDate time.Time) (decimal.Decimal, error)
   func Run(ctx context.Context, plan Plan, balance BalanceFn) ([]Movement, error)
   ```
2. Regra de iteração:
   - Para cada `i` em `0..DurationMonths-1`:
     - `movementDate = StartDate.AddDate(0, i, 0)` ajustado pro dia = `WithdrawalDay`
       (manter dia exato; vide §11 do PRD — sem ajuste pra fim de semana/feriado nesta POC).
     - `referenceMonth = primeiro dia do mês de movementDate`
   - Para cada `Allocation` na ordem fornecida:
     - `bal := balance(ctx, alloc.AccountID, movementDate)`
     - Se `bal >= alloc.MonthlyAmount` → `Movement{Status: COMPLETED, Amount: MonthlyAmount}`
     - Senão → `Movement{Status: FAILED_INSUFFICIENT_BALANCE, Amount: 0}` (PRD §7.2 nota)
   - **Não** tenta cobrir falha de uma conta com outra (RF-05).
3. **Saldo dinâmico:** o backtest precisa simular que o débito sintético aconteceu. Decisão: `BalanceFn` recebe o saldo histórico real; subtrai o que o backtest já "sacou" via um `map[accountID]decimal.Decimal` acumulado interno ao `Run`. Ou seja:
   ```
   effectiveBalance = realBalance(account, movementDate) - withdrawnSoFar[account]
   ```
   Documentar essa decisão em comentário Go no topo do arquivo.

## Critério de aceite
- [ ] `go build ./internal/backtest/...` ok.
- [ ] `Run` é pura: não importa `database/sql` nem `pgx`.
- [ ] Função aceita `BalanceFn` injetável → testes não precisam de DB.

## Referências PRD
- RF-05, §7.2 (nota sobre amount=0 em FAILED), §11 (dia exato sem ajuste)
