# 02-backend/14 — Documentação Swagger/OpenAPI dos endpoints

## Objetivo
Gerar um arquivo `swagger.yaml` (+ `swagger.json`) documentando todos os endpoints da API, e expor uma UI navegável em `/swagger/`.

## Pré-requisitos
- 02-backend/07
- 02-backend/08
- 02-backend/09
- 02-backend/10
- 02-backend/13
- (idealmente também o `GET /config` mencionado em 03-frontend/03)

## Passos

1. Instalar tooling do `swaggo/swag`:
   ```
   go install github.com/swaggo/swag/cmd/swag@latest
   cd backend && go get github.com/swaggo/http-swagger/v2 github.com/swaggo/files
   ```
   Confirmar que `$(go env GOPATH)/bin` está no PATH (ou usar `mise` se preferir).

2. Anotar `cmd/api/main.go` com o bloco geral da API:
   ```go
   // @title       Centralizador de Saldo (POC) — API
   // @version     0.1
   // @description Backend da POC do centralizador de saldo. Toda regra temporal usa POC_REFERENCE_DATE.
   // @BasePath    /
   ```

3. Anotar cada handler existente com tags `@Summary`, `@Tags`, `@Param`, `@Success`, `@Failure`, `@Router`. Cobertura mínima obrigatória:
   - `GET /healthz`
   - `GET /config` (se já existir)
   - `GET /users` (+ query `q`)
   - `GET /users/{userID}/accounts`
   - `POST /users/{userID}/goals`
   - `GET /users/{userID}/goals`
   - `GET /goals/{goalID}`
   - `POST /goals/{goalID}/backtest`
   - `GET /goals/{goalID}/backtest`

   Para schemas dos bodies, declarar structs Go dedicadas (ex: `CreateGoalRequest`, `UserDTO`, `AccountDTO`, `GoalDTO`, `BacktestResultDTO`) no pacote `httpapi` e referenciá-las com `@Param body body httpapi.CreateGoalRequest true "..."` e `@Success 200 {object} httpapi.BacktestResultDTO`.

   Reutilizar as structs já usadas nos handlers — não duplicar tipos.

4. Rodar o gerador a partir da raiz do backend:
   ```
   cd backend && swag init --generalInfo cmd/api/main.go --output docs --parseDependency --parseInternal
   ```
   Isso produz:
   - `backend/docs/docs.go`
   - `backend/docs/swagger.json`
   - `backend/docs/swagger.yaml`

5. Montar UI no router (`internal/httpapi/router.go`):
   ```go
   import (
       httpSwagger "github.com/swaggo/http-swagger/v2"
       _ "github.com/moronisauner/hackai/backend/docs" // ajustar ao module path real
   )
   // ...
   r.Get("/swagger/*", httpSwagger.Handler(
       httpSwagger.URL("/swagger/doc.json"),
   ))
   ```

6. Adicionar task no `mise.toml` pra regerar facilmente:
   ```toml
   [tasks.swagger]
   description = "Regera backend/docs/swagger.{json,yaml} a partir das annotations"
   run = "cd backend && swag init --generalInfo cmd/api/main.go --output docs --parseDependency --parseInternal"
   ```

7. Adicionar nota curta no `00-overview.md` (ou README) apontando pra `/swagger/` quando a API estiver de pé.

## Critério de aceite

- [ ] `cd backend && swag init ...` roda sem warning crítico (warnings de "could not parse" em pacotes terceiros toleráveis).
- [ ] `backend/docs/swagger.yaml` e `backend/docs/swagger.json` existem e contêm **todos** os endpoints listados na seção Passos item 3.
- [ ] Abrir `http://localhost:8080/swagger/` no browser mostra a UI com a lista completa de endpoints.
- [ ] Clicar em "Try it out" no `GET /users` na UI executa de verdade e retorna a lista.
- [ ] Schemas dos bodies de request/response aparecem expandíveis na UI (não apenas `object` opaco).
- [ ] `mise run swagger` regera os arquivos.

## Referências PRD
- §9 (boa prática pra POC — facilita o time de produto/dados validar o critério §10 sem ler código)
