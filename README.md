# Centralizador de Saldo — POC (Baú Inteligente)

POC que valida um **centralizador de saldo**: olha para as várias contas de um cliente
(dados de Open Finance) e permite estruturar um objetivo financeiro com reservas mensais
programadas a partir dessas contas. Toda a simulação roda sobre uma massa de dados
históricos — não há ingestão em tempo real nem execução real de saques.

Detalhes do produto em [PRD.md](PRD.md).

## Stack

| Camada    | Tecnologia                                  |
| --------- | ------------------------------------------- |
| Backend   | Go (`backend/`, entrypoint `cmd/api`)       |
| Frontend  | Vite + React + TypeScript (`frontend/`)     |
| Banco     | PostgreSQL 16 (Docker Compose, `infra/`)    |
| Tooling   | [mise](https://mise.jdx.dev/)               |

## Pré-requisitos

- [**mise**](https://mise.jdx.dev/getting-started.html) — gerencia as versões de Go, Node
  e ferramentas auxiliares, além de carregar as variáveis de ambiente.
- [**Docker**](https://docs.docker.com/get-docker/) com **Docker Compose** — sobe o
  Postgres da POC.

> Go, Node e o `swag` **não** precisam ser instalados manualmente: o `mise install`
> resolve tudo a partir do [mise.toml](mise.toml).

## Setup

### 1. Instalar as ferramentas com o mise

A partir da raiz do projeto:

```bash
mise install
```

Isso instala as versões de Go, Node e demais ferramentas declaradas no [mise.toml](mise.toml).

### 2. Configurar o `.env`

O projeto carrega segredos de um arquivo `.env` (não versionado). Crie-o na raiz com a
chave da LLM usada pelo assistente de planejamento:

```bash
echo 'LLM_API_KEY=sua-chave-aqui' > .env
```

As demais variáveis (`DATABASE_URL`, `LLM_BASE_URL`, `LLM_MODEL`, `POC_REFERENCE_DATE`,
etc.) já têm defaults definidos no [mise.toml](mise.toml) e não precisam ser alteradas
para rodar localmente. Por padrão a LLM aponta para a API da Groq
(`https://api.groq.com/openai/v1`, modelo `llama-3.3-70b-versatile`).

### 3. Baixar a massa de dados

Baixe os dois CSVs do dataset e coloque-os na pasta `raw_data/`:

- `01_bank_accounts.csv`
- `18_transactions_events.csv`

```
raw_data/
├── 01_bank_accounts.csv
└── 18_transactions_events.csv
```

> Esses arquivos são grandes e **não são versionados** (vide [.gitignore](.gitignore)). O
> Postgres monta a pasta `raw_data/` como read-only e carrega os CSVs automaticamente na
> primeira subida do banco (scripts em [infra/initdb/](infra/initdb/)).

### 4. Subir o banco

```bash
mise run db-up
```

Sobe o Postgres via Docker Compose em background e espera ficar *healthy*. Na primeira
execução, os scripts de `initdb` criam o schema e carregam a massa dos CSVs.

## Rodando o projeto

### Stack completo (banco + backend + frontend)

```bash
mise run up
```

Sobe o Postgres (se ainda não estiver de pé), a API em `:8080` e o frontend em `:5173`.
`Ctrl+C` derruba os servers (o banco continua de pé).

Acesse o frontend em **http://localhost:5173**.

### Componentes separados

| Comando             | O que faz                                                       |
| ------------------- | --------------------------------------------------------------- |
| `mise run db-up`    | Sobe o Postgres em background e espera ficar healthy            |
| `mise run db-down`  | Para o Postgres (mantém o volume com os dados)                  |
| `mise run db-reset` | Apaga o volume e recarrega a massa do zero                      |
| `mise run api`      | Sobe só a API HTTP (`:8080`)                                    |
| `mise run web`      | Sobe só o frontend (Vite dev server, proxy `/api` → `:8080`)    |
| `mise run api-test` | Roda os testes do backend (`go test ./...`)                     |
| `mise run swagger`  | Regera a documentação OpenAPI em `backend/docs/`                |
| `mise run db-cleanup` | Filtra a massa, mantendo só as contas com dados mais próximos da realidade — simplifica os testes |

Liste todas as tasks disponíveis com `mise tasks`.

## Estrutura do projeto

```
.
├── backend/    # API Go (cmd/api, internal/, docs/ OpenAPI)
├── frontend/   # SPA Vite + React + TS
├── infra/      # docker-compose.yml e scripts de initdb
├── raw_data/   # CSVs do dataset (não versionados — vide passo 3)
├── mise.toml   # tooling + variáveis de ambiente + tasks
└── PRD.md      # documento de produto
```

## Troubleshooting

- **`mise: command not found`** — instale o mise e reinicie o shell (vide
  [docs de instalação](https://mise.jdx.dev/getting-started.html)).
- **Banco não carregou os dados** — a carga só roda na primeira subida (volume vazio).
  Se você subiu o banco antes de colocar os CSVs em `raw_data/`, rode `mise run db-reset`
  para recriar o volume e recarregar a massa.
- **Erro de conexão com a LLM** — confira se `LLM_API_KEY` está definida no `.env` e se a
  chave é válida para o provedor configurado em `LLM_BASE_URL`.
