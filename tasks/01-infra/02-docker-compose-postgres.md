# 01-infra/02 — docker-compose com Postgres

## Objetivo
Subir Postgres 16 local via docker-compose, com volume persistente e mapeamento da pasta `infra/initdb/`.

## Pré-requisitos
- 01-infra/01

## Passos
1. Criar `infra/docker-compose.yml` com um serviço `postgres`:
   - imagem `postgres:16-alpine`
   - env: `POSTGRES_USER=hackai`, `POSTGRES_PASSWORD=hackai`, `POSTGRES_DB=hackai`
   - volume nomeado `pgdata:/var/lib/postgresql/data`
   - mount read-only `./initdb:/docker-entrypoint-initdb.d:ro`
   - porta `5432:5432`
   - healthcheck com `pg_isready`
2. Criar `infra/.env.example` documentando a `DATABASE_URL`:
   ```
   DATABASE_URL=postgres://hackai:hackai@localhost:5432/hackai?sslmode=disable
   ```

## Critério de aceite
- [ ] `cd infra && docker compose up -d` sobe sem erro.
- [ ] `docker compose exec postgres pg_isready -U hackai` retorna `accepting connections`.
- [ ] `docker compose down` para o serviço (sem `-v`, mantém dados).

## Referências PRD
- §9
