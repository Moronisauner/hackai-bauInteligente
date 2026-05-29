package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/moronisauner/hackai/backend/internal/goal"
)

// AllocationRequest é uma alocação no corpo de criação de objetivo.
type AllocationRequest struct {
	AccountID  string `json:"account_id"`
	Percentage int    `json:"percentage" example:"50"`
}

// CreateGoalRequest é o corpo do POST /users/{userID}/goals.
type CreateGoalRequest struct {
	Name           string              `json:"name" example:"Viagem"`
	TargetAmount   string              `json:"target_amount" example:"10000.00"`
	DurationMonths int                 `json:"duration_months" example:"12"`
	StartDate      string              `json:"start_date" example:"2025-01-01"`
	Allocations    []AllocationRequest `json:"allocations"`
}

// CreateGoalResponse é o corpo de resposta de criação.
type CreateGoalResponse struct {
	GoalID  string `json:"goal_id"`
	VaultID string `json:"vault_id"`
}

// CreateGoal cria atomicamente um objetivo, seu baú e as alocações.
//
//	@Summary	Cria um objetivo (goal + vault + allocations)
//	@Tags		goals
//	@Accept		json
//	@Produce	json
//	@Param		userID	path		string						true	"ID do cliente"
//	@Param		body	body		httpapi.CreateGoalRequest	true	"Dados do objetivo"
//	@Success	201		{object}	httpapi.CreateGoalResponse
//	@Failure	400		{object}	map[string]string
//	@Failure	500		{object}	map[string]string
//	@Router		/users/{userID}/goals [post]
func (s *Server) CreateGoal(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	target, err := decimal.NewFromString(req.TargetAmount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "target_amount must be a decimal string")
		return
	}
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "start_date must be YYYY-MM-DD")
		return
	}

	allocs := make([]goal.Allocation, len(req.Allocations))
	for i, a := range req.Allocations {
		allocs[i] = goal.Allocation{AccountID: a.AccountID, Percentage: a.Percentage}
	}

	in := goal.CreateInput{
		UserID:         userID,
		Name:           req.Name,
		TargetAmount:   target,
		DurationMonths: req.DurationMonths,
		StartDate:      startDate,
		Allocations:    allocs,
	}

	goalID, vaultID, err := s.goals.Create(r.Context(), in)
	if err != nil {
		if goal.IsValidation(err) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create goal")
		return
	}

	writeJSON(w, http.StatusCreated, CreateGoalResponse{GoalID: goalID, VaultID: vaultID})
}

// GoalSummaryDTO é o resumo de um objetivo na listagem.
type GoalSummaryDTO struct {
	GoalID         string `json:"goal_id"`
	Name           string `json:"name"`
	TargetAmount   string `json:"target_amount"`
	DurationMonths int    `json:"duration_months"`
	StartDate      string `json:"start_date"`
}

