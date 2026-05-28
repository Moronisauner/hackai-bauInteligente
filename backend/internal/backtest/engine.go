// Package backtest contém a engine pura do backtest do baú (PRD RF-05, §7.2).
//
// A engine NÃO toca o banco: recebe um Plan e uma BalanceFn injetável que
// devolve o saldo histórico real de uma conta numa data. Para simular que os
// saques sintéticos do backtest de fato saíram da conta, a engine acumula
// internamente, por conta, quanto já foi "sacado" e calcula o saldo efetivo:
//
//	effectiveBalance = realBalance(account, movementDate) - withdrawnSoFar[account]
//
// Assim um saque que sucede num mês reduz o saldo disponível dos meses
// seguintes, sem que a BalanceFn precise saber do backtest.
//
// A movimentação NÃO tenta cobrir a falha de uma conta com outra (RF-05): cada
// alocação é avaliada isoladamente.
package backtest

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// Status possíveis de um movimento.
const (
	StatusCompleted = "COMPLETED"
	StatusFailed    = "FAILED_INSUFFICIENT_BALANCE"
)

// Allocation é uma alocação de conta com o valor mensal já calculado (RF-04).
type Allocation struct {
	AccountID     string
	Percentage    int
	MonthlyAmount decimal.Decimal
}

// Plan descreve o backtest a executar.
type Plan struct {
	StartDate      time.Time
	DurationMonths int
	WithdrawalDay  int
	TargetAmount   decimal.Decimal
	Allocations    []Allocation
}

// Movement é o resultado de uma tentativa de saque de uma conta num mês.
type Movement struct {
	ReferenceMonth time.Time       // primeiro dia do mês de competência
	MovementDate   time.Time       // data efetiva (mês de competência com dia = WithdrawalDay)
	AccountID      string          // conta-fonte
	Amount         decimal.Decimal // valor sacado (0 em falha)
	Status         string          // StatusCompleted | StatusFailed
}

// BalanceFn devolve o saldo histórico real de uma conta numa data.
type BalanceFn func(ctx context.Context, accountID string, atDate time.Time) (decimal.Decimal, error)

// Run executa o backtest e devolve a sequência de movimentos, na ordem
// (mês, alocação).
func Run(ctx context.Context, plan Plan, balance BalanceFn) ([]Movement, error) {
	withdrawn := make(map[string]decimal.Decimal) // acumulado sacado por conta
	movements := make([]Movement, 0, plan.DurationMonths*len(plan.Allocations))

	for i := 0; i < plan.DurationMonths; i++ {
		// movementDate = StartDate + i meses, mantendo o dia = WithdrawalDay.
		// §11: dia exato, sem ajuste para fim de semana/feriado.
		base := plan.StartDate.AddDate(0, i, 0)
		movementDate := time.Date(base.Year(), base.Month(), plan.WithdrawalDay,
			0, 0, 0, 0, plan.StartDate.Location())
		referenceMonth := time.Date(movementDate.Year(), movementDate.Month(), 1,
			0, 0, 0, 0, plan.StartDate.Location())

		for _, alloc := range plan.Allocations {
			realBal, err := balance(ctx, alloc.AccountID, movementDate)
			if err != nil {
				return nil, err
			}
			effective := realBal.Sub(withdrawn[alloc.AccountID])

			mv := Movement{
				ReferenceMonth: referenceMonth,
				MovementDate:   movementDate,
				AccountID:      alloc.AccountID,
			}
			if effective.GreaterThanOrEqual(alloc.MonthlyAmount) {
				mv.Status = StatusCompleted
				mv.Amount = alloc.MonthlyAmount
				withdrawn[alloc.AccountID] = withdrawn[alloc.AccountID].Add(alloc.MonthlyAmount)
			} else {
				mv.Status = StatusFailed
				mv.Amount = decimal.Zero
			}
			movements = append(movements, mv)
		}
	}

	return movements, nil
}
