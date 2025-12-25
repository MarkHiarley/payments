package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/markHiarley/payments/internal/cache"
	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/repository"
)

var (
	ErrDuplicateTransaction   = errors.New("transaction still being processed")
	ErrTransactionConcluded   = errors.New("transaction already completed")
	ErrTransactionConcludedDB = errors.New("transaction conflict in database")
	ErrAccountNotFound        = errors.New("origin account not found for this user")
)

type TransactionUseCase struct {
	repo      repository.TransactionRepository
	accRepo   repository.AccountRepository
	rediStore *cache.RedisTransactionStore
}

func NewTransactionUsecase(repo repository.TransactionRepository, accRepo repository.AccountRepository, rediStore *cache.RedisTransactionStore) *TransactionUseCase {
	return &TransactionUseCase{
		repo:      repo,
		accRepo:   accRepo,
		rediStore: rediStore,
	}
}

func (r *TransactionUseCase) Transfer(ctx context.Context, t *models.Transaction, email string) error {

	acc, err := r.accRepo.FindByEmail(ctx, email)
	if err != nil {
		return ErrAccountNotFound
	}
	t.FromAccountID = acc.ID

	isNew, err := r.rediStore.IsNew(ctx, t.ExternalID, "PROCESSING")
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}

	if !isNew {
		status, _ := r.rediStore.Get(ctx, t.ExternalID).Result()
		if status == "COMPLETED" {
			return ErrTransactionConcluded
		}
		return ErrDuplicateTransaction
	}

	if err := r.repo.Transfer(ctx, t); err != nil {
		_ = r.rediStore.Delete(ctx, t.ExternalID)
		if strings.Contains(err.Error(), "unique constraint") {
			return ErrTransactionConcludedDB
		}
		return err
	}

	_ = r.rediStore.SetStatusCompleted(ctx, t.ExternalID)
	return nil
}
