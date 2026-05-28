# 02-backend/06 — Testes do cálculo de saldo

## Objetivo
Cobrir o `Reconstruct` com testes que pegam regressões nas regras de §6 (sinal de débito/crédito, filtro de status, filtro de data).

## Pré-requisitos
- 02-backend/05

## Passos
1. Em `internal/balance/repo_test.go`, escrever testes de integração contra o Postgres do compose.
2. Estratégia: criar um `account_id` sintético no `SetUp` do teste e inserir 5–6 `transaction_events` cobrindo:
   - 1 crédito antes de `refDate` → entra no saldo.
   - 1 débito antes de `refDate` → subtrai.
   - 1 crédito **após** `refDate` → **não** entra.
   - 1 evento com `completed_authorised_payment_type != 'TRANSACAO_EFETIVADA'` → **não** entra.
   - 1 evento com `transaction_date_time` igual a `refDate` exatamente → **entra** (limite inclusivo).
3. `TearDown`: deletar os eventos sintéticos pelo `account_id` sintético (não tocar massa real).
4. Pular o teste se `DATABASE_URL` não estiver setada (`t.Skip`).

## Critério de aceite
- [ ] `cd backend && DATABASE_URL=... go test ./internal/balance/... -v` passa.
- [ ] Cada caso descrito acima tem assert explícito (5 sub-tests via `t.Run`).
- [ ] Massa real (`SELECT COUNT(*) FROM transaction_events`) inalterada após `go test`.

## Referências PRD
- §6
