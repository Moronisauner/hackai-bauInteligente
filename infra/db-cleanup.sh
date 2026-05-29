#!/usr/bin/env bash
# Lê infra/keep.conf, popula as temp tables kept_users/kept_accounts no psql e
# executa infra/cleanup.sql (manual; apaga toda a massa fora das listas).
set -euo pipefail

cd "$(dirname "$0")/.."

KEEP_FILE="infra/keep.conf"
COMPOSE="infra/docker-compose.yml"

[ -f "$KEEP_FILE" ] || { echo "Arquivo não encontrado: $KEEP_FILE" >&2; exit 1; }

# Extrai os ids de uma seção [nome], ignorando comentários (#) e linhas vazias.
section() {
  awk -v want="$1" '
    {
      line = $0
      sub(/#.*/, "", line)                              # remove comentário inline
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)     # trim
      if (line == "") next
      if (line ~ /^\[.*\]$/) { cur = substr(line, 2, length(line) - 2); next }
      if (cur == want) print line
    }
  ' "$KEEP_FILE"
}

users=$(section users)
accounts=$(section accounts)

[ -n "$users" ]    || { echo "Seção [users] vazia em $KEEP_FILE"    >&2; exit 1; }
[ -n "$accounts" ] || { echo "Seção [accounts] vazia em $KEEP_FILE" >&2; exit 1; }

echo ">>> Mantendo $(printf '%s\n' "$users" | wc -l) cliente(s) e $(printf '%s\n' "$accounts" | wc -l) conta(s)."

# Monta o stream: cria/popula as temp tables via \copy FROM STDIN e anexa o SQL.
# Usa printf (não echo) para não depender do shell interpretar as barras de \copy.
{
  printf '%s\n' 'CREATE TEMP TABLE kept_users(user_id text);' '\copy kept_users FROM STDIN'
  printf '%s\n' "$users"
  printf '%s\n' '\.'
  printf '%s\n' 'CREATE TEMP TABLE kept_accounts(id text);' '\copy kept_accounts FROM STDIN'
  printf '%s\n' "$accounts"
  printf '%s\n' '\.'
  cat infra/cleanup.sql
} | docker compose -f "$COMPOSE" exec -T postgres \
      psql -U hackai -d hackai -v ON_ERROR_STOP=1
