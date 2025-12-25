package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markHiarley/payments/internal/models"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) GetProfile(c *gin.Context) {

	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Usuário não autenticado",
		})
		return
	}

	claims, _ := c.Get("claims")
	jwtClaims := claims.(*models.JWTClaims)

	c.JSON(http.StatusOK, gin.H{
		"email":      email,
		"expires_at": jwtClaims.ExpiresAt.Time,
		"message":    "Perfil do usuário autenticado",
	})
}
