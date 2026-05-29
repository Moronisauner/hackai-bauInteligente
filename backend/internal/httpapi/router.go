// Package httpapi monta o servidor HTTP: router Chi, middlewares e handlers.
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/moronisauner/hackai/backend/internal/assistant"
	"github.com/moronisauner/hackai/backend/internal/balance"
	"github.com/moronisauner/hackai/backend/internal/config"
	"github.com/moronisauner/hackai/backend/internal/goal"

	_ "github.com/moronisauner/hackai/backend/docs" // docs gerados pelo swag
)

// Server agrega as dependências compartilhadas pelos handlers.
type Server struct {
	pool      *pgxpool.Pool
	cfg       config.Config
	balance   *balance.Repo
	goals     *goal.Service
	assistant *assistant.Client
}

// New monta o http.Handler com todas as rotas da API.
func New(pool *pgxpool.Pool, cfg config.Config) http.Handler {
	s := &Server{
		pool:      pool,
		cfg:       cfg,
		balance:   &balance.Repo{Pool: pool},
		goals:     &goal.Service{Pool: pool},
		assistant: assistant.NewClient(cfg.LLMBaseURL, cfg.LLMModel, cfg.LLMAPIKey),
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", s.healthz)
	r.Get("/config", s.getConfig)

	// UI navegável do Swagger em /swagger/.
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	r.Get("/users", s.ListUsers)
	r.Get("/users/{userID}/accounts", s.ListAccountsByUser)
	r.Post("/users/{userID}/goals", s.CreateGoal)
	r.Get("/users/{userID}/goals", s.ListGoalsByUser)
	r.Post("/users/{userID}/assistant/plan", s.AssistantPlan)

	r.Get("/goals/{goalID}", s.GetGoal)
	r.Post("/goals/{goalID}/allocations", s.AddGoalAllocation)
	r.Delete("/goals/{goalID}/allocations/{accountID}", s.RemoveGoalAllocation)
	r.Post("/goals/{goalID}/backtest", s.RunBacktest)
	r.Get("/goals/{goalID}/backtest", s.GetBacktest)

	return r
}

// healthz responde 200 {"status":"ok"} se o DB responde a um Ping (timeout 1s),
// ou 503 caso contrário.
//
//	@Summary	Health check
//	@Tags		infra
//	@Produce	json
//	@Success	200	{object}	map[string]string
//	@Failure	503	{object}	map[string]string
//	@Router		/healthz [get]
func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	if err := s.pool.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ConfigDTO expõe a configuração relevante ao frontend.
type ConfigDTO struct {
	POCReferenceDate string `json:"poc_reference_date" example:"2025-01-01"`
}

// getConfig devolve a configuração pública da POC (em especial a data de
// referência usada em todo cálculo temporal).
//
//	@Summary	Configuração pública da POC
//	@Tags		infra
//	@Produce	json
//	@Success	200	{object}	httpapi.ConfigDTO
//	@Router		/config [get]
func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, ConfigDTO{
		POCReferenceDate: s.cfg.POCReferenceDate.Format("2006-01-02"),
	})
}

// writeJSON serializa v como JSON com o status code dado.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError devolve um corpo de erro padronizado {"error": "..."}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
