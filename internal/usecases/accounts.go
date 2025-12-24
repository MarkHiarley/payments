package usecases

import (
	"context"
	"errors"

	"strings"

	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrAccountAlreadyExists = errors.New("Account with given email or CPF/CNPJ already exists")

type AccountUseCase struct {
	repo repository.AccountRepository
}

func NewAccountUseCase(repo repository.AccountRepository) *AccountUseCase {
	return &AccountUseCase{
		repo: repo,
	}
}

func (u *AccountUseCase) Create(ctx context.Context, t models.Account) (models.Account, error) {
	var account models.Account
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(t.Password), bcrypt.DefaultCost)
	if err != nil {
		return account, err
	}
	t.Password = string(passwordHash)

	account, err = u.repo.Create(ctx, t)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return models.Account{}, ErrAccountAlreadyExists
		}
		return models.Account{}, err
	}

	return account, nil
}
