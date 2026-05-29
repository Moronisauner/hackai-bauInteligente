// Package backtest contém a engine pura do backtest do baú (PRD RF-04, RF-05, §7.2).
//
// A engine NÃO toca o banco: recebe um Plan e uma BalanceFn injetável que
// devolve o saldo histórico real de uma conta numa data.
//
// A reserva mensal de cada conta é uma fatia da EVOLUÇÃO da própria conta no
// mês (RF-04, §11): mede-se a evolução = saldo(fim do mês) - saldo(início do
// mês) e reserva-se `evolution × percentage / 100`. Só evolução positiva gera
// reserva; mês sem crescimento é SKIPPED_NO_GROWTH.
//
// O saque ocorre sempre no dia 1 do mês de competência (movementDate =
// primeiro dia do mês). O alvo reservado é limitado pelo saldo disponível
// nesse dia. Para simular que as reservas sintéticas do backtest de fato
// saíram da conta, a engine acumula internamente, por conta, quanto já foi
// "reservado" e calcula o saldo disponível:
//
//	available = realBalance(account, movementDate) - reserved[account]
//
// Assim uma reserva que sucede num mês reduz o disponível dos meses seguintes,
// sem que a BalanceFn precise saber do backtest. Se o disponível for menor que
// o alvo, reserva-se o que houver (PARTIAL); se for <= 0, falha
// (FAILED_INSUFFICIENT_BALANCE).
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
	StatusPartial   = "PARTIAL"           // reservou menos que o alvo (saldo limitou)
	StatusSkipped   = "SKIPPED_NO_GROWTH" // conta não evoluiu no mês
	StatusFailed    = "FAILED_INSUFFICIENT_BALANCE"
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
	Amount         decimal.Decimal // valor reservado (0 em SKIPPED/FAILED)
	Status         string          // StatusCompleted | StatusPartial | StatusSkipped | StatusFailed
}

// BalanceFn devolve o saldo histórico real de uma conta numa data.
type BalanceFn func(ctx context.Context, accountID string, atDate time.Time) (decimal.Decimal, error)

// Run executa o backtest e devolve a sequência de movimentos, na ordem
// (mês, alocação).
func Run(ctx context.Context, plan Plan, balance BalanceFn) ([]Movement, error) {
	reserved := make(map[string]decimal.Decimal) // acumulado reservado por conta
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

			// movementDate == monthOpen (dia 1), logo o saldo no dia do saque é o
			// próprio openBal já consultado.
			available := openBal.Sub(reserved[alloc.AccountID])

			switch {
			case available.LessThanOrEqual(decimal.Zero):
				mv.Status = StatusFailed
				mv.Amount = decimal.Zero
			case available.GreaterThanOrEqual(target):
				mv.Status = StatusCompleted
				mv.Amount = target
				reserved[alloc.AccountID] = reserved[alloc.AccountID].Add(target)
			default:
				mv.Status = StatusPartial
				mv.Amount = available.Round(2)
				reserved[alloc.AccountID] = reserved[alloc.AccountID].Add(mv.Amount)
			}
			movements = append(movements, mv)
		}
	}

	return movements, nil
}
