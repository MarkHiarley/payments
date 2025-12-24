package repository

import (
	"context"
	"database/sql"

	"github.com/markHiarley/payments/internal/models"
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
	status := "COMPLETED"
	queryDebito := "UPDATE account SET balance = balance - $1 WHERE id = $2"
	queryCredito := "UPDATE account SET balance = balance + $1 WHERE id = $2 "
	queryTransa := `INSERT INTO transactions (
             external_id, from_account_id, to_account_id, 
            type, amount, currency, status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, queryDebito, t.Amount, t.FromAccountID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, queryCredito, t.Amount, t.ToAccountID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, queryTransa, t.ExternalID, t.FromAccountID, t.ToAccountID, t.Type, t.Amount, t.Currency, status); err != nil {
		return err
	}

	return tx.Commit()
}
