# 01-infra/03 — Carga inicial do schema e dados

## Objetivo
Garantir que o Postgres do compose suba já com o `schema.sql` aplicado e a massa de `raw_data/` carregada.

## Pré-requisitos
- 01-infra/02

## Passos
1. Copiar (ou symlinkar) `schema.sql` para `infra/initdb/01-schema.sql`.
2. Inspecionar `raw_data/` e identificar o formato dos arquivos (CSV? SQL? JSON?):
   - `ls raw_data/`
   - `file raw_data/* | head`
3. Conforme o formato:
   - **SQL dumps:** copiar para `infra/initdb/02-data.sql` (ou múltiplos arquivos numerados após `02-`).
   - **CSV:** criar `infra/initdb/02-load.sh` que faz `psql ... -c "\copy <tabela> FROM <arquivo> WITH CSV HEADER"` para cada arquivo, e dar `chmod +x`.
4. Recriar o volume (carga só roda no primeiro boot):
   ```
   cd infra && docker compose down -v && docker compose up -d
   ```

## Critério de aceite
- [ ] `docker compose exec postgres psql -U hackai -d hackai -c "SELECT COUNT(*) FROM bank_accounts;"` retorna > 0.
- [ ] `... -c "SELECT COUNT(*) FROM transaction_events;"` retorna > 0.
- [ ] `... -c "SELECT COUNT(DISTINCT user_id) FROM bank_accounts;"` retorna >= 5 (necessário pra §10 do PRD).

## Referências PRD
- §6 (cálculo de saldo depende dessa massa estar carregada)
- §10 (critério de sucesso pede 5+ clientes)
