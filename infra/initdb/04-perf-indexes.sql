-- Índice composto para o cálculo de saldo reconstruído (PRD §6 e §9).
-- Suporta o predicado WHERE account_id = ? AND transaction_date_time <= ? do recálculo.

CREATE INDEX IF NOT EXISTS idx_transaction_events_account_date
    ON transaction_events(account_id, transaction_date_time);
