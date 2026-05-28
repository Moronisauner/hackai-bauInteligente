package httpapi

import (
	"net/http"
)

// UserDTO representa um cliente na listagem, com o nome legível e a contagem de
// contas.
type UserDTO struct {
	UserID        string `json:"user_id"`
	Nome          string `json:"nome"`
	AccountsCount int    `json:"accounts_count"`
}

// ListUsers lista os clientes (contas AVAILABLE) com o nome legível e a
// contagem de contas vinculadas, ordenados pelo nome. Aceita ?q=<substring>
// para filtrar pelo nome.
//
//	@Summary	Lista clientes
//	@Tags		users
//	@Produce	json
//	@Param		q	query		string	false	"Filtro por substring do nome"
//	@Success	200	{array}		httpapi.UserDTO
//	@Failure	500	{object}	map[string]string
//	@Router		/users [get]
func (s *Server) ListUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	const sql = `
		SELECT ba.user_id, COALESCE(c.nome, ba.user_id) AS nome, COUNT(*) AS accounts_count
		FROM bank_accounts ba
		LEFT JOIN clientes c ON c.user_id = ba.user_id
		WHERE ba.status = 'AVAILABLE'
		  AND ($1 = '' OR COALESCE(c.nome, ba.user_id) ILIKE '%' || $1 || '%')
		GROUP BY ba.user_id, c.nome
		ORDER BY nome ASC, ba.user_id ASC`

	rows, err := s.pool.Query(r.Context(), sql, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	defer rows.Close()

	users := make([]UserDTO, 0)
	for rows.Next() {
		var u UserDTO
		if err := rows.Scan(&u.UserID, &u.Nome, &u.AccountsCount); err != nil {
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
