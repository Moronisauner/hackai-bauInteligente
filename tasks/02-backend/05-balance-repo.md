# 02-backend/05 — Cálculo de saldo reconstruído

## Objetivo
Implementar a função core que reconstrói saldo de uma conta numa data `D`, conforme §6 do PRD.

## Pré-requisitos
- 02-backend/03
- 01-infra/05

## Passos
1. Em `internal/balance/repo.go`, expor:
   ```go
   type Repo struct { Pool *pgxpool.Pool }
   func (r *Repo) Reconstruct(ctx context.Context, accountID string, refDate time.Time) (decimal.Decimal, error)
   ```
2. Usar `github.com/shopspring/decimal` (não `float64`) para evitar erros de arredondamento. `go get github.com/shopspring/decimal`.
3. SQL exato (1 round-trip):
   ```sql
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
     AND completed_authorised_payment_type = 'TRANSACAO_EFETIVADA';
   ```
3. Função auxiliar `ReconstructMany(ctx, accountIDs []string, refDate)` que devolve `map[string]decimal.Decimal` — uma única query com `GROUP BY account_id` (será usada na listagem de contas).
4. Filtro de moeda fica fora desta task (vide §11 do PRD — pergunta em aberto). Se for trivial filtrar `currency='BRL'`, adicionar com comentário `// §11: filtrando BRL por padrão`.

## Critério de aceite
- [ ] `go build ./...` ok.
- [ ] Smoke manual: escrever um `cmd/balance-smoke/main.go` descartável que carrega config, pega um `account_id` real de `bank_accounts` e imprime o saldo na `POC_REFERENCE_DATE`. Conferir mentalmente contra `SELECT * FROM transaction_events WHERE account_id = '...' ORDER BY transaction_date_time DESC LIMIT 5;`.
- [ ] Query `EXPLAIN` (vide 01-infra/05) ainda usa Index Scan.

## Referências PRD
- §6 (regra inteira: créditos − débitos, só `TRANSACAO_EFETIVADA`, saldo inicial 0)
- §11 (BRL como default)
