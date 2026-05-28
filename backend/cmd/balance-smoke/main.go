// Command balance-smoke é um utilitário descartável: carrega a config, pega
// um account_id real de bank_accounts e imprime o saldo na POC_REFERENCE_DATE.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/moronisauner/hackai/backend/internal/balance"
	"github.com/moronisauner/hackai/backend/internal/config"
	"github.com/moronisauner/hackai/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	var accountID string
	if err := pool.QueryRow(ctx,
		`SELECT account_id FROM transaction_events GROUP BY account_id ORDER BY COUNT(*) DESC LIMIT 1`,
	).Scan(&accountID); err != nil {
		log.Fatal(err)
	}

	repo := &balance.Repo{Pool: pool}
	bal, err := repo.Reconstruct(ctx, accountID, cfg.POCReferenceDate)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("account_id=%s  balance@%s = %s\n",
		accountID, cfg.POCReferenceDate.Format("2006-01-02"), bal.StringFixed(2))
}
