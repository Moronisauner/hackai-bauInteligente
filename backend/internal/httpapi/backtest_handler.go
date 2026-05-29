package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/moronisauner/hackai/backend/internal/backtest"
	"github.com/moronisauner/hackai/backend/internal/balance"
)

// BacktestSummaryDTO agrega os KPIs do backtest (RF-06).
type BacktestSummaryDTO struct {
	CompletedMonthsPct float64 `json:"completed_months_pct" example:"0.83"`
	CompletedCount     int     `json:"completed_count"`
	FailedCount        int     `json:"failed_count"`
	VaultBalance       string  `json:"vault_balance" example:"8400.00"`
	TargetAmount       string  `json:"target_amount" example:"10000.00"`
	GoalReached        bool    `json:"goal_reached"`
	WorstAccountID     string  `json:"worst_account_id"`
}

// BacktestMovementDTO é um movimento do baú na resposta. Entradas e Saidas são
// os totais movimentados pela conta-fonte naquele mês (créditos e débitos
// efetivados), exibidos no detalhe da célula da tabela mês a mês.
type BacktestMovementDTO struct {
	ReferenceMonth string `json:"reference_month" example:"2025-06-01"`
	AccountID      string `json:"account_id"`
	Status         string `json:"status"`
	Amount         string `json:"amount" example:"500.00"`
	Entradas       string `json:"entradas" example:"2000.00"`
	Saidas         string `json:"saidas" example:"1500.00"`
}

// VaultEvolutionDTO é o saldo acumulado do baú ao fim de um mês.
type VaultEvolutionDTO struct {
	Month   string `json:"month" example:"2025-06-01"`
	Balance string `json:"balance" example:"1000.00"`
}

// BacktestResultDTO é a resposta completa do backtest.
type BacktestResultDTO struct {
	Summary        BacktestSummaryDTO    `json:"summary"`
	Movements      []BacktestMovementDTO `json:"movements"`
	VaultEvolution []VaultEvolutionDTO   `json:"vault_evolution"`
}

// goalPlan agrega os dados de um goal necessários ao backtest.
type goalPlan struct {
	vaultID     string
	target      decimal.Decimal
	duration    int
	startDate   time.Time
	allocations []backtest.Allocation
}

