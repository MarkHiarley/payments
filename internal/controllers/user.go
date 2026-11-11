package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/platform/apperr"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) GetProfile(c *gin.Context) {
	email, ok := c.Get("email")
	userEmail, isString := email.(string)
	if !ok || !isString {
		_ = c.Error(apperr.Unauthorized(apperr.CodeUnauthorized, "usuário não autenticado"))
		return
	}

	rawClaims, ok := c.Get("claims")
	if !ok {
		_ = c.Error(apperr.Unauthorized(apperr.CodeUnauthorized, "credenciais de sessão ausentes"))
		return
	}

	jwtClaims, ok := rawClaims.(*models.JWTClaims)
	if !ok {
		_ = c.Error(apperr.Unauthorized(apperr.CodeUnauthorized, "credenciais de sessão inválidas"))
		return
	}

	var expiresAt time.Time
	if jwtClaims.ExpiresAt != nil {
		expiresAt = jwtClaims.ExpiresAt.Time
	}

	c.JSON(http.StatusOK, gin.H{
		"email":      userEmail,
		"expires_at": expiresAt,
		"message":    "Perfil do usuário autenticado",
	})
}
