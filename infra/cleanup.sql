-- Limpeza da massa: mantém APENAS um cliente (user_id) e, dentro dele, apenas um
-- conjunto fixo de contas (agência/número-dígito). Apaga todo o resto.
--
-- Execução MANUAL apenas (NÃO fica em infra/initdb/, então não roda no init).
-- Uso:        mise run db-cleanup
-- Override:   KEEP_USER_ID=<hash> mise run db-cleanup
-- A task passa o user_id via -v keep_user_id; rodando psql direto, defina antes:
--   \set keep_user_id '695a8cddb2c3fcf4d0536663f8cfca8f4d147f14d1af5e4835c2aeb82b80ce84'
--
-- Roda em transação única com ON_ERROR_STOP=1 (vide task): ou apaga tudo, ou nada.

\echo '>>> Cliente mantido:' :keep_user_id

BEGIN;

-- Contas do cliente que devem ser mantidas (agência / número-dígito). Restringe
-- ao keep_user_id por segurança, embora os pares já sejam únicos na massa.
CREATE TEMP TABLE kept_accounts ON COMMIT DROP AS
SELECT id FROM bank_accounts
 WHERE user_id = :'keep_user_id'
   AND (branch_code, number, check_digit) IN (
       ('0390', '17358201', '8'),
       ('1000', '25066767', '6'),
       ('2663', '25673713', '6'),
       ('9227', '24530924', '8'),
       ('8765', '89990558', '7')
   );

-- 1) Tabelas da POC, na ordem das FKs (filhos antes dos pais).
--    goal_vault_movements -> goal_vaults -> goals ; goal_allocations -> goals.
DELETE FROM goal_vault_movements
 WHERE vault_id IN (SELECT id FROM goal_vaults WHERE user_id <> :'keep_user_id');

DELETE FROM goal_allocations
 WHERE goal_id IN (SELECT id FROM goals WHERE user_id <> :'keep_user_id');

DELETE FROM goal_vaults WHERE user_id <> :'keep_user_id';
DELETE FROM goals        WHERE user_id <> :'keep_user_id';

-- 2) Massa Open Finance / cadastro (sem FKs entre si).
--    transaction_events: mantém só as transações das contas mantidas (por
--    account_id). Roda antes de apagar bank_accounts.
DELETE FROM transaction_events
 WHERE account_id NOT IN (SELECT id FROM kept_accounts);

DELETE FROM accounts_balances_events WHERE user_id <> :'keep_user_id';
DELETE FROM balances_history         WHERE user_id <> :'keep_user_id';
-- bank_accounts: mantém só as contas listadas (remove demais contas do cliente
-- e contas de outros clientes de uma vez).
DELETE FROM bank_accounts            WHERE id NOT IN (SELECT id FROM kept_accounts);
DELETE FROM clientes                 WHERE user_id <> :'keep_user_id';

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
