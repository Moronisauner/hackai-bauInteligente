// Package balance reconstrói o saldo de uma conta numa data de referência a
// partir dos eventos de transação (PRD §6): saldo inicial 0, soma de créditos
// menos débitos, considerando apenas transações TRANSACAO_EFETIVADA até a data.
package balance

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// Repo expõe as consultas de reconstrução de saldo.
type Repo struct {
	Pool *pgxpool.Pool
}

// reconstructOneSQL soma créditos − débitos efetivados de uma conta até refDate.
// O limite de data é inclusivo (<=). §11: filtrando BRL por padrão.
const reconstructOneSQL = `
SELECT COALESCE(SUM(
    CASE
        WHEN credit_debit_type = 'CREDITO' THEN amount
        WHEN credit_debit_type = 'DEBITO'  THEN -amount
        ELSE 0
    END
), 0)::numeric
FROM transaction_events
WHERE account_id = $1
  AND transaction_date_time <= $2
  AND completed_authorised_payment_type = 'TRANSACAO_EFETIVADA'
  AND currency = 'BRL';` // §11: filtrando BRL por padrão

// reconstructManySQL faz o mesmo agrupando por conta, em uma única query.
const reconstructManySQL = `
SELECT account_id, COALESCE(SUM(
    CASE
        WHEN credit_debit_type = 'CREDITO' THEN amount
        WHEN credit_debit_type = 'DEBITO'  THEN -amount
        ELSE 0
    END
), 0)::numeric
FROM transaction_events
WHERE account_id = ANY($1)
  AND transaction_date_time <= $2
  AND completed_authorised_payment_type = 'TRANSACAO_EFETIVADA'
  AND currency = 'BRL'
GROUP BY account_id;` // §11: filtrando BRL por padrão

// Reconstruct devolve o saldo reconstruído de uma conta na data refDate.
func (r *Repo) Reconstruct(ctx context.Context, accountID string, refDate time.Time) (decimal.Decimal, error) {
	var bal decimal.Decimal
	err := r.Pool.QueryRow(ctx, reconstructOneSQL, accountID, refDate).Scan(&bal)
	if err != nil {
		return decimal.Zero, fmt.Errorf("balance: reconstruct account %s: %w", accountID, err)
	}
	return bal, nil
}

// ReconstructMany devolve o saldo de várias contas em uma única query. Contas
// sem transações não aparecem no resultado da query; o chamador deve tratar
// ausência como saldo zero.
func (r *Repo) ReconstructMany(ctx context.Context, accountIDs []string, refDate time.Time) (map[string]decimal.Decimal, error) {
	out := make(map[string]decimal.Decimal, len(accountIDs))
	if len(accountIDs) == 0 {
		return out, nil
	}

	rows, err := r.Pool.Query(ctx, reconstructManySQL, accountIDs, refDate)
	if err != nil {
		return nil, fmt.Errorf("balance: reconstruct many: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var bal decimal.Decimal
		if err := rows.Scan(&id, &bal); err != nil {
			return nil, fmt.Errorf("balance: scanning row: %w", err)
		}
		out[id] = bal
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("balance: iterating rows: %w", err)
	}

	// Garante saldo zero explícito para contas sem transações.
	for _, id := range accountIDs {
		if _, ok := out[id]; !ok {
			out[id] = decimal.Zero
		}
	}

	return out, nil
}
