# 01-infra/05 — Índice de performance em transaction_events

## Objetivo
Garantir que o cálculo de saldo reconstruído fique abaixo de 2s para contas com até 10k transações (RNF de §9).

## Pré-requisitos
- 01-infra/03

## Passos
1. Inspecionar índices existentes em `transaction_events`:
   ```
   docker compose exec postgres psql -U hackai -d hackai -c "\d transaction_events"
   ```
2. Se **não** existir índice cobrindo `(account_id, transaction_date_time)`, criar via `infra/initdb/04-perf-indexes.sql`:
   ```sql
   CREATE INDEX IF NOT EXISTS idx_transaction_events_account_date
       ON transaction_events(account_id, transaction_date_time);
   ```
3. Aplicar:
   - Em DB já populado: rodar o SQL direto via `docker compose exec postgres psql ... -f`.
   - Ou recriar volume: `cd infra && docker compose down -v && docker compose up -d`.

## Critério de aceite
- [ ] `\d transaction_events` lista `idx_transaction_events_account_date`.
- [ ] `EXPLAIN ANALYZE SELECT SUM(amount) FROM transaction_events WHERE account_id = '<um_id_real>' AND transaction_date_time <= NOW();` mostra **Index Scan** (não Seq Scan).

## Referências PRD
- §9 (performance: < 2s pra 10k transações)
