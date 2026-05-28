# 03-frontend/04 — Página de contas do cliente

## Objetivo
Tela `/users/:userID` que mostra as contas com saldo na data de referência (jornada item 2).

## Pré-requisitos
- 03-frontend/03
- 02-backend/08

## Passos
1. Em `src/pages/AccountsPage.tsx`:
   - Pegar `userID` via `useParams`.
   - Carregar `listAccounts(userID)` no mount.
   - Renderizar tabela com colunas: Instituição (`brand_name`), Tipo (`type` traduzido pra label amigável via mapa estático), Agência/Número (`branch_code` + `number` + `check_digit`), Saldo (formatado em BRL).
   - Linha de total: saldo somado das contas.
2. Botão "Criar objetivo" no topo → navega pra `/users/:userID/goals/new`.
3. Mostrar `balance_reference_date` em destaque ("Saldo em DD/MM/YYYY").
4. Mapa de tipos amigáveis:
   ```ts
   const ACCOUNT_TYPE_LABEL: Record<string, string> = {
     CONTA_DEPOSITO_A_VISTA: 'Conta corrente',
     CONTA_POUPANCA: 'Poupança',
     CONTA_PAGAMENTO_PRE_PAGA: 'Conta de pagamento',
   };
   ```
5. Se a lista vier vazia, mostrar empty state com link de volta.

## Critério de aceite
- [ ] Abrir `/users/<id>` mostra contas + saldos.
- [ ] Total na base da tabela = soma dos saldos.
- [ ] Botão "Criar objetivo" funciona.
- [ ] Cliente sem contas mostra empty state, sem quebrar.

## Referências PRD
- Jornada §4 item 2, RF-02, §6
