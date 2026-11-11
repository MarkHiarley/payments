package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/platform/apperr"
	"github.com/markHiarley/payments/internal/repository"
)

const testSchemaDDL = `
CREATE TABLE account (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	balance BIGINT NOT NULL CHECK (balance >= 0) DEFAULT 0,
	user_name VARCHAR(120) NOT NULL,
	user_cpf_cnpj VARCHAR(14) UNIQUE NOT NULL,
	user_email VARCHAR(120) NOT NULL UNIQUE,
	password_hash VARCHAR(255) NOT NULL,
	blocked BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	external_id VARCHAR(255) NOT NULL UNIQUE,
	from_account_id UUID NOT NULL,
	to_account_id UUID NOT NULL,
	type VARCHAR(50) NOT NULL,
	amount BIGINT NOT NULL CHECK (amount >= 1),
	currency VARCHAR(10) NOT NULL,
	status VARCHAR(50) NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	CONSTRAINT fk_from_account FOREIGN KEY (from_account_id) REFERENCES account(id),
	CONSTRAINT fk_to_account FOREIGN KEY (to_account_id) REFERENCES account(id)
);`

// TestTransferErrorPrecedence pins the order of the domain checks: a missing
// counterpart is always ACCOUNT_NOT_FOUND, even when the origin cannot fund the
// amount. Requires a reachable PostgreSQL; skipped otherwise.
func TestTransferErrorPrecedence(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewPostgresTransactionRepository(db)
	ctx := context.Background()

	origin := insertAccount(t, db, 0)
	existingDest := insertAccount(t, db, 0)

	t.Run("destino inexistente", func(t *testing.T) {
		err := repo.Transfer(ctx, &models.Transaction{
			ExternalID:    uuid.NewString(),
			FromAccountID: origin,
			ToAccountID:   uuid.New(),
			Type:          models.Transfer,
			Amount:        1000,
			Currency:      "BRL",
		})

		assertAppError(t, err, apperr.KindNotFound, apperr.CodeAccountNotFound)
	})

	t.Run("saldo insuficiente com destino existente", func(t *testing.T) {
		err := repo.Transfer(ctx, &models.Transaction{
			ExternalID:    uuid.NewString(),
			FromAccountID: origin,
			ToAccountID:   existingDest,
			Type:          models.Transfer,
			Amount:        1000,
			Currency:      "BRL",
		})

		assertAppError(t, err, apperr.KindPrecondition, apperr.CodeInsufficientFunds)
	})

	if balance := balanceOf(t, db, origin); balance != 0 {
		t.Fatalf("origin balance after rejected transfers = %d, want 0", balance)
	}

	t.Run("transferência válida", func(t *testing.T) {
		funded := insertAccount(t, db, 5000)
		dest := insertAccount(t, db, 0)

		err := repo.Transfer(ctx, &models.Transaction{
			ExternalID:    uuid.NewString(),
			FromAccountID: funded,
			ToAccountID:   dest,
			Type:          models.Transfer,
			Amount:        1500,
			Currency:      "BRL",
		})
		if err != nil {
			t.Fatalf("Transfer() error = %v, want nil", err)
		}

		if balance := balanceOf(t, db, funded); balance != 3500 {
			t.Errorf("origin balance = %d, want 3500", balance)
		}
		if balance := balanceOf(t, db, dest); balance != 1500 {
			t.Errorf("destination balance = %d, want 1500", balance)
		}
	})
}

func assertAppError(t *testing.T, err error, kind apperr.Kind, code string) {
	t.Helper()

	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("Transfer() error = %v (%T), want *apperr.Error", err, err)
	}
	if appErr.Kind != kind || appErr.Code != code {
		t.Fatalf("Transfer() error = %s/%s, want %s/%s", appErr.Kind, appErr.Code, kind, code)
	}
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := testDSN()
	if dsn == "" {
		t.Skip("integração com PostgreSQL desligada: defina TEST_DATABASE_URL ou POSTGRES_USER/POSTGRES_DB")
	}

	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := admin.PingContext(ctx); err != nil {
		_ = admin.Close()
		t.Skipf("postgres indisponível: %v", err)
	}

	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := admin.ExecContext(context.Background(), "CREATE SCHEMA "+schema); err != nil {
		_ = admin.Close()
		t.Skipf("não foi possível criar o schema de teste: %v", err)
	}

	scoped, err := sql.Open("pgx", withSearchPath(dsn, schema))
	if err != nil {
		_, _ = admin.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		_ = admin.Close()
		t.Fatalf("open scoped database: %v", err)
	}

	t.Cleanup(func() {
		_ = scoped.Close()
		_, _ = admin.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
		_ = admin.Close()
	})

	ddlCtx, cancelDDL := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelDDL()

	if _, err := scoped.ExecContext(ddlCtx, testSchemaDDL); err != nil {
		t.Fatalf("create test tables: %v", err)
	}

	return scoped
}

func testDSN() string {
	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		return dsn
	}

	user := os.Getenv("POSTGRES_USER")
	dbname := os.Getenv("POSTGRES_DB")
	if user == "" || dbname == "" {
		return ""
	}

	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("POSTGRES_PORT_EXTERNAL")
	if port == "" {
		port = "5432"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, os.Getenv("POSTGRES_PASSWORD"), dbname)
}

func withSearchPath(dsn, schema string) string {
	if strings.Contains(dsn, "://") {
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		return dsn + separator + "search_path=" + schema
	}

	return dsn + " search_path=" + schema
}

func insertAccount(t *testing.T, db *sql.DB, balance int64) uuid.UUID {
	t.Helper()

	id := uuid.New()
	unique := strings.ReplaceAll(uuid.NewString(), "-", "")

	_, err := db.Exec(
		`INSERT INTO account (id, balance, user_name, user_cpf_cnpj, user_email, password_hash)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, balance, "Conta de teste", unique[:14], unique+"@example.com", "hash",
	)
	if err != nil {
		t.Fatalf("insert account: %v", err)
	}

	return id
}

func balanceOf(t *testing.T, db *sql.DB, id uuid.UUID) int64 {
	t.Helper()

	var balance int64
	if err := db.QueryRow("SELECT balance FROM account WHERE id = $1", id).Scan(&balance); err != nil {
		t.Fatalf("query balance: %v", err)
	}

	return balance
}
