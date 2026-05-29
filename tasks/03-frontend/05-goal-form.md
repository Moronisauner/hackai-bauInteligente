# 03-frontend/05 — Formulário de criação de objetivo (wizard step 1)

## Objetivo
Tela `/users/:userID/goals/new` com formulário dos dados do objetivo (jornada item 3).

## Pré-requisitos
- 03-frontend/04

## Passos
1. Em `src/pages/GoalFormPage.tsx`, montar formulário controlado com campos:
   - `name` (texto, obrigatório, max 255)
   - `target_amount` (input texto com máscara BRL; converter via `parseBRLInput` antes de enviar)
   - `duration_months` (number, 1–60)
   - `start_date` (date input; default = `POC_REFERENCE_DATE`)
2. Validações client-side com mensagens inline:
   - `name` não vazio
   - `target_amount > 0`
   - `duration_months` 1–60
3. **Não** submete ainda ao backend — esta tela é o **step 1 do wizard**. Ao clicar "Próximo", guardar o estado parcial em `useState` no componente pai (ou em `useNavigate(state)`) e ir pra `/users/:userID/goals/new/allocations`.
4. Botão "Voltar" volta pra `/users/:userID`.

## Critério de aceite
- [ ] Campos validam visualmente (borda vermelha + mensagem).
- [ ] Botão "Próximo" desabilitado enquanto inválido.
- [ ] Ao clicar "Próximo" com tudo válido, navega pra step de alocações carregando os dados.

## Referências PRD
- Jornada §4 item 3, RF-03
