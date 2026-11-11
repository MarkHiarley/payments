package usecases

import (
	"context"

	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/platform/apperr"
	"github.com/markHiarley/payments/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AccountUseCase struct {
	repo repository.AccountRepository
}

func NewAccountUseCase(repo repository.AccountRepository) *AccountUseCase {
	return &AccountUseCase{
		repo: repo,
	}
}

func (u *AccountUseCase) Create(ctx context.Context, t models.Account) (models.Account, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(t.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.Account{}, apperr.Internal(err)
	}
	t.Password = string(passwordHash)

	return u.repo.Create(ctx, t)
}

func (u *AccountUseCase) FindByEmail(ctx context.Context, email string) (models.Account, error) {
	return u.repo.FindByEmail(ctx, email)
}
