package usecases

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/redis/go-redis/v9"

	"github.com/markHiarley/payments/internal/cache"
	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/platform/apperr"
	"github.com/markHiarley/payments/internal/repository"
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
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.NotFound(apperr.CodeAccountNotFound, "conta de origem não encontrada")
		}
		return err
	}
	t.FromAccountID = acc.ID

	isNew, err := r.rediStore.IsNew(ctx, t.ExternalID, "PROCESSING")
	if err != nil {
		return apperr.Unavailable(apperr.CodeServiceUnavailable, "serviço de idempotência indisponível")
	}

	if !isNew {
		status, err := r.rediStore.Get(ctx, t.ExternalID).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return apperr.Unavailable(apperr.CodeServiceUnavailable, "serviço de idempotência indisponível")
		}

		if status == string(models.Completed) {
			return apperr.Conflict(apperr.CodeDuplicateExternalID, "transação já processada para este external_id")
		}
		return apperr.Conflict(apperr.CodeIdempotency, "transação com este external_id já está em processamento")
	}

	if err := r.repo.Transfer(ctx, t); err != nil {
		if delErr := r.rediStore.Delete(ctx, t.ExternalID); delErr != nil {
			slog.Warn("failed to release idempotency key", "external_id", t.ExternalID, "error", delErr)
		}

		if appErr := apperr.From(err); appErr.Kind == apperr.KindConflict && appErr.Meta["field"] == "external_id" {
			return apperr.Conflict(apperr.CodeDuplicateExternalID, "transação já processada para este external_id")
		}
		return err
	}

	if err := r.rediStore.SetStatusCompleted(ctx, t.ExternalID); err != nil {
		slog.Warn("failed to mark idempotency key as completed", "external_id", t.ExternalID, "error", err)
	}

	return nil
}
