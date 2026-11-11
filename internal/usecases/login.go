package usecases

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/markHiarley/payments/internal/auth"
	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/platform/apperr"
	"github.com/markHiarley/payments/internal/repository"
)

type LoginUseCase struct {
	repo repository.LoginRepository
}

func NewLoginUseCase(repo repository.LoginRepository) *LoginUseCase {
	return &LoginUseCase{
		repo: repo,
	}
}

func (uc *LoginUseCase) AuthenticateUser(body models.LoginUser) (string, string, error) {
	passwordHash, err := uc.repo.AuthenticateUser(body.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", apperr.Unauthorized(apperr.CodeUnauthorized, "credenciais inválidas")
		}
		return "", "", err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(body.Password)); err != nil {
		return "", "", apperr.Unauthorized(apperr.CodeUnauthorized, "credenciais inválidas")
	}

	accessToken, err := auth.GenerateAccessToken(body.Email)
	if err != nil {
		return "", "", apperr.Internal(err)
	}

	refreshToken, err := auth.GenerateRefreshToken(body.Email)
	if err != nil {
		return "", "", apperr.Internal(err)
	}

	return accessToken, refreshToken, nil
}