// loadGoalPlan carrega goal + vault + allocations e monta o backtest.Plan.
func (s *Server) loadGoalPlan(ctx context.Context, goalID string) (goalPlan, error) {
	var gp goalPlan
	err := s.pool.QueryRow(ctx, `
		SELECT v.id, g.target_amount, g.duration_months, g.start_date
		FROM goals g
		JOIN goal_vaults v ON v.goal_id = g.id
		WHERE g.id = $1`, goalID,
	).Scan(&gp.vaultID, &gp.target, &gp.duration, &gp.startDate)
	if err != nil {
		return goalPlan{}, err
	}

	rows, err := s.pool.Query(ctx,
		`SELECT account_id, percentage FROM goal_allocations WHERE goal_id = $1`, goalID)
	if err != nil {
		return goalPlan{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var accountID string
		var pct int
		if err := rows.Scan(&accountID, &pct); err != nil {
			return goalPlan{}, err
		}
		gp.allocations = append(gp.allocations, backtest.Allocation{
			AccountID:  accountID,
			Percentage: pct,
		})
	}
	return gp, rows.Err()
}

// flowsByAccountMonth devolve, indexado por "accountID|YYYY-MM-01", o total de
// entradas e saídas de cada conta-fonte em cada mês de competência dos
// movimentos. É enriquecimento de exibição da tabela mês a mês: derivamos a
// janela e as contas dos próprios movimentos e consultamos transaction_events.
func (s *Server) flowsByAccountMonth(ctx context.Context, movements []backtest.Movement) (map[string]balance.MonthlyFlow, error) {
	if len(movements) == 0 {
		return nil, nil
	}

	accSet := make(map[string]struct{})
	minM, maxM := movements[0].ReferenceMonth, movements[0].ReferenceMonth
	for _, mv := range movements {
		accSet[mv.AccountID] = struct{}{}
		if mv.ReferenceMonth.Before(minM) {
			minM = mv.ReferenceMonth
		}
		if mv.ReferenceMonth.After(maxM) {
			maxM = mv.ReferenceMonth
		}
	}
	accounts := make([]string, 0, len(accSet))
	for a := range accSet {
		accounts = append(accounts, a)
	}

	// Janela [primeiro dia do mês mais antigo, primeiro dia do mês seguinte ao
	// mais recente) — exclusiva no fim para fechar o último mês inteiro.
	flows, err := s.balance.MonthlyFlows(ctx, accounts, minM, maxM.AddDate(0, 1, 0))
	if err != nil {
		return nil, err
	}

	out := make(map[string]balance.MonthlyFlow, len(flows))
	for _, f := range flows {
		out[f.AccountID+"|"+f.Month.Format("2006-01-02")] = f
	}
	return out, nil
}

// RunBacktest executa o backtest do objetivo e persiste os movimentos de forma
// idempotente (apaga os anteriores do vault e insere os novos numa transação).
//
//	@Summary	Executa o backtest de um objetivo
//	@Tags		backtest
//	@Produce	json
//	@Param		goalID	path		string	true	"ID do objetivo"
//	@Success	200		{object}	httpapi.BacktestResultDTO
//	@Failure	404		{object}	map[string]string
//	@Failure	500		{object}	map[string]string
//	@Router		/goals/{goalID}/backtest [post]
func (s *Server) RunBacktest(w http.ResponseWriter, r *http.Request) {
	goalID := chi.URLParam(r, "goalID")
	ctx := r.Context()

	gp, err := s.loadGoalPlan(ctx, goalID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "goal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load goal")
		return
	}

	plan := backtest.Plan{
		StartDate:      gp.startDate,
		DurationMonths: gp.duration,
		TargetAmount:   gp.target,
		Allocations:    gp.allocations,
	}

	balanceFn := func(ctx context.Context, accountID string, atDate time.Time) (decimal.Decimal, error) {
		return s.balance.Reconstruct(ctx, accountID, atDate)
	}

	movements, err := backtest.Run(ctx, plan, balanceFn)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "backtest run failed")
		return
	}

	// Persiste de forma idempotente: apaga anteriores e insere os novos.
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to begin tx")
		return
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx,
		`DELETE FROM goal_vault_movements WHERE vault_id = $1`, gp.vaultID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to clear previous movements")
		return
	}
	for _, mv := range movements {
		if _, err := tx.Exec(ctx, `
			INSERT INTO goal_vault_movements
			    (id, vault_id, source_account_id, reference_month, movement_date, amount, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.NewString(), gp.vaultID, mv.AccountID, mv.ReferenceMonth, mv.MovementDate, mv.Amount, mv.Status,
		); err != nil {
			log.Printf("backtest insert movement failed: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to insert movement")
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit movements")
		return
	}

	flows, err := s.flowsByAccountMonth(ctx, movements)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load monthly flows")
		return
	}

	writeJSON(w, http.StatusOK, buildResult(gp.target, movements, flows))
}

// GetBacktest lê os movimentos persistidos do baú e devolve KPIs + séries.
// 404 se nenhum movimento foi persistido ainda.
//
//	@Summary	Consulta o resultado do backtest
//	@Tags		backtest
//	@Produce	json
//	@Param		goalID	path		string	true	"ID do objetivo"
//	@Success	200		{object}	httpapi.BacktestResultDTO
//	@Failure	404		{object}	map[string]string
//	@Failure	500		{object}	map[string]string
//	@Router		/goals/{goalID}/backtest [get]
func (s *Server) GetBacktest(w http.ResponseWriter, r *http.Request) {
	goalID := chi.URLParam(r, "goalID")
	ctx := r.Context()

	var vaultID string
	var target decimal.Decimal
	err := s.pool.QueryRow(ctx, `
		SELECT v.id, g.target_amount
		FROM goals g JOIN goal_vaults v ON v.goal_id = g.id
		WHERE g.id = $1`, goalID,
	).Scan(&vaultID, &target)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "goal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load goal")
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT source_account_id, reference_month, movement_date, amount, status
		FROM goal_vault_movements
		WHERE vault_id = $1
		ORDER BY reference_month, source_account_id`, vaultID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load movements")
		return
	}
	defer rows.Close()

	var movements []backtest.Movement
	for rows.Next() {
		var mv backtest.Movement
		if err := rows.Scan(&mv.AccountID, &mv.ReferenceMonth, &mv.MovementDate, &mv.Amount, &mv.Status); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan movement")
			return
		}
		movements = append(movements, mv)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read movements")
		return
	}

	if len(movements) == 0 {
		writeError(w, http.StatusNotFound, "no backtest found; run POST /goals/"+goalID+"/backtest first")
		return
	}

	flows, err := s.flowsByAccountMonth(ctx, movements)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load monthly flows")
		return
	}

	writeJSON(w, http.StatusOK, buildResult(target, movements, flows))
}

