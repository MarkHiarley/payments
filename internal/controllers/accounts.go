package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/platform/httpapi"
	"github.com/markHiarley/payments/internal/usecases"
)

type AccountController struct {
	usecase *usecases.AccountUseCase
}

func NewAccountController(usecase *usecases.AccountUseCase) *AccountController {
	return &AccountController{
		usecase: usecase,
	}
}

func (ac *AccountController) Create(c *gin.Context) {
	var input models.Account

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(httpapi.BindError(err))
		return
	}

	account, err := ac.usecase.Create(c.Request.Context(), input)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Conta criada com sucesso",
		"account": account,
	})
}
