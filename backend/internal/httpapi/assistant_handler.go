package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/moronisauner/hackai/backend/internal/assistant"
)

// maxLLMAccounts limita quantas contas (as de maior saldo) são oferecidas ao
// assistente para ele escolher, mantendo o prompt enxuto.
const maxLLMAccounts = 10

// ChatMessageDTO é uma mensagem da conversa (role: "user" | "assistant").
type ChatMessageDTO struct {
	Role    string `json:"role" example:"user"`
	Content string `json:"content" example:"Quero juntar 10 mil para uma viagem"`
}

// AssistantPlanRequest é o corpo do POST /users/{userID}/assistant/plan: o
// histórico completo da conversa até aqui (o frontend mantém o estado).
type AssistantPlanRequest struct {
	Messages []ChatMessageDTO `json:"messages"`
}

// ProposedAllocationDTO é uma alocação sugerida pelo assistente, enriquecida
// com dados da conta-fonte para exibição.
type ProposedAllocationDTO struct {
	AccountID  string `json:"account_id"`
	Percentage int    `json:"percentage"`
	Reason     string `json:"reason"`
	BrandName  string `json:"brand_name"`
	Number     string `json:"number"`
}

// PlanOptionDTO é uma das opções de plano (variação de valor/prazo) do objetivo.
type PlanOptionDTO struct {
	Label          string `json:"label"`
	Summary        string `json:"summary"`
	TargetAmount   string `json:"target_amount"`
	DurationMonths int    `json:"duration_months"`
}

// PlanProposalDTO é o objetivo proposto: nome, alocações (comuns a todas as
// opções) e as opções de plano entre as quais o cliente escolhe.
type PlanProposalDTO struct {
	Name        string                  `json:"name"`
	StartDate   string                  `json:"start_date"`
	Allocations []ProposedAllocationDTO `json:"allocations"`
	Plans       []PlanOptionDTO         `json:"plans"`
}

// AssistantPlanResponse é a resposta de um turno do assistente.
type AssistantPlanResponse struct {
	Reply    string           `json:"reply"`
	Done     bool             `json:"done"`
	Proposal *PlanProposalDTO `json:"proposal"`
}

// AssistantPlan roda um turno do assistente de planejamento: dado o histórico
// da conversa e as contas do cliente, devolve a próxima fala e, quando há
// informação suficiente, uma proposta de objetivo + alocações pronta para
// confirmar (que vira um POST /users/{userID}/goals).
//
//	@Summary	Turno do assistente de planejamento (chat)
//	@Tags		assistant
//	@Accept		json
//	@Produce	json
//	@Param		userID	path		string							true	"ID do cliente"
//	@Param		body	body		httpapi.AssistantPlanRequest	true	"Histórico da conversa"
//	@Success	200		{object}	httpapi.AssistantPlanResponse
//	@Failure	400		{object}	map[string]string
//	@Failure	500		{object}	map[string]string
//	@Failure	503		{object}	map[string]string
//	@Router		/users/{userID}/assistant/plan [post]
func (s *Server) AssistantPlan(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	ctx := r.Context()

	var req AssistantPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	accountDTOs, err := s.loadAccounts(ctx, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(accountDTOs) == 0 {
		writeError(w, http.StatusBadRequest, "client has no available accounts to plan with")
		return
	}

	// Índice de todas as contas para enriquecer a proposta na resposta.
	byID := make(map[string]AccountDTO, len(accountDTOs))
	for _, a := range accountDTOs {
		byID[a.AccountID] = a
	}

	// Só as contas de maior saldo vão pro LLM (já vêm ordenadas desc): clientes
	// têm dezenas de contas, a maioria zerada — alocar numa conta sem evolução é
	// inócuo (vira SKIPPED no backtest) e o prompt gigante trava um modelo local.
	top := accountDTOs
	if len(top) > maxLLMAccounts {
		top = top[:maxLLMAccounts]
	}
	accounts := make([]assistant.AccountContext, 0, len(top))
	for _, a := range top {
		accounts = append(accounts, assistant.AccountContext{
			AccountID: a.AccountID,
			BrandName: a.BrandName,
			Type:      a.Type,
			Number:    a.Number,
			Balance:   a.Balance,
		})
	}

	history := make([]assistant.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		role := assistant.RoleUser
		if m.Role == string(assistant.RoleAssistant) {
			role = assistant.RoleAssistant
		}
		history = append(history, assistant.Message{Role: role, Content: m.Content})
	}

	turn, err := s.assistant.Plan(ctx, accounts, s.cfg.POCReferenceDate, history)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "assistant unavailable")
		return
	}

	resp := AssistantPlanResponse{Reply: turn.Reply, Done: turn.Done}
	if turn.Proposal != nil {
		p := &PlanProposalDTO{
			Name:        turn.Proposal.Name,
			StartDate:   turn.Proposal.StartDate,
			Allocations: make([]ProposedAllocationDTO, 0, len(turn.Proposal.Allocations)),
			Plans:       make([]PlanOptionDTO, 0, len(turn.Proposal.Plans)),
		}
		for _, a := range turn.Proposal.Allocations {
			acc := byID[a.AccountID]
			p.Allocations = append(p.Allocations, ProposedAllocationDTO{
				AccountID:  a.AccountID,
				Percentage: a.Percentage,
				Reason:     a.Reason,
				BrandName:  acc.BrandName,
				Number:     acc.Number,
			})
		}
		for _, pl := range turn.Proposal.Plans {
			p.Plans = append(p.Plans, PlanOptionDTO{
				Label:          pl.Label,
				Summary:        pl.Summary,
				TargetAmount:   pl.TargetAmount,
				DurationMonths: pl.DurationMonths,
			})
		}
		resp.Proposal = p
	}

	writeJSON(w, http.StatusOK, resp)
}
