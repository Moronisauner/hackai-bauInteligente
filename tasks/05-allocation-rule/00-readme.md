# 05-allocation-rule — Refino da regra de alocação

Refino da semântica de **alocação** (RF-04/RF-05): o percentual deixa de ser a
fatia de um aporte mensal fixo (`target/prazo × %`) e passa a significar **quanto
da evolução mensal da própria conta** (entradas − saídas do mês) é reservado para
a meta.

Consequências que estas tasks implementam:

- Reserva mensal = `min(evolução_positiva × %, saldo disponível no dia do saque)`.
- Percentuais **independentes por conta** — não somam mais 100%.
- 4 status de movimento: `COMPLETED`, `PARTIAL`, `SKIPPED_NO_GROWTH`,
  `FAILED_INSUFFICIENT_BALANCE`.
- Some o "valor mensal" fixo projetado na criação (não é mais projetável).

Rodar **em ordem** (DB → engine → testes → validação → handlers → frontend).

## Referências PRD
- RF-04, RF-05, RF-06, §7.2, §11 (janela da evolução)
