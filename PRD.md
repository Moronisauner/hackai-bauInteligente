# PRD — Centralizador de Saldo (POC)

## 1. Visão geral

Esta POC valida a hipótese de um **centralizador de saldo**: uma ferramenta que olha para as várias contas que um cliente possui (potencialmente em diferentes instituições, conforme dados de Open Finance) e permite estruturar um **objetivo financeiro** com saques mensais programados a partir dessas contas.

A POC trabalha sobre uma **massa de dados históricos** já presente no banco ([schema.sql](schema.sql)) — não há ingestão em tempo real e não há execução real de saques. Toda a simulação ocorre sobre eventos passados.

## 2. Objetivo da POC

Responder a três perguntas:

1. Conseguimos, a partir dos dados disponíveis, reconstruir uma visão coerente das contas e do saldo de um cliente em uma data de referência?
2. Um cliente consegue definir um objetivo (meta + prazo) e distribuir contribuições entre suas contas de forma simples?
3. Dado o histórico real de transações, o plano definido teria sido cumprido mês a mês? Onde teria falhado?

**Fora de escopo:** integração com instituições, execução real de transferências, autenticação/onboarding, recomendação automatizada de produtos.

## 3. Personas

- **Cliente final (primário):** quer juntar dinheiro para um objetivo (viagem, reserva, compra) e tem mais de uma conta. Tecnicamente leigo.
- **Time de produto/dados (secundário):** vai analisar o resultado da POC para decidir investimento na próxima etapa.

## 4. Jornada do usuário

1. **Selecionar cliente** — entrada por `user_id`. Sistema lista os `user_id` distintos disponíveis na massa com contagem de contas vinculadas. (CPF/`customer_document` não é usado nesta POC — pode entrar em versão futura.)
2. **Visualizar contas** — sistema mostra todas as contas de `bank_accounts` daquele cliente com:
   - Instituição (`brand_name`), tipo (`type`), agência e número
   - **Saldo na data de referência** (ver §6)
3. **Criar objetivo + conta baú** — usuário informa:
   - Nome do objetivo (texto livre, ex: "Viagem Japão")
   - Valor-alvo (R$)
   - Prazo em meses
   - Data de início (default: data de referência configurada)

   Ao salvar, o sistema **cria automaticamente uma "conta baú" associada ao objetivo**. Essa conta é um cofre virtual: não existe na instituição, vive apenas no banco da POC, e é onde o saldo do objetivo se acumula mês a mês.
4. **Selecionar contas-fonte** — usuário marca uma ou mais contas reais (de `bank_accounts`) que servirão de origem para os saques mensais. A conta baú é sempre o destino.
5. **Distribuir percentuais** — para cada conta-fonte selecionada, usuário define o **% de contribuição da conta**: quanto da *evolução mensal* daquela conta (o quanto o saldo cresceu no mês — entradas menos saídas) será reservado para a meta. Os percentuais são **independentes por conta e não precisam somar 100%**. Como o valor reservado depende de quanto a conta crescer em cada mês, **não há um valor mensal fixo** definido na criação — a projeção só existe após o backtest.
6. **Resumo + simulação histórica** — sistema mostra:
   - Plano consolidado (qual valor sai de qual conta, em qual dia do mês, todos com destino à conta baú)
   - **Backtest mês a mês** rodando contra o histórico real do cliente: em cada mês mede-se **quanto cada conta-fonte evoluiu** (entradas − saídas do mês) e reserva-se o percentual definido sobre essa evolução, limitado ao saldo disponível no dia do saque. Quantos meses teriam reservado o valor cheio, quantos parcialmente, e quantos nada (conta sem evolução ou sem saldo disponível)?
   - **Evolução da conta baú**: saldo acumulado mês a mês, mostrando a trajetória até a meta.
   - Indicador final: meta seria atingida? Quanto teria sido acumulado no baú ao fim do prazo?

## 5. Requisitos funcionais

### RF-01 — Seleção de cliente
- Listar `user_id` distintos com contagem de contas vinculadas.
- Filtro por `user_id` (busca textual).

### RF-02 — Listagem de contas
- Retornar todas as contas de `bank_accounts` do cliente com `status = 'AVAILABLE'`.
- Exibir saldo na data de referência (calculado conforme §6).

### RF-03 — Criação de objetivo + conta baú
- Persistir objetivos em nova tabela `goals` (sugestão de schema em §7).
- Validar: valor-alvo > 0, prazo entre 1 e 60 meses, data de início ≥ data mínima dos dados do cliente.
- **Ao criar o objetivo, criar automaticamente uma `goal_vault` (conta baú) atrelada**. Saldo inicial = 0. Uma conta baú por objetivo.

