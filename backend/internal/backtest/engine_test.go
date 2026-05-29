package backtest

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

// phasedBalance constrói uma BalanceFn a partir do saldo de cada conta nos dois
// instantes que a engine consulta a cada mês:
//   - [0] open  = primeiro instante do mês (dia 1, 00:00:00.000000000); como o
//     saque ocorre sempre no dia 1, este é também o saldo no dia do saque.
//   - [1] close = último instante do mês (…23:59:59.999999999, Nanosecond != 0)
//
// O saldo independe do mês de competência, o que basta para exercitar a regra
// (evolução → fatia → limite do disponível) e a acumulação entre meses.
func phasedBalance(byAccount map[string][2]string) BalanceFn {
	return func(_ context.Context, accountID string, at time.Time) (decimal.Decimal, error) {
		v := byAccount[accountID]
		idx := 0 // open / dia do saque
		if at.Nanosecond() != 0 {
			idx = 1 // close
		}
		return dec(v[idx]), nil
	}
}

func TestRun(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("evolução positiva, saldo sobra → COMPLETED", func(t *testing.T) {
		// open 5000 → close 6000: evolução 1000; pct 50 → alvo 500; disponível
		// (= saldo de abertura, pois o saque é no dia 1) 5000.
		plan := Plan{
			StartDate:      start,
			DurationMonths: 1,
			TargetAmount:   dec("10000"),
			Allocations:    []Allocation{{AccountID: "A", Percentage: 50}},
		}
		got, err := Run(ctx, plan, phasedBalance(map[string][2]string{
			"A": {"5000", "6000"},
		}))
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 movement, got %d", len(got))
		}
		if got[0].Status != StatusCompleted {
			t.Errorf("expected COMPLETED, got %s", got[0].Status)
		}
		if !got[0].Amount.Equal(dec("500")) {
			t.Errorf("expected amount 500, got %s", got[0].Amount)
		}
	})

	t.Run("sem evolução → SKIPPED_NO_GROWTH (estável e em queda)", func(t *testing.T) {
		// EQ: open == close (estável). DROP: close < open (caiu). Ambas SKIPPED.
		plan := Plan{
			StartDate:      start,
			DurationMonths: 1,
			TargetAmount:   dec("10000"),
			Allocations: []Allocation{
				{AccountID: "EQ", Percentage: 50},
				{AccountID: "DROP", Percentage: 50},
			},
		}
		got, err := Run(ctx, plan, phasedBalance(map[string][2]string{
			"EQ":   {"2000", "2000"},
			"DROP": {"2000", "1000"},
		}))
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 movements, got %d", len(got))
		}
		for _, mv := range got {
			if mv.Status != StatusSkipped {
				t.Errorf("conta %s: esperado SKIPPED_NO_GROWTH, got %s", mv.AccountID, mv.Status)
			}
			if !mv.Amount.Equal(decimal.Zero) {
				t.Errorf("conta %s: esperado amount 0, got %s", mv.AccountID, mv.Amount)
			}
		}
	})

	t.Run("evolução positiva, saldo no dia do saque < alvo → PARTIAL", func(t *testing.T) {
		// open 300 → close 1300: evolução 1000; pct 100 → alvo 1000; disponível
		// (= saldo de abertura) só 300.
		plan := Plan{
			StartDate:      start,
			DurationMonths: 1,
			TargetAmount:   dec("10000"),
			Allocations:    []Allocation{{AccountID: "A", Percentage: 100}},
		}
		got, err := Run(ctx, plan, phasedBalance(map[string][2]string{
			"A": {"300", "1300"},
		}))
		if err != nil {
			t.Fatal(err)
		}
		if got[0].Status != StatusPartial {
			t.Errorf("expected PARTIAL, got %s", got[0].Status)
		}
		if !got[0].Amount.Equal(dec("300")) {
			t.Errorf("expected amount 300, got %s", got[0].Amount)
		}
	})

	t.Run("evolução positiva, disponível <= 0 → FAILED_INSUFFICIENT_BALANCE", func(t *testing.T) {
		// open 0 → close 1000: evolução 1000; pct 100 → alvo 1000; disponível
		// (= saldo de abertura) 0.
		plan := Plan{
			StartDate:      start,
			DurationMonths: 1,
			TargetAmount:   dec("10000"),
			Allocations:    []Allocation{{AccountID: "A", Percentage: 100}},
		}
		got, err := Run(ctx, plan, phasedBalance(map[string][2]string{
			"A": {"0", "1000"},
		}))
		if err != nil {
			t.Fatal(err)
		}
		if got[0].Status != StatusFailed {
			t.Errorf("expected FAILED, got %s", got[0].Status)
		}
		if !got[0].Amount.Equal(decimal.Zero) {
			t.Errorf("expected amount 0, got %s", got[0].Amount)
		}
	})

	t.Run("acumulação entre meses: reservas anteriores reduzem o disponível", func(t *testing.T) {
		// 1 conta, pct 100, 3 meses. Cada mês: evolução 1000 (alvo 1000).
		// Saldo de abertura (= dia do saque) constante 2500.
		//   mês 0: disp 2500 >= 1000 → COMPLETED (reservado 1000)
		//   mês 1: disp 2500-1000=1500 >= 1000 → COMPLETED (reservado 2000)
		//   mês 2: disp 2500-2000=500 < 1000 → PARTIAL (amount 500)
		plan := Plan{
			StartDate:      start,
			DurationMonths: 3,
			TargetAmount:   dec("10000"),
			Allocations:    []Allocation{{AccountID: "A", Percentage: 100}},
		}
		got, err := Run(ctx, plan, phasedBalance(map[string][2]string{
			"A": {"2500", "3500"},
		}))
		if err != nil {
			t.Fatal(err)
		}
		wantStatus := []string{StatusCompleted, StatusCompleted, StatusPartial}
		wantAmount := []string{"1000", "1000", "500"}
		for i, mv := range got {
			if mv.Status != wantStatus[i] {
				t.Errorf("mês %d: esperado %s, got %s", i, wantStatus[i], mv.Status)
			}
			if !mv.Amount.Equal(dec(wantAmount[i])) {
				t.Errorf("mês %d: esperado amount %s, got %s", i, wantAmount[i], mv.Amount)
			}
		}
	})

	t.Run("MovementDate é sempre o dia 1 do mês de competência", func(t *testing.T) {
		// O saque ocorre sempre no dia 1, mesmo quando StartDate cai noutro dia:
		// MovementDate deve coincidir com ReferenceMonth (primeiro dia do mês).
		plan := Plan{
			StartDate:      time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			DurationMonths: 3,
			TargetAmount:   dec("10000"),
			Allocations:    []Allocation{{AccountID: "A", Percentage: 50}},
		}
		got, err := Run(ctx, plan, phasedBalance(map[string][2]string{
			"A": {"100000", "200000"},
		}))
		if err != nil {
			t.Fatal(err)
		}
		for i, mv := range got {
			if mv.MovementDate.Day() != 1 {
				t.Errorf("movimento %d: MovementDate %s não cai no dia 1",
					i, mv.MovementDate.Format("2006-01-02"))
			}
			if !mv.MovementDate.Equal(mv.ReferenceMonth) {
				t.Errorf("movimento %d: MovementDate %s difere de ReferenceMonth %s",
					i, mv.MovementDate.Format("2006-01-02"), mv.ReferenceMonth.Format("2006-01-02"))
			}
		}
	})
}
