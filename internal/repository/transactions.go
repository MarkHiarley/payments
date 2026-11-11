package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/platform/apperr"
	"github.com/markHiarley/payments/internal/platform/pgxerr"
)

type TransactionRepository interface {
	Transfer(ctx context.Context, t *models.Transaction) error
}

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{
		db: db,
	}
}

func (r *PostgresTransactionRepository) Transfer(ctx context.Context, t *models.Transaction) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return pgxerr.Translate(err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	// Existence of both accounts is checked before any balance rule so a
	// missing counterpart always reports ACCOUNT_NOT_FOUND, never 422.
	if err := lockAccount(ctx, tx, t.FromAccountID, "origem"); err != nil {
		return err
	}
	if err := lockAccount(ctx, tx, t.ToAccountID, "destino"); err != nil {
		return err
	}

	status := "COMPLETED"
	queryDebito := "UPDATE account SET balance = balance - $1 WHERE id = $2"
	queryCredito := "UPDATE account SET balance = balance + $1 WHERE id = $2 "
	queryTransa := `INSERT INTO transactions (
             external_id, from_account_id, to_account_id, 
            type, amount, currency, status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if _, err := tx.ExecContext(ctx, queryDebito, t.Amount, t.FromAccountID); err != nil {
		return pgxerr.Translate(err)
	}

	if _, err := tx.ExecContext(ctx, queryCredito, t.Amount, t.ToAccountID); err != nil {
		return pgxerr.Translate(err)
	}

	if _, err := tx.ExecContext(ctx, queryTransa, t.ExternalID, t.FromAccountID, t.ToAccountID, t.Type, t.Amount, t.Currency, status); err != nil {
		return pgxerr.Translate(err)
	}

	if err := tx.Commit(); err != nil {
		return pgxerr.Translate(err)
	}

	return nil
}

func lockAccount(ctx context.Context, tx *sql.Tx, id uuid.UUID, role string) error {
	var locked uuid.UUID

	err := tx.QueryRowContext(ctx, "SELECT id FROM account WHERE id = $1 FOR SHARE", id).Scan(&locked)
	if errors.Is(err, sql.ErrNoRows) {
		return apperr.NotFound(apperr.CodeAccountNotFound, "conta de "+role+" não encontrada")
	}
	if err != nil {
		return pgxerr.Translate(err)
	}

	return nil
}
