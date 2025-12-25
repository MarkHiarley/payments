package usecases

import (
	"github.com/markHiarley/payments/internal/auth"
	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/repository"
	"golang.org/x/crypto/bcrypt"
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
		return "", "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(body.Password))
	if err != nil {
		return "", "", err
	}

	accessToken, err := auth.GenerateAccessToken(body.Email)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := auth.GenerateRefreshToken(body.Email)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
