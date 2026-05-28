# 03-frontend/06 — Wizard step 2: seleção de contas-fonte e alocação por %

## Objetivo
Tela que coleta contas-fonte e percentuais, valida soma = 100, e submete tudo via `createGoal` (jornada itens 4 e 5).

## Pré-requisitos
- 03-frontend/05
- 02-backend/09

## Passos
1. Em `src/pages/AllocationFormPage.tsx` (rota `/users/:userID/goals/new/allocations`):
   - Recuperar estado da etapa anterior (via `location.state` ou contexto).
   - Carregar `listAccounts(userID)` no mount.
   - Pra cada conta: checkbox + input numérico (% inteiro 1–100, desabilitado se checkbox off).
2. Mostrar **em tempo real** o valor mensal calculado por conta: `(target_amount / duration_months) * pct / 100`, formatado em BRL.
3. Rodapé fixo com **"Soma atual: X% / 100%"**. Cor verde se 100, vermelho se != 100.
4. Botão "Criar objetivo" desabilitado enquanto soma != 100 ou nenhuma conta selecionada.
5. Ao submeter, chamar `createGoal(userID, { ...dadosStep1, allocations: [{ account_id, percentage }] })`.
6. Em sucesso, navegar pra `/goals/<goal_id>`.
7. Em erro 400 do backend, mostrar mensagem no topo do form.

## Critério de aceite
- [ ] Marcar contas, distribuir 50/50, ver soma "100%" em verde.
- [ ] Tentar 60/30 mantém botão desabilitado e mostra "90% / 100%" em vermelho.
- [ ] Submit com sucesso navega pra tela de resultados.
- [ ] Forçar erro (ex: tirar uma conta pelo DevTools) e ver mensagem de erro do backend.

## Referências PRD
- Jornada §4 itens 4–5, RF-04
