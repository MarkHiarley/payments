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

func NewAccountController(uc *usecases.AccountUseCase) *AccountController {
	return &AccountController{
		usecase: uc,
	}
}

func (ac *AccountController) Create(c *gin.Context) {
	var body models.Account

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON: " + err.Error()})
		return
	}

	account, err := ac.usecase.Create(c.Request.Context(), body)
	if err != nil {

		if errors.Is(err, usecases.ErrAccountAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Account already exists",
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"account": account,
	})
}
