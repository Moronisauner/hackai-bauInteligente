package httpapi

import (
	"net/http"
)

// UserDTO representa um cliente na listagem, com a contagem de contas.
type UserDTO struct {
	UserID        string `json:"user_id"`
	AccountsCount int    `json:"accounts_count"`
}

// ListUsers lista user_id distintos (contas AVAILABLE) com a contagem de contas
// vinculadas, ordenados por contagem desc. Aceita ?q=<substring> para filtrar.
//
//	@Summary	Lista clientes
//	@Tags		users
//	@Produce	json
//	@Param		q	query		string	false	"Filtro por substring do user_id"
//	@Success	200	{array}		httpapi.UserDTO
//	@Failure	500	{object}	map[string]string
//	@Router		/users [get]
func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	const sql = `
		SELECT user_id, COUNT(*) AS accounts_count
		FROM bank_accounts
		WHERE status = 'AVAILABLE'
		  AND ($1 = '' OR user_id ILIKE '%' || $1 || '%')
		GROUP BY user_id
		ORDER BY accounts_count DESC, user_id ASC`

	rows, err := s.pool.Query(r.Context(), sql, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	defer rows.Close()

	users := make([]UserDTO, 0)
	for rows.Next() {
		var u UserDTO
		if err := rows.Scan(&u.UserID, &u.AccountsCount); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan user")
			return
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read users")
		return
	}

	writeJSON(w, http.StatusOK, users)
}
