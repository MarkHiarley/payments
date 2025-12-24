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

// errors var

var ErrDuplicateTransaction = errors.New("This transaction is still being processed")
var ErrTransactionConcluded = errors.New("This transaction has already been completed, please wait a moment")
var ErrTransactionConcludedDB = errors.New("Transaction has already been processed previously (database conflict)")

type TransactionUseCase struct {
	repo      repository.TransactionRepository
	rediStore *cache.RedisTransactionStore
}

func NewTransactionUsecase(repo repository.TransactionRepository, rediStore *cache.RedisTransactionStore) *TransactionUseCase {
	return &TransactionUseCase{
		repo:      repo,
		rediStore: rediStore,
	}
}

func (r *TransactionUseCase) Transfer(ctx context.Context, t *models.Transaction) error {

	IsNew, err := r.rediStore.IsNew(ctx, t.ExternalID, "PROCESSING")

	if !IsNew {
		status, err := r.rediStore.Get(ctx, t.ExternalID).Result()
		if err != nil {
			return ErrDuplicateTransaction
		}

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

		return fmt.Errorf("error transferring: %w", err)
	}

	_, err = r.rediStore.SetStatusCompleted(ctx, t.ExternalID).Result()

	if err != nil {
		_ = r.rediStore.Delete(ctx, t.ExternalID)
		return fmt.Errorf("error updating transaction status in cache: %w", err)
	}

	return nil
}
