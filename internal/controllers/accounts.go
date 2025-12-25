package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markHiarley/payments/internal/models"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido: " + err.Error()})
		return
	}

	account, err := ac.usecase.Create(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, usecases.ErrAccountAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Conta já existe",
				"message": "Email ou CPF/CNPJ já cadastrado",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Conta criada com sucesso",
		"account": account,
	})
}
