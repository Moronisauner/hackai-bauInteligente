# 05-allocation-rule/04 — Validação: percentuais não somam mais 100%

## Objetivo
Remover a exigência de que as alocações somem exatamente 100% na criação do objetivo.

## Pré-requisitos
- nenhuma (independente da engine)

## Passos
1. Em `internal/goal/service.go`, função `validate`:
   - **Remover** o acumulador `sum` e o bloco `if sum != 100 { ... }`.
   - **Manter**: pelo menos 1 alocação; cada `percentage` entre 1 e 100; `account_id` obrigatório.

## Critério de aceite
- [ ] Criar objetivo com 1 conta a 20% retorna **201** (antes era 400).
- [ ] Criar objetivo com 2 contas 30% + 40% (soma 70%) retorna **201**.
- [ ] `percentage = 0` ou `101` ainda retorna **400**.
- [ ] `go test ./...` passa (ajustar testes que esperavam erro de soma, se houver).

## Referências PRD
- RF-04
