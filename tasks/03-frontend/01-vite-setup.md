# 03-frontend/01 — Setup Vite + React + TS

## Objetivo
Inicializar SPA React em TypeScript com Vite, configurar proxy `/api` → backend, e router básico.

## Pré-requisitos
- 02-backend/04 (precisa do backend rodando em :8080)

## Passos
1. `cd frontend && npm create vite@latest . -- --template react-ts` (responder pra usar a pasta atual).
2. `npm install`
3. Adicionar libs:
   ```
   npm install react-router-dom
   npm install -D tailwindcss postcss autoprefixer
   npx tailwindcss init -p
   ```
4. Configurar Tailwind (`content: ['./index.html', './src/**/*.{ts,tsx}']`, importar `@tailwind base/components/utilities` em `src/index.css`).
5. Em `vite.config.ts`, adicionar proxy:
   ```ts
   server: {
     proxy: {
       '/api': { target: 'http://localhost:8080', rewrite: p => p.replace(/^\/api/, '') }
     }
   }
   ```
6. Criar `src/routes.tsx` com `createBrowserRouter` e rotas placeholder:
   - `/` → `<SelectUserPage />` (placeholder com "TODO")
   - `/users/:userID` → `<AccountsPage />` (placeholder)
   - `/users/:userID/goals/new` → `<GoalFormPage />` (placeholder)
   - `/goals/:goalID` → `<BacktestResultsPage />` (placeholder)
7. Apagar boilerplate do Vite (logo, contador, CSS demo).

## Critério de aceite
- [ ] `cd frontend && npm run dev` sobe em `http://localhost:5173`.
- [ ] Acessar `/` mostra o placeholder do `SelectUserPage`.
- [ ] No browser console: `fetch('/api/healthz').then(r=>r.json())` retorna `{"status":"ok"}` (proxy funcionando).
- [ ] Classe Tailwind (ex: `text-red-500`) aplica estilo numa página de teste.

## Referências PRD
- §9
