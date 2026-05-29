-- Limpeza da massa: mantém APENAS os clientes e contas listados em infra/keep.conf
-- e apaga todo o resto. Execução MANUAL apenas (NÃO fica em infra/initdb/, então
-- não roda no init).
--
-- Uso: mise run db-cleanup   (chama infra/db-cleanup.sh, que lê infra/keep.conf)
--
-- Este script ASSUME que as temp tables já existem e estão populadas:
--   kept_users(user_id)  -> clientes a manter
--   kept_accounts(id)    -> contas (bank_accounts.id) a manter
-- O db-cleanup.sh as cria via \copy a partir do keep.conf antes de rodar este SQL.
--
-- Roda em transação única com ON_ERROR_STOP=1 (vide task): ou apaga tudo, ou nada.

BEGIN;

-- 1) Tabelas da POC, na ordem das FKs (filhos antes dos pais).
--    goal_vault_movements -> goal_vaults -> goals ; goal_allocations -> goals.
DELETE FROM goal_vault_movements
 WHERE vault_id IN (SELECT id FROM goal_vaults WHERE user_id NOT IN (SELECT user_id FROM kept_users));

DELETE FROM goal_allocations
 WHERE goal_id IN (SELECT id FROM goals WHERE user_id NOT IN (SELECT user_id FROM kept_users));

DELETE FROM goal_vaults WHERE user_id NOT IN (SELECT user_id FROM kept_users);
DELETE FROM goals        WHERE user_id NOT IN (SELECT user_id FROM kept_users);

-- 2) Massa Open Finance / cadastro (sem FKs entre si).
--    transaction_events: mantém só as transações das contas mantidas (por
--    account_id). Roda antes de apagar bank_accounts.
DELETE FROM transaction_events
 WHERE account_id NOT IN (SELECT id FROM kept_accounts);

DELETE FROM accounts_balances_events WHERE user_id NOT IN (SELECT user_id FROM kept_users);
DELETE FROM balances_history         WHERE user_id NOT IN (SELECT user_id FROM kept_users);
-- bank_accounts: mantém só as contas listadas (remove demais contas dos clientes
-- mantidos e contas de outros clientes de uma vez).
DELETE FROM bank_accounts            WHERE id NOT IN (SELECT id FROM kept_accounts);
DELETE FROM clientes                 WHERE user_id NOT IN (SELECT user_id FROM kept_users);

COMMIT;

\echo '>>> Limpeza concluída. Contagem restante por tabela:'
SELECT 'clientes'                AS tabela, count(*) FROM clientes
UNION ALL SELECT 'bank_accounts',            count(*) FROM bank_accounts
UNION ALL SELECT 'transaction_events',       count(*) FROM transaction_events
UNION ALL SELECT 'balances_history',         count(*) FROM balances_history
UNION ALL SELECT 'accounts_balances_events', count(*) FROM accounts_balances_events
UNION ALL SELECT 'goals',                    count(*) FROM goals
UNION ALL SELECT 'goal_allocations',         count(*) FROM goal_allocations
UNION ALL SELECT 'goal_vaults',              count(*) FROM goal_vaults
UNION ALL SELECT 'goal_vault_movements',     count(*) FROM goal_vault_movements
ORDER BY tabela;
