# 04-polish/01 — Banner de boot com POC_REFERENCE_DATE em destaque

## Objetivo
Imprimir, no startup do backend, um banner óbvio com a `POC_REFERENCE_DATE` em uso (§8 do PRD pede destaque pra evitar confusão).

## Pré-requisitos
- 02-backend/04

## Passos
1. Em `cmd/api/main.go`, logo após `config.Load()`, antes de iniciar o servidor, imprimir:
   ```
   ╔════════════════════════════════════════╗
   ║  POC_REFERENCE_DATE = 2024-06-01       ║
   ║  Toda regra de negócio usa essa data,  ║
   ║  NÃO time.Now()                        ║
   ╚════════════════════════════════════════╝
   ```
2. Usar `fmt.Println` direto (não `log.Println`) pra ficar visível antes dos logs estruturados.
3. Largura do banner ajusta pra acomodar a data dinâmica.

## Critério de aceite
- [ ] `go run ./cmd/api` imprime o banner no topo.
- [ ] A data impressa é exatamente o valor de `POC_REFERENCE_DATE` do env.

## Referências PRD
- §8 (destaque pra evitar confusão com `NOW()`)
