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
   - **Evolução positiva** → `COMPLETED`, `Amount = round(evolução × pct, 2)`.
   - **Sem evolução** (saldo fim == saldo abertura, ou caiu) → `SKIPPED_NO_GROWTH`, `Amount = 0`.
   - **Saldo de abertura baixo não limita a reserva**: mesmo com saldo de abertura 0, a fatia da evolução é reservada por inteiro (`COMPLETED`).
   - **Vários meses com evolução** → todos `COMPLETED`.
3. Validar que `MovementDate` é sempre o dia 1 do mês de competência (= `ReferenceMonth`).

## Critério de aceite
- [ ] `go test ./internal/backtest/...` passa.
- [ ] Há ao menos um teste por status (`COMPLETED`, `SKIPPED_NO_GROWTH`).

## Referências PRD
- RF-05
