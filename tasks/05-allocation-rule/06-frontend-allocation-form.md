# 05-allocation-rule/06 — Frontend: form de alocação (fatia da evolução)

## Objetivo
Refletir na UI que o % é a fatia da evolução mensal da conta — sem exigir soma
100% e sem projetar valor mensal fixo.

## Pré-requisitos
- 05-allocation-rule/04

## Passos
1. `src/api/types.ts`:
   - `BacktestMovementStatus`: `'COMPLETED' | 'SKIPPED_NO_GROWTH'`.
   - `Allocation`: remover `monthly_amount`.
2. `src/pages/AllocationFormPage.tsx`:
   - Remover a regra `totalPct === 100`. `canSubmit` = ao menos 1 conta marcada,
     todas as marcadas com `percentage` válido (1–100), e `!submitting`.
   - Remover o cálculo `monthlyTotal` e a coluna **"Valor mensal"** (não há projeção).
   - Trocar o rótulo da coluna `%` / subtítulo para deixar claro: **"% da evolução
     mensal da conta"** (quanto do que a conta crescer no mês vai pra meta).
   - Rodapé: trocar "Soma atual: X% / 100%" por algo neutro (ex: nº de contas
     selecionadas) — sem semáforo de 100%.

## Critério de aceite
- [ ] Marcar 1 conta a 30% habilita "Criar objetivo" (soma != 100 não bloqueia).
- [ ] Não há mais coluna/projeção de valor mensal fixo.
- [ ] Texto da tela explica que o % incide sobre a evolução da conta.
- [ ] `npm run build` (ou `tsc`) sem erros de tipo.

## Referências PRD
- RF-04, jornada §4 item 5
