-- Carga inicial da massa de dados a partir de /raw_data/ (mount read-only do compose).
-- Escopo POC (PRD §7.1): apenas bank_accounts e transaction_events.
-- balances_history (1.1GB) e accounts_balances_events não são usados pela POC.
--
-- Os CSVs têm BOM UTF-8 (3 bytes) na primeira linha e colunas extras de CDC
-- (__source_ts_ms, __op, __extraction_type, ...) que não existem no schema.
-- Solução: COPY para tabela TEMP refletindo o CSV, depois INSERT só das colunas úteis.
-- HEADER true descarta a primeira linha (incluindo o BOM).

\echo '>>> Carregando bank_accounts...'

CREATE TEMP TABLE bank_accounts_raw (
    id                   TEXT,
    consent_id           TEXT,
    user_id              TEXT,
    branch_code          TEXT,
    number               TEXT,
    check_digit          TEXT,
    customer_document    TEXT,
    created_at           TEXT,
    updated_at           TEXT,
    status               TEXT,
    brand_name           TEXT,
    company_cnpj         TEXT,
    type                 TEXT,
    compe_code           TEXT,
    src_ts_ms            TEXT,
    op                   TEXT,
    delta_ingestion_date NUMERIC(20, 4),
    extraction_date_alt  TEXT,
    extraction_type      TEXT,
    extraction_date      TEXT
);

COPY bank_accounts_raw FROM '/raw_data/01_bank_accounts.csv' WITH (FORMAT csv, HEADER true);

INSERT INTO bank_accounts (
    id, consent_id, user_id, branch_code, number, check_digit, customer_document,
    created_at, updated_at, status, brand_name, company_cnpj, type, compe_code,
    delta_ingestion_date
)
SELECT
    id, consent_id, user_id, branch_code, number, check_digit, customer_document,
    created_at::timestamptz, updated_at::timestamptz,
    status, brand_name, company_cnpj, type, compe_code,
    delta_ingestion_date
FROM bank_accounts_raw;

\echo '>>> Carregando transaction_events...'

CREATE TEMP TABLE transaction_events_raw (
    id                                TEXT,
    transaction_id                    TEXT,
    user_id                           TEXT,
    consent_id                        TEXT,
    account_id                        TEXT,
    amount                            NUMERIC(18, 4),
    partie_cnpj_cpf                   TEXT,
    partie_compe_code                 TEXT,
    partie_branch_code                TEXT,
    partie_number                     TEXT,
    partie_check_digit                TEXT,
    customer_document                 TEXT,
    transaction_date_time             TEXT,
    created_at                        TEXT,
    updated_at                        TEXT,
    completed_authorised_payment_type TEXT,
    credit_debit_type                 TEXT,
    transaction_name                  TEXT,
    type                              TEXT,
    currency                          TEXT,
    partie_person_type                TEXT,
    src_ts_ms                         TEXT,
    op                                TEXT,
    delta_ingestion_date              NUMERIC(20, 4),
    extraction_date_alt               TEXT,
    extraction_type                   TEXT,
    extraction_date                   TEXT
);

COPY transaction_events_raw FROM '/raw_data/18_transactions_events.csv' WITH (FORMAT csv, HEADER true);

INSERT INTO transaction_events (
    id, transaction_id, user_id, consent_id, account_id, amount,
    partie_cnpj_cpf, partie_compe_code, partie_branch_code, partie_number, partie_check_digit,
    customer_document, transaction_date_time, created_at, updated_at,
    completed_authorised_payment_type, credit_debit_type, transaction_name, type, currency,
    partie_person_type, delta_ingestion_date
)
SELECT
    id, transaction_id, user_id, consent_id, account_id, amount,
    partie_cnpj_cpf, partie_compe_code, partie_branch_code, partie_number, partie_check_digit,
    customer_document,
    transaction_date_time::timestamptz, created_at::timestamptz, updated_at::timestamptz,
    completed_authorised_payment_type, credit_debit_type, transaction_name, type, currency,
    partie_person_type, delta_ingestion_date
FROM transaction_events_raw;

\echo '>>> Carga concluída.'
SELECT 'bank_accounts'      AS tabela, COUNT(*) AS linhas FROM bank_accounts
UNION ALL
SELECT 'transaction_events' AS tabela, COUNT(*) AS linhas FROM transaction_events;
