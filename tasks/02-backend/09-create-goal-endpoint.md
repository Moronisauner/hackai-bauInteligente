# 02-backend/09 — Endpoint POST /users/:userID/goals (cria goal + vault + allocations)

## Objetivo
Criar atomicamente: 1 `goal`, 1 `goal_vault`, N `goal_allocations` (RF-03, RF-04).

## Pré-requisitos
- 02-backend/04

## Passos
1. `internal/goal/service.go` com:
   ```go
   type CreateInput struct {
       UserID         string
       Name           string
       TargetAmount   decimal.Decimal
       DurationMonths int
       StartDate      time.Time
       Allocations    []struct{ AccountID string; Percentage int }
   }
   func (s *Service) Create(ctx context.Context, in CreateInput) (goalID, vaultID string, err error)
   ```
2. Validações (retornar erro tipado, mapear pra HTTP 400):
   - `TargetAmount > 0`
   - `DurationMonths BETWEEN 1 AND 60`
   - `len(Allocations) >= 1`
   - cada `Percentage BETWEEN 1 AND 100`
   - **soma dos percentuais == 100** (RF-04)
   - todos os `AccountID` pertencem ao `UserID` (verificar via `SELECT id FROM bank_accounts WHERE user_id=$1 AND id = ANY($2)`)
3. Persistir tudo em **1 transação** (`pool.BeginTx`):
   - INSERT em `goals` com id = `uuid.NewString()`
   - INSERT em `goal_vaults` com id = `uuid.NewString()`, `goal_id` = id acima
   - INSERTs em `goal_allocations`
4. Rota `POST /users/{userID}/goals`. Body JSON. Resposta `201` com `{"goal_id": "...", "vault_id": "..."}`.
5. `go get github.com/google/uuid`.

## Critério de aceite
- [ ] POST válido → 201, IDs no body, registros nas 3 tabelas.
- [ ] POST com soma != 100 → 400 `{"error":"allocations must sum to 100"}`.
- [ ] POST com `account_id` de outro usuário → 400.
- [ ] POST com `duration_months = 0` → 400.
- [ ] Falha em qualquer INSERT → nenhum registro permanece (transação).

## Referências PRD
- RF-03, RF-04, §7.2
