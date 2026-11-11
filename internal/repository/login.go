package repository

import (
	"database/sql"
	"errors"

	"github.com/markHiarley/payments/internal/platform/pgxerr"
)

type LoginRepository interface {
	AuthenticateUser(email string) (string, error)
}

type PostgresLoginRepository struct {
	db *sql.DB
}

func NewPostgresLoginRepository(db *sql.DB) *PostgresLoginRepository {
	return &PostgresLoginRepository{
		db: db,
	}
}

func (r *PostgresLoginRepository) AuthenticateUser(email string) (string, error) {
	var passwordHash string
	query := "SELECT password_hash FROM account WHERE user_email = $1"

	err := r.db.QueryRow(query, email).Scan(&passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
		return "", pgxerr.Translate(err)
	}

	return passwordHash, nil
}
