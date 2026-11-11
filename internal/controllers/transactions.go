package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/markHiarley/payments/internal/models"
	"github.com/markHiarley/payments/internal/platform/apperr"
	"github.com/markHiarley/payments/internal/platform/httpapi"
	"github.com/markHiarley/payments/internal/usecases"
)

type TransactionController struct {
	usecase *usecases.TransactionUseCase
}

func NewTransactionController(usecase *usecases.TransactionUseCase) *TransactionController {
	return &TransactionController{usecase: usecase}
}

func (tc *TransactionController) Transfer(c *gin.Context) {
	userEmail, ok := c.Get("email")
	email, isString := userEmail.(string)
	if !ok || !isString {
		_ = c.Error(apperr.Unauthorized(apperr.CodeUnauthorized, "usuário não autenticado"))
		return
	}

	var input models.Transaction
	if err := c.ShouldBindJSON(&input); err != nil {
		_ = c.Error(httpapi.BindError(err))
		return
	}

	if err := tc.usecase.Transfer(c.Request.Context(), &input, email); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "transaction processed successfully"})
}
