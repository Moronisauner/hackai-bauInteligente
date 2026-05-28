---
name: mise
description: Ambiente de desenvolvimento do projeto — versões de Go e ferramentas auxiliares são gerenciadas via mise. Use ao rodar comandos, instalar dependências do sistema, ou propor mudanças no stack.
---

# Ambiente de desenvolvimento — mise

## Regra principal

**Todo o ambiente de desenvolvimento é gerenciado por [mise](https://mise.jdx.dev/).**

Versões de linguagens, runtimes e ferramentas auxiliares ficam declaradas em [mise.toml](../../../mise.toml). Não instale ou invoque versões de ferramentas fora desse fluxo — se a ferramenta não está no `mise.toml`, ela não existe para o projeto.

## Stack

O projeto é **Go**. O `go` resolvido sempre é o do `mise` — não confie em `go` do sistema.

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

O `mise.toml` ainda está vazio. Conforme o stack da POC (ver [PRD.md](../../../PRD.md)) for materializando, registre aqui — exemplo de base esperada:

```toml
[tools]
go = "1.23"
golangci-lint = "latest"
# goimports = "latest"
# air = "latest"          # live reload
# sqlc = "latest"         # se for usar geração a partir do schema.sql
```

Atualize esta skill quando o `mise.toml` for preenchido, listando as ferramentas reais e as tasks relevantes.
