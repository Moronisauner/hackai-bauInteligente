# 02-backend/10 — Endpoints de leitura de goals

## Objetivo
Permitir consultar objetivos criados e seus detalhes (alocações + vault).

## Pré-requisitos
- 02-backend/09

## Passos
1. Rotas:
   - `GET /users/{userID}/goals` → lista resumida.
   - `GET /goals/{goalID}` → detalhe com `allocations` (com `brand_name` + `number` da conta) e `vault_id`.
2. Para o detalhe, juntar com `bank_accounts` para enriquecer cada allocation com dados da conta-fonte:
   ```sql
   SELECT ga.account_id, ga.percentage, ba.brand_name, ba.number, ba.type
   FROM goal_allocations ga
   JOIN bank_accounts ba ON ba.id = ga.account_id
   WHERE ga.goal_id = $1;
   ```
3. `404` se `goal_id` não existe.
4. Calcular `monthly_amount` por alocação na resposta: `round(target_amount / duration_months * percentage / 100, 2)` (RF-04).

## Critério de aceite
- [ ] Criar um goal pelo endpoint POST e ver ele em `GET /users/<id>/goals`.
- [ ] `GET /goals/<id>` retorna `allocations` com `monthly_amount` calculado.
- [ ] Soma de `monthly_amount` ≈ `target_amount / duration_months` (tolerância de centavos por arredondamento).
- [ ] `GET /goals/inexistente` → 404.

## Referências PRD
- RF-04 (cálculo do valor mensal)
