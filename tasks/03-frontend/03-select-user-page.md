# 03-frontend/03 — Página de seleção de cliente

## Objetivo
Implementar a tela `/` que lista `user_id` com contagem de contas e permite buscar e selecionar (jornada item 1).

## Pré-requisitos
- 03-frontend/02
- 02-backend/07

## Passos
1. Em `src/pages/SelectUserPage.tsx`:
   - `useEffect` carrega `listUsers()` ao montar.
   - Input de busca (debounce 300ms) refaz `listUsers(q)`.
   - Renderiza lista (tabela ou cards) com `user_id` e `accounts_count`.
   - Click em um item navega pra `/users/:userID` via `useNavigate`.
2. Estados visuais: loading (skeleton), empty (sem resultados), error.
3. Mostrar a `POC_REFERENCE_DATE` em destaque no header. Pra isso, expor um endpoint `GET /config` no backend (uma rota mínima que retorna `{ "poc_reference_date": "..." }`) — adicionar ao backend se ainda não existe e atualizar `api/client.ts`.

## Critério de aceite
- [ ] Abrir `/` mostra lista de usuários.
- [ ] Digitar no campo de busca filtra (com debounce).
- [ ] Header mostra a `POC_REFERENCE_DATE`.
- [ ] Click em um usuário navega pra `/users/<id>` (mesmo que a próxima página ainda esteja em placeholder).

## Referências PRD
- Jornada §4 item 1, RF-01, §8 (destaque da `POC_REFERENCE_DATE`)
