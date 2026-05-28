# 02-backend/02 — Carregamento de configuração

## Objetivo
Implementar leitura de configuração via env vars, com falha explícita se obrigatórias estiverem ausentes. Em especial: `POC_REFERENCE_DATE`.

## Pré-requisitos
- 02-backend/01

## Passos
1. Em `internal/config/config.go`, expor:
   ```go
   type Config struct {
       DatabaseURL      string    // obrigatória
       POCReferenceDate time.Time // obrigatória, formato YYYY-MM-DD
       HTTPPort         string    // default ":8080"
   }
   func Load() (Config, error)
   ```
2. `POC_REFERENCE_DATE` deve ser parseada com `time.Parse("2006-01-02", v)`. Erro de parse → retornar erro descritivo.
3. Erros devem ser `fmt.Errorf("config: ...")`, sem `log.Fatal` dentro do package (o main decide).
4. Em `cmd/api/main.go`, chamar `config.Load()` e dar `log.Fatal` se erro.
5. Adicionar `backend/.env.example` com as 3 variáveis.

## Critério de aceite
- [ ] `cd backend && POC_REFERENCE_DATE=2024-06-01 DATABASE_URL=... go run ./cmd/api` imprime config sem erro.
- [ ] Rodar sem `POC_REFERENCE_DATE` → falha com mensagem clara.
- [ ] Rodar com `POC_REFERENCE_DATE=2024-13-99` → falha com mensagem clara.

## Referências PRD
- §8 (tratamento de datas — `POC_REFERENCE_DATE` é crítico)
