package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/usecases"
)

type LoginController struct {
	usecase *usecases.LoginUseCase
}

func NewLoginController(usecase *usecases.LoginUseCase) *LoginController {
	return &LoginController{
		usecase: usecase,
	}
}

func (lc *LoginController) AuthenticateUser(c *gin.Context) {
	var input models.LoginUser

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	accessToken, refreshToken, err := lc.usecase.AuthenticateUser(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Credenciais inválidas",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
