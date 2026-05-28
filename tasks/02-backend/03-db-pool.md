# 02-backend/03 — Pool pgx e ping

## Objetivo
Estabelecer um `pgxpool.Pool` na inicialização da app e validar conectividade.

## Pré-requisitos
- 02-backend/02
- 01-infra/02 (Postgres rodando)

## Passos
1. Adicionar dependência: `cd backend && go get github.com/jackc/pgx/v5/pgxpool`.
2. Em `internal/db/db.go`, expor:
   ```go
   func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error)
   ```
   - Aplica `pool.Ping(ctx)` antes de retornar; erro de ping é erro de retorno.
   - Configurar `MaxConns` modesto (ex: 10) para POC.
3. Em `cmd/api/main.go`: criar pool após `config.Load()`, com `context.Background()` e timeout de 5s no ping. `defer pool.Close()`.
4. Em caso de erro de conexão, log e `os.Exit(1)`.

## Critério de aceite
- [ ] Com Postgres rodando: `go run ./cmd/api` sobe sem erro e fica rodando (não sai).
- [ ] Com Postgres parado: `go run ./cmd/api` falha em < 6s com mensagem "ping failed".

## Referências PRD
- §9
