# 02-backend/01 — Esqueleto do módulo Go

## Objetivo
Inicializar o módulo Go do backend com a estrutura de pastas e um `main.go` que compila.

## Pré-requisitos
- 01-infra/01

## Passos
1. `cd backend && go mod init github.com/moronisauner/hackai/backend` (ajustar o path se preferir outro).
2. Criar diretórios:
   - `cmd/api/` (entrypoint)
   - `internal/config/`
   - `internal/db/`
   - `internal/httpapi/` (handlers, router) — evita conflito com `net/http`
   - `internal/balance/`
   - `internal/goal/`
   - `internal/backtest/`
3. Criar `cmd/api/main.go` com `func main() { fmt.Println("hackai api") }`.
4. Rodar `go mod tidy`.

## Critério de aceite
- [ ] `cd backend && go build ./...` termina sem erro.
- [ ] `cd backend && go run ./cmd/api` imprime `hackai api`.
- [ ] `ls backend/internal` lista as 6 subpastas.

## Referências PRD
- §9 (stack)
