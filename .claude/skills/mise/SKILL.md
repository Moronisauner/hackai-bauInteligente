---
name: mise
description: Ambiente de desenvolvimento do projeto — versões de Go e ferramentas auxiliares são gerenciadas via mise. Use ao rodar comandos, instalar dependências do sistema, ou propor mudanças no stack.
---

# Ambiente de desenvolvimento — mise

## Regra principal

**Todo o ambiente de desenvolvimento é gerenciado por [mise](https://mise.jdx.dev/).**

Versões de linguagens, runtimes e ferramentas auxiliares ficam declaradas em [mise.toml](../../../mise.toml). Não instale ou invoque versões de ferramentas fora desse fluxo — se a ferramenta não está no `mise.toml`, ela não existe para o projeto.

## Stack

O projeto é **Go** (backend) + **Node/React** (frontend SPA Vite). Tanto o `go`
quanto o `node`/`npm` resolvidos vêm do `mise` — não confie nos do sistema.

## Como usar

- **Executar uma ferramenta:** `mise exec -- <comando>` (ou apenas `<comando>` se o `mise` já está ativo no shell).
- **Instalar/sincronizar tudo que o projeto pede:** `mise install`.
- **Adicionar uma nova ferramenta:** edite o `mise.toml` (seção `[tools]`) e rode `mise install`. Não use `apt`, `brew`, `go install` global, etc. para o que pode ser declarado no `mise`.
- **Tasks do projeto** (quando existirem): declaradas em `[tasks.*]` no `mise.toml` e executadas via `mise run <task>`.

## O que isso implica para o Claude

- Antes de rodar qualquer comando Go (`go build`, `go test`, `go run`, etc.), assuma que ele é resolvido via `mise`. Se `go` não estiver em `mise.toml`, proponha adicioná-lo em vez de usar o do sistema.
- Para ferramentas comuns do ecossistema Go (`golangci-lint`, `goimports`, `air`, `mockgen`, `sqlc`, etc.), a sugestão padrão é declará-las no `mise.toml`, não `go install` global.
- Para reproduzir builds/testes/lint, prefira `mise run <task>` a comandos soltos — assim a definição vive no repo.
- Ao sugerir uma dependência de sistema (cliente de banco, ferramenta de build), o caminho é "adicione no `mise.toml`", não "rode `apt install`".

## Estado atual

Ferramentas declaradas em [mise.toml](../../../mise.toml):

```toml
[tools]
go = "latest"
node = "lts"                                        # runtime do frontend (Vite)
"go:github.com/swaggo/swag/cmd/swag" = "v1.16.4"    # gerador de docs OpenAPI
```

Tasks disponíveis (`mise run <task>`):

- `api` — sobe a API HTTP (`backend/cmd/api`, porta `:8080`).
- `web` — sobe o frontend (Vite dev server em `:5173`, proxy `/api` → `:8080`).
- `swagger` — regera `backend/docs/swagger.{json,yaml}` a partir das annotations.

Comandos `npm`/`node` do frontend rodam via `mise exec -- npm ...` (ou direto, se
o `mise` já está ativo no shell), a partir de `frontend/`.

Atualize esta skill ao mudar `mise.toml` (novas ferramentas ou tasks).
