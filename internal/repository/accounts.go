package repository

import (
	"context"
	"database/sql"

	"github.com/markHiarley/payments/internal/models"
)

type AccountRepository interface {
	Create(ctx context.Context, a models.Account) (models.Account, error)
}

type PostgresAccountRepository struct {
	db *sql.DB
}

func NewPostgresAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{
		db: db,
	}
}

func (r *PostgresAccountRepository) Create(ctx context.Context, t models.Account) (models.Account, error) {
	query := `INSERT INTO account (user_name, user_cpf_cnpj, user_email, password_hash) 
	          VALUES ($1, $2, $3, $4) 
	          RETURNING id, balance, created_at;`

	err := r.db.QueryRowContext(ctx, query, t.UserName, t.UserCpfCnpj, t.UserEmail, t.Password).Scan(
		&t.ID,
		&t.Balance,
		&t.CreatedAt,
	)

	if err != nil {
		return models.Account{}, err
	}

	t.Password = ""

	return t, nil
}
