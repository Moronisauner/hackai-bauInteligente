# 04-polish/02 — mise tasks pros comandos do dia-a-dia

## Objetivo
Centralizar comandos comuns (subir DB, rodar API, rodar frontend, resetar massa) em tasks do mise.

## Pré-requisitos
- 01-infra/02
- 02-backend/04
- 03-frontend/01

## Passos
1. Adicionar tasks no `mise.toml` da raiz:
   ```toml
   [tools]
   go = "latest"
   node = "20"

   [tasks.db-up]
   description = "Sobe Postgres via docker-compose"
   run = "cd infra && docker compose up -d"

   [tasks.db-down]
   description = "Para Postgres (mantém dados)"
   run = "cd infra && docker compose down"

   [tasks.db-reset]
   description = "Apaga volume e recarrega massa do zero"
   run = "cd infra && docker compose down -v && docker compose up -d"

   [tasks.api]
   description = "Roda o backend Go (precisa do .env carregado)"
   run = "cd backend && go run ./cmd/api"

   [tasks.api-test]
   run = "cd backend && go test ./..."

   [tasks.web]
   description = "Roda o frontend (Vite dev server)"
   run = "cd frontend && npm run dev"
   ```
2. Documentar no `00-overview.md` (ou novo `README.md`) a sequência: `mise run db-up && mise run api` (terminal 1) + `mise run web` (terminal 2).

## Critério de aceite
- [ ] `mise run db-up` sobe Postgres.
- [ ] `mise run api` sobe a API.
- [ ] `mise run web` sobe o frontend.
- [ ] `mise tasks` lista todas as tasks acima com descrição.

## Referências PRD
- §9
