package backtest

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

// constBalance devolve uma BalanceFn que sempre retorna o mesmo saldo real.
func constBalance(amount string) BalanceFn {
	return func(_ context.Context, _ string, _ time.Time) (decimal.Decimal, error) {
		return dec(amount), nil
	}
}

func TestRun(t *testing.T) {
	ctx := context.Background()

	t.Run("happy path: 3 meses todos COMPLETED", func(t *testing.T) {
		plan := Plan{
			StartDate:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			DurationMonths: 3,
			WithdrawalDay:  5,
			TargetAmount:   dec("3000"),
			Allocations: []Allocation{
				{AccountID: "A", Percentage: 100, MonthlyAmount: dec("1000")},
			},
		}
		got, err := Run(ctx, plan, constBalance("5000"))
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 3 {
			t.Fatalf("expected 3 movements, got %d", len(got))
		}
		for i, mv := range got {
			if mv.Status != StatusCompleted {
				t.Errorf("movement %d: expected COMPLETED, got %s", i, mv.Status)
			}
			if !mv.Amount.Equal(dec("1000")) {
				t.Errorf("movement %d: expected amount 1000, got %s", i, mv.Amount)
			}
		}
	})

	t.Run("falha total: saldo insuficiente em todos os meses", func(t *testing.T) {
		plan := Plan{
			StartDate:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			DurationMonths: 3,
			WithdrawalDay:  5,
			TargetAmount:   dec("3000"),
			Allocations: []Allocation{
				{AccountID: "A", Percentage: 100, MonthlyAmount: dec("1000")},
			},
		}
		got, err := Run(ctx, plan, constBalance("10"))
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 3 {
			t.Fatalf("expected 3 movements, got %d", len(got))
		}
		for i, mv := range got {
			if mv.Status != StatusFailed {
				t.Errorf("movement %d: expected FAILED, got %s", i, mv.Status)
			}
			if !mv.Amount.Equal(decimal.Zero) {
				t.Errorf("movement %d: expected amount 0, got %s", i, mv.Amount)
			}
		}
	})

	t.Run("falha intermitente: A so tem saldo em meses pares", func(t *testing.T) {
		// 2 contas, 50% cada, R$500 cada/mês, 4 meses. Start em Jan/2025.
		// Conta A: saldo alto só nos meses de número par (Fev, Abr); 0 nos ímpares.
		// Conta B: saldo sempre alto.
		balance := func(_ context.Context, accountID string, atDate time.Time) (decimal.Decimal, error) {
			if accountID == "A" {
				if int(atDate.Month())%2 == 0 {
					return dec("100000"), nil
				}
				return decimal.Zero, nil
			}
			return dec("100000"), nil // B
		}
		plan := Plan{
			StartDate:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			DurationMonths: 4,
			WithdrawalDay:  5,
			TargetAmount:   dec("4000"),
			Allocations: []Allocation{
				{AccountID: "A", Percentage: 50, MonthlyAmount: dec("500")},
				{AccountID: "B", Percentage: 50, MonthlyAmount: dec("500")},
			},
		}
		got, err := Run(ctx, plan, balance)
		if err != nil {
			t.Fatal(err)
		}

		aCompleted, aFailed, bCompleted := 0, 0, 0
		for _, mv := range got {
			switch {
			case mv.AccountID == "A" && mv.Status == StatusCompleted:
				aCompleted++
			case mv.AccountID == "A" && mv.Status == StatusFailed:
				aFailed++
			case mv.AccountID == "B" && mv.Status == StatusCompleted:
				bCompleted++
			}
		}
		if aCompleted != 2 || aFailed != 2 {
			t.Errorf("conta A: esperado 2 COMPLETED + 2 FAILED, got %d/%d", aCompleted, aFailed)
		}
		if bCompleted != 4 {
			t.Errorf("conta B: esperado 4 COMPLETED, got %d", bCompleted)
		}
	})

	t.Run("saldo decrescente: engine considera saques anteriores", func(t *testing.T) {
		// 1 conta, R$1000/mês, saldo real constante R$2500.
		// Mês 0: 2500>=1000 ok (sacado=1000). Mês 1: 2500-1000=1500>=1000 ok
		// (sacado=2000). Mês 2: 2500-2000=500<1000 FALHA.
		plan := Plan{
			StartDate:      time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			DurationMonths: 3,
			WithdrawalDay:  5,
			TargetAmount:   dec("3000"),
			Allocations: []Allocation{
				{AccountID: "A", Percentage: 100, MonthlyAmount: dec("1000")},
			},
		}
		got, err := Run(ctx, plan, constBalance("2500"))
		if err != nil {
			t.Fatal(err)
		}
		want := []string{StatusCompleted, StatusCompleted, StatusFailed}
		for i, mv := range got {
			if mv.Status != want[i] {
				t.Errorf("mês %d: esperado %s, got %s", i, want[i], mv.Status)
			}
		}
	})

	t.Run("withdrawal_day fora do mes: start 2024-01-31, day 31", func(t *testing.T) {
		// Documenta o comportamento de time.AddDate: ao somar meses sobre um dia
		// 31, fevereiro "transborda" para março (Jan31 + 1mês = Mar 2). Como o
		// engine então fixa o dia em WithdrawalDay=31 sobre base.Month(), todas as
		// datas caem em meses de 31 dias e ficam exatamente no dia 31.
		plan := Plan{
			StartDate:      time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			DurationMonths: 3,
			WithdrawalDay:  31,
			TargetAmount:   dec("3000"),
			Allocations: []Allocation{
				{AccountID: "A", Percentage: 100, MonthlyAmount: dec("1000")},
			},
		}
		got, err := Run(ctx, plan, constBalance("100000"))
		if err != nil {
			t.Fatal(err)
		}
		for i, mv := range got {
			day := mv.MovementDate.Day()
			lastValid := lastDayOfMonth(mv.MovementDate)
			if day != 31 && day != lastValid {
				t.Errorf("movimento %d: MovementDate %s não cai no dia 31 nem no último dia válido (%d)",
					i, mv.MovementDate.Format("2006-01-02"), lastValid)
			}
			t.Logf("movimento %d -> MovementDate=%s ReferenceMonth=%s",
				i, mv.MovementDate.Format("2006-01-02"), mv.ReferenceMonth.Format("2006-01-02"))
		}
	})
}

// lastDayOfMonth devolve o último dia do mês de t.
func lastDayOfMonth(t time.Time) int {
	firstOfNext := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).AddDate(0, 1, 0)
	return firstOfNext.AddDate(0, 0, -1).Day()
}
