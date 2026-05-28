# 02-backend/08 — Endpoint GET /users/:userID/accounts

## Objetivo
Listar contas de um cliente com saldo na `POC_REFERENCE_DATE` (RF-02).

## Pré-requisitos
- 02-backend/05
- 02-backend/07

## Passos
1. Handler `ListAccountsByUser`. Rota `GET /users/{userID}/accounts`.
2. SQL para contas:
   ```sql
   SELECT id, brand_name, type, branch_code, number, check_digit, compe_code
   FROM bank_accounts
   WHERE user_id = $1 AND status = 'AVAILABLE'
   ORDER BY brand_name, number;
   ```
3. Para o saldo: usar `balance.Repo.ReconstructMany(ctx, ids, cfg.POCReferenceDate)`.
4. Resposta JSON:
   ```json
   [
     {
       "account_id": "...",
       "brand_name": "Mercado Pago",
       "type": "CONTA_PAGAMENTO_PRE_PAGA",
       "branch_code": "0001",
       "number": "1234567",
       "check_digit": "8",
       "compe_code": "323",
       "balance": "1234.56",
       "balance_reference_date": "2024-06-01"
     }
   ]
   ```
5. Se o cliente não tem contas: retornar `200 []` (não 404).

## Critério de aceite
- [ ] Pegar um `user_id` de `GET /users` e chamar `GET /users/<id>/accounts` retorna lista.
- [ ] Cada item tem `balance` como string decimal (ex: `"1234.56"`).
- [ ] `balance_reference_date` = `POC_REFERENCE_DATE` configurado.
- [ ] User inexistente retorna `200 []`.

## Referências PRD
- RF-02, §8
