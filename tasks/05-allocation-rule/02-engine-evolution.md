# 05-allocation-rule/02 — Engine: reserva como fatia da evolução mensal

## Objetivo
Reescrever a engine de backtest para reservar uma fatia da **evolução mensal** de
cada conta, em vez de um valor mensal fixo.

## Pré-requisitos
- 05-allocation-rule/01

## Passos
1. Em `internal/backtest/engine.go`:
   - `Allocation`: **remover** `MonthlyAmount`; manter `AccountID` e `Percentage`
     (o percentual agora é a fatia da evolução da conta).
   - Constantes de status: adicionar
     ```go
     StatusPartial = "PARTIAL"             // reservou menos que o alvo (saldo limitou)
     StatusSkipped = "SKIPPED_NO_GROWTH"   // conta não evoluiu no mês
     ```
     manter `StatusCompleted` e `StatusFailed`.
2. Em `Run`, para cada mês `i` e cada `Allocation`:
   - `monthOpen` = primeiro instante do mês de competência (`referenceMonth`, 00:00:00).
   - `monthClose` = último instante do mês = `referenceMonth.AddDate(0,1,0).Add(-time.Nanosecond)`.
   - `evolution = balance(close) - balance(open)`.
   - Se `evolution <= 0` → `StatusSkipped`, `Amount = 0`.
   - `target = round(evolution * Percentage / 100, 2)`; se `target <= 0` → `StatusSkipped`.
   - `available = balance(movementDate) - reserved[accountID]` (reservas anteriores
     do backtest descontam o disponível — manter o acumulador interno).
     - `available <= 0` → `StatusFailed`, `Amount = 0`.
     - `available >= target` → `StatusCompleted`, `Amount = target`, soma em `reserved`.
     - senão → `StatusPartial`, `Amount = round(available, 2)`, soma em `reserved`.
3. Atualizar o doc-comment do topo do arquivo descrevendo a nova regra
   (evolução → fatia → limite do disponível) e o acumulador `reserved`.

## Critério de aceite
- [ ] `go build ./internal/backtest/...` ok.
- [ ] `Run` continua **puro** (sem importar `pgx`/`database/sql`).
- [ ] Nenhuma referência a `MonthlyAmount` resta no pacote.

## Referências PRD
- RF-04, RF-05, §11 (janela da evolução)
