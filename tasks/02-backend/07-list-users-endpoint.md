# 02-backend/07 — Endpoint GET /users

## Objetivo
Listar `user_id` distintos com contagem de contas vinculadas (RF-01).

## Pré-requisitos
- 02-backend/04

## Passos
1. Em `internal/httpapi/users_handler.go`, criar handler `ListUsers`.
2. SQL:
   ```sql
   SELECT user_id, COUNT(*) AS accounts_count
   FROM bank_accounts
   WHERE status = 'AVAILABLE'
   GROUP BY user_id
   ORDER BY accounts_count DESC, user_id ASC;
   ```
3. Resposta JSON:
   ```json
   [{"user_id": "...", "accounts_count": 3}, ...]
   ```
4. Registrar rota `GET /users` no router.
5. Suportar query param opcional `?q=<substring>` que filtra `user_id ILIKE '%' || $1 || '%'` (busca textual mencionada em RF-01).

## Critério de aceite
- [ ] `curl -s http://localhost:8080/users | jq 'length'` retorna > 0.
- [ ] `curl -s 'http://localhost:8080/users?q=<prefixo>' | jq` filtra a lista.
- [ ] Ordenação: primeiro user na resposta tem o maior `accounts_count`.

## Referências PRD
- RF-01
