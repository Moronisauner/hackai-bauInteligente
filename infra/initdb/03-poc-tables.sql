-- Tabelas da POC (PRD §7.2): objetivo, alocações, conta baú e movimentos do baú.

CREATE TABLE goals (
    id              VARCHAR(64) PRIMARY KEY,
    user_id         VARCHAR(64) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    target_amount   NUMERIC(18, 2) NOT NULL,
    duration_months INTEGER NOT NULL CHECK (duration_months BETWEEN 1 AND 60),
    start_date      DATE NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE goal_allocations (
    id              VARCHAR(64) PRIMARY KEY,
    goal_id         VARCHAR(64) NOT NULL REFERENCES goals(id),
    account_id      VARCHAR(64) NOT NULL,
    percentage      INTEGER NOT NULL CHECK (percentage BETWEEN 1 AND 100)
);

CREATE TABLE goal_vaults (
    id          VARCHAR(64) PRIMARY KEY,
    goal_id     VARCHAR(64) NOT NULL UNIQUE REFERENCES goals(id),
    user_id     VARCHAR(64) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE goal_vault_movements (
    id                VARCHAR(64) PRIMARY KEY,
    vault_id          VARCHAR(64) NOT NULL REFERENCES goal_vaults(id),
    source_account_id VARCHAR(64) NOT NULL,
    reference_month   DATE NOT NULL,
    movement_date     DATE NOT NULL,
    amount            NUMERIC(18, 2) NOT NULL,
    -- Status do movimento: COMPLETED (reservou o alvo), PARTIAL (saldo limitou),
    -- SKIPPED_NO_GROWTH (conta não evoluiu no mês), FAILED_INSUFFICIENT_BALANCE
    -- (sem saldo disponível). 'FAILED_INSUFFICIENT_BALANCE' tem 27 chars; VARCHAR(40) dá folga.
    status            VARCHAR(40) NOT NULL CHECK (status IN ('COMPLETED', 'PARTIAL', 'SKIPPED_NO_GROWTH', 'FAILED_INSUFFICIENT_BALANCE')),
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_goal_vault_movements_vault_id  ON goal_vault_movements(vault_id);
CREATE INDEX idx_goal_vault_movements_ref_month ON goal_vault_movements(reference_month);
