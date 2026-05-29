// Package backtest contém a engine pura do backtest do baú (PRD RF-04, RF-05, §7.2).
//
// A engine NÃO toca o banco: recebe um Plan e uma BalanceFn injetável que
// devolve o saldo histórico real de uma conta numa data.
//
// A reserva mensal de cada conta é uma fatia da EVOLUÇÃO da própria conta no
// mês (RF-04, §11): mede-se a evolução = saldo(fim do mês) - saldo(início do
// mês) e reserva-se `evolution × percentage / 100`. Só evolução positiva gera
// reserva (SKIPPED_NO_GROWTH quando o mês não cresce); como a fatia reservada é
// parte do próprio crescimento do mês, ela é sempre reservada por inteiro
// (COMPLETED). O saque ocorre sempre no dia 1 do mês de competência
// (movementDate = primeiro dia do mês).
//
// Cada alocação é avaliada isoladamente (RF-05).
package backtest

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// Status possíveis de um movimento.
const (
	StatusCompleted = "COMPLETED"         // reservou a fatia da evolução do mês
	StatusSkipped   = "SKIPPED_NO_GROWTH" // conta não evoluiu no mês
)

// Allocation é uma alocação de conta: o percentual é a fatia da evolução
// mensal da conta que vai para a meta (RF-04).
type Allocation struct {
	AccountID  string
	Percentage int
}

// Plan descreve o backtest a executar.
type Plan struct {
	StartDate      time.Time
	DurationMonths int
	TargetAmount   decimal.Decimal
	Allocations    []Allocation
}

// Movement é o resultado de uma tentativa de saque de uma conta num mês.
type Movement struct {
	ReferenceMonth time.Time       // primeiro dia do mês de competência
	MovementDate   time.Time       // data efetiva (sempre o dia 1 do mês de competência)
	AccountID      string          // conta-fonte
	Amount         decimal.Decimal // valor reservado (0 em SKIPPED)
	Status         string          // StatusCompleted | StatusSkipped
}

// BalanceFn devolve o saldo histórico real de uma conta numa data.
type BalanceFn func(ctx context.Context, accountID string, atDate time.Time) (decimal.Decimal, error)

// Run executa o backtest e devolve a sequência de movimentos, na ordem
// (mês, alocação).
func Run(ctx context.Context, plan Plan, balance BalanceFn) ([]Movement, error) {
	movements := make([]Movement, 0, plan.DurationMonths*len(plan.Allocations))

	for i := 0; i < plan.DurationMonths; i++ {
		// O saque ocorre sempre no dia 1: movementDate = primeiro dia do mês de
		// competência (StartDate + i meses). §11: dia exato, sem ajuste.
		base := plan.StartDate.AddDate(0, i, 0)
		referenceMonth := time.Date(base.Year(), base.Month(), 1,
			0, 0, 0, 0, plan.StartDate.Location())
		movementDate := referenceMonth
		// Janela da evolução (§11): saldo no primeiro instante do mês de
		// competência vs. saldo no último instante do mês.
		monthOpen := referenceMonth
		monthClose := referenceMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

		for _, alloc := range plan.Allocations {
			mv := Movement{
				ReferenceMonth: referenceMonth,
				MovementDate:   movementDate,
				AccountID:      alloc.AccountID,
			}

			openBal, err := balance(ctx, alloc.AccountID, monthOpen)
			if err != nil {
				return nil, err
			}
			closeBal, err := balance(ctx, alloc.AccountID, monthClose)
			if err != nil {
				return nil, err
			}
			evolution := closeBal.Sub(openBal)

			// Sem crescimento no mês → nada a reservar.
			if evolution.LessThanOrEqual(decimal.Zero) {
				mv.Status = StatusSkipped
				mv.Amount = decimal.Zero
				movements = append(movements, mv)
				continue
			}

			target := evolution.Mul(decimal.NewFromInt(int64(alloc.Percentage))).
				Div(decimal.NewFromInt(100)).Round(2)
			if target.LessThanOrEqual(decimal.Zero) {
				mv.Status = StatusSkipped
				mv.Amount = decimal.Zero
				movements = append(movements, mv)
				continue
			}

			// A fatia reservada é parte do próprio crescimento do mês, logo é
			// sempre reservada por inteiro.
			mv.Status = StatusCompleted
			mv.Amount = target
			movements = append(movements, mv)
		}
	}

	return movements, nil
}
