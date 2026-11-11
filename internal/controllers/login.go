package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/platform/httpapi"
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
		_ = c.Error(httpapi.BindError(err))
		return
	}

	accessToken, refreshToken, err := lc.usecase.AuthenticateUser(input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