// ListGoalsByUser lista os objetivos de um cliente (resumo).
//
//	@Summary	Lista objetivos de um cliente
//	@Tags		goals
//	@Produce	json
//	@Param		userID	path		string	true	"ID do cliente"
//	@Success	200		{array}		httpapi.GoalSummaryDTO
//	@Failure	500		{object}	map[string]string
//	@Router		/users/{userID}/goals [get]
func (s *Server) ListGoalsByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	const sql = `
		SELECT id, name, target_amount, duration_months, start_date
		FROM goals
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(r.Context(), sql, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list goals")
		return
	}
	defer rows.Close()

	out := make([]GoalSummaryDTO, 0)
	for rows.Next() {
		var (
			g         GoalSummaryDTO
			target    decimal.Decimal
			startDate time.Time
		)
		if err := rows.Scan(&g.GoalID, &g.Name, &target, &g.DurationMonths, &startDate); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan goal")
			return
		}
		g.TargetAmount = target.StringFixed(2)
		g.StartDate = startDate.Format("2006-01-02")
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read goals")
		return
	}

	writeJSON(w, http.StatusOK, out)
}

// GoalAllocationDTO é uma alocação enriquecida com dados da conta-fonte. O
// percentual é a fatia da evolução mensal da conta (RF-04) — não há mais valor
// mensal fixo projetado.
type GoalAllocationDTO struct {
	AccountID  string `json:"account_id"`
	Percentage int    `json:"percentage"`
	BrandName  string `json:"brand_name"`
	Number     string `json:"number"`
	Type       string `json:"type"`
}

// GoalDTO é o detalhe de um objetivo com alocações e vault.
type GoalDTO struct {
	GoalID         string              `json:"goal_id"`
	UserID         string              `json:"user_id"`
	Name           string              `json:"name"`
	TargetAmount   string              `json:"target_amount"`
	DurationMonths int                 `json:"duration_months"`
	StartDate      string              `json:"start_date"`
	VaultID        string              `json:"vault_id"`
	Allocations    []GoalAllocationDTO `json:"allocations"`
}

// GetGoal devolve o detalhe de um objetivo, com alocações (incluindo dados da
// conta-fonte) e o vault_id. 404 se não existir.
//
//	@Summary	Detalhe de um objetivo
//	@Tags		goals
//	@Produce	json
//	@Param		goalID	path		string	true	"ID do objetivo"
//	@Success	200		{object}	httpapi.GoalDTO
//	@Failure	404		{object}	map[string]string
//	@Failure	500		{object}	map[string]string
//	@Router		/goals/{goalID} [get]
func (s *Server) GetGoal(w http.ResponseWriter, r *http.Request) {
	goalID := chi.URLParam(r, "goalID")
	ctx := r.Context()

	var (
		dto       GoalDTO
		target    decimal.Decimal
		startDate time.Time
	)
	err := s.pool.QueryRow(ctx, `
		SELECT g.id, g.user_id, g.name, g.target_amount, g.duration_months,
		       g.start_date, v.id
		FROM goals g
		JOIN goal_vaults v ON v.goal_id = g.id
		WHERE g.id = $1`, goalID,
	).Scan(&dto.GoalID, &dto.UserID, &dto.Name, &target, &dto.DurationMonths,
		&startDate, &dto.VaultID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "goal not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load goal")
		return
	}
	dto.TargetAmount = target.StringFixed(2)
	dto.StartDate = startDate.Format("2006-01-02")

	rows, err := s.pool.Query(ctx, `
		SELECT ga.account_id, ga.percentage, ba.brand_name, ba.number, ba.type
		FROM goal_allocations ga
		JOIN bank_accounts ba ON ba.id = ga.account_id
		WHERE ga.goal_id = $1
		ORDER BY ga.percentage DESC`, goalID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load allocations")
		return
	}
	defer rows.Close()

	dto.Allocations = make([]GoalAllocationDTO, 0)
	for rows.Next() {
		var (
			a                            GoalAllocationDTO
			brand, number, accType       *string
		)
		if err := rows.Scan(&a.AccountID, &a.Percentage, &brand, &number, &accType); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan allocation")
			return
		}
		a.BrandName = deref(brand)
		a.Number = deref(number)
		a.Type = deref(accType)
		dto.Allocations = append(dto.Allocations, a)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read allocations")
		return
	}

	writeJSON(w, http.StatusOK, dto)
}

// AddAllocationRequest é o corpo do POST /goals/{goalID}/allocations.
type AddAllocationRequest struct {
	AccountID  string `json:"account_id"`
	Percentage int    `json:"percentage" example:"50"`
}

// AddGoalAllocation adiciona uma conta-fonte a um objetivo existente. O
// percentual é a fatia da evolução mensal da conta que vai pra meta (RF-04).
//
//	@Summary	Adiciona uma alocação (conta-fonte) a um objetivo
//	@Tags		goals
//	@Accept		json
//	@Produce	json
//	@Param		goalID	path	string						true	"ID do objetivo"
//	@Param		body	body	httpapi.AddAllocationRequest	true	"Conta e percentual"
//	@Success	204
//	@Failure	400	{object}	map[string]string
//	@Failure	404	{object}	map[string]string
//	@Failure	500	{object}	map[string]string
//	@Router		/goals/{goalID}/allocations [post]
func (s *Server) AddGoalAllocation(w http.ResponseWriter, r *http.Request) {
	goalID := chi.URLParam(r, "goalID")

	var req AddAllocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := s.goals.AddAllocation(r.Context(), goalID, req.AccountID, req.Percentage); err != nil {
		switch {
		case goal.IsNotFound(err):
			writeError(w, http.StatusNotFound, err.Error())
		case goal.IsValidation(err):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to add allocation")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveGoalAllocation remove uma conta-fonte de um objetivo. Mantém ao menos
// uma alocação por objetivo.
//
//	@Summary	Remove uma alocação (conta-fonte) de um objetivo
//	@Tags		goals
//	@Produce	json
//	@Param		goalID		path	string	true	"ID do objetivo"
//	@Param		accountID	path	string	true	"ID da conta-fonte"
//	@Success	204
//	@Failure	400	{object}	map[string]string
//	@Failure	404	{object}	map[string]string
//	@Failure	500	{object}	map[string]string
//	@Router		/goals/{goalID}/allocations/{accountID} [delete]
func (s *Server) RemoveGoalAllocation(w http.ResponseWriter, r *http.Request) {
	goalID := chi.URLParam(r, "goalID")
	accountID := chi.URLParam(r, "accountID")

	if err := s.goals.RemoveAllocation(r.Context(), goalID, accountID); err != nil {
		switch {
		case goal.IsNotFound(err):
			writeError(w, http.StatusNotFound, err.Error())
		case goal.IsValidation(err):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to remove allocation")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
