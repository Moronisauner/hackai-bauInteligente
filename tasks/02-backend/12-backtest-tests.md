# 02-backend/12 — Testes unitários do backtest

## Objetivo
Cobrir o `backtest.Run` com cenários sintéticos que exercitam falha, sucesso, e saldo decrescente.

## Pré-requisitos
- 02-backend/11

## Passos
1. Em `internal/backtest/engine_test.go`, escrever testes com `BalanceFn` mockada (closure sobre `map[string]decimal.Decimal`).
2. Cenários (cada um um `t.Run`):
   - **happy path:** 1 conta, 100%, R$1000/mês por 3 meses, saldo sempre R$5000 → 3 `COMPLETED`.
   - **falha total:** saldo R$10 sempre → 3 `FAILED_INSUFFICIENT_BALANCE` com `Amount = 0`.
   - **falha intermitente:** 2 contas, 50% cada, R$500 cada/mês por 4 meses; conta A tem saldo só nos meses pares → A tem 2 COMPLETED + 2 FAILED, B tem 4 COMPLETED.
   - **saldo decrescente:** 1 conta, R$1000/mês, saldo inicial R$2500 — espera 2 COMPLETED e 1 FAILED no terceiro mês (saldo - sacado = 500 < 1000).
   - **withdrawal_day fora do mês corrente:** start `2024-01-31`, withdrawal_day `31`, duration `3` — confirmar que cada `MovementDate` cai no dia 31 ou no último dia válido (documentar comportamento de `time.AddDate`).

## Critério de aceite
- [ ] `go test ./internal/backtest/... -v` passa em todos os sub-tests.
- [ ] Cenário "falha total" assegura que cada `Movement.Amount.Equal(decimal.Zero)`.
- [ ] Cenário "saldo decrescente" prova que o engine considera saques anteriores.

## Referências PRD
- RF-05, §10 (coerência interna do backtest é critério de sucesso)
