package pgxerr

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/markHiarley/payments/internal/platform/apperr"
)

// constraintFields maps each unique key to the request field it protects.
var constraintFields = map[string]string{
	"account_user_email_key":       "email",
	"account_user_cpf_cnpj_key":    "cpf_cnpj",
	"transactions_external_id_key": "external_id",
}

// Translate converts a PostgreSQL error into a domain error. Every error that is
// not a recognizable PgError becomes internal, so driver text never reaches the
// HTTP layer, and every translated error keeps its cause for logging.
func Translate(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return apperr.Internal(err)
	}

	switch {
	case pgErr.Code == pgerrcode.UniqueViolation:
		conflict := apperr.Conflict(apperr.CodeAlreadyExists, "registro já cadastrado")
		if field, ok := constraintFields[pgErr.ConstraintName]; ok {
			conflict.Meta = map[string]string{"field": field}
		}
		return conflict

	case pgErr.Code == pgerrcode.CheckViolation && pgErr.ConstraintName == "account_balance_check":
		return apperr.Precondition(apperr.CodeInsufficientFunds, "saldo insuficiente na conta de origem")

	case pgErr.Code == pgerrcode.ForeignKeyViolation:
		return apperr.NotFound(apperr.CodeAccountNotFound, "conta referenciada não encontrada")

	case pgErr.Code == pgerrcode.InvalidTextRepresentation:
		return apperr.Validation(apperr.CodeValidation, "identificador com formato inválido")

	case pgerrcode.IsConnectionException(pgErr.Code):
		return apperr.Unavailable(apperr.CodeServiceUnavailable, "banco de dados indisponível")

	default:
		return apperr.Internal(err)
	}
}
