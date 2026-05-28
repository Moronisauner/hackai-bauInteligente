# 04-polish/03 — Smoke E2E com 5 clientes

## Objetivo
Cumprir o critério §10 do PRD: rodar o fluxo completo (seleção → objetivo → simulação) com pelo menos 5 clientes diferentes da massa.

## Pré-requisitos
- 03-frontend/07

## Passos
1. Selecionar 5 `user_id` candidatos da massa, priorizando:
   - histórico longo (`MIN(transaction_date_time)` bem antes da `POC_REFERENCE_DATE`)
   - múltiplas contas
   - movimento expressivo (`SUM(amount) > algum threshold`)

   Query sugerida:
   ```sql
   SELECT ba.user_id, COUNT(DISTINCT ba.id) AS contas,
          COUNT(te.*) AS transacoes,
          MIN(te.transaction_date_time) AS primeira_tx,
          MAX(te.transaction_date_time) AS ultima_tx
   FROM bank_accounts ba
   LEFT JOIN transaction_events te ON te.account_id = ba.id
   GROUP BY ba.user_id
   HAVING COUNT(DISTINCT ba.id) >= 2
      AND MIN(te.transaction_date_time) < '<POC_REFERENCE_DATE menos 6 meses>'
   ORDER BY transacoes DESC
   LIMIT 10;
   ```
2. Criar `docs/SMOKE.md` documentando, pra cada um dos 5 clientes escolhidos:
   - `user_id`
   - número de contas
   - intervalo de histórico
   - objetivo testado (nome, target, prazo, % por conta)
   - resultado: meses cumpridos, saldo final do baú, meta atingida?
   - observação curta (algo surpreendeu?)
3. Anexar screenshots da tela de resultados (`/goals/<id>`) pra cada um.
4. Concluir com **veredito go/no-go** (RF do §10 item 3).

## Critério de aceite
- [ ] `docs/SMOKE.md` existe com 5 clientes documentados.
- [ ] Cada cliente tem screenshot do backtest.
- [ ] Para cada um, `summary.vault_balance` reportado bate com `SUM(amount) WHERE status='COMPLETED'` no DB (validação cruzada do §10 item 2).
- [ ] Documento termina com seção "Recomendação" (go ou no-go + por quê).

## Referências PRD
- §10 (critérios de sucesso completos), §12 item 3
