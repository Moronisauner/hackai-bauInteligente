# 01-infra/01 — Estrutura de pastas

## Objetivo
Criar o esqueleto de diretórios do monorepo (backend Go + frontend React + infra Docker).

## Pré-requisitos
Nenhum.

## Passos
1. Criar diretórios:
   - `backend/` — código Go
   - `frontend/` — código React
   - `infra/` — docker-compose e scripts de inicialização
   - `infra/initdb/` — SQLs que o Postgres roda no primeiro boot
2. Atualizar `.gitignore` na raiz com entradas para:
   - `backend/bin/`, `backend/tmp/`
   - `frontend/node_modules/`, `frontend/dist/`
   - `.env`, `*.local`
   - `infra/pgdata/` (volume Postgres)

## Critério de aceite
- [ ] `ls backend frontend infra infra/initdb` lista os 4 diretórios sem erro.
- [ ] `grep -E '(node_modules|pgdata|\.env)' .gitignore` retorna match em cada padrão.

## Referências PRD
- §9 (stack a definir — já fechada: Go + React + Postgres)