### RF-04 — Alocação por percentual (fatia da evolução)
- Para cada conta-fonte selecionada, aceitar percentual inteiro de 1 a 100.
- O percentual **não** é a fatia de um aporte fixo: ele define **quanto da evolução mensal da própria conta** (entradas − saídas do mês) é reservado para a meta.
- Os percentuais são **independentes por conta** — **não precisam somar 100%**. Cada conta pode contribuir com qualquer % de 1 a 100 do seu próprio crescimento.
- Não há valor mensal projetável na criação: o valor reservado depende de quanto cada conta evoluir em cada mês (vide RF-05).
- `target_amount` e `duration_months` passam a ser a **meta a atingir e a janela do backtest** — o baú acumula o que as evoluções renderem, podendo bater a meta antes, depois, ou não bater.

### RF-05 — Simulação histórica (backtest)

**Evolução do mês** de uma conta (entradas − saídas do mês de competência):

```
evolução(conta, mês) = saldo(conta, fim do mês) − saldo(conta, abertura do mês)
```

Ambos os saldos via §6. "Abertura" = primeiro instante do mês de competência; "fim" = último instante do mês de competência (captura todas as transações do mês-calendário). O saque ocorre sempre no **dia 1** do mês de competência.

Para cada mês `m` do plano, para cada conta-fonte:
1. Calcular a evolução do mês.
2. Se `evolução <= 0` → **`SKIPPED_NO_GROWTH`**: a conta não cresceu no mês, nada é reservado (`amount = 0`).
3. Senão, `reserva_alvo = round(evolução × percentual / 100, 2)`.
4. Calcular o **saldo disponível no dia do saque** = `saldo(conta, movement_date)` (= saldo na abertura do mês, pois o saque é no dia 1) − reservas já feitas por essa conta nos meses anteriores do backtest (o débito sintético reduz o disponível dos meses seguintes). Então:
   - `disponível <= 0` → **`FAILED_INSUFFICIENT_BALANCE`**: a conta evoluiu mas não há saldo disponível; reserva 0.
   - `disponível >= reserva_alvo` → **`COMPLETED`**: reserva o alvo cheio.
   - `0 < disponível < reserva_alvo` → **`PARTIAL`**: reserva apenas o disponível (nunca move mais do que há na conta).
5. O valor reservado vira **crédito na conta baú** (`goal_vault_movements`). Não tenta cobrir a falha de uma conta com outra (cada alocação é avaliada isoladamente).
- Resultado armazenado: array de movimentos `{ mes, conta_fonte_id, vault_id, status, valor_reservado }`.

### RF-06 — Visualização do resultado
- KPIs: **% de meses cumpridos** (status `COMPLETED` sobre o total), **saldo final do baú vs. meta**, conta-fonte com maior taxa de não-cumprimento.
- Tabela detalhada mês a mês: para cada mês, status por conta-fonte (cheio / parcial / sem evolução / sem saldo) + valor reservado + saldo acumulado do baú ao final daquele mês.
- Gráfico de evolução do saldo do baú no tempo (linha simples).

## 6. Cálculo de saldo (decisão crítica)

O saldo de cada conta em uma data `D` é **reconstruído a partir de `transaction_events`**:

```
saldo(conta, D) = Σ (créditos até D) − Σ (débitos até D)
```

Onde:
- `credit_debit_type = 'CREDITO'` soma `amount`
- `credit_debit_type = 'DEBITO'` subtrai `amount`
- Apenas transações com `transaction_date_time <= D`
- Apenas transações com `completed_authorised_payment_type = 'TRANSACAO_EFETIVADA'`

**Por que reconstruir e não usar `accounts_balances_events` ou `balances_history`?** Auditabilidade: o usuário do POC consegue justificar exatamente como chegamos a cada número, e o backtest fica internamente consistente (saldo e transações sempre batem).

**Saldo inicial:** assume-se 0 no início do histórico. Esta é uma simplificação aceita para POC — significa que clientes cujo histórico não começa do zero terão saldos potencialmente negativos no início. Mitigação: filtrar clientes cuja primeira transação ocorre em data anterior ao início do plano.

## 7. Modelo de dados

