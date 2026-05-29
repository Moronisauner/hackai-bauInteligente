// Package goal contém a regra de negócio para criar e consultar objetivos
// (goals), suas alocações e o baú (vault) — PRD RF-03, RF-04, §7.2.
package goal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// ErrValidation marca falhas de validação de entrada (mapeadas para HTTP 400).
type ErrValidation struct{ Msg string }

func (e ErrValidation) Error() string { return e.Msg }

// validationf cria um ErrValidation formatado.
func validationf(format string, args ...any) error {
	return ErrValidation{Msg: fmt.Sprintf(format, args...)}
}

// Service expõe as operações de objetivos sobre o pool.
type Service struct {
	Pool *pgxpool.Pool
}

// Allocation é uma alocação de percentual sobre uma conta-fonte.
type Allocation struct {
	AccountID  string
	Percentage int
}

// CreateInput agrega os dados necessários para criar um objetivo completo.
type CreateInput struct {
	UserID         string
	Name           string
	TargetAmount   decimal.Decimal
	DurationMonths int
	StartDate      time.Time
	Allocations    []Allocation
}

// Create persiste atomicamente 1 goal, 1 goal_vault e N goal_allocations.
// Retorna ErrValidation para entradas inválidas.
func (s *Service) Create(ctx context.Context, in CreateInput) (goalID, vaultID string, err error) {
	if err := validate(in); err != nil {
		return "", "", err
	}

	// Todos os account_id das alocações precisam pertencer ao usuário.
	accountIDs := make([]string, len(in.Allocations))
	for i, a := range in.Allocations {
		accountIDs[i] = a.AccountID
	}
	if err := s.assertAccountsOwnedBy(ctx, in.UserID, accountIDs); err != nil {
		return "", "", err
	}

	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", "", fmt.Errorf("goal: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback é no-op após commit

	goalID = uuid.NewString()
	vaultID = uuid.NewString()

	_, err = tx.Exec(ctx, `
		INSERT INTO goals (id, user_id, name, target_amount, duration_months, start_date)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		goalID, in.UserID, in.Name, in.TargetAmount, in.DurationMonths, in.StartDate)
	if err != nil {
		return "", "", fmt.Errorf("goal: insert goal: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO goal_vaults (id, goal_id, user_id)
		VALUES ($1, $2, $3)`,
		vaultID, goalID, in.UserID)
	if err != nil {
		return "", "", fmt.Errorf("goal: insert vault: %w", err)
	}

	for _, a := range in.Allocations {
		_, err = tx.Exec(ctx, `
			INSERT INTO goal_allocations (id, goal_id, account_id, percentage)
			VALUES ($1, $2, $3, $4)`,
			uuid.NewString(), goalID, a.AccountID, a.Percentage)
		if err != nil {
			return "", "", fmt.Errorf("goal: insert allocation: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", fmt.Errorf("goal: commit: %w", err)
	}

	return goalID, vaultID, nil
}

// validate aplica as regras de RF-03/RF-04 sobre a entrada.
func validate(in CreateInput) error {
	if in.UserID == "" {
		return validationf("user_id is required")
	}
	if in.Name == "" {
		return validationf("name is required")
	}
	if in.TargetAmount.LessThanOrEqual(decimal.Zero) {
		return validationf("target_amount must be greater than 0")
	}
	if in.DurationMonths < 1 || in.DurationMonths > 60 {
		return validationf("duration_months must be between 1 and 60")
	}
	if len(in.Allocations) < 1 {
		return validationf("at least one allocation is required")
	}
	// O percentual é a fatia da evolução mensal de cada conta (RF-04); as
	// alocações são independentes e NÃO precisam somar 100%.
	for _, a := range in.Allocations {
		if a.AccountID == "" {
			return validationf("allocation account_id is required")
		}
		if a.Percentage < 1 || a.Percentage > 100 {
			return validationf("each allocation percentage must be between 1 and 100")
		}
	}
	return nil
}

// assertAccountsOwnedBy garante que todos os accountIDs existem e pertencem ao
// userID (status AVAILABLE).
func (s *Service) assertAccountsOwnedBy(ctx context.Context, userID string, accountIDs []string) error {
	rows, err := s.Pool.Query(ctx,
		`SELECT id FROM bank_accounts WHERE user_id = $1 AND status = 'AVAILABLE' AND id = ANY($2)`,
		userID, accountIDs)
	if err != nil {
		return fmt.Errorf("goal: checking account ownership: %w", err)
	}
	defer rows.Close()

	owned := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("goal: scanning account id: %w", err)
		}
		owned[id] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("goal: iterating accounts: %w", err)
	}

	for _, id := range accountIDs {
		if !owned[id] {
			return validationf("account %s does not belong to user", id)
		}
	}
	return nil
}

// IsValidation reporta se err é um erro de validação (HTTP 400).
func IsValidation(err error) bool {
	var v ErrValidation
	return errors.As(err, &v)
}
