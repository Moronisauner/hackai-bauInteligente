package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

// AccountDTO representa uma conta do cliente com o saldo reconstruído na
// POC_REFERENCE_DATE.
type AccountDTO struct {
	AccountID            string `json:"account_id"`
	BrandName            string `json:"brand_name"`
	Type                 string `json:"type"`
	BranchCode           string `json:"branch_code"`
	Number               string `json:"number"`
	CheckDigit           string `json:"check_digit"`
	CompeCode            string `json:"compe_code"`
	Balance              string `json:"balance" example:"1234.56"`
	BalanceReferenceDate string `json:"balance_reference_date" example:"2025-01-01"`
}

// ListAccountsByUser lista as contas AVAILABLE de um cliente, cada uma com o
// saldo reconstruído na POC_REFERENCE_DATE, ordenadas pelo saldo (maior para o
// menor). Cliente sem contas → 200 [].
//
//	@Summary	Lista contas de um cliente com saldo
//	@Tags		users
//	@Produce	json
//	@Param		userID	path		string	true	"ID do cliente"
//	@Success	200		{array}		httpapi.AccountDTO
//	@Failure	500		{object}	map[string]string
//	@Router		/users/{userID}/accounts [get]
func (s *Server) ListAccountsByUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	out, err := s.loadAccounts(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// loadAccounts devolve as contas AVAILABLE de um cliente com o saldo
// reconstruído na POC_REFERENCE_DATE, ordenadas do maior saldo para o menor.
// Reaproveitado pelo handler de listagem e pelo assistente de planejamento.
func (s *Server) loadAccounts(ctx context.Context, userID string) ([]AccountDTO, error) {
	const sql = `
		SELECT id, brand_name, type, branch_code, number, check_digit, compe_code
		FROM bank_accounts
		WHERE user_id = $1 AND status = 'AVAILABLE'
		ORDER BY brand_name, number`

	rows, err := s.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts")
	}
	defer rows.Close()

	type accountRow struct {
		id, brandName, accType, branchCode, number, checkDigit, compeCode *string
	}
	var accounts []accountRow
	var ids []string
	for rows.Next() {
		var a accountRow
		if err := rows.Scan(&a.id, &a.brandName, &a.accType, &a.branchCode, &a.number, &a.checkDigit, &a.compeCode); err != nil {
			return nil, fmt.Errorf("failed to scan account")
		}
		accounts = append(accounts, a)
		ids = append(ids, deref(a.id))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read accounts")
	}

	balances, err := s.balance.ReconstructMany(ctx, ids, s.cfg.POCReferenceDate)
	if err != nil {
		return nil, fmt.Errorf("failed to compute balances")
	}

	refDate := s.cfg.POCReferenceDate.Format("2006-01-02")
	out := make([]AccountDTO, 0, len(accounts))
	for _, a := range accounts {
		id := deref(a.id)
		bal, ok := balances[id]
		if !ok {
			bal = decimal.Zero
		}
		out = append(out, AccountDTO{
			AccountID:            id,
			BrandName:            deref(a.brandName),
			Type:                 deref(a.accType),
			BranchCode:           deref(a.branchCode),
			Number:               deref(a.number),
			CheckDigit:           deref(a.checkDigit),
			CompeCode:            deref(a.compeCode),
			Balance:              bal.StringFixed(2),
			BalanceReferenceDate: refDate,
		})
	}

	// Ordena pelo saldo do maior para o menor (a query não pode ordenar pois o
	// saldo é reconstruído em memória a partir das transações).
	sort.SliceStable(out, func(i, j int) bool {
		return balances[out[i].AccountID].GreaterThan(balances[out[j].AccountID])
	})

	return out, nil
}

// deref devolve o valor de um *string ou "" se nil.
func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
