# 02-backend/04 — Router Chi e endpoint /healthz

## Objetivo
Montar o servidor HTTP com Chi, middlewares básicos, e um `/healthz` que valida o DB.

## Pré-requisitos
- 02-backend/03

## Passos
1. `cd backend && go get github.com/go-chi/chi/v5`.
2. Em `internal/httpapi/router.go`, expor:
   ```go
   func New(pool *pgxpool.Pool, cfg config.Config) http.Handler
   ```
   - Usar `chi.NewRouter()` com middlewares: `middleware.RequestID`, `middleware.Logger`, `middleware.Recoverer`.
   - Registrar `GET /healthz` que faz `pool.Ping` (timeout 1s) e devolve `{"status":"ok"}` ou 503.
   - Aceitar JSON via `Content-Type: application/json` em handlers POST (próximas tasks).
3. Em `cmd/api/main.go`: usar `http.Server` com `Addr: cfg.HTTPPort`, `Handler: router`, e shutdown gracioso via `signal.NotifyContext` (SIGINT/SIGTERM).
4. Logar URL final do servidor no boot.

## Critério de aceite
- [ ] `curl -s http://localhost:8080/healthz` retorna `{"status":"ok"}` com HTTP 200.
- [ ] Parar Postgres → `curl /healthz` retorna 503.
- [ ] `Ctrl+C` para o processo de forma limpa (log "shutting down").

## Referências PRD
- §9
