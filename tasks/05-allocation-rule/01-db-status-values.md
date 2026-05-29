# 05-allocation-rule/01 — Novos status de movimento no schema

## Objetivo
Restringir os status da tabela de movimentos do baú a `COMPLETED` e `SKIPPED_NO_GROWTH`.

## Pré-requisitos
- nenhuma (infra)

## Passos
1. Em `infra/initdb/03-poc-tables.sql`, na coluna `goal_vault_movements.status`:
   - Atualizar o `CHECK (...)` para aceitar os dois valores:
     `'COMPLETED'`, `'SKIPPED_NO_GROWTH'`.
   - `VARCHAR(40)` já cobre o maior valor (`SKIPPED_NO_GROWTH` = 17 chars).
   - Atualizar o comentário acima da coluna para listar os dois status.
2. Como os scripts de `initdb` só rodam em volume novo, recriar o banco:
   ```sh
   mise run db-reset
   ```

## Critério de aceite
- [ ] `mise run db-reset` recria o banco sem erro.
- [ ] `INSERT` em `goal_vault_movements` com `status = 'COMPLETED'` e com
      `status = 'SKIPPED_NO_GROWTH'` **não** viola o CHECK.
- [ ] Um status inválido (ex: `'FOO'`) ainda é rejeitado.

## Referências PRD
- §7.2 (status e nota), RF-05