### 7.1 Tabelas existentes (consumo)
- [bank_accounts](schema.sql#L6) — cadastro de contas
- [transaction_events](schema.sql#L97) — fonte primária do saldo reconstruído
- `balances_history` e `accounts_balances_events` — **não usadas na POC** (mantidas para validação cruzada futura)

### 7.2 Tabelas novas (POC)

```sql
CREATE TABLE goals (
    id              VARCHAR(64) PRIMARY KEY,
    user_id         VARCHAR(64) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    target_amount   NUMERIC(18, 2) NOT NULL,
    duration_months INTEGER NOT NULL,
    start_date      DATE NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE goal_allocations (
    id              VARCHAR(64) PRIMARY KEY,
    goal_id         VARCHAR(64) NOT NULL REFERENCES goals(id),
    account_id      VARCHAR(64) NOT NULL,  -- bank_accounts.id (conta-fonte)
    percentage      INTEGER NOT NULL CHECK (percentage BETWEEN 1 AND 100)
);

-- Conta baú: cofre virtual atrelado ao objetivo. Criada junto com o goal.
CREATE TABLE goal_vaults (
    id          VARCHAR(64) PRIMARY KEY,
    goal_id     VARCHAR(64) NOT NULL UNIQUE REFERENCES goals(id),
    user_id     VARCHAR(64) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Movimentos da conta baú gerados pelo backtest. Saldo do baú = SUM(amount).
CREATE TABLE goal_vault_movements (
    id                VARCHAR(64) PRIMARY KEY,
    vault_id          VARCHAR(64) NOT NULL REFERENCES goal_vaults(id),
    source_account_id VARCHAR(64) NOT NULL,  -- bank_accounts.id de origem
    reference_month   DATE NOT NULL,         -- mês de competência (primeiro dia)
    movement_date     DATE NOT NULL,         -- data efetiva do saque simulado (sempre o dia 1)
    amount            NUMERIC(18, 2) NOT NULL,
    -- COMPLETED | PARTIAL | SKIPPED_NO_GROWTH | FAILED_INSUFFICIENT_BALANCE
    status            VARCHAR(40) NOT NULL,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_goal_vault_movements_vault_id ON goal_vault_movements(vault_id);
CREATE INDEX idx_goal_vault_movements_ref_month ON goal_vault_movements(reference_month);
```

> Nota: movimentos `SKIPPED_NO_GROWTH` (conta não evoluiu) e `FAILED_INSUFFICIENT_BALANCE` (evoluiu mas sem saldo disponível) são persistidos com `amount = 0` para manter o histórico de tentativas. `PARTIAL` guarda o valor efetivamente reservado (menor que o alvo). O saldo do baú = `SUM(amount)`, ou seja, soma de `COMPLETED` + `PARTIAL`.

## 8. Tratamento de datas (atenção redobrada)

Como a massa é histórica, **a aplicação não pode usar `NOW()` como referência temporal de produto**.

- **Configuração:** variável de ambiente `POC_REFERENCE_DATE` (formato `YYYY-MM-DD`).
- Toda regra de negócio que precisaria de "hoje" deve consultar essa variável.
- `created_at` técnico (auditoria de quando o registro foi inserido no banco) **continua usando `NOW()`** — não confundir com data de negócio.
- Sugestão: log de inicialização da app deve imprimir a `POC_REFERENCE_DATE` em destaque para evitar confusão.
- A `start_date` do objetivo, por padrão, é `POC_REFERENCE_DATE`; usuário pode adiantar ou atrasar dentro do range disponível na massa.

## 9. Requisitos não-funcionais

- **Stack:** a definir (sugestão: Python + FastAPI + Postgres, frontend leve em React/HTMX/Streamlit — a escolher conforme familiaridade do time).
- **Performance:** consultas de saldo reconstruído devem retornar em < 2s para um cliente com até 10k transações. Índice em `transaction_events(account_id, transaction_date_time)` provavelmente necessário.
- **Sem autenticação real** na POC. Acesso restrito por rede interna.
- **Dados anonimizados:** a massa de dados já é anônima, então **não é necessário mascaramento** na UI nem em logs. `user_id`, números de conta e demais identificadores podem ser exibidos integralmente.

## 10. Critérios de sucesso da POC

1. Conseguimos rodar o fluxo completo (seleção → objetivo → simulação) com **pelo menos 5 clientes diferentes** da massa.
2. O backtest produz resultados **internamente coerentes** (soma de saques cumpridos = valor acumulado).
3. Time de produto consegue, vendo um relatório de backtest, **decidir go/no-go** para a próxima fase.

## 11. Riscos e perguntas em aberto

- **Saldo inicial assumido como zero:** quão distorcido fica o backtest? Talvez precisemos calibrar com um saldo inicial sintético baseado em `balances_history` mais antigo.
- **Múltiplas moedas:** schema permite, mas a POC assume `BRL`. Filtrar `currency = 'BRL'`?
- **Dia do saque:** o saque ocorre sempre no **dia 1** do mês de competência — não é mais configurável.
- **Granularidade do percentual:** inteiro (1–100) é suficiente, ou precisamos de uma casa decimal?
- **Janela da evolução mensal (RF-05):** medimos a evolução pelo mês-calendário cheio (saldo do fim − saldo da abertura do mês de competência). O saque/reserva ocorre no dia 1 e é limitado ao saldo disponível nesse dia (= saldo na abertura do mês).
- **Múltiplos objetivos por cliente:** suportado pelo schema, mas a UI da POC pode focar em um objetivo de cada vez para simplificar.

## 12. Próximos passos

1. Validação deste PRD com o time.
2. Definição de stack e setup do projeto.
3. Identificação de 5–10 clientes "ricos" na massa (com histórico longo e várias contas) para servirem de casos-piloto.
4. Implementação iterativa: começar pelo cálculo de saldo reconstruído isolado, validar, depois construir o fluxo por cima.
