# 01-infra/04 — Tabelas da POC (goals, vaults, allocations, movements)

## Objetivo
Criar as 4 tabelas novas da POC, descritas em §7.2 do PRD, via arquivo SQL versionado.

## Pré-requisitos
- 01-infra/03

## Passos
1. Criar `infra/initdb/03-poc-tables.sql` contendo os `CREATE TABLE` exatos de §7.2 do PRD:
   - `goals`
   - `goal_allocations`
   - `goal_vaults` (com `UNIQUE` em `goal_id`)
   - `goal_vault_movements`
   - índices `idx_goal_vault_movements_vault_id` e `idx_goal_vault_movements_ref_month`
2. Adicionar `CHECK` extra em `goal_vault_movements.status` (whitelist: `'COMPLETED'`, `'FAILED_INSUFFICIENT_BALANCE'`).
3. Adicionar `CHECK` em `goals.duration_months BETWEEN 1 AND 60` (RF-03).
4. Recriar volume e subir: `cd infra && docker compose down -v && docker compose up -d`.

## Critério de aceite
- [ ] `docker compose exec postgres psql -U hackai -d hackai -c "\dt"` lista as 4 tabelas novas.
- [ ] Tentar inserir `duration_months = 100` em `goals` falha por CHECK.
- [ ] Tentar inserir 2 `goal_vaults` com o mesmo `goal_id` falha por UNIQUE.

## Referências PRD
- §7.2 (DDL completo)
- RF-03, RF-04
