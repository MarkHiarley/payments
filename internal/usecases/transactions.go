package usecases

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/markHiarley/payments/internal/middleware"
	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/repository"
)

var ErrDuplicateTransaction = errors.New("transação duplicada ou já em processamento")

type TransactionUseCase struct {
	repo      repository.TransactionRepository
	rediStore *middleware.RedisTransactionStore
}

func NewTransactionUsecase(repo repository.TransactionRepository, rediStore *middleware.RedisTransactionStore) *TransactionUseCase {
	return &TransactionUseCase{
		repo:      repo,
		rediStore: rediStore,
	}
}

func (r *TransactionUseCase) Transfer(ctx context.Context, t *models.Transaction) error {
	rawKey := fmt.Sprintf("%s-%s-%d", t.FromAccountID, t.ToAccountID, t.Amount)

	hash := sha256.Sum256([]byte(rawKey))

	idempotencyKey := fmt.Sprintf("%x", hash)

	IsNew, err := r.rediStore.IsNew(ctx, idempotencyKey, "PROCESSING")
	if err != nil {
		return fmt.Errorf("erro ao verificar idempotência: %w", err)
	}

	if !IsNew {
		return ErrDuplicateTransaction
	}

	if err := r.repo.Transfer(ctx, t); err != nil {
		_ = r.rediStore.Delete(ctx, idempotencyKey)
		return fmt.Errorf("erro ao transferir: %w", err)
	}

	return nil
}
