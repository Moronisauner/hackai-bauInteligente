package balance

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// Estes são testes de integração contra o Postgres do compose. Pulam quando
// DATABASE_URL não está setada. Cada sub-test usa um account_id sintético com
// prefixo conhecido e limpa seus próprios eventos no fim — a massa real nunca
// é tocada (só removemos linhas com account_id LIKE 'test-synthetic-%').

const synthPrefix = "test-synthetic-"

// refDate de referência usada em todos os sub-tests (meia-noite UTC).
var refDate = time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping balance integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connecting to db: %v", err)
	}
	return pool
}

// insertEvent insere um transaction_event sintético com os campos relevantes
// para o cálculo de saldo. id é derivado de accountID + idx para ser único.
func insertEvent(t *testing.T, pool *pgxpool.Pool, accountID string, idx int, when time.Time, cdType, payType string, amount string) {
	t.Helper()
	id := accountID + "-evt-" + decimal.NewFromInt(int64(idx)).String()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO transaction_events
		    (id, transaction_id, user_id, consent_id, account_id, amount,
		     transaction_date_time, completed_authorised_payment_type, credit_debit_type, currency)
		VALUES ($1, $1, 'synth-user', 'synth-consent', $2, $3, $4, $5, $6, 'BRL')`,
		id, accountID, amount, when, payType, cdType)
	if err != nil {
		t.Fatalf("inserting synthetic event: %v", err)
	}
}

// cleanup remove todos os eventos sintéticos do accountID dado.
func cleanup(t *testing.T, pool *pgxpool.Pool, accountID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		`DELETE FROM transaction_events WHERE account_id = $1`, accountID)
	if err != nil {
		t.Fatalf("cleaning up synthetic events: %v", err)
	}
}

func TestReconstruct(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	repo := &Repo{Pool: pool}
	ctx := context.Background()

	// Massa real deve permanecer intacta após todos os sub-tests.
	before := countRealEvents(t, pool)

	beforeRef := refDate.Add(-30 * 24 * time.Hour) // ~1 mês antes
	afterRef := refDate.Add(30 * 24 * time.Hour)   // ~1 mês depois

	t.Run("credito antes da refDate entra", func(t *testing.T) {
		acc := synthPrefix + "credit-before"
		defer cleanup(t, pool, acc)
		insertEvent(t, pool, acc, 1, beforeRef, "CREDITO", "TRANSACAO_EFETIVADA", "100.00")

		got, err := repo.Reconstruct(ctx, acc, refDate)
		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, got, "100")
	})

	t.Run("debito antes da refDate subtrai", func(t *testing.T) {
		acc := synthPrefix + "debit-before"
		defer cleanup(t, pool, acc)
		insertEvent(t, pool, acc, 1, beforeRef, "CREDITO", "TRANSACAO_EFETIVADA", "100.00")
		insertEvent(t, pool, acc, 2, beforeRef, "DEBITO", "TRANSACAO_EFETIVADA", "30.00")

		got, err := repo.Reconstruct(ctx, acc, refDate)
		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, got, "70")
	})

	t.Run("credito apos a refDate nao entra", func(t *testing.T) {
		acc := synthPrefix + "credit-after"
		defer cleanup(t, pool, acc)
		insertEvent(t, pool, acc, 1, beforeRef, "CREDITO", "TRANSACAO_EFETIVADA", "100.00")
		insertEvent(t, pool, acc, 2, afterRef, "CREDITO", "TRANSACAO_EFETIVADA", "500.00")

		got, err := repo.Reconstruct(ctx, acc, refDate)
		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, got, "100")
	})

	t.Run("evento nao efetivado nao entra", func(t *testing.T) {
		acc := synthPrefix + "not-effective"
		defer cleanup(t, pool, acc)
		insertEvent(t, pool, acc, 1, beforeRef, "CREDITO", "TRANSACAO_EFETIVADA", "100.00")
		insertEvent(t, pool, acc, 2, beforeRef, "CREDITO", "TRANSACAO_PROCESSANDO", "1000.00")

		got, err := repo.Reconstruct(ctx, acc, refDate)
		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, got, "100")
	})

	t.Run("evento exatamente na refDate entra (limite inclusivo)", func(t *testing.T) {
		acc := synthPrefix + "exactly-ref"
		defer cleanup(t, pool, acc)
		insertEvent(t, pool, acc, 1, refDate, "CREDITO", "TRANSACAO_EFETIVADA", "7.00")

		got, err := repo.Reconstruct(ctx, acc, refDate)
		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, got, "7")
	})

	// Verifica que a massa real não mudou.
	if after := countRealEvents(t, pool); after != before {
		t.Fatalf("real transaction mass changed: before=%d after=%d", before, after)
	}
}

func countRealEvents(t *testing.T, pool *pgxpool.Pool) int64 {
	t.Helper()
	var n int64
	if err := pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM transaction_events WHERE account_id NOT LIKE $1`, synthPrefix+"%",
	).Scan(&n); err != nil {
		t.Fatalf("counting real events: %v", err)
	}
	return n
}

func assertEqual(t *testing.T, got decimal.Decimal, want string) {
	t.Helper()
	w := decimal.RequireFromString(want)
	if !got.Equal(w) {
		t.Fatalf("balance mismatch: got %s, want %s", got.String(), w.String())
	}
}
