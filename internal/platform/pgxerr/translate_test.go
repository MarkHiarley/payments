package pgxerr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/markHiarley/payments/internal/platform/apperr"
	"github.com/markHiarley/payments/internal/platform/pgxerr"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		name  string
		in    error
		kind  apperr.Kind
		code  string
		field string
	}{
		{
			name:  "unique violation on account email",
			in:    &pgconn.PgError{Code: pgerrcode.UniqueViolation, ConstraintName: "account_user_email_key"},
			kind:  apperr.KindConflict,
			code:  apperr.CodeAlreadyExists,
			field: "email",
		},
		{
			name:  "unique violation on account cpf cnpj",
			in:    &pgconn.PgError{Code: pgerrcode.UniqueViolation, ConstraintName: "account_user_cpf_cnpj_key"},
			kind:  apperr.KindConflict,
			code:  apperr.CodeAlreadyExists,
			field: "cpf_cnpj",
		},
		{
			name:  "unique violation on transaction external id",
			in:    &pgconn.PgError{Code: pgerrcode.UniqueViolation, ConstraintName: "transactions_external_id_key"},
			kind:  apperr.KindConflict,
			code:  apperr.CodeAlreadyExists,
			field: "external_id",
		},
		{
			name: "unique violation on unknown constraint",
			in:   &pgconn.PgError{Code: pgerrcode.UniqueViolation, ConstraintName: "some_future_key"},
			kind: apperr.KindConflict,
			code: apperr.CodeAlreadyExists,
		},
		{
			name: "check violation on account balance",
			in:   &pgconn.PgError{Code: pgerrcode.CheckViolation, ConstraintName: "account_balance_check"},
			kind: apperr.KindPrecondition,
			code: apperr.CodeInsufficientFunds,
		},
		{
			name: "foreign key violation",
			in:   &pgconn.PgError{Code: pgerrcode.ForeignKeyViolation, ConstraintName: "fk_to_account"},
			kind: apperr.KindNotFound,
			code: apperr.CodeAccountNotFound,
		},
		{
			name: "invalid text representation",
			in:   &pgconn.PgError{Code: pgerrcode.InvalidTextRepresentation},
			kind: apperr.KindValidation,
			code: apperr.CodeValidation,
		},
		{
			name: "connection exception",
			in:   &pgconn.PgError{Code: pgerrcode.ConnectionFailure},
			kind: apperr.KindUnavailable,
			code: apperr.CodeServiceUnavailable,
		},
		{
			name: "unmapped postgres error",
			in:   &pgconn.PgError{Code: pgerrcode.SyntaxError},
			kind: apperr.KindInternal,
			code: apperr.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pgxerr.Translate(tt.in)

			var appErr *apperr.Error
			if !errors.As(err, &appErr) {
				t.Fatalf("Translate() returned %T, want *apperr.Error", err)
			}
			if appErr.Kind != tt.kind {
				t.Errorf("Kind = %q, want %q", appErr.Kind, tt.kind)
			}
			if appErr.Code != tt.code {
				t.Errorf("Code = %q, want %q", appErr.Code, tt.code)
			}
			if tt.field == "" {
				if len(appErr.Meta) != 0 {
					t.Errorf("Meta = %v, want empty", appErr.Meta)
				}
				return
			}
			if got := appErr.Meta["field"]; got != tt.field {
				t.Errorf("Meta[field] = %q, want %q", got, tt.field)
			}
		})
	}
}

func TestTranslateWrappedPgError(t *testing.T) {
	pgErr := &pgconn.PgError{Code: pgerrcode.UniqueViolation, ConstraintName: "transactions_external_id_key"}

	appErr := apperr.From(pgxerr.Translate(fmt.Errorf("insert transaction: %w", pgErr)))

	if appErr.Kind != apperr.KindConflict {
		t.Fatalf("Kind = %q, want %q", appErr.Kind, apperr.KindConflict)
	}
	if got := appErr.Meta["field"]; got != "external_id" {
		t.Fatalf("Meta[field] = %q, want %q", got, "external_id")
	}
}

func TestTranslatePreservesCauseForInternalErrors(t *testing.T) {
	cause := errors.New("pgx: connection reset by peer")

	err := pgxerr.Translate(cause)

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("Translate() returned %T, want *apperr.Error", err)
	}
	if appErr.Kind != apperr.KindInternal {
		t.Errorf("Kind = %q, want %q", appErr.Kind, apperr.KindInternal)
	}
	if !errors.Is(err, cause) {
		t.Error("Translate() dropped the original cause")
	}
}

func TestTranslateKeepsCauseOfUnmappedPgError(t *testing.T) {
	pgErr := &pgconn.PgError{Code: pgerrcode.SyntaxError, Message: "syntax error at or near \"slect\""}

	err := pgxerr.Translate(pgErr)

	if !errors.Is(err, pgErr) {
		t.Fatal("Translate() dropped the original PgError cause")
	}
}

func TestTranslateNil(t *testing.T) {
	if err := pgxerr.Translate(nil); err != nil {
		t.Fatalf("Translate(nil) = %v, want nil", err)
	}
}
