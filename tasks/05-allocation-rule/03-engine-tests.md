# 05-allocation-rule/03 — Testes da engine para a nova regra

## Objetivo
Reescrever `engine_test.go` para cobrir reserva baseada em evolução, com uma
`BalanceFn` que varia o saldo por (conta, data).

## Pré-requisitos
- 05-allocation-rule/02

## Passos
1. Substituir o `constBalance` por um helper que devolve saldo dependente da data
   (ex: um `map[string]decimal` por marco temporal, ou função que cresce o saldo
   ao longo dos meses) — necessário para simular *evolução*.
2. Cobrir os casos:
   - **Evolução positiva, saldo sobra** → `COMPLETED`, `Amount = evolução × pct`.
   - **Sem evolução** (saldo fim == saldo abertura, ou caiu) → `SKIPPED_NO_GROWTH`, `Amount = 0`.
   - **Evolução positiva, mas saldo no dia do saque < alvo** → `PARTIAL`, `Amount = disponível`.
   - **Evolução positiva, mas saldo disponível <= 0** → `FAILED_INSUFFICIENT_BALANCE`.
   - **Acumulação entre meses**: reservas anteriores reduzem o disponível dos meses seguintes.
3. Manter o caso de `withdrawal_day` fora do mês (dia 31) se ainda fizer sentido.

## Critério de aceite
- [ ] `go test ./internal/backtest/...` passa.
- [ ] Há ao menos um teste por status (`COMPLETED`, `PARTIAL`, `SKIPPED_NO_GROWTH`, `FAILED_INSUFFICIENT_BALANCE`).

## Referências PRD
- RF-05
