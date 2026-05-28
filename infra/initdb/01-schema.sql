
-- 
-- Contas bancárias
-- 

CREATE TABLE bank_accounts (
    id                   VARCHAR(64) PRIMARY KEY,
    consent_id           VARCHAR(64) NOT NULL,
    user_id              VARCHAR(64) NOT NULL,
    branch_code          VARCHAR(10),
    number               VARCHAR(20),
    check_digit          VARCHAR(2),
    customer_document    VARCHAR(20),
    created_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status               VARCHAR(20),
    brand_name           VARCHAR(100),
    company_cnpj         VARCHAR(14),
    type                 VARCHAR(50),
    compe_code           VARCHAR(10),
    delta_ingestion_date NUMERIC(20, 4)
);

-- Índices para os campos de relacionamento e filtros comuns
CREATE INDEX idx_bank_accounts_consent_id ON bank_accounts(consent_id);
CREATE INDEX idx_bank_accounts_user_id    ON bank_accounts(user_id);
CREATE INDEX idx_bank_accounts_status     ON bank_accounts(status);

-- Comentários para documentação
COMMENT ON TABLE bank_accounts IS 'Dados cadastrais das contas de depósito à vista, poupança e pagamento pré-pago dos clientes';
COMMENT ON COLUMN bank_accounts.status IS 'Ex: AVAILABLE';
COMMENT ON COLUMN bank_accounts.type IS 'Ex: CONTA_PAGAMENTO_PRE_PAGA, CONTA_DEPOSITO_A_VISTA, CONTA_POUPANCA';
COMMENT ON COLUMN bank_accounts.brand_name IS 'Nome da instituição, ex: Mercado Pago';
COMMENT ON COLUMN bank_accounts.compe_code IS 'Código COMPE da instituição financeira';

-- 
-- Dados de Saldo
-- 

CREATE TABLE balances_history (
    id                   VARCHAR(64) PRIMARY KEY,
    resource_id          VARCHAR(64) NOT NULL,
    consent_id           VARCHAR(64) NOT NULL,
    user_id              VARCHAR(64) NOT NULL,
    balances             JSONB NOT NULL,
    created_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    customer_document    VARCHAR(20),
    delta_ingestion_date NUMERIC(20, 4)
);

-- Índices para os campos de relacionamento e consultas comuns
CREATE INDEX idx_balances_history_resource_id ON balances_history(resource_id);
CREATE INDEX idx_balances_history_consent_id  ON balances_history(consent_id);
CREATE INDEX idx_balances_history_user_id     ON balances_history(user_id);
CREATE INDEX idx_balances_history_created_at  ON balances_history(created_at);

-- Índice GIN para consultas dentro do JSON de saldos
CREATE INDEX idx_balances_history_balances_gin ON balances_history USING GIN (balances);

-- Comentários para documentação
COMMENT ON TABLE balances_history IS 'Histórico consolidado de saldos extraído de payloads JSON da conta';
COMMENT ON COLUMN balances_history.balances IS 'Payload JSON com os saldos, ex: {"type":"CONTA_POUPA...}';


CREATE TABLE accounts_balances_events (
    id                            VARCHAR(64) PRIMARY KEY,
    resource_id                   VARCHAR(64) NOT NULL,
    consent_id                    VARCHAR(64) NOT NULL,
    user_id                       VARCHAR(64) NOT NULL,
    available_amount              NUMERIC(18, 4) NOT NULL,
    automatically_invested_amount NUMERIC(18, 4) NOT NULL,
    blocked_amount                NUMERIC(18, 4) NOT NULL,
    updated_at_origin             TIMESTAMP WITH TIME ZONE,
    updated_at                    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    customer_document             VARCHAR(20),
    updated_blocked_reason        VARCHAR(255)
);

-- Índices para os campos de relacionamento e consultas comuns
CREATE INDEX idx_accounts_balances_events_resource_id ON accounts_balances_events(resource_id);
CREATE INDEX idx_accounts_balances_events_consent_id  ON accounts_balances_events(consent_id);
CREATE INDEX idx_accounts_balances_events_user_id     ON accounts_balances_events(user_id);
CREATE INDEX idx_accounts_balances_events_updated_at  ON accounts_balances_events(updated_at);

-- Comentários para documentação
COMMENT ON TABLE accounts_balances_events IS 'Eventos de saldo das contas: valor disponível, bloqueado e automaticamente investido';
COMMENT ON COLUMN accounts_balances_events.available_amount IS 'Saldo disponível na conta';
COMMENT ON COLUMN accounts_balances_events.automatically_invested_amount IS 'Valor automaticamente investido';
COMMENT ON COLUMN accounts_balances_events.blocked_amount IS 'Valor bloqueado na conta';
COMMENT ON COLUMN accounts_balances_events.updated_at_origin IS 'Data/hora de atualização na origem (instituição)';
COMMENT ON COLUMN accounts_balances_events.updated_at IS 'Data/hora de atualização no nosso sistema';

-- 
-- 
-- 

CREATE TABLE transaction_events (
    id                                VARCHAR(64) PRIMARY KEY,
    transaction_id                    VARCHAR(64) NOT NULL,
    user_id                           VARCHAR(64) NOT NULL,
    consent_id                        VARCHAR(64) NOT NULL,
    account_id                        VARCHAR(64) NOT NULL,
    amount                            NUMERIC(18, 4) NOT NULL,
    partie_cnpj_cpf                   VARCHAR(14),
    partie_compe_code                 VARCHAR(10),
    partie_branch_code                VARCHAR(10),
    partie_number                     VARCHAR(20),
    partie_check_digit                VARCHAR(2),
    customer_document                 VARCHAR(20),
    transaction_date_time             TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at                        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at                        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_authorised_payment_type VARCHAR(30),
    credit_debit_type                 VARCHAR(10),
    transaction_name                  VARCHAR(255),
    type                              VARCHAR(50),
    currency                          CHAR(3),
    partie_person_type                VARCHAR(20),
    delta_ingestion_date              NUMERIC(20, 4)
);

-- Índices para os campos de relacionamento (úteis para filtros e joins futuros)
CREATE INDEX idx_transaction_events_transaction_id ON transaction_events(transaction_id);
CREATE INDEX idx_transaction_events_user_id        ON transaction_events(user_id);
CREATE INDEX idx_transaction_events_consent_id     ON transaction_events(consent_id);
CREATE INDEX idx_transaction_events_account_id     ON transaction_events(account_id);
CREATE INDEX idx_transaction_events_date_time      ON transaction_events(transaction_date_time);

-- Comentários para documentação
COMMENT ON TABLE transaction_events IS 'Eventos de transações financeiras';
COMMENT ON COLUMN transaction_events.completed_authorised_payment_type IS 'Ex: TRANSACAO_EFETIVADA';
COMMENT ON COLUMN transaction_events.credit_debit_type IS 'Ex: DEBITO, CREDITO';
COMMENT ON COLUMN transaction_events.type IS 'Ex: CONVENIO_ARRECADACAO';
COMMENT ON COLUMN transaction_events.partie_person_type IS 'Ex: PESSOA_JURIDICA, PESSOA_FISICA';
COMMENT ON COLUMN transaction_events.currency IS 'Código ISO 4217, ex: BRL';