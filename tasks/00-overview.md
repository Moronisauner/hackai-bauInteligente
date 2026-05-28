# Tasks — Centralizador de Saldo (POC)

Quebra do [PRD.md](../PRD.md) em tarefas atômicas e verificáveis.

## Stack escolhida

- **Backend:** Go + Chi (router) + pgx (Postgres)
- **Frontend:** React + Vite (TypeScript), SPA separada
- **Infra:** docker-compose com Postgres 16; seed via `docker-entrypoint-initdb.d`

## Convenções de cada task

Cada arquivo `.md` segue o formato:

- **Objetivo** — uma frase
- **Pré-requisitos** — outras tasks que precisam estar prontas
- **Passos** — instruções diretas, sem prosa
- **Critério de aceite** — comandos/checks que provam que terminou
- **Referências PRD** — seções do PRD que motivam a task

Tasks são pra ser rodadas **em ordem** dentro de cada fase. Fases podem se sobrepor depois que a infra (`01-infra`) está pronta.

## Fases

1. [01-infra/](01-infra/) — Postgres + schema + seed + migrations da POC
2. [02-backend/](02-backend/) — API Go (config → DB → endpoints → backtest)
3. [03-frontend/](03-frontend/) — UI React (wizard de objetivo + resultados)
4. [04-polish/](04-polish/) — Boot banner, dev runner, smoke E2E

## Documentação da API (Swagger)

Com o backend de pé, a documentação navegável fica em **http://localhost:8080/swagger/**
(spec em `/swagger/doc.json`). Para regerar os arquivos `backend/docs/swagger.{json,yaml}`
a partir das annotations dos handlers: `mise run swagger`.

## Como executar uma task com Claude

Aponte Claude pra uma task específica:

> "Execute a task `tasks/02-backend/05-balance-repo.md`"

Cada task é projetada pra caber numa sessão curta. Se uma task crescer demais durante execução, pare e quebre antes de continuar.
