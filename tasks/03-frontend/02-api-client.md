# 03-frontend/02 — Cliente HTTP e tipos compartilhados

## Objetivo
Centralizar chamadas ao backend em um cliente tipado, evitando `fetch` espalhado pelos componentes.

## Pré-requisitos
- 03-frontend/01

## Passos
1. Criar `src/api/types.ts` com tipos espelhando os endpoints já definidos:
   ```ts
   export type User = { user_id: string; accounts_count: number };
   export type Account = {
     account_id: string;
     brand_name: string;
     type: string;
     branch_code: string;
     number: string;
     check_digit: string;
     compe_code: string;
     balance: string; // string decimal — manter como string e formatar na UI
     balance_reference_date: string;
   };
   export type Allocation = {
     account_id: string;
     percentage: number;
     monthly_amount: string;
     brand_name?: string;
     number?: string;
   };
   export type Goal = {
     id: string;
     name: string;
     target_amount: string;
     duration_months: number;
     start_date: string;
     vault_id: string;
     allocations: Allocation[];
   };
   export type BacktestResult = {
     summary: {
       completed_months_pct: number;
       vault_balance: string;
       target_amount: string;
       goal_reached: boolean;
       worst_account_id: string | null;
     };
     movements: { reference_month: string; account_id: string; status: 'COMPLETED' | 'SKIPPED_NO_GROWTH'; amount: string }[];
     vault_evolution: { month: string; balance: string }[];
   };
   ```
2. Criar `src/api/client.ts` com funções:
   ```ts
   listUsers(q?: string): Promise<User[]>
   listAccounts(userID: string): Promise<Account[]>
   createGoal(userID: string, body: CreateGoalInput): Promise<{ goal_id: string; vault_id: string }>
   getGoal(goalID: string): Promise<Goal>
   runBacktest(goalID: string): Promise<BacktestResult>
   getBacktest(goalID: string): Promise<BacktestResult>
   ```
3. Implementar com `fetch` apontando pra `/api/...`. Em erro HTTP, `throw new Error(<status> <body>)`.
4. Criar `src/lib/format.ts` com helpers:
   - `formatBRL(value: string): string` — `"1234.56"` → `"R$ 1.234,56"`
   - `parseBRLInput(value: string): string` — input do usuário (vírgula) → string decimal pra API.

## Critério de aceite
- [ ] `tsc --noEmit` (via `npm run build`) passa sem erro de tipo.
- [ ] Em uma página de teste, chamar `await listUsers()` retorna array (verificar no DevTools).
- [ ] `formatBRL('1234.56')` retorna `'R$ 1.234,56'`.

## Referências PRD
- (corte transversal)