// buildResult calcula os KPIs e séries a partir dos movimentos (uma passada
// principal + agregação por mês), conforme RF-05/RF-06 e §10.
func buildResult(target decimal.Decimal, movements []backtest.Movement, flows map[string]balance.MonthlyFlow) BacktestResultDTO {
	var (
		completed   int
		vaultBal    = decimal.Zero
		monthOrder  []string
		monthCumEOM = map[string]decimal.Decimal{}
		failByAcc   = map[string]int{}
		totalByAcc  = map[string]int{}
		dtoMoves    = make([]BacktestMovementDTO, 0, len(movements))
	)

	for _, mv := range movements {
		month := mv.ReferenceMonth.Format("2006-01-02")
		if _, seen := monthCumEOM[month]; !seen {
			monthOrder = append(monthOrder, month)
		}

		totalByAcc[mv.AccountID]++
		// vaultBal soma o valor de TODOS os movimentos; SKIPPED tem Amount = 0,
		// logo o saldo é a soma dos COMPLETED.
		vaultBal = vaultBal.Add(mv.Amount)
		if mv.Status == backtest.StatusCompleted {
			completed++
		} else {
			failByAcc[mv.AccountID]++
		}
		// Saldo acumulado do baú ao fim de cada mês (carrega o acumulado total).
		monthCumEOM[month] = vaultBal

		// flows ausente para o par conta/mês → entradas/saídas zeradas (zero value
		// de decimal.Decimal formata como "0.00").
		f := flows[mv.AccountID+"|"+month]
		dtoMoves = append(dtoMoves, BacktestMovementDTO{
			ReferenceMonth: month,
			AccountID:      mv.AccountID,
			Status:         mv.Status,
			Amount:         mv.Amount.StringFixed(2),
			Entradas:       f.Entradas.StringFixed(2),
			Saidas:         f.Saidas.StringFixed(2),
		})
	}

	total := len(movements)
	failed := total - completed
	pct := 0.0
	if total > 0 {
		pct = float64(completed) / float64(total)
	}

	// worst_account = maior taxa de movimentos não-COMPLETED por conta (RF-06).
	worst := ""
	worstRate := -1.0
	for acc, tot := range totalByAcc {
		rate := float64(failByAcc[acc]) / float64(tot)
		if rate > worstRate {
			worstRate = rate
			worst = acc
		}
	}

	evolution := make([]VaultEvolutionDTO, 0, len(monthOrder))
	for _, m := range monthOrder {
		evolution = append(evolution, VaultEvolutionDTO{
			Month:   m,
			Balance: monthCumEOM[m].StringFixed(2),
		})
	}

	return BacktestResultDTO{
		Summary: BacktestSummaryDTO{
			CompletedMonthsPct: pct,
			CompletedCount:     completed,
			FailedCount:        failed,
			VaultBalance:       vaultBal.StringFixed(2),
			TargetAmount:       target.StringFixed(2),
			GoalReached:        vaultBal.GreaterThanOrEqual(target),
			WorstAccountID:     worst,
		},
		Movements:      dtoMoves,
		VaultEvolution: evolution,
	}
}
