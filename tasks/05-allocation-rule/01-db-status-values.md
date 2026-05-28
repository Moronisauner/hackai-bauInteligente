# 05-allocation-rule/01 — Novos status de movimento no schema

## Objetivo
Permitir os status `PARTIAL` e `SKIPPED_NO_GROWTH` na tabela de movimentos do baú.

## Pré-requisitos
- nenhuma (infra)

## Passos
1. Em `infra/initdb/03-poc-tables.sql`, na coluna `goal_vault_movements.status`:
   - Atualizar o `CHECK (...)` para aceitar os quatro valores:
     `'COMPLETED'`, `'PARTIAL'`, `'SKIPPED_NO_GROWTH'`, `'FAILED_INSUFFICIENT_BALANCE'`.
   - `VARCHAR(40)` já cobre o maior valor (`FAILED_INSUFFICIENT_BALANCE` = 27 chars).
   - Atualizar o comentário acima da coluna para listar os quatro status.
2. Como os scripts de `initdb` só rodam em volume novo, recriar o banco:
   ```sh
   mise run db-reset
   ```

## Critério de aceite
- [ ] `mise run db-reset` recria o banco sem erro.
- [ ] `INSERT` em `goal_vault_movements` com `status = 'PARTIAL'` e com
      `status = 'SKIPPED_NO_GROWTH'` **não** viola o CHECK.
- [ ] Um status inválido (ex: `'FOO'`) ainda é rejeitado.

## Referências PRD
- §7.2 (status e nota), RF-05
